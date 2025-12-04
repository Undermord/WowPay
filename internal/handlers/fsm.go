package handlers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tgwow/internal/fsm"
	"tgwow/internal/validation"
)

// handleFSMState обрабатывает сообщения в зависимости от состояния FSM
func (h *Handler) handleFSMState(msg *tgbotapi.Message, userState *fsm.UserState) {
	switch userState.State {
	case fsm.StateWaitingForPrice:
		h.handlePriceInput(msg, userState.ProductID)
	case fsm.StateWaitingForName:
		h.handleNameInput(msg, userState.ProductID)
	case fsm.StateWaitingForDesc:
		h.handleDescInput(msg, userState.ProductID)
	case fsm.StateWaitingForWelcomeMsg:
		h.handleWelcomeMsgInput(msg)
	case fsm.StateWaitingForBroadcastText:
		h.handleBroadcastTextInput(msg, userState)
	case fsm.StateWaitingForBroadcastPhoto:
		h.handleBroadcastPhotoInput(msg, userState)
	}
}

// handleAdminStartEditPrice начинает диалог изменения цены
func (h *Handler) handleAdminStartEditPrice(query *tgbotapi.CallbackQuery, productID int) {
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

	// Устанавливаем состояние FSM
	h.fsmManager.SetState(query.From.ID, fsm.StateWaitingForPrice, productID)

	text := fmt.Sprintf(
		"💰 <b>Изменение цены товара</b>\n\n"+
			"Товар: <b>%s</b>\n"+
			"Текущая цена: %.2f руб.\n\n"+
			"Введите новую цену в рублях (например: 2500 или 2500.50)\n\n"+
			"Для отмены используйте /cancel",
		product.Name, product.Price,
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	h.bot.Send(msg)
}

// handleAdminStartEditName начинает диалог изменения названия
func (h *Handler) handleAdminStartEditName(query *tgbotapi.CallbackQuery, productID int) {
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

	// Устанавливаем состояние FSM
	h.fsmManager.SetState(query.From.ID, fsm.StateWaitingForName, productID)

	text := fmt.Sprintf(
		"✏️ <b>Изменение названия товара</b>\n\n"+
			"Текущее название: <b>%s</b>\n\n"+
			"Введите новое название товара\n\n"+
			"Для отмены используйте /cancel",
		product.Name,
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	h.bot.Send(msg)
}

// handleAdminStartEditDesc начинает диалог изменения описания
func (h *Handler) handleAdminStartEditDesc(query *tgbotapi.CallbackQuery, productID int) {
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

	// Устанавливаем состояние FSM
	h.fsmManager.SetState(query.From.ID, fsm.StateWaitingForDesc, productID)

	text := fmt.Sprintf(
		"📝 <b>Изменение описания товара</b>\n\n"+
			"Товар: <b>%s</b>\n\n"+
			"Текущее описание:\n%s\n\n"+
			"Введите новое описание товара\n\n"+
			"Для отмены используйте /cancel",
		product.Name, product.Description,
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	h.bot.Send(msg)
}

// handleAdminStartEditWelcome начинает диалог редактирования приветственного сообщения
func (h *Handler) handleAdminStartEditWelcome(query *tgbotapi.CallbackQuery) {
	if !h.isAdmin(query.From.ID) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), DBContextTimeout)
	defer cancel()

	settings, err := h.storage.GetBotSettings(ctx)
	if err != nil {
		log.Printf("Error fetching bot settings: %v", err)
		h.sendMessage(query.Message.Chat.ID, "❌ Ошибка при загрузке настроек")
		return
	}

	// Устанавливаем состояние FSM (productID не используется, ставим 0)
	h.fsmManager.SetState(query.From.ID, fsm.StateWaitingForWelcomeMsg, 0)

	text := fmt.Sprintf(
		"✏️ <b>Редактирование приветственного сообщения</b>\n\n"+
			"<b>Текущее сообщение:</b>\n%s\n\n"+
			"Введите новое приветственное сообщение.\n\n"+
			"💡 <b>Подсказки:</b>\n"+
			"• Используйте <code>{name}</code> для вставки имени пользователя\n"+
			"• Можно использовать HTML-теги: &lt;b&gt;жирный&lt;/b&gt;, &lt;i&gt;курсив&lt;/i&gt;, &lt;a href=\"url\"&gt;ссылка&lt;/a&gt;\n"+
			"• Для отмены используйте /cancel",
		settings.WelcomeMessage,
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	h.bot.Send(msg)
}

// handlePriceInput обрабатывает ввод новой цены
func (h *Handler) handlePriceInput(msg *tgbotapi.Message, productID int) {
	if !h.isAdmin(msg.From.ID) {
		return
	}

	// Парсим цену
	newPrice, err := strconv.ParseFloat(strings.TrimSpace(msg.Text), 64)
	if err != nil {
		h.sendMessage(msg.Chat.ID, "❌ Неверный формат цены. Введите число (например: 2500 или 2500.50)")
		return
	}

	// Валидация цены
	if err := validation.ValidatePrice(newPrice); err != nil {
		h.sendMessage(msg.Chat.ID, fmt.Sprintf("❌ %s\n\nПопробуйте еще раз или используйте /cancel для отмены.", err.Error()))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), DBContextTimeout)
	defer cancel()

	// Обновляем цену
	if err := h.storage.UpdateProductPrice(ctx, productID, newPrice); err != nil {
		log.Printf("Error updating price: %v", err)
		h.sendMessage(msg.Chat.ID, "❌ Ошибка при обновлении цены")
		h.fsmManager.ClearState(msg.From.ID)
		return
	}

	product, _ := h.storage.GetProductByID(ctx, productID)

	h.sendMessage(msg.Chat.ID, fmt.Sprintf("✅ Цена товара \"%s\" обновлена на %.2f руб.", product.Name, newPrice))
	h.fsmManager.ClearState(msg.From.ID)
}

// handleNameInput обрабатывает ввод нового названия
func (h *Handler) handleNameInput(msg *tgbotapi.Message, productID int) {
	if !h.isAdmin(msg.From.ID) {
		return
	}

	newName := strings.TrimSpace(msg.Text)
	if newName == "" {
		h.sendMessage(msg.Chat.ID, "❌ Название не может быть пустым")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), DBContextTimeout)
	defer cancel()

	product, err := h.storage.GetProductByID(ctx, productID)
	if err != nil {
		log.Printf("Error fetching product: %v", err)
		h.fsmManager.ClearState(msg.From.ID)
		return
	}

	// Обновляем название
	if err := h.storage.UpdateProduct(ctx, productID, newName, product.Price, product.Description); err != nil {
		log.Printf("Error updating name: %v", err)
		h.sendMessage(msg.Chat.ID, "❌ Ошибка при обновлении названия")
		h.fsmManager.ClearState(msg.From.ID)
		return
	}

	h.sendMessage(msg.Chat.ID, fmt.Sprintf("✅ Название товара обновлено на: \"%s\"", newName))
	h.fsmManager.ClearState(msg.From.ID)
}

// handleDescInput обрабатывает ввод нового описания
func (h *Handler) handleDescInput(msg *tgbotapi.Message, productID int) {
	if !h.isAdmin(msg.From.ID) {
		return
	}

	newDesc := strings.TrimSpace(msg.Text)
	if newDesc == "" {
		h.sendMessage(msg.Chat.ID, "❌ Описание не может быть пустым")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), DBContextTimeout)
	defer cancel()

	product, err := h.storage.GetProductByID(ctx, productID)
	if err != nil {
		log.Printf("Error fetching product: %v", err)
		h.fsmManager.ClearState(msg.From.ID)
		return
	}

	// Обновляем описание
	if err := h.storage.UpdateProduct(ctx, productID, product.Name, product.Price, newDesc); err != nil {
		log.Printf("Error updating description: %v", err)
		h.sendMessage(msg.Chat.ID, "❌ Ошибка при обновлении описания")
		h.fsmManager.ClearState(msg.From.ID)
		return
	}

	h.sendMessage(msg.Chat.ID, "✅ Описание товара обновлено")
	h.fsmManager.ClearState(msg.From.ID)
}

// handleWelcomeMsgInput обрабатывает ввод нового приветственного сообщения
func (h *Handler) handleWelcomeMsgInput(msg *tgbotapi.Message) {
	if !h.isAdmin(msg.From.ID) {
		return
	}

	newMessage := strings.TrimSpace(msg.Text)
	if newMessage == "" {
		h.sendMessage(msg.Chat.ID, "❌ Приветственное сообщение не может быть пустым")
		return
	}

	// Валидация HTML
	if err := validation.ValidateHTML(newMessage); err != nil {
		errorText := fmt.Sprintf(
			"❌ <b>Ошибка валидации HTML:</b>\n%s\n\n"+
				"💡 <b>Разрешенные теги:</b>\n"+
				"• &lt;b&gt;жирный&lt;/b&gt;\n"+
				"• &lt;i&gt;курсив&lt;/i&gt;\n"+
				"• &lt;u&gt;подчеркнутый&lt;/u&gt;\n"+
				"• &lt;s&gt;зачеркнутый&lt;/s&gt;\n"+
				"• &lt;code&gt;код&lt;/code&gt;\n"+
				"• &lt;pre&gt;блок кода&lt;/pre&gt;\n"+
				"• &lt;a href=\"url\"&gt;ссылка&lt;/a&gt;\n\n"+
				"Попробуйте еще раз или используйте /cancel для отмены.",
			err.Error(),
		)
		response := tgbotapi.NewMessage(msg.Chat.ID, errorText)
		response.ParseMode = "HTML"
		h.bot.Send(response)
		return
	}

	// Санитизация HTML (удаление опасных тегов)
	sanitizedMessage := validation.SanitizeHTML(newMessage)

	ctx, cancel := context.WithTimeout(context.Background(), DBContextTimeout)
	defer cancel()

	// Обновляем приветственное сообщение
	if err := h.storage.UpdateWelcomeMessage(ctx, sanitizedMessage); err != nil {
		log.Printf("Error updating welcome message: %v", err)
		h.sendMessage(msg.Chat.ID, "❌ Ошибка при обновлении приветственного сообщения")
		h.fsmManager.ClearState(msg.From.ID)
		return
	}

	h.sendMessage(msg.Chat.ID, "✅ Приветственное сообщение успешно обновлено!")
	h.fsmManager.ClearState(msg.From.ID)

	log.Printf("Welcome message updated by admin %d", msg.From.ID)
}
