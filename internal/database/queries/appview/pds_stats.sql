-- name: GetPDSServerStats :many
SELECT 
    COALESCE(pds_source, 'Unknown PDS') as pds_endpoint,
    COUNT(*) as user_count,
    MAX(last_seen_at) as last_activity,
    COUNT(CASE WHEN last_seen_at > NOW() - INTERVAL '24 hours' THEN 1 END) as active_users_24h
FROM appview_users
GROUP BY pds_source
ORDER BY user_count DESC;

-- name: GetPDSServerDetails :one
SELECT 
    COALESCE(pds_source, 'Unknown PDS') as pds_endpoint,
    COUNT(*) as user_count,
    MAX(last_seen_at) as last_activity,
    MIN(created_at) as first_user_created,
    COUNT(CASE WHEN last_seen_at > NOW() - INTERVAL '24 hours' THEN 1 END) as active_users_24h,
    COUNT(CASE WHEN last_seen_at > NOW() - INTERVAL '7 days' THEN 1 END) as active_users_7d,
    COUNT(CASE WHEN last_seen_at > NOW() - INTERVAL '30 days' THEN 1 END) as active_users_30d
FROM appview_users
WHERE pds_source = $1
GROUP BY pds_source;

-- name: GetUsersByPDSSource :many
SELECT 
    u.did,
    u.handle,
    u.display_name,
    u.created_at,
    u.last_seen_at,
    COALESCE(COUNT(ur.id), 0)::bigint as role_count
FROM appview_users u
LEFT JOIN user_roles ur ON u.did = ur.user_did 
    AND ur.is_active = TRUE 
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
WHERE u.pds_source = $1
GROUP BY u.did, u.handle, u.display_name, u.created_at, u.last_seen_at
ORDER BY u.last_seen_at DESC NULLS LAST
LIMIT $2 OFFSET $3;
