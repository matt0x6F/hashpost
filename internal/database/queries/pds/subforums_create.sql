-- name: CreateSubforum :one
INSERT INTO subforums (name, slug, description, created_by, prefix_type)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
