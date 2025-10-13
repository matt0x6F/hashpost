-- name: GetAppViewPostByID :one
SELECT 
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
    score
FROM appview_posts
WHERE id = $1;

-- name: GetAppViewPostByURI :one
SELECT 
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
    score
FROM appview_posts
WHERE atproto_uri = $1;
