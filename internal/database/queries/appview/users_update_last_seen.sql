-- name: UpdateUserLastSeen :exec
UPDATE appview_users 
SET last_seen_at = NOW()
WHERE did = $1;
