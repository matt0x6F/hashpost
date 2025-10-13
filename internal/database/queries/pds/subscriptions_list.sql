-- name: ListSubscriptions :many
SELECT 
    ss.id,
    ss.user_id,
    ss.subforum_id,
    ss.created_at,
    u.handle as subscriber_handle,
    s.name as subforum_name,
    s.slug as subforum_slug
FROM subforum_subscriptions ss
JOIN users u ON ss.user_id = u.id
JOIN subforums s ON ss.subforum_id = s.id
WHERE ($1::text IS NULL OR ss.user_id::text = $1)
   OR ($2::text IS NULL OR ss.subforum_id::text = $2)
ORDER BY ss.created_at DESC
LIMIT $3 OFFSET $4;

