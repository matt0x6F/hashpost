-- name: GetCommentByID :one
SELECT 
    c.id,
    c.user_id,
    c.post_id,
    c.parent_id,
    c.content,
    c.atproto_uri,
    c.created_at,
    c.updated_at,
    u.handle as author_handle
FROM comments c
JOIN users u ON c.user_id = u.id
WHERE c.id = $1;

