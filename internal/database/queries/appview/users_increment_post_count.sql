-- name: IncrementUserPostCount :exec
UPDATE appview_users 
SET post_count = post_count + 1,
    updated_at = NOW()
WHERE did = $1;
