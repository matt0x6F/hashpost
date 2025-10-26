-- name: CreateOrUpdateUserFromDID :one
INSERT INTO appview_users (
    did,
    handle,
    display_name,
    avatar_url,
    pds_source,
    last_seen_at,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, NOW(), NOW()
)
ON CONFLICT (did) 
DO UPDATE SET 
    handle = EXCLUDED.handle,
    display_name = EXCLUDED.display_name,
    avatar_url = EXCLUDED.avatar_url,
    pds_source = EXCLUDED.pds_source,
    last_seen_at = EXCLUDED.last_seen_at,
    updated_at = NOW()
RETURNING id, did, handle, display_name, bio, avatar_url, created_at, updated_at, post_count, comment_count, reputation, pds_source, last_seen_at;
