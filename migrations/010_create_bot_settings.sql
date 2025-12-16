-- Create bot_settings table
CREATE TABLE IF NOT EXISTS bot_settings (
    id SERIAL PRIMARY KEY,
    welcome_message TEXT NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert default welcome message
INSERT INTO bot_settings (id, welcome_message, updated_at) VALUES (
    1,
    '👋 Добро пожаловать, {name}!

🎮 Я бот для продажи игровых подписок World of Warcraft.',
    CURRENT_TIMESTAMP
) ON CONFLICT (id) DO NOTHING;
