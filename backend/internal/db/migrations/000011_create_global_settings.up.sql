CREATE TABLE IF NOT EXISTS global_settings (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

INSERT INTO global_settings (key, value) VALUES ('items_limit', '150')
ON CONFLICT DO NOTHING;
