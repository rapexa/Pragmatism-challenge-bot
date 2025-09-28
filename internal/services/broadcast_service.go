package services

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"telegram-bot/internal/database"
	"telegram-bot/internal/models"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BroadcastService struct {
	db  *database.Database
	bot *tgbotapi.BotAPI
}

func NewBroadcastService(db *database.Database, bot *tgbotapi.BotAPI) *BroadcastService {
	return &BroadcastService{
		db:  db,
		bot: bot,
	}
}

// CreateBroadcast creates a new broadcast message
func (s *BroadcastService) CreateBroadcast(adminID int64, contentType, text, fileID, fileURL string) (*models.BroadcastMessage, error) {
	broadcast := &models.BroadcastMessage{
		AdminID:     adminID,
		ContentType: contentType,
		Text:        text,
		FileID:      fileID,
		FileURL:     fileURL,
		Status:      "pending",
	}

	err := s.db.DB.Create(broadcast).Error
	return broadcast, err
}

// GetBroadcastByID gets a broadcast by ID
func (s *BroadcastService) GetBroadcastByID(id uint) (*models.BroadcastMessage, error) {
	var broadcast models.BroadcastMessage
	err := s.db.DB.First(&broadcast, id).Error
	return &broadcast, err
}

// GetBroadcastHistory gets broadcast history with pagination
func (s *BroadcastService) GetBroadcastHistory(limit, offset int) ([]models.BroadcastMessage, error) {
	var broadcasts []models.BroadcastMessage
	err := s.db.DB.Order("created_at DESC").Limit(limit).Offset(offset).Find(&broadcasts).Error
	return broadcasts, err
}

// GetBroadcastStats gets statistics for a broadcast
func (s *BroadcastService) GetBroadcastStats(broadcastID uint) (map[string]int, error) {
	stats := make(map[string]int)

	// Total deliveries
	var total int64
	s.db.DB.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ?", broadcastID).Count(&total)
	stats["total"] = int(total)

	// Sent count
	var sent int64
	s.db.DB.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ? AND status = ?", broadcastID, "sent").Count(&sent)
	stats["sent"] = int(sent)

	// Failed count
	var failed int64
	s.db.DB.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ? AND status = ?", broadcastID, "failed").Count(&failed)
	stats["failed"] = int(failed)

	// Blocked count
	var blocked int64
	s.db.DB.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ? AND status = ?", broadcastID, "blocked").Count(&blocked)
	stats["blocked"] = int(blocked)

	// Pending count
	var pending int64
	s.db.DB.Model(&models.BroadcastDelivery{}).Where("broadcast_id = ? AND status = ?", broadcastID, "pending").Count(&pending)
	stats["pending"] = int(pending)

	return stats, nil
}

// SendBroadcast sends broadcast message to all active users
func (s *BroadcastService) SendBroadcast(broadcastID uint) error {
	// Get broadcast message
	broadcast, err := s.GetBroadcastByID(broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast: %v", err)
	}

	// Get all active users
	var users []models.User
	err = s.db.DB.Where("is_active = ?", true).Find(&users).Error
	if err != nil {
		return fmt.Errorf("failed to get users: %v", err)
	}

	// Update broadcast status to sending
	s.db.DB.Model(broadcast).Update("status", "sending")

	// Create delivery records
	var deliveries []models.BroadcastDelivery
	for _, user := range users {
		delivery := models.BroadcastDelivery{
			BroadcastID: broadcastID,
			UserID:      user.TelegramID,
			Status:      "pending",
		}
		deliveries = append(deliveries, delivery)
	}

	// Batch create delivery records
	if len(deliveries) > 0 {
		s.db.DB.CreateInBatches(deliveries, 100)
	}

	// Send messages concurrently
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrent sends

	for _, user := range users {
		wg.Add(1)
		go func(u models.User) {
			defer wg.Done()
			semaphore <- struct{}{}        // Acquire semaphore
			defer func() { <-semaphore }() // Release semaphore

			s.sendToUser(broadcast, u.TelegramID)
		}(user)
	}

	wg.Wait()

	// Update final status
	s.db.DB.Model(broadcast).Update("status", "sent")

	return nil
}

// sendToUser sends broadcast message to a specific user
func (s *BroadcastService) sendToUser(broadcast *models.BroadcastMessage, userID int64) {
	var err error
	var delivery models.BroadcastDelivery

	// Get delivery record
	s.db.DB.Where("broadcast_id = ? AND user_id = ?", broadcast.ID, userID).First(&delivery)

	switch broadcast.ContentType {
	case "text":
		err = s.sendTextMessage(userID, broadcast.Text)
	case "photo":
		err = s.sendPhotoMessage(userID, broadcast.Text, broadcast.FileID, broadcast.FileURL)
	case "video":
		err = s.sendVideoMessage(userID, broadcast.Text, broadcast.FileID, broadcast.FileURL)
	case "document":
		err = s.sendDocumentMessage(userID, broadcast.Text, broadcast.FileID, broadcast.FileURL)
	case "voice":
		err = s.sendVoiceMessage(userID, broadcast.FileID, broadcast.FileURL)
	case "audio":
		err = s.sendAudioMessage(userID, broadcast.Text, broadcast.FileID, broadcast.FileURL)
	case "sticker":
		err = s.sendStickerMessage(userID, broadcast.FileID)
	case "animation":
		err = s.sendAnimationMessage(userID, broadcast.Text, broadcast.FileID, broadcast.FileURL)
	default:
		err = fmt.Errorf("unsupported content type: %s", broadcast.ContentType)
	}

	// Update delivery status
	if err != nil {
		log.Printf("Failed to send broadcast %d to user %d: %v", broadcast.ID, userID, err)
		delivery.Status = "failed"
		delivery.ErrorMessage = err.Error()
		s.db.DB.Model(&delivery).Updates(delivery)
	} else {
		delivery.Status = "sent"
		s.db.DB.Model(&delivery).Update("status", "sent")
	}
}

// sendTextMessage sends a text message
func (s *BroadcastService) sendTextMessage(userID int64, text string) error {
	msg := tgbotapi.NewMessage(userID, text)
	_, err := s.bot.Send(msg)
	return err
}

// sendPhotoMessage sends a photo with optional caption
func (s *BroadcastService) sendPhotoMessage(userID int64, caption, fileID, fileURL string) error {
	var photo tgbotapi.PhotoConfig

	if fileID != "" {
		photo = tgbotapi.NewPhoto(userID, tgbotapi.FileID(fileID))
	} else if fileURL != "" {
		if strings.HasPrefix(fileURL, "uploads/") || (!strings.HasPrefix(fileURL, "http")) {
			photo = tgbotapi.NewPhoto(userID, tgbotapi.FilePath(fileURL))
		} else {
			photo = tgbotapi.NewPhoto(userID, tgbotapi.FileURL(fileURL))
		}
	} else {
		return fmt.Errorf("no file ID or URL provided")
	}

	if caption != "" {
		photo.Caption = caption
	}

	_, err := s.bot.Send(photo)
	return err
}

// sendVideoMessage sends a video with optional caption
func (s *BroadcastService) sendVideoMessage(userID int64, caption, fileID, fileURL string) error {
	if fileID != "" {
		log.Printf("Sending video to user %d: FileID=%s, Caption=%s", userID, fileID, caption)

		// Try to send as regular video first
		video := tgbotapi.NewVideo(userID, tgbotapi.FileID(fileID))
		if caption != "" {
			video.Caption = caption
		}

		_, err := s.bot.Send(video)
		if err != nil {
			log.Printf("Failed to send video as regular video, trying as document: %v", err)
			// If regular video fails, try as document
			document := tgbotapi.NewDocument(userID, tgbotapi.FileID(fileID))
			if caption != "" {
				document.Caption = caption
			}
			_, err = s.bot.Send(document)
			if err != nil {
				log.Printf("Failed to send video as document: %v", err)
			}
		}
		return err
	} else if fileURL != "" {
		log.Printf("Sending video to user %d: FileURL=%s, Caption=%s", userID, fileURL, caption)

		if strings.HasPrefix(fileURL, "uploads/") || (!strings.HasPrefix(fileURL, "http")) {
			video := tgbotapi.NewVideo(userID, tgbotapi.FilePath(fileURL))
			if caption != "" {
				video.Caption = caption
			}
			_, err := s.bot.Send(video)
			return err
		} else {
			video := tgbotapi.NewVideo(userID, tgbotapi.FileURL(fileURL))
			if caption != "" {
				video.Caption = caption
			}
			_, err := s.bot.Send(video)
			return err
		}
	} else {
		return fmt.Errorf("no file ID or URL provided")
	}
}

// sendDocumentMessage sends a document with optional caption
func (s *BroadcastService) sendDocumentMessage(userID int64, caption, fileID, fileURL string) error {
	var doc tgbotapi.DocumentConfig

	if fileID != "" {
		doc = tgbotapi.NewDocument(userID, tgbotapi.FileID(fileID))
	} else if fileURL != "" {
		if strings.HasPrefix(fileURL, "uploads/") || (!strings.HasPrefix(fileURL, "http")) {
			doc = tgbotapi.NewDocument(userID, tgbotapi.FilePath(fileURL))
		} else {
			doc = tgbotapi.NewDocument(userID, tgbotapi.FileURL(fileURL))
		}
	} else {
		return fmt.Errorf("no file ID or URL provided")
	}

	if caption != "" {
		doc.Caption = caption
	}

	_, err := s.bot.Send(doc)
	return err
}

// sendVoiceMessage sends a voice message
func (s *BroadcastService) sendVoiceMessage(userID int64, fileID, fileURL string) error {
	var voice tgbotapi.VoiceConfig

	if fileID != "" {
		voice = tgbotapi.NewVoice(userID, tgbotapi.FileID(fileID))
	} else if fileURL != "" {
		if strings.HasPrefix(fileURL, "uploads/") || (!strings.HasPrefix(fileURL, "http")) {
			voice = tgbotapi.NewVoice(userID, tgbotapi.FilePath(fileURL))
		} else {
			voice = tgbotapi.NewVoice(userID, tgbotapi.FileURL(fileURL))
		}
	} else {
		return fmt.Errorf("no file ID or URL provided")
	}

	_, err := s.bot.Send(voice)
	return err
}

// sendAudioMessage sends an audio file with optional caption
func (s *BroadcastService) sendAudioMessage(userID int64, caption, fileID, fileURL string) error {
	var audio tgbotapi.AudioConfig

	if fileID != "" {
		audio = tgbotapi.NewAudio(userID, tgbotapi.FileID(fileID))
	} else if fileURL != "" {
		if strings.HasPrefix(fileURL, "uploads/") || (!strings.HasPrefix(fileURL, "http")) {
			audio = tgbotapi.NewAudio(userID, tgbotapi.FilePath(fileURL))
		} else {
			audio = tgbotapi.NewAudio(userID, tgbotapi.FileURL(fileURL))
		}
	} else {
		return fmt.Errorf("no file ID or URL provided")
	}

	if caption != "" {
		audio.Caption = caption
	}

	_, err := s.bot.Send(audio)
	return err
}

// sendStickerMessage sends a sticker
func (s *BroadcastService) sendStickerMessage(userID int64, fileID string) error {
	if fileID == "" {
		return fmt.Errorf("no file ID provided")
	}

	sticker := tgbotapi.NewSticker(userID, tgbotapi.FileID(fileID))
	_, err := s.bot.Send(sticker)
	return err
}

// sendAnimationMessage sends an animation (GIF) with optional caption
func (s *BroadcastService) sendAnimationMessage(userID int64, caption, fileID, fileURL string) error {
	var animation tgbotapi.AnimationConfig

	if fileID != "" {
		animation = tgbotapi.NewAnimation(userID, tgbotapi.FileID(fileID))
	} else if fileURL != "" {
		if strings.HasPrefix(fileURL, "uploads/") || (!strings.HasPrefix(fileURL, "http")) {
			animation = tgbotapi.NewAnimation(userID, tgbotapi.FilePath(fileURL))
		} else {
			animation = tgbotapi.NewAnimation(userID, tgbotapi.FileURL(fileURL))
		}
	} else {
		return fmt.Errorf("no file ID or URL provided")
	}

	if caption != "" {
		animation.Caption = caption
	}

	_, err := s.bot.Send(animation)
	return err
}

// GetUserCount returns the total number of active users
func (s *BroadcastService) GetUserCount() (int64, error) {
	var count int64
	err := s.db.DB.Model(&models.User{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}

// DeleteBroadcast soft deletes a broadcast
func (s *BroadcastService) DeleteBroadcast(id uint) error {
	return s.db.DB.Delete(&models.BroadcastMessage{}, id).Error
}
