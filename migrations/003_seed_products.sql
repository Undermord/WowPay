-- Seed initial products
INSERT INTO products (name, duration_months, price, description) VALUES
    ('WoW Подписка 1 месяц', 1, 990.00, 'Игровое время World of Warcraft на 1 месяц. Доступ ко всем дополнениям, включая Dragonflight.'),
    ('WoW Подписка 2 месяца', 2, 1890.00, 'Игровое время World of Warcraft на 2 месяца. Доступ ко всем дополнениям. Выгода 90 рублей!'),
    ('WoW Подписка 6 месяцев', 6, 5490.00, 'Игровое время World of Warcraft на 6 месяцев. Доступ ко всем дополнениям. Выгода 450 рублей!'),
    ('WoW Подписка 12 месяцев', 12, 10490.00, 'Игровое время World of Warcraft на 12 месяцев. Доступ ко всем дополнениям. Максимальная выгода - 1390 рублей!')
ON CONFLICT DO NOTHING;
