package services

import (
	"math/rand"
	"telegram-bot/internal/database"
	"telegram-bot/internal/models"
	"time"
)

type SupportService struct {
	db   *database.Database
	rand *rand.Rand
}

func NewSupportService(db *database.Database) *SupportService {
	return &SupportService{
		db:   db,
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *SupportService) GetRandomSupport() *models.SupportStaff {
	var staff []models.SupportStaff
	result := s.db.DB.Where("is_active = ?", true).Find(&staff)
	if result.Error != nil || len(staff) == 0 {
		return nil
	}

	index := s.rand.Intn(len(staff))
	return &staff[index]
}

func (s *SupportService) CreateSupportStaff(staff *models.SupportStaff) error {
	return s.db.DB.Create(staff).Error
}

func (s *SupportService) GetAllSupportStaff() ([]models.SupportStaff, error) {
	var staff []models.SupportStaff
	err := s.db.DB.Where("is_active = ?", true).Find(&staff).Error
	return staff, err
}

func (s *SupportService) InitializeTestData() error {
	// Check if support staff already exists
	var count int64
	s.db.DB.Model(&models.SupportStaff{}).Count(&count)
	if count > 0 {
		return nil // Data already exists
	}

	// Add test support staff
	testStaff := []models.SupportStaff{
		{
			Name:     "خانم فاطمه تقی زاده",
			Username: "@rapexam",
			PhotoURL: "https://d1uuxsymbea74i.cloudfront.net/images/cms/1_6_passport_photo_young_female_9061ba5533.jpg",
			IsActive: true,
		},
		{
			Name:     "خانم بهار قربانی",
			Username: "@rapexam",
			PhotoURL: "https://d1uuxsymbea74i.cloudfront.net/images/cms/1_6_passport_photo_young_female_9061ba5533.jpg",
			IsActive: true,
		},
		{
			Name:     "خانم راضیه مزینانی",
			Username: "@rapexam",
			PhotoURL: "https://d1uuxsymbea74i.cloudfront.net/images/cms/1_6_passport_photo_young_female_9061ba5533.jpg",
			IsActive: true,
		},
	}

	for _, staff := range testStaff {
		if err := s.CreateSupportStaff(&staff); err != nil {
			return err
		}
	}

	return nil
}
