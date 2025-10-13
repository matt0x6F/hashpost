-- name: GetSubforumBySlug :one
SELECT * FROM appview_subforums 
WHERE slug = $1;
