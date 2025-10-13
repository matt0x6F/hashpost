# Event Streaming

## Overview

HashPost implements event-driven communication between PDS and AppView using NATS JetStream. This enables loose coupling, scalability, and reliability in the distributed system.

## Event Architecture

### Event Flow

```
PDS → NATS JetStream → AppView
```

**Flow**:
1. **PDS**: User performs action (create/update/delete atproto record)
2. **PDS**: Publishes event to NATS JetStream
3. **NATS**: Stores event in stream with delivery guarantees
4. **AppView**: Consumes event from NATS JetStream
5. **AppView**: Updates denormalized data based on event

### Event Types

**Record Events**:
- `record.created` - New atproto record created
- `record.updated` - Existing record updated
- `record.deleted` - Record deleted

**Identity Events**:
- `identity.resolved` - DID resolved to handle
- `session.created` - User session created

**Event Structure**:
```json
{
  "type": "record.created",
  "repo": "did:plc:hashpost-binding-test",
  "collection": "app.hashpost.feed.post",
  "record": {
    "$type": "app.hashpost.feed.post",
    "text": "Hello, HashPost!",
    "createdAt": "2025-01-01T00:00:00.000Z"
  },
  "uri": "at://did:plc:hashpost-binding-test/app.hashpost.feed.post/123",
  "cid": "bafybeigdyrzt5sfpudm7hu76uh7y26nf3efuylqabf3oclgtqy55fbzdi-123",
  "timestamp": "2025-01-01T00:00:00.000Z",
  "metadata": {
    "source": "hashpost-pds"
  }
}
```

## Event Publishing (PDS)

### Event Streamer Service

**File**: `internal/pds/events.go`  
**Service**: `EventStreamer`

```go
type EventStreamer struct {
    nc      NatsConn
    js      JetStreamContext
    logger  *slog.Logger
    stream  string
    subject string
}

func NewEventStreamer(natsURL string, logger *slog.Logger) (*EventStreamer, error) {
    // Connect to NATS
    nc, err := nats.Connect(natsURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }

    // Create JetStream context
    js, err := nc.JetStream()
    if err != nil {
        return nil, fmt.Errorf("failed to create JetStream context: %w", err)
    }

    streamer := &EventStreamer{
        nc:      nc,
        js:      js,
        logger:  logger,
        stream:  "HASHPOST_EVENTS",
        subject: "hashpost.events.record.created",
    }

    // Create the stream
    if err := streamer.createStream(); err != nil {
        return nil, fmt.Errorf("failed to create stream: %w", err)
    }

    return streamer, nil
}
```

### Event Publishing Methods

**Record Event Publishing**:
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

**Identity Event Publishing**:
```go
func (es *EventStreamer) PublishIdentityEvent(ctx context.Context, eventType EventType, handle, did string) error {
    event := &AtprotoEvent{
        Type:      eventType,
        Repo:      did,
        Timestamp: time.Now(),
        Metadata: map[string]interface{}{
            "source": "hashpost-pds",
            "handle": handle,
            "did":    did,
        },
    }

    return es.publishEvent(ctx, event)
}
```

**Core Publishing Method**:
```go
func (es *EventStreamer) publishEvent(ctx context.Context, event *AtprotoEvent) error {
    // Serialize event to JSON
    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
    }

    // Create subject based on event type
    subject := "hashpost.events.record.created"

    // Check if stream exists before publishing
    streamInfo, err := es.js.StreamInfo(es.stream)
    if err != nil {
        return fmt.Errorf("stream does not exist: %w", err)
    }

    // Publish to NATS JetStream
    ack, err := es.js.Publish(subject, data)
    if err != nil {
        return fmt.Errorf("failed to publish event: %w", err)
    }

    es.logger.Debug("Event published successfully",
        "stream", ack.Stream,
        "sequence", ack.Sequence,
        "subject", subject,
    )

    return nil
}
```

## Event Consumption (AppView)

### Event Consumer Service

**File**: `internal/appview/events.go`  
**Service**: `EventConsumer`

```go
type EventConsumer struct {
    nc               *nats.Conn
    js               nats.JetStreamContext
    logger           *slog.Logger
    db               *Database
    identityResolver   *IdentityResolver

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

**Start Consuming**:
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

**Message Processing**:
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

## Error Handling

### Retry Logic

**Message Processing with Retry**:
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

**DLQ Implementation**:
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

```go
func (ec *EventConsumer) isEventProcessed(eventID string) bool {
    return ec.processedEvents[eventID]
}

func (ec *EventConsumer) markEventProcessed(eventID string) {
    ec.processedEvents[eventID] = true
}
```

## Performance Considerations

### Message Processing

**Batch Processing**:
```go
// Process multiple messages in batch
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

**Concurrent Processing**:
```go
// Process messages concurrently
func (ec *EventConsumer) processConcurrently(msgs []*nats.Msg) error {
    var wg sync.WaitGroup
    errChan := make(chan error, len(msgs))

    for _, msg := range msgs {
        wg.Add(1)
        go func(m *nats.Msg) {
            defer wg.Done()
            if err := ec.processMessageWithRetry(m); err != nil {
                errChan <- err
            }
        }(msg)
    }

    wg.Wait()
    close(errChan)

    // Check for errors
    for err := range errChan {
        if err != nil {
            return err
        }
    }

    return nil
}
```

## Monitoring and Observability

### Event Metrics

```go
type EventMetrics struct {
    ProcessedEvents    int64
    FailedEvents      int64
    RetryAttempts     int64
    DLQMessages       int64
    TransientErrors   int64
    PermanentErrors   int64
}

func (ec *EventConsumer) GetEventMetrics() EventMetrics {
    return EventMetrics{
        ProcessedEvents:  ec.processedEvents,
        FailedEvents:     ec.failedEvents,
        RetryAttempts:    ec.retryAttempts,
        DLQMessages:      ec.dlqMessages,
        TransientErrors:  ec.transientErrors,
        PermanentErrors:  ec.permanentErrors,
    }
}
```

### Health Checks

```go
// Health check for event consumer
func (ec *EventConsumer) HealthCheck() error {
    // Check NATS connection
    if !ec.nc.IsConnected() {
        return fmt.Errorf("NATS connection lost")
    }

    // Check database connection
    if err := ec.db.Ping(); err != nil {
        return fmt.Errorf("database connection lost: %w", err)
    }

    // Check DLQ queue size
    if ec.dlqMessages > 100 {
        return fmt.Errorf("too many messages in DLQ: %d", ec.dlqMessages)
    }

    return nil
}
```

## Configuration

### Environment Variables

```bash
# NATS configuration
NATS_URL=nats://nats:4222
NATS_STREAM=HASHPOST_EVENTS
NATS_SUBJECT=hashpost.events.record.created

# Error handling configuration
EVENT_MAX_RETRIES=3
EVENT_RETRY_DELAY=1s
EVENT_MAX_RETRY_DELAY=30s
EVENT_DLQ_SUBJECT=hashpost.events.dlq
```

### Docker Compose

```yaml
services:
  nats:
    image: nats:2.10-alpine
    ports:
      - "4222:4222"
    command: ["-js"]  # Enable JetStream
    volumes:
      - nats_data:/data

  hashpost-pds:
    build: .
    environment:
      - NATS_URL=nats://nats:4222
    depends_on:
      - nats

  hashpost-appview:
    build: .
    environment:
      - NATS_URL=nats://nats:4222
    depends_on:
      - nats
```

## Best Practices

### Event Design

1. **Immutable Events**: Events should be immutable once published
2. **Event Versioning**: Version events for backward compatibility
3. **Event Schema**: Define clear event schemas
4. **Event Ordering**: Handle event ordering correctly

### Error Handling

1. **Retry Logic**: Implement retry logic for transient errors
2. **Dead Letter Queue**: Handle permanent failures gracefully
3. **Idempotency**: Make event processing idempotent
4. **Monitoring**: Monitor event processing continuously

### Performance

1. **Batch Processing**: Process events in batches when possible
2. **Concurrent Processing**: Use concurrent processing for high throughput
3. **Connection Pooling**: Optimize database connections
4. **Monitoring**: Monitor performance metrics

## References

- [NATS JetStream Documentation](https://docs.nats.io/jetstream)
- [PDS Event Publishing](../PDS/Events/event-publishing.md)
- [AppView Event Consumption](../AppView/Events/event-consumption.md)
- [Event Processing Architecture](../designs/archive/event-processing-architecture.md)
