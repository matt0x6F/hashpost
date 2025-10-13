-- name: DeleteAppViewSubforum :exec
DELETE FROM appview_subforums WHERE slug = $1;
