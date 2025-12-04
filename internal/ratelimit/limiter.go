package ratelimit

import (
	"sync"
	"time"
)

// Limiter управляет ограничением скорости запросов
type Limiter struct {
	users      map[int64]*userLimit
	mu         sync.RWMutex
	maxReqs    int           // Максимум запросов в окно времени
	window     time.Duration // Размер окна времени
	banDuration time.Duration // Длительность блокировки при превышении
	cleanupInterval time.Duration
	stopCh     chan struct{}
}

// userLimit хранит информацию о запросах пользователя
type userLimit struct {
	requests  []time.Time
	bannedUntil time.Time
}

// Config содержит настройки rate limiter
type Config struct {
	MaxRequests     int           // Максимум запросов
	Window          time.Duration // Окно времени
	BanDuration     time.Duration // Время блокировки
	CleanupInterval time.Duration // Интервал очистки
}

// DefaultConfig возвращает стандартную конфигурацию
func DefaultConfig() Config {
	return Config{
		MaxRequests:     20,                // 20 запросов
		Window:          time.Minute,       // за минуту
		BanDuration:     5 * time.Minute,   // блокировка на 5 минут
		CleanupInterval: 10 * time.Minute,  // очистка каждые 10 минут
	}
}

// AdminConfig возвращает конфигурацию для администраторов (более мягкие лимиты)
func AdminConfig() Config {
	return Config{
		MaxRequests:     100,               // 100 запросов
		Window:          time.Minute,       // за минуту
		BanDuration:     time.Minute,       // блокировка на 1 минуту
		CleanupInterval: 10 * time.Minute,
	}
}

// NewLimiter создает новый rate limiter
func NewLimiter(config Config) *Limiter {
	l := &Limiter{
		users:      make(map[int64]*userLimit),
		maxReqs:    config.MaxRequests,
		window:     config.Window,
		banDuration: config.BanDuration,
		cleanupInterval: config.CleanupInterval,
		stopCh:     make(chan struct{}),
	}

	// Запускаем фоновую очистку
	go l.cleanup()

	return l
}

// Allow проверяет, разрешен ли запрос от пользователя
func (l *Limiter) Allow(userID int64) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Получаем или создаем запись пользователя
	user, exists := l.users[userID]
	if !exists {
		user = &userLimit{
			requests: []time.Time{},
		}
		l.users[userID] = user
	}

	// Проверяем, не заблокирован ли пользователь
	if now.Before(user.bannedUntil) {
		return false
	}

	// Очищаем устаревшие запросы (вне окна времени)
	cutoff := now.Add(-l.window)
	validRequests := []time.Time{}
	for _, reqTime := range user.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	user.requests = validRequests

	// Проверяем лимит
	if len(user.requests) >= l.maxReqs {
		// Превышен лимит - блокируем пользователя
		user.bannedUntil = now.Add(l.banDuration)
		return false
	}

	// Разрешаем запрос
	user.requests = append(user.requests, now)
	return true
}

// IsBanned проверяет, заблокирован ли пользователь
func (l *Limiter) IsBanned(userID int64) (bool, time.Duration) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	user, exists := l.users[userID]
	if !exists {
		return false, 0
	}

	now := time.Now()
	if now.Before(user.bannedUntil) {
		remaining := user.bannedUntil.Sub(now)
		return true, remaining
	}

	return false, 0
}

// Reset сбрасывает лимиты для пользователя (для администраторов)
func (l *Limiter) Reset(userID int64) {
	l.mu.Lock()
	defer l.mu.Unlock()

	delete(l.users, userID)
}

// cleanup периодически удаляет неактивных пользователей
func (l *Limiter) cleanup() {
	ticker := time.NewTicker(l.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			l.mu.Lock()
			now := time.Now()
			cutoff := now.Add(-l.cleanupInterval)

			for userID, user := range l.users {
				// Удаляем пользователей без недавней активности
				if len(user.requests) == 0 ||
				   (len(user.requests) > 0 && user.requests[len(user.requests)-1].Before(cutoff) &&
				    user.bannedUntil.Before(now)) {
					delete(l.users, userID)
				}
			}
			l.mu.Unlock()
		case <-l.stopCh:
			return
		}
	}
}

// Stop останавливает фоновую очистку
func (l *Limiter) Stop() {
	close(l.stopCh)
}

// GetStats возвращает статистику использования
func (l *Limiter) GetStats(userID int64) (requestCount int, remaining int) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	user, exists := l.users[userID]
	if !exists {
		return 0, l.maxReqs
	}

	now := time.Now()
	cutoff := now.Add(-l.window)

	// Считаем активные запросы
	activeCount := 0
	for _, reqTime := range user.requests {
		if reqTime.After(cutoff) {
			activeCount++
		}
	}

	remaining = l.maxReqs - activeCount
	if remaining < 0 {
		remaining = 0
	}

	return activeCount, remaining
}
