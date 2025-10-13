-- name: GetUserRoles :many
SELECT 
    r.name as role_name,
    r.description as role_description,
    r.is_platform_role,
    ur.subforum_id,
    sf.slug as subforum_slug,
    sf.name as subforum_name,
    ur.granted_by,
    ur.granted_at,
    ur.expires_at,
    ur.is_active
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
LEFT JOIN appview_subforums sf ON ur.subforum_id = sf.id
WHERE ur.user_did = $1
    AND ur.is_active = TRUE
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
ORDER BY r.is_platform_role DESC, sf.name ASC;
