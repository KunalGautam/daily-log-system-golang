package entries

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EntryFilter struct {
	EntryType  string   `form:"entry_type"`
	Visibility string   `form:"visibility"`
	StartDate  string   `form:"start_date"`
	EndDate    string   `form:"end_date"`
	TagIDs     []string `form:"tag_ids"`
	MoodMin    int      `form:"mood_min"`
	MoodMax    int      `form:"mood_max"`
	Search     string   `form:"search"`
	Sort       string   `form:"sort"`
	SortDir    string   `form:"sort_dir"`
	Page       int      `form:"page"`
	PageSize   int      `form:"page_size"`
}

type PublicFilter struct {
	EntryType string `form:"entry_type"`
	Tag       string `form:"tag"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	Search    string `form:"search"`
}

type PublicStats struct {
	TotalEntries     int64          `json:"total_entries"`
	TotalUsers       int64          `json:"total_users"`
	EntriesByType    map[string]int64 `json:"entries_by_type"`
	MoodDistribution map[int]int64  `json:"mood_distribution"`
	RecentEntries    int64          `json:"recent_entries"`
}

type HeatmapData struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type EntryService struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *EntryService {
	return &EntryService{db: db}
}

func (s *EntryService) Create(entry *Entry) error {
	return s.db.Create(entry).Error
}

func (s *EntryService) GetByID(id uuid.UUID) (*Entry, error) {
	var entry Entry
	err := s.db.Preload("Tags").Preload("Attachments").Where("id = ?", id).First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *EntryService) Update(entry *Entry) error {
	return s.db.Model(&Entry{}).Where("id = ? AND user_id = ?", entry.ID, entry.UserID).Updates(map[string]interface{}{
		"title":             entry.Title,
		"description":       entry.Description,
		"markdown_content":  entry.MarkdownContent,
		"mood_score":        entry.MoodScore,
		"activities":        entry.Activities,
		"visibility":        entry.Visibility,
		"entry_type":        entry.EntryType,
		"entry_date":        entry.EntryDate,
		"metadata":          entry.Metadata,
		"is_edited":         true,
	}).Error
}

func (s *EntryService) Delete(id, userID uuid.UUID) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Entry{}).Error
}

func (s *EntryService) List(userID uuid.UUID, filter EntryFilter) ([]Entry, int64, error) {
	query := s.db.Model(&Entry{}).Where("user_id = ?", userID)

	if filter.EntryType != "" {
		query = query.Where("entry_type = ?", filter.EntryType)
	}
	if filter.Visibility != "" {
		query = query.Where("visibility = ?", filter.Visibility)
	}
	if filter.StartDate != "" {
		query = query.Where("entry_date >= ?", filter.StartDate)
	}
	if filter.EndDate != "" {
		query = query.Where("entry_date <= ?", filter.EndDate)
	}
	if filter.MoodMin > 0 {
		query = query.Where("mood_score >= ?", filter.MoodMin)
	}
	if filter.MoodMax > 0 {
		query = query.Where("mood_score <= ?", filter.MoodMax)
	}
	if filter.Search != "" {
		search := "%" + filter.Search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ? OR markdown_content ILIKE ?", search, search, search)
	}
	if len(filter.TagIDs) > 0 {
		query = query.Joins("JOIN entry_tags ON entry_tags.entry_id = entries.id").
			Where("entry_tags.tag_id IN ?", filter.TagIDs)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortField := "created_at"
	if filter.Sort != "" {
		allowed := map[string]bool{"created_at": true, "entry_date": true, "mood_score": true}
		if allowed[filter.Sort] {
			sortField = filter.Sort
		}
	}
	sortDir := "DESC"
	if filter.SortDir == "asc" {
		sortDir = "ASC"
	}

	page := filter.Page
	if page < 1 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var entries []Entry
	err := query.Preload("Tags").Preload("Attachments").
		Order(fmt.Sprintf("%s %s", sortField, sortDir)).
		Offset(offset).Limit(pageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (s *EntryService) GetPublicFeed(page, pageSize int) ([]Entry, int64, error) {
	query := s.db.Model(&Entry{}).Where("visibility = ?", VisibilityPublic)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var entries []Entry
	err := query.Preload("Tags").Preload("Attachments").
		Order("entry_date DESC").
		Offset(offset).Limit(pageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (s *EntryService) GetPublicTimeline(page, pageSize int, filters PublicFilter) ([]Entry, int64, error) {
	query := s.db.Model(&Entry{}).Where("visibility = ?", VisibilityPublic)

	if filters.EntryType != "" {
		query = query.Where("entry_type = ?", filters.EntryType)
	}
	if filters.StartDate != "" {
		query = query.Where("entry_date >= ?", filters.StartDate)
	}
	if filters.EndDate != "" {
		query = query.Where("entry_date <= ?", filters.EndDate)
	}
	if filters.Search != "" {
		search := "%" + filters.Search + "%"
		query = query.Where("title ILIKE ? OR description ILIKE ? OR markdown_content ILIKE ?", search, search, search)
	}
	if filters.Tag != "" {
		query = query.Joins("JOIN entry_tags ON entry_tags.entry_id = entries.id").
			Joins("JOIN tags ON tags.id = entry_tags.tag_id").
			Where("tags.name = ?", filters.Tag)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var entries []Entry
	err := query.Preload("Tags").Preload("Attachments").
		Order("entry_date DESC").
		Offset(offset).Limit(pageSize).
		Find(&entries).Error
	if err != nil {
		return nil, 0, err
	}

	return entries, total, nil
}

func (s *EntryService) GetPublicStats() (*PublicStats, error) {
	stats := &PublicStats{
		EntriesByType:    make(map[string]int64),
		MoodDistribution: make(map[int]int64),
	}

	if err := s.db.Model(&Entry{}).Where("visibility = ?", VisibilityPublic).Count(&stats.TotalEntries).Error; err != nil {
		return nil, err
	}

	if err := s.db.Model(&Entry{}).Where("visibility = ?", VisibilityPublic).
		Select("COUNT(DISTINCT user_id)").Scan(&stats.TotalUsers).Error; err != nil {
		return nil, err
	}

	type typeCount struct {
		EntryType string
		Count     int64
	}
	var byType []typeCount
	if err := s.db.Model(&Entry{}).Where("visibility = ?", VisibilityPublic).
		Select("entry_type, COUNT(*) as count").
		Group("entry_type").Scan(&byType).Error; err != nil {
		return nil, err
	}
	for _, tc := range byType {
		stats.EntriesByType[tc.EntryType] = tc.Count
	}

	type moodCount struct {
		MoodScore int
		Count     int64
	}
	var byMood []moodCount
	if err := s.db.Model(&Entry{}).Where("visibility = ? AND mood_score IS NOT NULL", VisibilityPublic).
		Select("mood_score, COUNT(*) as count").
		Group("mood_score").Scan(&byMood).Error; err != nil {
		return nil, err
	}
	for _, mc := range byMood {
		stats.MoodDistribution[mc.MoodScore] = mc.Count
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	if err := s.db.Model(&Entry{}).Where("visibility = ? AND created_at >= ?", VisibilityPublic, sevenDaysAgo).
		Count(&stats.RecentEntries).Error; err != nil {
		return nil, err
	}

	return stats, nil
}

func (s *EntryService) GetHeatmapData(year int) ([]HeatmapData, error) {
	startDate := fmt.Sprintf("%d-01-01", year)
	endDate := fmt.Sprintf("%d-12-31", year)

	type result struct {
		Date  string
		Count int64
	}
	var rows []result
	err := s.db.Model(&Entry{}).Where("visibility = ? AND entry_date >= ? AND entry_date <= ?", VisibilityPublic, startDate, endDate).
		Select("TO_CHAR(entry_date, 'YYYY-MM-DD') as date, COUNT(*) as count").
		Group("date").
		Order("date ASC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	data := make([]HeatmapData, len(rows))
	for i, r := range rows {
		data[i] = HeatmapData{Date: r.Date, Count: int(r.Count)}
	}
	return data, nil
}

func (s *EntryService) GetRSSFeed() (string, error) {
	entries, _, err := s.GetPublicFeed(1, 50)
	if err != nil {
		return "", err
	}

	now := time.Now().Format(time.RFC1123Z)
	var items strings.Builder
	for _, e := range entries {
		title := "Untitled"
		if e.Title != nil {
			title = *e.Title
		}
		description := ""
		if e.Description != nil {
			description = *e.Description
		}
		pubDate := e.CreatedAt.Format(time.RFC1123Z)
		guid := e.ID.String()
		items.WriteString(fmt.Sprintf(`    <item>
      <title><![CDATA[%s]]></title>
      <description><![CDATA[%s]]></description>
      <pubDate>%s</pubDate>
      <guid>%s</guid>
    </item>
`, title, description, pubDate, guid))
	}

	rss := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Life Log Public Feed</title>
    <description>Public entries from Life Log</description>
    <lastBuildDate>%s</lastBuildDate>
    <link>https://lifelog.app/public/feed</link>
%s  </channel>
</rss>`, now, items.String())

	return rss, nil
}

func (s *EntryService) GetJSONFeed() (string, error) {
	entries, _, err := s.GetPublicFeed(1, 50)
	if err != nil {
		return "", err
	}

	items := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		title := "Untitled"
		if e.Title != nil {
			title = *e.Title
		}
		description := ""
		if e.Description != nil {
			description = *e.Description
		}
		items = append(items, map[string]interface{}{
			"id":            e.ID.String(),
			"url":           fmt.Sprintf("https://lifelog.app/public/entries/%s", e.ID),
			"title":         title,
			"content_text":  description,
			"date_published": e.CreatedAt.Format(time.RFC3339),
		})
	}

	feed := map[string]interface{}{
		"version":     "https://jsonfeed.org/version/1",
		"title":       "Life Log Public Feed",
		"description": "Public entries from Life Log",
		"home_page_url": "https://lifelog.app",
		"feed_url":     "https://lifelog.app/public/feed.json",
		"items":        items,
	}

	b, err := json.MarshalIndent(feed, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (s *EntryService) CreateTag(tag *Tag) error {
	return s.db.Create(tag).Error
}

func (s *EntryService) GetTags(userID uuid.UUID) ([]Tag, error) {
	var tags []Tag
	err := s.db.Where("user_id = ? OR user_id IS NULL", userID).Order("name ASC").Find(&tags).Error
	return tags, err
}

func (s *EntryService) DeleteTag(id, userID uuid.UUID) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Tag{}).Error
}

func (s *EntryService) AttachFile(entryID, userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (*Attachment, error) {
	uploadDir := filepath.Join(".", "uploads", userID.String(), entryID.String())
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	filePath := filepath.Join(uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return nil, err
	}

	mimeType := header.Header.Get("Content-Type")
	attachment := &Attachment{
		EntryID:  &entryID,
		UserID:   userID,
		FileName: header.Filename,
		FilePath: filePath,
		FileSize: header.Size,
		MimeType: mimeType,
		IsImage:  strings.HasPrefix(mimeType, "image/"),
	}

	if err := s.db.Create(attachment).Error; err != nil {
		return nil, err
	}

	return attachment, nil
}

func getUserID(c *gin.Context) (uuid.UUID, error) {
	uid, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, fmt.Errorf("user ID not found in context")
	}
	uidStr, ok := uid.(string)
	if !ok {
		return uuid.Nil, fmt.Errorf("user ID is not a string")
	}
	return uuid.Parse(uidStr)
}

type createEntryRequest struct {
	EntryType       string      `json:"entry_type" binding:"required"`
	Title           *string     `json:"title"`
	Description     *string     `json:"description"`
	MarkdownContent *string     `json:"markdown_content"`
	MoodScore       *int        `json:"mood_score"`
	Activities      interface{} `json:"activities"`
	Visibility      string      `json:"visibility"`
	TagIDs          []string    `json:"tag_ids"`
	EntryDate       string      `json:"entry_date"`
	Metadata        interface{} `json:"metadata"`
}

func (s *EntryService) HandleCreate(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req createEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entry := &Entry{
		UserID:          userID,
		EntryType:       req.EntryType,
		Title:           req.Title,
		Description:     req.Description,
		MarkdownContent: req.MarkdownContent,
		MoodScore:       req.MoodScore,
		Visibility:      req.Visibility,
	}

	if entry.Visibility == "" {
		entry.Visibility = VisibilityPrivate
	}

	if req.EntryDate != "" {
		parsed, err := time.Parse("2006-01-02", req.EntryDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry_date format, expected YYYY-MM-DD"})
			return
		}
		entry.EntryDate = parsed
	} else {
		entry.EntryDate = time.Now().Truncate(24 * time.Hour)
	}

	if req.Activities != nil {
		switch v := req.Activities.(type) {
		case string:
			entry.Activities = &v
		case []interface{}:
			parts := make([]string, len(v))
			for i, a := range v {
				parts[i] = fmt.Sprintf("%v", a)
			}
			joined := strings.Join(parts, ",")
			entry.Activities = &joined
		}
	}

	if req.Metadata != nil {
		switch v := req.Metadata.(type) {
		case string:
			entry.Metadata = &v
		default:
			b, err := json.Marshal(v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metadata"})
				return
			}
			s := string(b)
			entry.Metadata = &s
		}
	}

	if err := s.Create(entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(req.TagIDs) > 0 {
		for _, tid := range req.TagIDs {
			tagID, err := uuid.Parse(tid)
			if err != nil {
				continue
			}
			s.db.Create(&EntryTag{EntryID: entry.ID, TagID: tagID})
		}
	}

	s.db.Preload("Tags").Preload("Attachments").First(entry, entry.ID)
	c.JSON(http.StatusCreated, entry)
}

func (s *EntryService) HandleGet(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry ID"})
		return
	}

	entry, err := s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}

	if entry.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, entry)
}

func (s *EntryService) HandleUpdate(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry ID"})
		return
	}

	existing, err := s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}
	if existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req createEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != nil {
		existing.Title = req.Title
	}
	if req.Description != nil {
		existing.Description = req.Description
	}
	if req.MarkdownContent != nil {
		existing.MarkdownContent = req.MarkdownContent
	}
	if req.MoodScore != nil {
		existing.MoodScore = req.MoodScore
	}
	if req.EntryType != "" {
		existing.EntryType = req.EntryType
	}
	if req.Visibility != "" {
		existing.Visibility = req.Visibility
	}
	if req.EntryDate != "" {
		parsed, err := time.Parse("2006-01-02", req.EntryDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry_date format"})
			return
		}
		existing.EntryDate = parsed
	}
	if req.Activities != nil {
		switch v := req.Activities.(type) {
		case string:
			existing.Activities = &v
		case []interface{}:
			parts := make([]string, len(v))
			for i, a := range v {
				parts[i] = fmt.Sprintf("%v", a)
			}
			joined := strings.Join(parts, ",")
			existing.Activities = &joined
		}
	}
	if req.Metadata != nil {
		switch v := req.Metadata.(type) {
		case string:
			existing.Metadata = &v
		default:
			b, err := json.Marshal(v)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid metadata"})
				return
			}
			s := string(b)
			existing.Metadata = &s
		}
	}

	if err := s.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(req.TagIDs) > 0 {
		s.db.Where("entry_id = ?", existing.ID).Delete(&EntryTag{})
		for _, tid := range req.TagIDs {
			tagID, err := uuid.Parse(tid)
			if err != nil {
				continue
			}
			s.db.Create(&EntryTag{EntryID: existing.ID, TagID: tagID})
		}
	}

	s.db.Preload("Tags").Preload("Attachments").First(existing, existing.ID)
	c.JSON(http.StatusOK, existing)
}

func (s *EntryService) HandleDelete(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry ID"})
		return
	}

	if err := s.Delete(id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "entry not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "entry deleted"})
}

func (s *EntryService) HandleList(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var filter EntryFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entries, total, err := s.List(userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  entries,
		"total": total,
	})
}

func (s *EntryService) HandleCreateTag(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var tag Tag
	if err := c.ShouldBindJSON(&tag); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tag.UserID = &userID

	if err := s.CreateTag(&tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tag)
}

func (s *EntryService) HandleListTags(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	tags, err := s.GetTags(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tags)
}

func (s *EntryService) HandleDeleteTag(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tag ID"})
		return
	}

	if err := s.DeleteTag(id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "tag not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "tag deleted"})
}

func (s *EntryService) HandleUploadAttachment(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	entryID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid entry ID"})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	attachment, err := s.AttachFile(entryID, userID, file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, attachment)
}

func (s *EntryService) HandlePublicFeed(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	entries, total, err := s.GetPublicFeed(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  entries,
		"total": total,
	})
}

func (s *EntryService) HandlePublicTimeline(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var filter PublicFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entries, total, err := s.GetPublicTimeline(page, pageSize, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  entries,
		"total": total,
	})
}

func (s *EntryService) HandlePublicStats(c *gin.Context) {
	stats, err := s.GetPublicStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func (s *EntryService) HandlePublicHeatmap(c *gin.Context) {
	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}

	data, err := s.GetHeatmapData(year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (s *EntryService) HandleRSS(c *gin.Context) {
	rss, err := s.GetRSSFeed()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/rss+xml")
	c.String(http.StatusOK, rss)
}

func (s *EntryService) HandleJSONFeed(c *gin.Context) {
	feed, err := s.GetJSONFeed()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "application/feed+json")
	c.String(http.StatusOK, feed)
}

func RegisterRoutes(rg *gin.RouterGroup, entrySvc *EntryService, mw gin.HandlerFunc) {
	entries := rg.Group("/entries").Use(mw)
	{
		entries.GET("", entrySvc.HandleList)
		entries.POST("", entrySvc.HandleCreate)
		entries.GET("/:id", entrySvc.HandleGet)
		entries.PUT("/:id", entrySvc.HandleUpdate)
		entries.DELETE("/:id", entrySvc.HandleDelete)
		entries.GET("/tags", entrySvc.HandleListTags)
		entries.POST("/tags", entrySvc.HandleCreateTag)
		entries.DELETE("/tags/:id", entrySvc.HandleDeleteTag)
		entries.POST("/:id/attachments", entrySvc.HandleUploadAttachment)
	}
}
