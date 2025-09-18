package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"telegram-bot/internal/config"
)

type SMSService struct {
	config *config.SMSConfig
}

type ippanelPatternRequest struct {
	Code      string            `json:"code"`
	Sender    string            `json:"sender"`
	Recipient string            `json:"recipient"`
	Variable  map[string]string `json:"variable"`
}

func NewSMSService(smsConfig *config.SMSConfig) *SMSService {
	return &SMSService{
		config: smsConfig,
	}
}

// SendSMS sends SMS using IPPanel pattern
func (s *SMSService) SendSMS(phone string, params map[string]string, patternKey string) error {
	// Get pattern code
	patternCode, exists := s.config.Patterns[patternKey]
	if !exists {
		return fmt.Errorf("pattern %s not found", patternKey)
	}

	body := ippanelPatternRequest{
		Code:      patternCode,
		Sender:    s.config.FromNumber,
		Recipient: phone,
		Variable:  params,
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		log.Printf("Failed to marshal SMS request: %v", err)
		return err
	}

	requestURL := s.config.BaseURL + "/sms/pattern/normal/send"
	log.Printf("Sending SMS to URL: %s", requestURL)
	log.Printf("SMS request body: %s", string(jsonBody))

	req, err := http.NewRequest("POST", requestURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Failed to create SMS request: %v", err)
		return err
	}

	req.Header.Set("apikey", s.config.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("SMS request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Printf("SMS response body: %s", string(respBody))

	if resp.StatusCode != 200 {
		log.Printf("SMS non-200 response: %d", resp.StatusCode)
		return fmt.Errorf("SMS API returned status %d", resp.StatusCode)
	}

	log.Printf("SMS sent successfully to %s: %+v (pattern: %s)", phone, params, patternKey)
	return nil
}

// SendRegistrationSMS sends registration success SMS
func (s *SMSService) SendRegistrationSMS(phone, firstName string) error {
	params := map[string]string{
		"name": firstName,
	}

	return s.SendSMS(phone, params, "registration")
}
