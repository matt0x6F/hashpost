-- name: AssignSubforumRole :one
INSERT INTO user_roles (
    user_did,
    role_id,
    subforum_id,
    granted_by,
    expires_at,
    is_active
) VALUES (
    $1, -- user_did
    (SELECT id FROM roles WHERE roles.name = $2), -- role_name
    (SELECT id FROM appview_subforums WHERE slug = $3), -- subforum_slug
    $4, -- granted_by
    $5, -- expires_at
    TRUE
)
ON CONFLICT (user_did, role_id, subforum_id) 
DO UPDATE SET 
    granted_by = EXCLUDED.granted_by,
    expires_at = EXCLUDED.expires_at,
    is_active = TRUE,
    updated_at = NOW()
RETURNING id, user_did, role_id, subforum_id, granted_by, granted_at, expires_at, is_active;
