-- Таблица для хранения всех пользователей бота
CREATE TABLE IF NOT EXISTS users (
    user_id BIGINT PRIMARY KEY,
    username VARCHAR(255),
    first_name VARCHAR(255),
    last_name VARCHAR(255),
    is_blocked BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_is_blocked ON users(is_blocked);
CREATE INDEX idx_users_last_activity ON users(last_activity DESC);

COMMENT ON TABLE users IS 'Таблица всех пользователей бота для рассылок';
COMMENT ON COLUMN users.is_blocked IS 'Флаг: true если пользователь заблокировал бота';
