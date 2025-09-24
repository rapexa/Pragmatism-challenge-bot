package database

import (
	"fmt"
	"log"
	"telegram-bot/internal/config"
	"telegram-bot/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Database struct {
	DB *gorm.DB
}

func New(cfg *config.DatabaseConfig) *Database {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate models
	err = db.AutoMigrate(&models.User{}, &models.SupportStaff{}, &models.Admin{}, &models.BroadcastMessage{}, &models.BroadcastDelivery{})
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("Database connection established successfully")

	return &Database{DB: db}
}

func (d *Database) Close() {
	sqlDB, err := d.DB.DB()
	if err != nil {
		log.Printf("Error getting SQL connection: %v", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}
}
