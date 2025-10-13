-- Vote operations for HashPost AppView
-- These queries handle voting on posts and comments with proper upsert logic

-- name: CreateVote :one
INSERT INTO appview_votes (user_did, post_id, comment_id, vote_type)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_did, post_id) DO UPDATE SET
    vote_type = EXCLUDED.vote_type,
    created_at = NOW()
WHERE appview_votes.post_id IS NOT NULL
RETURNING *;

-- name: CreateVoteOnComment :one
INSERT INTO appview_votes (user_did, post_id, comment_id, vote_type)
VALUES ($1, $2, $3, $4)
ON CONFLICT (user_did, comment_id) DO UPDATE SET
    vote_type = EXCLUDED.vote_type,
    created_at = NOW()
WHERE appview_votes.comment_id IS NOT NULL
RETURNING *;

-- name: DeleteVote :exec
DELETE FROM appview_votes 
WHERE user_did = $1 AND post_id = $2;

-- name: DeleteVoteOnComment :exec
DELETE FROM appview_votes 
WHERE user_did = $1 AND comment_id = $2;

-- name: GetUserVoteOnPost :one
SELECT * FROM appview_votes 
WHERE user_did = $1 AND post_id = $2;

-- name: GetUserVoteOnComment :one
SELECT * FROM appview_votes 
WHERE user_did = $1 AND comment_id = $2;

-- name: UpdatePostVoteCounts :exec
UPDATE appview_posts 
SET 
    upvotes = (
        SELECT COUNT(*) 
        FROM appview_votes v
        WHERE v.post_id = appview_posts.id AND v.vote_type = 'up'
    ),
    downvotes = (
        SELECT COUNT(*) 
        FROM appview_votes v
        WHERE v.post_id = appview_posts.id AND v.vote_type = 'down'
    ),
    score = (
        SELECT COUNT(*) 
        FROM appview_votes v
        WHERE v.post_id = appview_posts.id AND v.vote_type = 'up'
    ) - (
        SELECT COUNT(*) 
        FROM appview_votes v
        WHERE v.post_id = appview_posts.id AND v.vote_type = 'down'
    )
WHERE appview_posts.id = $1;

-- name: UpdateCommentVoteCounts :exec
UPDATE appview_comments 
SET 
    upvotes = (
        SELECT COUNT(*) 
        FROM appview_votes v
        WHERE v.comment_id = appview_comments.id AND v.vote_type = 'up'
    ),
    downvotes = (
        SELECT COUNT(*) 
        FROM appview_votes v
        WHERE v.comment_id = appview_comments.id AND v.vote_type = 'down'
    ),
    score = (
        SELECT COUNT(*) 
        FROM appview_votes v
        WHERE v.comment_id = appview_comments.id AND v.vote_type = 'up'
    ) - (
        SELECT COUNT(*) 
        FROM appview_votes v
        WHERE v.comment_id = appview_comments.id AND v.vote_type = 'down'
    )
WHERE appview_comments.id = $1;

-- name: GetPostVoteCounts :one
SELECT 
    upvotes,
    downvotes,
    score
FROM appview_posts 
WHERE id = $1;

-- name: GetCommentVoteCounts :one
SELECT 
    upvotes,
    downvotes,
    score
FROM appview_comments 
WHERE id = $1;
