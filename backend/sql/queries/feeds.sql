-- name: GetFeeds :many
SELECT * FROM feeds
WHERE user_id = $1
ORDER BY title;

-- name: GetAllFeeds :many
SELECT * FROM feeds
ORDER BY title;

-- name: GetFeedByID :one
SELECT * FROM feeds
WHERE id = $1 AND user_id = $2;

-- name: GetFeedByIDAny :one
SELECT * FROM feeds
WHERE id = $1;

-- name: CreateFeed :one
INSERT INTO feeds (user_id, url, title, description, site_link, icon_url, etag)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: DeleteFeed :exec
DELETE FROM feeds
WHERE id = $1 AND user_id = $2;

-- name: UpdateFeedLastFetched :exec
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: UpdateFeedEtag :exec
UPDATE feeds
SET etag = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateFeedMetadata :exec
UPDATE feeds
SET title = $2, description = $3, site_link = $4, icon_url = $5, updated_at = NOW()
WHERE id = $1;

-- name: SetFeedFetchError :exec
UPDATE feeds
SET last_fetch_error = $2, updated_at = NOW()
WHERE id = $1;

-- name: ClearFeedFetchError :exec
UPDATE feeds
SET last_fetch_error = NULL, updated_at = NOW()
WHERE id = $1;