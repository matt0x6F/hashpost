-- name: CreateAppViewComment :one
INSERT INTO appview_comments (
    atproto_uri,
    author_did,
    author_handle,
    post_id,
    parent_id,
    content
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING 
    id,
    atproto_uri,
    author_did,
    author_handle,
    post_id,
    parent_id,
    content,
    created_at,
    updated_at,
    upvotes,
    downvotes,
    score;
