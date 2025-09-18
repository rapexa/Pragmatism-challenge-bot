package main

import (
	"log"
	"telegram-bot/internal/config"
	"telegram-bot/internal/database"
	"telegram-bot/internal/handlers"
	"telegram-bot/internal/server"
	"telegram-bot/internal/services"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	// Load configuration
	cfg := config.Load()
	log.Println("Configuration loaded successfully")

	// Initialize database
	db := database.New(&cfg.Database)
	defer db.Close()

	// Initialize Telegram bot first (needed for FileService)
	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.Token)
	if err != nil {
		log.Fatalf("Error creating Telegram bot: %v", err)
	}

	bot.Debug = true
	log.Printf("Bot %s started successfully", bot.Self.UserName)

	// Start file server
	fileServer := server.NewFileServer("uploads", cfg.Server.Port)
	fileServer.Start()

	// Initialize services
	userService := services.NewUserService(db)
	supportService := services.NewSupportService(db)
	adminPanelService := services.NewAdminPanelService(db)
	configService := services.NewConfigService(cfg)
	fileService := services.NewFileService(bot, cfg.Server.URL)

	// Initialize test support staff data
	err = supportService.InitializeTestData()
	if err != nil {
		log.Printf("Error initializing test support staff data: %v", err)
	}

	// Initialize default admin
	err = adminPanelService.InitializeDefaultAdmin()
	if err != nil {
		log.Printf("Error initializing default admin: %v", err)
	}

	// Initialize bot handler
	botHandler := handlers.NewBotHandler(bot, userService, supportService, adminPanelService, configService, fileService, cfg)

	// Set up webhook or polling
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	log.Println("Bot is now running...")

	for update := range updates {
		go botHandler.HandleUpdate(update)
	}
}
