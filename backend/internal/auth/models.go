package auth

import (
	"time"

	"github.com/google/uuid"
	"github.com/kunal/life-log/backend/internal/users"
	"gorm.io/gorm"
)

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleUser   Role = "user"
	RoleViewer Role = "viewer"
)

type AuditAction string

const (
	ActionLogin         AuditAction = "login"
	ActionLogout        AuditAction = "logout"
	ActionRegister      AuditAction = "register"
	ActionPasswordReset AuditAction = "password_reset"
	ActionEmailVerify   AuditAction = "email_verify"
	ActionTOTPEnable    AuditAction = "totp_enable"
	ActionTOTPDisable   AuditAction = "totp_disable"
	ActionPasskeyRegister AuditAction = "passkey_register"
	ActionPasskeyLogin    AuditAction = "passkey_login"
	ActionSettingsChange AuditAction = "settings_change"
	ActionEntryCreate   AuditAction = "entry_create"
	ActionEntryUpdate   AuditAction = "entry_update"
	ActionEntryDelete   AuditAction = "entry_delete"
	ActionAdminAction   AuditAction = "admin_action"
)

type PasskeyCredential struct {
	ID              uuid.UUID      `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null;index"`
	CredentialID    string         `gorm:"type:text;unique;not null"`
	PublicKey       string         `gorm:"type:text;not null"`
	AttestationType string         `gorm:"type:text;not null"`
	Transport       string         `gorm:"type:text"`
	AAGUID          string         `gorm:"type:text"`
	SignCount       int            `gorm:"default:0;not null"`
	CloneWarning    bool           `gorm:"default:false;not null"`
	Name            *string        `gorm:"type:text"`
	CreatedAt       time.Time      `gorm:"autoCreateTime"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime"`
	DeletedAt       gorm.DeletedAt `gorm:"index"`

	User users.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (pc *PasskeyCredential) BeforeCreate(tx *gorm.DB) error {
	if pc.ID == uuid.Nil {
		pc.ID = uuid.New()
	}
	return nil
}

type RecoveryCode struct {
	ID        uuid.UUID  `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null;index"`
	Code      string     `gorm:"type:text;unique;not null"`
	UsedAt    *time.Time
	IsUsed    bool       `gorm:"default:false;not null"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`

	User users.User `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
}

func (rc *RecoveryCode) BeforeCreate(tx *gorm.DB) error {
	if rc.ID == uuid.Nil {
		rc.ID = uuid.New()
	}
	return nil
}

type AuditLog struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey"`
	UserID    *uuid.UUID  `gorm:"type:uuid;index"`
	Action    AuditAction `gorm:"type:varchar(50);not null;index"`
	IPAddress string      `gorm:"type:varchar(45)"`
	UserAgent string      `gorm:"type:varchar(512)"`
	Details   string      `gorm:"type:jsonb"`
	CreatedAt time.Time   `gorm:"autoCreateTime;index"`

	User *users.User `gorm:"foreignKey:UserID;constraint:OnDelete:SET NULL"`
}

func (al *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return nil
}
