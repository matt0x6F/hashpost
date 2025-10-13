-- name: RevokeSubforumRole :exec
UPDATE user_roles 
SET is_active = FALSE, updated_at = NOW()
WHERE user_did = $1 
    AND role_id = (SELECT id FROM roles WHERE roles.name = $2)
    AND subforum_id = (SELECT id FROM appview_subforums WHERE slug = $3);
