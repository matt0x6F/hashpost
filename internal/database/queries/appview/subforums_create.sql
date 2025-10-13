-- name: CreateAppViewSubforum :one
INSERT INTO appview_subforums (
    name,
    slug,
    description,
    created_by_did,
    created_by_handle
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING 
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
