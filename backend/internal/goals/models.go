package goals

import (
	"time"

	"github.com/google/uuid"
	"github.com/kunal/life-log/backend/internal/users"
	"gorm.io/gorm"
)

type Goal struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID        uuid.UUID      `gorm:"type:uuid;not null;index"`
	Title         string         `gorm:"type:varchar(255);not null"`
	Description   string         `gorm:"type:text"`
	GoalType      string         `gorm:"type:varchar(20);not null"`
	TargetValue   float64        `gorm:"not null"`
	Unit          *string        `gorm:"type:varchar(50)"`
	CurrentValue  float64        `gorm:"default:0"`
	StartDate     time.Time      `gorm:"not null"`
	EndDate       *time.Time
	IsCompleted   bool           `gorm:"default:false"`
	CompletedAt   *time.Time
	Color         string         `gorm:"type:varchar(7);default:'#8B5CF6'"`
	Icon          *string        `gorm:"type:varchar(100)"`
	IsArchived    bool           `gorm:"default:false"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`

	User       users.User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Progress   []GoalProgress  `gorm:"foreignKey:GoalID;constraint:OnDelete:CASCADE"`
	Milestones []Milestone     `gorm:"foreignKey:GoalID;constraint:OnDelete:CASCADE"`
}

func (g *Goal) BeforeCreate(tx *gorm.DB) error {
	if g.ID == uuid.Nil {
		g.ID = uuid.New()
	}
	return nil
}

type GoalProgress struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey"`
	GoalID      uuid.UUID `gorm:"type:uuid;not null;index"`
	Value       float64   `gorm:"not null"`
	Note        *string   `gorm:"type:text"`
	RecordedAt  time.Time `gorm:"not null;index"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

func (gp *GoalProgress) BeforeCreate(tx *gorm.DB) error {
	if gp.ID == uuid.Nil {
		gp.ID = uuid.New()
	}
	return nil
}

type Milestone struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey"`
	GoalID       uuid.UUID      `gorm:"type:uuid;not null;index"`
	Title        string         `gorm:"type:varchar(255);not null"`
	Description  *string        `gorm:"type:text"`
	TargetValue  float64        `gorm:"not null"`
	IsReached    bool           `gorm:"default:false"`
	ReachedAt    *time.Time
	CreatedAt    time.Time      `gorm:"autoCreateTime"`
	UpdatedAt    time.Time      `gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (m *Milestone) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}
