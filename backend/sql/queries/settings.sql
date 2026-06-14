-- name: GetSettings :many
SELECT key, value
FROM settings;

-- name: GetSetting :one
SELECT key, value
FROM settings
WHERE key = $1;

-- name: UpsertSetting :exec
INSERT INTO settings (key, value)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;

-- name: DeleteSetting :exec
DELETE FROM settings
WHERE key = $1;
