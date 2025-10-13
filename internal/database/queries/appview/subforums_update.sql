-- name: UpdateAppViewSubforum :one
UPDATE appview_subforums 
SET 
    name = $2,
    description = $3,
    updated_at = NOW()
WHERE slug = $1
RETURNING 
    id,
    name,
    slug,
    description,
    created_by_did,
    created_by_handle,
    created_at,
    updated_at,
    subscriber_count,
    post_count;
