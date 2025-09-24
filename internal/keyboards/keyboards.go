package keyboards

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Admin keyboards

// AdminMainKeyboard returns the main admin panel keyboard
func AdminMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📊 آمار کاربران"),
			tgbotapi.NewKeyboardButton("📋 خروجی اکسل کاربران"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("👥 مدیریت پشتیبان‌ها"),
			tgbotapi.NewKeyboardButton("📢 ارسال پیام همگانی"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎬 تنظیمات ویدیو"),
			tgbotapi.NewKeyboardButton("🔗 تنظیمات گروه"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// SupportManagementKeyboard returns the support management keyboard
func SupportManagementKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("➕ افزودن پشتیبان"),
			tgbotapi.NewKeyboardButton("📝 ویرایش پشتیبان"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🗑 حذف پشتیبان"),
			tgbotapi.NewKeyboardButton("🔙 بازگشت به پنل مدیریت"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// BackToAdminKeyboard returns a keyboard with back to admin panel button
func BackToAdminKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🔙 بازگشت به پنل مدیریت"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// CancelOperationKeyboard returns a keyboard with cancel button
func CancelOperationKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ لغو عملیات"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// User keyboards

// PhoneRequestKeyboard returns keyboard for requesting phone number
func PhoneRequestKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButtonContact("📱 ارسال شماره تماس"),
		),
	)
	keyboard.OneTimeKeyboard = true
	keyboard.ResizeKeyboard = true
	return keyboard
}

// PhotoUploadKeyboard returns keyboard for photo upload options
func PhotoUploadKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📤 آپلود عکس جدید"),
			tgbotapi.NewKeyboardButton("🔗 وارد کردن لینک"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ لغو عملیات"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// RemoveKeyboard returns an empty keyboard to remove current keyboard
func RemoveKeyboard() tgbotapi.ReplyKeyboardRemove {
	return tgbotapi.NewRemoveKeyboard(true)
}

// Broadcast keyboards

// BroadcastMainKeyboard returns the broadcast main menu keyboard
func BroadcastMainKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📝 ارسال متن"),
			tgbotapi.NewKeyboardButton("📷 ارسال عکس"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎥 ارسال ویدیو"),
			tgbotapi.NewKeyboardButton("📄 ارسال فایل"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎵 ارسال صدا"),
			tgbotapi.NewKeyboardButton("🎤 ارسال ویس"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("😀 ارسال استیکر"),
			tgbotapi.NewKeyboardButton("🎬 ارسال انیمیشن"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📋 تاریخچه پیام‌ها"),
			tgbotapi.NewKeyboardButton("🔙 بازگشت به پنل مدیریت"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}

// BroadcastConfirmationKeyboard returns the confirmation keyboard for broadcast
func BroadcastConfirmationKeyboard() tgbotapi.InlineKeyboardMarkup {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ تأیید و ارسال", "confirm_broadcast"),
			tgbotapi.NewInlineKeyboardButtonData("❌ لغو", "cancel_broadcast"),
		),
	)
	return keyboard
}

// BroadcastContentTypeKeyboard returns keyboard for selecting content type
func BroadcastContentTypeKeyboard() tgbotapi.ReplyKeyboardMarkup {
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("📝 فقط متن"),
			tgbotapi.NewKeyboardButton("📷 عکس + متن"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎥 ویدیو + متن"),
			tgbotapi.NewKeyboardButton("📄 فایل + متن"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("🎵 صدا + متن"),
			tgbotapi.NewKeyboardButton("🎤 ویس"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("😀 استیکر"),
			tgbotapi.NewKeyboardButton("🎬 انیمیشن + متن"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("❌ لغو عملیات"),
		),
	)
	keyboard.ResizeKeyboard = true
	return keyboard
}
