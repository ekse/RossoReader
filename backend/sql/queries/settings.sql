-- name: GetSettings :many
SELECT key, value
FROM settings
WHERE user_id = $1;

-- name: GetSetting :one
SELECT key, value
FROM settings
WHERE user_id = $1 AND key = $2;

-- name: UpsertSetting :exec
INSERT INTO settings (user_id, key, value)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, key) DO UPDATE SET value = EXCLUDED.value;

-- name: DeleteSetting :exec
DELETE FROM settings
WHERE user_id = $1 AND key = $2;