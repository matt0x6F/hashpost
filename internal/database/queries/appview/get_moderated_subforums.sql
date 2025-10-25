-- name: GetModeratedSubforums :many
SELECT DISTINCT
    sf.id,
    sf.name,
    sf.slug,
    sf.description,
    sf.created_by_did,
    sf.created_by_handle,
    sf.created_at,
    sf.updated_at,
    sf.subscriber_count,
    sf.post_count,
    sf.comment_count,
    sf.prefix_type
FROM appview_subforums sf
INNER JOIN user_roles ur ON sf.id = ur.subforum_id
INNER JOIN roles r ON ur.role_id = r.id
WHERE ur.user_did = $1
    AND ur.is_active = TRUE
    AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
    AND r.name IN ('subforum_owner', 'subforum_moderator')
ORDER BY sf.name ASC;
