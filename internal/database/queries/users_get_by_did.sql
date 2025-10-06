-- name: GetUserByDID :one
SELECT * FROM users
WHERE did = $1;
