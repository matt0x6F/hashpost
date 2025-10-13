-- name: DeletePostByAtprotoURI :exec
DELETE FROM appview_posts 
WHERE atproto_uri = $1;
