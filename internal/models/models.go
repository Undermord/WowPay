package models

import (
	"time"
)

type Region struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"created_at"`
}

type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	RegionID    int       `json:"region_id"`
	Description string    `json:"description"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	CategoryID  int       `json:"category_id"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	IsVisible   bool      `json:"is_visible"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}

type Order struct {
	OrderID   string    `json:"order_id"`
	UserID    int64     `json:"user_id"`
	ProductID int       `json:"product_id"`
	Price     float64   `json:"price"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type BotSettings struct {
	ID             int       `json:"id"`
	WelcomeMessage string    `json:"welcome_message"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// User представляет пользователя бота
type User struct {
	UserID       int64     `json:"user_id"`
	Username     string    `json:"username"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	IsBlocked    bool      `json:"is_blocked"`
	CreatedAt    time.Time `json:"created_at"`
	LastActivity time.Time `json:"last_activity"`
}

// Broadcast представляет рассылку
type Broadcast struct {
	ID          int        `json:"id"`
	AdminID     int64      `json:"admin_id"`
	Text        string     `json:"text"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Status      string     `json:"status"` // draft, sending, completed, failed
	TotalUsers  int        `json:"total_users"`
	SentCount   int        `json:"sent_count"`
	FailedCount int        `json:"failed_count"`
}

// BroadcastPhoto представляет фотографию для рассылки
type BroadcastPhoto struct {
	ID          int       `json:"id"`
	BroadcastID int       `json:"broadcast_id"`
	FileID      string    `json:"file_id"`
	SortOrder   int       `json:"sort_order"`
	CreatedAt   time.Time `json:"created_at"`
}
