package handlers

import (
	"context"
	"fmt"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleBuyProduct обрабатывает покупку товара
func (h *Handler) handleBuyProduct(query *tgbotapi.CallbackQuery, productID int) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	product, err := h.storage.GetProductByID(ctx, productID)
	if err != nil {
		log.Printf("Error fetching product: %v", err)
		h.sendMessage(query.Message.Chat.ID, "❌ Ошибка при загрузке товара.")
		return
	}

	// Create order
	order, err := h.storage.CreateOrder(ctx, query.From.ID, product.ID, product.Price)
	if err != nil {
		log.Printf("Error creating order: %v", err)
		h.sendMessage(query.Message.Chat.ID, "❌ Ошибка при создании заказа. Попробуйте позже.")
		return
	}

	log.Printf("Order created: %+v", order)

	// Send payment instructions to user
	userText := fmt.Sprintf(
		"✅ <b>Заказ успешно создан!</b>\n\n"+
			"📦 <b>Заказ №:</b> <code>%s</code>\n"+
			"🎮 <b>Товар:</b> %s\n"+
			"💰 <b>Сумма:</b> %.2f руб.\n\n"+
			"💳 <b>Инструкция по оплате:</b>\n"+
			"1. Переведите %.2f руб. на карту: <code>%s</code>\n"+
			"2. В комментарии к переводу укажите номер заказа: <code>%s</code>\n"+
			"3. Отправьте скриншот оплаты администратору\n\n"+
			"После проверки оплаты вы получите доступ к подписке.\n\n"+
			"По всем вопросам обращайтесь к администратору.",
		order.OrderID, product.Name, product.Price,
		product.Price, h.paymentCardNumber, order.OrderID,
	)

	msg := tgbotapi.NewMessage(query.Message.Chat.ID, userText)
	msg.ParseMode = "HTML"
	if _, err := h.bot.Send(msg); err != nil {
		log.Printf("Error sending order confirmation: %v", err)
	}

	// Notify all admins
	// Convert time to Moscow timezone (MSK, UTC+3)
	moscowLocation, _ := time.LoadLocation("Europe/Moscow")
	moscowTime := order.CreatedAt.In(moscowLocation)

	adminText := fmt.Sprintf(
		"🔔 <b>Новый заказ!</b>\n\n"+
			"📦 <b>Заказ №:</b> <code>%s</code>\n"+
			"👤 <b>Пользователь:</b> @%s (ID: %d)\n"+
			"🎮 <b>Товар:</b> %s\n"+
			"💰 <b>Сумма:</b> %.2f руб.\n"+
			"📅 <b>Дата:</b> %s (МСК)\n\n"+
			"Ожидает оплаты.",
		order.OrderID,
		query.From.UserName, query.From.ID,
		product.Name, product.Price,
		moscowTime.Format("02.01.2006 15:04"),
	)

	for _, adminID := range h.adminChatIDs {
		adminMsg := tgbotapi.NewMessage(adminID, adminText)
		adminMsg.ParseMode = "HTML"
		if _, err := h.bot.Send(adminMsg); err != nil {
			log.Printf("Error sending admin notification to %d: %v", adminID, err)
		}
	}
}

// handleConfirmPayment подтверждает оплату заказа
func (h *Handler) handleConfirmPayment(query *tgbotapi.CallbackQuery, orderIDStr string) {
	// Проверка что пользователь - админ
	if !h.isAdmin(query.From.ID) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Получаем заказ
	order, err := h.storage.GetOrderByID(ctx, orderIDStr)
	if err != nil {
		log.Printf("Error fetching order: %v", err)
		h.sendMessage(query.Message.Chat.ID, "❌ Заказ не найден.")
		return
	}

	// Обновляем статус
	if err := h.storage.UpdateOrderStatus(ctx, orderIDStr, "paid"); err != nil {
		log.Printf("Error updating order status: %v", err)
		h.sendMessage(query.Message.Chat.ID, "❌ Ошибка при обновлении статуса.")
		return
	}

	// Получаем информацию о товаре
	product, err := h.storage.GetProductByID(ctx, order.ProductID)
	if err != nil {
		log.Printf("Error fetching product: %v", err)
		return
	}

	// Уведомляем пользователя
	userText := fmt.Sprintf(
		"✅ <b>Оплата подтверждена!</b>\n\n"+
			"📦 Заказ №: <code>%s</code>\n"+
			"🎮 %s\n"+
			"💰 %.2f руб.\n\n"+
			"Ваша подписка активирована! Спасибо за покупку! 🎉",
		order.OrderID,
		product.Name,
		order.Price,
	)

	userMsg := tgbotapi.NewMessage(order.UserID, userText)
	userMsg.ParseMode = "HTML"
	if _, err := h.bot.Send(userMsg); err != nil {
		log.Printf("Error notifying user: %v", err)
	}

	// Подтверждаем админу
	h.sendMessage(query.Message.Chat.ID, fmt.Sprintf("✅ Оплата подтверждена для заказа %s", orderIDStr))

	log.Printf("Payment confirmed for order %s by admin %d", orderIDStr, query.From.ID)
}
