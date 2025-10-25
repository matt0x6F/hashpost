-- name: UpdateUserProfile :one
UPDATE appview_users 
SET display_name = $2, bio = $3, avatar_url = $4, updated_at = NOW()
WHERE did = $1
RETURNING *;
