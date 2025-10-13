-- name: GetUserByDID :one
SELECT * FROM appview_users 
WHERE did = $1;
