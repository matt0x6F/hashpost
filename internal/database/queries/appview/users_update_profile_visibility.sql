-- name: UpdateUserProfileVisibility :one
UPDATE appview_users 
SET profile_visibility = $2, updated_at = NOW()
WHERE did = $1
RETURNING id, did, handle, display_name, bio, avatar_url, created_at, updated_at, post_count, comment_count, reputation, pds_source, last_seen_at, profile_visibility;
