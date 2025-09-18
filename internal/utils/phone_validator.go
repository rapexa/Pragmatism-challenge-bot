package utils

import (
	"regexp"
	"strings"
)

// ValidateIranianPhoneNumber validates Iranian phone numbers
func ValidateIranianPhoneNumber(phone string) (bool, string) {
	// Remove all spaces, dashes, and other non-digit characters except +
	cleanPhone := regexp.MustCompile(`[^\d+]`).ReplaceAllString(phone, "")

	// Define valid Iranian mobile prefixes
	validPrefixes := []string{
		// Hamrah-e Avval (MCI)
		"0910", "0911", "0912", "0913", "0914", "0915", "0916", "0917", "0918", "0919",
		"0990", "0991", "0992", "0993", "0994", "0995", "0996", "0997", "0998", "0999",
		// Irancell
		"0901", "0902", "0903", "0905", "0930", "0933", "0934", "0935", "0936", "0937", "0938", "0939",
		// Rightel
		"0920", "0921", "0922",
		// TeleKish
		"0934",
		// MTCE
		"0932",
	}

	// Remove country code if present
	if strings.HasPrefix(cleanPhone, "+98") {
		cleanPhone = "0" + cleanPhone[3:]
	} else if strings.HasPrefix(cleanPhone, "0098") {
		cleanPhone = "0" + cleanPhone[4:]
	} else if strings.HasPrefix(cleanPhone, "98") && len(cleanPhone) == 12 {
		cleanPhone = "0" + cleanPhone[2:]
	}

	// Check if it starts with 09 and has 11 digits
	if !strings.HasPrefix(cleanPhone, "09") || len(cleanPhone) != 11 {
		return false, ""
	}

	// Check if it matches valid Iranian mobile patterns
	isValid := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(cleanPhone, prefix) {
			isValid = true
			break
		}
	}

	if !isValid {
		return false, ""
	}

	// Return normalized phone number
	return true, cleanPhone
}

// FormatIranianPhoneNumber formats Iranian phone number for display
func FormatIranianPhoneNumber(phone string) string {
	if len(phone) == 11 && strings.HasPrefix(phone, "09") {
		// Format as: 0912 345 6789
		return phone[:4] + " " + phone[4:7] + " " + phone[7:]
	}
	return phone
}

// GetPhoneNumberError returns appropriate error message for invalid phone
func GetPhoneNumberError() string {
	return `❌ شماره تماس وارد شده معتبر نیست!

📱 شماره تماس باید:
• شماره موبایل ایرانی باشد
• با 09 شروع شود
• 11 رقم داشته باشد

✅ مثال‌های صحیح:
• 09123456789
• 0912 345 6789
• +98 912 345 6789

💡 یا از دکمه "📱 ارسال شماره تماس" استفاده کنید`
}
