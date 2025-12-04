-- Seed products for all regions and categories

-- WoW KZ - Подписка WOW
INSERT INTO products (name, category_id, price, description, sort_order)
SELECT '1 месяц', c.id, 670.00, 'Игровое время World of Warcraft на 1 месяц. Доступ ко всем дополнениям.', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '3 месяца', c.id, 1900.00, 'Игровое время World of Warcraft на 3 месяца. Выгодное предложение!', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '6 месяцев', c.id, 3369.00, 'Игровое время World of Warcraft на 6 месяцев. Максимальная выгода!', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '12 месяцев', c.id, 6729.00, 'Игровое время World of Warcraft на 12 месяцев. Лучшая цена!', 4
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Подписка WOW'

-- WoW KZ - Midnight
UNION ALL
SELECT 'Base Edition', c.id, 5299.00, 'Базовое издание World of Warcraft: Midnight', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Midnight'
UNION ALL
SELECT 'Heroic Edition', c.id, 7349.00, 'Героическое издание World of Warcraft: Midnight с дополнительными бонусами', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Midnight'
UNION ALL
SELECT 'Epic Edition', c.id, 9299.00, 'Эпическое издание World of Warcraft: Midnight с эксклюзивными наградами', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Midnight'

-- WoW KZ - The War Within
UNION ALL
SELECT 'The War Within', c.id, 1749.00, 'Дополнение World of Warcraft: The War Within', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'The War Within'

-- WoW KZ - Услуги WOW
UNION ALL
SELECT 'Повышение до 80 уровня', c.id, 3459.00, 'Мгновенное повышение персонажа до 80 уровня', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Перенос персонажа', c.id, 1749.00, 'Перенос персонажа на другой сервер', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена фракции', c.id, 2099.00, 'Изменение фракции персонажа', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена расы', c.id, 1749.00, 'Изменение расы персонажа', 4
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена имени', c.id, 699.00, 'Изменение имени персонажа', 5
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Услуги WOW'

-- WoW UA - Подписка WOW (цены пока 0)
UNION ALL
SELECT '1 месяц', c.id, 0.00, 'Игровое время World of Warcraft на 1 месяц. Доступ ко всем дополнениям.', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '3 месяца', c.id, 0.00, 'Игровое время World of Warcraft на 3 месяца. Выгодное предложение!', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '6 месяцев', c.id, 0.00, 'Игровое время World of Warcraft на 6 месяцев. Максимальная выгода!', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '12 месяцев', c.id, 0.00, 'Игровое время World of Warcraft на 12 месяцев. Лучшая цена!', 4
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Подписка WOW'

-- WoW UA - Midnight (цены пока 0)
UNION ALL
SELECT 'Base Edition', c.id, 0.00, 'Базовое издание World of Warcraft: Midnight', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Midnight'
UNION ALL
SELECT 'Heroic Edition', c.id, 0.00, 'Героическое издание World of Warcraft: Midnight с дополнительными бонусами', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Midnight'
UNION ALL
SELECT 'Epic Edition', c.id, 0.00, 'Эпическое издание World of Warcraft: Midnight с эксклюзивными наградами', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Midnight'

-- WoW UA - The War Within (цены пока 0)
UNION ALL
SELECT 'The War Within', c.id, 0.00, 'Дополнение World of Warcraft: The War Within', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'The War Within'

-- WoW UA - Услуги WOW (цены пока 0)
UNION ALL
SELECT 'Повышение до 80 уровня', c.id, 0.00, 'Мгновенное повышение персонажа до 80 уровня', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Перенос персонажа', c.id, 0.00, 'Перенос персонажа на другой сервер', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена фракции', c.id, 0.00, 'Изменение фракции персонажа', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена расы', c.id, 0.00, 'Изменение расы персонажа', 4
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена имени', c.id, 0.00, 'Изменение имени персонажа', 5
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'UA' AND c.name = 'Услуги WOW'

-- WoW EU - Подписка WOW (цены пока 0)
UNION ALL
SELECT '1 месяц', c.id, 0.00, 'Игровое время World of Warcraft на 1 месяц. Доступ ко всем дополнениям.', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '3 месяца', c.id, 0.00, 'Игровое время World of Warcraft на 3 месяца. Выгодное предложение!', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '6 месяцев', c.id, 0.00, 'Игровое время World of Warcraft на 6 месяцев. Максимальная выгода!', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '12 месяцев', c.id, 0.00, 'Игровое время World of Warcraft на 12 месяцев. Лучшая цена!', 4
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Подписка WOW'

-- WoW EU - Midnight (цены пока 0)
UNION ALL
SELECT 'Base Edition', c.id, 0.00, 'Базовое издание World of Warcraft: Midnight', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Midnight'
UNION ALL
SELECT 'Heroic Edition', c.id, 0.00, 'Героическое издание World of Warcraft: Midnight с дополнительными бонусами', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Midnight'
UNION ALL
SELECT 'Epic Edition', c.id, 0.00, 'Эпическое издание World of Warcraft: Midnight с эксклюзивными наградами', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Midnight'

-- WoW EU - The War Within (цены пока 0)
UNION ALL
SELECT 'The War Within', c.id, 0.00, 'Дополнение World of Warcraft: The War Within', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'The War Within'

-- WoW EU - Услуги WOW (цены пока 0)
UNION ALL
SELECT 'Повышение до 80 уровня', c.id, 0.00, 'Мгновенное повышение персонажа до 80 уровня', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Перенос персонажа', c.id, 0.00, 'Перенос персонажа на другой сервер', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена фракции', c.id, 0.00, 'Изменение фракции персонажа', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена расы', c.id, 0.00, 'Изменение расы персонажа', 4
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена имени', c.id, 0.00, 'Изменение имени персонажа', 5
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'EU' AND c.name = 'Услуги WOW'

-- WoW TUR - Подписка WOW (цены пока 0)
UNION ALL
SELECT '1 месяц', c.id, 0.00, 'Игровое время World of Warcraft на 1 месяц. Доступ ко всем дополнениям.', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '3 месяца', c.id, 0.00, 'Игровое время World of Warcraft на 3 месяца. Выгодное предложение!', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '6 месяцев', c.id, 0.00, 'Игровое время World of Warcraft на 6 месяцев. Максимальная выгода!', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Подписка WOW'
UNION ALL
SELECT '12 месяцев', c.id, 0.00, 'Игровое время World of Warcraft на 12 месяцев. Лучшая цена!', 4
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Подписка WOW'

-- WoW TUR - Midnight (цены пока 0)
UNION ALL
SELECT 'Base Edition', c.id, 0.00, 'Базовое издание World of Warcraft: Midnight', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Midnight'
UNION ALL
SELECT 'Heroic Edition', c.id, 0.00, 'Героическое издание World of Warcraft: Midnight с дополнительными бонусами', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Midnight'
UNION ALL
SELECT 'Epic Edition', c.id, 0.00, 'Эпическое издание World of Warcraft: Midnight с эксклюзивными наградами', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Midnight'

-- WoW TUR - The War Within (цены пока 0)
UNION ALL
SELECT 'The War Within', c.id, 0.00, 'Дополнение World of Warcraft: The War Within', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'The War Within'

-- WoW TUR - Услуги WOW (цены пока 0)
UNION ALL
SELECT 'Повышение до 80 уровня', c.id, 0.00, 'Мгновенное повышение персонажа до 80 уровня', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Перенос персонажа', c.id, 0.00, 'Перенос персонажа на другой сервер', 2
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена фракции', c.id, 0.00, 'Изменение фракции персонажа', 3
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена расы', c.id, 0.00, 'Изменение расы персонажа', 4
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Услуги WOW'
UNION ALL
SELECT 'Смена имени', c.id, 0.00, 'Изменение имени персонажа', 5
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'TUR' AND c.name = 'Услуги WOW'

-- Системный товар "Сменить регион"
UNION ALL
SELECT 'Сменить регион', c.id, 1749.00, 'Смена региона аккаунта World of Warcraft. Услуга включает полный перенос вашего аккаунта на другой игровой регион.', 1
FROM categories c
JOIN regions r ON c.region_id = r.id
WHERE r.code = 'KZ' AND c.name = 'Системные услуги';
