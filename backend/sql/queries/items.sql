-- name: GetItems :many
SELECT id, feed_id, guid, title, url, content, description, author, published_at, fetched_at, read, starred
FROM items
WHERE (sqlc.narg('feed_id')::int IS NULL OR feed_id = sqlc.narg('feed_id')::int)
  AND (sqlc.narg('read')::bool IS NULL OR read = sqlc.narg('read')::bool)
  AND (sqlc.narg('starred')::bool IS NULL OR starred = sqlc.narg('starred')::bool)
ORDER BY published_at DESC NULLS LAST
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountItems :one
SELECT COUNT(*)
FROM items
WHERE (sqlc.narg('feed_id')::int IS NULL OR feed_id = sqlc.narg('feed_id')::int)
  AND (sqlc.narg('read')::bool IS NULL OR read = sqlc.narg('read')::bool)
  AND (sqlc.narg('starred')::bool IS NULL OR starred = sqlc.narg('starred')::bool);

-- name: GetItemByID :one
SELECT id, feed_id, guid, title, url, content, description, author, published_at, fetched_at, read, starred
FROM items
WHERE id = $1;

-- name: UpsertItem :one
INSERT INTO items (feed_id, guid, title, url, content, description, author, published_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (feed_id, guid)
DO UPDATE SET
    title = EXCLUDED.title,
    url = EXCLUDED.url,
    content = EXCLUDED.content,
    description = EXCLUDED.description,
    author = EXCLUDED.author,
    published_at = EXCLUDED.published_at,
    fetched_at = NOW()
RETURNING id, feed_id, guid, title, url, content, description, author, published_at, fetched_at, read, starred;

-- name: MarkItemRead :exec
UPDATE items
SET read = $2
WHERE id = $1;

-- name: MarkItemStarred :exec
UPDATE items
SET starred = $2
WHERE id = $1;

-- name: MarkFeedItemsRead :exec
UPDATE items
SET read = TRUE
WHERE feed_id = $1 AND read = FALSE;

-- name: MarkAllItemsRead :exec
UPDATE items
SET read = TRUE
WHERE read = FALSE;

-- name: GetUnreadCountByFeed :many
SELECT feed_id, COUNT(*)::int AS count
FROM items
WHERE read = FALSE
GROUP BY feed_id;
