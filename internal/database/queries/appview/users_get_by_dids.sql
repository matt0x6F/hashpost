-- name: GetUsersByDIDs :many
SELECT id, did, handle, display_name, bio, avatar_url, created_at, updated_at, post_count, comment_count, reputation, pds_source, last_seen_at FROM appview_users WHERE did = ANY($1::text[]);
