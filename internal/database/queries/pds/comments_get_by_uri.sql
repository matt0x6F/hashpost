-- name: GetCommentByURI :one
SELECT * FROM comments WHERE atproto_uri = $1;
