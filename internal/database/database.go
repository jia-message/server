package database

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"jia/server/internal/config"
	"jia/server/internal/models"
)

var DB *gorm.DB

func ConnectDB() {
	cfg := config.AppConfig
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBSSLMode)

	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	if cfg.Env == "production" {
		gormConfig.Logger = logger.Default.LogMode(logger.Warn)
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Configure connection pooling
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("Failed to retrieve generic database object: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	log.Println("Database connection established")

	// Ensure pgcrypto extension is installed for gen_random_uuid()
	if err := DB.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
		log.Printf("Warning: failed to ensure pgcrypto extension is active: %v", err)
	}

	// Auto-migrations
	Migrate()
}

func Migrate() {
	err := DB.AutoMigrate(
		&models.User{},
		&models.ServerSetting{},
		&models.InviteCode{},
		&models.DeviceKey{},
		&models.OneTimePrekey{},
		&models.Conversation{},
		&models.ConversationParticipant{},
		&models.Message{},
		&models.Attachment{},
		&models.Reaction{},
		&models.Contact{},
		&models.Session{},
		&models.PushSubscription{},
	)
	if err != nil {
		log.Fatalf("Failed to perform auto-migration: %v", err)
	}

	log.Println("Database migrations applied successfully")
}
