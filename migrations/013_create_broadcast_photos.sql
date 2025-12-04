-- Таблица для хранения file_id фотографий в рассылках
CREATE TABLE IF NOT EXISTS broadcast_photos (
    id SERIAL PRIMARY KEY,
    broadcast_id INT NOT NULL REFERENCES broadcasts(id) ON DELETE CASCADE,
    file_id VARCHAR(255) NOT NULL,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_broadcast_photos_broadcast_id ON broadcast_photos(broadcast_id);

COMMENT ON TABLE broadcast_photos IS 'Фотографии для рассылок (Telegram file_id)';
COMMENT ON COLUMN broadcast_photos.file_id IS 'Telegram file_id для повторного использования';
