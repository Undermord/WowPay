package ratelimit

import (
	"testing"
	"time"
)

func TestLimiter_Allow(t *testing.T) {
	config := Config{
		MaxRequests:     3,
		Window:          time.Second,
		BanDuration:     2 * time.Second,
		CleanupInterval: time.Minute,
	}

	limiter := NewLimiter(config)
	defer limiter.Stop()

	userID := int64(12345)

	// Первые 3 запроса должны пройти
	for i := 0; i < 3; i++ {
		if !limiter.Allow(userID) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 4-й запрос должен быть заблокирован
	if limiter.Allow(userID) {
		t.Error("4th request should be blocked")
	}

	// Проверяем, что пользователь заблокирован
	banned, duration := limiter.IsBanned(userID)
	if !banned {
		t.Error("User should be banned")
	}
	if duration <= 0 || duration > config.BanDuration {
		t.Errorf("Ban duration should be between 0 and %v, got %v", config.BanDuration, duration)
	}
}

func TestLimiter_WindowExpiry(t *testing.T) {
	config := Config{
		MaxRequests:     2,
		Window:          500 * time.Millisecond,
		BanDuration:     time.Second,
		CleanupInterval: time.Minute,
	}

	limiter := NewLimiter(config)
	defer limiter.Stop()

	userID := int64(12345)

	// Делаем 2 запроса (макс лимит)
	for i := 0; i < 2; i++ {
		if !limiter.Allow(userID) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 3-й запрос сразу - должен быть заблокирован
	if limiter.Allow(userID) {
		t.Error("3rd request should be blocked immediately")
	}

	// Ждем окончания блокировки
	time.Sleep(config.BanDuration + 100*time.Millisecond)

	// После блокировки должен пройти
	if !limiter.Allow(userID) {
		t.Error("Request after ban should be allowed")
	}
}

func TestLimiter_Reset(t *testing.T) {
	config := Config{
		MaxRequests:     2,
		Window:          time.Second,
		BanDuration:     time.Second,
		CleanupInterval: time.Minute,
	}

	limiter := NewLimiter(config)
	defer limiter.Stop()

	userID := int64(12345)

	// Исчерпываем лимит
	limiter.Allow(userID)
	limiter.Allow(userID)
	if limiter.Allow(userID) {
		t.Error("Should be blocked")
	}

	// Сбрасываем лимиты
	limiter.Reset(userID)

	// Теперь должно пройти
	if !limiter.Allow(userID) {
		t.Error("Request after reset should be allowed")
	}
}

func TestLimiter_GetStats(t *testing.T) {
	config := Config{
		MaxRequests:     5,
		Window:          time.Second,
		BanDuration:     time.Second,
		CleanupInterval: time.Minute,
	}

	limiter := NewLimiter(config)
	defer limiter.Stop()

	userID := int64(12345)

	// Делаем 3 запроса
	for i := 0; i < 3; i++ {
		limiter.Allow(userID)
	}

	count, remaining := limiter.GetStats(userID)
	if count != 3 {
		t.Errorf("Expected 3 requests, got %d", count)
	}
	if remaining != 2 {
		t.Errorf("Expected 2 remaining, got %d", remaining)
	}
}

func TestLimiter_MultipleUsers(t *testing.T) {
	config := Config{
		MaxRequests:     2,
		Window:          time.Second,
		BanDuration:     time.Second,
		CleanupInterval: time.Minute,
	}

	limiter := NewLimiter(config)
	defer limiter.Stop()

	user1 := int64(111)
	user2 := int64(222)

	// User1 исчерпывает лимит
	limiter.Allow(user1)
	limiter.Allow(user1)
	if limiter.Allow(user1) {
		t.Error("User1 should be blocked")
	}

	// User2 должен иметь свои лимиты
	if !limiter.Allow(user2) {
		t.Error("User2 should be allowed")
	}
	if !limiter.Allow(user2) {
		t.Error("User2 should be allowed")
	}
}

func TestLimiter_IsBanned(t *testing.T) {
	config := Config{
		MaxRequests:     1,
		Window:          time.Second,
		BanDuration:     500 * time.Millisecond,
		CleanupInterval: time.Minute,
	}

	limiter := NewLimiter(config)
	defer limiter.Stop()

	userID := int64(12345)

	// Пользователь не заблокирован изначально
	banned, _ := limiter.IsBanned(userID)
	if banned {
		t.Error("User should not be banned initially")
	}

	// Исчерпываем лимит
	limiter.Allow(userID)
	limiter.Allow(userID) // Это должно вызвать блокировку

	// Теперь заблокирован
	banned, duration := limiter.IsBanned(userID)
	if !banned {
		t.Error("User should be banned")
	}
	if duration <= 0 {
		t.Error("Ban duration should be positive")
	}

	// Ждем окончания блокировки
	time.Sleep(config.BanDuration + 100*time.Millisecond)

	// Больше не заблокирован
	banned, _ = limiter.IsBanned(userID)
	if banned {
		t.Error("User should not be banned after expiry")
	}
}
