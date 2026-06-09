package entries

import (
	"time"

	"github.com/google/uuid"
	"github.com/kunal/life-log/backend/internal/users"
	"gorm.io/gorm"
)

const (
	EntryTypeMood         = "mood"
	EntryTypeActivity     = "activity"
	EntryTypeJournal      = "journal"
	EntryTypeHabitRecord  = "habit_record"
	EntryTypeGoalProgress = "goal_progress"
	EntryTypeSleep        = "sleep"
	EntryTypeExercise     = "exercise"
	EntryTypeWeight       = "weight"
	EntryTypeMedication   = "medication"
	EntryTypeCustom       = "custom"
)

const (
	VisibilityPrivate  = "private"
	VisibilityPublic   = "public"
	VisibilityUnlisted = "unlisted"
)

type Tag struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID     `gorm:"type:uuid;index"`
	Name      string         `gorm:"type:varchar(255);not null;index"`
	Color     string         `gorm:"type:varchar(7);default:'#3B82F6'"`
	Icon      string         `gorm:"type:varchar(100)"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`

	User *users.User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

func (t *Tag) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

type EntryTag struct {
	EntryID uuid.UUID `gorm:"type:uuid;primaryKey"`
	TagID   uuid.UUID `gorm:"type:uuid;primaryKey"`

	Entry *Entry `gorm:"foreignKey:EntryID;constraint:OnDelete:CASCADE"`
	Tag   *Tag   `gorm:"foreignKey:TagID;constraint:OnDelete:CASCADE"`
}

type Attachment struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	EntryID   *uuid.UUID `gorm:"type:uuid;index"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	FileName  string     `gorm:"type:varchar(255);not null"`
	FilePath  string     `gorm:"type:varchar(512);not null"`
	FileSize  int64      `gorm:"not null"`
	MimeType  string     `gorm:"type:varchar(127);not null"`
	IsImage   bool       `gorm:"default:false"`
	Width     *int
	Height    *int
	CreatedAt time.Time  `gorm:"autoCreateTime;index"`

	Entry *Entry      `gorm:"foreignKey:EntryID;constraint:OnDelete:SET NULL"`
	User  *users.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (a *Attachment) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

type Entry struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null;index;index:idx_user_entry_date,priority:1;index:idx_user_entry_type,priority:1"`
	EntryType       string         `gorm:"type:varchar(30);not null;index:idx_user_entry_type,priority:2"`
	Title           *string        `gorm:"type:varchar(255)"`
	Description     *string        `gorm:"type:text"`
	MarkdownContent *string        `gorm:"type:text"`
	MoodScore       *int           `gorm:"type:smallint"`
	Activities      *string        `gorm:"type:text"`
	Visibility      string         `gorm:"type:varchar(20);default:'private';not null;index"`
	IsEdited        bool           `gorm:"default:false"`
	EditHistory     *string        `gorm:"type:json"`
	Metadata        *string        `gorm:"type:json"`
	Source          string         `gorm:"type:varchar(20);default:'web';not null"`
	EntryDate       time.Time      `gorm:"not null;index;index:idx_user_entry_date,priority:2"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	User        users.User   `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	Tags        []Tag        `gorm:"many2many:entry_tags;foreignKey:ID;joinForeignKey:EntryID;References:ID;joinReferences:TagID"`
	Attachments []Attachment `gorm:"foreignKey:EntryID;constraint:OnDelete:CASCADE"`
}

func (e *Entry) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
