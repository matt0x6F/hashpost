# Event Publishing

## Overview

The PDS publishes atproto events to NATS JetStream for consumption by the AppView. This enables the event-driven architecture where the PDS stores canonical records and the AppView maintains denormalized data.

## Implementation

### Event Streamer Service

**File**: `internal/pds/events.go`  
**Service**: `EventStreamer`

The `EventStreamer` handles publishing atproto events to NATS JetStream:

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
    js, err := wrappedConn.JetStream()
    if err != nil {
        return nil, fmt.Errorf("failed to create JetStream context: %w", err)
    }

    streamer := &EventStreamer{
        nc:      wrappedConn,
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

### Event Types

**File**: `internal/pds/events.go`  
**Constants**: Event types for different atproto events

```go
type EventType string

const (
    EventTypeRecordCreated    EventType = "record.created"
    EventTypeRecordUpdated    EventType = "record.updated"
    EventTypeRecordDeleted    EventType = "record.deleted"
    EventTypeIdentityResolved EventType = "identity.resolved"
    EventTypeSessionCreated   EventType = "session.created"
)
```

### Event Structure

**Type**: `AtprotoEvent`

```go
type AtprotoEvent struct {
    Type       EventType              `json:"type"`
    Repo       string                 `json:"repo"`
    Collection string                 `json:"collection,omitempty"`
    Record     map[string]interface{} `json:"record,omitempty"`
    URI        string                 `json:"uri,omitempty"`
    CID        string                 `json:"cid,omitempty"`
    Timestamp  time.Time              `json:"timestamp"`
    Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
```

## Event Publishing Methods

### Record Events

**Method**: `PublishRecordEvent(ctx context.Context, eventType EventType, repo, collection string, record map[string]interface{}, uri, cid string) error`

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

### Identity Events

**Method**: `PublishIdentityEvent(ctx context.Context, eventType EventType, handle, did string) error`

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

### Session Events

**Method**: `PublishSessionEvent(ctx context.Context, eventType EventType, did, sessionID string) error`

```go
func (es *EventStreamer) PublishSessionEvent(ctx context.Context, eventType EventType, did, sessionID string) error {
    event := &AtprotoEvent{
        Type:      eventType,
        Repo:      did,
        Timestamp: time.Now(),
        Metadata: map[string]interface{}{
            "source":     "hashpost-pds",
            "session_id": sessionID,
        },
    }

    return es.publishEvent(ctx, event)
}
```

## Event Publishing Implementation

### Core Publishing Method

**Method**: `publishEvent(ctx context.Context, event *AtprotoEvent) error`

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

## Stream Configuration

### Stream Creation

**Method**: `createStream() error`

```go
func (es *EventStreamer) createStream() error {
    // Check if stream already exists
    streamInfo, err := es.js.StreamInfo(es.stream)
    if err == nil {
        es.logger.Info("NATS stream already exists",
            "stream", es.stream,
            "subjects", streamInfo.Config.Subjects,
            "state", streamInfo.State,
        )
        return nil
    }

    // Create stream configuration
    streamConfig := &nats.StreamConfig{
        Name:      es.stream,
        Subjects:  []string{es.subject},
        Retention: nats.LimitsPolicy, // Keep messages until they expire
        MaxAge:    24 * time.Hour,    // Keep events for 24 hours
        Storage:   nats.FileStorage,
        Replicas:  1,
    }

    // Create the stream
    _, err = es.js.AddStream(streamConfig)
    if err != nil {
        return fmt.Errorf("failed to add stream: %w", err)
    }

    es.logger.Info("Created NATS JetStream", "stream", es.stream, "subject", es.subject)
    return nil
}
```

### Stream Configuration Details

- **Stream Name**: `HASHPOST_EVENTS`
- **Subject Pattern**: `hashpost.events.record.created`
- **Retention Policy**: `LimitsPolicy` (keep until expiration)
- **Max Age**: 24 hours
- **Storage**: File-based storage
- **Replicas**: 1 (single instance)

## Integration with PDS Endpoints

### Record Creation Events

**File**: `internal/pds/repo.go`  
**Method**: `handleCreateRecord`

When a record is created via the atproto API, the PDS publishes an event:

```go
func (s *Server) handleCreateRecord(w http.ResponseWriter, r *http.Request) {
    // ... authentication and validation ...

    // Create record in database
    // ... database operations ...

    // Publish event
    err := s.eventStream.PublishRecordEvent(
        r.Context(),
        EventTypeRecordCreated,
        session.DID,
        req.Collection,
        req.Record,
        uri,
        cid,
    )
    if err != nil {
        s.logger.Error("Failed to publish record event", "error", err)
        // Don't fail the request, just log the error
    }

    // ... return response ...
}
```

### Session Creation Events

**File**: `internal/pds/auth.go`  
**Method**: `handleCreateSession`

When a user creates a session, the PDS publishes a session event:

```go
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
    // ... authentication ...

    // Create session
    session, err := s.authService.AuthenticateSession(r.Context(), req.Identifier, req.Password)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Publish session event
    err = s.eventStream.PublishSessionEvent(
        r.Context(),
        EventTypeSessionCreated,
        session.DID,
        session.ID,
    )
    if err != nil {
        s.logger.Error("Failed to publish session event", "error", err)
    }

    // ... generate tokens and return response ...
}
```

## Event Payload Examples

### Record Created Event

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

### Session Created Event

```json
{
  "type": "session.created",
  "repo": "did:plc:hashpost-binding-test",
  "timestamp": "2025-01-01T00:00:00.000Z",
  "metadata": {
    "source": "hashpost-pds",
    "session_id": "session-uuid-123"
  }
}
```

### Identity Resolved Event

```json
{
  "type": "identity.resolved",
  "repo": "did:plc:hashpost-binding-test",
  "timestamp": "2025-01-01T00:00:00.000Z",
  "metadata": {
    "source": "hashpost-pds",
    "handle": "testuser.hashpost.local",
    "did": "did:plc:hashpost-binding-test"
  }
}
```

## Error Handling

### Publishing Errors

```go
// Publish event with error handling
err := s.eventStream.PublishRecordEvent(ctx, eventType, repo, collection, record, uri, cid)
if err != nil {
    s.logger.Error("Failed to publish record event", 
        "error", err,
        "event_type", eventType,
        "repo", repo,
        "collection", collection,
    )
    // Don't fail the request, just log the error
    // Events are not critical for API responses
}
```

### Stream Errors

```go
// Check stream existence before publishing
streamInfo, err := es.js.StreamInfo(es.stream)
if err != nil {
    es.logger.Error("Stream does not exist", "error", err, "stream", es.stream)
    return fmt.Errorf("stream does not exist: %w", err)
}
```

### Connection Errors

```go
// Handle NATS connection errors
nc, err := nats.Connect(natsURL)
if err != nil {
    return nil, fmt.Errorf("failed to connect to NATS: %w", err)
}
```

## Performance Considerations

### Async Publishing
- Events are published asynchronously
- API responses don't wait for event publishing
- Failed events are logged but don't affect API responses

### Message Size
- Events are serialized to JSON
- Large records may impact message size
- Consider compression for large payloads

### Stream Limits
- 24-hour retention policy
- File-based storage for persistence
- Single replica for development

## Monitoring and Debugging

### Logging

```go
es.logger.Debug("Event published successfully",
    "stream", ack.Stream,
    "sequence", ack.Sequence,
    "subject", subject,
)

es.logger.Info("Published atproto event",
    "type", event.Type,
    "repo", event.Repo,
    "subject", subject,
)
```

### Stream Information

```go
// Get stream information
streamInfo, err := es.js.StreamInfo(es.stream)
if err == nil {
    es.logger.Info("Stream exists",
        "stream", streamInfo.Config.Name,
        "subjects", streamInfo.Config.Subjects,
        "state", streamInfo.State,
    )
}
```

## References

- [Event Streamer Implementation](internal/pds/events.go)
- [Repository Handlers](internal/pds/repo.go)
- [Authentication Handlers](internal/pds/auth.go)
- [NATS JetStream Documentation](https://docs.nats.io/jetstream)
