package habits

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HabitService struct {
	db *gorm.DB
}

type HabitFilter struct {
	HabitType  string `form:"habit_type"`
	IsArchived *bool  `form:"is_archived"`
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
}

type HabitHeatmap struct {
	Date   string         `json:"date"`
	Count  int            `json:"count"`
	Habits []HabitSummary `json:"habits"`
}

type HabitSummary struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Completed bool   `json:"completed"`
}

type HabitStats struct {
	TotalHabits      int64        `json:"total_habits"`
	ActiveHabits     int64        `json:"active_habits"`
	TotalCompletions int64        `json:"total_completions"`
	CurrentStreaks   []HabitStreak `json:"current_streaks"`
}

type HabitStreak struct {
	HabitID       string `json:"habit_id"`
	Name          string `json:"name"`
	CurrentStreak int    `json:"current_streak"`
	LongestStreak int    `json:"longest_streak"`
}

func NewService(db *gorm.DB) *HabitService {
	return &HabitService{db: db}
}

func (s *HabitService) Create(habit *Habit) error {
	if habit.StartDate.IsZero() {
		habit.StartDate = time.Now().Truncate(24 * time.Hour)
	}
	return s.db.Create(habit).Error
}

func (s *HabitService) GetByID(id uuid.UUID) (*Habit, error) {
	var habit Habit
	err := s.db.Where("id = ?", id).First(&habit).Error
	if err != nil {
		return nil, err
	}
	return &habit, nil
}

func (s *HabitService) Update(habit *Habit) error {
	return s.db.Model(&Habit{}).Where("id = ? AND user_id = ?", habit.ID, habit.UserID).Updates(map[string]interface{}{
		"name":         habit.Name,
		"description":  habit.Description,
		"habit_type":   habit.HabitType,
		"frequency":    habit.Frequency,
		"unit":         habit.Unit,
		"target_value": habit.TargetValue,
		"color":        habit.Color,
		"icon":         habit.Icon,
		"is_archived":  habit.IsArchived,
		"start_date":   habit.StartDate,
	}).Error
}

func (s *HabitService) Delete(id, userID uuid.UUID) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Habit{}).Error
}

func (s *HabitService) List(userID uuid.UUID, filter HabitFilter) ([]Habit, int64, error) {
	query := s.db.Model(&Habit{}).Where("user_id = ?", userID)

	if filter.HabitType != "" {
		query = query.Where("habit_type = ?", filter.HabitType)
	}
	if filter.IsArchived != nil {
		query = query.Where("is_archived = ?", *filter.IsArchived)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
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

	var habits []Habit
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&habits).Error
	if err != nil {
		return nil, 0, err
	}

	return habits, total, nil
}

func (s *HabitService) LogCompletion(habitID, userID uuid.UUID, value float64, note string) error {
	var habit Habit
	if err := s.db.Where("id = ? AND user_id = ?", habitID, userID).First(&habit).Error; err != nil {
		return err
	}

	var notePtr *string
	if note != "" {
		notePtr = &note
	}

	log := &HabitLog{
		HabitID:     habitID,
		UserID:      userID,
		Value:       value,
		Note:        notePtr,
		CompletedAt: time.Now(),
	}

	if err := s.db.Create(log).Error; err != nil {
		return err
	}

	habit.TotalCompletions++

	if err := s.RecalculateStreak(&habit); err != nil {
		return err
	}

	return s.db.Model(&habit).Updates(map[string]interface{}{
		"total_completions": habit.TotalCompletions,
		"current_streak":    habit.CurrentStreak,
		"longest_streak":    habit.LongestStreak,
	}).Error
}

func (s *HabitService) GetLogs(habitID uuid.UUID, startDate, endDate time.Time) ([]HabitLog, error) {
	var logs []HabitLog
	err := s.db.Where("habit_id = ? AND completed_at BETWEEN ? AND ?", habitID, startDate, endDate).
		Order("completed_at DESC").Find(&logs).Error
	return logs, err
}

func (s *HabitService) GetHeatmapData(userID uuid.UUID, year int) ([]HabitHeatmap, error) {
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(year, 12, 31, 23, 59, 59, 999999999, time.UTC)

	var habits []Habit
	if err := s.db.Where("user_id = ? AND is_archived = ?", userID, false).Find(&habits).Error; err != nil {
		return nil, err
	}

	habitMap := make(map[string]string)
	for _, h := range habits {
		habitMap[h.ID.String()] = h.Name
	}

	var logs []HabitLog
	if err := s.db.Where("user_id = ? AND completed_at BETWEEN ? AND ?", userID, startDate, endDate).
		Find(&logs).Error; err != nil {
		return nil, err
	}

	type dayData struct {
		habitIDs map[string]bool
		count    int
	}
	grouped := make(map[string]*dayData)

	for _, log := range logs {
		dateKey := log.CompletedAt.Format("2006-01-02")
		if _, ok := grouped[dateKey]; !ok {
			grouped[dateKey] = &dayData{
				habitIDs: make(map[string]bool),
			}
		}
		hid := log.HabitID.String()
		if !grouped[dateKey].habitIDs[hid] {
			grouped[dateKey].habitIDs[hid] = true
			grouped[dateKey].count++
		}
	}

	dateKeys := make([]string, 0, len(grouped))
	for k := range grouped {
		dateKeys = append(dateKeys, k)
	}
	sort.Strings(dateKeys)

	result := make([]HabitHeatmap, len(dateKeys))
	for i, dk := range dateKeys {
		dd := grouped[dk]
		summary := make([]HabitSummary, 0, len(dd.habitIDs))
		for id := range dd.habitIDs {
			summary = append(summary, HabitSummary{
				ID:        id,
				Name:      habitMap[id],
				Completed: true,
			})
		}
		result[i] = HabitHeatmap{
			Date:   dk,
			Count:  dd.count,
			Habits: summary,
		}
	}

	return result, nil
}

func (s *HabitService) RecalculateStreak(habit *Habit) error {
	var logs []HabitLog
	if err := s.db.Where("habit_id = ?", habit.ID).Order("completed_at DESC").Find(&logs).Error; err != nil {
		return err
	}

	if len(logs) == 0 {
		habit.CurrentStreak = 0
		habit.LongestStreak = 0
		return nil
	}

	now := time.Now().Truncate(24 * time.Hour)
	today := now
	yesterday := today.AddDate(0, 0, -1)

	currentStreak := 0
	latestDate := logs[0].CompletedAt.Truncate(24 * time.Hour)
	if latestDate.Equal(today) || latestDate.Equal(yesterday) {
		currentStreak = 1
		checkDate := latestDate.AddDate(0, 0, -1)
		for i := 1; i < len(logs); i++ {
			logDate := logs[i].CompletedAt.Truncate(24 * time.Hour)
			if logDate.Equal(checkDate) {
				currentStreak++
				checkDate = checkDate.AddDate(0, 0, -1)
			} else if logDate.Before(checkDate) {
				break
			}
		}
	}

	dateSet := make(map[string]bool)
	var dates []time.Time
	for _, log := range logs {
		dk := log.CompletedAt.Format("2006-01-02")
		if !dateSet[dk] {
			dateSet[dk] = true
			dates = append(dates, log.CompletedAt.Truncate(24*time.Hour))
		}
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	longestStreak := 0
	tempStreak := 0
	for i := 0; i < len(dates); i++ {
		if i == 0 {
			tempStreak = 1
		} else {
			diff := dates[i].Sub(dates[i-1]).Hours() / 24
			if diff == 1 {
				tempStreak++
			} else if diff > 1 {
				if tempStreak > longestStreak {
					longestStreak = tempStreak
				}
				tempStreak = 1
			}
		}
	}
	if tempStreak > longestStreak {
		longestStreak = tempStreak
	}

	habit.CurrentStreak = currentStreak
	habit.LongestStreak = longestStreak
	return nil
}

func (s *HabitService) GetStats(userID uuid.UUID) (*HabitStats, error) {
	var totalHabits int64
	if err := s.db.Model(&Habit{}).Where("user_id = ?", userID).Count(&totalHabits).Error; err != nil {
		return nil, err
	}

	var activeHabits int64
	if err := s.db.Model(&Habit{}).Where("user_id = ? AND is_archived = ?", userID, false).Count(&activeHabits).Error; err != nil {
		return nil, err
	}

	var totalCompletions int64
	if err := s.db.Model(&Habit{}).Where("user_id = ?", userID).
		Select("COALESCE(SUM(total_completions), 0)").Scan(&totalCompletions).Error; err != nil {
		return nil, err
	}

	var habits []Habit
	if err := s.db.Where("user_id = ? AND is_archived = ?", userID, false).Find(&habits).Error; err != nil {
		return nil, err
	}

	streaks := make([]HabitStreak, len(habits))
	for i, h := range habits {
		streaks[i] = HabitStreak{
			HabitID:       h.ID.String(),
			Name:          h.Name,
			CurrentStreak: h.CurrentStreak,
			LongestStreak: h.LongestStreak,
		}
	}

	return &HabitStats{
		TotalHabits:      totalHabits,
		ActiveHabits:     activeHabits,
		TotalCompletions: totalCompletions,
		CurrentStreaks:   streaks,
	}, nil
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

type createHabitRequest struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description"`
	HabitType   string   `json:"habit_type" binding:"required"`
	Frequency   *int     `json:"frequency"`
	Unit        *string  `json:"unit"`
	TargetValue *float64 `json:"target_value"`
	Color       *string  `json:"color"`
	Icon        *string  `json:"icon"`
	StartDate   string   `json:"start_date"`
}

type logCompletionRequest struct {
	Value float64 `json:"value"`
	Note  string  `json:"note"`
}

func (s *HabitService) HandleCreate(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req createHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	habit := &Habit{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		HabitType:   req.HabitType,
		Unit:        req.Unit,
		TargetValue: req.TargetValue,
		Icon:        req.Icon,
	}

	if req.Frequency != nil {
		habit.Frequency = *req.Frequency
	}
	if req.Color != nil {
		habit.Color = *req.Color
	}
	if req.StartDate != "" {
		parsed, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected YYYY-MM-DD"})
			return
		}
		habit.StartDate = parsed
	}

	if err := s.Create(habit); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, habit)
}

func (s *HabitService) HandleGet(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	habit, err := s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
		return
	}

	if habit.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, habit)
}

func (s *HabitService) HandleUpdate(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	existing, err := s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
		return
	}
	if existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req createHabitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.HabitType != "" {
		existing.HabitType = req.HabitType
	}
	existing.Description = req.Description
	existing.Unit = req.Unit
	existing.TargetValue = req.TargetValue
	existing.Icon = req.Icon
	if req.Frequency != nil {
		existing.Frequency = *req.Frequency
	}
	if req.Color != nil {
		existing.Color = *req.Color
	}
	if req.StartDate != "" {
		parsed, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format"})
			return
		}
		existing.StartDate = parsed
	}

	if err := s.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, existing)
}

func (s *HabitService) HandleDelete(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	if err := s.Delete(id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "habit deleted"})
}

func (s *HabitService) HandleList(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var filter HabitFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if c.Query("is_archived") != "" {
		val := c.Query("is_archived") == "true"
		filter.IsArchived = &val
	}

	habits, total, err := s.List(userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  habits,
		"total": total,
	})
}

func (s *HabitService) HandleLogCompletion(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	var req logCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.LogCompletion(habitID, userID, req.Value, req.Note); err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "completion logged"})
}

func (s *HabitService) HandleGetLogs(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	habitID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid habit ID"})
		return
	}

	habit, err := s.GetByID(habitID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "habit not found"})
		return
	}
	if habit.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	startDateStr := c.DefaultQuery("start_date", time.Now().AddDate(0, -1, 0).Format("2006-01-02"))
	endDateStr := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected YYYY-MM-DD"})
		return
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected YYYY-MM-DD"})
		return
	}
	endDate = endDate.Add(24*time.Hour - time.Nanosecond)

	logs, err := s.GetLogs(habitID, startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

func (s *HabitService) HandleGetHeatmap(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	yearStr := c.DefaultQuery("year", strconv.Itoa(time.Now().Year()))
	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid year"})
		return
	}

	data, err := s.GetHeatmapData(userID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}

func (s *HabitService) HandleGetStats(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	stats, err := s.GetStats(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func RegisterRoutes(rg *gin.RouterGroup, habitSvc *HabitService, mw gin.HandlerFunc) {
	habits := rg.Group("/habits").Use(mw)
	{
		habits.GET("", habitSvc.HandleList)
		habits.POST("", habitSvc.HandleCreate)
		habits.GET("/stats", habitSvc.HandleGetStats)
		habits.GET("/heatmap", habitSvc.HandleGetHeatmap)
		habits.GET("/:id", habitSvc.HandleGet)
		habits.PUT("/:id", habitSvc.HandleUpdate)
		habits.DELETE("/:id", habitSvc.HandleDelete)
		habits.POST("/:id/log", habitSvc.HandleLogCompletion)
		habits.GET("/:id/logs", habitSvc.HandleGetLogs)
	}
}
