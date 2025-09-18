package services

import (
	"fmt"
	"telegram-bot/internal/database"
	"telegram-bot/internal/models"
	"time"

	"github.com/xuri/excelize/v2"
)

type AdminPanelService struct {
	db *database.Database
}

func NewAdminPanelService(db *database.Database) *AdminPanelService {
	return &AdminPanelService{
		db: db,
	}
}

// IsAdmin checks if a user is an admin
func (s *AdminPanelService) IsAdmin(telegramID int64) bool {
	var admin models.Admin
	err := s.db.DB.Where("telegram_id = ? AND is_active = ?", telegramID, true).First(&admin).Error
	return err == nil
}

// CreateAdmin creates a new admin
func (s *AdminPanelService) CreateAdmin(telegramID int64, firstName, username string) error {
	admin := &models.Admin{
		TelegramID: telegramID,
		FirstName:  firstName,
		Username:   username,
		IsActive:   true,
	}
	return s.db.DB.Create(admin).Error
}

// InitializeDefaultAdmin creates the default admin if not exists
func (s *AdminPanelService) InitializeDefaultAdmin() error {
	// Initialize first admin (RAPEXA)
	var count1 int64
	s.db.DB.Model(&models.Admin{}).Where("telegram_id = ?", 76599340).Count(&count1)
	if count1 == 0 {
		err := s.CreateAdmin(76599340, "RAPEXA", "@Rapexam")
		if err != nil {
			return err
		}
	}

	// Initialize second admin (حسین مهری)
	var count2 int64
	s.db.DB.Model(&models.Admin{}).Where("telegram_id = ?", 363999066).Count(&count2)
	if count2 == 0 {
		err := s.CreateAdmin(363999066, "حسین مهری", "@hosseinmehri_ir")
		if err != nil {
			return err
		}
	}

	return nil
}

// ExportUsersToExcel exports all users to Excel file
func (s *AdminPanelService) ExportUsersToExcel() (string, error) {
	var users []models.User
	err := s.db.DB.Find(&users).Error
	if err != nil {
		return "", err
	}

	filename := fmt.Sprintf("users_export_%s.xlsx", time.Now().Format("2006-01-02_15-04-05"))

	// Create a new Excel file
	f := excelize.NewFile()
	defer f.Close()

	// Create a new sheet
	sheetName := "Users"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return "", err
	}

	// Set header
	headers := []string{"ID", "Telegram ID", "First Name", "Last Name", "Phone", "Job", "Username/ID", "Support Staff", "Active", "Created At"}
	for i, header := range headers {
		cell := fmt.Sprintf("%c1", 'A'+i)
		f.SetCellValue(sheetName, cell, header)
	}

	// Write data
	for i, user := range users {
		row := i + 2 // Start from row 2 (row 1 is header)

		var supportName string
		if user.SupportStaffID > 0 {
			var support models.SupportStaff
			s.db.DB.First(&support, user.SupportStaffID)
			supportName = support.Name
		}

		// Determine username or telegram ID
		usernameOrID := user.Username
		if usernameOrID == "" {
			usernameOrID = fmt.Sprintf("%d", user.TelegramID)
		}

		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), user.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), user.TelegramID)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), user.FirstName)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), user.LastName)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), user.PhoneNumber)
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), user.Job)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), usernameOrID)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), supportName)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), user.IsActive)
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), user.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// Set active sheet
	f.SetActiveSheet(index)

	// Save file
	if err := f.SaveAs(filename); err != nil {
		return "", err
	}

	return filename, nil
}

// GetUserStats returns user statistics
func (s *AdminPanelService) GetUserStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total users
	var total int64
	s.db.DB.Model(&models.User{}).Count(&total)
	stats["total"] = total

	// Active users
	var active int64
	s.db.DB.Model(&models.User{}).Where("is_active = ?", true).Count(&active)
	stats["active"] = active

	// Today's registrations
	var today int64
	todayDate := time.Now().Format("2006-01-02")
	s.db.DB.Model(&models.User{}).Where("DATE(created_at) = ?", todayDate).Count(&today)
	stats["today"] = today

	// This week's registrations
	var week int64
	weekStart := time.Now().AddDate(0, 0, -7)
	s.db.DB.Model(&models.User{}).Where("created_at >= ?", weekStart).Count(&week)
	stats["week"] = week

	// This month's registrations
	var month int64
	monthStart := time.Now().AddDate(0, -1, 0)
	s.db.DB.Model(&models.User{}).Where("created_at >= ?", monthStart).Count(&month)
	stats["month"] = month

	return stats, nil
}

// GetSupportStaffList returns all support staff
func (s *AdminPanelService) GetSupportStaffList() ([]models.SupportStaff, error) {
	var staff []models.SupportStaff
	err := s.db.DB.Find(&staff).Error
	return staff, err
}

// CreateSupportStaff creates new support staff
func (s *AdminPanelService) CreateSupportStaff(name, username, photoURL string) error {
	staff := &models.SupportStaff{
		Name:     name,
		Username: username,
		PhotoURL: photoURL,
		IsActive: true,
	}
	return s.db.DB.Create(staff).Error
}

// UpdateSupportStaff updates support staff
func (s *AdminPanelService) UpdateSupportStaff(id uint, name, username, photoURL string) error {
	return s.db.DB.Model(&models.SupportStaff{}).Where("id = ?", id).Updates(models.SupportStaff{
		Name:     name,
		Username: username,
		PhotoURL: photoURL,
	}).Error
}

// UpdateSupportStaffField updates a specific field of support staff
func (s *AdminPanelService) UpdateSupportStaffField(id uint, field, value string) error {
	return s.db.DB.Model(&models.SupportStaff{}).Where("id = ?", id).Update(field, value).Error
}

// DeleteSupportStaff soft deletes support staff
func (s *AdminPanelService) DeleteSupportStaff(id uint) error {
	return s.db.DB.Delete(&models.SupportStaff{}, id).Error
}

// ToggleSupportStaffStatus toggles support staff active status
func (s *AdminPanelService) ToggleSupportStaffStatus(id uint) error {
	var staff models.SupportStaff
	if err := s.db.DB.First(&staff, id).Error; err != nil {
		return err
	}

	return s.db.DB.Model(&staff).Update("is_active", !staff.IsActive).Error
}
