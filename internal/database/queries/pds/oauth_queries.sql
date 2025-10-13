-- name: CreateOAuthClient :one
INSERT INTO oauth_clients (client_id, client_name, redirect_uris, scopes, grant_types, response_types)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetOAuthClient :one
SELECT * FROM oauth_clients WHERE client_id = $1;

-- name: CreateAuthorizationCode :one
INSERT INTO oauth_authorization_codes (code, client_id, user_did, redirect_uri, scope, nonce, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetAuthorizationCode :one
SELECT * FROM oauth_authorization_codes WHERE code = $1 AND expires_at > NOW() AND used_at IS NULL;

-- name: MarkAuthorizationCodeUsed :exec
UPDATE oauth_authorization_codes SET used_at = NOW() WHERE code = $1;

-- name: CreateAccessToken :one
INSERT INTO oauth_access_tokens (access_token, refresh_token, client_id, user_did, scope, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAccessToken :one
SELECT * FROM oauth_access_tokens WHERE access_token = $1 AND expires_at > NOW();

-- name: GetRefreshToken :one
SELECT * FROM oauth_access_tokens WHERE refresh_token = $1 AND expires_at > NOW();

-- name: RevokeAccessToken :exec
UPDATE oauth_access_tokens SET expires_at = NOW() WHERE access_token = $1;

-- name: CreateDPoPNonce :one
INSERT INTO dpop_nonces (nonce, expires_at)
VALUES ($1, $2)
RETURNING *;

-- name: GetDPoPNonce :one
SELECT * FROM dpop_nonces WHERE nonce = $1 AND expires_at > NOW() AND used_at IS NULL;

-- name: MarkDPoPNonceUsed :exec
UPDATE dpop_nonces SET used_at = NOW() WHERE nonce = $1;

-- name: CreateUserSession :one
INSERT INTO user_sessions (session_id, user_did, handle, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserSession :one
SELECT * FROM user_sessions WHERE session_id = $1 AND expires_at > NOW();

-- name: UpdateUserSessionLastAccessed :exec
UPDATE user_sessions SET last_accessed_at = NOW() WHERE session_id = $1;

-- name: DeleteUserSession :exec
DELETE FROM user_sessions WHERE session_id = $1;

-- name: CleanupExpiredSessions :exec
DELETE FROM user_sessions WHERE expires_at < NOW();

-- name: CleanupExpiredAuthorizationCodes :exec
DELETE FROM oauth_authorization_codes WHERE expires_at < NOW();

-- name: CleanupExpiredAccessTokens :exec
DELETE FROM oauth_access_tokens WHERE expires_at < NOW();

-- name: CleanupExpiredDPoPNonces :exec
DELETE FROM dpop_nonces WHERE expires_at < NOW();
