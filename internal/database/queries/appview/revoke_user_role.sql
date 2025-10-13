-- name: RevokeUserRole :exec
UPDATE user_roles 
SET is_active = FALSE, updated_at = NOW()
WHERE user_did = $1
    AND role_id = (SELECT id FROM roles WHERE name = $2)
    AND (subforum_id = $3 OR (subforum_id IS NULL AND $3 IS NULL));
