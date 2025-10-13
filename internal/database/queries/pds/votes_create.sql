-- name: CreateVote :one
INSERT INTO votes (user_id, post_id, comment_id, vote_type)
VALUES ($1, $2, $3, $4)
RETURNING *;

