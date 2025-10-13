-- name: CreateAppViewPost :one
INSERT INTO appview_posts (
    atproto_uri,
    author_did,
    author_handle,
    subforum_slug,
    title,
    content
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING 
    id,
    atproto_uri,
    author_did,
    author_handle,
    subforum_slug,
    title,
    content,
    created_at,
    updated_at,
    upvotes,
    downvotes,
    comment_count,
    score;
