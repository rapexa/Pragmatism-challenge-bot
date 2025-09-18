package handlers

import (
	"fmt"
	"log"
	"strings"
	"telegram-bot/internal/config"
	"telegram-bot/internal/models"
	"telegram-bot/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot               *tgbotapi.BotAPI
	userService       *services.UserService
	supportService    *services.SupportService
	adminPanelService *services.AdminPanelService
	adminHandler      *AdminHandler
	config            *config.Config
}

func NewBotHandler(bot *tgbotapi.BotAPI, userService *services.UserService, supportService *services.SupportService, adminPanelService *services.AdminPanelService, configService *services.ConfigService, cfg *config.Config) *BotHandler {
	adminHandler := NewAdminHandler(bot, adminPanelService, configService, cfg)

	return &BotHandler{
		bot:               bot,
		userService:       userService,
		supportService:    supportService,
		adminPanelService: adminPanelService,
		adminHandler:      adminHandler,
		config:            cfg,
	}
}

func (h *BotHandler) HandleUpdate(update tgbotapi.Update) {
	var telegramID int64

	if update.Message != nil {
		telegramID = update.Message.From.ID
	} else if update.CallbackQuery != nil {
		telegramID = update.CallbackQuery.From.ID
	} else {
		return
	}

	// Check if user is admin
	if h.adminPanelService.IsAdmin(telegramID) {
		h.adminHandler.HandleAdminUpdate(update, telegramID)
		return
	}

	// Handle regular user updates
	if update.Message != nil {
		h.handleMessage(update.Message)
	}
}

func (h *BotHandler) handleMessage(message *tgbotapi.Message) {
	telegramID := message.From.ID

	// Check if user is already registered
	user, support, err := h.userService.GetUserWithSupport(telegramID)
	if err != nil {
		log.Printf("Error checking user: %v", err)
		h.sendMessage(telegramID, "خطایی رخ داده است. لطفاً دوباره تلاش کنید.")
		return
	}

	// If user is registered and sends /start, send video
	if user != nil && message.Command() == "start" {
		h.sendVideoWithSupport(telegramID, support)
		return
	}

	// If user is registered but sends other messages
	if user != nil {
		h.sendMessage(telegramID, "شما قبلاً ثبت نام کرده‌اید. از ربات استفاده کنید.")
		return
	}

	// Handle registration process
	h.handleRegistration(message)
}

func (h *BotHandler) handleRegistration(message *tgbotapi.Message) {
	telegramID := message.From.ID
	text := message.Text

	// Check if it's start command
	if message.Command() == "start" {
		h.userService.StartRegistration(telegramID)
		h.sendMessage(telegramID, "سلام! به ربات ما خوش آمدید 🌟\n\nلطفاً نام و نام خانوادگی خود را وارد کنید:\n(مثال: احمد احمدی)")
		return
	}

	// Get registration state
	state := h.userService.GetRegistrationState(telegramID)
	if state == nil {
		h.sendMessage(telegramID, "لطفاً ابتدا دستور /start را ارسال کنید.")
		return
	}

	switch state.Step {
	case "waiting_name":
		h.handleNameInput(telegramID, text)
	case "waiting_phone":
		h.handlePhoneInput(telegramID, message)
	case "waiting_job":
		h.handleJobInput(telegramID, text)
	}
}

func (h *BotHandler) handleNameInput(telegramID int64, text string) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		h.sendMessage(telegramID, "لطفاً نام و نام خانوادگی خود را به صورت کامل وارد کنید.\n(مثال: احمد احمدی)")
		return
	}

	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")

	h.userService.UpdateRegistrationState(telegramID, "waiting_phone", map[string]string{
		"first_name": firstName,
		"last_name":  lastName,
	})

	// Request phone number with keyboard
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 ارسال شماره تماس"),
		),
	)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true

	msg := tgbotapi.NewMessage(telegramID, fmt.Sprintf("سلام %s %s! 👋\n\nحالا لطفاً شماره تماس خود را ارسال کنید:", firstName, lastName))
	msg.ReplyMarkup = keyboard

	h.bot.Send(msg)
}

func (h *BotHandler) handlePhoneInput(telegramID int64, message *tgbotapi.Message) {
	var phoneNumber string

	// Check if contact was shared
	if message.Contact != nil {
		phoneNumber = message.Contact.PhoneNumber
	} else if message.Text != "" {
		phoneNumber = message.Text
	} else {
		h.sendMessage(telegramID, "لطفاً شماره تماس خود را ارسال کنید یا از دکمه زیر استفاده کنید.")
		return
	}

	// Validate phone number (basic validation)
	if len(phoneNumber) < 10 {
		h.sendMessage(telegramID, "شماره تماس وارد شده معتبر نیست. لطفاً دوباره تلاش کنید.")
		return
	}

	h.userService.UpdateRegistrationState(telegramID, "waiting_job", map[string]string{
		"phone_number": phoneNumber,
	})

	// Remove keyboard and ask for job
	removeKeyboard := tgbotapi.NewRemoveKeyboard(true)
	msg := tgbotapi.NewMessage(telegramID, "عالی! ✅\n\nحالا لطفاً شغل خود را وارد کنید:")
	msg.ReplyMarkup = removeKeyboard

	h.bot.Send(msg)
}

func (h *BotHandler) handleJobInput(telegramID int64, job string) {
	if strings.TrimSpace(job) == "" {
		h.sendMessage(telegramID, "لطفاً شغل خود را وارد کنید.")
		return
	}

	// Get registration state to retrieve saved data
	state := h.userService.GetRegistrationState(telegramID)
	if state == nil {
		h.sendMessage(telegramID, "خطایی رخ داده است. لطفاً دوباره شروع کنید.")
		return
	}

	// Complete registration
	username := ""
	if state.TelegramID != 0 {
		// Try to get username from Telegram (this would require storing it during registration)
		// For now, we'll leave it empty
	}

	err := h.userService.CompleteRegistration(telegramID, state.PhoneNumber, job, username)
	if err != nil {
		log.Printf("Error completing user registration: %v", err)
		h.sendMessage(telegramID, "خطایی در ثبت نام رخ داده است. لطفاً دوباره تلاش کنید.")
		return
	}

	h.sendMessage(telegramID, "🎉 ثبت نام شما با موفقیت تکمیل شد!\n\nدر حال ارسال ویدیو...")

	// Get user with support info for sending video
	_, support, err := h.userService.GetUserWithSupport(telegramID)
	if err != nil || support == nil {
		log.Printf("Error getting user support info: %v", err)
		h.sendMessage(telegramID, "خطایی در دریافت اطلاعات پشتیبانی رخ داده است.")
		return
	}

	// Send video with support info
	h.sendVideoWithSupport(telegramID, support)
}

func (h *BotHandler) sendVideoWithSupport(telegramID int64, support *models.SupportStaff) {
	// Copy the specific message from the channel with custom caption
	messageID := h.config.Video.MessageID
	if messageID == 0 {
		messageID = 2 // Default fallback
	}

	copyConfig := tgbotapi.CopyMessageConfig{
		BaseChat: tgbotapi.BaseChat{
			ChatID: telegramID,
		},
		FromChatID: h.config.Telegram.ChannelID,
		MessageID:  messageID,
		Caption:    "ثبت نام شما با موفقیت انجام شد",
	}

	_, err := h.bot.Send(copyConfig)
	if err != nil {
		log.Printf("Error copying message from channel: %v", err)
		h.sendMessage(telegramID, "خطایی در ارسال ویدیو رخ داده است.")
		return
	}

	// Check if support staff is available
	if support == nil {
		log.Println("No support staff assigned to user")
		h.sendMessage(telegramID, "ویدیو ارسال شد اما اطلاعات پشتیبانی در دسترس نیست.")
		return
	}

	// Send support photo with complete info as caption
	if support.PhotoURL != "" {
		photo := tgbotapi.NewPhoto(telegramID, tgbotapi.FileURL(support.PhotoURL))
		photo.Caption = fmt.Sprintf("👤 پشتیبان شما: %s\n📞 آیدی پشتیبان: %s\n🔗 لینک گروه: %s",
			support.Name,
			support.Username,
			h.config.Telegram.GroupLink,
		)
		h.bot.Send(photo)
	} else {
		// If no photo available, send text message
		supportMessage := fmt.Sprintf("👤 پشتیبان شما: %s\n📞 آیدی پشتیبان: %s\n🔗 لینک گروه: %s",
			support.Name,
			support.Username,
			h.config.Telegram.GroupLink,
		)
		h.sendMessage(telegramID, supportMessage)
	}
}

func (h *BotHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
