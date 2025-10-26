-- name: IncrementUserCommentCount :exec
UPDATE appview_users 
SET comment_count = comment_count + 1,
    updated_at = NOW()
WHERE did = $1;
