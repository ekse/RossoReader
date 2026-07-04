INSERT INTO global_settings (key, value) VALUES ('feeds_limit', '200')
ON CONFLICT DO NOTHING;
