-- name: CreateOrUpdateUserFromDID :one
INSERT INTO appview_users (
    did,
    handle,
    display_name,
    avatar_url,
    created_at,
    updated_at
) VALUES (
    $1, $2, $3, $4, NOW(), NOW()
)
ON CONFLICT (did) 
DO UPDATE SET 
    handle = EXCLUDED.handle,
    display_name = EXCLUDED.display_name,
    avatar_url = EXCLUDED.avatar_url,
    updated_at = NOW()
RETURNING *;
