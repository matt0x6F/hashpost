-- Comment operations for HashPost AppView
-- These queries handle CRUD operations for comments with proper denormalized counts

-- name: CreateComment :one
INSERT INTO appview_comments (
    atproto_uri,
    author_did,
    author_handle,
    post_id,
    parent_id,
    content
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM appview_comments 
WHERE id = $1;

-- name: ListCommentsByPost :many
SELECT 
    c.*,
    p.title as post_title,
    p.subforum_slug,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url
FROM appview_comments c
JOIN appview_posts p ON c.post_id = p.id
LEFT JOIN appview_users u ON c.author_did = u.did
WHERE c.post_id = $1
ORDER BY c.created_at ASC
LIMIT $2 OFFSET $3;

-- name: CountCommentsByPost :one
SELECT COUNT(*) as total
FROM appview_comments 
WHERE post_id = $1;

-- name: ListCommentsByPostWithReplies :many
WITH RECURSIVE comment_tree AS (
    -- Base case: top-level comments
    SELECT c.*, 0 as depth, p.title as post_title, p.subforum_slug, u.display_name as author_display_name, u.avatar_url as author_avatar_url
    FROM appview_comments c
    JOIN appview_posts p ON c.post_id = p.id
    LEFT JOIN appview_users u ON c.author_did = u.did
    WHERE c.post_id = $1 AND c.parent_id IS NULL
    
    UNION ALL
    
    -- Recursive case: replies
    SELECT c.*, ct.depth + 1, ct.post_title, ct.subforum_slug, u.display_name as author_display_name, u.avatar_url as author_avatar_url
    FROM appview_comments c
    JOIN comment_tree ct ON c.parent_id = ct.id
    LEFT JOIN appview_users u ON c.author_did = u.did
)
SELECT * FROM comment_tree
ORDER BY depth, created_at ASC
LIMIT $2 OFFSET $3;

-- name: UpdateComment :one
WITH updated AS (
    UPDATE appview_comments 
    SET 
        content = $2,
        updated_at = NOW()
    WHERE appview_comments.id = $1
    RETURNING *
)
SELECT 
    updated.*,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url
FROM updated
LEFT JOIN appview_users u ON updated.author_did = u.did;

-- name: DeleteComment :exec
DELETE FROM appview_comments 
WHERE id = $1;

-- name: UpdatePostCommentCount :exec
UPDATE appview_posts 
SET comment_count = (
    SELECT COUNT(*) 
    FROM appview_comments 
    WHERE post_id = $1
)
WHERE id = $1;

-- name: GetPostCommentCount :one
SELECT comment_count
FROM appview_posts 
WHERE id = $1;

-- name: ListCommentsByAuthor :many
SELECT 
    c.*,
    p.title as post_title,
    p.subforum_slug,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url
FROM appview_comments c
JOIN appview_posts p ON c.post_id = p.id
LEFT JOIN appview_users u ON c.author_did = u.did
WHERE c.author_did = $1
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountCommentsByAuthor :one
SELECT COUNT(*) as total
FROM appview_comments 
WHERE author_did = $1;

-- name: GetCommentWithPost :one
SELECT 
    c.*,
    p.title as post_title,
    p.subforum_slug,
    p.author_did as post_author_did,
    p.author_handle as post_author_handle,
    u.display_name as author_display_name,
    u.avatar_url as author_avatar_url
FROM appview_comments c
JOIN appview_posts p ON c.post_id = p.id
LEFT JOIN appview_users u ON c.author_did = u.did
WHERE c.id = $1;

-- name: GetCommentByURI :one
SELECT * FROM appview_comments WHERE atproto_uri = $1;

-- name: GetPostByAtprotoURI :one
SELECT * FROM appview_posts WHERE atproto_uri = $1;

-- name: UpdateCommentByURI :one
UPDATE appview_comments 
SET 
    content = $2,
    updated_at = NOW()
WHERE atproto_uri = $1
RETURNING *;

-- name: DeleteCommentByURI :exec
DELETE FROM appview_comments WHERE atproto_uri = $1;
