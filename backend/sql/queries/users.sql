-- name: CreateUser :one
INSERT INTO users (username, password_hash, is_admin)
VALUES ($1, $2, $3)
RETURNING id, username, password_hash, is_admin, created_at, updated_at;

-- name: GetUserByUsername :one
SELECT id, username, password_hash, is_admin, created_at, updated_at
FROM users
WHERE username = $1;

-- name: GetUserByID :one
SELECT id, username, password_hash, is_admin, created_at, updated_at
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT u.id, u.username, u.is_admin, u.created_at, u.updated_at,
       MAX(s.created_at)::timestamptz AS last_login_at
FROM users u
LEFT JOIN sessions s ON s.user_id = u.id
GROUP BY u.id
ORDER BY u.id;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;