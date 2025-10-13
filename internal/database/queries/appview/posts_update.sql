-- name: UpdateAppViewPost :one
UPDATE appview_posts 
SET 
    title = $2,
    content = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING 
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
