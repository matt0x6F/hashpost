-- name: ListAppViewPosts :many
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
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAppViewPostsBySubforum :many
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
WHERE subforum_slug = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
