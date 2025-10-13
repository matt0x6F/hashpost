-- name: UpdatePostByAtprotoURI :one
UPDATE appview_posts 
SET 
    title = $2,
    content = $3,
    updated_at = NOW()
WHERE atproto_uri = $1
RETURNING *;
