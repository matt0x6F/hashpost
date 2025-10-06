-- name: CreateUser :one
INSERT INTO users (did, handle, email)
VALUES ($1, $2, $3)
RETURNING *;
