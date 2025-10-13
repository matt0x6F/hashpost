-- name: ListPostsWithCursor :many
SELECT * FROM posts 
WHERE created_at < $1
ORDER BY created_at DESC 
LIMIT $2;

-- name: ListPostsWithCursorByUser :many
SELECT p.* FROM posts p
JOIN users u ON p.user_id = u.id
WHERE u.did = $1 AND p.created_at < $2
ORDER BY p.created_at DESC 
LIMIT $3;
