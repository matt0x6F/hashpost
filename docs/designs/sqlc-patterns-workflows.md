# sqlc Patterns and Workflows for HashPost

## Project Overview

This document defines the patterns and workflows for using sqlc in the HashPost rebuild. sqlc will be used from the start to provide better performance, type safety, and maintainability.

## Design Goals

### Primary Objectives
- **Zero Runtime Overhead**: sqlc generates pure Go code with no reflection
- **Type Safety**: Compile-time query validation and type checking
- **Performance**: Optimized queries for HashPost's complex permission system
- **Maintainability**: Clear separation between SQL and Go code

### Technical Goals
- Define query organization patterns
- Establish development workflows
- Create testing strategies
- Plan fresh implementation approach

## Architecture Decisions

### Query Organization
- **Directory Structure**: `internal/database/queries/` for SQL files
- **Naming Convention**: `{entity}_{operation}.sql` (e.g., `users_create.sql`)
- **Query Categories**: CRUD, complex queries, and analytics
- **Parameter Validation**: Use sqlc's built-in parameter validation

### Code Generation
- **Output Directory**: `internal/database/generated/`
- **Package Structure**: Separate packages for different domains
- **Type Safety**: Leverage sqlc's generated types
- **Interface Pattern**: Create interfaces for testability

## sqlc Patterns

### 1. Query File Organization

```
internal/database/queries/
├── users/
│   ├── create.sql
│   ├── get_by_id.sql
│   ├── get_by_email.sql
│   ├── update.sql
│   └── delete.sql
├── posts/
│   ├── create.sql
│   ├── get_by_id.sql
│   ├── get_by_subforum.sql
│   ├── get_by_pseudonym.sql
│   └── update_score.sql
├── permissions/
│   ├── check_permission.sql
│   ├── get_user_roles.sql
│   └── get_role_permissions.sql
└── analytics/
    ├── post_stats.sql
    ├── user_activity.sql
    └── moderation_stats.sql
```

### 2. Query Naming Patterns

```sql
-- users/create.sql
-- name: CreateUser :one
INSERT INTO users (email, password_hash, created_at)
VALUES ($1, $2, $3)
RETURNING *;

-- users/get_by_id.sql
-- name: GetUserByID :one
SELECT * FROM users WHERE user_id = $1;

-- posts/get_by_subforum.sql
-- name: GetPostsBySubforum :many
SELECT p.*, u.display_name, s.name as subforum_name
FROM posts p
JOIN pseudonyms u ON p.pseudonym_id = u.pseudonym_id
JOIN subforums s ON p.subforum_id = s.subforum_id
WHERE p.subforum_id = $1
  AND p.is_deleted = false
ORDER BY p.created_at DESC
LIMIT $2 OFFSET $3;
```

### 3. Complex Query Patterns

```sql
-- permissions/check_permission.sql
-- name: CheckUserPermission :one
SELECT EXISTS(
  SELECT 1
  FROM user_roles ur
  JOIN role_permissions rp ON ur.role_id = rp.role_id
  WHERE ur.user_id = $1
    AND rp.permission = $2
    AND ur.expires_at > NOW()
) as has_permission;

-- analytics/post_stats.sql
-- name: GetPostStats :one
SELECT 
  COUNT(*) as total_posts,
  COUNT(CASE WHEN created_at > $1 THEN 1 END) as recent_posts,
  AVG(score) as avg_score
FROM posts
WHERE subforum_id = $2
  AND is_deleted = false;
```

## Development Workflows

### 1. Query Development Workflow

```bash
# 1. Write SQL query
vim internal/database/queries/users/create.sql

# 2. Generate Go code
make sqlc-generate

# 3. Update DAO to use generated code
vim internal/database/dao/users.go

# 4. Run tests
make test

# 5. Update integration tests
vim internal/database/dao/users_test.go
```

### 2. Testing Workflow

```bash
# 1. Create test database
make test-db-setup

# 2. Run sqlc tests
make sqlc-test

# 3. Run DAO tests
make test-dao

# 4. Run integration tests
make test-integration

# 5. Cleanup
make test-db-cleanup
```

### 3. Schema Evolution Workflow

```bash
# 1. Create migration
make migrate-create name=add_user_preferences

# 2. Write migration SQL
vim internal/database/migrations/xxx_add_user_preferences.sql

# 3. Apply migration
make migrate-up

# 4. Update sqlc queries
vim internal/database/queries/users/update_preferences.sql

# 5. Regenerate code
make sqlc-generate

# 6. Update tests
make test
```

## DAO Patterns

### 1. Interface-Based DAOs

```go
// internal/database/dao/interfaces.go
type UserDAO interface {
    CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
    GetUserByID(ctx context.Context, userID int64) (User, error)
    GetUserByEmail(ctx context.Context, email string) (User, error)
    UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
    DeleteUser(ctx context.Context, userID int64) error
}

// internal/database/dao/users.go
type userDAO struct {
    db *sql.DB
    queries *db.Queries
}

func NewUserDAO(db *sql.DB) UserDAO {
    return &userDAO{
        db: db,
        queries: db.New(db),
    }
}

func (dao *userDAO) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
    return dao.queries.CreateUser(ctx, arg)
}
```

### 2. Transaction Support

```go
// internal/database/dao/posts.go
func (dao *postDAO) CreatePostWithVote(ctx context.Context, arg CreatePostParams) (Post, error) {
    tx, err := dao.db.BeginTx(ctx, nil)
    if err != nil {
        return Post{}, err
    }
    defer tx.Rollback()

    queries := dao.queries.WithTx(tx)
    
    post, err := queries.CreatePost(ctx, arg)
    if err != nil {
        return Post{}, err
    }

    // Create initial vote
    _, err = queries.CreateVote(ctx, CreateVoteParams{
        PostID: post.PostID,
        UserID: arg.UserID,
        VoteType: "upvote",
    })
    if err != nil {
        return Post{}, err
    }

    return post, tx.Commit()
}
```

### 3. Complex Query Composition

```go
// internal/database/dao/posts.go
func (dao *postDAO) GetPostsWithFilters(ctx context.Context, filters PostFilters) ([]PostWithDetails, error) {
    var posts []PostWithDetails
    
    if filters.SubforumID != nil {
        posts, err = dao.queries.GetPostsBySubforum(ctx, GetPostsBySubforumParams{
            SubforumID: *filters.SubforumID,
            Limit: filters.Limit,
            Offset: filters.Offset,
        })
    } else if filters.PseudonymID != nil {
        posts, err = dao.queries.GetPostsByPseudonym(ctx, GetPostsByPseudonymParams{
            PseudonymID: *filters.PseudonymID,
            Limit: filters.Limit,
            Offset: filters.Offset,
        })
    } else {
        posts, err = dao.queries.GetAllPosts(ctx, GetAllPostsParams{
            Limit: filters.Limit,
            Offset: filters.Offset,
        })
    }
    
    return posts, err
}
```

## Testing Patterns

### 1. Unit Tests

```go
// internal/database/dao/users_test.go
func TestUserDAO_CreateUser(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    dao := NewUserDAO(db)
    
    user, err := dao.CreateUser(ctx, CreateUserParams{
        Email: "test@example.com",
        PasswordHash: "hashed_password",
    })
    
    require.NoError(t, err)
    assert.Equal(t, "test@example.com", user.Email)
}
```

### 2. Integration Tests

```go
// internal/database/dao/posts_integration_test.go
func TestPostDAO_ComplexWorkflow(t *testing.T) {
    db := setupTestDB(t)
    defer cleanupTestDB(t, db)
    
    // Create test data
    user := createTestUser(t, db)
    subforum := createTestSubforum(t, db)
    
    dao := NewPostDAO(db)
    
    // Test post creation
    post, err := dao.CreatePost(ctx, CreatePostParams{
        SubforumID: subforum.SubforumID,
        PseudonymID: user.PseudonymID,
        Title: "Test Post",
        Content: "Test content",
    })
    
    require.NoError(t, err)
    assert.Equal(t, "Test Post", post.Title)
    
    // Test post retrieval
    retrieved, err := dao.GetPostByID(ctx, post.PostID)
    require.NoError(t, err)
    assert.Equal(t, post.PostID, retrieved.PostID)
}
```

## PDS-Only Database Access

### PDS-Only Database Architecture
- **Database Layer**: Only PDS component accesses PostgreSQL database
- **Query Organization**: PDS handles all queries - atproto protocol, business logic, and RBAC
- **Transaction Boundaries**: PDS manages all transaction boundaries
- **Data Consistency**: PDS ensures both atproto data integrity and business rule integrity
- **RBAC**: PDS handles all role-based access control and business logic
- **AppView**: Stateless aggregator with no database access

### PDS-Only Query Patterns
- **PDS**: Handles all database access - atproto protocol, business logic, and RBAC
- **AppView**: No database access, aggregates data from PDS via APIs
- **Integration**: AppView calls PDS APIs for all data operations
- **Benefits**: Atproto compliance, simpler architecture, proven pattern

### Component-Specific Query Patterns
```
internal/database/queries/
├── pds/                    # PDS handles all queries
│   ├── atproto/
│   │   ├── did_resolution.sql
│   │   ├── session_validation.sql
│   │   └── identity_management.sql
│   ├── protocol/
│   │   ├── profile_management.sql
│   │   └── content_management.sql
│   ├── business/
│   │   ├── user_management.sql
│   │   ├── post_management.sql
│   │   ├── moderation.sql
│   │   └── rbac_permissions.sql
│   └── analytics/
│       ├── engagement_stats.sql
│       └── content_analytics.sql
└── appview/                # AppView has no database queries
    └── (no database access)
```

## Implementation Strategy

### Phase 1: Setup
- [ ] Install sqlc and configure
- [ ] Set up query directory structure
- [ ] Create initial sqlc.yaml configuration
- [ ] Set up code generation workflow
- [ ] Design PDS-only query organization

### Phase 2: Core Queries
- [ ] Build user-related queries
- [ ] Build authentication queries
- [ ] Build basic CRUD operations
- [ ] Create corresponding DAOs

### Phase 3: Complex Queries
- [ ] Build permission system queries
- [ ] Build forum and post queries
- [ ] Build analytics queries
- [ ] Create integration tests

### Phase 4: Optimization
- [ ] Optimize query performance
- [ ] Add query-specific indexes
- [ ] Implement query caching where appropriate
- [ ] Performance testing

## Success Criteria

### Performance
- Query execution time < 50ms for simple queries
- Complex queries < 200ms
- Zero runtime overhead from sqlc
- Memory usage reduction compared to Bob

### Developer Experience
- Clear query organization
- Easy testing setup
- Good error messages
- Fast code generation

### Maintainability
- Single source of truth for SQL
- Type-safe generated code
- Clear separation of concerns
- Comprehensive test coverage

## Implementation Complete

### Phase 1: Foundation ✅
1. **✅ Set up sqlc configuration** - Created sqlc.yaml and directory structure
2. **✅ Build comprehensive schema** - Complete HashPost database schema with users, subforums, posts, comments, votes
3. **✅ Generate code** - sqlc generating Go code successfully
4. **✅ Type safety** - Using pgx/v5 with proper type overrides for UUID and timestamps

### Next Steps
1. **Build core queries** - Create CRUD queries for all entities
2. **Create DAOs** - Build DAOs using sqlc generated code
3. **Test integration** - Ensure everything works together
4. **Performance testing** - Validate performance improvements

## Notes

- sqlc provides excellent type safety and performance
- Focus on query organization and testing patterns
- Consider using transactions for complex operations
- Plan for fresh implementation with sqlc from the start

---

*Last Updated: [Current Date]*
*Status: Design Phase*
