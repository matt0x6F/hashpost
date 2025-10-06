-- name: GetUserByHandle :one
SELECT * FROM users
WHERE handle = $1;
