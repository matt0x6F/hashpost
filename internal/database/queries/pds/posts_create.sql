-- name: CreatePost :one
INSERT INTO posts (user_id, subforum_id, title, content, atproto_uri)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

