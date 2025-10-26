-- name: GetCommentByAtprotoURI :one
SELECT id, user_id, post_id, parent_id, content, atproto_uri, created_at, updated_at
FROM comments
WHERE atproto_uri = $1;
