package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleShowProductsCallback показывает каталог товаров при нажатии кнопки "Далее"
func (h *Handler) handleShowProductsCallback(query *tgbotapi.CallbackQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	regions, err := h.storage.ListRegions(ctx)
	if err != nil {
		log.Printf("Error fetching regions: %v", err)
		callback := tgbotapi.NewCallback(query.ID, "❌ Ошибка при загрузке каталога")
		callback.ShowAlert = true
		h.bot.Request(callback)
		return
	}

	if len(regions) == 0 {
		callback := tgbotapi.NewCallback(query.ID, "📦 Каталог пуст")
		callback.ShowAlert = true
		h.bot.Request(callback)
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

	keyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	// Редактируем существующее сообщение
	edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = &keyboardMarkup

	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// handleRegionSelection показывает категории для выбранного региона
func (h *Handler) handleRegionSelection(query *tgbotapi.CallbackQuery, regionID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	region, err := h.storage.GetRegionByID(ctx, regionID)
	if err != nil {
		log.Printf("Error fetching region: %v", err)
		return
	}

	categories, err := h.storage.ListCategoriesByRegion(ctx, regionID)
	if err != nil {
		log.Printf("Error fetching categories: %v", err)
		return
	}

	if len(categories) == 0 {
		h.sendMessage(query.Message.Chat.ID, "📦 В этом регионе пока нет категорий.")
		return
	}

	text := fmt.Sprintf("%s <b>%s</b>\n\nВыберите категорию:", getRegionFlag(region.Code), region.Name)

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, c := range categories {
		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("📁 %s", c.Name),
			fmt.Sprintf("category:%d", c.ID),
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	// Кнопка "Назад"
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("◀️ Назад к регионам", "back:regions:0"),
	})

	keyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = &keyboardMarkup

	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// handleCategorySelection показывает товары для выбранной категории
func (h *Handler) handleCategorySelection(query *tgbotapi.CallbackQuery, categoryID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	category, err := h.storage.GetCategoryByID(ctx, categoryID)
	if err != nil {
		log.Printf("Error fetching category: %v", err)
		return
	}

	region, err := h.storage.GetRegionByID(ctx, category.RegionID)
	if err != nil {
		log.Printf("Error fetching region: %v", err)
		return
	}

	products, err := h.storage.ListProductsByCategory(ctx, categoryID)
	if err != nil {
		log.Printf("Error fetching products: %v", err)
		return
	}

	if len(products) == 0 {
		h.sendMessage(query.Message.Chat.ID, "📦 В этой категории пока нет товаров.")
		return
	}

	text := fmt.Sprintf("%s %s → 📁 <b>%s</b>\n\n", region.Name, getRegionFlag(region.Code), category.Name)

	if category.Description != "" {
		text += fmt.Sprintf("📝 %s\n\n", category.Description)
	}

	text += "Выберите товар:\n\n"

	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, p := range products {
		priceText := ""
		if p.Price > 0 {
			priceText = fmt.Sprintf(" - %.2f руб.", p.Price)
		} else {
			priceText = " - цена уточняется"
		}

		button := tgbotapi.NewInlineKeyboardButtonData(
			fmt.Sprintf("%s%s", p.Name, priceText),
			fmt.Sprintf("product:%d", p.ID),
		)
		keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{button})
	}

	// Кнопка "Назад"
	keyboard = append(keyboard, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("◀️ Назад к категориям", fmt.Sprintf("back:categories:%d", region.ID)),
	})

	keyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = &keyboardMarkup

	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// handleProductSelection показывает карточку товара
func (h *Handler) handleProductSelection(query *tgbotapi.CallbackQuery, productID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product, err := h.storage.GetProductByID(ctx, productID)
	if err != nil {
		log.Printf("Error fetching product: %v", err)
		h.sendMessage(query.Message.Chat.ID, "❌ Ошибка при загрузке товара.")
		return
	}

	category, err := h.storage.GetCategoryByID(ctx, product.CategoryID)
	if err != nil {
		log.Printf("Error fetching category: %v", err)
		return
	}

	priceText := ""
	if product.Price > 0 {
		priceText = fmt.Sprintf("💰 <b>Цена:</b> %.2f руб.\n\n", product.Price)
	} else {
		priceText = "💰 <b>Цена:</b> уточняется\n\n"
	}

	text := fmt.Sprintf(
		"🎮 <b>%s</b>\n\n"+
			"%s"+
			"📝 <b>Описание:</b>\n%s",
		product.Name, priceText, product.Description,
	)

	var keyboard tgbotapi.InlineKeyboardMarkup

	if product.Price > 0 {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Купить",
					fmt.Sprintf("buy:%d", product.ID),
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"◀️ Назад к товарам",
					fmt.Sprintf("back:products:%d", category.ID),
				),
			),
		)
	} else {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"◀️ Назад к товарам",
					fmt.Sprintf("back:products:%d", category.ID),
				),
			),
		)
	}

	// Edit existing message
	edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = &keyboard

	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// handleChangeRegion показывает карточку товара "Сменить регион"
func (h *Handler) handleChangeRegion(query *tgbotapi.CallbackQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Получаем товар "Сменить регион" из БД
	product, err := h.storage.GetChangeRegionProduct(ctx)
	if err != nil {
		log.Printf("Error fetching change region product: %v", err)
		callback := tgbotapi.NewCallback(query.ID, "❌ Ошибка при загрузке услуги")
		callback.ShowAlert = true
		h.bot.Request(callback)
		return
	}

	priceText := ""
	if product.Price > 0 {
		priceText = fmt.Sprintf("💰 <b>Цена:</b> %.2f руб.\n\n", product.Price)
	} else {
		priceText = "💰 <b>Цена:</b> уточняется\n\n"
	}

	text := fmt.Sprintf(
		"🔄 <b>%s</b>\n\n"+
			"%s"+
			"📝 <b>Описание:</b>\n%s",
		product.Name, priceText, product.Description,
	)

	var keyboard tgbotapi.InlineKeyboardMarkup

	if product.Price > 0 {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"✅ Купить",
					fmt.Sprintf("buy:%d", product.ID),
				),
			),
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"◀️ Назад к регионам",
					"back:regions:0",
				),
			),
		)
	} else {
		keyboard = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(
					"◀️ Назад к регионам",
					"back:regions:0",
				),
			),
		)
	}

	// Редактируем существующее сообщение
	edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = &keyboard

	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// handleBackToRegions возвращает к списку регионов
func (h *Handler) handleBackToRegions(query *tgbotapi.CallbackQuery) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	regions, err := h.storage.ListRegions(ctx)
	if err != nil {
		log.Printf("Error fetching regions: %v", err)
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

	keyboardMarkup := tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = &keyboardMarkup

	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// handleBackToCategories возвращает к списку категорий региона
func (h *Handler) handleBackToCategories(query *tgbotapi.CallbackQuery, regionID int) {
	h.handleRegionSelection(query, regionID)
}

// handleBackToProducts возвращает к списку товаров категории
func (h *Handler) handleBackToProducts(query *tgbotapi.CallbackQuery, categoryID int) {
	h.handleCategorySelection(query, categoryID)
}
