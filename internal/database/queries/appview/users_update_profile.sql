-- name: UpdateUserProfile :one
UPDATE appview_users 
SET display_name = $2, bio = $3, avatar_url = $4, updated_at = NOW()
WHERE did = $1
RETURNING id, did, handle, display_name, bio, avatar_url, created_at, updated_at, post_count, comment_count, reputation, pds_source, last_seen_at;
