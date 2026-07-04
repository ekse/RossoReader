-- name: GetGlobalSetting :one
SELECT value FROM global_settings WHERE key = $1;

-- name: UpsertGlobalSetting :exec
INSERT INTO global_settings (key, value)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value;

-- name: CountItemsByFeed :one
SELECT COUNT(*)::int FROM items WHERE feed_id = $1;

-- name: DeleteExcessItems :many
WITH newest AS (
    SELECT id FROM items WHERE feed_id = sqlc.arg('feed_id')::int
    ORDER BY COALESCE(published_at, fetched_at) DESC
    LIMIT sqlc.arg('max_items')::int
)
DELETE FROM items
WHERE feed_id = sqlc.arg('feed_id')::int
  AND id NOT IN (SELECT id FROM newest)
  AND NOT EXISTS (
      SELECT 1 FROM user_item_states uis
      WHERE uis.item_id = items.id AND uis.starred = true
  )
RETURNING id;
