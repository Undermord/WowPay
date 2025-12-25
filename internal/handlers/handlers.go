package handlers

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tgwow/internal/fsm"
	"tgwow/internal/ratelimit"
	"tgwow/internal/storage"
)

// Handler управляет обработкой сообщений и callback'ов Telegram бота
type Handler struct {
	bot               *tgbotapi.BotAPI
	storage           *storage.PostgresStorage
	adminChatIDs      []int64            // Список ID администраторов
	fsmManager        *fsm.Manager       // Менеджер FSM состояний
	paymentCardNumber string             // Номер карты для оплаты
	userLimiter       *ratelimit.Limiter // Rate limiter для пользователей
	adminLimiter      *ratelimit.Limiter // Rate limiter для админов
}

// NewHandler создает новый Handler
func NewHandler(bot *tgbotapi.BotAPI, storage *storage.PostgresStorage, adminChatIDs []int64, paymentCardNumber string) *Handler {
	return &Handler{
		bot:               bot,
		storage:           storage,
		adminChatIDs:      adminChatIDs,
		fsmManager:        fsm.NewManager(),
		paymentCardNumber: paymentCardNumber,
		userLimiter:       ratelimit.NewLimiter(ratelimit.DefaultConfig()),
		adminLimiter:      ratelimit.NewLimiter(ratelimit.AdminConfig()),
	}
}

// Shutdown корректно останавливает все фоновые процессы Handler
func (h *Handler) Shutdown() {
	log.Println("Shutting down handler resources...")

	// Останавливаем FSM Manager (cleanup goroutine)
	if h.fsmManager != nil {
		h.fsmManager.Stop()
	}

	// Останавливаем rate limiters (cleanup goroutines)
	if h.userLimiter != nil {
		h.userLimiter.Stop()
	}
	if h.adminLimiter != nil {
		h.adminLimiter.Stop()
	}

	log.Println("Handler resources stopped")
}

// HandleMessage обрабатывает входящие сообщения
func (h *Handler) HandleMessage(msg *tgbotapi.Message) {
	log.Printf("Message from %s: %s", msg.From.UserName, msg.Text)

	// Проверка rate limit
	if !h.checkRateLimit(msg.From.ID, msg.Chat.ID) {
		return
	}

	// Сохраняем/обновляем пользователя в БД для рассылок
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		username := msg.From.UserName
		if username == "" {
			username = msg.From.FirstName
		}

		if err := h.storage.UpsertUser(ctx, msg.From.ID, username, msg.From.FirstName, msg.From.LastName); err != nil {
			log.Printf("Warning: Failed to upsert user %d: %v", msg.From.ID, err)
		}
	}()

	// Проверяем, находится ли пользователь в состоянии FSM
	if userState, exists := h.fsmManager.GetState(msg.From.ID); exists {
		h.handleFSMState(msg, userState)
		return
	}

	// Маршрутизация команд
	switch msg.Command() {
	case "start":
		h.handleStart(msg)
	case "products":
		h.handleProducts(msg)
	case "my_orders":
		h.handleMyOrders(msg)
	case "admin":
		h.handleAdmin(msg)
	case "cancel":
		h.handleCancel(msg)
	default:
		if msg.Command() != "" {
			h.sendMessage(msg.Chat.ID, "Неизвестная команда. Используйте /start, /products, /my_orders")
		}
	}
}

// HandleCallback обрабатывает callback запросы от inline кнопок
func (h *Handler) HandleCallback(query *tgbotapi.CallbackQuery) {
	log.Printf("Callback from %s: %s", query.From.UserName, query.Data)

	// Проверка rate limit
	if !h.checkRateLimit(query.From.ID, query.Message.Chat.ID) {
		// Answer callback with error
		callback := tgbotapi.NewCallback(query.ID, "⏳ Слишком много запросов")
		callback.ShowAlert = true
		h.bot.Request(callback)
		return
	}

	// Answer callback to remove loading state
	callback := tgbotapi.NewCallback(query.ID, "")
	if _, err := h.bot.Request(callback); err != nil {
		log.Printf("Error answering callback: %v", err)
	}

	// Сохраняем/обновляем пользователя в БД для рассылок
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		username := query.From.UserName
		if username == "" {
			username = query.From.FirstName
		}

		if err := h.storage.UpsertUser(ctx, query.From.ID, username, query.From.FirstName, query.From.LastName); err != nil {
			log.Printf("Warning: Failed to upsert user %d: %v", query.From.ID, err)
		}
	}()

	// Обрабатываем специальные случаи без разделителя
	if query.Data == "show_products" {
		h.handleShowProductsCallback(query)
		return
	}

	if query.Data == "change_region" {
		h.handleChangeRegion(query)
		return
	}

	// Парсим callback данные
	parts := strings.Split(query.Data, ":")
	if len(parts) < 2 {
		log.Printf("Invalid callback data: %s", query.Data)
		return
	}

	action := parts[0]
	value := parts[1]

	// Маршрутизация callback действий
	switch action {
	case "region":
		regionID, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Invalid region ID: %v", err)
			return
		}
		h.handleRegionSelection(query, regionID)

	case "category":
		categoryID, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Invalid category ID: %v", err)
			return
		}
		h.handleCategorySelection(query, categoryID)

	case "product":
		productID, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Invalid product ID: %v", err)
			return
		}
		h.handleProductSelection(query, productID)

	case "buy":
		productID, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Invalid product ID: %v", err)
			return
		}
		h.handleBuyProduct(query, productID)

	case "back":
		// Для back данные в формате back:type:id
		if len(parts) >= 3 {
			backType := parts[1]
			backID, _ := strconv.Atoi(parts[2])
			switch backType {
			case "regions":
				h.handleBackToRegions(query)
			case "categories":
				h.handleBackToCategories(query, backID)
			case "products":
				h.handleBackToProducts(query, backID)
			}
		} else if value == "catalog" {
			// Старый формат back:catalog
			h.handleBackToRegions(query)
		}

	case "confirm_payment":
		h.handleConfirmPayment(query, value)

	case "admin_edit_price":
		productID, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Invalid product ID: %v", err)
			return
		}
		h.handleAdminStartEditPrice(query, productID)

	case "admin_edit_name":
		productID, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Invalid product ID: %v", err)
			return
		}
		h.handleAdminStartEditName(query, productID)

	case "admin_edit_desc":
		productID, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Invalid product ID: %v", err)
			return
		}
		h.handleAdminStartEditDesc(query, productID)

	case "admin_toggle_visibility":
		h.handleAdminToggleVisibility(query, value)

	case "admin_products":
		h.handleAdminProducts(query)

	case "admin_edit_product":
		productID, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Invalid product ID: %v", err)
			return
		}
		h.handleAdminEditProduct(query, productID)

	case "admin_edit_welcome":
		h.handleAdminStartEditWelcome(query)

	case "broadcast_menu":
		h.handleAdminBroadcast(query)

	case "broadcast_start":
		h.handleBroadcastStart(query)

	case "broadcast_confirm":
		if userState, exists := h.fsmManager.GetState(query.From.ID); exists {
			h.handleBroadcastConfirm(query, userState)
		}

	case "broadcast_cancel":
		if userState, exists := h.fsmManager.GetState(query.From.ID); exists {
			h.handleBroadcastCancel(query, userState)
		}

	case "back_to_admin":
		fakeMsg := &tgbotapi.Message{
			Chat: &tgbotapi.Chat{ID: query.Message.Chat.ID},
			From: query.From,
		}
		h.handleAdmin(fakeMsg)
	}
}
