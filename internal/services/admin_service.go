package services

import (
	"fmt"
	"telegram-bot/internal/database"
	"telegram-bot/internal/models"
)

type AdminService struct {
	db *database.Database
}

func NewAdminService(db *database.Database) *AdminService {
	return &AdminService{
		db: db,
	}
}

// GetUserStats returns statistics about registered users
func (s *AdminService) GetUserStats() (map[string]interface{}, error) {
	var totalUsers int64
	var activeUsers int64
	var todayUsers int64

	// Total users
	err := s.db.DB.Model(&models.User{}).Count(&totalUsers).Error
	if err != nil {
		return nil, err
	}

	// Active users
	err = s.db.DB.Model(&models.User{}).Where("is_active = ?", true).Count(&activeUsers).Error
	if err != nil {
		return nil, err
	}

	// Today's registrations
	err = s.db.DB.Model(&models.User{}).Where("DATE(created_at) = CURDATE()").Count(&todayUsers).Error
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_users":  totalUsers,
		"active_users": activeUsers,
		"today_users":  todayUsers,
	}, nil
}

// GetSupportStats returns statistics about support staff
func (s *AdminService) GetSupportStats() (map[string]interface{}, error) {
	var totalSupport int64
	var activeSupport int64

	// Total support staff
	err := s.db.DB.Model(&models.SupportStaff{}).Count(&totalSupport).Error
	if err != nil {
		return nil, err
	}

	// Active support staff
	err = s.db.DB.Model(&models.SupportStaff{}).Where("is_active = ?", true).Count(&activeSupport).Error
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"total_support":  totalSupport,
		"active_support": activeSupport,
	}, nil
}

// FormatStatsMessage formats statistics into a readable message
func (s *AdminService) FormatStatsMessage() (string, error) {
	userStats, err := s.GetUserStats()
	if err != nil {
		return "", err
	}

	supportStats, err := s.GetSupportStats()
	if err != nil {
		return "", err
	}

	message := fmt.Sprintf(`📊 آمار ربات:

👥 کاربران:
• کل کاربران: %d
• کاربران فعال: %d
• ثبت نام امروز: %d

👨‍💼 پشتیبان‌ها:
• کل پشتیبان‌ها: %d
• پشتیبان‌های فعال: %d`,
		userStats["total_users"],
		userStats["active_users"],
		userStats["today_users"],
		supportStats["total_support"],
		supportStats["active_support"])

	return message, nil
}
