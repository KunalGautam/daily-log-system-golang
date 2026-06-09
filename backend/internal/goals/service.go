package goals

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GoalService struct {
	db *gorm.DB
}

type GoalFilter struct {
	GoalType    string `form:"goal_type"`
	IsCompleted *bool  `form:"is_completed"`
	IsArchived  *bool  `form:"is_archived"`
	Page        int    `form:"page"`
	PageSize    int    `form:"page_size"`
}

type GoalStats struct {
	TotalGoals     int64            `json:"total_goals"`
	ActiveGoals    int64            `json:"active_goals"`
	CompletedGoals int64            `json:"completed_goals"`
	ByType         map[string]int64 `json:"by_type"`
}

type createGoalRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	GoalType    string  `json:"goal_type" binding:"required"`
	TargetValue float64 `json:"target_value" binding:"required"`
	Unit        *string `json:"unit"`
	StartDate   string  `json:"start_date"`
	EndDate     string  `json:"end_date"`
	Color       *string `json:"color"`
	Icon        *string `json:"icon"`
}

type addProgressRequest struct {
	Value float64 `json:"value" binding:"required"`
	Note  string  `json:"note"`
}

type createMilestoneRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	TargetValue float64 `json:"target_value" binding:"required"`
}

func NewService(db *gorm.DB) *GoalService {
	return &GoalService{db: db}
}

func (s *GoalService) Create(goal *Goal) error {
	return s.db.Create(goal).Error
}

func (s *GoalService) GetByID(id uuid.UUID) (*Goal, error) {
	var goal Goal
	err := s.db.Where("id = ?", id).First(&goal).Error
	if err != nil {
		return nil, err
	}
	return &goal, nil
}

func (s *GoalService) Update(goal *Goal) error {
	return s.db.Model(&Goal{}).Where("id = ? AND user_id = ?", goal.ID, goal.UserID).Updates(map[string]interface{}{
		"title":         goal.Title,
		"description":   goal.Description,
		"goal_type":     goal.GoalType,
		"target_value":  goal.TargetValue,
		"unit":          goal.Unit,
		"current_value": goal.CurrentValue,
		"start_date":    goal.StartDate,
		"end_date":      goal.EndDate,
		"is_completed":  goal.IsCompleted,
		"completed_at":  goal.CompletedAt,
		"color":         goal.Color,
		"icon":          goal.Icon,
		"is_archived":   goal.IsArchived,
	}).Error
}

func (s *GoalService) Delete(id, userID uuid.UUID) error {
	return s.db.Where("id = ? AND user_id = ?", id, userID).Delete(&Goal{}).Error
}

func (s *GoalService) List(userID uuid.UUID, filter GoalFilter) ([]Goal, int64, error) {
	query := s.db.Model(&Goal{}).Where("user_id = ?", userID)

	if filter.GoalType != "" {
		query = query.Where("goal_type = ?", filter.GoalType)
	}
	if filter.IsCompleted != nil {
		query = query.Where("is_completed = ?", *filter.IsCompleted)
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

	var goals []Goal
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&goals).Error
	if err != nil {
		return nil, 0, err
	}

	return goals, total, nil
}

func (s *GoalService) AddProgress(goalID, userID uuid.UUID, value float64, note string) error {
	var goal Goal
	if err := s.db.Where("id = ? AND user_id = ?", goalID, userID).First(&goal).Error; err != nil {
		return err
	}

	var notePtr *string
	if note != "" {
		notePtr = &note
	}

	progress := &GoalProgress{
		GoalID:     goalID,
		Value:      value,
		Note:       notePtr,
		RecordedAt: time.Now(),
	}

	if err := s.db.Create(progress).Error; err != nil {
		return err
	}

	goal.CurrentValue += value

	updates := map[string]interface{}{
		"current_value": goal.CurrentValue,
	}

	if goal.CurrentValue >= goal.TargetValue && !goal.IsCompleted {
		goal.IsCompleted = true
		now := time.Now()
		goal.CompletedAt = &now
		updates["is_completed"] = true
		updates["completed_at"] = now
	}

	return s.db.Model(&goal).Where("id = ? AND user_id = ?", goalID, userID).Updates(updates).Error
}

func (s *GoalService) GetProgress(goalID uuid.UUID) ([]GoalProgress, error) {
	var progress []GoalProgress
	err := s.db.Where("goal_id = ?", goalID).Order("recorded_at DESC").Find(&progress).Error
	return progress, err
}

func (s *GoalService) CreateMilestone(milestone *Milestone) error {
	return s.db.Create(milestone).Error
}

func (s *GoalService) GetMilestones(goalID uuid.UUID) ([]Milestone, error) {
	var milestones []Milestone
	err := s.db.Where("goal_id = ?", goalID).Order("target_value ASC").Find(&milestones).Error
	return milestones, err
}

func (s *GoalService) UpdateMilestone(milestone *Milestone) error {
	return s.db.Model(&Milestone{}).Where("id = ?", milestone.ID).Updates(map[string]interface{}{
		"title":        milestone.Title,
		"description":  milestone.Description,
		"target_value": milestone.TargetValue,
		"is_reached":   milestone.IsReached,
		"reached_at":   milestone.ReachedAt,
	}).Error
}

func (s *GoalService) DeleteMilestone(id uuid.UUID) error {
	return s.db.Where("id = ?", id).Delete(&Milestone{}).Error
}

func (s *GoalService) GetStats(userID uuid.UUID) (*GoalStats, error) {
	var totalGoals int64
	if err := s.db.Model(&Goal{}).Where("user_id = ?", userID).Count(&totalGoals).Error; err != nil {
		return nil, err
	}

	var activeGoals int64
	if err := s.db.Model(&Goal{}).Where("user_id = ? AND is_archived = ? AND is_completed = ?", userID, false, false).Count(&activeGoals).Error; err != nil {
		return nil, err
	}

	var completedGoals int64
	if err := s.db.Model(&Goal{}).Where("user_id = ? AND is_completed = ?", userID, true).Count(&completedGoals).Error; err != nil {
		return nil, err
	}

	type typeCount struct {
		GoalType string
		Count    int64
	}
	var counts []typeCount
	if err := s.db.Model(&Goal{}).Where("user_id = ?", userID).Select("goal_type, COUNT(*) as count").Group("goal_type").Scan(&counts).Error; err != nil {
		return nil, err
	}

	byType := make(map[string]int64)
	for _, c := range counts {
		byType[c.GoalType] = c.Count
	}

	return &GoalStats{
		TotalGoals:     totalGoals,
		ActiveGoals:    activeGoals,
		CompletedGoals: completedGoals,
		ByType:         byType,
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

func (s *GoalService) HandleCreate(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var req createGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	goal := &Goal{
		UserID:      userID,
		Title:       req.Title,
		Description: req.Description,
		GoalType:    req.GoalType,
		TargetValue: req.TargetValue,
		Unit:        req.Unit,
		Icon:        req.Icon,
	}

	if req.Color != nil {
		goal.Color = *req.Color
	}
	if req.StartDate != "" {
		parsed, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, expected YYYY-MM-DD"})
			return
		}
		goal.StartDate = parsed
	}
	if req.EndDate != "" {
		parsed, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format, expected YYYY-MM-DD"})
			return
		}
		goal.EndDate = &parsed
	}

	if err := s.Create(goal); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, goal)
}

func (s *GoalService) HandleGet(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	goal, err := s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}

	if goal.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	c.JSON(http.StatusOK, goal)
}

func (s *GoalService) HandleUpdate(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	existing, err := s.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	if existing.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req createGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.GoalType != "" {
		existing.GoalType = req.GoalType
	}
	existing.Description = req.Description
	existing.TargetValue = req.TargetValue
	existing.Unit = req.Unit
	existing.Icon = req.Icon
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
	if req.EndDate != "" {
		parsed, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end_date format"})
			return
		}
		existing.EndDate = &parsed
	}

	if err := s.Update(existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, existing)
}

func (s *GoalService) HandleDelete(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	if err := s.Delete(id, userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "goal deleted"})
}

func (s *GoalService) HandleList(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	var filter GoalFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if c.Query("is_completed") != "" {
		val := c.Query("is_completed") == "true"
		filter.IsCompleted = &val
	}
	if c.Query("is_archived") != "" {
		val := c.Query("is_archived") == "true"
		filter.IsArchived = &val
	}

	goals, total, err := s.List(userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  goals,
		"total": total,
	})
}

func (s *GoalService) HandleAddProgress(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	var req addProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := s.AddProgress(goalID, userID, req.Value, req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "progress added"})
}

func (s *GoalService) HandleGetProgress(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	goal, err := s.GetByID(goalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	if goal.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	progress, err := s.GetProgress(goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, progress)
}

func (s *GoalService) HandleCreateMilestone(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	goal, err := s.GetByID(goalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	if goal.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req createMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	milestone := &Milestone{
		GoalID:      goalID,
		Title:       req.Title,
		Description: req.Description,
		TargetValue: req.TargetValue,
	}

	if err := s.CreateMilestone(milestone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, milestone)
}

func (s *GoalService) HandleGetMilestones(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	goalID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid goal ID"})
		return
	}

	goal, err := s.GetByID(goalID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	if goal.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	milestones, err := s.GetMilestones(goalID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, milestones)
}

func (s *GoalService) HandleUpdateMilestone(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid milestone ID"})
		return
	}

	var milestone Milestone
	if err := s.db.Where("id = ?", id).First(&milestone).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "milestone not found"})
		return
	}

	var goal Goal
	if err := s.db.Where("id = ?", milestone.GoalID).First(&goal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	if goal.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	var req createMilestoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title != "" {
		milestone.Title = req.Title
	}
	milestone.Description = req.Description
	milestone.TargetValue = req.TargetValue

	if err := s.UpdateMilestone(&milestone); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, milestone)
}

func (s *GoalService) HandleDeleteMilestone(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid milestone ID"})
		return
	}

	var milestone Milestone
	if err := s.db.Where("id = ?", id).First(&milestone).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "milestone not found"})
		return
	}

	var goal Goal
	if err := s.db.Where("id = ?", milestone.GoalID).First(&goal).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "goal not found"})
		return
	}
	if goal.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
		return
	}

	if err := s.DeleteMilestone(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "milestone deleted"})
}

func (s *GoalService) HandleGetStats(c *gin.Context) {
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

func RegisterRoutes(rg *gin.RouterGroup, goalSvc *GoalService, mw gin.HandlerFunc) {
	goals := rg.Group("/goals").Use(mw)
	{
		goals.GET("", goalSvc.HandleList)
		goals.POST("", goalSvc.HandleCreate)
		goals.GET("/stats", goalSvc.HandleGetStats)
		goals.GET("/:id", goalSvc.HandleGet)
		goals.PUT("/:id", goalSvc.HandleUpdate)
		goals.DELETE("/:id", goalSvc.HandleDelete)
		goals.POST("/:id/progress", goalSvc.HandleAddProgress)
		goals.GET("/:id/progress", goalSvc.HandleGetProgress)
		goals.GET("/:id/milestones", goalSvc.HandleGetMilestones)
		goals.POST("/:id/milestones", goalSvc.HandleCreateMilestone)
		goals.PUT("/milestones/:id", goalSvc.HandleUpdateMilestone)
		goals.DELETE("/milestones/:id", goalSvc.HandleDeleteMilestone)
	}
}
