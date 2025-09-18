package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-bot/internal/config"
	"telegram-bot/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminHandler struct {
	bot               *tgbotapi.BotAPI
	adminPanelService *services.AdminPanelService
	configService     *services.ConfigService
	config            *config.Config
	adminStates       map[int64]string // Track admin states for multi-step operations
}

func NewAdminHandler(bot *tgbotapi.BotAPI, adminPanelService *services.AdminPanelService, configService *services.ConfigService, cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		bot:               bot,
		adminPanelService: adminPanelService,
		configService:     configService,
		config:            cfg,
		adminStates:       make(map[int64]string),
	}
}

func (h *AdminHandler) HandleAdminUpdate(update tgbotapi.Update, telegramID int64) {
	if update.Message != nil {
		h.handleAdminMessage(update.Message, telegramID)
	} else if update.CallbackQuery != nil {
		h.handleAdminCallback(update.CallbackQuery, telegramID)
	}
}

func (h *AdminHandler) handleAdminMessage(message *tgbotapi.Message, telegramID int64) {
	text := message.Text

	// Handle states for multi-step operations
	if state, exists := h.adminStates[telegramID]; exists {
		h.handleAdminState(message, telegramID, state)
		return
	}

	// Handle commands and text
	switch {
	case message.Command() == "start" || text == "🏠 صفحه اصلی":
		h.sendAdminMainMenu(telegramID)
	case text == "📊 آمار کاربران":
		h.sendUserStats(telegramID)
	case text == "📋 خروجی اکسل کاربران":
		h.exportUsers(telegramID)
	case text == "👥 مدیریت پشتیبان‌ها":
		h.sendSupportManagementMenu(telegramID)
	case text == "➕ افزودن پشتیبان":
		h.startAddSupport(telegramID)
	case text == "📝 ویرایش پشتیبان":
		h.showSupportList(telegramID, "edit")
	case text == "🗑 حذف پشتیبان":
		h.showSupportList(telegramID, "delete")
	case text == "🎬 تنظیمات ویدیو":
		h.sendVideoSettings(telegramID)
	default:
		h.sendAdminMainMenu(telegramID)
	}
}

func (h *AdminHandler) sendAdminMainMenu(telegramID int64) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 آمار کاربران"),
			tgbotapi.NewKeyboardButton("📋 خروجی اکسل کاربران"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("👥 مدیریت پشتیبان‌ها"),
			tgbotapi.NewKeyboardButton("🎬 تنظیمات ویدیو"),
		),
	)
	keyboard.ResizeKeyboard = true

	msg := tgbotapi.NewMessage(telegramID, "🔧 پنل مدیریت ربات\n\nیکی از گزینه‌های زیر را انتخاب کنید:")
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *AdminHandler) sendUserStats(telegramID int64) {
	stats, err := h.adminPanelService.GetUserStats()
	if err != nil {
		log.Printf("Error getting user stats: %v", err)
		h.sendMessage(telegramID, "خطا در دریافت آمار کاربران")
		return
	}

	message := fmt.Sprintf(`📊 آمار کاربران:

👥 کل کاربران: %d
✅ کاربران فعال: %d
📅 ثبت نام امروز: %d
📈 ثبت نام این هفته: %d
📊 ثبت نام این ماه: %d`,
		stats["total"],
		stats["active"],
		stats["today"],
		stats["week"],
		stats["month"])

	h.sendMessage(telegramID, message)
}

func (h *AdminHandler) exportUsers(telegramID int64) {
	h.sendMessage(telegramID, "در حال تهیه فایل اکسل...")

	filename, err := h.adminPanelService.ExportUsersToExcel()
	if err != nil {
		log.Printf("Error exporting users: %v", err)
		h.sendMessage(telegramID, "خطا در تهیه فایل اکسل")
		return
	}

	// Send file
	doc := tgbotapi.NewDocument(telegramID, tgbotapi.FilePath(filename))
	doc.Caption = "📋 لیست تمام کاربران ثبت نام شده"

	_, err = h.bot.Send(doc)
	if err != nil {
		log.Printf("Error sending document: %v", err)
		h.sendMessage(telegramID, "خطا در ارسال فایل")
	}
}

func (h *AdminHandler) sendSupportManagementMenu(telegramID int64) {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ افزودن پشتیبان"),
			tgbotapi.NewKeyboardButton("📝 ویرایش پشتیبان"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🗑 حذف پشتیبان"),
			tgbotapi.NewKeyboardButton("🏠 صفحه اصلی"),
		),
	)
	keyboard.ResizeKeyboard = true

	msg := tgbotapi.NewMessage(telegramID, "👥 مدیریت پشتیبان‌ها\n\nیکی از گزینه‌های زیر را انتخاب کنید:")
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *AdminHandler) startAddSupport(telegramID int64) {
	h.adminStates[telegramID] = "add_support_name"
	h.sendMessage(telegramID, "نام پشتیبان جدید را وارد کنید:")
}

func (h *AdminHandler) showSupportList(telegramID int64, action string) {
	staff, err := h.adminPanelService.GetSupportStaffList()
	if err != nil {
		log.Printf("Error getting support staff: %v", err)
		h.sendMessage(telegramID, "خطا در دریافت لیست پشتیبان‌ها")
		return
	}

	if len(staff) == 0 {
		h.sendMessage(telegramID, "هیچ پشتیبانی یافت نشد")
		return
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, s := range staff {
		status := "🟢"
		if !s.IsActive {
			status = "🔴"
		}

		text := fmt.Sprintf("%s %s", status, s.Name)
		callbackData := fmt.Sprintf("%s_%d", action, s.ID)

		keyboard = append(keyboard, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(text, callbackData),
		))
	}

	markup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	var message string
	if action == "edit" {
		message = "📝 پشتیبان مورد نظر برای ویرایش را انتخاب کنید:"
	} else {
		message = "🗑 پشتیبان مورد نظر برای حذف را انتخاب کنید:"
	}

	msg := tgbotapi.NewMessage(telegramID, message)
	msg.ReplyMarkup = markup
	h.bot.Send(msg)
}

func (h *AdminHandler) sendVideoSettings(telegramID int64) {
	currentMessageID := h.configService.GetCurrentVideoMessageID()
	if currentMessageID == 0 {
		currentMessageID = 2 // Default value
	}

	message := fmt.Sprintf(`🎬 تنظیمات ویدیو

📺 کانال: %d
🆔 شماره پست فعلی: %d

برای تغییر شماره پست، عدد جدید را ارسال کنید:`,
		h.config.Telegram.ChannelID,
		currentMessageID)

	h.adminStates[telegramID] = "change_video_message_id"
	h.sendMessage(telegramID, message)
}

func (h *AdminHandler) handleAdminState(message *tgbotapi.Message, telegramID int64, state string) {
	switch state {
	case "add_support_name":
		h.adminStates[telegramID] = "add_support_username:" + message.Text
		h.sendMessage(telegramID, "نام کاربری پشتیبان را وارد کنید (مثال: @username):")

	case "change_video_message_id":
		if messageID, err := strconv.Atoi(message.Text); err == nil && messageID > 0 {
			// Update config and save to file
			err := h.configService.UpdateVideoMessageID(messageID)
			if err != nil {
				log.Printf("Error updating video message ID: %v", err)
				h.sendMessage(telegramID, "❌ خطا در ذخیره تنظیمات")
			} else {
				h.sendMessage(telegramID, fmt.Sprintf("✅ شماره پست ویدیو به %d تغییر یافت و ذخیره شد", messageID))
			}
		} else {
			h.sendMessage(telegramID, "❌ شماره پست نامعتبر است. لطفاً عدد صحیح وارد کنید:")
			return
		}
		delete(h.adminStates, telegramID)
		h.sendAdminMainMenu(telegramID)

	default:
		if strings.HasPrefix(state, "add_support_username:") {
			name := strings.TrimPrefix(state, "add_support_username:")
			username := message.Text

			// Use default photo URL
			photoURL := "https://d1uuxsymbea74i.cloudfront.net/images/cms/1_6_passport_photo_young_female_9061ba5533.jpg"

			err := h.adminPanelService.CreateSupportStaff(name, username, photoURL)
			if err != nil {
				log.Printf("Error creating support staff: %v", err)
				h.sendMessage(telegramID, "خطا در ایجاد پشتیبان")
			} else {
				h.sendMessage(telegramID, fmt.Sprintf("✅ پشتیبان %s با موفقیت اضافه شد", name))
			}

			delete(h.adminStates, telegramID)
			h.sendSupportManagementMenu(telegramID)
		}
	}
}

func (h *AdminHandler) handleAdminCallback(callback *tgbotapi.CallbackQuery, telegramID int64) {
	data := callback.Data

	if strings.HasPrefix(data, "edit_") {
		h.handleEditSupport(callback, telegramID)
	} else if strings.HasPrefix(data, "delete_") {
		h.handleDeleteSupport(callback, telegramID)
	} else if strings.HasPrefix(data, "toggle_") {
		h.handleToggleSupport(callback, telegramID)
	}

	// Answer callback query
	h.bot.Request(tgbotapi.NewCallback(callback.ID, ""))
}

func (h *AdminHandler) handleEditSupport(callback *tgbotapi.CallbackQuery, telegramID int64) {
	idStr := strings.TrimPrefix(callback.Data, "edit_")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendMessage(telegramID, "خطا در شناسایی پشتیبان")
		return
	}

	// Create inline keyboard for edit options
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🔄 تغییر وضعیت", fmt.Sprintf("toggle_%d", id)),
		),
	)

	msg := tgbotapi.NewMessage(telegramID, "گزینه مورد نظر را انتخاب کنید:")
	msg.ReplyMarkup = keyboard
	h.bot.Send(msg)
}

func (h *AdminHandler) handleDeleteSupport(callback *tgbotapi.CallbackQuery, telegramID int64) {
	idStr := strings.TrimPrefix(callback.Data, "delete_")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendMessage(telegramID, "خطا در شناسایی پشتیبان")
		return
	}

	err = h.adminPanelService.DeleteSupportStaff(uint(id))
	if err != nil {
		log.Printf("Error deleting support staff: %v", err)
		h.sendMessage(telegramID, "خطا در حذف پشتیبان")
	} else {
		h.sendMessage(telegramID, "✅ پشتیبان با موفقیت حذف شد")
	}
}

func (h *AdminHandler) handleToggleSupport(callback *tgbotapi.CallbackQuery, telegramID int64) {
	idStr := strings.TrimPrefix(callback.Data, "toggle_")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendMessage(telegramID, "خطا در شناسایی پشتیبان")
		return
	}

	err = h.adminPanelService.ToggleSupportStaffStatus(uint(id))
	if err != nil {
		log.Printf("Error toggling support staff status: %v", err)
		h.sendMessage(telegramID, "خطا در تغییر وضعیت پشتیبان")
	} else {
		h.sendMessage(telegramID, "✅ وضعیت پشتیبان تغییر یافت")
	}
}

func (h *AdminHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending admin message: %v", err)
	}
}
