package handlers

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tgwow/internal/models"
)

// newDBContext creates a context with standard timeout for database operations
func (h *Handler) newDBContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), DBContextTimeout)
}

// buildRegionsKeyboard creates keyboard with list of regions and "Change region" button
func (h *Handler) buildRegionsKeyboard(ctx context.Context) (string, [][]tgbotapi.InlineKeyboardButton, error) {
	regions, err := h.storage.ListRegions(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("failed to fetch regions: %w", err)
	}

	text := "🛒 <b>Каталог товаров World of Warcraft</b>\n\nВыберите регион:"
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, r := range regions {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s %s", r.Name, getRegionFlag(r.Code)),
			fmt.Sprintf("%s:%d", CallbackActionRegion, r.ID),
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	// Add "Change region" button
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🔄 Сменить регион", CallbackActionChangeRegion),
	})

	return text, keyboard, nil
}

// buildProductCard creates product card with price, description and buy/back buttons
func (h *Handler) buildProductCard(product *models.Product, backCallback string) (string, tgbotapi.InlineKeyboardMarkup) {
	priceText := ""
	if product.Price > 0 {
		priceText = fmt.Sprintf("💰 <b>Цена:</b> %.2f руб.\n\n", product.Price)
	} else {
		priceText = "💰 <b>Цена:</b> уточняется\n\n"
	}

	text := fmt.Sprintf(
		"🎮 <b>%s</b>\n\n%s📝 <b>Описание:</b>\n%s",
		product.Name, priceText, product.Description,
	)

	var keyboard tgbotapi.InlineKeyboardMarkup
	if product.Price > 0 {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Купить", fmt.Sprintf("%s:%d", CallbackActionBuy, product.ID)),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", backCallback),
			),
		)
	} else {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("◀️ Назад", backCallback),
			),
		)
	}

	return text, keyboard
}

// getUserDisplayName returns user's display name (username or first+last name)
func getUserDisplayName(user *tgbotapi.User) string {
	if user.UserName != "" {
		return "@" + user.UserName
	}
	name := user.FirstName
	if user.LastName != "" {
		name += " " + user.LastName
	}
	return name
}

// getRegionFlag returns flag emoji for region code
func getRegionFlag(code string) string {
	flags := map[string]string{
		"KZ":  "🇰🇿",
		"UA":  "🇺🇦",
		"EU":  "🇪🇺",
		"TUR": "🇹🇷",
	}
	if flag, ok := flags[code]; ok {
		return flag
	}
	return "🌍" // default flag for unknown regions
}

// checkRateLimit проверяет лимит запросов пользователя
func (h *Handler) checkRateLimit(userID int64, chatID int64) bool {
	// Админы не ограничены rate limiting
	if h.isAdmin(userID) {
		return true
	}

	// Для обычных пользователей используем rate limiter
	limiter := h.userLimiter

	// Проверяем лимит
	if !limiter.Allow(userID) {
		// Проверяем, заблокирован ли пользователь теперь
		banned, duration := limiter.IsBanned(userID)
		if banned {
			minutes := int(duration.Minutes())
			seconds := int(duration.Seconds()) % 60

			var timeStr string
			if minutes > 0 {
				timeStr = fmt.Sprintf("%d мин %d сек", minutes, seconds)
			} else {
				timeStr = fmt.Sprintf("%d сек", seconds)
			}

			msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(
				"⏳ <b>Превышен лимит запросов</b>\n\n"+
					"Вы отправляете слишком много сообщений.\n"+
					"Пожалуйста, подождите <b>%s</b> перед следующей попыткой.",
				timeStr,
			))
			msg.ParseMode = "HTML"
			h.bot.Send(msg)
		}
		return false
	}

	return true
}
