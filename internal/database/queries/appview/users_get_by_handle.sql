-- name: GetUserByHandle :one
SELECT id, did, handle, display_name, bio, avatar_url, created_at, updated_at, post_count, comment_count, reputation, pds_source, last_seen_at, profile_visibility 
FROM appview_users 
WHERE handle = $1;
