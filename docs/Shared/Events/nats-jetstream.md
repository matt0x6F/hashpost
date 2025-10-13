# NATS JetStream

## Overview

HashPost uses NATS JetStream for reliable event streaming between PDS and AppView services. NATS JetStream provides at-least-once delivery guarantees, persistence, and scalability for the event-driven architecture.

## NATS Configuration

### Docker Compose Setup

```yaml
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

### Connection Configuration

**PDS Connection**:
```go
// internal/pds/events.go
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

    return &EventStreamer{
        nc:     nc,
        js:     js,
        logger: logger,
        stream: "HASHPOST_EVENTS",
        subject: "hashpost.events.record.created",
    }, nil
}
```

**AppView Connection**:
```go
// internal/appview/events.go
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

    return &EventConsumer{
        nc:     nc,
        js:     js,
        logger: logger,
        db:     db,
    }, nil
}
```

## Stream Configuration

### Stream Creation

**PDS Stream Creation**:
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
        Retention: nats.LimitsPolicy,
        MaxAge:    24 * time.Hour,
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
// PDS subject configuration
streamer := &EventStreamer{
    nc:      wrappedConn,
    js:      js,
    logger:  logger,
    stream:  "HASHPOST_EVENTS",
    subject: "hashpost.events.record.created", // Primary subject
}

// AppView subject configuration
consumer := &EventConsumer{
    nc:      nc,
    js:      js,
    logger:  logger,
    db:      db,
    subject: "hashpost.events.record.created", // Consume from same subject
}
```

## Message Publishing

### Publishing Configuration

**PDS Publishing**:
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

## Message Consumption

### Consumer Configuration

**AppView Consumer**:
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

### Consumer Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Subject** | `hashpost.events.record.created` | Subject to consume from |
| **Consumer Name** | `hashpost-appview` | Consumer identifier |
| **Stream Binding** | `HASHPOST_EVENTS` | Explicit stream binding |
| **Batch Size** | `10` | Messages per fetch |
| **Timeout** | `5s` | Max wait time for messages |

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

### Consumption Errors

```go
// Handle consumption failures
msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
if err != nil {
    if err == nats.ErrTimeout {
        continue
    }
    ec.logger.Error("Failed to fetch messages", "error", err)
    continue
}
```

## Dead Letter Queue

### DLQ Configuration

**DLQ Subject**: `hashpost.events.dlq`

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

### DLQ Message Structure

```json
{
  "original_subject": "hashpost.events.record.created",
  "original_data": {
    "type": "record.created",
    "repo": "did:plc:hashpost-binding-test",
    "collection": "app.hashpost.feed.post",
    "record": {
      "$type": "app.hashpost.feed.post",
      "text": "Hello, HashPost!"
    },
    "uri": "at://did:plc:hashpost-binding-test/app.hashpost.feed.post/123",
    "timestamp": "2025-01-01T00:00:00.000Z"
  },
  "error": "failed to create post in AppView: database connection timeout",
  "timestamp": "2025-01-01T00:00:00.000Z"
}
```

## Monitoring and Debugging

### Stream Information

```go
// Get stream information
streamInfo, err := es.js.StreamInfo(es.stream)
if err == nil {
    es.logger.Info("Stream exists",
        "stream", streamInfo.Config.Name,
        "subjects", streamInfo.Config.Subjects,
        "state", streamInfo.State,
        "messages", streamInfo.State.Msgs,
        "bytes", streamInfo.State.Bytes,
    )
}
```

### Consumer Information

```go
// Get consumer information
consumerInfo, err := ec.js.ConsumerInfo(es.stream, "hashpost-appview")
if err == nil {
    ec.logger.Info("Consumer exists",
        "stream", consumerInfo.Stream,
        "consumer", consumerInfo.Name,
        "delivered", consumerInfo.Delivered.Consumer,
        "ack_floor", consumerInfo.AckFloor.Consumer,
    )
}
```

### Health Checks

```go
// Health check for NATS connection
func (es *EventStreamer) HealthCheck() error {
    if !es.nc.IsConnected() {
        return fmt.Errorf("NATS connection lost")
    }

    // Check stream exists
    streamInfo, err := es.js.StreamInfo(es.stream)
    if err != nil {
        return fmt.Errorf("stream does not exist: %w", err)
    }

    return nil
}
```

## Performance Considerations

### Message Retention

- **MaxAge**: 24 hours for development
- **Storage**: File-based for persistence
- **Retention**: LimitsPolicy for automatic cleanup

### Connection Management

- **Connection Pooling**: Single connection per service
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

# Check consumers
nats consumer list HASHPOST_EVENTS

# Monitor consumer
nats consumer monitor HASHPOST_EVENTS hashpost-appview
```

## Production Considerations

### High Availability

- **Clustering**: Use NATS clustering for high availability
- **Replication**: Configure stream replication
- **Monitoring**: Monitor cluster health

### Scalability

- **Horizontal Scaling**: Scale NATS cluster horizontally
- **Load Balancing**: Distribute load across cluster
- **Performance**: Monitor and optimize performance

### Backup and Recovery

- **Stream Backup**: Backup stream data
- **Consumer State**: Backup consumer state
- **Recovery**: Implement recovery procedures

## References

- [NATS JetStream Documentation](https://docs.nats.io/jetstream)
- [NATS Configuration](https://docs.nats.io/running-a-nats-service/configuration)
- [Event Streaming](event-streaming.md)
- [PDS Event Publishing](../PDS/Events/event-publishing.md)
- [AppView Event Consumption](../AppView/Events/event-consumption.md)
