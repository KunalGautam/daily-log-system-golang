package notifications

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kunal/life-log/backend/configs"
	"github.com/kunal/life-log/backend/internal/users"
	"gorm.io/gorm"
)

type NotificationService struct {
	httpClient *http.Client
	cfg        *configs.NtfyConfig
	db         *gorm.DB
}

func NewService(db *gorm.DB, cfg *configs.NtfyConfig) *NotificationService {
	return &NotificationService{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		cfg:        cfg,
		db:         db,
	}
}

func (s *NotificationService) Send(title, message string, priority int, tags []string) error {
	if !s.cfg.Enabled {
		return nil
	}

	url := s.cfg.URL + "/" + s.cfg.Topic

	body := bytes.NewBufferString(message)

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Title", title)
	req.Header.Set("Priority", fmt.Sprintf("%d", priority))
	req.Header.Set("Content-Type", "text/plain")

	if len(tags) > 0 {
		for _, tag := range tags {
			req.Header.Add("Tags", tag)
		}
	}

	if s.cfg.Token != "" {
		req.Header.Set("Authorization", "Bearer "+s.cfg.Token)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("ntfy returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *NotificationService) SendToUser(userID uuid.UUID, title, message string, priority int) error {
	notif := users.Notification{
		UserID: userID,
		Type:   users.NotifSystemAlert,
		Title:  title,
		Body:   message,
	}

	if err := s.db.Create(&notif).Error; err != nil {
		return fmt.Errorf("failed to create notification record: %w", err)
	}

	if !s.cfg.Enabled {
		return nil
	}

	return s.Send(title, message, priority, nil)
}

func (s *NotificationService) StartDailyReminder() {
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now()
			currentTime := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

			var settings []users.UserSettings
			if err := s.db.Where("daily_reminder_time = ? AND notification_enabled = ?", currentTime, true).
				Find(&settings).Error; err != nil {
				log.Printf("failed to query users for daily reminder: %v", err)
				continue
			}

			for _, setting := range settings {
				title := "Daily Reminder"
				msg := "Time to log your day! Don't forget to record your entries."

				if err := s.SendToUser(setting.UserID, title, msg, 3); err != nil {
					log.Printf("failed to send daily reminder to user %s: %v", setting.UserID, err)
				}
			}
		}
	}()
}

func (s *NotificationService) SendWeeklySummary(userID uuid.UUID) error {
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	endOfWeek := startOfWeek.AddDate(0, 0, 6)

	title := "Weekly Summary"
	body := fmt.Sprintf("Summary for %s to %s", startOfWeek.Format("Jan 2"), endOfWeek.Format("Jan 2, 2006"))

	if err := s.db.Create(&users.Notification{
		UserID: userID,
		Type:   users.NotifWeeklySummary,
		Title:  title,
		Body:   body,
	}).Error; err != nil {
		return fmt.Errorf("failed to create weekly summary notification: %w", err)
	}

	return s.Send(title, body, 2, []string{"chart_with_upwards_trend"})
}

func (s *NotificationService) SendMonthlySummary(userID uuid.UUID) error {
	now := time.Now()
	month := now.Month()
	year := now.Year()

	title := "Monthly Summary"
	body := fmt.Sprintf("Summary for %s %d", month.String(), year)

	if err := s.db.Create(&users.Notification{
		UserID: userID,
		Type:   users.NotifMonthlySummary,
		Title:  title,
		Body:   body,
	}).Error; err != nil {
		return fmt.Errorf("failed to create monthly summary notification: %w", err)
	}

	return s.Send(title, body, 2, []string{"bar_chart"})
}

func (s *NotificationService) HandleSend(c *gin.Context) {
	var req struct {
		Title    string   `json:"title"`
		Message  string   `json:"message"`
		Priority int      `json:"priority"`
		Tags     []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Title == "" || req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "title and message are required"})
		return
	}

	if req.Priority < 1 || req.Priority > 5 {
		req.Priority = 3
	}

	if err := s.Send(req.Title, req.Message, req.Priority, req.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification sent"})
}

func RegisterRoutes(rg *gin.RouterGroup, notifSvc *NotificationService, mw gin.HandlerFunc) {
	protected := rg.Group("/notifications")
	protected.Use(mw)
	{
		protected.POST("/send", notifSvc.HandleSend)
	}
}
