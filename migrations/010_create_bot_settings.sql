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

🎮 Я бот для продажи игровых подписок World of Warcraft.


Впервые покупаешь в телеграмме? Тогда обязательно ознакомься 👇🏻

 - <a href="https://teletype.in/@grek_blckdl/_O-0ZkNtk_W">Инструкция как пополнить бота и как приобрести товар</a>
 - <a href="https://teletype.in/@grek_blckdl/DiN9Ag8IsNL">Как найти инструкцию по товару</a>
 - <a href="https://teletype.in/@grek_blckdl/z_lyOXdW7k0">Как правильно обратиться в поддержку</a>

<a href="https://t.me/wowpaysupp">Поддержка</a> | Время работы 10:00-22:00

- <a href="https://teletype.in/@grek_blckdl/FY8EkS5Wen3">Полезные инструкции</a> 👇🏻Настоятельно рекомендуем ознакомиться!',
    CURRENT_TIMESTAMP
) ON CONFLICT (id) DO NOTHING;
