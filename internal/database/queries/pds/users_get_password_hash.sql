-- name: GetUserPasswordHash :one
SELECT password_hash 
FROM users 
WHERE id = $1;
