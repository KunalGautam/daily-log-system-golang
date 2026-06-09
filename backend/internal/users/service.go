package users

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetByID(id uuid.UUID) (*User, error) {
	var user User
	err := s.db.Preload("UserSettings").First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetByEmail(email string) (*User, error) {
	var user User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) GetByUsername(username string) (*User, error) {
	var user User
	err := s.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserService) Create(user *User) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		settings := UserSettings{UserID: user.ID}
		if err := tx.Create(&settings).Error; err != nil {
			return err
		}
		user.UserSettings = &settings
		return nil
	})
}

func (s *UserService) Update(user *User) error {
	return s.db.Save(user).Error
}

func (s *UserService) Delete(id uuid.UUID) error {
	return s.db.Delete(&User{}, "id = ?", id).Error
}

func (s *UserService) UpdateLastLogin(id uuid.UUID, ip, userAgent string) error {
	return s.db.Model(&User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_login_at":  time.Now(),
		"last_ip":        ip,
		"last_user_agent": userAgent,
	}).Error
}

func (s *UserService) GetSettings(userID uuid.UUID) (*UserSettings, error) {
	var settings UserSettings
	err := s.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	return &settings, nil
}

func (s *UserService) UpdateSettings(userID uuid.UUID, settings *UserSettings) error {
	return s.db.Where("user_id = ?", userID).Updates(settings).Error
}

func (s *UserService) ListNotifications(userID uuid.UUID, limit, offset int) ([]Notification, int64, error) {
	var total int64
	if err := s.db.Model(&Notification{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var notifications []Notification
	err := s.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&notifications).Error
	if err != nil {
		return nil, 0, err
	}
	return notifications, total, nil
}

func (s *UserService) CreateNotification(userID uuid.UUID, nType, title, body, data string) error {
	notification := Notification{
		UserID: userID,
		Type:   NotificationType(nType),
		Title:  title,
		Body:   body,
		Data:   &data,
	}
	return s.db.Create(&notification).Error
}

func (s *UserService) MarkNotificationRead(id, userID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (s *UserService) MarkAllNotificationsRead(userID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Updates(map[string]interface{}{
			"is_read": true,
			"read_at": &now,
		}).Error
}

func (s *UserService) CountUnreadNotifications(userID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Count(&count).Error
	return count, err
}

func (s *UserService) AdminListUsers(page, pageSize int) ([]User, int64, error) {
	var total int64
	if err := s.db.Model(&User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * pageSize
	var users []User
	err := s.db.Preload("UserSettings").
		Order("created_at DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (s *UserService) AdminDeleteUser(id uuid.UUID) error {
	return s.db.Unscoped().Delete(&User{}, "id = ?", id).Error
}

func (s *UserService) AdminUpdateUserRole(id uuid.UUID, role string) error {
	return s.db.Model(&User{}).Where("id = ?", id).Update("role", Role(role)).Error
}

func (s *UserService) HandleGetProfile(c *gin.Context) {
	userIDStr := c.MustGet("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	user, err := s.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (s *UserService) HandleUpdateProfile(c *gin.Context) {
	userIDStr := c.MustGet("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	user, err := s.GetByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	var input struct {
		DisplayName string `json:"display_name"`
		Bio         string `json:"bio"`
		AvatarURL   string `json:"avatar_url"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user.DisplayName = input.DisplayName
	user.Bio = input.Bio
	user.AvatarURL = input.AvatarURL
	if err := s.Update(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update profile"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (s *UserService) HandleGetSettings(c *gin.Context) {
	userIDStr := c.MustGet("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	settings, err := s.GetSettings(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "settings not found"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

func (s *UserService) HandleUpdateSettings(c *gin.Context) {
	userIDStr := c.MustGet("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	var settings UserSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.UpdateSettings(userID, &settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update settings"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "settings updated"})
}

func (s *UserService) HandleListNotifications(c *gin.Context) {
	userIDStr := c.MustGet("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit
	notifications, total, err := s.ListNotifications(userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"notifications": notifications,
		"total":         total,
		"page":          page,
		"limit":         limit,
	})
}

func (s *UserService) HandleMarkNotificationRead(c *gin.Context) {
	userIDStr := c.MustGet("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	notifIDStr := c.Param("id")
	notifID, err := uuid.Parse(notifIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification ID"})
		return
	}
	if err := s.MarkNotificationRead(notifID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark notification as read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

func (s *UserService) HandleMarkAllNotificationsRead(c *gin.Context) {
	userIDStr := c.MustGet("userID").(string)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	if err := s.MarkAllNotificationsRead(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to mark all notifications as read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "all notifications marked as read"})
}

func (s *UserService) HandleAdminListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	users, total, err := s.AdminListUsers(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"users":     users,
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	})
}

func (s *UserService) HandleAdminDeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	if err := s.AdminDeleteUser(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user deleted"})
}

func (s *UserService) HandleAdminUpdateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	var input struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := s.AdminUpdateUserRole(id, input.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update user role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "user role updated"})
}

func RegisterRoutes(rg *gin.RouterGroup, userSvc *UserService, mw gin.HandlerFunc) {
	protected := rg.Group("")
	protected.Use(mw)
	{
		protected.GET("/users/me", userSvc.HandleGetProfile)
		protected.PUT("/users/me", userSvc.HandleUpdateProfile)
		protected.GET("/users/me/settings", userSvc.HandleGetSettings)
		protected.PUT("/users/me/settings", userSvc.HandleUpdateSettings)
		protected.GET("/users/me/notifications", userSvc.HandleListNotifications)
		protected.PUT("/users/me/notifications/read", userSvc.HandleMarkAllNotificationsRead)
		protected.PUT("/users/me/notifications/:id/read", userSvc.HandleMarkNotificationRead)
	}
}
