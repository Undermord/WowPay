package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleStart обрабатывает команду /start
func (h *Handler) handleStart(msg *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Получаем настройки из БД
	settings, err := h.storage.GetBotSettings(ctx)
	if err != nil {
		log.Printf("Error fetching bot settings: %v", err)
		// Fallback сообщение, если не удалось получить из БД
		text := fmt.Sprintf(
			"👋 Добро пожаловать, %s!\n\n"+
				"🎮 Я бот для продажи игровых подписок World of Warcraft.\n\n"+
				"📋 Доступные команды:\n"+
				"/products - Посмотреть каталог подписок\n"+
				"/my_orders - Просмотреть мои заказы",
			msg.From.FirstName,
		)
		h.sendMessage(msg.Chat.ID, text)
		return
	}

	// Заменяем {name} на имя пользователя
	text := strings.ReplaceAll(settings.WelcomeMessage, "{name}", msg.From.FirstName)

	// Создаём inline-клавиатуру с кнопкой "Далее"
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("➡️ Далее", "show_products"),
		),
	)

	// Отправляем с HTML разметкой
	response := tgbotapi.NewMessage(msg.Chat.ID, text)
	response.ParseMode = "HTML"
	response.DisableWebPagePreview = false
	response.ReplyMarkup = keyboard

	if _, err := h.bot.Send(response); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}
}

// handleProducts обрабатывает команду /products
func (h *Handler) handleProducts(msg *tgbotapi.Message) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	regions, err := h.storage.ListRegions(ctx)
	if err != nil {
		log.Printf("Error fetching regions: %v", err)
		h.sendMessage(msg.Chat.ID, "❌ Ошибка при загрузке каталога. Попробуйте позже.")
		return
	}

	if len(regions) == 0 {
		h.sendMessage(msg.Chat.ID, "📦 Каталог пуст.")
		return
	}

	text := "🛒 <b>Каталог товаров World of Warcraft</b>\n\n" +
		"Выберите регион:"

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, r := range regions {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s %s", r.Name, getRegionFlag(r.Code)),
			fmt.Sprintf("region:%d", r.ID),
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	// Добавляем кнопку "Сменить регион"
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🔄 Сменить регион", "change_region"),
	})

	response := tgbotapi.NewMessage(msg.Chat.ID, text)
	response.ParseMode = "HTML"
	response.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	if _, err := h.bot.Send(response); err != nil {
		log.Printf("Error sending regions: %v", err)
	}
}

// handleMyOrders обрабатывает команду /my_orders
func (h *Handler) handleMyOrders(msg *tgbotapi.Message) {
	ctx, cancel := h.newDBContext()
	defer cancel()

	orders, err := h.storage.GetUserOrders(ctx, msg.From.ID)
	if err != nil {
		log.Printf("Error fetching user orders: %v", err)
		h.sendMessage(msg.Chat.ID, "❌ Ошибка при загрузке заказов.")
		return
	}

	if len(orders) == 0 {
		h.sendMessage(msg.Chat.ID, "📦 У вас пока нет заказов.\n\nИспользуйте /products чтобы посмотреть каталог подписок.")
		return
	}

	// Собираем все ID товаров для batch-загрузки (решение N+1 проблемы)
	productIDs := make([]int, 0, len(orders))
	for _, order := range orders {
		productIDs = append(productIDs, order.ProductID)
	}

	// Загружаем все товары одним запросом
	products, err := h.storage.GetProductsByIDs(ctx, productIDs)
	if err != nil {
		log.Printf("Error fetching products: %v", err)
		h.sendMessage(msg.Chat.ID, "❌ Ошибка при загрузке информации о товарах.")
		return
	}

	text := "📋 <b>Ваши заказы:</b>\n\n"

	// Load Moscow timezone once for all orders
	moscowLocation, _ := time.LoadLocation("Europe/Moscow")

	for i, order := range orders {
		product, exists := products[order.ProductID]
		if !exists {
			continue
		}

		// Convert time to Moscow timezone
		moscowTime := order.CreatedAt.In(moscowLocation)

		text += fmt.Sprintf(
			"%s <b>Заказ №%d</b>\n"+
				"🆔 <code>%s</code>\n"+
				"🎮 %s\n"+
				"💰 %.2f руб.\n"+
				"📊 Статус: %s %s\n"+
				"📅 %s (МСК)\n\n",
			StatusEmojis[order.Status],
			i+1,
			order.OrderID,
			product.Name,
			order.Price,
			StatusEmojis[order.Status],
			StatusTexts[order.Status],
			moscowTime.Format("02.01.2006 15:04"),
		)
	}

	response := tgbotapi.NewMessage(msg.Chat.ID, text)
	response.ParseMode = "HTML"

	if _, err := h.bot.Send(response); err != nil {
		log.Printf("Error sending my orders: %v", err)
	}
}

// handleCancel обрабатывает команду /cancel
func (h *Handler) handleCancel(msg *tgbotapi.Message) {
	h.fsmManager.ClearState(msg.From.ID)
	h.sendMessage(msg.Chat.ID, "❌ Действие отменено")
}

// sendMessage отправляет простое текстовое сообщение
func (h *Handler) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending message: %v", err)
	}
}
