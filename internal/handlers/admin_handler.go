package handlers

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"telegram-bot/internal/config"
	"telegram-bot/internal/keyboards"
	"telegram-bot/internal/models"
	"telegram-bot/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type AdminHandler struct {
	bot               *tgbotapi.BotAPI
	adminPanelService *services.AdminPanelService
	configService     *services.ConfigService
	fileService       *services.FileService
	broadcastService  *services.BroadcastService
	config            *config.Config
	adminStates       map[int64]string                   // Track admin states for multi-step operations
	broadcastPreviews map[int64]*models.BroadcastPreview // Track broadcast previews
}

func NewAdminHandler(bot *tgbotapi.BotAPI, adminPanelService *services.AdminPanelService, configService *services.ConfigService, fileService *services.FileService, broadcastService *services.BroadcastService, cfg *config.Config) *AdminHandler {
	return &AdminHandler{
		bot:               bot,
		adminPanelService: adminPanelService,
		configService:     configService,
		fileService:       fileService,
		broadcastService:  broadcastService,
		config:            cfg,
		adminStates:       make(map[int64]string),
		broadcastPreviews: make(map[int64]*models.BroadcastPreview),
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

	// Handle states for multi-step operations first (including broadcast states)
	if state, exists := h.adminStates[telegramID]; exists {
		// Check if it's a broadcast state
		if strings.HasPrefix(state, "broadcast_") {
			h.handleBroadcastContent(message, telegramID, state)
			return
		}
		// Check if it's a photo upload state for support staff editing
		if strings.HasPrefix(state, "edit_support_photo_upload:") {
			h.handlePhotoUpload(message, telegramID)
			return
		}
		h.handleAdminState(message, telegramID, state)
		return
	}

	// Handle media uploads (only if not in any state)
	if message.Photo != nil {
		h.handlePhotoUpload(message, telegramID)
		return
	}

	// Handle video uploads for broadcast
	if message.Video != nil || message.VideoNote != nil || (message.Document != nil && strings.HasPrefix(message.Document.MimeType, "video/")) {
		// Check if we're in a broadcast state
		if state, exists := h.adminStates[telegramID]; exists && strings.HasPrefix(state, "broadcast_") {
			h.handleBroadcastContent(message, telegramID, state)
			return
		}
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
	case "📢 ارسال پیام همگانی":
		delete(h.adminStates, telegramID)
		h.sendBroadcastMainMenu(telegramID)
		return
	case "📝 ارسال متن":
		h.startBroadcastText(telegramID)
		return
	case "📷 ارسال عکس":
		h.startBroadcastPhoto(telegramID)
		return
	case "🎥 ارسال ویدیو":
		h.startBroadcastVideo(telegramID)
		return
	case "📄 ارسال فایل":
		h.startBroadcastDocument(telegramID)
		return
	case "🎵 ارسال صدا":
		h.startBroadcastAudio(telegramID)
		return
	case "🎤 ارسال ویس":
		h.startBroadcastVoice(telegramID)
		return
	case "😀 ارسال استیکر":
		h.startBroadcastSticker(telegramID)
		return
	case "🎬 ارسال انیمیشن":
		h.startBroadcastAnimation(telegramID)
		return
	case "📋 تاریخچه پیام‌ها":
		h.showBroadcastHistory(telegramID)
		return
	case "📤 آپلود عکس جدید":
		h.handlePhotoUploadRequest(telegramID)
		return
	case "🔗 وارد کردن لینک":
		h.handlePhotoURLRequest(telegramID)
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

	if data == "confirm_broadcast" {
		h.confirmBroadcast(telegramID)
	} else if data == "cancel_broadcast" {
		h.cancelBroadcast(telegramID)
	} else if strings.HasPrefix(data, "edit_name_") {
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

// Broadcast handling methods

func (h *AdminHandler) sendBroadcastMainMenu(telegramID int64) {
	msg := tgbotapi.NewMessage(telegramID, "📢 سیستم ارسال پیام همگانی\n\nیکی از گزینه‌های زیر را انتخاب کنید:")
	msg.ReplyMarkup = keyboards.BroadcastMainKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) startBroadcastText(telegramID int64) {
	h.adminStates[telegramID] = "broadcast_text"
	h.broadcastPreviews[telegramID] = &models.BroadcastPreview{
		ContentType: "text",
	}

	msg := tgbotapi.NewMessage(telegramID, "📝 متن پیام همگانی را وارد کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) startBroadcastPhoto(telegramID int64) {
	h.adminStates[telegramID] = "broadcast_photo"
	h.broadcastPreviews[telegramID] = &models.BroadcastPreview{
		ContentType: "photo",
	}

	msg := tgbotapi.NewMessage(telegramID, "📷 عکس پیام همگانی را ارسال کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) startBroadcastVideo(telegramID int64) {
	h.adminStates[telegramID] = "broadcast_video"
	h.broadcastPreviews[telegramID] = &models.BroadcastPreview{
		ContentType: "video",
	}

	msg := tgbotapi.NewMessage(telegramID, `🎥 ویدیو پیام همگانی را ارسال کنید:

📱 انواع ویدیو قابل قبول:
• ویدیو عادی (Video)
• ویدیو دایره‌ای (Video Note)
• فایل ویدیو (Document)

💡 برای لغو، روی ❌ لغو عملیات کلیک کنید`)
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) startBroadcastDocument(telegramID int64) {
	h.adminStates[telegramID] = "broadcast_document"
	h.broadcastPreviews[telegramID] = &models.BroadcastPreview{
		ContentType: "document",
	}

	msg := tgbotapi.NewMessage(telegramID, "📄 فایل پیام همگانی را ارسال کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) startBroadcastAudio(telegramID int64) {
	h.adminStates[telegramID] = "broadcast_audio"
	h.broadcastPreviews[telegramID] = &models.BroadcastPreview{
		ContentType: "audio",
	}

	msg := tgbotapi.NewMessage(telegramID, "🎵 فایل صوتی پیام همگانی را ارسال کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) startBroadcastVoice(telegramID int64) {
	h.adminStates[telegramID] = "broadcast_voice"
	h.broadcastPreviews[telegramID] = &models.BroadcastPreview{
		ContentType: "voice",
	}

	msg := tgbotapi.NewMessage(telegramID, "🎤 پیام صوتی همگانی را ارسال کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) startBroadcastSticker(telegramID int64) {
	h.adminStates[telegramID] = "broadcast_sticker"
	h.broadcastPreviews[telegramID] = &models.BroadcastPreview{
		ContentType: "sticker",
	}

	msg := tgbotapi.NewMessage(telegramID, "😀 استیکر پیام همگانی را ارسال کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) startBroadcastAnimation(telegramID int64) {
	h.adminStates[telegramID] = "broadcast_animation"
	h.broadcastPreviews[telegramID] = &models.BroadcastPreview{
		ContentType: "animation",
	}

	msg := tgbotapi.NewMessage(telegramID, "🎬 انیمیشن (GIF) پیام همگانی را ارسال کنید:\n\n💡 برای لغو، روی ❌ لغو عملیات کلیک کنید")
	msg.ReplyMarkup = keyboards.CancelOperationKeyboard()
	h.bot.Send(msg)
}

func (h *AdminHandler) showBroadcastHistory(telegramID int64) {
	broadcasts, err := h.broadcastService.GetBroadcastHistory(10, 0)
	if err != nil {
		log.Printf("Error getting broadcast history: %v", err)
		h.sendMessage(telegramID, "خطا در دریافت تاریخچه پیام‌ها")
		return
	}

	if len(broadcasts) == 0 {
		h.sendMessage(telegramID, "📋 هیچ پیام همگانی ارسال نشده است")
		return
	}

	message := "📋 تاریخچه پیام‌های همگانی:\n\n"
	for i, broadcast := range broadcasts {
		status := "⏳ در انتظار"
		if broadcast.Status == "sent" {
			status = "✅ ارسال شده"
		} else if broadcast.Status == "sending" {
			status = "📤 در حال ارسال"
		} else if broadcast.Status == "failed" {
			status = "❌ ناموفق"
		}

		message += fmt.Sprintf("%d. %s - %s\n", i+1, broadcast.ContentType, status)
		message += fmt.Sprintf("   📅 %s\n", broadcast.CreatedAt.Format("2006-01-02 15:04"))
		message += fmt.Sprintf("   📊 ارسال: %d | ناموفق: %d\n\n", broadcast.SentCount, broadcast.FailedCount)
	}

	h.sendMessage(telegramID, message)
}

func (h *AdminHandler) handleBroadcastContent(message *tgbotapi.Message, telegramID int64, state string) {
	preview, exists := h.broadcastPreviews[telegramID]
	if !exists {
		h.sendMessage(telegramID, "خطا در پردازش پیام. لطفاً دوباره شروع کنید.")
		return
	}

	switch state {
	case "broadcast_text":
		preview.Text = message.Text
		h.showBroadcastPreview(telegramID, preview)

	case "broadcast_photo":
		if message.Photo != nil {
			largestPhoto := message.Photo[len(message.Photo)-1]
			preview.FileID = largestPhoto.FileID
			preview.HasFile = true
		}
		h.adminStates[telegramID] = "broadcast_photo_caption"
		msg := tgbotapi.NewMessage(telegramID, "📝 کپشن عکس را وارد کنید (اختیاری):\n\n💡 برای رد کردن کپشن، روی ⏭ رد کردن کپشن کلیک کنید")
		msg.ReplyMarkup = keyboards.SkipCaptionKeyboard()
		h.bot.Send(msg)

	case "broadcast_video":
		// Handle different types of video content
		if message.Video != nil {
			preview.FileID = message.Video.FileID
			preview.HasFile = true
			log.Printf("Video received: FileID=%s, Duration=%d, Width=%d, Height=%d",
				message.Video.FileID, message.Video.Duration, message.Video.Width, message.Video.Height)
		} else if message.VideoNote != nil {
			// VideoNote (circular video) - treat as video
			preview.FileID = message.VideoNote.FileID
			preview.HasFile = true
			log.Printf("VideoNote received: FileID=%s, Duration=%d, Length=%d",
				message.VideoNote.FileID, message.VideoNote.Duration, message.VideoNote.Length)
		} else if message.Document != nil {
			// Check if document is a video file
			if strings.HasPrefix(message.Document.MimeType, "video/") {
				preview.FileID = message.Document.FileID
				preview.HasFile = true
				log.Printf("Video Document received: FileID=%s, MimeType=%s, FileName=%s",
					message.Document.FileID, message.Document.MimeType, message.Document.FileName)
			} else {
				h.sendMessage(telegramID, "❌ فایل ارسال شده ویدیو نیست. لطفاً یک ویدیو ارسال کنید.")
				return
			}
		} else {
			h.sendMessage(telegramID, "❌ ویدیو دریافت نشد. لطفاً یک ویدیو ارسال کنید.")
			return
		}

		h.adminStates[telegramID] = "broadcast_video_caption"
		msg := tgbotapi.NewMessage(telegramID, "📝 کپشن ویدیو را وارد کنید (اختیاری):\n\n💡 برای رد کردن کپشن، روی ⏭ رد کردن کپشن کلیک کنید")
		msg.ReplyMarkup = keyboards.SkipCaptionKeyboard()
		h.bot.Send(msg)

	case "broadcast_document":
		if message.Document != nil {
			preview.FileID = message.Document.FileID
			preview.HasFile = true
		}
		h.adminStates[telegramID] = "broadcast_document_caption"
		msg := tgbotapi.NewMessage(telegramID, "📝 کپشن فایل را وارد کنید (اختیاری):\n\n💡 برای رد کردن کپشن، روی ⏭ رد کردن کپشن کلیک کنید")
		msg.ReplyMarkup = keyboards.SkipCaptionKeyboard()
		h.bot.Send(msg)

	case "broadcast_audio":
		if message.Audio != nil {
			preview.FileID = message.Audio.FileID
			preview.HasFile = true
		}
		h.adminStates[telegramID] = "broadcast_audio_caption"
		msg := tgbotapi.NewMessage(telegramID, "📝 کپشن فایل صوتی را وارد کنید (اختیاری):\n\n💡 برای رد کردن کپشن، روی ⏭ رد کردن کپشن کلیک کنید")
		msg.ReplyMarkup = keyboards.SkipCaptionKeyboard()
		h.bot.Send(msg)

	case "broadcast_voice":
		if message.Voice != nil {
			preview.FileID = message.Voice.FileID
			preview.HasFile = true
		}
		h.showBroadcastPreview(telegramID, preview)

	case "broadcast_sticker":
		if message.Sticker != nil {
			preview.FileID = message.Sticker.FileID
			preview.HasFile = true
		}
		h.showBroadcastPreview(telegramID, preview)

	case "broadcast_animation":
		if message.Animation != nil {
			preview.FileID = message.Animation.FileID
			preview.HasFile = true
		}
		h.adminStates[telegramID] = "broadcast_animation_caption"
		msg := tgbotapi.NewMessage(telegramID, "📝 کپشن انیمیشن را وارد کنید (اختیاری):\n\n💡 برای رد کردن کپشن، روی ⏭ رد کردن کپشن کلیک کنید")
		msg.ReplyMarkup = keyboards.SkipCaptionKeyboard()
		h.bot.Send(msg)

	case "broadcast_photo_caption", "broadcast_video_caption", "broadcast_document_caption", "broadcast_audio_caption", "broadcast_animation_caption":
		// Check if user wants to skip caption or cancel operation
		if message.Text == "❌ لغو عملیات" {
			preview.Text = "" // Clear any existing text
		} else if message.Text == "⏭ رد کردن کپشن" {
			preview.Text = "" // Skip caption
		} else {
			// Trim whitespace and check if text is empty
			trimmedText := strings.TrimSpace(message.Text)
			if trimmedText == "" {
				preview.Text = "" // No caption
			} else {
				preview.Text = trimmedText
			}
		}
		h.showBroadcastPreview(telegramID, preview)
	}
}

func (h *AdminHandler) showBroadcastPreview(telegramID int64, preview *models.BroadcastPreview) {
	// Get user count
	userCount, err := h.broadcastService.GetUserCount()
	if err != nil {
		log.Printf("Error getting user count: %v", err)
		userCount = 0
	}

	message := "📢 پیش‌نمایش پیام همگانی\n\n"
	message += fmt.Sprintf("📊 تعداد کاربران: %d\n", userCount)
	message += fmt.Sprintf("📝 نوع محتوا: %s\n\n", preview.ContentType)

	if preview.HasFile {
		message += "📎 فایل ضمیمه شده است"
		if preview.Text != "" {
			message += "\n📄 کپشن:\n" + preview.Text
		} else {
			message += " (بدون کپشن)"
		}
		message += "\n\n"
	}

	if preview.Text != "" && !preview.HasFile {
		message += fmt.Sprintf("📄 متن:\n%s\n\n", preview.Text)
	}

	message += "⚠️ آیا می‌خواهید این پیام به همه کاربران ارسال شود؟"

	msg := tgbotapi.NewMessage(telegramID, message)
	msg.ReplyMarkup = keyboards.BroadcastConfirmationKeyboard()
	h.bot.Send(msg)

	// Store preview for confirmation
	h.broadcastPreviews[telegramID] = preview
}

func (h *AdminHandler) confirmBroadcast(telegramID int64) {
	preview, exists := h.broadcastPreviews[telegramID]
	if !exists {
		h.sendMessage(telegramID, "خطا در پردازش پیام. لطفاً دوباره شروع کنید.")
		return
	}

	// Create broadcast message
	broadcast, err := h.broadcastService.CreateBroadcast(
		telegramID,
		preview.ContentType,
		preview.Text,
		preview.FileID,
		preview.FileURL,
	)
	if err != nil {
		log.Printf("Error creating broadcast: %v", err)
		h.sendMessage(telegramID, "خطا در ایجاد پیام همگانی")
		return
	}

	// Send broadcast
	h.sendMessage(telegramID, "📤 در حال ارسال پیام همگانی...")

	go func() {
		err := h.broadcastService.SendBroadcast(broadcast.ID)
		if err != nil {
			log.Printf("Error sending broadcast: %v", err)
			h.sendMessage(telegramID, "❌ خطا در ارسال پیام همگانی")
		} else {
			// Get final stats
			stats, _ := h.broadcastService.GetBroadcastStats(broadcast.ID)
			message := fmt.Sprintf("✅ پیام همگانی با موفقیت ارسال شد!\n\n📊 آمار:\n✅ ارسال شده: %d\n❌ ناموفق: %d", stats["sent"], stats["failed"])
			h.sendMessage(telegramID, message)
		}
	}()

	// Clean up
	delete(h.broadcastPreviews, telegramID)
	delete(h.adminStates, telegramID)
	h.sendBroadcastMainMenu(telegramID)
}

func (h *AdminHandler) cancelBroadcast(telegramID int64) {
	delete(h.broadcastPreviews, telegramID)
	delete(h.adminStates, telegramID)
	h.sendMessage(telegramID, "❌ ارسال پیام همگانی لغو شد")
	h.sendBroadcastMainMenu(telegramID)
}

// Helper methods for testing and external access
func (h *AdminHandler) StartBroadcastPhoto(telegramID int64) {
	h.startBroadcastPhoto(telegramID)
}

func (h *AdminHandler) GetAdminState(telegramID int64) string {
	if state, exists := h.adminStates[telegramID]; exists {
		return state
	}
	return ""
}

func (h *AdminHandler) GetBroadcastPreview(telegramID int64) *models.BroadcastPreview {
	if preview, exists := h.broadcastPreviews[telegramID]; exists {
		return preview
	}
	return nil
}
