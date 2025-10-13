-- name: HasUserRole :one
SELECT EXISTS(
    SELECT 1
    FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.user_did = $1
        AND r.name = $2
        AND ur.is_active = TRUE
        AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
        AND (ur.subforum_id = $3 OR (ur.subforum_id IS NULL AND $3 IS NULL))
) as has_role;
