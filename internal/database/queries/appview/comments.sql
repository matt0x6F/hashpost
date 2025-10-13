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
    p.subforum_slug
FROM appview_comments c
JOIN appview_posts p ON c.post_id = p.id
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
    SELECT c.*, 0 as depth, p.title as post_title, p.subforum_slug
    FROM appview_comments c
    JOIN appview_posts p ON c.post_id = p.id
    WHERE c.post_id = $1 AND c.parent_id IS NULL
    
    UNION ALL
    
    -- Recursive case: replies
    SELECT c.*, ct.depth + 1, ct.post_title, ct.subforum_slug
    FROM appview_comments c
    JOIN comment_tree ct ON c.parent_id = ct.id
)
SELECT * FROM comment_tree
ORDER BY depth, created_at ASC
LIMIT $2 OFFSET $3;

-- name: UpdateComment :one
UPDATE appview_comments 
SET 
    content = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

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
    p.subforum_slug
FROM appview_comments c
JOIN appview_posts p ON c.post_id = p.id
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
    p.author_handle as post_author_handle
FROM appview_comments c
JOIN appview_posts p ON c.post_id = p.id
WHERE c.id = $1;
