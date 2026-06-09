package habits

import (
	"time"

	"github.com/google/uuid"
	"github.com/kunal/life-log/backend/internal/users"
	"gorm.io/gorm"
)

type Habit struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null;index"`
	Name            string         `gorm:"type:varchar(255);not null"`
	Description     string         `gorm:"type:text"`
	HabitType       string         `gorm:"type:varchar(20);not null"`
	Frequency       int            `gorm:"default:1;not null"`
	Unit            *string        `gorm:"type:varchar(50)"`
	TargetValue     *float64
	Color           string         `gorm:"type:varchar(7);default:'#10B981'"`
	Icon            *string        `gorm:"type:varchar(100)"`
	IsArchived      bool           `gorm:"default:false"`
	CurrentStreak   int            `gorm:"default:0;not null"`
	LongestStreak   int            `gorm:"default:0;not null"`
	TotalCompletions int           `gorm:"default:0;not null"`
	SuccessRate     float64        `gorm:"default:0;not null"`
	StartDate       time.Time      `gorm:"not null"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	User      users.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	HabitLogs []HabitLog `gorm:"foreignKey:HabitID;constraint:OnDelete:CASCADE"`
}

func (h *Habit) BeforeCreate(tx *gorm.DB) error {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return nil
}

type HabitLog struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	HabitID     uuid.UUID `gorm:"type:uuid;not null;index;index:idx_habit_completed,priority:1;index:idx_user_habit_completed,priority:2"`
	UserID      uuid.UUID `gorm:"type:uuid;not null;index;index:idx_user_habit_completed,priority:1"`
	Value       float64   `gorm:"default:0"`
	Note        *string   `gorm:"type:text"`
	CompletedAt time.Time `gorm:"not null;index;index:idx_habit_completed,priority:2;index:idx_user_habit_completed,priority:3"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

func (hl *HabitLog) BeforeCreate(tx *gorm.DB) error {
	if hl.ID == uuid.Nil {
		hl.ID = uuid.New()
	}
	return nil
}
