# PDS Database Queries

## Overview

The PDS uses SQLC for type-safe database operations. All queries are defined in SQL files and generated into Go code, ensuring type safety and preventing SQL injection vulnerabilities.

## Query Organization

### File Structure
- **Location**: `internal/database/queries/pds/`
- **Generated Code**: `internal/database/generated/pds/`
- **Configuration**: `sqlc-pds.yaml`

### Query Categories
- **User Operations**: User CRUD, authentication, session management
- **Content Operations**: Posts, comments, subforums
- **OAuth Operations**: Client management, token handling
- **Session Operations**: Session lifecycle management

## User Operations

### User Creation and Retrieval

**File**: `users_create.sql`

```sql
-- name: CreateUser :one
INSERT INTO users (did, handle, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserByDID :one
SELECT * FROM users WHERE did = $1;

-- name: GetUserByHandle :one
SELECT * FROM users WHERE handle = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: UpdateUser :one
UPDATE users 
SET handle = $2, email = $3, updated_at = NOW()
WHERE did = $1
RETURNING *;
```

**Generated Go Code**:
```go
func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
    row := q.db.QueryRow(ctx, createUser, arg.Did, arg.Handle, arg.Email, arg.PasswordHash)
    var i User
    err := row.Scan(
        &i.ID,
        &i.Did,
        &i.Handle,
        &i.Email,
        &i.PasswordHash,
        &i.CreatedAt,
        &i.UpdatedAt,
    )
    return i, err
}
```

### User Authentication

**File**: `users_auth.sql`

```sql
-- name: ValidateUserPassword :one
SELECT did, password_hash FROM users WHERE did = $1;

-- name: UpdateUserPassword :exec
UPDATE users 
SET password_hash = $2, updated_at = NOW()
WHERE did = $1;
```

## Session Management

### Session Operations

**File**: `sessions.sql`

```sql
-- name: CreateUserSession :one
INSERT INTO user_sessions (session_id, user_did, handle, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserSession :one
SELECT * FROM user_sessions WHERE session_id = $1;

-- name: UpdateUserSessionLastAccessed :exec
UPDATE user_sessions 
SET last_accessed_at = NOW() 
WHERE session_id = $1;

-- name: DeleteUserSession :exec
DELETE FROM user_sessions WHERE session_id = $1;

-- name: ListUserSessions :many
SELECT * FROM user_sessions WHERE user_did = $1 ORDER BY created_at DESC;

-- name: CleanupExpiredSessions :exec
DELETE FROM user_sessions WHERE expires_at < NOW();
```

**Generated Go Code**:
```go
type CreateUserSessionParams struct {
    SessionID string    `json:"session_id"`
    UserDid   string    `json:"user_did"`
    Handle    string    `json:"handle"`
    ExpiresAt time.Time `json:"expires_at"`
}

func (q *Queries) CreateUserSession(ctx context.Context, arg CreateUserSessionParams) (UserSession, error) {
    row := q.db.QueryRow(ctx, createUserSession, arg.SessionID, arg.UserDid, arg.Handle, arg.ExpiresAt)
    var i UserSession
    err := row.Scan(
        &i.ID,
        &i.SessionID,
        &i.UserDid,
        &i.Handle,
        &i.ExpiresAt,
        &i.CreatedAt,
        &i.LastAccessedAt,
    )
    return i, err
}
```

## Content Operations

### Post Management

**File**: `posts.sql`

```sql
-- name: CreatePost :one
INSERT INTO posts (user_id, subforum_id, title, content, atproto_uri)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1;

-- name: GetPostByAtprotoURI :one
SELECT * FROM posts WHERE atproto_uri = $1;

-- name: ListPostsByUser :many
SELECT * FROM posts WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: ListPostsBySubforum :many
SELECT * FROM posts WHERE subforum_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdatePost :one
UPDATE posts 
SET title = $2, content = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePost :exec
DELETE FROM posts WHERE id = $1;
```

### Comment Management

**File**: `comments.sql`

```sql
-- name: CreateComment :one
INSERT INTO comments (user_id, post_id, parent_id, content, atproto_uri)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM comments WHERE id = $1;

-- name: GetCommentByAtprotoURI :one
SELECT * FROM comments WHERE atproto_uri = $1;

-- name: ListCommentsByPost :many
SELECT * FROM comments WHERE post_id = $1 ORDER BY created_at ASC;

-- name: ListCommentsByUser :many
SELECT * FROM comments WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateComment :one
UPDATE comments 
SET content = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments WHERE id = $1;
```

### Subforum Management

**File**: `subforums.sql`

```sql
-- name: CreateSubforum :one
INSERT INTO subforums (name, slug, description, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetSubforumBySlug :one
SELECT * FROM subforums WHERE slug = $1;

-- name: ListSubforums :many
SELECT * FROM subforums ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateSubforum :one
UPDATE subforums 
SET name = $2, description = $3, updated_at = NOW()
WHERE slug = $1
RETURNING *;

-- name: DeleteSubforum :exec
DELETE FROM subforums WHERE slug = $1;
```

## OAuth Operations

### Client Management

**File**: `oauth_clients.sql`

```sql
-- name: CreateOAuthClient :one
INSERT INTO oauth_clients (client_id, client_secret, name, redirect_uri)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetOAuthClient :one
SELECT * FROM oauth_clients WHERE client_id = $1;

-- name: ListOAuthClients :many
SELECT * FROM oauth_clients ORDER BY created_at DESC;

-- name: UpdateOAuthClient :one
UPDATE oauth_clients 
SET name = $2, redirect_uri = $3, updated_at = NOW()
WHERE client_id = $1
RETURNING *;

-- name: DeleteOAuthClient :exec
DELETE FROM oauth_clients WHERE client_id = $1;
```

### Authorization Code Management

**File**: `oauth_authorization_codes.sql`

```sql
-- name: CreateAuthorizationCode :one
INSERT INTO oauth_authorization_codes (code, client_id, user_did, redirect_uri, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAuthorizationCode :one
SELECT * FROM oauth_authorization_codes WHERE code = $1;

-- name: DeleteAuthorizationCode :exec
DELETE FROM oauth_authorization_codes WHERE code = $1;

-- name: CleanupExpiredAuthorizationCodes :exec
DELETE FROM oauth_authorization_codes WHERE expires_at < NOW();
```

### Access Token Management

**File**: `oauth_access_tokens.sql`

```sql
-- name: CreateAccessToken :one
INSERT INTO oauth_access_tokens (access_token, client_id, user_did, scope, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetAccessToken :one
SELECT * FROM oauth_access_tokens WHERE access_token = $1;

-- name: ListAccessTokensByUser :many
SELECT * FROM oauth_access_tokens WHERE user_did = $1 ORDER BY created_at DESC;

-- name: DeleteAccessToken :exec
DELETE FROM oauth_access_tokens WHERE access_token = $1;

-- name: CleanupExpiredAccessTokens :exec
DELETE FROM oauth_access_tokens WHERE expires_at < NOW();
```

## Subscription Operations

### Subforum Subscriptions

**File**: `subforum_subscriptions.sql`

```sql
-- name: CreateSubforumSubscription :one
INSERT INTO subforum_subscriptions (user_id, subforum_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetSubforumSubscription :one
SELECT * FROM subforum_subscriptions WHERE user_id = $1 AND subforum_id = $2;

-- name: ListUserSubscriptions :many
SELECT s.*, sf.name, sf.slug, sf.description
FROM subforum_subscriptions s
JOIN subforums sf ON s.subforum_id = sf.id
WHERE s.user_id = $1
ORDER BY s.created_at DESC;

-- name: ListSubforumSubscribers :many
SELECT s.*, u.handle, u.did
FROM subforum_subscriptions s
JOIN users u ON s.user_id = u.id
WHERE s.subforum_id = $1
ORDER BY s.created_at DESC;

-- name: DeleteSubforumSubscription :exec
DELETE FROM subforum_subscriptions WHERE user_id = $1 AND subforum_id = $2;
```

## Query Patterns

### Pagination

Most list queries support pagination:

```sql
-- name: ListPostsWithPagination :many
SELECT * FROM posts 
ORDER BY created_at DESC 
LIMIT $1 OFFSET $2;
```

**Usage**:
```go
posts, err := queries.ListPostsWithPagination(ctx, generated.ListPostsWithPaginationParams{
    Limit:  20,
    Offset: 0,
})
```

### Filtering

Queries support various filtering options:

```sql
-- name: ListPostsByUserAndSubforum :many
SELECT * FROM posts 
WHERE user_id = $1 AND subforum_id = $2
ORDER BY created_at DESC 
LIMIT $3 OFFSET $4;
```

### Search

Full-text search capabilities:

```sql
-- name: SearchPosts :many
SELECT * FROM posts 
WHERE title ILIKE '%' || $1 || '%' OR content ILIKE '%' || $1 || '%'
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;
```

## Error Handling

### Common Error Patterns

```go
// Handle not found errors
user, err := queries.GetUserByDID(ctx, did)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, fmt.Errorf("user not found")
    }
    return nil, fmt.Errorf("database error: %w", err)
}

// Handle constraint violations
_, err = queries.CreateUser(ctx, params)
if err != nil {
    if strings.Contains(err.Error(), "duplicate key") {
        return nil, fmt.Errorf("user already exists")
    }
    return nil, fmt.Errorf("failed to create user: %w", err)
}
```

## Performance Considerations

### Index Usage
- All foreign key columns are indexed
- Frequently queried columns have indexes
- Composite indexes for multi-column queries

### Query Optimization
- Use LIMIT for pagination
- Avoid SELECT * in production
- Use prepared statements (automatic with SQLC)

### Connection Management
- Connection pooling configured
- Proper context usage for timeouts
- Transaction management for complex operations

## Code Generation

### SQLC Configuration

**File**: `sqlc-pds.yaml`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/database/queries/pds"
    schema: "internal/database/migrations/pds"
    gen:
      go:
        package: "generated"
        out: "internal/database/generated/pds"
        sql_package: "pgx/v5"
        emit_json_tags: true
        emit_prepared_queries: true
        emit_interface: true
        emit_exact_table_names: true
```

### Generation Commands

```bash
# Generate SQLC code
task generate:sqlc

# Run migrations
task migrate:up

# Rollback migrations
task migrate:down
```

## References

- [SQLC Documentation](https://docs.sqlc.dev/)
- [Query Files](internal/database/queries/pds/)
- [Generated Code](internal/database/generated/pds/)
- [Migration Files](internal/database/migrations/pds/)
