-- name: UpdateSubforumByURI :one
UPDATE subforums 
SET name = $2, slug = $3, description = $4, updated_at = NOW()
WHERE atproto_uri = $1
RETURNING id, name, slug, description, created_by, created_at, updated_at;
