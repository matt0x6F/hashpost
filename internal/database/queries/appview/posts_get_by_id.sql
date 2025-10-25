-- name: GetAppViewPostByID :one
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
WHERE p.id = $1;

-- name: GetAppViewPostByURI :one
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
WHERE p.atproto_uri = $1;
