-- name: UpdateSubforumPostCount :exec
UPDATE appview_subforums 
SET post_count = post_count + $2, updated_at = NOW()
WHERE slug = $1;

-- name: UpdateSubforumCommentCount :exec
UPDATE appview_subforums 
SET comment_count = comment_count + $2, updated_at = NOW()
WHERE slug = $1;

-- name: GetSubforumStats :one
SELECT 
    subscriber_count,
    post_count,
    comment_count
FROM appview_subforums 
WHERE slug = $1;
