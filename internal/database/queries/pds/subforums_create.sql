-- name: CreateSubforum :one
INSERT INTO subforums (name, slug, description, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;
