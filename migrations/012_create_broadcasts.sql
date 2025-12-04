-- Таблица для истории рассылок
CREATE TABLE IF NOT EXISTS broadcasts (
    id SERIAL PRIMARY KEY,
    admin_id BIGINT NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    started_at TIMESTAMP,
    completed_at TIMESTAMP,
    status VARCHAR(50) DEFAULT 'draft',
    total_users INT DEFAULT 0,
    sent_count INT DEFAULT 0,
    failed_count INT DEFAULT 0
);

CREATE INDEX idx_broadcasts_status ON broadcasts(status);
CREATE INDEX idx_broadcasts_created_at ON broadcasts(created_at DESC);

COMMENT ON TABLE broadcasts IS 'История рассылок администраторов';
COMMENT ON COLUMN broadcasts.status IS 'Статус: draft, sending, completed, failed';
