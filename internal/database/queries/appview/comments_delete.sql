-- name: DeleteAppViewComment :exec
DELETE FROM appview_comments WHERE id = $1;
