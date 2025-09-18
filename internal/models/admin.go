package models

import (
	"time"

	"gorm.io/gorm"
)

type Admin struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	TelegramID int64          `gorm:"uniqueIndex;not null" json:"telegram_id"`
	FirstName  string         `gorm:"not null" json:"first_name"`
	Username   string         `json:"username"`
	IsActive   bool           `gorm:"default:true" json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
