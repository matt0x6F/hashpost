-- name: GetSubforumBySlug :one
SELECT 
    id,
    name,
    slug,
    description,
    created_by,
    created_at,
    updated_at
FROM subforums
WHERE slug = $1;
