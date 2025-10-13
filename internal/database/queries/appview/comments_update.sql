-- name: UpdateAppViewComment :one
UPDATE appview_comments 
SET 
    content = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING 
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
