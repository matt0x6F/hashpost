-- name: CheckUserPermission :one
SELECT EXISTS(
    SELECT 1
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.user_did = $1
        AND p.name = $2
        AND ur.is_active = TRUE
        AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
        AND (
            -- Platform permissions (no subforum restriction)
            (p.resource_type = 'platform' AND ur.subforum_id IS NULL)
            OR
            -- Subforum permissions (specific subforum or any subforum)
            (p.resource_type IN ('subforum', 'post', 'comment', 'vote') AND (
                ur.subforum_id = $3 OR ur.subforum_id IS NULL
            ))
        )
) as has_permission;
