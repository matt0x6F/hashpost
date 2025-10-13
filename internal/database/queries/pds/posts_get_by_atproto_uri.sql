-- name: GetPostByAtprotoURI :one
SELECT id, user_id, subforum_id, title, content, atproto_uri, created_at, updated_at
FROM posts
WHERE atproto_uri = $1;
