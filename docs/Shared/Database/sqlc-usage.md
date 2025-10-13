# SQLC Usage

## Overview

HashPost uses SQLC for type-safe database operations across both PDS and AppView services. SQLC generates Go code from SQL queries, ensuring type safety and preventing SQL injection vulnerabilities.

## SQLC Configuration

### PDS Configuration

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

### AppView Configuration

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

## Query Organization

### File Structure

**PDS Queries**:
```
internal/database/queries/pds/
├── users.sql              # User operations
├── sessions.sql           # Session management
├── posts.sql              # Post operations
├── comments.sql           # Comment operations
├── subforums.sql          # Subforum operations
├── oauth_clients.sql      # OAuth client management
├── oauth_authorization_codes.sql
└── oauth_access_tokens.sql
```

**AppView Queries**:
```
internal/database/queries/appview/
├── users.sql              # User operations
├── posts.sql              # Post operations
├── comments.sql           # Comment operations
├── subforums.sql          # Subforum operations
├── subscriptions.sql      # Subscription operations
├── votes.sql              # Vote operations
├── roles.sql              # Role management
├── permissions.sql        # Permission management
├── user_roles.sql         # User role assignments
└── check_user_permission.sql
```

### Query Categories

**CRUD Operations**:
- **Create**: `CreateUser`, `CreatePost`, `CreateComment`
- **Read**: `GetUserByDID`, `GetPostByID`, `ListPosts`
- **Update**: `UpdateUser`, `UpdatePost`, `UpdateComment`
- **Delete**: `DeleteUser`, `DeletePost`, `DeleteComment`

**Specialized Operations**:
- **Authentication**: `ValidateUserPassword`, `CreateUserSession`
- **RBAC**: `CheckUserPermission`, `AssignRole`, `RevokeRole`
- **Statistics**: `UpdateUserStats`, `UpdatePostStats`
- **Search**: `SearchPosts`, `SearchUsers`

## Query Patterns

### Basic CRUD Operations

**User Creation**:
```sql
-- name: CreateUser :one
INSERT INTO users (did, handle, email, password_hash)
VALUES ($1, $2, $3, $4)
RETURNING *;
```

**Generated Go Code**:
```go
type CreateUserParams struct {
    Did          string `json:"did"`
    Handle       string `json:"handle"`
    Email        string `json:"email"`
    PasswordHash string `json:"password_hash"`
}

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

**User Retrieval**:
```sql
-- name: GetUserByDID :one
SELECT * FROM users WHERE did = $1;
```

**Generated Go Code**:
```go
func (q *Queries) GetUserByDID(ctx context.Context, did string) (User, error) {
    row := q.db.QueryRow(ctx, getUserByDID, did)
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

### Pagination Queries

**List Posts with Pagination**:
```sql
-- name: ListPosts :many
SELECT * FROM posts 
ORDER BY created_at DESC 
LIMIT $1 OFFSET $2;
```

**Generated Go Code**:
```go
type ListPostsParams struct {
    Limit  int32 `json:"limit"`
    Offset int32 `json:"offset"`
}

func (q *Queries) ListPosts(ctx context.Context, arg ListPostsParams) ([]Post, error) {
    rows, err := q.db.Query(ctx, listPosts, arg.Limit, arg.Offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    var items []Post
    for rows.Next() {
        var i Post
        if err := rows.Scan(
            &i.ID,
            &i.UserID,
            &i.SubforumID,
            &i.Title,
            &i.Content,
            &i.AtprotoURI,
            &i.CreatedAt,
            &i.UpdatedAt,
        ); err != nil {
            return nil, err
        }
        items = append(items, i)
    }
    if err := rows.Err(); err != nil {
        return nil, err
    }
    return items, nil
}
```

### Complex Queries

**Permission Checking**:
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

**Generated Go Code**:
```go
type CheckUserPermissionParams struct {
    UserDid    string  `json:"user_did"`
    Permission string  `json:"permission"`
    SubforumID *string `json:"subforum_id"`
}

type CheckUserPermissionRow struct {
    HasPermission bool `json:"has_permission"`
}

func (q *Queries) CheckUserPermission(ctx context.Context, arg CheckUserPermissionParams) (CheckUserPermissionRow, error) {
    row := q.db.QueryRow(ctx, checkUserPermission, arg.UserDid, arg.Permission, arg.SubforumID)
    var i CheckUserPermissionRow
    err := row.Scan(&i.HasPermission)
    return i, err
}
```

### Statistics Queries

**Update User Statistics**:
```sql
-- name: UpdateUserStats :exec
UPDATE users 
SET post_count = $2, comment_count = $3, reputation = $4, updated_at = NOW()
WHERE did = $1;
```

**Generated Go Code**:
```go
type UpdateUserStatsParams struct {
    Did           string `json:"did"`
    PostCount     int32  `json:"post_count"`
    CommentCount  int32  `json:"comment_count"`
    Reputation    int32  `json:"reputation"`
}

func (q *Queries) UpdateUserStats(ctx context.Context, arg UpdateUserStatsParams) error {
    _, err := q.db.Exec(ctx, updateUserStats, arg.Did, arg.PostCount, arg.CommentCount, arg.Reputation)
    return err
}
```

## Code Generation

### Generation Commands

**PDS Code Generation**:
```bash
# Generate PDS SQLC code
task generate:sqlc

# Run PDS migrations
task migrate:up

# Rollback PDS migrations
task migrate:down
```

**AppView Code Generation**:
```bash
# Generate AppView SQLC code
task generate:sqlc

# Run AppView migrations
task migrate:up

# Rollback AppView migrations
task migrate:down
```

### Generated Code Structure

**PDS Generated Code**:
```
internal/database/generated/pds/
├── db.go              # Database connection
├── models.go          # Generated models
├── querier.go         # Query interface
├── users.sql.go       # User queries
├── sessions.sql.go    # Session queries
├── posts.sql.go       # Post queries
└── ...
```

**AppView Generated Code**:
```
internal/database/generated/appview/
├── db.go              # Database connection
├── models.go          # Generated models
├── querier.go         # Query interface
├── users.sql.go       # User queries
├── posts.sql.go       # Post queries
├── rbac.sql.go        # RBAC queries
└── ...
```

## Database Connection

### Connection Management

**PDS Database Connection**:
```go
// internal/database/connection.go
func NewPDSConnection(databaseURL string) (*sql.DB, error) {
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)

    // Test connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return db, nil
}
```

**AppView Database Connection**:
```go
// internal/database/connection.go
func NewAppViewConnection(databaseURL string) (*sql.DB, error) {
    db, err := sql.Open("postgres", databaseURL)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

    // Configure connection pool
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(5 * time.Minute)

    // Test connection
    if err := db.Ping(); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }

    return db, nil
}
```

### Query Interface

**PDS Query Interface**:
```go
// internal/database/generated/pds/querier.go
type Querier interface {
    CreateUser(ctx context.Context, arg CreateUserParams) (User, error)
    GetUserByDID(ctx context.Context, did string) (User, error)
    GetUserByHandle(ctx context.Context, handle string) (User, error)
    UpdateUser(ctx context.Context, arg UpdateUserParams) (User, error)
    DeleteUser(ctx context.Context, id uuid.UUID) error
    // ... other methods
}
```

**AppView Query Interface**:
```go
// internal/database/generated/appview/querier.go
type Querier interface {
    CreateUser(ctx context.Context, arg CreateUserParams) (AppviewUser, error)
    GetUserByDID(ctx context.Context, did string) (AppviewUser, error)
    GetUserByHandle(ctx context.Context, handle string) (AppviewUser, error)
    UpdateUser(ctx context.Context, arg UpdateUserParams) (AppviewUser, error)
    DeleteUser(ctx context.Context, id uuid.UUID) error
    // ... other methods
}
```

## Error Handling

### Common Error Patterns

**Not Found Errors**:
```go
// Handle not found errors
user, err := queries.GetUserByDID(ctx, did)
if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        return nil, fmt.Errorf("user not found")
    }
    return nil, fmt.Errorf("database error: %w", err)
}
```

**Constraint Violations**:
```go
// Handle constraint violations
_, err = queries.CreateUser(ctx, params)
if err != nil {
    if strings.Contains(err.Error(), "duplicate key") {
        return nil, fmt.Errorf("user already exists")
    }
    return nil, fmt.Errorf("failed to create user: %w", err)
}
```

**Transaction Errors**:
```go
// Handle transaction errors
tx, err := db.BeginTx(ctx, nil)
if err != nil {
    return fmt.Errorf("failed to begin transaction: %w", err)
}
defer tx.Rollback()

// Perform operations
if err := queries.CreateUser(ctx, params); err != nil {
    return fmt.Errorf("failed to create user: %w", err)
}

if err := tx.Commit(); err != nil {
    return fmt.Errorf("failed to commit transaction: %w", err)
}
```

## Performance Considerations

### Query Optimization

**Index Usage**:
```sql
-- Create indexes for common queries
CREATE INDEX idx_users_did ON users(did);
CREATE INDEX idx_users_handle ON users(handle);
CREATE INDEX idx_posts_author_did ON posts(author_did);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
```

**Query Performance**:
```go
// Use prepared statements (automatic with SQLC)
func (q *Queries) GetUserByDID(ctx context.Context, did string) (User, error) {
    row := q.db.QueryRow(ctx, getUserByDID, did)
    // ... scan results
}
```

### Connection Pooling

**Pool Configuration**:
```go
// Configure connection pool
db.SetMaxOpenConns(25)    // Maximum open connections
db.SetMaxIdleConns(25)    // Maximum idle connections
db.SetConnMaxLifetime(5 * time.Minute)  // Connection lifetime
```

**Connection Health**:
```go
// Health check for database connection
func (q *Queries) HealthCheck() error {
    if err := q.db.Ping(); err != nil {
        return fmt.Errorf("database connection lost: %w", err)
    }
    return nil
}
```

## Testing

### Unit Testing

**Mock Queries**:
```go
// Create mock queries for testing
type MockQueries struct {
    CreateUserFunc func(ctx context.Context, arg CreateUserParams) (User, error)
    GetUserByDIDFunc func(ctx context.Context, did string) (User, error)
}

func (m *MockQueries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
    if m.CreateUserFunc != nil {
        return m.CreateUserFunc(ctx, arg)
    }
    return User{}, fmt.Errorf("not implemented")
}
```

**Test Setup**:
```go
func TestCreateUser(t *testing.T) {
    // Setup test database
    db := setupTestDB(t)
    queries := generated.New(db)

    // Test user creation
    user, err := queries.CreateUser(ctx, generated.CreateUserParams{
        Did:    "did:plc:test",
        Handle: "testuser",
        Email:  "test@example.com",
    })

    require.NoError(t, err)
    assert.Equal(t, "did:plc:test", user.Did)
    assert.Equal(t, "testuser", user.Handle)
}
```

### Integration Testing

**Database Setup**:
```go
func setupTestDB(t *testing.T) *sql.DB {
    // Create test database
    db, err := sql.Open("postgres", testDatabaseURL)
    require.NoError(t, err)

    // Run migrations
    err = runMigrations(db)
    require.NoError(t, err)

    // Cleanup after test
    t.Cleanup(func() {
        db.Close()
    })

    return db
}
```

## Best Practices

### Query Design

1. **Single Responsibility**: Each query should have a single purpose
2. **Parameter Validation**: Validate parameters before query execution
3. **Error Handling**: Handle errors appropriately
4. **Performance**: Optimize queries for performance

### Code Organization

1. **File Structure**: Organize queries by domain
2. **Naming Conventions**: Use consistent naming patterns
3. **Documentation**: Document complex queries
4. **Versioning**: Version query changes

### Development Workflow

1. **Write SQL**: Write SQL queries first
2. **Generate Code**: Generate Go code with SQLC
3. **Test Queries**: Test queries thoroughly
4. **Deploy**: Deploy with confidence

## References

- [SQLC Documentation](https://docs.sqlc.dev/)
- [PDS Queries](../PDS/Database/queries.md)
- [AppView Queries](../AppView/Database/queries.md)
- [Database Migrations](../database/migrations/)
