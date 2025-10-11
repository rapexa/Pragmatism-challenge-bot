package services

import (
	"fmt"
	"log"
	"telegram-bot/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChannelService struct {
	bot    *tgbotapi.BotAPI
	config *config.TelegramConfig
}

func NewChannelService(bot *tgbotapi.BotAPI, telegramConfig *config.TelegramConfig) *ChannelService {
	return &ChannelService{
		bot:    bot,
		config: telegramConfig,
	}
}

// IsUserMemberOfMandatoryChannel checks if user is member of mandatory channel
func (s *ChannelService) IsUserMemberOfMandatoryChannel(userID int64) (bool, error) {
	// Get chat member status
	chatMemberConfig := tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{
			ChatID: s.config.MandatoryChannelID,
			UserID: userID,
		},
	}

	chatMember, err := s.bot.GetChatMember(chatMemberConfig)
	if err != nil {
		log.Printf("Error checking channel membership for user %d: %v", userID, err)
		return false, err
	}

	// Check if user is a member (not left, kicked, or restricted)
	isMember := chatMember.Status == "member" ||
		chatMember.Status == "administrator" ||
		chatMember.Status == "creator"

	return isMember, nil
}

// SendMandatoryChannelJoinMessage sends message asking user to join mandatory channel
func (s *ChannelService) SendMandatoryChannelJoinMessage(chatID int64) error {
	message := fmt.Sprintf(`🔔 عضویت در کانال

برای استفاده از ربات، ابتدا باید در کانال زیر عضو شوید:

📢 کانال: %s

مراحل عضویت:
1️⃣ روی دکمه "📢 عضویت در کانال" کلیک کنید
2️⃣ در کانال عضو شوید
3️⃣ روی "✅ عضو شدم" کلیک کنید

💡 پس از عضویت، می‌توانید از تمام قابلیت‌های ربات استفاده کنید.`, s.config.MandatoryChannelUsername)

	// Create inline keyboard with channel join button
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📢 عضویت در کانال", fmt.Sprintf("https://t.me/%s", s.config.MandatoryChannelUsername[1:])), // Remove @ from username
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ عضو شدم", "check_membership"),
		),
	)

	msg := tgbotapi.NewMessage(chatID, message)
	msg.ReplyMarkup = keyboard

	_, err := s.bot.Send(msg)
	return err
}

// GetMandatoryChannelInfo returns mandatory channel information
func (s *ChannelService) GetMandatoryChannelInfo() (int64, string) {
	return s.config.MandatoryChannelID, s.config.MandatoryChannelUsername
}
