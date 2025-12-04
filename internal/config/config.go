package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	BotToken             string
	AdminChatIDs         []int64 // Список ID администраторов
	DatabaseURL          string
	PaymentProviderToken string // Опционально для Telegram Payments
	PaymentCardNumber    string // Номер карты для оплаты
}

func Load() (*Config, error) {
	// Try to load .env file (optional in Docker)
	_ = godotenv.Load()

	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return nil, fmt.Errorf("BOT_TOKEN is required")
	}

	adminChatIDStr := os.Getenv("ADMIN_CHAT_ID")
	if adminChatIDStr == "" {
		return nil, fmt.Errorf("ADMIN_CHAT_ID is required")
	}

	// Парсим список ID через запятую
	adminChatIDs := []int64{}
	idStrings := strings.Split(adminChatIDStr, ",")
	for _, idStr := range idStrings {
		idStr = strings.TrimSpace(idStr)
		if idStr == "" {
			continue
		}
		adminID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid ADMIN_CHAT_ID '%s': %w", idStr, err)
		}
		adminChatIDs = append(adminChatIDs, adminID)
	}

	if len(adminChatIDs) == 0 {
		return nil, fmt.Errorf("at least one ADMIN_CHAT_ID is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	// Опциональный payment provider token
	paymentToken := os.Getenv("PAYMENT_PROVIDER_TOKEN")

	// Номер карты для оплаты (обязательно)
	paymentCard := os.Getenv("PAYMENT_CARD_NUMBER")
	if paymentCard == "" {
		return nil, fmt.Errorf("PAYMENT_CARD_NUMBER is required")
	}

	return &Config{
		BotToken:             botToken,
		AdminChatIDs:         adminChatIDs,
		DatabaseURL:          databaseURL,
		PaymentProviderToken: paymentToken,
		PaymentCardNumber:    paymentCard,
	}, nil
}
