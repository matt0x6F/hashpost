-- name: GetAllPermissions :many
SELECT 
    id,
    name,
    description,
    resource_type,
    created_at
FROM permissions
ORDER BY resource_type, name;
