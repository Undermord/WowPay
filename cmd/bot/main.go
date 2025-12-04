package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tgwow/internal/config"
	"tgwow/internal/handlers"
	"tgwow/internal/storage"
)

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
	log.Println("Starting WoW Subscription Bot...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Wait for database to be ready
	time.Sleep(3 * time.Second)

	ctx := context.Background()
	db, err := storage.NewPostgresStorage(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("Database connected successfully")

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
