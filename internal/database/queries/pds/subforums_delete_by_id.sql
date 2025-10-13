-- name: DeleteSubforumByID :exec
DELETE FROM subforums WHERE id = $1;
