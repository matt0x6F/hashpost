-- name: GetSubforumMembers :many
SELECT DISTINCT
    ur.user_did,
    COUNT(ur.id) as role_count
FROM user_roles ur
WHERE ur.is_active = TRUE
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
    AND ur.subforum_id = (SELECT id FROM appview_subforums WHERE slug = $1)
GROUP BY ur.user_did 
ORDER BY ur.user_did 
LIMIT $2 OFFSET $3;
