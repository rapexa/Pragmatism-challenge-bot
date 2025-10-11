package services

import (
	"fmt"
	"log"
	"telegram-bot/internal/database"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type DelayedMessageService struct {
	db            *database.Database
	bot           *tgbotapi.BotAPI
	avanakService *AvanakService
}

func NewDelayedMessageService(db *database.Database, bot *tgbotapi.BotAPI, avanakService *AvanakService) *DelayedMessageService {
	return &DelayedMessageService{
		db:            db,
		bot:           bot,
		avanakService: avanakService,
	}
}

// ScheduleDelayedMessage schedules a message to be sent after a delay
func (s *DelayedMessageService) ScheduleDelayedMessage(userID int64, message string, delay time.Duration) {
	go func() {
		// Wait for the specified delay
		time.Sleep(delay)

		// Send the message
		msg := tgbotapi.NewMessage(userID, message)
		_, err := s.bot.Send(msg)
		if err != nil {
			log.Printf("Error sending delayed message to user %d: %v", userID, err)
		} else {
			log.Printf("Delayed message sent successfully to user %d", userID)
		}
	}()
}

// ScheduleGroupLinkFollowUp schedules the follow-up message after group link
func (s *DelayedMessageService) ScheduleGroupLinkFollowUp(userID int64) {
	followUpMessage := `چــیــشــد گلادیاتور دکمه شیشه ای رو کلیک کردی برا ورود به ربات⁉️`

	// Schedule message to be sent after 1 minute
	s.ScheduleDelayedMessage(userID, followUpMessage, 1*time.Minute)
}

// ScheduleWelcomeFollowUp schedules a welcome follow-up message
func (s *DelayedMessageService) ScheduleWelcomeFollowUp(userID int64, userName string) {
	followUpMessage := fmt.Sprintf(`سلام %s عزیز! 👋

چــیــشــد گلادیاتور دکمه شیشه ای رو کلیک کردی برا ورود به ربات⁉️

امیدوارم از چالش 3 روزه عملگرایی لذت ببری! 🌟`, userName)

	// Schedule message to be sent after 1 minute
	s.ScheduleDelayedMessage(userID, followUpMessage, 1*time.Minute)
}

// ScheduleCustomDelayedMessage schedules a custom message with custom delay
func (s *DelayedMessageService) ScheduleCustomDelayedMessage(userID int64, message string, delayMinutes int) {
	delay := time.Duration(delayMinutes) * time.Minute
	s.ScheduleDelayedMessage(userID, message, delay)
}

// ScheduleVoiceCallAfterRegistration schedules a voice call after user registration
func (s *DelayedMessageService) ScheduleVoiceCallAfterRegistration(userID int64, phoneNumber string) {
	if !s.avanakService.IsEnabled() {
		log.Printf("Avanak service is disabled, skipping voice call for user %d", userID)
		return
	}

	go func() {
		// Wait for 1 minute after registration
		time.Sleep(1 * time.Minute)

		// Send voice call
		err := s.avanakService.SendVoiceCall(phoneNumber)
		if err != nil {
			log.Printf("Error sending voice call to user %d (%s): %v", userID, phoneNumber, err)
		} else {
			log.Printf("Voice call sent successfully to user %d (%s)", userID, phoneNumber)
		}
	}()
}
