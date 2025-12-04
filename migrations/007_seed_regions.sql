-- Seed regions
INSERT INTO regions (name, code) VALUES
    ('WoW KZ', 'KZ'),
    ('WoW UA', 'UA'),
    ('WoW EU', 'EU'),
    ('WoW TUR', 'TUR')
ON CONFLICT (code) DO NOTHING;
