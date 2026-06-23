-- name: GetItems :many
SELECT i.id, i.feed_id, i.guid, i.title, i.url, i.content, i.description, i.author, i.published_at, i.fetched_at,
       COALESCE(uis.read, false) AS is_read, COALESCE(uis.starred, false) AS is_starred
FROM items i
JOIN feeds f ON f.id = i.feed_id AND f.user_id = sqlc.arg('user_id')::bigint
LEFT JOIN user_item_states uis ON uis.item_id = i.id AND uis.user_id = sqlc.arg('user_id')::bigint
WHERE (sqlc.narg('feed_id')::int IS NULL OR i.feed_id = sqlc.narg('feed_id')::int)
  AND (sqlc.narg('read')::bool IS NULL OR COALESCE(uis.read, false) = sqlc.narg('read')::bool)
  AND (sqlc.narg('starred')::bool IS NULL OR COALESCE(uis.starred, false) = sqlc.narg('starred')::bool)
ORDER BY i.published_at DESC NULLS LAST
LIMIT sqlc.arg('limit') OFFSET sqlc.arg('offset');

-- name: CountItems :one
SELECT COUNT(*)
FROM items i
JOIN feeds f ON f.id = i.feed_id AND f.user_id = sqlc.arg('user_id')::bigint
LEFT JOIN user_item_states uis ON uis.item_id = i.id AND uis.user_id = sqlc.arg('user_id')::bigint
WHERE (sqlc.narg('feed_id')::int IS NULL OR i.feed_id = sqlc.narg('feed_id')::int)
  AND (sqlc.narg('read')::bool IS NULL OR COALESCE(uis.read, false) = sqlc.narg('read')::bool)
  AND (sqlc.narg('starred')::bool IS NULL OR COALESCE(uis.starred, false) = sqlc.narg('starred')::bool);

-- name: GetItemByID :one
SELECT i.id, i.feed_id, i.guid, i.title, i.url, i.content, i.description, i.author, i.published_at, i.fetched_at,
       COALESCE(uis.read, false) AS is_read, COALESCE(uis.starred, false) AS is_starred
FROM items i
JOIN feeds f ON f.id = i.feed_id AND f.user_id = sqlc.arg('user_id')::bigint
LEFT JOIN user_item_states uis ON uis.item_id = i.id AND uis.user_id = sqlc.arg('user_id')::bigint
WHERE i.id = $1;

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
RETURNING id, feed_id, guid, title, url, content, description, author, published_at, fetched_at;

-- name: SetItemRead :exec
INSERT INTO user_item_states (user_id, item_id, read, starred)
SELECT $1, i.id, $3, false
FROM items i JOIN feeds f ON f.id = i.feed_id AND f.user_id = $1
WHERE i.id = $2
ON CONFLICT (user_id, item_id) DO UPDATE SET read = EXCLUDED.read;

-- name: SetItemStarred :exec
INSERT INTO user_item_states (user_id, item_id, read, starred)
SELECT $1, i.id, false, $3
FROM items i JOIN feeds f ON f.id = i.feed_id AND f.user_id = $1
WHERE i.id = $2
ON CONFLICT (user_id, item_id) DO UPDATE SET starred = EXCLUDED.starred;

-- name: MarkFeedItemsReadForUser :exec
INSERT INTO user_item_states (user_id, item_id, read, starred)
SELECT $1, i.id, true, false FROM items i WHERE i.feed_id = $2
ON CONFLICT (user_id, item_id) DO UPDATE SET read = true;

-- name: MarkAllItemsReadForUser :exec
INSERT INTO user_item_states (user_id, item_id, read, starred)
SELECT $1, i.id, true, false FROM items i
  JOIN feeds f ON f.id = i.feed_id AND f.user_id = $1
ON CONFLICT (user_id, item_id) DO UPDATE SET read = true;

-- name: GetUnreadCountByFeedForUser :many
SELECT i.feed_id, COUNT(*)::int AS count
FROM items i
JOIN feeds f ON f.id = i.feed_id AND f.user_id = $1
LEFT JOIN user_item_states uis ON uis.user_id = $1 AND uis.item_id = i.id
WHERE uis.item_id IS NULL OR uis.read = false
GROUP BY i.feed_id;