-- name: UpdatePostByAtprotoURI :one
UPDATE posts 
SET title = $2, content = $3, updated_at = NOW()
WHERE atproto_uri = $1
RETURNING id, user_id, subforum_id, title, content, atproto_uri, created_at, updated_at;
