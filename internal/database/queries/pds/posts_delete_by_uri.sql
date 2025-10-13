-- name: DeletePostByAtprotoURI :exec
DELETE FROM posts WHERE atproto_uri = $1;
