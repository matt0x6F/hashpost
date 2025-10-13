-- name: GetAllRoles :many
SELECT 
    id,
    name,
    description,
    is_platform_role,
    created_at
FROM roles
ORDER BY is_platform_role DESC, name ASC;
