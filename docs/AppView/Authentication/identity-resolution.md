# Identity Resolution

## Overview

The AppView implements identity resolution to convert DIDs to handles for user-facing operations. This service provides caching and fallback mechanisms for efficient identity lookups.

## Implementation

### Identity Resolver Service

**File**: `internal/appview/identity.go`  
**Service**: `IdentityResolver`

The `IdentityResolver` handles DID-to-handle resolution with caching:

```go
type IdentityResolver struct {
    logger *slog.Logger
    cache  map[string]string // DID -> Handle cache
    mutex  sync.RWMutex
}

func NewIdentityResolver(logger *slog.Logger) *IdentityResolver {
    return &IdentityResolver{
        logger: logger,
        cache:  make(map[string]string),
    }
}
```

### Handle Resolution

**Method**: `ResolveHandleFromDID(ctx context.Context, did string) (string, error)`

```go
func (ir *IdentityResolver) ResolveHandleFromDID(ctx context.Context, did string) (string, error) {
    // Check cache first
    ir.mutex.RLock()
    if handle, exists := ir.cache[did]; exists {
        ir.mutex.RUnlock()
        ir.logger.Debug("Handle resolved from cache", "did", did, "handle", handle)
        return handle, nil
    }
    ir.mutex.RUnlock()

    // Resolve from PDS (this would be implemented based on your PDS integration)
    handle, err := ir.resolveFromPDS(ctx, did)
    if err != nil {
        ir.logger.Error("Failed to resolve handle from PDS", "error", err, "did", did)
        return "", fmt.Errorf("failed to resolve handle: %w", err)
    }

    // Cache the result
    ir.mutex.Lock()
    ir.cache[did] = handle
    ir.mutex.Unlock()

    ir.logger.Info("Handle resolved and cached", "did", did, "handle", handle)
    return handle, nil
}
```

### Cache Management

**Method**: `CacheHandle(did, handle string)`

```go
func (ir *IdentityResolver) CacheHandle(did, handle string) {
    ir.mutex.Lock()
    defer ir.mutex.Unlock()
    
    ir.cache[did] = handle
    ir.logger.Debug("Handle cached", "did", did, "handle", handle)
}
```

**Method**: `GetCachedHandle(did string) (string, bool)`

```go
func (ir *IdentityResolver) GetCachedHandle(did string) (string, bool) {
    ir.mutex.RLock()
    defer ir.mutex.RUnlock()
    
    handle, exists := ir.cache[did]
    return handle, exists
}
```

## Integration with Event Processing

### Event Handler Integration

**File**: `internal/appview/events.go`  
**Method**: `handlePostCreated`

The identity resolver is used during event processing to resolve author handles:

```go
func (ec *EventConsumer) handlePostCreated(event AtprotoEvent) error {
    // Resolve author handle from DID
    authorHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
    if err != nil {
        ec.logger.Warn("Failed to resolve author handle", "error", err, "did", event.Repo)
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

### Subforum Creation Integration

**Method**: `handleSubforumCreated`

```go
func (ec *EventConsumer) handleSubforumCreated(event AtprotoEvent) error {
    // Resolve creator handle from DID
    createdByHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
    if err != nil {
        ec.logger.Warn("Failed to resolve creator handle", "error", err, "did", event.Repo)
        createdByHandle = "unknown"
    }

    // Create AppView subforum data
    subforum := &AppViewSubforum{
        ID:              uuid.New(),
        Name:            name,
        Slug:            slug,
        Description:     description,
        CreatedByDID:    event.Repo,
        CreatedByHandle: createdByHandle,
        CreatedAt:       event.Timestamp,
        UpdatedAt:       event.Timestamp,
    }

    // Store in AppView database
    if err := ec.db.CreateSubforum(subforum); err != nil {
        return fmt.Errorf("failed to create subforum in AppView: %w", err)
    }

    return nil
}
```

## Database Integration

### User Data Storage

**Table**: `appview_users`  
**Purpose**: Store denormalized user information

```sql
CREATE TABLE appview_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
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
```

### User Creation from Identity Resolution

**Method**: `CreateUser(user *AppViewUser) error`

```go
func (ec *EventConsumer) handleIdentityResolved(event AtprotoEvent) error {
    // Update user information in AppView database
    did := event.Repo
    handle, ok := event.Metadata["handle"].(string)
    if !ok {
        ec.logger.Warn("Invalid handle in identity resolved event", "metadata", event.Metadata)
        return nil
    }

    // Create or update user in AppView database
    user := &AppViewUser{
        ID:          uuid.New(),
        DID:         did,
        Handle:      handle,
        DisplayName: handle, // Use handle as display name for now
        CreatedAt:   event.Timestamp,
        UpdatedAt:   event.Timestamp,
    }

    if err := ec.db.CreateUser(user); err != nil {
        ec.logger.Error("Failed to update user from identity resolution", "error", err, "did", did)
        return fmt.Errorf("failed to update user: %w", err)
    }

    ec.logger.Info("User information updated from identity resolution", "did", did, "handle", handle)
    return nil
}
```

## Caching Strategy

### In-Memory Cache

**Implementation**: Simple map-based cache with mutex protection

```go
type IdentityResolver struct {
    logger *slog.Logger
    cache  map[string]string // DID -> Handle cache
    mutex  sync.RWMutex
}
```

### Cache Operations

**Read Operations**: Use read lock for concurrent access
**Write Operations**: Use write lock for exclusive access
**Cache Miss**: Resolve from PDS and cache result

### Cache Benefits

- **Performance**: Avoid repeated PDS lookups
- **Reliability**: Fallback for PDS unavailability
- **Efficiency**: Reduce network calls during event processing

## Error Handling

### Resolution Failures

```go
// Handle resolution failures gracefully
authorHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
if err != nil {
    ec.logger.Warn("Failed to resolve author handle", "error", err, "did", event.Repo)
    authorHandle = "unknown" // Fallback to unknown
}
```

### Cache Errors

```go
// Handle cache errors
ir.mutex.Lock()
defer ir.mutex.Unlock()

ir.cache[did] = handle
ir.logger.Debug("Handle cached", "did", did, "handle", handle)
```

### PDS Integration Errors

```go
// Handle PDS resolution errors
handle, err := ir.resolveFromPDS(ctx, did)
if err != nil {
    ir.logger.Error("Failed to resolve handle from PDS", "error", err, "did", did)
    return "", fmt.Errorf("failed to resolve handle: %w", err)
}
```

## Performance Considerations

### Cache Hit Rate
- **Target**: High cache hit rate for frequently accessed DIDs
- **Monitoring**: Log cache hits vs misses
- **Optimization**: Pre-populate cache with common users

### Memory Usage
- **Growth**: Cache grows with unique DIDs
- **Cleanup**: Consider TTL for cache entries
- **Monitoring**: Track memory usage of cache

### Network Calls
- **Reduction**: Cache reduces PDS API calls
- **Fallback**: Graceful degradation on PDS unavailability
- **Batching**: Consider batch resolution for multiple DIDs

## Configuration

### Environment Variables

```bash
# AppView configuration
ENVIRONMENT=development
PDS_URL=http://hashpost-pds:8080
NATS_URL=nats://nats:4222
```

### Cache Configuration

```go
// Cache configuration
type IdentityResolver struct {
    logger    *slog.Logger
    cache     map[string]string
    mutex     sync.RWMutex
    maxSize   int           // Optional: max cache size
    ttl       time.Duration // Optional: cache TTL
}
```

## Monitoring and Debugging

### Logging

```go
// Cache hit logging
ir.logger.Debug("Handle resolved from cache", "did", did, "handle", handle)

// Cache miss logging
ir.logger.Info("Handle resolved and cached", "did", did, "handle", handle)

// Error logging
ir.logger.Error("Failed to resolve handle from PDS", "error", err, "did", did)
```

### Metrics

```go
// Cache metrics
type CacheMetrics struct {
    Hits   int64
    Misses int64
    Size   int
}

func (ir *IdentityResolver) GetMetrics() CacheMetrics {
    ir.mutex.RLock()
    defer ir.mutex.RUnlock()
    
    return CacheMetrics{
        Hits:   ir.hits,
        Misses: ir.misses,
        Size:   len(ir.cache),
    }
}
```

## Future Enhancements

### Persistent Cache
- **Redis Integration**: Use Redis for distributed caching
- **TTL Support**: Automatic cache expiration
- **Persistence**: Survive application restarts

### Batch Resolution
- **Multiple DIDs**: Resolve multiple DIDs in single call
- **Efficiency**: Reduce network overhead
- **Batching**: Group resolution requests

### Advanced Caching
- **LRU Eviction**: Least recently used eviction policy
- **Size Limits**: Maximum cache size
- **Metrics**: Detailed cache performance metrics

## References

- [Identity Resolver Implementation](internal/appview/identity.go)
- [Event Processing Integration](internal/appview/events.go)
- [Database User Operations](internal/database/queries/appview/)
- [AppView Database Schema](internal/database/migrations/appview/)
