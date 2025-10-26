-- name: GetPostsByAuthorDID :many
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
    s.name as subforum_name
FROM appview_posts p
LEFT JOIN appview_subforums s ON p.subforum_slug = s.slug
WHERE p.author_did = $1
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;
