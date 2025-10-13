-- name: CreateUserWithPassword :one
INSERT INTO users (handle, did, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;
