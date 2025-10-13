-- name: CreateOrUpdateSession :one
INSERT INTO appview_sessions (
    session_id,
    user_did,
    handle,
    created_at,
    expires_at
) VALUES (
    $1, $2, $3, NOW(), $4
)
ON CONFLICT (session_id) 
DO UPDATE SET 
    user_did = EXCLUDED.user_did,
    handle = EXCLUDED.handle,
    expires_at = EXCLUDED.expires_at,
    updated_at = NOW()
RETURNING *;
