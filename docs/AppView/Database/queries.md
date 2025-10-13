# AppView Database Queries

## Overview

The AppView uses SQLC for type-safe database operations. All queries are defined in SQL files and generated into Go code, ensuring type safety and preventing SQL injection vulnerabilities.

## Query Organization

### File Structure
- **Location**: `internal/database/queries/appview/`
- **Generated Code**: `internal/database/generated/appview/`
- **Configuration**: `sqlc-appview.yaml`

### Query Categories
- **User Operations**: User CRUD, profile management
- **Content Operations**: Posts, comments, subforums
- **Subscription Operations**: User subscriptions to subforums
- **Vote Operations**: User votes on posts and comments
- **RBAC Operations**: Role and permission management

## User Operations

### User Management

**File**: `users.sql`

```sql
-- name: CreateUser :one
INSERT INTO appview_users (did, handle, display_name, bio, avatar_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetUserByDID :one
SELECT * FROM appview_users WHERE did = $1;

-- name: GetUserByHandle :one
SELECT * FROM appview_users WHERE handle = $1;

-- name: UpdateUser :one
UPDATE appview_users 
SET display_name = $2, bio = $3, avatar_url = $4, updated_at = NOW()
WHERE did = $1
RETURNING *;

-- name: UpdateUserStats :exec
UPDATE appview_users 
SET post_count = $2, comment_count = $3, reputation = $4, updated_at = NOW()
WHERE did = $1;

-- name: ListUsers :many
SELECT * FROM appview_users ORDER BY created_at DESC LIMIT $1 OFFSET $2;
```

**Generated Go Code**:
```go
func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (AppviewUser, error) {
    row := q.db.QueryRow(ctx, createUser, arg.Did, arg.Handle, arg.DisplayName, arg.Bio, arg.AvatarUrl)
    var i AppviewUser
    err := row.Scan(
        &i.ID,
        &i.Did,
        &i.Handle,
        &i.DisplayName,
        &i.Bio,
        &i.AvatarUrl,
        &i.CreatedAt,
        &i.UpdatedAt,
        &i.PostCount,
        &i.CommentCount,
        &i.Reputation,
    )
    return i, err
}
```

## Content Operations

### Post Management

**File**: `posts.sql`

```sql
-- name: CreatePost :one
INSERT INTO appview_posts (atproto_uri, author_did, author_handle, subforum_slug, title, content)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM appview_posts WHERE id = $1;

-- name: GetPostByAtprotoURI :one
SELECT * FROM appview_posts WHERE atproto_uri = $1;

-- name: ListPosts :many
SELECT * FROM appview_posts ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: ListPostsBySubforum :many
SELECT * FROM appview_posts WHERE subforum_slug = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: ListPostsByAuthor :many
SELECT * FROM appview_posts WHERE author_did = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdatePost :one
UPDATE appview_posts 
SET title = $2, content = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdatePostStats :exec
UPDATE appview_posts 
SET upvotes = $2, downvotes = $3, comment_count = $4, score = $5, updated_at = NOW()
WHERE id = $1;

-- name: DeletePost :exec
DELETE FROM appview_posts WHERE id = $1;
```

### Comment Management

**File**: `comments.sql`

```sql
-- name: CreateComment :one
INSERT INTO appview_comments (atproto_uri, author_did, author_handle, post_id, parent_id, content)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM appview_comments WHERE id = $1;

-- name: GetCommentByAtprotoURI :one
SELECT * FROM appview_comments WHERE atproto_uri = $1;

-- name: ListCommentsByPost :many
SELECT * FROM appview_comments WHERE post_id = $1 ORDER BY created_at ASC;

-- name: ListCommentsByAuthor :many
SELECT * FROM appview_comments WHERE author_did = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: UpdateComment :one
UPDATE appview_comments 
SET content = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateCommentStats :exec
UPDATE appview_comments 
SET upvotes = $2, downvotes = $3, score = $4, updated_at = NOW()
WHERE id = $1;

-- name: DeleteComment :exec
DELETE FROM appview_comments WHERE id = $1;
```

### Subforum Management

**File**: `subforums.sql`

```sql
-- name: CreateSubforum :one
INSERT INTO appview_subforums (name, slug, description, created_by_did, created_by_handle)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetSubforumBySlug :one
SELECT * FROM appview_subforums WHERE slug = $1;

-- name: ListSubforums :many
SELECT * FROM appview_subforums ORDER BY created_at DESC LIMIT $1 OFFSET $2;

-- name: UpdateSubforum :one
UPDATE appview_subforums 
SET name = $2, description = $3, updated_at = NOW()
WHERE slug = $1
RETURNING *;

-- name: UpdateSubforumStats :exec
UPDATE appview_subforums 
SET subscriber_count = $2, post_count = $3, comment_count = $4, updated_at = NOW()
WHERE slug = $1;

-- name: DeleteSubforum :exec
DELETE FROM appview_subforums WHERE slug = $1;
```

## Subscription Operations

### Subforum Subscriptions

**File**: `subscriptions.sql`

```sql
-- name: CreateSubscription :one
INSERT INTO appview_subscriptions (user_did, user_handle, subforum_slug)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSubscription :one
SELECT * FROM appview_subscriptions WHERE user_did = $1 AND subforum_slug = $2;

-- name: ListUserSubscriptions :many
SELECT s.*, sf.name, sf.description
FROM appview_subscriptions s
JOIN appview_subforums sf ON s.subforum_slug = sf.slug
WHERE s.user_did = $1
ORDER BY s.created_at DESC;

-- name: ListSubforumSubscribers :many
SELECT s.*, u.handle, u.display_name
FROM appview_subscriptions s
JOIN appview_users u ON s.user_did = u.did
WHERE s.subforum_slug = $1
ORDER BY s.created_at DESC;

-- name: DeleteSubscription :exec
DELETE FROM appview_subscriptions WHERE user_did = $1 AND subforum_slug = $2;
```

## Vote Operations

### Post Votes

**File**: `votes.sql`

```sql
-- name: CreatePostVote :one
INSERT INTO appview_votes (user_did, post_id, vote_type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPostVote :one
SELECT * FROM appview_votes WHERE user_did = $1 AND post_id = $2;

-- name: UpdatePostVote :one
UPDATE appview_votes 
SET vote_type = $3, created_at = NOW()
WHERE user_did = $1 AND post_id = $2
RETURNING *;

-- name: DeletePostVote :exec
DELETE FROM appview_votes WHERE user_did = $1 AND post_id = $2;

-- name: ListPostVotes :many
SELECT * FROM appview_votes WHERE post_id = $1 ORDER BY created_at DESC;
```

### Comment Votes

```sql
-- name: CreateCommentVote :one
INSERT INTO appview_votes (user_did, comment_id, vote_type)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCommentVote :one
SELECT * FROM appview_votes WHERE user_did = $1 AND comment_id = $2;

-- name: UpdateCommentVote :one
UPDATE appview_votes 
SET vote_type = $3, created_at = NOW()
WHERE user_did = $1 AND comment_id = $2
RETURNING *;

-- name: DeleteCommentVote :exec
DELETE FROM appview_votes WHERE user_did = $1 AND comment_id = $2;

-- name: ListCommentVotes :many
SELECT * FROM appview_votes WHERE comment_id = $1 ORDER BY created_at DESC;
```

## RBAC Operations

### Permission Checking

**File**: `check_user_permission.sql`

```sql
-- name: CheckUserPermission :one
SELECT EXISTS(
    SELECT 1
    FROM user_roles ur
    JOIN role_permissions rp ON ur.role_id = rp.role_id
    JOIN permissions p ON rp.permission_id = p.id
    WHERE ur.user_did = $1
        AND p.name = $2
        AND ur.is_active = TRUE
        AND (ur.expires_at IS NULL OR ur.expires_at > NOW())
        AND (
            -- Platform permissions (no subforum restriction)
            (p.resource_type = 'platform' AND ur.subforum_id IS NULL)
            OR
            -- Subforum permissions (specific subforum or any subforum)
            (p.resource_type IN ('subforum', 'post', 'comment', 'vote') AND (
                ur.subforum_id = $3 OR ur.subforum_id IS NULL
            ))
        )
) as has_permission;
```

### Role Management

**File**: `roles.sql`

```sql
-- name: CreateRole :one
INSERT INTO roles (name, description, is_platform_role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetRoleByName :one
SELECT * FROM roles WHERE name = $1;

-- name: ListRoles :many
SELECT * FROM roles ORDER BY name;

-- name: UpdateRole :one
UPDATE roles 
SET description = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles WHERE id = $1;
```

### User Role Management

**File**: `user_roles.sql`

```sql
-- name: AssignRole :one
INSERT INTO user_roles (user_did, role_id, subforum_id, granted_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: RevokeRole :exec
DELETE FROM user_roles WHERE user_did = $1 AND role_id = $2 AND subforum_id = $3;

-- name: GetUserRoles :many
SELECT ur.*, r.name as role_name, r.description as role_description
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.user_did = $1 AND ur.is_active = TRUE;

-- name: GetUserRolesBySubforum :many
SELECT ur.*, r.name as role_name, r.description as role_description
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
WHERE ur.user_did = $1 AND ur.subforum_id = $2 AND ur.is_active = TRUE;
```

## Statistics Queries

### User Statistics

**File**: `user_stats.sql`

```sql
-- name: GetUserStats :one
SELECT 
    post_count,
    comment_count,
    reputation
FROM appview_users 
WHERE did = $1;

-- name: UpdateUserPostCount :exec
UPDATE appview_users 
SET post_count = post_count + $2, updated_at = NOW()
WHERE did = $1;

-- name: UpdateUserCommentCount :exec
UPDATE appview_users 
SET comment_count = comment_count + $2, updated_at = NOW()
WHERE did = $1;
```

### Subforum Statistics

**File**: `subforum_stats.sql`

```sql
-- name: GetSubforumStats :one
SELECT 
    subscriber_count,
    post_count,
    comment_count
FROM appview_subforums 
WHERE slug = $1;

-- name: UpdateSubforumSubscriberCount :exec
UPDATE appview_subforums 
SET subscriber_count = subscriber_count + $2, updated_at = NOW()
WHERE slug = $1;

-- name: UpdateSubforumPostCount :exec
UPDATE appview_subforums 
SET post_count = post_count + $2, updated_at = NOW()
WHERE slug = $1;
```

### Post Statistics

**File**: `post_stats.sql`

```sql
-- name: GetPostStats :one
SELECT 
    upvotes,
    downvotes,
    comment_count,
    score
FROM appview_posts 
WHERE id = $1;

-- name: UpdatePostVoteCounts :exec
UPDATE appview_posts 
SET upvotes = $2, downvotes = $3, score = $4, updated_at = NOW()
WHERE id = $1;

-- name: UpdatePostCommentCount :exec
UPDATE appview_posts 
SET comment_count = comment_count + $2, updated_at = NOW()
WHERE id = $1;
```

## Query Patterns

### Pagination

Most list queries support pagination:

```sql
-- name: ListPostsWithPagination :many
SELECT * FROM appview_posts 
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
-- name: ListPostsBySubforumAndAuthor :many
SELECT * FROM appview_posts 
WHERE subforum_slug = $1 AND author_did = $2
ORDER BY created_at DESC 
LIMIT $3 OFFSET $4;
```

### Search

Full-text search capabilities:

```sql
-- name: SearchPosts :many
SELECT * FROM appview_posts 
WHERE title ILIKE '%' || $1 || '%' OR content ILIKE '%' || $1 || '%'
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;
```

### Aggregation

Statistics and aggregation queries:

```sql
-- name: GetSubforumStats :one
SELECT 
    COUNT(*) as total_posts,
    AVG(score) as avg_score,
    MAX(created_at) as latest_post
FROM appview_posts 
WHERE subforum_slug = $1;
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

**File**: `sqlc-appview.yaml`

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "internal/database/queries/appview"
    schema: "internal/database/migrations/appview"
    gen:
      go:
        package: "generated"
        out: "internal/database/generated/appview"
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
- [Query Files](internal/database/queries/appview/)
- [Generated Code](internal/database/generated/appview/)
- [Migration Files](internal/database/migrations/appview/)
