-- name: DeleteAppViewPost :exec
DELETE FROM appview_posts WHERE id = $1;
