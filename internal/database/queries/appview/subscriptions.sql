-- Subscription operations for HashPost AppView
-- These queries handle user subscriptions to subforums

-- name: CreateSubscription :one
INSERT INTO appview_subscriptions (user_did, user_handle, subforum_slug)
VALUES ($1, $2, $3)
ON CONFLICT (user_did, subforum_slug) DO NOTHING
RETURNING *;

-- name: DeleteSubscription :exec
DELETE FROM appview_subscriptions 
WHERE user_did = $1 AND subforum_slug = $2;

-- name: GetUserSubscription :one
SELECT * FROM appview_subscriptions 
WHERE user_did = $1 AND subforum_slug = $2;

-- name: ListUserSubscriptions :many
SELECT 
    s.*,
    sf.name as subforum_name,
    sf.description as subforum_description
FROM appview_subscriptions s
JOIN appview_subforums sf ON s.subforum_slug = sf.slug
WHERE s.user_did = $1
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountUserSubscriptions :one
SELECT COUNT(*) as total
FROM appview_subscriptions 
WHERE user_did = $1;

-- name: UpdateSubforumSubscriberCount :exec
UPDATE appview_subforums 
SET subscriber_count = (
    SELECT COUNT(*) 
    FROM appview_subscriptions 
    WHERE subforum_slug = $1
)
WHERE slug = $1;

-- name: GetSubforumSubscriberCount :one
SELECT subscriber_count
FROM appview_subforums 
WHERE slug = $1;

-- name: ListSubforumSubscribers :many
SELECT 
    s.user_did,
    s.user_handle,
    s.created_at as subscribed_at
FROM appview_subscriptions s
WHERE s.subforum_slug = $1
ORDER BY s.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountSubforumSubscribers :one
SELECT COUNT(*) as total
FROM appview_subscriptions 
WHERE subforum_slug = $1;
