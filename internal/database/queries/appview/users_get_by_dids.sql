-- name: GetUsersByDIDs :many
SELECT * FROM appview_users WHERE did = ANY($1::text[]);
