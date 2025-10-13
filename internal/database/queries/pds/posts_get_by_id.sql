-- name: GetPostByID :one
SELECT 
    p.id,
    p.user_id,
    p.subforum_id,
    p.title,
    p.content,
    p.atproto_uri,
    p.created_at,
    p.updated_at,
    u.handle as author_handle,
    s.slug as subforum_slug,
    s.name as subforum_name
FROM posts p
JOIN users u ON p.user_id = u.id
JOIN subforums s ON p.subforum_id = s.id
WHERE p.id = $1;

