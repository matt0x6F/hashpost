-- name: UpdateUserPasswordHash :exec
UPDATE users 
SET password_hash = $1, updated_at = NOW()
WHERE id = $2;
