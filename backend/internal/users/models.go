package users

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
	RoleViewer Role = "viewer"
)

type NotificationType string

const (
	NotifHabitReminder   NotificationType = "habit_reminder"
	NotifGoalReminder    NotificationType = "goal_reminder"
	NotifDailyReminder   NotificationType = "daily_reminder"
	NotifWeeklySummary   NotificationType = "weekly_summary"
	NotifMonthlySummary  NotificationType = "monthly_summary"
	NotifSystemAlert     NotificationType = "system_alert"
)

type User struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey"`
	Email          string         `gorm:"type:varchar(255);uniqueIndex;not null"`
	Username       string         `gorm:"type:varchar(100);uniqueIndex;not null"`
	PasswordHash   string         `gorm:"type:varchar(255);not null"`
	DisplayName    string         `gorm:"type:varchar(255)"`
	Role           Role           `gorm:"type:varchar(20);default:'user';not null"`
	Bio            string         `gorm:"type:text"`
	AvatarURL      string         `gorm:"type:varchar(512)"`
	EmailVerifiedAt *time.Time    `gorm:"index"`
	TOTPSecret     string         `gorm:"type:varchar(255)"`
	TOTPEnabled    bool           `gorm:"default:false"`
	TOTPVerifiedAt *time.Time
	PasskeyEnabled bool           `gorm:"default:false"`
	LastLoginAt    *time.Time
	LastIP         string         `gorm:"type:varchar(45)"`
	LastUserAgent  string         `gorm:"type:varchar(512)"`
	CreatedAt      time.Time      `gorm:"autoCreateTime"`
	UpdatedAt      time.Time      `gorm:"autoUpdateTime"`
	DeletedAt      gorm.DeletedAt `gorm:"index"`

	Sessions     []Session     `gorm:"foreignKey:UserID"`
	UserSettings *UserSettings `gorm:"foreignKey:UserID"`
	Notifications []Notification `gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (u *User) BeforeDelete(tx *gorm.DB) error {
	return nil
}

type Session struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID       uuid.UUID  `gorm:"type:uuid;not null;index"`
	Token        string     `gorm:"type:varchar(512);uniqueIndex;not null"`
	RefreshToken string     `gorm:"type:varchar(512);uniqueIndex;not null"`
	DeviceInfo   string     `gorm:"type:varchar(255)"`
	IPAddress    string     `gorm:"type:varchar(45)"`
	UserAgent    string     `gorm:"type:varchar(512)"`
	IsRevoked    bool       `gorm:"default:false;not null"`
	ExpiresAt    time.Time  `gorm:"not null;index"`
	CreatedAt    time.Time  `gorm:"autoCreateTime"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (s *Session) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

type UserSettings struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID               uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null"`
	Theme                string     `gorm:"type:varchar(50);default:'system';not null"`
	Language             string     `gorm:"type:varchar(10);default:'en';not null"`
	Timezone             string     `gorm:"type:varchar(50);default:'UTC';not null"`
	WeekStartDay         int        `gorm:"default:1;not null"`
	DateFormat           string     `gorm:"type:varchar(20);default:'2006-01-02';not null"`
	TimeFormat           string     `gorm:"type:varchar(10);default:'15:04';not null"`
	DarkMode             bool       `gorm:"default:false"`
	NotificationEnabled  bool       `gorm:"default:true"`
	DailyReminderTime    *string    `gorm:"type:varchar(5)"`
	WeeklyDigestEnabled  bool       `gorm:"default:false"`
	MonthlyDigestEnabled bool       `gorm:"default:false"`
	PublicProfile        bool       `gorm:"default:false"`
	ShowOnTimeline       bool       `gorm:"default:true"`
	CreatedAt            time.Time  `gorm:"autoCreateTime"`
	UpdatedAt            time.Time  `gorm:"autoUpdateTime"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (us *UserSettings) BeforeCreate(tx *gorm.DB) error {
	if us.ID == uuid.Nil {
		us.ID = uuid.New()
	}
	return nil
}

type Notification struct {
	ID        uuid.UUID        `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID        `gorm:"type:uuid;not null;index"`
	Type      NotificationType `gorm:"type:varchar(50);not null;index"`
	Title     string           `gorm:"type:varchar(255);not null"`
	Body      string           `gorm:"type:text"`
	Data      *string          `gorm:"type:jsonb"`
	ReadAt    *time.Time       `gorm:"index"`
	IsRead    bool             `gorm:"default:false;not null;index"`
	CreatedAt time.Time        `gorm:"autoCreateTime;index"`

	User User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (n *Notification) BeforeCreate(tx *gorm.DB) error {
	if n.ID == uuid.Nil {
		n.ID = uuid.New()
	}
	return nil
}
