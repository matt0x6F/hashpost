-- name: ListPosts :many
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
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListPostsBySubforum :many
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
WHERE s.slug = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPostsByUser :many
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
WHERE u.did = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;