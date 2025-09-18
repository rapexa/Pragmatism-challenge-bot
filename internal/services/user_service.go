package services

import (
	"errors"
	"math/rand"
	"sync"
	"telegram-bot/internal/database"
	"telegram-bot/internal/models"
	"time"

	"gorm.io/gorm"
)

type UserService struct {
	db             *database.Database
	registrations  map[int64]*models.UserRegistrationState
	registrationMu sync.RWMutex
	rand           *rand.Rand
}

func NewUserService(db *database.Database) *UserService {
	return &UserService{
		db:            db,
		registrations: make(map[int64]*models.UserRegistrationState),
		rand:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *UserService) GetUser(telegramID int64) (*models.User, error) {
	var user models.User
	err := s.db.DB.Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (s *UserService) CreateUser(user *models.User) error {
	return s.db.DB.Create(user).Error
}

func (s *UserService) StartRegistration(telegramID int64) {
	s.registrationMu.Lock()
	defer s.registrationMu.Unlock()

	s.registrations[telegramID] = &models.UserRegistrationState{
		TelegramID: telegramID,
		Step:       "waiting_name",
	}
}

func (s *UserService) GetRegistrationState(telegramID int64) *models.UserRegistrationState {
	s.registrationMu.RLock()
	defer s.registrationMu.RUnlock()

	return s.registrations[telegramID]
}

func (s *UserService) UpdateRegistrationState(telegramID int64, step string, data map[string]string) {
	s.registrationMu.Lock()
	defer s.registrationMu.Unlock()

	state, exists := s.registrations[telegramID]
	if !exists {
		return
	}

	state.Step = step

	if firstName, ok := data["first_name"]; ok {
		state.FirstName = firstName
	}
	if lastName, ok := data["last_name"]; ok {
		state.LastName = lastName
	}
	if phoneNumber, ok := data["phone_number"]; ok {
		state.PhoneNumber = phoneNumber
	}
}

func (s *UserService) CompleteRegistration(telegramID int64, phoneNumber, job, username string) error {
	s.registrationMu.Lock()
	defer s.registrationMu.Unlock()

	state, exists := s.registrations[telegramID]
	if !exists {
		return errors.New("registration state not found")
	}

	// Get a random support staff for this user (will be fixed for this user)
	var supportStaff []models.SupportStaff
	result := s.db.DB.Where("is_active = ?", true).Find(&supportStaff)
	if result.Error != nil || len(supportStaff) == 0 {
		return errors.New("no active support staff available")
	}

	// Select support staff randomly for initial assignment (but will be fixed for this user)
	supportIndex := s.rand.Intn(len(supportStaff))
	selectedSupport := supportStaff[supportIndex]

	user := &models.User{
		TelegramID:     telegramID,
		FirstName:      state.FirstName,
		LastName:       state.LastName,
		PhoneNumber:    phoneNumber,
		Job:            job,
		Username:       username,
		SupportStaffID: selectedSupport.ID,
		IsActive:       true,
	}

	err := s.CreateUser(user)
	if err != nil {
		return err
	}

	// Clean up registration state
	delete(s.registrations, telegramID)

	return nil
}

func (s *UserService) GetUserWithSupport(telegramID int64) (*models.User, *models.SupportStaff, error) {
	var user models.User
	err := s.db.DB.Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var support models.SupportStaff
	err = s.db.DB.First(&support, user.SupportStaffID).Error
	if err != nil {
		return &user, nil, err
	}

	return &user, &support, nil
}

func (s *UserService) CancelRegistration(telegramID int64) {
	s.registrationMu.Lock()
	defer s.registrationMu.Unlock()

	delete(s.registrations, telegramID)
}
