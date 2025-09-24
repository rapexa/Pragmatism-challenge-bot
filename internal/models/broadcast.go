package models

import (
	"time"

	"gorm.io/gorm"
)

// BroadcastMessage represents a broadcast message sent to all users
type BroadcastMessage struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	AdminID     int64          `gorm:"not null" json:"admin_id"`        // Telegram ID of admin who sent
	ContentType string         `gorm:"not null" json:"content_type"`    // text, photo, video, document, etc.
	Text        string         `json:"text"`                            // Caption or text content
	FileID      string         `json:"file_id"`                         // Telegram file ID
	FileURL     string         `json:"file_url"`                        // Local file URL
	Status      string         `gorm:"default:'pending'" json:"status"` // pending, sent, failed
	SentCount   int            `gorm:"default:0" json:"sent_count"`     // Number of users who received it
	FailedCount int            `gorm:"default:0" json:"failed_count"`   // Number of failed deliveries
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// BroadcastDelivery tracks delivery status for each user
type BroadcastDelivery struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	BroadcastID  uint           `gorm:"not null" json:"broadcast_id"`
	UserID       int64          `gorm:"not null" json:"user_id"` // Telegram ID
	Status       string         `gorm:"not null" json:"status"`  // pending, sent, failed, blocked
	ErrorMessage string         `json:"error_message"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

// BroadcastPreview represents a preview of broadcast message
type BroadcastPreview struct {
	ContentType string `json:"content_type"`
	Text        string `json:"text"`
	FileID      string `json:"file_id"`
	FileURL     string `json:"file_url"`
	HasFile     bool   `json:"has_file"`
}
