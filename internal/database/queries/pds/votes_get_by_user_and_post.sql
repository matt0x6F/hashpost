-- name: GetVoteByUserAndPost :one
SELECT id, user_id, post_id, comment_id, vote_type, created_at
FROM votes
WHERE user_id = $1 AND post_id = $2;
