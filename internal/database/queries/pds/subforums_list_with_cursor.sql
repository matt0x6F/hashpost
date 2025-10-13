-- name: ListSubforumsWithCursor :many
SELECT * FROM subforums 
WHERE created_at < $1
ORDER BY created_at DESC 
LIMIT $2;
