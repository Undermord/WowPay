package handlers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tgwow/internal/fsm"
	"tgwow/internal/validation"
)

const (
	// BroadcastDelay - задержка между отправками для защиты от бана (50ms = 20 msg/sec)
	BroadcastDelay = 50 * time.Millisecond

	// BroadcastBatchSize - размер батча для обновления прогресса
	BroadcastBatchSize = 10
)

// handleAdminBroadcast показывает меню создания рассылки
func (h *Handler) handleAdminBroadcast(query *tgbotapi.CallbackQuery) {
	if !h.isAdmin(query.From.ID) {
		return
	}

	ctx, cancel := h.newDBContext()
	defer cancel()

	usersCount, err := h.storage.GetUsersCount(ctx)
	if err != nil {
		log.Printf("Error getting users count: %v", err)
		usersCount = 0
	}

	text := fmt.Sprintf(
		"📢 <b>Создание рассылки</b>\n\n"+
			"👥 Активных пользователей: <b>%d</b>\n\n"+
			"Нажмите кнопку ниже для создания новой рассылки.",
		usersCount,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✍️ Создать рассылку", "broadcast_start:0"),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("◀️ Назад в админ-панель", "back_to_admin:0"),
		),
	)

	edit := tgbotapi.NewEditMessageText(query.Message.Chat.ID, query.Message.MessageID, text)
	edit.ParseMode = "HTML"
	edit.ReplyMarkup = &keyboard

	if _, err := h.bot.Send(edit); err != nil {
		log.Printf("Error editing message: %v", err)
	}
}

// handleBroadcastStart начинает диалог создания рассылки
func (h *Handler) handleBroadcastStart(query *tgbotapi.CallbackQuery) {
	if !h.isAdmin(query.From.ID) {
		return
	}

	h.fsmManager.SetState(query.From.ID, fsm.StateWaitingForBroadcastText, 0)

	text := "✍️ <b>Шаг 1/3: Введите текст рассылки</b>\n\n" +
		"Вы можете использовать HTML-форматирование:\n" +
		"• <code>&lt;b&gt;жирный&lt;/b&gt;</code>\n" +
		"• <code>&lt;i&gt;курсив&lt;/i&gt;</code>\n" +
		"• <code>&lt;a href=\"url\"&gt;ссылка&lt;/a&gt;</code>\n\n" +
		"Для отмены используйте /cancel"

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, text)
	msg.ParseMode = "HTML"
	h.bot.Send(msg)
}

// handleBroadcastTextInput обрабатывает ввод текста рассылки
func (h *Handler) handleBroadcastTextInput(msg *tgbotapi.Message, userState *fsm.UserState) {
	if !h.isAdmin(msg.From.ID) {
		return
	}

	text := strings.TrimSpace(msg.Text)
	if text == "" {
		h.sendMessage(msg.Chat.ID, "❌ Текст рассылки не может быть пустым")
		return
	}

	if err := validation.ValidateHTML(text); err != nil {
		errorText := fmt.Sprintf(
			"❌ <b>Ошибка валидации HTML:</b>\n%s\n\n"+
				"Попробуйте еще раз или используйте /cancel",
			err.Error(),
		)
		response := tgbotapi.NewMessage(msg.Chat.ID, errorText)
		response.ParseMode = "HTML"
		h.bot.Send(response)
		return
	}

	sanitizedText := validation.SanitizeHTML(text)

	ctx, cancel := h.newDBContext()
	defer cancel()

	broadcast, err := h.storage.CreateBroadcast(ctx, msg.From.ID, sanitizedText)
	if err != nil {
		log.Printf("Error creating broadcast: %v", err)
		h.sendMessage(msg.Chat.ID, "❌ Ошибка при создании рассылки")
		h.fsmManager.ClearState(msg.From.ID)
		return
	}

	userState.Data["broadcast_id"] = broadcast.ID
	userState.Data["broadcast_text"] = sanitizedText

	h.fsmManager.SetBroadcastState(msg.From.ID, fsm.StateWaitingForBroadcastPhoto, broadcast.ID)

	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("⏭ Пропустить фото"),
		),
	)

	response := tgbotapi.NewMessage(msg.Chat.ID,
		"✅ Текст сохранен!\n\n"+
			"📸 <b>Шаг 2/3: Отправьте фотографию</b>\n\n"+
			"Вы можете отправить одно изображение или нажать \"Пропустить фото\".\n\n"+
			"Для отмены используйте /cancel",
	)
	response.ParseMode = "HTML"
	response.ReplyMarkup = keyboard
	h.bot.Send(response)
}

// handleBroadcastPhotoInput обрабатывает загрузку фото
func (h *Handler) handleBroadcastPhotoInput(msg *tgbotapi.Message, userState *fsm.UserState) {
	if !h.isAdmin(msg.From.ID) {
		return
	}

	broadcastID, ok := userState.Data["broadcast_id"].(int)
	if !ok {
		h.sendMessage(msg.Chat.ID, "❌ Ошибка: рассылка не найдена")
		h.fsmManager.ClearState(msg.From.ID)
		return
	}

	if msg.Text == "⏭ Пропустить фото" {
		userState.Data["skip_photos"] = true
		h.showBroadcastPreview(msg.Chat.ID, msg.From.ID, userState)
		return
	}

	if msg.Photo == nil || len(msg.Photo) == 0 {
		h.sendMessage(msg.Chat.ID, "❌ Пожалуйста, отправьте фотографию или нажмите \"Пропустить фото\"")
		return
	}

	photo := msg.Photo[len(msg.Photo)-1]
	fileID := photo.FileID

	ctx, cancel := h.newDBContext()
	defer cancel()

	if err := h.storage.SaveBroadcastPhoto(ctx, broadcastID, fileID, 0); err != nil {
		log.Printf("Error saving broadcast photo: %v", err)
		h.sendMessage(msg.Chat.ID, "❌ Ошибка при сохранении фото")
		return
	}

	fileIDs, _ := userState.Data["photo_file_ids"].([]string)
	fileIDs = append(fileIDs, fileID)
	userState.Data["photo_file_ids"] = fileIDs

	h.showBroadcastPreview(msg.Chat.ID, msg.From.ID, userState)
}

// showBroadcastPreview показывает предпросмотр рассылки
func (h *Handler) showBroadcastPreview(chatID int64, userID int64, userState *fsm.UserState) {
	broadcastText, _ := userState.Data["broadcast_text"].(string)
	photoFileIDs, _ := userState.Data["photo_file_ids"].([]string)
	skipPhotos, _ := userState.Data["skip_photos"].(bool)

	h.fsmManager.SetBroadcastState(userID, fsm.StateConfirmingBroadcast, 0)

	ctx, cancel := h.newDBContext()
	defer cancel()

	usersCount, _ := h.storage.GetUsersCount(ctx)

	previewText := fmt.Sprintf(
		"👁 <b>Шаг 3/3: Предпросмотр рассылки</b>\n\n"+
			"📝 <b>Текст:</b>\n%s\n\n"+
			"📸 <b>Фото:</b> %s\n"+
			"👥 <b>Получателей:</b> %d\n\n"+
			"Все верно?",
		broadcastText,
		getPhotoStatus(photoFileIDs, skipPhotos),
		usersCount,
	)

	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Отправить", "broadcast_confirm:0"),
			tgbotapi.NewInlineKeyboardButtonData("❌ Отменить", "broadcast_cancel:0"),
		),
	)

	removeKeyboard := tgbotapi.NewRemoveKeyboard(true)

	if len(photoFileIDs) > 0 && !skipPhotos {
		// С фото: отправляем фото с кнопками
		photoMsg := tgbotapi.NewPhoto(chatID, tgbotapi.FileID(photoFileIDs[0]))
		photoMsg.Caption = previewText
		photoMsg.ParseMode = "HTML"
		photoMsg.ReplyMarkup = keyboard

		// Убираем reply клавиатуру перед отправкой
		cleanupMsg := tgbotapi.NewMessage(chatID, "Подготовка предпросмотра...")
		cleanupMsg.ReplyMarkup = removeKeyboard
		h.bot.Send(cleanupMsg)

		h.bot.Send(photoMsg)
	} else {
		// Без фото: убираем reply клавиатуру и отправляем сообщение с inline кнопками
		cleanupMsg := tgbotapi.NewMessage(chatID, "📋 Подготовка предпросмотра...")
		cleanupMsg.ReplyMarkup = removeKeyboard
		h.bot.Send(cleanupMsg)

		msg := tgbotapi.NewMessage(chatID, previewText)
		msg.ParseMode = "HTML"
		msg.ReplyMarkup = keyboard
		h.bot.Send(msg)
	}
}

// handleBroadcastConfirm подтверждает и начинает рассылку
func (h *Handler) handleBroadcastConfirm(query *tgbotapi.CallbackQuery, userState *fsm.UserState) {
	if !h.isAdmin(query.From.ID) {
		return
	}

	broadcastID, ok := userState.Data["broadcast_id"].(int)
	if !ok {
		h.sendMessage(query.Message.Chat.ID, "❌ Ошибка: рассылка не найдена")
		h.fsmManager.ClearState(query.From.ID)
		return
	}

	h.fsmManager.ClearState(query.From.ID)

	h.sendMessage(query.Message.Chat.ID, "🚀 Рассылка начата! Это может занять некоторое время...")

	go h.executeBroadcast(broadcastID, query.Message.Chat.ID)
}

// handleBroadcastCancel отменяет рассылку
func (h *Handler) handleBroadcastCancel(query *tgbotapi.CallbackQuery, userState *fsm.UserState) {
	if !h.isAdmin(query.From.ID) {
		return
	}

	broadcastID, ok := userState.Data["broadcast_id"].(int)
	if ok {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := h.storage.DeleteBroadcast(ctx, broadcastID); err != nil {
			log.Printf("Error deleting broadcast: %v", err)
		}
	}

	h.fsmManager.ClearState(query.From.ID)
	h.sendMessage(query.Message.Chat.ID, "❌ Рассылка отменена")
}

// executeBroadcast выполняет массовую рассылку
func (h *Handler) executeBroadcast(broadcastID int, adminChatID int64) {
	ctx := context.Background()

	broadcast, err := h.storage.GetBroadcastByID(ctx, broadcastID)
	if err != nil {
		log.Printf("Error getting broadcast: %v", err)
		h.sendMessage(adminChatID, "❌ Ошибка при загрузке рассылки")
		return
	}

	photos, err := h.storage.GetBroadcastPhotos(ctx, broadcastID)
	if err != nil {
		log.Printf("Error getting photos: %v", err)
	}

	hasPhoto := len(photos) > 0
	var photoFileID string
	if hasPhoto {
		photoFileID = photos[0].FileID
	}

	users, err := h.storage.GetActiveUsers(ctx)
	if err != nil {
		log.Printf("Error getting users: %v", err)
		h.sendMessage(adminChatID, "❌ Ошибка при загрузке пользователей")
		return
	}

	totalUsers := len(users)
	sentCount := 0
	failedCount := 0

	h.storage.UpdateBroadcastStatus(ctx, broadcastID, "sending", totalUsers, 0, 0)

	log.Printf("Starting broadcast %d to %d users", broadcastID, totalUsers)

	for i, user := range users {
		var err error

		if hasPhoto {
			photoMsg := tgbotapi.NewPhoto(user.UserID, tgbotapi.FileID(photoFileID))
			photoMsg.Caption = broadcast.Text
			photoMsg.ParseMode = "HTML"
			_, err = h.bot.Send(photoMsg)
		} else {
			msg := tgbotapi.NewMessage(user.UserID, broadcast.Text)
			msg.ParseMode = "HTML"
			_, err = h.bot.Send(msg)
		}

		if err != nil {
			log.Printf("Failed to send to user %d: %v", user.UserID, err)
			failedCount++

			if strings.Contains(err.Error(), "Forbidden: bot was blocked by the user") {
				h.storage.MarkUserAsBlocked(ctx, user.UserID)
			}
		} else {
			sentCount++
		}

		time.Sleep(BroadcastDelay)

		if (i+1)%BroadcastBatchSize == 0 || i == totalUsers-1 {
			h.storage.UpdateBroadcastStatus(ctx, broadcastID, "sending", totalUsers, sentCount, failedCount)

			progressText := fmt.Sprintf(
				"📊 Прогресс: %d/%d (%.1f%%)\n"+
					"✅ Отправлено: %d\n"+
					"❌ Ошибок: %d",
				i+1, totalUsers, float64(i+1)/float64(totalUsers)*100,
				sentCount, failedCount,
			)
			h.sendMessage(adminChatID, progressText)
		}
	}

	finalStatus := "completed"
	if sentCount == 0 {
		finalStatus = "failed"
	}

	h.storage.UpdateBroadcastStatus(ctx, broadcastID, finalStatus, totalUsers, sentCount, failedCount)

	resultText := fmt.Sprintf(
		"✅ <b>Рассылка завершена!</b>\n\n"+
			"👥 Всего пользователей: %d\n"+
			"✅ Отправлено: %d\n"+
			"❌ Ошибок: %d\n"+
			"📊 Успешность: %.1f%%",
		totalUsers, sentCount, failedCount,
		float64(sentCount)/float64(totalUsers)*100,
	)

	msg := tgbotapi.NewMessage(adminChatID, resultText)
	msg.ParseMode = "HTML"
	h.bot.Send(msg)

	log.Printf("Broadcast %d completed: %d sent, %d failed", broadcastID, sentCount, failedCount)
}

// getPhotoStatus возвращает статус фото для предпросмотра
func getPhotoStatus(photoFileIDs []string, skipPhotos bool) string {
	if skipPhotos {
		return "Без фото"
	}
	if len(photoFileIDs) > 0 {
		return "1 фото"
	}
	return "Без фото"
}
