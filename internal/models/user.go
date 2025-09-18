package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	TelegramID     int64          `gorm:"uniqueIndex;not null" json:"telegram_id"`
	FirstName      string         `gorm:"not null" json:"first_name"`
	LastName       string         `gorm:"not null" json:"last_name"`
	PhoneNumber    string         `gorm:"not null" json:"phone_number"`
	Job            string         `gorm:"not null" json:"job"`
	Username       string         `json:"username"`
	SupportStaffID uint           `json:"support_staff_id"` // Fixed support staff for this user
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type UserRegistrationState struct {
	TelegramID  int64  `json:"telegram_id"`
	Step        string `json:"step"` // "waiting_name", "waiting_phone", "waiting_job"
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
}
