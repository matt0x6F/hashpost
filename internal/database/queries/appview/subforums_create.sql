-- name: CreateAppViewSubforum :one
INSERT INTO appview_subforums (
    name,
    slug,
    description,
    created_by_did,
    created_by_handle,
    prefix_type
) VALUES (
    $1, $2, $3, $4, $5, $6
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
    post_count,
    comment_count,
    prefix_type;
