package services

import (
	"fmt"
	"os"
	"telegram-bot/internal/config"

	"gopkg.in/yaml.v3"
)

type ConfigService struct {
	config     *config.Config
	configPath string
}

func NewConfigService(cfg *config.Config) *ConfigService {
	return &ConfigService{
		config:     cfg,
		configPath: "config.yaml", // Default config path
	}
}

// UpdateVideoMessageID updates the video message ID in config and saves to file
func (s *ConfigService) UpdateVideoMessageID(messageID int) error {
	s.config.Video.MessageID = messageID

	// Save to file
	return s.saveConfigToFile()
}

// UpdateGroupLink updates the group link in config and saves to file
func (s *ConfigService) UpdateGroupLink(groupLink string) error {
	s.config.Telegram.GroupLink = groupLink

	// Save to file
	return s.saveConfigToFileForTelegram()
}

// saveConfigToFile saves the current config to YAML file
func (s *ConfigService) saveConfigToFile() error {
	// Read the current config file to preserve structure and comments
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the YAML
	var configData map[string]interface{}
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	// Update the video message_id
	if videoConfig, ok := configData["video"].(map[string]interface{}); ok {
		videoConfig["message_id"] = s.config.Video.MessageID
	} else {
		// Create video section if it doesn't exist
		configData["video"] = map[string]interface{}{
			"message_id": s.config.Video.MessageID,
		}
	}

	// Write back to file
	updatedData, err := yaml.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(s.configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// saveConfigToFileForTelegram saves telegram config to YAML file
func (s *ConfigService) saveConfigToFileForTelegram() error {
	// Read the current config file to preserve structure and comments
	data, err := os.ReadFile(s.configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	// Parse the YAML
	var configData map[string]interface{}
	if err := yaml.Unmarshal(data, &configData); err != nil {
		return fmt.Errorf("failed to parse config: %v", err)
	}

	// Update the telegram group_link
	if telegramConfig, ok := configData["telegram"].(map[string]interface{}); ok {
		telegramConfig["group_link"] = s.config.Telegram.GroupLink
	}

	// Write back to file
	updatedData, err := yaml.Marshal(configData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	if err := os.WriteFile(s.configPath, updatedData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %v", err)
	}

	return nil
}

// GetCurrentVideoMessageID returns the current video message ID
func (s *ConfigService) GetCurrentVideoMessageID() int {
	return s.config.Video.MessageID
}

// GetCurrentGroupLink returns the current group link
func (s *ConfigService) GetCurrentGroupLink() string {
	return s.config.Telegram.GroupLink
}
