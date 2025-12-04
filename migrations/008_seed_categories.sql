-- Seed categories for all regions
-- WoW KZ categories
INSERT INTO categories (name, region_id, description, sort_order)
SELECT 'Подписка WOW', id, '', 1 FROM regions WHERE code = 'KZ'
UNION ALL
SELECT 'Midnight', id, '', 2 FROM regions WHERE code = 'KZ'
UNION ALL
SELECT 'The War Within', id, '', 3 FROM regions WHERE code = 'KZ'
UNION ALL
SELECT 'Услуги WOW', id, '', 4 FROM regions WHERE code = 'KZ'

-- WoW UA categories
UNION ALL
SELECT 'Подписка WOW', id, '', 1 FROM regions WHERE code = 'UA'
UNION ALL
SELECT 'Midnight', id, '', 2 FROM regions WHERE code = 'UA'
UNION ALL
SELECT 'The War Within', id, '', 3 FROM regions WHERE code = 'UA'
UNION ALL
SELECT 'Услуги WOW', id, '', 4 FROM regions WHERE code = 'UA'

-- WoW EU categories
UNION ALL
SELECT 'Подписка WOW', id, '', 1 FROM regions WHERE code = 'EU'
UNION ALL
SELECT 'Midnight', id, '', 2 FROM regions WHERE code = 'EU'
UNION ALL
SELECT 'The War Within', id, '', 3 FROM regions WHERE code = 'EU'
UNION ALL
SELECT 'Услуги WOW', id, '', 4 FROM regions WHERE code = 'EU'

-- WoW TUR categories
UNION ALL
SELECT 'Подписка WOW', id, '', 1 FROM regions WHERE code = 'TUR'
UNION ALL
SELECT 'Midnight', id, '', 2 FROM regions WHERE code = 'TUR'
UNION ALL
SELECT 'The War Within', id, '', 3 FROM regions WHERE code = 'TUR'
UNION ALL
SELECT 'Услуги WOW', id, '', 4 FROM regions WHERE code = 'TUR'

-- Системная категория (не отображается в обычном каталоге)
UNION ALL
SELECT 'Системные услуги', id, 'Специальные услуги системы', 99 FROM regions WHERE code = 'KZ';
