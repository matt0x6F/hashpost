# Data Flow

## Overview

HashPost implements an event-driven data flow architecture where the PDS maintains canonical atproto records and the AppView maintains denormalized data synchronized through NATS JetStream events.

## End-to-End Data Flow

### User Registration Flow

```
1. User → PDS: POST /xrpc/com.atproto.server.createAccount
2. PDS → Database: Store user record
3. PDS → NATS: Publish identity.resolved event
4. NATS → AppView: Consume identity.resolved event
5. AppView → Database: Store denormalized user data
```

### Post Creation Flow

```
1. User → PDS: POST /xrpc/com.atproto.repo.createRecord
2. PDS → Database: Store canonical post record
3. PDS → NATS: Publish record.created event
4. NATS → AppView: Consume record.created event
5. AppView → Database: Store denormalized post data
6. AppView → User: Return success response
```

### Post Retrieval Flow

```
1. User → AppView: GET /api/v1/posts/{id}
2. AppView → Database: Query denormalized post data
3. AppView → User: Return post with statistics
```

## Event-Driven Architecture

### Event Publishing (PDS)

**File**: `internal/pds/events.go`  
**Service**: `EventStreamer`

```go
func (es *EventStreamer) PublishRecordEvent(ctx context.Context, eventType EventType, repo, collection string, record map[string]interface{}, uri, cid string) error {
    event := &AtprotoEvent{
        Type:       eventType,
        Repo:       repo,
        Collection: collection,
        Record:     record,
        URI:        uri,
        CID:        cid,
        Timestamp:  time.Now(),
        Metadata: map[string]interface{}{
            "source": "hashpost-pds",
        },
    }

    return es.publishEvent(ctx, event)
}
```

### Event Consumption (AppView)

**File**: `internal/appview/events.go`  
**Service**: `EventConsumer`

```go
func (ec *EventConsumer) StartConsuming(ctx context.Context) error {
    // Subscribe to all hashpost events
    subject := "hashpost.events.record.created"
    streamName := "HASHPOST_EVENTS"

    // Create pull subscription with explicit stream binding
    sub, err := ec.js.PullSubscribe(subject, "hashpost-appview",
        nats.BindStream(streamName),
    )
    if err != nil {
        return fmt.Errorf("failed to create subscription: %w", err)
    }

    // Process messages in a loop
    for {
        select {
        case <-ctx.Done():
            return nil
        default:
            // Fetch and process messages
            msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
            if err != nil {
                continue
            }

            for _, msg := range msgs {
                if err := ec.processMessageWithRetry(msg); err != nil {
                    ec.sendToDeadLetterQueue(msg, err)
                    continue
                }

                if err := msg.Ack(); err != nil {
                    ec.logger.Error("Failed to ack message", "error", err)
                }
            }
        }
    }
}
```

## Data Synchronization

### PDS to AppView Synchronization

**Event Types**:
- **Record Events**: `record.created`, `record.updated`, `record.deleted`
- **Identity Events**: `identity.resolved`, `session.created`

**Synchronization Process**:
1. **PDS**: User performs action (create/update/delete)
2. **PDS**: Stores canonical record in PDS database
3. **PDS**: Publishes event to NATS JetStream
4. **AppView**: Consumes event from NATS JetStream
5. **AppView**: Updates denormalized data in AppView database

### Data Consistency

**Eventual Consistency**:
- **PDS**: Source of truth for canonical records
- **AppView**: Eventually consistent denormalized data
- **Sync Delay**: Minimal delay due to event processing

**Consistency Guarantees**:
- **At-Least-Once Delivery**: NATS JetStream ensures event delivery
- **Idempotency**: AppView handles duplicate events gracefully
- **Error Handling**: Failed events go to dead letter queue

## Database Operations

### PDS Database Operations

**Canonical Record Storage**:
```sql
-- Store canonical atproto record
INSERT INTO posts (user_id, subforum_id, title, content, atproto_uri)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
```

**Event Publishing**:
```go
// Publish event after database operation
err := s.eventStream.PublishRecordEvent(
    r.Context(),
    EventTypeRecordCreated,
    session.DID,
    req.Collection,
    req.Record,
    uri,
    cid,
)
```

### AppView Database Operations

**Denormalized Data Storage**:
```sql
-- Store denormalized post data
INSERT INTO appview_posts (atproto_uri, author_did, author_handle, subforum_slug, title, content)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;
```

**Statistics Updates**:
```sql
-- Update post statistics
UPDATE appview_posts 
SET upvotes = $2, downvotes = $3, comment_count = $4, score = $5
WHERE id = $1;
```

## Event Processing Patterns

### Record Created Processing

**PDS Side**:
```go
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

**AppView Side**:
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

### Identity Resolution Processing

**PDS Side**:
```go
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
    // ... authentication ...

    // Publish identity resolved event
    err = s.eventStream.PublishIdentityEvent(
        r.Context(),
        EventTypeIdentityResolved,
        session.Handle,
        session.DID,
    )
}
```

**AppView Side**:
```go
func (ec *EventConsumer) handleIdentityResolved(event AtprotoEvent) error {
    // Update user information in AppView database
    did := event.Repo
    handle, ok := event.Metadata["handle"].(string)
    if !ok {
        return fmt.Errorf("invalid handle in identity resolved event")
    }

    // Create or update user in AppView database
    user := &AppViewUser{
        ID:          uuid.New(),
        DID:         did,
        Handle:      handle,
        DisplayName: handle,
        CreatedAt:   event.Timestamp,
        UpdatedAt:   event.Timestamp,
    }

    if err := ec.db.CreateUser(user); err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }

    return nil
}
```

## Error Handling and Recovery

### Event Processing Errors

**Retry Logic**:
```go
func (ec *EventConsumer) processMessageWithRetry(msg *nats.Msg) error {
    // Check for idempotency
    eventID := ec.generateEventID(msg)
    if ec.isEventProcessed(eventID) {
        return nil
    }

    // Retry logic with exponential backoff
    for attempt := 0; attempt < ec.maxRetries; attempt++ {
        if err := ec.processMessage(msg); err != nil {
            if attempt == ec.maxRetries-1 {
                return fmt.Errorf("failed after %d attempts: %w", ec.maxRetries, err)
            }

            delay := ec.calculateBackoffDelay(attempt)
            time.Sleep(delay)
            continue
        }

        ec.markEventProcessed(eventID)
        return nil
    }

    return fmt.Errorf("exhausted all retry attempts")
}
```

**Dead Letter Queue**:
```go
func (ec *EventConsumer) sendToDeadLetterQueue(msg *nats.Msg, err error) {
    dlqSubject := "hashpost.events.dlq"

    dlqData := map[string]interface{}{
        "original_subject": msg.Subject,
        "original_data":    string(msg.Data),
        "error":            err.Error(),
        "timestamp":        time.Now(),
    }

    dlqBytes, _ := json.Marshal(dlqData)
    ec.js.Publish(dlqSubject, dlqBytes)
}
```

### Data Consistency Recovery

**Manual Reconciliation**:
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

### Event Processing Performance

**Batch Processing**:
```go
// Process multiple events in batch
func (ec *EventConsumer) processBatch(msgs []*nats.Msg) error {
    for _, msg := range msgs {
        if err := ec.processMessageWithRetry(msg); err != nil {
            ec.sendToDeadLetterQueue(msg, err)
            continue
        }

        if err := msg.Ack(); err != nil {
            ec.logger.Error("Failed to ack message", "error", err)
        }
    }
    return nil
}
```

**Database Optimization**:
```sql
-- Use indexes for fast lookups
CREATE INDEX idx_appview_posts_atproto_uri ON appview_posts(atproto_uri);
CREATE INDEX idx_appview_posts_author_did ON appview_posts(author_did);
CREATE INDEX idx_appview_posts_subforum_slug ON appview_posts(subforum_slug);
```

### Caching Strategy

**Identity Resolution Caching**:
```go
func (ir *IdentityResolver) ResolveHandleFromDID(ctx context.Context, did string) (string, error) {
    // Check cache first
    ir.mutex.RLock()
    if handle, exists := ir.cache[did]; exists {
        ir.mutex.RUnlock()
        return handle, nil
    }
    ir.mutex.RUnlock()

    // Resolve from PDS
    handle, err := ir.resolveFromPDS(ctx, did)
    if err != nil {
        return "", fmt.Errorf("failed to resolve handle: %w", err)
    }

    // Cache the result
    ir.mutex.Lock()
    ir.cache[did] = handle
    ir.mutex.Unlock()

    return handle, nil
}
```

## Monitoring and Observability

### Event Flow Metrics

**PDS Metrics**:
- **Events Published**: Number of events published per second
- **Event Types**: Breakdown by event type
- **Publishing Errors**: Failed event publishing attempts

**AppView Metrics**:
- **Events Consumed**: Number of events consumed per second
- **Processing Time**: Average event processing time
- **Error Rate**: Failed event processing rate
- **DLQ Messages**: Messages in dead letter queue

### Data Consistency Monitoring

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

## Best Practices

### Event Design

1. **Immutable Events**: Events should be immutable once published
2. **Event Versioning**: Version events for backward compatibility
3. **Event Schema**: Define clear event schemas
4. **Event Ordering**: Handle event ordering correctly

### Data Consistency

1. **Idempotency**: Make event processing idempotent
2. **Error Handling**: Implement comprehensive error handling
3. **Monitoring**: Monitor data consistency continuously
4. **Recovery**: Implement data reconciliation procedures

### Performance

1. **Batch Processing**: Process events in batches when possible
2. **Caching**: Cache frequently accessed data
3. **Database Optimization**: Optimize database queries
4. **Monitoring**: Monitor performance metrics

## References

- [Event Processing Architecture](../designs/archive/event-processing-architecture.md)
- [PDS Event Publishing](../PDS/Events/event-publishing.md)
- [AppView Event Consumption](../AppView/Events/event-consumption.md)
- [NATS JetStream Documentation](https://docs.nats.io/jetstream)
