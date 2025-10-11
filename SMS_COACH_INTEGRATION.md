# ادغام نام پشتیبان در SMS فعال‌سازی

## 🎯 هدف

اضافه کردن نام پشتیبان کاربر به SMS فعال‌سازی به جای کلمه "coach".

## 📋 تغییرات انجام شده

### 1. اضافه کردن متد جدید در SMS Service

```go
// SendRegistrationSMSWithCoach sends registration success SMS with coach name
func (s *SMSService) SendRegistrationSMSWithCoach(phone, firstName, coachName string) error {
	params := map[string]string{
		"name":   firstName,
		"coach":  coachName,
	}
	return s.SendSMS(phone, params, "registration")
}
```

### 2. آپدیت Bot Handler

```go
// Send SMS notification with coach name
go func() {
	smsErr := h.smsService.SendRegistrationSMSWithCoach(state.PhoneNumber, state.FirstName, support.Name)
	if smsErr != nil {
		log.Printf("Error sending registration SMS: %v", smsErr)
	} else {
		log.Printf("Registration SMS sent successfully to %s with coach %s", state.PhoneNumber, support.Name)
	}
}()
```

## 🔧 نحوه کارکرد

### قبل از تغییر:
```
SMS Pattern Variables:
- name: "احمد احمدی"
- coach: "coach" (کلمه ثابت)
```

### بعد از تغییر:
```
SMS Pattern Variables:
- name: "احمد احمدی"
- coach: "خانم فاطمه تقی زاده" (نام واقعی پشتیبان)
```

## 📱 مثال SMS

### قبل:
```
سلام احمد احمدی عزیز!
ثبت نام شما با موفقیت انجام شد.
پشتیبان شما: coach
```

### بعد:
```
سلام احمد احمدی عزیز!
ثبت نام شما با موفقیت انجام شد.
پشتیبان شما: خانم فاطمه تقی زاده
```

## 🛠️ تنظیمات IPPanel

### Pattern Variables
در پنل IPPanel، pattern "registration" باید شامل متغیرهای زیر باشد:
- `{name}`: نام کاربر
- `{coach}`: نام پشتیبان

### مثال Pattern:
```
سلام {name} عزیز!
ثبت نام شما با موفقیت انجام شد.
پشتیبان شما: {coach}
```

## 📁 فایل‌های تغییر یافته

### فایل‌های آپدیت شده:
- `internal/services/sms_service.go` - اضافه کردن متد جدید
- `internal/handlers/bot_handler.go` - استفاده از متد جدید

### فایل‌های جدید:
- `SMS_COACH_INTEGRATION.md` - مستندات

## 🔍 تست

### سناریوهای تست:
1. **کاربر جدید**: باید SMS با نام پشتیبان دریافت کند
2. **کاربر قدیمی**: باید SMS با نام پشتیبان دریافت کند
3. **عدم وجود پشتیبان**: باید SMS بدون نام پشتیبان ارسال شود

### لاگ‌های مورد انتظار:
```
Registration SMS sent successfully to 09123456789 with coach خانم فاطمه تقی زاده
```

## ⚠️ نکات مهم

1. **Pattern IPPanel**: باید متغیر `{coach}` در pattern تعریف شده باشد
2. **Fallback**: اگر پشتیبان یافت نشود، SMS بدون نام پشتیبان ارسال می‌شود
3. **Performance**: SMS در goroutine ارسال می‌شود تا blocking نباشد

## 🎉 نتیجه

حالا SMS فعال‌سازی:
- ✅ نام واقعی پشتیبان را نمایش می‌دهد
- ✅ شخصی‌سازی شده است
- ✅ تجربه کاربری بهتری دارد
- ✅ قابل تنظیم و توسعه است
