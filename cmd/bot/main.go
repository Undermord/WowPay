package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tgwow/internal/config"
	"tgwow/internal/handlers"
	"tgwow/internal/logger"
	"tgwow/internal/storage"
)

// waitForDB пытается подключиться к БД с retry логикой
// Использует экспоненциальный backoff: 1s, 2s, 3s, 4s, 5s
func waitForDB(databaseURL string, maxRetries int) (*storage.PostgresStorage, error) {
	var db *storage.PostgresStorage
	var err error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		db, err = storage.NewPostgresStorage(ctx, databaseURL)
		cancel()

		if err == nil {
			log.Println("Database connected successfully")
			return db, nil
		}

		if attempt < maxRetries {
			waitTime := time.Duration(attempt) * time.Second
			log.Printf("Database not ready (attempt %d/%d): %v. Retrying in %v...",
				attempt, maxRetries, err, waitTime)
			time.Sleep(waitTime)
		}
	}

	return nil, fmt.Errorf("failed to connect to database after %d attempts: %w", maxRetries, err)
}

// setupBotCommands configures the bot command menu for users and admins
func setupBotCommands(bot *tgbotapi.BotAPI, adminChatIDs []int64) error {
	// Commands for all users
	userCommands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начать работу с ботом"},
		{Command: "products", Description: "Посмотреть каталог подписок"},
		{Command: "my_orders", Description: "Мои заказы"},
	}

	// Set commands for all users (default scope)
	defaultConfig := tgbotapi.NewSetMyCommands(userCommands...)
	if _, err := bot.Request(defaultConfig); err != nil {
		return err
	}

	// Commands for admin (includes admin panel)
	adminCommands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начать работу с ботом"},
		{Command: "products", Description: "Посмотреть каталог подписок"},
		{Command: "my_orders", Description: "Мои заказы"},
		{Command: "admin", Description: "Админ-панель"},
	}

	// Set commands for each admin
	for _, adminID := range adminChatIDs {
		adminScope := tgbotapi.NewBotCommandScopeChat(adminID)
		adminConfig := tgbotapi.NewSetMyCommandsWithScope(adminScope, adminCommands...)
		if _, err := bot.Request(adminConfig); err != nil {
			log.Printf("Warning: Failed to set commands for admin %d: %v", adminID, err)
		}
	}

	return nil
}

func main() {
	// Инициализируем логирование в файл + stdout
	if err := logger.Init(); err != nil {
		log.Printf("Warning: Failed to init file logging: %v", err)
		log.Println("Continuing with stdout only...")
	} else {
		defer logger.Close()
		// Проверяем ротацию при старте (без фоновой проверки)
		if err := logger.CheckAndRotate(); err != nil {
			log.Printf("Warning: Failed to rotate logs: %v", err)
		}
	}

	log.Println("Starting WoW Subscription Bot...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Wait for database to be ready with retry logic
	db, err := waitForDB(cfg.DatabaseURL, 5)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	bot, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	bot.Debug = false
	log.Printf("Authorized as @%s", bot.Self.UserName)

	// Set bot commands for menu
	if err := setupBotCommands(bot, cfg.AdminChatIDs); err != nil {
		log.Printf("Warning: Failed to set bot commands: %v", err)
	} else {
		log.Println("Bot commands configured successfully")
	}

	h := handlers.NewHandler(bot, db, cfg.AdminChatIDs, cfg.PaymentCardNumber)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping bot...")
		bot.StopReceivingUpdates()
		os.Exit(0)
	}()

	log.Println("Bot is running. Press Ctrl+C to stop.")

	for update := range updates {
		log.Printf("Received update: %+v", update)

		if update.Message != nil {
			h.HandleMessage(update.Message)
		} else if update.CallbackQuery != nil {
			h.HandleCallback(update.CallbackQuery)
		}
	}
}
