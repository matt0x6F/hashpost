-- name: DeleteVote :exec
DELETE FROM votes WHERE id = $1;
