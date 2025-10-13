# NATS Configuration

## Overview

The PDS uses NATS JetStream for event streaming to the AppView. This document covers the NATS configuration, stream setup, and connection management.

## NATS Connection

### Connection Setup

**File**: `internal/pds/events.go`  
**Method**: `NewEventStreamer`

```go
func NewEventStreamer(natsURL string, logger *slog.Logger) (*EventStreamer, error) {
    // Connect to NATS
    nc, err := nats.Connect(natsURL)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }

    // Create JetStream context
    js, err := wrappedConn.JetStream()
    if err != nil {
        nc.Close()
        return nil, fmt.Errorf("failed to create JetStream context: %w", err)
    }

    return streamer, nil
}
```

### Connection Configuration

**Environment Variable**: `NATS_URL`  
**Default**: `nats://localhost:4222`

```yaml
# docker-compose.yml
services:
  nats:
    image: nats:2.10-alpine
    ports:
      - "4222:4222"
    command: ["-js"]  # Enable JetStream
    volumes:
      - nats_data:/data
```

## Stream Configuration

### Stream Creation

**Method**: `createStream() error`

```go
func (es *EventStreamer) createStream() error {
    // Create stream configuration
    streamConfig := &nats.StreamConfig{
        Name:      es.stream,                    // "HASHPOST_EVENTS"
        Subjects:  []string{es.subject},         // ["hashpost.events.record.created"]
        Retention: nats.LimitsPolicy,            // Keep messages until they expire
        MaxAge:    24 * time.Hour,               // Keep events for 24 hours
        Storage:   nats.FileStorage,             // File-based storage
        Replicas:  1,                            // Single replica for development
    }

    // Create the stream
    _, err = es.js.AddStream(streamConfig)
    if err != nil {
        return fmt.Errorf("failed to add stream: %w", err)
    }

    return nil
}
```

### Stream Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Name** | `HASHPOST_EVENTS` | Stream identifier |
| **Subjects** | `hashpost.events.record.created` | Subject pattern for events |
| **Retention** | `LimitsPolicy` | Keep messages until expiration |
| **MaxAge** | `24h` | Message retention time |
| **Storage** | `FileStorage` | Persistent file storage |
| **Replicas** | `1` | Single replica for development |

## Subject Patterns

### Event Subjects

**Pattern**: `hashpost.events.{event_type}`

- `hashpost.events.record.created` - Record creation events
- `hashpost.events.record.updated` - Record update events  
- `hashpost.events.record.deleted` - Record deletion events
- `hashpost.events.identity.resolved` - Identity resolution events
- `hashpost.events.session.created` - Session creation events

### Subject Configuration

```go
streamer := &EventStreamer{
    nc:      wrappedConn,
    js:      js,
    logger:  logger,
    stream:  "HASHPOST_EVENTS",
    subject: "hashpost.events.record.created", // Primary subject
}
```

## Message Publishing

### Publishing Configuration

**Method**: `publishEvent(ctx context.Context, event *AtprotoEvent) error`

```go
func (es *EventStreamer) publishEvent(ctx context.Context, event *AtprotoEvent) error {
    // Serialize event to JSON
    data, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("failed to marshal event: %w", err)
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

### Message Acknowledgment

- **Synchronous Publishing**: Uses `js.Publish()` for immediate acknowledgment
- **Error Handling**: Returns error if publishing fails
- **Logging**: Logs successful publishing with stream and sequence info

## Docker Compose Configuration

### NATS Service

```yaml
# docker-compose.yml
services:
  nats:
    image: nats:2.10-alpine
    ports:
      - "4222:4222"
    command: ["-js"]  # Enable JetStream
    volumes:
      - nats_data:/data
    environment:
      - NATS_LOG_LEVEL=debug

volumes:
  nats_data:
```

### PDS Service Configuration

```yaml
# docker-compose.yml
services:
  hashpost-pds:
    build: .
    ports:
      - "8080:8080"
    environment:
      - NATS_URL=nats://nats:4222
      - ENVIRONMENT=development
    depends_on:
      - nats
      - postgres
```

## Environment Configuration

### Development Environment

```bash
# .env
NATS_URL=nats://localhost:4222
ENVIRONMENT=development
```

### Production Environment

```bash
# .env
NATS_URL=nats://nats-cluster:4222
ENVIRONMENT=production
```

## Stream Management

### Stream Information

**Method**: `StreamInfo(streamName string)`

```go
// Check if stream exists
streamInfo, err := es.js.StreamInfo(es.stream)
if err != nil {
    es.logger.Error("Stream does not exist", "error", err, "stream", es.stream)
    return fmt.Errorf("stream does not exist: %w", err)
}

es.logger.Debug("Stream exists",
    "stream", streamInfo.Config.Name,
    "subjects", streamInfo.Config.Subjects,
    "state", streamInfo.State,
)
```

### Stream Recreation

**Method**: `recreateStream() error`

```go
func (es *EventStreamer) recreateStream() error {
    // Delete the existing stream
    err := es.js.DeleteStream(es.stream)
    if err != nil {
        es.logger.Warn("Failed to delete existing stream", "error", err)
    }

    // Create new stream with updated configuration
    streamConfig := &nats.StreamConfig{
        Name:      es.stream,
        Subjects:  []string{es.subject},
        Retention: nats.LimitsPolicy,
        MaxAge:    24 * time.Hour,
        Storage:   nats.FileStorage,
        Replicas:  1,
    }

    _, err = es.js.AddStream(streamConfig)
    if err != nil {
        return fmt.Errorf("failed to recreate stream: %w", err)
    }

    return nil
}
```

## Error Handling

### Connection Errors

```go
// Handle NATS connection failures
nc, err := nats.Connect(natsURL)
if err != nil {
    return nil, fmt.Errorf("failed to connect to NATS: %w", err)
}

// Handle JetStream context creation failures
js, err := nc.JetStream()
if err != nil {
    nc.Close()
    return nil, fmt.Errorf("failed to create JetStream context: %w", err)
}
```

### Publishing Errors

```go
// Handle publishing failures
ack, err := es.js.Publish(subject, data)
if err != nil {
    es.logger.Error("Failed to publish event to NATS",
        "error", err,
        "subject", subject,
        "stream", es.stream,
    )
    return fmt.Errorf("failed to publish event: %w", err)
}
```

### Stream Errors

```go
// Handle stream creation failures
_, err = es.js.AddStream(streamConfig)
if err != nil {
    es.logger.Error("Failed to create NATS stream", "error", err, "stream", es.stream)
    return fmt.Errorf("failed to add stream: %w", err)
}
```

## Monitoring and Debugging

### Logging Configuration

```go
// Structured logging for NATS operations
es.logger.Info("Created NATS JetStream", 
    "stream", es.stream, 
    "subject", es.subject,
)

es.logger.Debug("Event published successfully",
    "stream", ack.Stream,
    "sequence", ack.Sequence,
    "subject", subject,
)
```

### Stream Monitoring

```go
// Monitor stream state
streamInfo, err := es.js.StreamInfo(es.stream)
if err == nil {
    es.logger.Info("Stream state",
        "stream", streamInfo.Config.Name,
        "subjects", streamInfo.Config.Subjects,
        "state", streamInfo.State,
        "messages", streamInfo.State.Msgs,
        "bytes", streamInfo.State.Bytes,
    )
}
```

## Performance Considerations

### Message Retention

- **MaxAge**: 24 hours for development
- **Storage**: File-based for persistence
- **Retention**: LimitsPolicy for automatic cleanup

### Connection Management

- **Connection Pooling**: Single connection per PDS instance
- **Reconnection**: Automatic reconnection on connection loss
- **Timeout**: Default NATS timeouts

### Message Size

- **JSON Serialization**: Events serialized to JSON
- **Compression**: Consider compression for large payloads
- **Batch Publishing**: Single message per event

## Security Considerations

### Network Security

- **Internal Network**: NATS accessible only within Docker network
- **Authentication**: No authentication required for development
- **TLS**: Consider TLS for production deployments

### Message Security

- **Event Validation**: Events validated before publishing
- **Source Identification**: Events tagged with source information
- **Error Handling**: Failed events logged but not retried

## Troubleshooting

### Common Issues

1. **Stream Not Found**
   - Check if stream exists before publishing
   - Verify stream configuration
   - Check NATS JetStream is enabled

2. **Connection Failures**
   - Verify NATS_URL is correct
   - Check NATS service is running
   - Verify network connectivity

3. **Publishing Failures**
   - Check stream exists and is accessible
   - Verify subject pattern matches stream configuration
   - Check message size limits

### Debug Commands

```bash
# Check NATS connection
nats server info

# List streams
nats stream list

# Check stream info
nats stream info HASHPOST_EVENTS

# Monitor messages
nats stream monitor HASHPOST_EVENTS
```

## References

- [NATS JetStream Documentation](https://docs.nats.io/jetstream)
- [Event Streamer Implementation](internal/pds/events.go)
- [Docker Compose Configuration](docker-compose.yml)
- [NATS Configuration](config/dev.yaml)
