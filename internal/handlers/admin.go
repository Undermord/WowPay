package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tgwow/internal/models"
)

// isAdmin проверяет, является ли пользователь администратором
func (h *Handler) isAdmin(userID int64) bool {
	for _, adminID := range h.adminChatIDs {
		if adminID == userID {
			return true
		}
	}
	return false
}

// handleAdmin показывает админ-панель
func (h *Handler) handleAdmin(msg *tgbotapi.Message) {
	// Проверка что пользователь - админ
	if !h.isAdmin(msg.From.ID) {
		h.sendMessage(msg.Chat.ID, "❌ У вас нет доступа к этой команде.")
		return
	}

	ctx, cancel := h.newDBContext()
	defer cancel()

	// Получаем статистику
	stats, err := h.storage.GetOrderStats(ctx)
	if err != nil {
		log.Printf("Error fetching stats: %v", err)
		h.sendMessage(msg.Chat.ID, "❌ Ошибка при загрузке статистики.")
		return
	}

	// Получаем последние заказы
	recentOrders, err := h.storage.GetRecentOrders(ctx, RecentOrdersLimit)
	if err != nil {
		log.Printf("Error fetching recent orders: %v", err)
		recentOrders = []models.Order{}
	}

	// Собираем все ID товаров для batch-загрузки (решение N+1 проблемы)
	productIDs := make([]int, 0, len(recentOrders))
	for _, order := range recentOrders {
		productIDs = append(productIDs, order.ProductID)
	}

	// Загружаем все товары одним запросом
	products := make(map[int]*models.Product)
	if len(productIDs) > 0 {
		productsMap, err := h.storage.GetProductsByIDs(ctx, productIDs)
		if err != nil {
			log.Printf("Error fetching products: %v", err)
		} else {
			products = productsMap
		}
	}

	text := fmt.Sprintf(
		"👨‍💼 <b>Админ-панель</b>\n\n"+
			"📊 <b>Статистика:</b>\n"+
			"📦 Всего заказов: %d\n"+
			"⏳ Ожидают оплаты: %d\n"+
			"✅ Оплачено: %d\n"+
			"🎉 Завершено: %d\n"+
			"💰 Общая выручка: %.2f руб.\n\n"+
			"📋 <b>Последние заказы:</b>\n\n",
		stats["total_orders"],
		stats["pending_orders"],
		stats["paid_orders"],
		stats["completed_orders"],
		stats["total_revenue"],
	)

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for i, order := range recentOrders {
		if i >= DisplayedOrdersLimit {
			break
		}

		product, exists := products[order.ProductID]
		productName := "Товар"
		if exists {
			productName = product.Name
		}

		text += fmt.Sprintf(
			"%s <code>%s</code>\n"+
				"   %s - %.2f руб.\n"+
				"   User ID: %d\n\n",
			StatusEmojis[order.Status],
			order.OrderID,
			productName,
			order.Price,
			order.UserID,
		)

		// Добавляем кнопки для заказов в статусе "created"
		if order.Status == "created" {
			button := tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("✅ Подтвердить %s", order.OrderID),
				fmt.Sprintf("%s:%s", CallbackActionConfirmPayment, order.OrderID),
			)
			keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
		}
	}

	// Добавляем кнопки управления
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("📢 Создать рассылку", "broadcast_menu:0"),
	})
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("🛠 Управление товарами", CallbackActionAdminProducts+":0"),
	})
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("✏️ Редактировать приветствие", CallbackActionAdminEditWelcome+":0"),
	})

	response := tgbotapi.NewMessage(msg.Chat.ID, text)
	response.ParseMode = "HTML"

	if len(keyboard) > 0 {
		response.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
	}

	if _, err := h.bot.Send(response); err != nil {
		log.Printf("Error sending admin panel: %v", err)
	}
}

// handleAdminProducts показывает список товаров сгруппированных по регионам и категориям
func (h *Handler) handleAdminProducts(query *tgbotapi.CallbackQuery) {
	if !h.isAdmin(query.From.ID) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), DBContextTimeout)
	defer cancel()

	// Загружаем всё одним пакетом для решения N+1 проблемы
	regions, err := h.storage.ListRegions(ctx)
	if err != nil {
		log.Printf("Error fetching regions: %v", err)
		h.sendMessage(query.Message.Chat.ID, "❌ Ошибка при загрузке регионов.")
		return
	}

	allCategories, err := h.storage.ListAllCategories(ctx)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		h.sendMessage(query.Message.Chat.ID, "❌ Ошибка при загрузке категорий.")
		return
	}

	allProducts, err := h.storage.ListAllProducts(ctx)
	if err != nil {
		log.Printf("Error fetching products: %v", err)
		h.sendMessage(query.Message.Chat.ID, "❌ Ошибка при загрузке товаров.")
		return
	}

	// Группируем категории по region_id
	categoriesByRegion := make(map[int][]models.Category)
	for _, cat := range allCategories {
		categoriesByRegion[cat.RegionID] = append(categoriesByRegion[cat.RegionID], cat)
	}

	// Группируем товары по category_id
	productsByCategory := make(map[int][]models.Product)
	for _, prod := range allProducts {
		productsByCategory[prod.CategoryID] = append(productsByCategory[prod.CategoryID], prod)
	}

	// Создаем map регионов для быстрого доступа
	regionMap := make(map[int]models.Region)
	for _, reg := range regions {
		regionMap[reg.ID] = reg
	}

	text := "🛠 <b>Управление товарами</b>\n\n"
	var keyboard [][]tgbotapi.InlineKeyboardButton

	// Строим UI на основе загруженных данных
	for _, region := range regions {
		categories := categoriesByRegion[region.ID]

		text += fmt.Sprintf("%s <b>%s</b>\n", getRegionFlag(region.Code), region.Name)

		for _, category := range categories {
			products := productsByCategory[category.ID]

			// Подсчитываем видимые товары
			visibleCount := 0
			for _, p := range products {
				if p.IsVisible {
					visibleCount++
				}
			}

			text += fmt.Sprintf("  📁 %s: %d/%d товаров видно\n", category.Name, visibleCount, len(products))

			// Добавляем кнопки для каждого товара
			for _, p := range products {
				visibilityEmoji := "✅"
				if !p.IsVisible {
					visibilityEmoji = "❌"
				}

				priceText := fmt.Sprintf("%.0f₽", p.Price)
				if p.Price == 0 {
					priceText = "не указана"
				}

				button := tgbotapi.NewInlineKeyboardButtonData(
					fmt.Sprintf("%s [%s] %s - %s", visibilityEmoji, region.Code, p.Name, priceText),
					fmt.Sprintf("admin_edit_product:%d", p.ID),
				)
				keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
			}
		}
		text += "\n"
	}

	text += "Нажмите на товар для редактирования"

	keyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = &keyboardMarkup

	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// handleAdminEditProduct показывает детальную информацию о товаре с возможностью редактирования
func (h *Handler) handleAdminEditProduct(query *tgbotapi.CallbackQuery, productID int) {
	if !h.isAdmin(query.From.ID) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), DBContextTimeout)
	defer cancel()

	product, err := h.storage.GetProductByID(ctx, productID)
	if err != nil {
		log.Printf("Error fetching product: %v", err)
		return
	}

	category, err := h.storage.GetCategoryByID(ctx, product.CategoryID)
	if err != nil {
		log.Printf("Error fetching category: %v", err)
		return
	}

	region, err := h.storage.GetRegionByID(ctx, category.RegionID)
	if err != nil {
		log.Printf("Error fetching region: %v", err)
		return
	}

	visibilityStatus := "Видимый ✅"
	if !product.IsVisible {
		visibilityStatus = "Скрытый ❌"
	}

	text := fmt.Sprintf(
		"📦 <b>Редактирование товара</b>\n\n"+
			"%s <b>Регион:</b> %s\n"+
			"📁 <b>Категория:</b> %s\n"+
			"🏷 <b>Название:</b> %s\n"+
			"💰 <b>Цена:</b> %.2f руб.\n"+
			"👁 <b>Статус:</b> %s\n"+
			"🆔 <b>ID:</b> %d\n\n"+
			"📝 <b>Описание:</b>\n%s",
		getRegionFlag(region.Code), region.Name, category.Name, product.Name, product.Price, visibilityStatus, product.ID, product.Description,
	)

	toggleText := "Скрыть товар"
	if !product.IsVisible {
		toggleText = "Показать товар"
	}

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"💰 Изменить цену",
				fmt.Sprintf("admin_edit_price:%d", product.ID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"✏️ Изменить название",
				fmt.Sprintf("admin_edit_name:%d", product.ID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"📝 Изменить описание",
				fmt.Sprintf("admin_edit_desc:%d", product.ID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				fmt.Sprintf("👁 %s", toggleText),
				fmt.Sprintf("admin_toggle_visibility:%d", product.ID),
			),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(
				"◀️ Назад к списку",
				"admin_products:0",
			),
		),
	)

	edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = &keyboard

	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// handleAdminToggleVisibility переключает видимость товара
func (h *Handler) handleAdminToggleVisibility(query *tgbotapi.CallbackQuery, productIDStr string) {
	if !h.isAdmin(query.From.ID) {
		return
	}

	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		log.Printf("Invalid product ID: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), DBContextTimeout)
	defer cancel()

	product, err := h.storage.GetProductByID(ctx, productID)
	if err != nil {
		log.Printf("Error fetching product: %v", err)
		return
	}

	// Переключаем видимость
	newVisibility := !product.IsVisible
	if err := h.storage.UpdateProductVisibility(ctx, productID, newVisibility); err != nil {
		log.Printf("Error updating visibility: %v", err)
		// Показываем alert с ошибкой
		callback := tgbotapi.NewCallback(query.ID, "❌ Ошибка при изменении видимости")
		callback.ShowAlert = true
		h.bot.Request(callback)
		return
	}

	// Показываем уведомление об успехе
	status := "скрыт"
	if newVisibility {
		status = "показан"
	}
	successCallback := tgbotapi.NewCallback(query.ID, fmt.Sprintf("✅ Товар %s", status))
	h.bot.Request(successCallback)

	// Возвращаемся к редактированию товара с обновленной информацией
	h.handleAdminEditProduct(query, productID)
}
