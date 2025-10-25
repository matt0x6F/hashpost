-- name: GetSubforumByURI :one
SELECT * FROM appview_subforums 
WHERE atproto_uri = $1;
