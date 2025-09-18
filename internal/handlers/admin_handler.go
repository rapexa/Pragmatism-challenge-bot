package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-bot/internal/config"
	"telegram-bot/internal/keyboards"
	"telegram-bot/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminHandler struct {
	bot               *tgbotapi.BotAPI
	adminPanelService *services.AdminPanelService
	configService     *services.ConfigService
	fileService       *services.FileService
	config            *config.Config
	adminStates       map[int64]string // Track admin states for multi-step operations
}

func NewAdminHandler(bot *tgbotapi.BotAPI, adminPanelService *services.AdminPanelService, configService *services.ConfigService, fileService *services.FileService, cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		bot:               bot,
		adminPanelService: adminPanelService,
		configService:     configService,
		fileService:       fileService,
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

	// Handle photo uploads
	if message.Photo != nil {
		h.handlePhotoUpload(message, telegramID)
		return
	}

	// Handle navigation buttons that should clear any state
	switch text {
	case "🔙 بازگشت به پنل مدیریت":
		delete(h.adminStates, telegramID)
		h.sendAdminMainMenu(telegramID)
		return
	case "❌ لغو عملیات":
		delete(h.adminStates, telegramID)
		h.sendAdminMainMenu(telegramID)
		return
	case "📊 آمار کاربران":
		delete(h.adminStates, telegramID)
		h.sendUserStats(telegramID)
		return
	case "📋 خروجی اکسل کاربران":
		delete(h.adminStates, telegramID)
		h.exportUsers(telegramID)
		return
	case "👥 مدیریت پشتیبان‌ها":
		delete(h.adminStates, telegramID)
		h.sendSupportManagementMenu(telegramID)
		return
	case "➕ افزودن پشتیبان":
		delete(h.adminStates, telegramID)
		h.startAddSupport(telegramID)
		return
	case "📝 ویرایش پشتیبان":
		delete(h.adminStates, telegramID)
		h.showSupportList(telegramID, "edit")
		return
	case "🗑 حذف پشتیبان":
		delete(h.adminStates, telegramID)
		h.showSupportList(telegramID, "delete")
		return
	case "🎬 تنظیمات ویدیو":
		delete(h.adminStates, telegramID)
		h.sendVideoSettings(telegramID)
		return
	case "🔗 تنظیمات گروه":
		delete(h.adminStates, telegramID)
		h.sendGroupSettings(telegramID)
		return
	case "📤 آپلود عکس جدید":
		h.handlePhotoUploadRequest(telegramID)
		return
	case "🔗 وارد کردن لینک":
		h.handlePhotoURLRequest(telegramID)
		return
	}

	// Handle states for multi-step operations
	if state, exists := h.adminStates[telegramID]; exists {
		h.handleAdminState(message, telegramID, state)
		return
	}

	// Handle start command
	if message.Command() == "start" {
		h.sendAdminMainMenu(telegramID)
		return
	}

	// Default: show main menu
	h.sendAdminMainMenu(telegramID)
}

func (h *AdminHandler) sendAdminMainMenu(telegramID int64) {
	msg := tgbotapi.NewMessage(telegramID, "🔧 پنل مدیریت ربات\n\nیکی از گزینه‌های زیر را انتخاب کنید:")
	msg.ReplyMarkup = keyboards.AdminMainKeyboard()
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
	msg := tgbotapi.NewMessage(telegramID, "👥 مدیریت پشتیبان‌ها\n\nیکی از گزینه‌های زیر را انتخاب کنید:")
	msg.ReplyMarkup = keyboards.SupportManagementKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) startAddSupport(telegramID int64) {
	h.adminStates[telegramID] = "add_support_name"

	msg := tgbotapi.NewMessage(telegramID, "نام پشتیبان جدید را وارد کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
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

برای تغییر شماره پست، عدد جدید را ارسال کنید:

💡 برای لغو، روی ❌ لغو عملیات کلیک کنید`,
		h.config.Telegram.ChannelID,
		currentMessageID)

	h.adminStates[telegramID] = "change_video_message_id"

	msg := tgbotapi.NewMessage(telegramID, message)
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) sendGroupSettings(telegramID int64) {
	currentGroupLink := h.configService.GetCurrentGroupLink()

	message := fmt.Sprintf(`🔗 تنظیمات گروه

🔗 لینک گروه فعلی: %s

برای تغییر لینک گروه، لینک جدید را ارسال کنید:

💡 برای لغو، روی ❌ لغو عملیات کلیک کنید`,
		currentGroupLink)

	h.adminStates[telegramID] = "change_group_link"

	msg := tgbotapi.NewMessage(telegramID, message)
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) handleAdminState(message *tgbotapi.Message, telegramID int64, state string) {
	switch state {
	case "add_support_name":
		h.adminStates[telegramID] = "add_support_username:" + message.Text

		msg := tgbotapi.NewMessage(telegramID, "نام کاربری پشتیبان را وارد کنید (مثال: @username):\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
		msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
		h.bot.Send(msg)
		return

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

	case "change_group_link":
		// Validate group link format
		groupLink := message.Text
		if !strings.HasPrefix(groupLink, "https://t.me/") && !strings.HasPrefix(groupLink, "@") {
			h.sendMessage(telegramID, "❌ فرمت لینک گروه نامعتبر است. لطفاً لینک معتبر وارد کنید (مثال: https://t.me/group_name یا @group_name):")
			return
		}

		err := h.configService.UpdateGroupLink(groupLink)
		if err != nil {
			log.Printf("Error updating group link: %v", err)
			h.sendMessage(telegramID, "❌ خطا در ذخیره لینک گروه")
		} else {
			h.sendMessage(telegramID, fmt.Sprintf("✅ لینک گروه به %s تغییر یافت و ذخیره شد", groupLink))
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
		} else if strings.HasPrefix(state, "edit_support_name:") {
			idStr := strings.TrimPrefix(state, "edit_support_name:")
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				h.sendMessage(telegramID, "خطا در شناسایی پشتیبان")
				return
			}

			newName := message.Text
			err = h.adminPanelService.UpdateSupportStaffField(uint(id), "name", newName)
			if err != nil {
				log.Printf("Error updating support staff name: %v", err)
				h.sendMessage(telegramID, "خطا در ویرایش نام پشتیبان")
			} else {
				h.sendMessage(telegramID, fmt.Sprintf("✅ نام پشتیبان به %s تغییر یافت", newName))
			}

			delete(h.adminStates, telegramID)
			h.sendSupportManagementMenu(telegramID)
		} else if strings.HasPrefix(state, "edit_support_username:") {
			idStr := strings.TrimPrefix(state, "edit_support_username:")
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				h.sendMessage(telegramID, "خطا در شناسایی پشتیبان")
				return
			}

			newUsername := message.Text
			err = h.adminPanelService.UpdateSupportStaffField(uint(id), "username", newUsername)
			if err != nil {
				log.Printf("Error updating support staff username: %v", err)
				h.sendMessage(telegramID, "خطا در ویرایش یوزرنیم پشتیبان")
			} else {
				h.sendMessage(telegramID, fmt.Sprintf("✅ یوزرنیم پشتیبان به %s تغییر یافت", newUsername))
			}

			delete(h.adminStates, telegramID)
			h.sendSupportManagementMenu(telegramID)
		} else if strings.HasPrefix(state, "edit_support_photo_url:") {
			idStr := strings.TrimPrefix(state, "edit_support_photo_url:")
			id, err := strconv.ParseUint(idStr, 10, 32)
			if err != nil {
				h.sendMessage(telegramID, "خطا در شناسایی پشتیبان")
				return
			}

			newPhotoURL := message.Text
			err = h.adminPanelService.UpdateSupportStaffField(uint(id), "photo_url", newPhotoURL)
			if err != nil {
				log.Printf("Error updating support staff photo: %v", err)
				h.sendMessage(telegramID, "خطا در ویرایش عکس پشتیبان")
			} else {
				h.sendMessage(telegramID, "✅ عکس پشتیبان با موفقیت تغییر یافت")
			}

			delete(h.adminStates, telegramID)
			h.sendSupportManagementMenu(telegramID)
		}
	}
}

func (h *AdminHandler) handleAdminCallback(callback *tgbotapi.CallbackQuery, telegramID int64) {
	data := callback.Data

	if strings.HasPrefix(data, "edit_name_") {
		h.handleEditSupportName(callback, telegramID)
	} else if strings.HasPrefix(data, "edit_username_") {
		h.handleEditSupportUsername(callback, telegramID)
	} else if strings.HasPrefix(data, "edit_photo_") {
		h.handleEditSupportPhoto(callback, telegramID)
	} else if strings.HasPrefix(data, "edit_") {
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
			tgbotapi.NewInlineKeyboardButtonData("📝 ویرایش نام", fmt.Sprintf("edit_name_%d", id)),
			tgbotapi.NewInlineKeyboardButtonData("👤 ویرایش یوزرنیم", fmt.Sprintf("edit_username_%d", id)),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("🖼 ویرایش عکس", fmt.Sprintf("edit_photo_%d", id)),
			tgbotapi.NewInlineKeyboardButtonData("🔄 تغییر وضعیت", fmt.Sprintf("toggle_%d", id)),
		),
	)

	msg := tgbotapi.NewMessage(telegramID, "چه چیزی را می‌خواهید ویرایش کنید:")
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

func (h *AdminHandler) handleEditSupportName(callback *tgbotapi.CallbackQuery, telegramID int64) {
	idStr := strings.TrimPrefix(callback.Data, "edit_name_")
	h.adminStates[telegramID] = "edit_support_name:" + idStr

	msg := tgbotapi.NewMessage(telegramID, "نام جدید پشتیبان را وارد کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) handleEditSupportUsername(callback *tgbotapi.CallbackQuery, telegramID int64) {
	idStr := strings.TrimPrefix(callback.Data, "edit_username_")
	h.adminStates[telegramID] = "edit_support_username:" + idStr

	msg := tgbotapi.NewMessage(telegramID, "یوزرنیم جدید پشتیبان را وارد کنید (مثال: @username):\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) handleEditSupportPhoto(callback *tgbotapi.CallbackQuery, telegramID int64) {
	idStr := strings.TrimPrefix(callback.Data, "edit_photo_")
	h.adminStates[telegramID] = "edit_support_photo_method:" + idStr

	msg := tgbotapi.NewMessage(telegramID, "روش تغییر عکس پشتیبان را انتخاب کنید:")
	msg.ReplyMarkup = keyboards.PhotoUploadKeyboard()
	h.bot.Send(msg)
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

func (h *AdminHandler) handlePhotoUploadRequest(telegramID int64) {
	if state, exists := h.adminStates[telegramID]; exists && strings.HasPrefix(state, "edit_support_photo_method:") {
		idStr := strings.TrimPrefix(state, "edit_support_photo_method:")
		h.adminStates[telegramID] = "edit_support_photo_upload:" + idStr

		msg := tgbotapi.NewMessage(telegramID, "عکس جدید پشتیبان را ارسال کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
		msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
		h.bot.Send(msg)
	}
}

func (h *AdminHandler) handlePhotoURLRequest(telegramID int64) {
	if state, exists := h.adminStates[telegramID]; exists && strings.HasPrefix(state, "edit_support_photo_method:") {
		idStr := strings.TrimPrefix(state, "edit_support_photo_method:")
		h.adminStates[telegramID] = "edit_support_photo_url:" + idStr

		msg := tgbotapi.NewMessage(telegramID, "URL عکس جدید پشتیبان را وارد کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
		msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
		h.bot.Send(msg)
	}
}

func (h *AdminHandler) handlePhotoUpload(message *tgbotapi.Message, telegramID int64) {
	state, exists := h.adminStates[telegramID]
	if !exists || !strings.HasPrefix(state, "edit_support_photo_upload:") {
		h.sendMessage(telegramID, "لطفاً ابتدا گزینه ویرایش عکس را انتخاب کنید")
		return
	}

	idStr := strings.TrimPrefix(state, "edit_support_photo_upload:")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.sendMessage(telegramID, "خطا در شناسایی پشتیبان")
		return
	}

	// Get the largest photo size
	photos := message.Photo
	if len(photos) == 0 {
		h.sendMessage(telegramID, "خطا در دریافت عکس")
		return
	}

	largestPhoto := photos[len(photos)-1] // Last photo is usually the largest

	// Download and save photo
	h.sendMessage(telegramID, "در حال ذخیره عکس...")

	localPath, err := h.fileService.DownloadPhoto(largestPhoto.FileID)
	if err != nil {
		log.Printf("Error downloading photo: %v", err)
		h.sendMessage(telegramID, "خطا در ذخیره عکس")
		return
	}

	// Update support staff photo URL with local path
	err = h.adminPanelService.UpdateSupportStaffField(uint(id), "photo_url", localPath)
	if err != nil {
		log.Printf("Error updating support staff photo: %v", err)
		h.sendMessage(telegramID, "خطا در ویرایش عکس پشتیبان")
		return
	}

	h.sendMessage(telegramID, "✅ عکس پشتیبان با موفقیت آپلود و ذخیره شد")

	delete(h.adminStates, telegramID)
	h.sendSupportManagementMenu(telegramID)
}

func (h *AdminHandler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := h.bot.Send(msg)
	if err != nil {
		log.Printf("Error sending admin message: %v", err)
	}
}
