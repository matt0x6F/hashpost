-- name: GetCommentsByAuthorDID :many
SELECT 
    c.id,
    c.atproto_uri,
    c.author_did,
    c.author_handle,
    c.post_id,
    c.parent_id,
    c.content,
    c.created_at,
    c.updated_at,
    c.upvotes,
    c.downvotes,
    c.score,
    p.title as post_title,
    p.subforum_slug,
    s.name as subforum_name
FROM appview_comments c
LEFT JOIN appview_posts p ON c.post_id = p.id
LEFT JOIN appview_subforums s ON p.subforum_slug = s.slug
WHERE c.author_did = $1
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;
