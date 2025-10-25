-- name: ListAppViewPosts :many
SELECT 
    p.id,
    p.atproto_uri,
    p.author_did,
    p.author_handle,
    p.subforum_slug,
    p.title,
    p.content,
    p.created_at,
    p.updated_at,
    p.upvotes,
    p.downvotes,
    p.comment_count,
    p.score,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url
FROM appview_posts p
LEFT JOIN appview_users u ON p.author_did = u.did
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAppViewPostsBySubforum :many
SELECT 
    p.id,
    p.atproto_uri,
    p.author_did,
    p.author_handle,
    p.subforum_slug,
    p.title,
    p.content,
    p.created_at,
    p.updated_at,
    p.upvotes,
    p.downvotes,
    p.comment_count,
    p.score,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url
FROM appview_posts p
LEFT JOIN appview_users u ON p.author_did = u.did
WHERE p.subforum_slug = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;
