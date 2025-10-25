# Database Separation Strategy

## Overview

HashPost implements a database separation strategy where the PDS and AppView maintain separate databases optimized for their specific use cases. This separation enables independent scaling, optimization, and maintenance of each service.

## Database Architecture

### PDS Database

**Database Name**: `hashpost_pds_dev` (development) / `hashpost_pds` (production)

**Purpose**: Store canonical atproto records plus forum-specific tables

**Characteristics**:
- **Canonical Data**: Source of truth for all atproto records and forum content
- **Protocol Compliance**: Follows atproto data model plus forum extensions
- **Write-Heavy**: Optimized for record creation and updates
- **Data Integrity**: Ensures protocol compliance and consistency

**Schema Design**:
```sql
-- Canonical user records
CREATE TABLE users (
    id UUID PRIMARY KEY,
    did VARCHAR(255) UNIQUE NOT NULL,
    handle VARCHAR(255) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE,
    password_hash TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Forum-specific tables in PDS
CREATE TABLE subforums (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    subforum_id UUID REFERENCES subforums(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    atproto_uri VARCHAR(500) UNIQUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    vote_type VARCHAR(10) NOT NULL CHECK (vote_type IN ('up', 'down')),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### AppView Database

**Database Name**: `hashpost_appview_dev` (development) / `hashpost_appview` (production)

**Purpose**: Store denormalized data optimized for user-facing queries

**Characteristics**:
- **Denormalized Data**: Pre-computed statistics and aggregated data
- **Read-Heavy**: Optimized for user-facing queries
- **Performance**: Indexes and denormalization for fast queries
- **User Experience**: Optimized for forum functionality

**Schema Design**:
```sql
-- Denormalized user data
CREATE TABLE appview_users (
    id UUID PRIMARY KEY,
    did VARCHAR(255) UNIQUE NOT NULL,
    handle VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    bio TEXT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized stats
    post_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    reputation INTEGER DEFAULT 0
);

-- Denormalized post data
CREATE TABLE appview_posts (
    id UUID PRIMARY KEY,
    atproto_uri VARCHAR(500) UNIQUE NOT NULL,
    author_did VARCHAR(255) NOT NULL,
    author_handle VARCHAR(255) NOT NULL,
    subforum_slug VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized stats
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    score INTEGER DEFAULT 0
);
```

## Separation Rationale

### Data Ownership

**PDS Database**:
- **Owns**: Canonical atproto records
- **Responsibility**: Protocol compliance and data integrity
- **Access Pattern**: Direct database access for atproto operations
- **Consistency**: Strong consistency for protocol compliance

**AppView Database**:
- **Owns**: Denormalized application data
- **Responsibility**: User experience and performance
- **Access Pattern**: Event-driven updates from PDS
- **Consistency**: Eventual consistency for user experience

### Performance Optimization

**PDS Database**:
- **Write Optimization**: Optimized for record creation and updates
- **Protocol Compliance**: Follows atproto data model
- **Data Integrity**: Ensures canonical record structure
- **Indexing**: Indexes for atproto protocol operations

**AppView Database**:
- **Read Optimization**: Optimized for user-facing queries
- **Denormalization**: Pre-computed statistics and aggregated data
- **Performance**: Indexes for forum-specific queries
- **Caching**: Application-level caching for frequently accessed data

### Scalability

**Independent Scaling**:
- **PDS**: Can scale based on atproto protocol load
- **AppView**: Can scale based on user-facing application load
- **Database**: Each service can optimize its database independently

**Load Distribution**:
- **PDS**: Handles protocol compliance and data storage
- **AppView**: Handles user interactions and business logic
- **Separation**: Clear boundaries between services

## Data Synchronization

### Event-Driven Synchronization

**Flow**:
1. **PDS**: User performs action (create/update/delete)
2. **PDS**: Stores canonical record in PDS database
3. **PDS**: Publishes event to NATS JetStream
4. **AppView**: Consumes event from NATS JetStream
5. **AppView**: Updates denormalized data in AppView database

**Event Types**:
- **Record Events**: `record.created`, `record.updated`, `record.deleted`
- **Identity Events**: `identity.resolved`, `session.created`

### Consistency Model

**Eventual Consistency**:
- **PDS**: Source of truth for canonical records
- **AppView**: Eventually consistent denormalized data
- **Sync Delay**: Minimal delay due to event processing

**Consistency Guarantees**:
- **At-Least-Once Delivery**: NATS JetStream ensures event delivery
- **Idempotency**: AppView handles duplicate events gracefully
- **Error Handling**: Failed events go to dead letter queue

## Implementation Details

### PDS Database Operations

**Canonical Record Storage**:
```go
// Store canonical atproto record
func (s *Server) handleCreateRecord(w http.ResponseWriter, r *http.Request) {
    // ... authentication and validation ...

    // Create record in database
    post, err := s.db.CreatePost(ctx, &generated.CreatePostParams{
        UserID:     userID,
        SubforumID: subforumID,
        Title:      req.Record["title"].(string),
        Content:    req.Record["content"].(string),
        AtprotoURI: uri,
    })

    // Publish event
    err = s.eventStream.PublishRecordEvent(
        r.Context(),
        EventTypeRecordCreated,
        session.DID,
        req.Collection,
        req.Record,
        uri,
        cid,
    )

    // ... return response ...
}
```

**SQLC Queries**:
```sql
-- name: CreatePost :one
INSERT INTO posts (user_id, subforum_id, title, content, atproto_uri)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetPostByAtprotoURI :one
SELECT * FROM posts WHERE atproto_uri = $1;

-- name: UpdatePost :one
UPDATE posts 
SET title = $2, content = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;
```

### AppView Database Operations

**Denormalized Data Storage**:
```go
func (ec *EventConsumer) handlePostCreated(event AtprotoEvent) error {
    // Extract post data from the record
    text, ok := event.Record["text"].(string)
    if !ok {
        return fmt.Errorf("invalid text field in post record")
    }

    // Resolve author handle from DID
    authorHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
    if err != nil {
        authorHandle = "unknown"
    }

    // Create AppView post data
    post := &AppViewPost{
        ID:           uuid.New(),
        AtprotoURI:   event.URI,
        AuthorDID:    event.Repo,
        AuthorHandle: authorHandle,
        SubforumSlug: subforumSlug,
        Title:        text,
        Content:      text,
        CreatedAt:    createdAt,
        UpdatedAt:    createdAt,
    }

    // Store in AppView database
    if err := ec.db.CreatePost(post); err != nil {
        return fmt.Errorf("failed to create post in AppView: %w", err)
    }

    return nil
}
```

**SQLC Queries**:
```sql
-- name: CreatePost :one
INSERT INTO appview_posts (atproto_uri, author_did, author_handle, subforum_slug, title, content)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: ListPostsBySubforum :many
SELECT * FROM appview_posts 
WHERE subforum_slug = $1 
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: UpdatePostStats :exec
UPDATE appview_posts 
SET upvotes = $2, downvotes = $3, comment_count = $4, score = $5
WHERE id = $1;
```

## Migration Management

### PDS Migrations

**Location**: `internal/database/migrations/pds/`

**Migration Files**:
- `001_initial_schema.up.sql` - Initial PDS schema
- `002_oauth_schema.up.sql` - OAuth tables
- `003_oauth_schema.up.sql` - Additional OAuth tables

**Migration Commands**:
```bash
# Apply PDS migrations
task migrate:up

# Rollback PDS migrations
task migrate:down

# Generate PDS SQLC code
task generate:sqlc
```

### AppView Migrations

**Location**: `internal/database/migrations/appview/`

**Migration Files**:
- `001_appview_schema.up.sql` - Initial AppView schema
- `002_rbac_schema.up.sql` - RBAC tables
- `003_appview_sessions.up.sql` - Session tables
- `004_performance_indexes.up.sql` - Performance indexes

**Migration Commands**:
```bash
# Apply AppView migrations
task migrate:up

# Rollback AppView migrations
task migrate:down

# Generate AppView SQLC code
task generate:sqlc
```

## Data Consistency

### Consistency Guarantees

**PDS Database**:
- **Strong Consistency**: All reads return the most recent write
- **ACID Properties**: Full ACID compliance for data integrity
- **Protocol Compliance**: Ensures atproto protocol compliance

**AppView Database**:
- **Eventual Consistency**: Eventually consistent with PDS
- **Performance**: Optimized for user-facing queries
- **Denormalization**: Pre-computed statistics for performance

### Consistency Monitoring

**Sync Status**:
```go
type SyncStatus struct {
    PDSRecords     int64
    AppViewRecords int64
    SyncLag        time.Duration
    LastSync       time.Time
}

func (ec *EventConsumer) GetSyncStatus() SyncStatus {
    return SyncStatus{
        PDSRecords:     ec.getPDSCount(),
        AppViewRecords: ec.getAppViewCount(),
        SyncLag:        ec.getSyncLag(),
        LastSync:       ec.getLastSyncTime(),
    }
}
```

**Data Reconciliation**:
```go
// Reconcile data between PDS and AppView
func (ec *EventConsumer) ReconcileData() error {
    // Get all posts from PDS
    pdsPosts, err := ec.pdsClient.ListPosts()
    if err != nil {
        return fmt.Errorf("failed to get PDS posts: %w", err)
    }

    // Get all posts from AppView
    appViewPosts, err := ec.db.ListPosts()
    if err != nil {
        return fmt.Errorf("failed to get AppView posts: %w", err)
    }

    // Compare and sync missing data
    for _, pdsPost := range pdsPosts {
        found := false
        for _, appViewPost := range appViewPosts {
            if appViewPost.AtprotoURI == pdsPost.AtprotoURI {
                found = true
                break
            }
        }

        if !found {
            // Create missing AppView post
            err := ec.createAppViewPostFromPDS(pdsPost)
            if err != nil {
                return fmt.Errorf("failed to create AppView post: %w", err)
            }
        }
    }

    return nil
}
```

## Performance Considerations

### Database Optimization

**PDS Database**:
- **Indexes**: Indexes for atproto protocol operations
- **Query Optimization**: Optimized for write operations
- **Connection Pooling**: Proper connection management
- **Transaction Management**: ACID compliance

**AppView Database**:
- **Indexes**: Indexes for user-facing queries
- **Denormalization**: Pre-computed statistics
- **Caching**: Application-level caching
- **Query Optimization**: Optimized for read operations

### Scaling Strategies

**Horizontal Scaling**:
- **PDS**: Scale based on atproto protocol load
- **AppView**: Scale based on user-facing application load
- **Database**: Scale databases independently

**Vertical Scaling**:
- **PDS**: Optimize for write-heavy workloads
- **AppView**: Optimize for read-heavy workloads
- **Database**: Optimize each database for its use case

## Security Considerations

### Database Security

**PDS Database**:
- **Access Control**: Restricted access to PDS service only
- **Data Encryption**: Encrypt sensitive data
- **Audit Logging**: Log all database operations
- **Backup Security**: Secure backup procedures

**AppView Database**:
- **Access Control**: Restricted access to AppView service only
- **Data Encryption**: Encrypt sensitive data
- **Audit Logging**: Log all database operations
- **Backup Security**: Secure backup procedures

### Network Security

**Database Isolation**:
- **Network Segmentation**: Isolate database networks
- **Firewall Rules**: Restrict database access
- **TLS Encryption**: Encrypt database connections
- **Authentication**: Strong authentication mechanisms

## Monitoring and Observability

### Database Metrics

**PDS Database Metrics**:
- **Connection Count**: Active database connections
- **Query Performance**: Query execution times
- **Transaction Rate**: Transactions per second
- **Error Rate**: Database error rate

**AppView Database Metrics**:
- **Connection Count**: Active database connections
- **Query Performance**: Query execution times
- **Cache Hit Rate**: Cache hit ratio
- **Error Rate**: Database error rate

### Health Checks

**PDS Health Check**:
```go
func (s *Server) HealthCheck() error {
    // Check database connectivity
    if err := s.db.Ping(); err != nil {
        return fmt.Errorf("PDS database connection lost: %w", err)
    }

    // Check NATS connection
    if !s.eventStream.IsConnected() {
        return fmt.Errorf("NATS connection lost")
    }

    return nil
}
```

**AppView Health Check**:
```go
func (s *Server) HealthCheck() error {
    // Check database connectivity
    if err := s.db.Ping(); err != nil {
        return fmt.Errorf("AppView database connection lost: %w", err)
    }

    // Check NATS connection
    if !s.eventConsumer.IsConnected() {
        return fmt.Errorf("NATS connection lost")
    }

    return nil
}
```

## Best Practices

### Development

1. **Clear Boundaries**: Maintain clear separation between PDS and AppView databases
2. **Event Design**: Design events for loose coupling
3. **Database Design**: Optimize each database for its use case
4. **Error Handling**: Implement comprehensive error handling

### Deployment

1. **Independent Deployment**: Deploy PDS and AppView independently
2. **Database Management**: Manage databases separately
3. **Monitoring**: Monitor each database independently
4. **Scaling**: Scale databases based on their specific load patterns

### Maintenance

1. **Database Maintenance**: Maintain each database independently
2. **Backup Strategy**: Implement separate backup strategies
3. **Monitoring**: Monitor database health continuously
4. **Performance**: Optimize database performance independently

## References

- [PDS Database Schema](../PDS/Database/schema.md)
- [AppView Database Schema](../AppView/Database/schema.md)
- [SQLC Usage](sqlc-usage.md)
- [Event Processing Architecture](../designs/archive/event-processing-architecture.md)
