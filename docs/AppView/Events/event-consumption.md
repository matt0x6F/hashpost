# Event Consumption

## Overview

The AppView consumes atproto events from the PDS via NATS JetStream (limited current usage). This enables partial event-driven architecture where the AppView maintains denormalized data synchronized with the PDS canonical records.

## Implementation

### Event Consumer Service

**File**: `internal/appview/events.go`  
**Service**: `EventConsumer`

The `EventConsumer` handles consuming atproto events from NATS JetStream:

```go
type EventConsumer struct {
    nc               *nats.Conn
    js               nats.JetStreamContext
    logger           *slog.Logger
    db               *Database
    identityResolver *IdentityResolver

    // Enhanced error handling
    maxRetries      int
    retryDelay      time.Duration
    maxRetryDelay   time.Duration
    processedEvents map[string]bool // For idempotency
}

func NewEventConsumer(natsURL string, db *Database, logger *slog.Logger) (*EventConsumer, error) {
    // Connect to NATS
    nc, err := nats.Connect(natsURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }

    // Create JetStream context
    js, err := nc.JetStream()
    if err != nil {
        nc.Close()
        return nil, fmt.Errorf("failed to create JetStream context: %w", err)
    }

    // Create identity resolver
    identityResolver := NewIdentityResolver(logger)

    consumer := &EventConsumer{
        nc:               nc,
        js:               js,
        logger:           logger,
        db:               db,
        identityResolver: identityResolver,

        // Enhanced error handling configuration
        maxRetries:      3,
        retryDelay:      1 * time.Second,
        maxRetryDelay:   30 * time.Second,
        processedEvents: make(map[string]bool),
    }

    return consumer, nil
}
```

### Event Consumption

**Method**: `StartConsuming(ctx context.Context) error`

```go
func (ec *EventConsumer) StartConsuming(ctx context.Context) error {
    // Subscribe to all hashpost events
    subject := "hashpost.events.record.created"
    streamName := "HASHPOST_EVENTS"

    ec.logger.Info("Subscribing to atproto events", "subject", subject, "stream", streamName)

    // Create pull subscription with explicit stream binding
    sub, err := ec.js.PullSubscribe(subject, "hashpost-appview",
        nats.BindStream(streamName),
    )
    if err != nil {
        return fmt.Errorf("failed to create subscription: %w", err)
    }

    ec.logger.Info("Started consuming atproto events", "subject", subject)

    // Process messages in a loop
    for {
        select {
        case <-ctx.Done():
            ec.logger.Info("Stopping event consumer")
            sub.Unsubscribe()
            return nil
        default:
            // Fetch messages
            msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
            if err != nil {
                if err == nats.ErrTimeout {
                    continue
                }
                ec.logger.Error("Failed to fetch messages", "error", err)
                continue
            }

            // Process each message with enhanced error handling
            for _, msg := range msgs {
                if err := ec.processMessageWithRetry(msg); err != nil {
                    ec.logger.Error("Failed to process message after retries", "error", err)
                    ec.sendToDeadLetterQueue(msg, err)
                    continue
                }

                // Acknowledge the message
                if err := msg.Ack(); err != nil {
                    ec.logger.Error("Failed to ack message", "error", err)
                }
            }
        }
    }
}
```

## Event Processing

### Event Structure

**Type**: `AtprotoEvent`

```go
type AtprotoEvent struct {
    Type       string                 `json:"type"`
    Repo       string                 `json:"repo"`
    Collection string                 `json:"collection,omitempty"`
    Record     map[string]interface{} `json:"record,omitempty"`
    URI        string                 `json:"uri,omitempty"`
    CID        string                 `json:"cid,omitempty"`
    Timestamp  time.Time              `json:"timestamp"`
    Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
```

### Event Routing

**Method**: `processMessage(msg *nats.Msg) error`

```go
func (ec *EventConsumer) processMessage(msg *nats.Msg) error {
    ec.logger.Info("Processing NATS message", "subject", msg.Subject, "data_length", len(msg.Data))

    // Parse the event
    var event AtprotoEvent
    if err := json.Unmarshal(msg.Data, &event); err != nil {
        return fmt.Errorf("failed to unmarshal event: %w", err)
    }

    ec.logger.Info("Processing atproto event",
        "type", event.Type,
        "repo", event.Repo,
        "collection", event.Collection,
        "subject", msg.Subject,
    )

    // Route event to appropriate handler
    switch event.Type {
    case "record.created":
        return ec.handleRecordCreated(event)
    case "record.updated":
        return ec.handleRecordUpdated(event)
    case "record.deleted":
        return ec.handleRecordDeleted(event)
    case "identity.resolved":
        return ec.handleIdentityResolved(event)
    case "session.created":
        return ec.handleSessionCreated(event)
    default:
        ec.logger.Warn("Unknown event type", "type", event.Type)
        return nil
    }
}
```

## Event Handlers

### Record Created Handler

**Method**: `handleRecordCreated(event AtprotoEvent) error`

```go
func (ec *EventConsumer) handleRecordCreated(event AtprotoEvent) error {
    ec.logger.Info("Record created",
        "repo", event.Repo,
        "collection", event.Collection,
        "uri", event.URI,
    )

    // Process based on collection type
    switch event.Collection {
    case lexicons.CollectionFeedPost:
        return ec.handlePostCreated(event)
    case lexicons.CollectionFeedSubforum:
        return ec.handleSubforumCreated(event)
    default:
        ec.logger.Debug("Unknown collection type", "collection", event.Collection)
    }

    return nil
}
```

### Post Created Handler

**Method**: `handlePostCreated(event AtprotoEvent) error`

```go
func (ec *EventConsumer) handlePostCreated(event AtprotoEvent) error {
    ec.logger.Info("HashPost feed post created",
        "repo", event.Repo,
        "uri", event.URI,
        "text", event.Record["text"],
    )

    // Extract post data from the record
    text, ok := event.Record["text"].(string)
    if !ok {
        return fmt.Errorf("invalid text field in post record")
    }

    createdAtStr, ok := event.Record["createdAt"].(string)
    if !ok {
        return fmt.Errorf("invalid createdAt field in post record")
    }

    // Parse timestamp
    createdAt, err := time.Parse(time.RFC3339, createdAtStr)
    if err != nil {
        createdAt = time.Now()
        ec.logger.Warn("Failed to parse createdAt, using current time", "createdAt", createdAtStr, "error", err)
    }

    // Resolve author handle from DID
    authorHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
    if err != nil {
        ec.logger.Warn("Failed to resolve author handle", "error", err, "did", event.Repo)
        authorHandle = "unknown"
    }

    // Extract subforum slug from record context or use default
    subforumSlug := "general"
    if subforum, ok := event.Record["subforum"].(string); ok && subforum != "" {
        subforumSlug = subforum
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

    ec.logger.Info("Post stored in AppView database",
        "uri", event.URI,
        "text", text,
        "created_at", createdAt,
    )

    return nil
}
```

### Subforum Created Handler

**Method**: `handleSubforumCreated(event AtprotoEvent) error`

```go
func (ec *EventConsumer) handleSubforumCreated(event AtprotoEvent) error {
    ec.logger.Info("HashPost subforum created",
        "repo", event.Repo,
        "uri", event.URI,
        "name", event.Record["name"],
    )

    // Extract subforum data from the record
    name, ok := event.Record["name"].(string)
    if !ok {
        return fmt.Errorf("invalid name field in subforum record")
    }

    slug, ok := event.Record["slug"].(string)
    if !ok {
        return fmt.Errorf("invalid slug field in subforum record")
    }

    description, ok := event.Record["description"].(string)
    if !ok {
        return fmt.Errorf("invalid description field in subforum record")
    }

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

    ec.logger.Info("Subforum stored in AppView database",
        "name", name,
        "slug", slug,
        "description", description,
    )

    return nil
}
```

### Identity Resolved Handler

**Method**: `handleIdentityResolved(event AtprotoEvent) error`

```go
func (ec *EventConsumer) handleIdentityResolved(event AtprotoEvent) error {
    ec.logger.Info("Identity resolved",
        "repo", event.Repo,
        "handle", event.Metadata["handle"],
    )

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

## Error Handling

### Retry Logic

**Method**: `processMessageWithRetry(msg *nats.Msg) error`

```go
func (ec *EventConsumer) processMessageWithRetry(msg *nats.Msg) error {
    // Check for idempotency
    eventID := ec.generateEventID(msg)
    if ec.isEventProcessed(eventID) {
        ec.logger.Debug("Event already processed", "event_id", eventID)
        return nil
    }

    // Retry logic with exponential backoff
    for attempt := 0; attempt < ec.maxRetries; attempt++ {
        if err := ec.processMessage(msg); err != nil {
            if attempt == ec.maxRetries-1 {
                return fmt.Errorf("failed after %d attempts: %w", ec.maxRetries, err)
            }

            // Calculate exponential backoff delay
            delay := ec.calculateBackoffDelay(attempt)
            ec.logger.Warn("Message processing failed, retrying",
                "attempt", attempt+1,
                "max_retries", ec.maxRetries,
                "delay", delay,
                "error", err,
            )
            time.Sleep(delay)
            continue
        }

        // Success - mark as processed
        ec.markEventProcessed(eventID)
        return nil
    }

    return fmt.Errorf("exhausted all retry attempts")
}
```

### Dead Letter Queue

**Method**: `sendToDeadLetterQueue(msg *nats.Msg, err error)`

```go
func (ec *EventConsumer) sendToDeadLetterQueue(msg *nats.Msg, err error) {
    dlqSubject := "hashpost.events.dlq"

    // Create DLQ message with original data and error info
    dlqData := map[string]interface{}{
        "original_subject": msg.Subject,
        "original_data":    string(msg.Data),
        "error":            err.Error(),
        "timestamp":        time.Now(),
    }

    dlqBytes, marshalErr := json.Marshal(dlqData)
    if marshalErr != nil {
        ec.logger.Error("Failed to marshal DLQ data", "error", marshalErr)
        return
    }

    // Publish to dead letter queue
    if _, pubErr := ec.js.Publish(dlqSubject, dlqBytes); pubErr != nil {
        ec.logger.Error("Failed to publish to dead letter queue", "error", pubErr)
        return
    }

    ec.logger.Error("Message sent to dead letter queue",
        "dlq_subject", dlqSubject,
        "original_subject", msg.Subject,
        "error", err,
    )
}
```

## Idempotency

### Event ID Generation

**Method**: `generateEventID(msg *nats.Msg) string`

```go
func (ec *EventConsumer) generateEventID(msg *nats.Msg) string {
    // Use subject + sequence + timestamp for uniqueness
    sequence := msg.Header.Get("Nats-Sequence")
    if sequence == "" {
        sequence = "unknown"
    }
    return fmt.Sprintf("%s-%s-%d", msg.Subject, sequence, time.Now().Unix())
}
```

### Event Tracking

**Methods**: `isEventProcessed`, `markEventProcessed`

```go
func (ec *EventConsumer) isEventProcessed(eventID string) bool {
    return ec.processedEvents[eventID]
}

func (ec *EventConsumer) markEventProcessed(eventID string) {
    ec.processedEvents[eventID] = true
}
```

## Configuration

### Environment Variables

```bash
# AppView configuration
ENVIRONMENT=development
NATS_URL=nats://nats:4222
PDS_URL=http://hashpost-pds:8080
```

### Consumer Configuration

```go
consumer := &EventConsumer{
    nc:               nc,
    js:               js,
    logger:           logger,
    db:               db,
    identityResolver: identityResolver,

    // Enhanced error handling configuration
    maxRetries:      3,
    retryDelay:      1 * time.Second,
    maxRetryDelay:   30 * time.Second,
    processedEvents: make(map[string]bool),
}
```

## Performance Considerations

### Message Processing
- **Batch Processing**: Process multiple messages in single batch
- **Concurrent Processing**: Consider worker pools for high throughput
- **Memory Management**: Monitor processed events map size

### Database Operations
- **Transaction Batching**: Batch database operations for efficiency
- **Connection Pooling**: Proper database connection management
- **Query Optimization**: Use efficient queries for data storage

### Error Handling
- **Retry Limits**: Configurable retry attempts
- **Backoff Strategy**: Exponential backoff for retries
- **Dead Letter Queue**: Handle permanently failed messages

## Monitoring and Debugging

### Logging

```go
// Event processing logging
ec.logger.Info("Processing atproto event",
    "type", event.Type,
    "repo", event.Repo,
    "collection", event.Collection,
    "subject", msg.Subject,
)

// Error logging
ec.logger.Error("Failed to process message after retries", "error", err)
```

### Metrics

```go
// Event processing metrics
type EventMetrics struct {
    ProcessedEvents int64
    FailedEvents    int64
    RetryAttempts   int64
    DLQMessages     int64
}
```

## References

- [Event Consumer Implementation](internal/appview/events.go)
- [Database Operations](internal/appview/database.go)
- [Identity Resolution](internal/appview/identity.go)
- [NATS JetStream Documentation](https://docs.nats.io/jetstream)
