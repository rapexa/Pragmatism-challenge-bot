package handlers

import (
	"fmt"
	"log"
	"strings"
	"telegram-bot/internal/config"
	"telegram-bot/internal/keyboards"
	"telegram-bot/internal/models"
	"telegram-bot/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type BotHandler struct {
	bot               *tgbotapi.BotAPI
	userService       *services.UserService
	supportService    *services.SupportService
	adminPanelService *services.AdminPanelService
	fileService       *services.FileService
	adminHandler      *AdminHandler
	config            *config.Config
}

func NewBotHandler(bot *tgbotapi.BotAPI, userService *services.UserService, supportService *services.SupportService, adminPanelService *services.AdminPanelService, configService *services.ConfigService, fileService *services.FileService, cfg *config.Config) *BotHandler {
	adminHandler := NewAdminHandler(bot, adminPanelService, configService, fileService, cfg)

	return &BotHandler{
		bot:               bot,
		userService:       userService,
		supportService:    supportService,
		adminPanelService: adminPanelService,
		fileService:       fileService,
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
		h.sendWelcomeBackMessage(telegramID, user, support)
		return
	}

	// If user is registered but sends other messages
	if user != nil {
		h.sendMessage(telegramID, "سلام دوست عزیز! 👋\n\nشما قبلاً در ربات ثبت نام کرده‌اید.\n\nبرای مشاهده مجدد ویدیو و اطلاعات پشتیبان، دستور /start را ارسال کنید.")
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
		welcomeMessage := `سلام! به ربات چالش عملگرایی خوش آمدید! 🌟

برای شروع، لطفاً نام و نام خانوادگی خود را وارد کنید:

📝 مثال: احمد احمدی`
		h.sendMessage(telegramID, welcomeMessage)
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
		errorMessage := `❌ لطفاً نام و نام خانوادگی خود را به صورت کامل وارد کنید.

📝 مثال صحیح: احمد احمدی`
		h.sendMessage(telegramID, errorMessage)
		return
	}

	firstName := parts[0]
	lastName := strings.Join(parts[1:], " ")

	h.userService.UpdateRegistrationState(telegramID, "waiting_phone", map[string]string{
		"first_name": firstName,
		"last_name":  lastName,
	})

	// Request phone number with keyboard
	phoneMessage := fmt.Sprintf(`عالی %s عزیز! ✅

حالا لطفاً شماره تماس خود را ارسال کنید:

📱 می‌توانید از دکمه زیر استفاده کنید یا شماره را تایپ کنید`, firstName)

	msg := tgbotapi.NewMessage(telegramID, phoneMessage)
	msg.ReplyMarkup = keyboards.PhoneRequestKeyboard()

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
		retryMessage := `📱 شماره تماس دریافت نشد!

لطفاً یکی از روش‌های زیر را انتخاب کنید:
• از دکمه "📱 ارسال شماره تماس" استفاده کنید
• یا شماره خود را تایپ کنید (مثال: 09123456789)`

		msg := tgbotapi.NewMessage(telegramID, retryMessage)
		msg.ReplyMarkup = keyboards.PhoneRequestKeyboard()
		h.bot.Send(msg)
		return
	}

	// Validate phone number (basic validation)
	if len(phoneNumber) < 10 {
		errorMessage := `❌ شماره تماس وارد شده معتبر نیست!

📱 لطفاً شماره تماس معتبر وارد کنید:
• مثال: 09123456789
• یا از دکمه ارسال شماره استفاده کنید`

		msg := tgbotapi.NewMessage(telegramID, errorMessage)
		msg.ReplyMarkup = keyboards.PhoneRequestKeyboard()
		h.bot.Send(msg)
		return
	}

	h.userService.UpdateRegistrationState(telegramID, "waiting_job", map[string]string{
		"phone_number": phoneNumber,
	})

	// Remove keyboard and ask for job
	jobMessage := `عالی! شماره تماس شما ثبت شد ✅

حالا لطفاً شغل یا تخصص خود را وارد کنید:

💼 مثال: مهندس نرم‌افزار، معلم، پزشک، دانشجو و ...`

	msg := tgbotapi.NewMessage(telegramID, jobMessage)
	msg.ReplyMarkup = keyboards.RemoveKeyboard()

	h.bot.Send(msg)
}

func (h *BotHandler) handleJobInput(telegramID int64, job string) {
	if strings.TrimSpace(job) == "" {
		jobErrorMessage := `💼 شغل وارد نشده!

لطفاً شغل یا تخصص خود را وارد کنید:

📝 مثال‌هایی از شغل‌ها:
• مهندس نرم‌افزار
• معلم ریاضی  
• پزشک عمومی
• دانشجوی پزشکی
• کارمند اداری`
		h.sendMessage(telegramID, jobErrorMessage)
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

	successMessage := `🎉 تبریک! ثبت نام شما با موفقیت تکمیل شد!

لطفاً ویدیو آموزشی بالا را مشاهده کنید 👆

در حال ارسال اطلاعات پشتیبان شما...`

	h.sendMessage(telegramID, successMessage)

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

func (h *BotHandler) sendWelcomeBackMessage(telegramID int64, user *models.User, support *models.SupportStaff) {
	welcomeBackMessage := fmt.Sprintf(`سلام مجدد %s عزیز! 👋

خوش برگشتید! 🌟

در حال ارسال مجدد ویدیو آموزشی و اطلاعات پشتیبان شما...`, user.FirstName)

	h.sendMessage(telegramID, welcomeBackMessage)

	// Send video with different caption for returning users
	h.sendVideoWithSupportAndCaption(telegramID, support, "ویدیو آموزشی ربات چالش عملگرایی")
}

func (h *BotHandler) sendVideoWithSupport(telegramID int64, support *models.SupportStaff) {
	h.sendVideoWithSupportAndCaption(telegramID, support, "ثبت نام شما با موفقیت انجام شد")
}

func (h *BotHandler) sendVideoWithSupportAndCaption(telegramID int64, support *models.SupportStaff, caption string) {
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
		Caption:    caption,
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
		var photo tgbotapi.PhotoConfig

		// Check if it's a local file or external URL
		if strings.HasPrefix(support.PhotoURL, "uploads/") || (!strings.HasPrefix(support.PhotoURL, "http")) {
			// Local file - send as file path
			photo = tgbotapi.NewPhoto(telegramID, tgbotapi.FilePath(support.PhotoURL))
		} else {
			// External URL
			photo = tgbotapi.NewPhoto(telegramID, tgbotapi.FileURL(support.PhotoURL))
		}

		photo.Caption = fmt.Sprintf(`👨‍💼 پشتیبان اختصاصی شما:

👤 نام: %s
📞 آیدی تلگرام: %s

🔗 لینک گروه VIP:
%s

💬 برای ارتباط با پشتیبان، روی آیدی بالا کلیک کنید`,
			support.Name,
			support.Username,
			h.config.Telegram.GroupLink,
		)
		h.bot.Send(photo)
	} else {
		// If no photo available, send text message
		supportMessage := fmt.Sprintf(`👨‍💼 پشتیبان اختصاصی شما:

👤 نام: %s
📞 آیدی تلگرام: %s

🔗 لینک گروه VIP:
%s

💬 برای ارتباط با پشتیبان، روی آیدی بالا کلیک کنید`,
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
