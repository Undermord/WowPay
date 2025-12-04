package handlers

import "time"

// Timeouts and limits
const (
	DBContextTimeout     = 5 * time.Second
	RecentOrdersLimit    = 10
	DisplayedOrdersLimit = 5
)

// Callback action constants
const (
	CallbackActionRegion           = "region"
	CallbackActionCategory         = "category"
	CallbackActionProduct          = "product"
	CallbackActionBuy              = "buy"
	CallbackActionBack             = "back"
	CallbackActionConfirmPayment   = "confirm_payment"
	CallbackActionAdminEditPrice   = "admin_edit_price"
	CallbackActionAdminEditName    = "admin_edit_name"
	CallbackActionAdminEditDesc    = "admin_edit_desc"
	CallbackActionAdminToggleVis   = "admin_toggle_visibility"
	CallbackActionAdminProducts    = "admin_products"
	CallbackActionAdminEditProduct = "admin_edit_product"
	CallbackActionAdminEditWelcome = "admin_edit_welcome"
	CallbackActionShowProducts     = "show_products"
	CallbackActionChangeRegion     = "change_region"
	CallbackActionBroadcastMenu    = "broadcast_menu"
	CallbackActionBroadcastStart   = "broadcast_start"
	CallbackActionBroadcastConfirm = "broadcast_confirm"
	CallbackActionBroadcastCancel  = "broadcast_cancel"
	CallbackActionBackToAdmin      = "back_to_admin"
)

// Status emoji and text maps
var (
	StatusEmojis = map[string]string{
		"created":   "⏳",
		"paid":      "✅",
		"completed": "🎉",
		"cancelled": "❌",
	}

	StatusTexts = map[string]string{
		"created":   "Ожидает оплаты",
		"paid":      "Оплачен",
		"completed": "Завершен",
		"cancelled": "Отменён",
	}
)
