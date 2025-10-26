-- name: DeleteCommentByAtprotoURI :exec
DELETE FROM comments WHERE atproto_uri = $1;
