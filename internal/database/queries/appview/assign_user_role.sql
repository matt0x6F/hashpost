-- name: AssignUserRole :one
INSERT INTO user_roles (
    user_did,
    role_id,
    subforum_id,
    granted_by,
    expires_at
) VALUES (
    $1, -- user_did
    (SELECT id FROM roles WHERE name = $2), -- role_name
    $3, -- subforum_id (NULL for platform roles)
    $4, -- granted_by (DID of user granting the role)
    $5  -- expires_at (NULL for permanent)
)
RETURNING id, user_did, role_id, subforum_id, granted_by, granted_at, expires_at, is_active;
