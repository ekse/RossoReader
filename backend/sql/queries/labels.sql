-- name: GetLabels :many
SELECT * FROM labels
WHERE user_id = $1
ORDER BY name;

-- name: GetLabelByID :one
SELECT * FROM labels
WHERE id = $1 AND user_id = $2;

-- name: CreateLabel :one
INSERT INTO labels (user_id, name)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateLabel :exec
UPDATE labels
SET name = $2
WHERE id = $1 AND user_id = $3;

-- name: DeleteLabel :exec
DELETE FROM labels
WHERE id = $1 AND user_id = $2;

-- name: AddFeedLabel :exec
INSERT INTO feed_labels (feed_id, label_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveFeedLabel :exec
DELETE FROM feed_labels
WHERE feed_id = $1 AND label_id = $2;

-- name: GetFeedLabels :many
SELECT l.* FROM labels l
INNER JOIN feed_labels fl ON fl.label_id = l.id
WHERE fl.feed_id = $1
  AND l.user_id = $2
ORDER BY l.name;

-- name: GetFeedsByLabel :many
SELECT f.* FROM feeds f
INNER JOIN feed_labels fl ON fl.feed_id = f.id
WHERE fl.label_id = $1
  AND f.user_id = $2
ORDER BY f.title;

-- name: GetUserLabelAssignments :many
SELECT label_id, feed_id FROM feed_labels;
