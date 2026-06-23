-- name: CreateSession :one
INSERT INTO sessions (id, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING id, user_id, expires_at, created_at;

-- name: GetSessionWithUser :one
SELECT s.id AS session_id, s.expires_at, s.created_at,
       u.id AS user_id, u.username, u.password_hash, u.is_admin, u.created_at, u.updated_at
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.id = $1;

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE id = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE expires_at < NOW();