-- name: GetSubforumByID :one
SELECT * FROM appview_subforums 
WHERE id = $1;
