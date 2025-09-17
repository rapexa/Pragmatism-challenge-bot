package services

import (
	"telegram-bot/internal/database"
	"telegram-bot/internal/models"
)

type SupportAdminService struct {
	db *database.Database
}

func NewSupportAdminService(db *database.Database) *SupportAdminService {
	return &SupportAdminService{
		db: db,
	}
}

// CreateSupportStaff creates a new support staff member
func (s *SupportAdminService) CreateSupportStaff(name, username, photoURL string) (*models.SupportStaff, error) {
	staff := &models.SupportStaff{
		Name:     name,
		Username: username,
		PhotoURL: photoURL,
		IsActive: true,
	}

	err := s.db.DB.Create(staff).Error
	return staff, err
}

// GetSupportStaffByID gets a support staff by ID
func (s *SupportAdminService) GetSupportStaffByID(id uint) (*models.SupportStaff, error) {
	var staff models.SupportStaff
	err := s.db.DB.First(&staff, id).Error
	return &staff, err
}

// UpdateSupportStaff updates support staff information
func (s *SupportAdminService) UpdateSupportStaff(id uint, name, username, photoURL string) error {
	return s.db.DB.Model(&models.SupportStaff{}).Where("id = ?", id).Updates(models.SupportStaff{
		Name:     name,
		Username: username,
		PhotoURL: photoURL,
	}).Error
}

// DeactivateSupportStaff deactivates a support staff member
func (s *SupportAdminService) DeactivateSupportStaff(id uint) error {
	return s.db.DB.Model(&models.SupportStaff{}).Where("id = ?", id).Update("is_active", false).Error
}

// ActivateSupportStaff activates a support staff member
func (s *SupportAdminService) ActivateSupportStaff(id uint) error {
	return s.db.DB.Model(&models.SupportStaff{}).Where("id = ?", id).Update("is_active", true).Error
}

// GetAllSupportStaff gets all support staff (active and inactive)
func (s *SupportAdminService) GetAllSupportStaff() ([]models.SupportStaff, error) {
	var staff []models.SupportStaff
	err := s.db.DB.Find(&staff).Error
	return staff, err
}

// GetActiveSupportStaff gets only active support staff
func (s *SupportAdminService) GetActiveSupportStaff() ([]models.SupportStaff, error) {
	var staff []models.SupportStaff
	err := s.db.DB.Where("is_active = ?", true).Find(&staff).Error
	return staff, err
}
