package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"telegram-bot/internal/config"
)

type AvanakService struct {
	config *config.AvanakConfig
}

type AvanakRequest struct {
	MessageID int    `json:"MessageID"`
	Number    string `json:"Number"`
	Vote      bool   `json:"Vote"`
	ServerID  int    `json:"ServerID"`
}

type AvanakResponse struct {
	ReturnValue int `json:"ReturnValue"`
}

func NewAvanakService(avanakConfig *config.AvanakConfig) *AvanakService {
	return &AvanakService{
		config: avanakConfig,
	}
}

// SendVoiceCall sends a voice call using Avanak API
func (s *AvanakService) SendVoiceCall(phoneNumber string) error {
	if !s.config.Enabled {
		log.Printf("Avanak service is disabled, skipping voice call to %s", phoneNumber)
		return nil
	}

	// Prepare request data
	data := url.Values{}
	data.Set("MessageID", fmt.Sprintf("%d", s.config.MessageID))
	data.Set("Number", phoneNumber)
	data.Set("Vote", "false")
	data.Set("ServerID", "0")

	// Create HTTP request
	req, err := http.NewRequest("POST", s.config.BaseURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		log.Printf("Failed to create Avanak request: %v", err)
		return err
	}

	// Set headers
	req.Header.Set("Authorization", s.config.Token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	log.Printf("Sending voice call to %s via Avanak", phoneNumber)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Avanak request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Failed to read Avanak response: %v", err)
		return err
	}

	log.Printf("Avanak response: %s", string(respBody))

	// Parse response
	var avanakResp AvanakResponse
	if err := json.Unmarshal(respBody, &avanakResp); err != nil {
		log.Printf("Failed to parse Avanak response: %v", err)
		return err
	}

	// Check result
	if avanakResp.ReturnValue > 0 {
		log.Printf("Voice call sent successfully to %s, QuickSend ID: %d", phoneNumber, avanakResp.ReturnValue)
		return nil
	} else {
		errorMsg := s.getErrorMessage(avanakResp.ReturnValue)
		log.Printf("Avanak error for %s: %s (code: %d)", phoneNumber, errorMsg, avanakResp.ReturnValue)
		return fmt.Errorf("Avanak error: %s (code: %d)", errorMsg, avanakResp.ReturnValue)
	}
}

// getErrorMessage returns human-readable error message for Avanak error codes
func (s *AvanakService) getErrorMessage(errorCode int) string {
	switch errorCode {
	case -25:
		return "ثبت ارسال سریع غیرفعال میباشد"
	case -2:
		return "شماره اشتباه میباشد"
	case -3:
		return "عدم موجودی کافی"
	case -6:
		return "زمان ارسال غیرمجاز میباشد"
	case -8:
		return "کد فایل صوتی اشتباه میباشد"
	case -71:
		return "مدت ضبط صدا غیرمجاز میباشد"
	case -72:
		return "عدم مجوز ضبط صدا"
	default:
		return fmt.Sprintf("خطای نامشخص (کد: %d)", errorCode)
	}
}

// IsEnabled returns whether Avanak service is enabled
func (s *AvanakService) IsEnabled() bool {
	return s.config.Enabled
}
