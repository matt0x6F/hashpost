-- name: GetUsersWithRoles :many
SELECT 
    u.did as user_did,
    u.handle,
    u.display_name,
    u.created_at,
    u.pds_source,
    u.last_seen_at,
    COALESCE(COUNT(ur.id), 0)::bigint as role_count
FROM appview_users u
LEFT JOIN user_roles ur ON u.did = ur.user_did 
    AND ur.is_active = TRUE 
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
    AND (
        $1::text IS NULL 
        OR $1::text = '' 
        OR ur.subforum_id IS NULL
        OR (LENGTH($1::text) > 0 AND ur.subforum_id::text = $1::text)
    )
GROUP BY u.did, u.handle, u.display_name, u.created_at, u.pds_source, u.last_seen_at
ORDER BY u.created_at DESC
LIMIT $2 OFFSET $3;
