package storage

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"tgwow/internal/models"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, databaseURL string) (*PostgresStorage, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresStorage{pool: pool}, nil
}

func (s *PostgresStorage) Close() {
	s.pool.Close()
}

// generateOrderID генерирует короткий номер заказа формата: WOW + YYMMDD + 3 цифры
// Пример: WOW241204123
func generateOrderID() string {
	now := time.Now()
	dateStr := now.Format("060102") // YYMMDD
	randomNum := rand.Intn(1000)    // 0-999
	return fmt.Sprintf("WOW%s%03d", dateStr, randomNum)
}

// ListRegions возвращает все регионы
func (s *PostgresStorage) ListRegions(ctx context.Context) ([]models.Region, error) {
	query := `
		SELECT id, name, code, created_at
		FROM regions
		ORDER BY id ASC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query regions: %w", err)
	}
	defer rows.Close()

	var regions []models.Region
	for rows.Next() {
		var r models.Region
		if err := rows.Scan(&r.ID, &r.Name, &r.Code, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan region: %w", err)
		}
		regions = append(regions, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return regions, nil
}

// GetRegionByID возвращает регион по ID
func (s *PostgresStorage) GetRegionByID(ctx context.Context, regionID int) (*models.Region, error) {
	query := `
		SELECT id, name, code, created_at
		FROM regions
		WHERE id = $1
	`

	var r models.Region
	err := s.pool.QueryRow(ctx, query, regionID).Scan(&r.ID, &r.Name, &r.Code, &r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get region: %w", err)
	}

	return &r, nil
}

// ListCategoriesByRegion возвращает категории для региона
func (s *PostgresStorage) ListCategoriesByRegion(ctx context.Context, regionID int) ([]models.Category, error) {
	query := `
		SELECT id, name, region_id, description, sort_order, created_at
		FROM categories
		WHERE region_id = $1 AND name != 'Системные услуги'
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := s.pool.Query(ctx, query, regionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.RegionID, &c.Description, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return categories, nil
}

// ListAllCategoriesByRegion возвращает все категории для региона (включая системные) - для админа
func (s *PostgresStorage) ListAllCategoriesByRegion(ctx context.Context, regionID int) ([]models.Category, error) {
	query := `
		SELECT id, name, region_id, description, sort_order, created_at
		FROM categories
		WHERE region_id = $1
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := s.pool.Query(ctx, query, regionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.RegionID, &c.Description, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return categories, nil
}

// ListAllCategories возвращает все категории (для batch-загрузки в админке)
func (s *PostgresStorage) ListAllCategories(ctx context.Context) ([]models.Category, error) {
	query := `
		SELECT id, name, region_id, description, sort_order, created_at
		FROM categories
		ORDER BY region_id ASC, sort_order ASC, id ASC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.RegionID, &c.Description, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return categories, nil
}

// GetCategoryByID возвращает категорию по ID
func (s *PostgresStorage) GetCategoryByID(ctx context.Context, categoryID int) (*models.Category, error) {
	query := `
		SELECT id, name, region_id, description, sort_order, created_at
		FROM categories
		WHERE id = $1
	`

	var c models.Category
	err := s.pool.QueryRow(ctx, query, categoryID).Scan(
		&c.ID, &c.Name, &c.RegionID, &c.Description, &c.SortOrder, &c.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return &c, nil
}

// ListProductsByCategory возвращает товары для категории
func (s *PostgresStorage) ListProductsByCategory(ctx context.Context, categoryID int) ([]models.Product, error) {
	query := `
		SELECT id, name, category_id, price, description, is_visible, sort_order, created_at
		FROM products
		WHERE category_id = $1 AND is_visible = true
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := s.pool.Query(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.CategoryID, &p.Price, &p.Description, &p.IsVisible, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

// ListAllProductsByCategory возвращает все товары для категории (включая скрытые) - для админа
func (s *PostgresStorage) ListAllProductsByCategory(ctx context.Context, categoryID int) ([]models.Product, error) {
	query := `
		SELECT id, name, category_id, price, description, is_visible, sort_order, created_at
		FROM products
		WHERE category_id = $1
		ORDER BY sort_order ASC, id ASC
	`

	rows, err := s.pool.Query(ctx, query, categoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.CategoryID, &p.Price, &p.Description, &p.IsVisible, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

// ListProducts возвращает все видимые товары (для совместимости)
func (s *PostgresStorage) ListProducts(ctx context.Context) ([]models.Product, error) {
	query := `
		SELECT id, name, category_id, price, description, is_visible, sort_order, created_at
		FROM products
		WHERE is_visible = true
		ORDER BY sort_order ASC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.CategoryID, &p.Price, &p.Description, &p.IsVisible, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

func (s *PostgresStorage) GetProductByID(ctx context.Context, productID int) (*models.Product, error) {
	query := `
		SELECT id, name, category_id, price, description, is_visible, sort_order, created_at
		FROM products
		WHERE id = $1
	`

	var p models.Product
	err := s.pool.QueryRow(ctx, query, productID).Scan(
		&p.ID, &p.Name, &p.CategoryID, &p.Price, &p.Description, &p.IsVisible, &p.SortOrder, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get product: %w", err)
	}

	return &p, nil
}

// GetProductsByIDs возвращает товары по списку ID (для решения N+1 проблемы)
func (s *PostgresStorage) GetProductsByIDs(ctx context.Context, productIDs []int) (map[int]*models.Product, error) {
	if len(productIDs) == 0 {
		return make(map[int]*models.Product), nil
	}

	query := `
		SELECT id, name, category_id, price, description, is_visible, sort_order, created_at
		FROM products
		WHERE id = ANY($1)
	`

	rows, err := s.pool.Query(ctx, query, productIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query products: %w", err)
	}
	defer rows.Close()

	products := make(map[int]*models.Product, len(productIDs))
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.CategoryID, &p.Price, &p.Description, &p.IsVisible, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products[p.ID] = &p
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

func (s *PostgresStorage) CreateOrder(ctx context.Context, userID int64, productID int, price float64) (*models.Order, error) {
	orderID := generateOrderID()
	createdAt := time.Now()

	query := `
		INSERT INTO orders (order_id, user_id, product_id, price, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING order_id, user_id, product_id, price, status, created_at
	`

	var order models.Order
	err := s.pool.QueryRow(
		ctx, query,
		orderID, userID, productID, price, "created", createdAt,
	).Scan(
		&order.OrderID, &order.UserID, &order.ProductID,
		&order.Price, &order.Status, &order.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &order, nil
}

// GetUserOrders возвращает заказы пользователя
func (s *PostgresStorage) GetUserOrders(ctx context.Context, userID int64) ([]models.Order, error) {
	query := `
		SELECT order_id, user_id, product_id, price, status, created_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.OrderID, &o.UserID, &o.ProductID, &o.Price, &o.Status, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// GetOrderByID возвращает заказ по ID
func (s *PostgresStorage) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	query := `
		SELECT order_id, user_id, product_id, price, status, created_at
		FROM orders
		WHERE order_id = $1
	`

	var o models.Order
	err := s.pool.QueryRow(ctx, query, orderID).Scan(
		&o.OrderID, &o.UserID, &o.ProductID, &o.Price, &o.Status, &o.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &o, nil
}

// UpdateOrderStatus обновляет статус заказа
func (s *PostgresStorage) UpdateOrderStatus(ctx context.Context, orderID string, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE order_id = $3
	`

	_, err := s.pool.Exec(ctx, query, status, time.Now(), orderID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

// GetRecentOrders возвращает последние заказы (для админа)
func (s *PostgresStorage) GetRecentOrders(ctx context.Context, limit int) ([]models.Order, error) {
	query := `
		SELECT order_id, user_id, product_id, price, status, created_at
		FROM orders
		ORDER BY created_at DESC
		LIMIT $1
	`

	rows, err := s.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query recent orders: %w", err)
	}
	defer rows.Close()

	var orders []models.Order
	for rows.Next() {
		var o models.Order
		if err := rows.Scan(&o.OrderID, &o.UserID, &o.ProductID, &o.Price, &o.Status, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return orders, nil
}

// GetOrderStats возвращает статистику заказов
func (s *PostgresStorage) GetOrderStats(ctx context.Context) (map[string]interface{}, error) {
	query := `
		SELECT
			COUNT(*) as total_orders,
			COUNT(CASE WHEN status = 'created' THEN 1 END) as pending_orders,
			COUNT(CASE WHEN status = 'paid' THEN 1 END) as paid_orders,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_orders,
			COALESCE(SUM(CASE WHEN status IN ('paid', 'completed') THEN price ELSE 0 END), 0) as total_revenue
		FROM orders
	`

	var totalOrders, pendingOrders, paidOrders, completedOrders int
	var totalRevenue float64

	err := s.pool.QueryRow(ctx, query).Scan(
		&totalOrders, &pendingOrders, &paidOrders, &completedOrders, &totalRevenue,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get order stats: %w", err)
	}

	stats := map[string]interface{}{
		"total_orders":     totalOrders,
		"pending_orders":   pendingOrders,
		"paid_orders":      paidOrders,
		"completed_orders": completedOrders,
		"total_revenue":    totalRevenue,
	}

	return stats, nil
}

// Admin methods for managing catalog

// UpdateProductPrice обновляет цену товара
func (s *PostgresStorage) UpdateProductPrice(ctx context.Context, productID int, newPrice float64) error {
	query := `
		UPDATE products
		SET price = $1
		WHERE id = $2
	`

	_, err := s.pool.Exec(ctx, query, newPrice, productID)
	if err != nil {
		return fmt.Errorf("failed to update product price: %w", err)
	}

	return nil
}

// UpdateProductVisibility изменяет видимость товара
func (s *PostgresStorage) UpdateProductVisibility(ctx context.Context, productID int, isVisible bool) error {
	query := `
		UPDATE products
		SET is_visible = $1
		WHERE id = $2
	`

	_, err := s.pool.Exec(ctx, query, isVisible, productID)
	if err != nil {
		return fmt.Errorf("failed to update product visibility: %w", err)
	}

	return nil
}

// CreateProduct создает новый товар
func (s *PostgresStorage) CreateProduct(ctx context.Context, name string, categoryID int, price float64, description string) (*models.Product, error) {
	query := `
		INSERT INTO products (name, category_id, price, description, is_visible, sort_order)
		VALUES ($1, $2, $3, $4, true, 0)
		RETURNING id, name, category_id, price, description, is_visible, sort_order, created_at
	`

	var p models.Product
	err := s.pool.QueryRow(ctx, query, name, categoryID, price, description).Scan(
		&p.ID, &p.Name, &p.CategoryID, &p.Price, &p.Description, &p.IsVisible, &p.SortOrder, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create product: %w", err)
	}

	return &p, nil
}

// DeleteProduct удаляет товар
func (s *PostgresStorage) DeleteProduct(ctx context.Context, productID int) error {
	query := `
		DELETE FROM products
		WHERE id = $1
	`

	_, err := s.pool.Exec(ctx, query, productID)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// UpdateProduct обновляет информацию о товаре
func (s *PostgresStorage) UpdateProduct(ctx context.Context, productID int, name string, price float64, description string) error {
	query := `
		UPDATE products
		SET name = $1, price = $2, description = $3
		WHERE id = $4
	`

	_, err := s.pool.Exec(ctx, query, name, price, description, productID)
	if err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// ListAllProducts возвращает все товары (включая скрытые) для админа
func (s *PostgresStorage) ListAllProducts(ctx context.Context) ([]models.Product, error) {
	query := `
		SELECT id, name, category_id, price, description, is_visible, sort_order, created_at
		FROM products
		ORDER BY category_id ASC, sort_order ASC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all products: %w", err)
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name, &p.CategoryID, &p.Price, &p.Description, &p.IsVisible, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return products, nil
}

// Bot settings methods

// GetBotSettings возвращает настройки бота
func (s *PostgresStorage) GetBotSettings(ctx context.Context) (*models.BotSettings, error) {
	query := `
		SELECT id, welcome_message, updated_at
		FROM bot_settings
		WHERE id = 1
	`

	var settings models.BotSettings
	err := s.pool.QueryRow(ctx, query).Scan(
		&settings.ID, &settings.WelcomeMessage, &settings.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get bot settings: %w", err)
	}

	return &settings, nil
}

// UpdateWelcomeMessage обновляет приветственное сообщение
func (s *PostgresStorage) UpdateWelcomeMessage(ctx context.Context, message string) error {
	query := `
		UPDATE bot_settings
		SET welcome_message = $1, updated_at = $2
		WHERE id = 1
	`

	_, err := s.pool.Exec(ctx, query, message, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update welcome message: %w", err)
	}

	return nil
}

// GetChangeRegionProduct возвращает товар "Сменить регион" из категории "Системные услуги"
func (s *PostgresStorage) GetChangeRegionProduct(ctx context.Context) (*models.Product, error) {
	query := `
		SELECT p.id, p.name, p.category_id, p.price, p.description, p.is_visible, p.sort_order, p.created_at
		FROM products p
		JOIN categories c ON p.category_id = c.id
		WHERE c.name = 'Системные услуги' AND p.name = 'Сменить регион'
		LIMIT 1
	`

	var p models.Product
	err := s.pool.QueryRow(ctx, query).Scan(
		&p.ID, &p.Name, &p.CategoryID, &p.Price, &p.Description, &p.IsVisible, &p.SortOrder, &p.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get change region product: %w", err)
	}

	return &p, nil
}

// ==================== USER METHODS ====================

// UpsertUser создает или обновляет пользователя
func (s *PostgresStorage) UpsertUser(ctx context.Context, userID int64, username, firstName, lastName string) error {
	query := `
		INSERT INTO users (user_id, username, first_name, last_name, last_activity)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id) DO UPDATE SET
			username = EXCLUDED.username,
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			last_activity = EXCLUDED.last_activity
	`

	_, err := s.pool.Exec(ctx, query, userID, username, firstName, lastName, time.Now())
	if err != nil {
		return fmt.Errorf("failed to upsert user: %w", err)
	}

	return nil
}

// GetActiveUsers возвращает всех пользователей, которые не заблокировали бота
func (s *PostgresStorage) GetActiveUsers(ctx context.Context) ([]models.User, error) {
	query := `
		SELECT user_id, username, first_name, last_name, is_blocked, created_at, last_activity
		FROM users
		WHERE is_blocked = false
		ORDER BY user_id ASC
	`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.UserID, &u.Username, &u.FirstName, &u.LastName, &u.IsBlocked, &u.CreatedAt, &u.LastActivity); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return users, nil
}

// MarkUserAsBlocked помечает пользователя как заблокировавшего бота
func (s *PostgresStorage) MarkUserAsBlocked(ctx context.Context, userID int64) error {
	query := `UPDATE users SET is_blocked = true WHERE user_id = $1`

	_, err := s.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to mark user as blocked: %w", err)
	}

	return nil
}

// GetUsersCount возвращает количество активных пользователей
func (s *PostgresStorage) GetUsersCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM users WHERE is_blocked = false`

	var count int
	err := s.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get users count: %w", err)
	}

	return count, nil
}

// ==================== BROADCAST METHODS ====================

// CreateBroadcast создает новую рассылку (статус draft)
func (s *PostgresStorage) CreateBroadcast(ctx context.Context, adminID int64, text string) (*models.Broadcast, error) {
	query := `
		INSERT INTO broadcasts (admin_id, text, status)
		VALUES ($1, $2, 'draft')
		RETURNING id, admin_id, text, created_at, started_at, completed_at, status, total_users, sent_count, failed_count
	`

	var b models.Broadcast
	err := s.pool.QueryRow(ctx, query, adminID, text).Scan(
		&b.ID, &b.AdminID, &b.Text, &b.CreatedAt, &b.StartedAt, &b.CompletedAt,
		&b.Status, &b.TotalUsers, &b.SentCount, &b.FailedCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create broadcast: %w", err)
	}

	return &b, nil
}

// SaveBroadcastPhoto сохраняет file_id фотографии для рассылки
func (s *PostgresStorage) SaveBroadcastPhoto(ctx context.Context, broadcastID int, fileID string, sortOrder int) error {
	query := `INSERT INTO broadcast_photos (broadcast_id, file_id, sort_order) VALUES ($1, $2, $3)`

	_, err := s.pool.Exec(ctx, query, broadcastID, fileID, sortOrder)
	if err != nil {
		return fmt.Errorf("failed to save broadcast photo: %w", err)
	}

	return nil
}

// GetBroadcastPhotos возвращает все фотографии рассылки
func (s *PostgresStorage) GetBroadcastPhotos(ctx context.Context, broadcastID int) ([]models.BroadcastPhoto, error) {
	query := `
		SELECT id, broadcast_id, file_id, sort_order, created_at
		FROM broadcast_photos
		WHERE broadcast_id = $1
		ORDER BY sort_order ASC
	`

	rows, err := s.pool.Query(ctx, query, broadcastID)
	if err != nil {
		return nil, fmt.Errorf("failed to query broadcast photos: %w", err)
	}
	defer rows.Close()

	var photos []models.BroadcastPhoto
	for rows.Next() {
		var p models.BroadcastPhoto
		if err := rows.Scan(&p.ID, &p.BroadcastID, &p.FileID, &p.SortOrder, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan photo: %w", err)
		}
		photos = append(photos, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return photos, nil
}

// UpdateBroadcastStatus обновляет статус рассылки и статистику
func (s *PostgresStorage) UpdateBroadcastStatus(ctx context.Context, broadcastID int, status string, totalUsers, sentCount, failedCount int) error {
	query := `
		UPDATE broadcasts
		SET status = $1, total_users = $2, sent_count = $3, failed_count = $4,
			started_at = CASE WHEN status = 'draft' AND $1 = 'sending' THEN NOW() ELSE started_at END,
			completed_at = CASE WHEN $1 IN ('completed', 'failed') THEN NOW() ELSE completed_at END
		WHERE id = $5
	`

	_, err := s.pool.Exec(ctx, query, status, totalUsers, sentCount, failedCount, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to update broadcast status: %w", err)
	}

	return nil
}

// GetBroadcastByID возвращает рассылку по ID
func (s *PostgresStorage) GetBroadcastByID(ctx context.Context, broadcastID int) (*models.Broadcast, error) {
	query := `
		SELECT id, admin_id, text, created_at, started_at, completed_at, status, total_users, sent_count, failed_count
		FROM broadcasts
		WHERE id = $1
	`

	var b models.Broadcast
	err := s.pool.QueryRow(ctx, query, broadcastID).Scan(
		&b.ID, &b.AdminID, &b.Text, &b.CreatedAt, &b.StartedAt, &b.CompletedAt,
		&b.Status, &b.TotalUsers, &b.SentCount, &b.FailedCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get broadcast: %w", err)
	}

	return &b, nil
}

// DeleteBroadcast удаляет рассылку (только в статусе draft)
func (s *PostgresStorage) DeleteBroadcast(ctx context.Context, broadcastID int) error {
	query := `DELETE FROM broadcasts WHERE id = $1 AND status = 'draft'`

	result, err := s.pool.Exec(ctx, query, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to delete broadcast: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("broadcast not found or not in draft status")
	}

	return nil
}
