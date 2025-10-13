-- name: CreateSubscription :one
INSERT INTO subforum_subscriptions (user_id, subforum_id)
VALUES ($1, $2)
RETURNING *;

