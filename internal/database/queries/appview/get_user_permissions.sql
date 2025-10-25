-- name: GetUserPermissions :many
SELECT DISTINCT p.name
FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_did = $1
    AND ur.is_active = TRUE
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
    AND ($2::text = '' OR $2::text IS NULL OR ur.subforum_id = $2::uuid OR ur.subforum_id IS NULL)
ORDER BY p.name;
