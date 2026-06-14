-- name: GetFeeds :many
SELECT id, url, title, description, site_link, last_fetched_at, created_at, updated_at, icon_url
FROM feeds
ORDER BY title;

-- name: GetFeedByID :one
SELECT id, url, title, description, site_link, last_fetched_at, created_at, updated_at, icon_url
FROM feeds
WHERE id = $1;

-- name: GetFeedByURL :one
SELECT id, url, title, description, site_link, last_fetched_at, created_at, updated_at, icon_url
FROM feeds
WHERE url = $1;

-- name: CreateFeed :one
INSERT INTO feeds (url, title, description, site_link, icon_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, url, title, description, site_link, last_fetched_at, created_at, updated_at, icon_url;

-- name: DeleteFeed :exec
DELETE FROM feeds
WHERE id = $1;

-- name: UpdateFeedLastFetched :exec
UPDATE feeds
SET last_fetched_at = NOW(), updated_at = NOW()
WHERE id = $1;

-- name: UpdateFeedMetadata :exec
UPDATE feeds
SET title = $2, description = $3, site_link = $4, icon_url = $5, updated_at = NOW()
WHERE id = $1;
