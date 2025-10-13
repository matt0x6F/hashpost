-- name: CreateComment :one
INSERT INTO comments (user_id, post_id, parent_id, content, atproto_uri)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

