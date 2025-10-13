-- name: ListSubforums :many
SELECT 
    id,
    name,
    slug,
    description,
    created_by,
    created_at,
    updated_at
FROM subforums
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
