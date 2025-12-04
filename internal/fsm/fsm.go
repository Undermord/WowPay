package fsm

import (
	"sync"
	"time"
)

// State представляет состояние пользователя в диалоге
type State string

const (
	StateNone                   State = ""
	StateWaitingForPrice        State = "waiting_for_price"
	StateWaitingForName         State = "waiting_for_name"
	StateWaitingForDesc         State = "waiting_for_description"
	StateWaitingForCategoryDesc State = "waiting_for_category_description"
	StateWaitingForWelcomeMsg   State = "waiting_for_welcome_message"
	// Broadcast FSM states
	StateWaitingForBroadcastText  State = "waiting_for_broadcast_text"
	StateWaitingForBroadcastPhoto State = "waiting_for_broadcast_photo"
	StateConfirmingBroadcast      State = "confirming_broadcast"
)

const (
	// StateTTL - время жизни состояния (30 минут)
	StateTTL = 30 * time.Minute
	// CleanupInterval - интервал очистки устаревших состояний (5 минут)
	CleanupInterval = 5 * time.Minute
)

// UserState хранит состояние пользователя
type UserState struct {
	State      State
	ProductID  int
	CategoryID int
	Data       map[string]interface{}
	ExpiresAt  time.Time // TTL для автоматической очистки
}

// Manager управляет состояниями пользователей
type Manager struct {
	states map[int64]*UserState
	mu     sync.RWMutex
	stopCh chan struct{}
}

// NewManager создает новый FSM менеджер с автоматической очисткой
func NewManager() *Manager {
	m := &Manager{
		states: make(map[int64]*UserState),
		stopCh: make(chan struct{}),
	}

	// Запускаем фоновую очистку устаревших состояний
	go m.cleanupExpired()

	return m
}

// cleanupExpired периодически удаляет устаревшие состояния
func (m *Manager) cleanupExpired() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for userID, state := range m.states {
				if now.After(state.ExpiresAt) {
					delete(m.states, userID)
				}
			}
			m.mu.Unlock()
		case <-m.stopCh:
			return
		}
	}
}

// Stop останавливает фоновую очистку
func (m *Manager) Stop() {
	close(m.stopCh)
}

// SetState устанавливает состояние для пользователя с TTL
func (m *Manager) SetState(userID int64, state State, productID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[userID] = &UserState{
		State:     state,
		ProductID: productID,
		Data:      make(map[string]interface{}),
		ExpiresAt: time.Now().Add(StateTTL),
	}
}

// SetCategoryState устанавливает состояние для редактирования категории с TTL
func (m *Manager) SetCategoryState(userID int64, state State, categoryID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.states[userID] = &UserState{
		State:      state,
		CategoryID: categoryID,
		Data:       make(map[string]interface{}),
		ExpiresAt:  time.Now().Add(StateTTL),
	}
}

// GetState возвращает состояние пользователя (проверяет TTL)
func (m *Manager) GetState(userID int64) (*UserState, bool) {
	m.mu.RLock()
	state, exists := m.states[userID]
	m.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Проверяем, не истекло ли время жизни состояния
	if time.Now().After(state.ExpiresAt) {
		m.ClearState(userID)
		return nil, false
	}

	return state, true
}

// ClearState очищает состояние пользователя
func (m *Manager) ClearState(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.states, userID)
}

// IsInState проверяет, находится ли пользователь в определенном состоянии
func (m *Manager) IsInState(userID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.states[userID]
	return exists && state.State != StateNone
}

// SetBroadcastState устанавливает состояние для создания рассылки с сохранением данных
func (m *Manager) SetBroadcastState(userID int64, state State, broadcastID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, exists := m.states[userID]
	if exists && existing.State != StateNone {
		// Сохраняем существующие данные broadcast
		existing.State = state
		existing.ExpiresAt = time.Now().Add(StateTTL)
	} else {
		m.states[userID] = &UserState{
			State:     state,
			ProductID: broadcastID, // Переиспользуем поле для broadcast_id
			Data:      make(map[string]interface{}),
			ExpiresAt: time.Now().Add(StateTTL),
		}
	}
}
