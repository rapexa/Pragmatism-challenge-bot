package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type FileService struct {
	bot        *tgbotapi.BotAPI
	uploadPath string
	serverURL  string
}

func NewFileService(bot *tgbotapi.BotAPI, serverURL string) *FileService {
	uploadPath := "uploads"

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		log.Printf("Error creating uploads directory: %v", err)
	}

	return &FileService{
		bot:        bot,
		uploadPath: uploadPath,
		serverURL:  serverURL,
	}
}

// DownloadPhoto downloads a photo from Telegram and saves it locally, returns the local path
func (s *FileService) DownloadPhoto(fileID string) (string, error) {
	// Get file info from Telegram
	file, err := s.bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %v", err)
	}

	// Generate unique filename
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("support_photo_%d.jpg", timestamp)
	localPath := filepath.Join(s.uploadPath, filename)

	// Download file from Telegram
	fileURL := file.Link(s.bot.Token)
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// Create local file
	localFile, err := os.Create(localPath)
	if err != nil {
		return "", fmt.Errorf("failed to create local file: %v", err)
	}
	defer localFile.Close()

	// Copy file content
	_, err = io.Copy(localFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// Return local path for Telegram file sending
	log.Printf("Photo saved successfully: %s", localPath)
	return localPath, nil
}

// GetPhotoPath returns the full path for a photo file
func (s *FileService) GetPhotoPath(filename string) string {
	return filepath.Join(s.uploadPath, filename)
}

// DeletePhoto deletes a photo file
func (s *FileService) DeletePhoto(filepath string) error {
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	return os.Remove(filepath)
}

// GetPhotoURL returns the URL for accessing the photo
func (s *FileService) GetPhotoURL(filename string) string {
	// This would be your server's URL where photos are accessible
	// For now, return the local path - you'll need to serve these files via HTTP
	return fmt.Sprintf("/uploads/%s", filename)
}
