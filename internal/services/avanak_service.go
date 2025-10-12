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
	QuickSendID              int    `json:"QuickSendID"`
	MessageID                int    `json:"MessageID"`
	MessageLength            int    `json:"MessageLength"`
	CreditDecrease_InSeconds int    `json:"CreditDecrease_InSeconds"`
	CreditDecrease_InPulses  int    `json:"CreditDecrease_InPulses"`
	CreditDecrease_InPrice   int    `json:"CreditDecrease_InPrice"`
	Status                   string `json:"Status"`
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
	if avanakResp.Status == "Success" && avanakResp.QuickSendID > 0 {
		log.Printf("Voice call sent successfully to %s, QuickSend ID: %d", phoneNumber, avanakResp.QuickSendID)
		return nil
	} else {
		errorCode := avanakResp.QuickSendID
		errorMsg := s.getErrorMessage(errorCode)
		log.Printf("Avanak error for %s: %s (code: %d), Status: %s", phoneNumber, errorMsg, errorCode, avanakResp.Status)
		return fmt.Errorf("Avanak error: %s (code: %d)", errorMsg, errorCode)
	}
}

// getErrorMessage returns human-readable error message for Avanak error codes
func (s *AvanakService) getErrorMessage(errorCode int) string {
	switch errorCode {
	// خطاهای اصلی
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

	// خطاهای احراز هویت
	case -1:
		return "نام کاربری یا گذرواژه اشتباه است"
	case -20:
		return "خطای ناشناخته"
	case -102:
		return "عدم احراز موبایل"
	case -103:
		return "کاربری غیرفعال شده"
	case -104:
		return "کاربری منقضی شده"
	case -105:
		return "دسترسی به وب سرویس مسدود شده"
	case -106:
		return "عدم مجوز وب سرویس"
	case -107:
		return "آی پی غیرمجاز"
	case -108:
		return "عدم مجوز متد"
	case -109:
		return "عدم مجوز استفاده از پروتکل Http"

	// خطاهای توکن
	case -201:
		return "توکن اشتباه است"
	case -202:
		return "توکن اشتباه است"
	case -203:
		return "توکن غیرفعال است"
	case -204:
		return "توکن منقضی شده است"
	case -207:
		return "آی پی غیرمجاز"
	case -208:
		return "عدم مجوز متد"

	// خطاهای سیستم
	case -300:
		return "سیستم غیرفعال است"

	default:
		return fmt.Sprintf("خطای نامشخص (کد: %d)", errorCode)
	}
}

// IsEnabled returns whether Avanak service is enabled
func (s *AvanakService) IsEnabled() bool {
	return s.config.Enabled
}

// TestConnection tests the Avanak API connection by sending a real voice call
func (s *AvanakService) TestConnection() error {
	if !s.config.Enabled {
		return fmt.Errorf("Avanak service is disabled")
	}

	// Real test with your phone number
	testNumber := "09155520952"

	log.Printf("🔔 ارسال تماس تست به شماره %s...", testNumber)

	// Use the same SendVoiceCall method for consistency
	err := s.SendVoiceCall(testNumber)
	if err != nil {
		log.Printf("❌ تست اتصال اوانک ناموفق: %v", err)
		return fmt.Errorf("Avanak test failed: %v", err)
	}

	log.Printf("✅ تست اتصال اوانک موفق - تماس به %s ارسال شد", testNumber)
	return nil
}
