-- name: ListAppViewSubforums :many
SELECT 
    id,
    name,
    slug,
    description,
    created_by_did,
    created_by_handle,
    created_at,
    updated_at,
    subscriber_count,
    post_count,
    comment_count
FROM appview_subforums
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAppViewSubforumBySlug :one
SELECT 
    id,
    name,
    slug,
    description,
    created_by_did,
    created_by_handle,
    created_at,
    updated_at,
    subscriber_count,
    post_count,
    comment_count
FROM appview_subforums
WHERE slug = $1;
