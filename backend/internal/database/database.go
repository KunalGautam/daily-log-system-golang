package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kunal/life-log/backend/configs"
	"github.com/kunal/life-log/backend/internal/auth"
	"github.com/kunal/life-log/backend/internal/entries"
	"github.com/kunal/life-log/backend/internal/goals"
	"github.com/kunal/life-log/backend/internal/habits"
	"github.com/kunal/life-log/backend/internal/users"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(cfg *configs.DatabaseConfig) *gorm.DB {
	var dialector gorm.Dialector

	switch cfg.Driver {
	case "postgres", "postgresql":
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
		if cfg.DSN != "" {
			dsn = cfg.DSN
		}
		dialector = postgres.Open(dsn)
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		if cfg.DSN != "" {
			dsn = cfg.DSN
		}
		dialector = mysql.Open(dsn)
	default:
		dbPath := cfg.DSN
		if dbPath == "" {
			dbPath = "life_log.db"
		}
		dir := filepath.Dir(dbPath)
		if dir != "." && dir != "" {
			os.MkdirAll(dir, 0755)
		}
		dialector = sqlite.Open(dbPath)
	}

	newLogger := gormlogger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		gormlogger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  gormlogger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	var err error
	DB, err = gorm.Open(dialector, &gorm.Config{
		Logger:                 newLogger,
		SkipDefaultTransaction: true,
		PrepareStmt:            true,
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("failed to get database instance: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return DB
}

func Migrate() {
	err := DB.AutoMigrate(
		&users.User{},
		&users.Session{},
		&users.UserSettings{},
		&auth.PasskeyCredential{},
		&auth.RecoveryCode{},
		&auth.AuditLog{},
		&entries.Entry{},
		&entries.Tag{},
		&entries.EntryTag{},
		&entries.Attachment{},
		&habits.Habit{},
		&habits.HabitLog{},
		&goals.Goal{},
		&goals.GoalProgress{},
		&goals.Milestone{},
		&users.Notification{},
	)
	if err != nil {
		log.Fatalf("failed to migrate database: %v", err)
	}
}
