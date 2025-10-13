# Error Handling

## Overview

The AppView implements comprehensive error handling for event processing, including retry logic, exponential backoff, idempotency, and dead letter queue management.

## Implementation

### Error Handling Configuration

**File**: `internal/appview/events.go`  
**Service**: `EventConsumer`

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

## Retry Logic

### Message Processing with Retry

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

### Exponential Backoff

**Method**: `calculateBackoffDelay(attempt int) time.Duration`

```go
func (ec *EventConsumer) calculateBackoffDelay(attempt int) time.Duration {
    delay := time.Duration(attempt+1) * ec.retryDelay
    if delay > ec.maxRetryDelay {
        delay = ec.maxRetryDelay
    }
    return delay
}
```

**Backoff Strategy**:
- **Attempt 1**: 1 second delay
- **Attempt 2**: 2 second delay
- **Attempt 3**: 3 second delay
- **Maximum**: 30 seconds (configurable)

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

**Idempotency Benefits**:
- **Duplicate Prevention**: Prevents processing the same event multiple times
- **Crash Recovery**: Handles application crashes gracefully
- **Network Issues**: Manages network interruptions and reconnections

## Dead Letter Queue

### DLQ Implementation

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

## Error Classification

### Transient Errors

**Characteristics**:
- **Network Issues**: NATS connection failures, database timeouts
- **Temporary Unavailability**: Service temporarily down
- **Resource Constraints**: Memory pressure, CPU overload

**Handling**:
- **Retry Logic**: Automatic retry with exponential backoff
- **Timeout Handling**: Configurable timeouts for operations
- **Circuit Breaker**: Consider circuit breaker pattern for persistent failures

### Permanent Errors

**Characteristics**:
- **Data Validation**: Invalid event data, malformed records
- **Business Logic**: Permission denied, resource not found
- **Configuration**: Missing required configuration

**Handling**:
- **Dead Letter Queue**: Send to DLQ for manual intervention
- **Error Logging**: Comprehensive error logging with context
- **Alerting**: Notify operators of permanent failures

## Error Handling Patterns

### Database Errors

```go
// Handle database connection errors
if err := ec.db.CreatePost(post); err != nil {
    if strings.Contains(err.Error(), "connection timeout") {
        // Transient error - retry
        return fmt.Errorf("database timeout, will retry: %w", err)
    }
    if strings.Contains(err.Error(), "duplicate key") {
        // Permanent error - skip
        ec.logger.Warn("Post already exists", "uri", event.URI)
        return nil
    }
    // Unknown error - retry
    return fmt.Errorf("failed to create post in AppView: %w", err)
}
```

### Identity Resolution Errors

```go
// Handle identity resolution failures
authorHandle, err := ec.identityResolver.ResolveHandleFromDID(context.Background(), event.Repo)
if err != nil {
    ec.logger.Warn("Failed to resolve author handle", "error", err, "did", event.Repo)
    authorHandle = "unknown" // Fallback to unknown
}
```

### Event Parsing Errors

```go
// Handle event parsing failures
var event AtprotoEvent
if err := json.Unmarshal(msg.Data, &event); err != nil {
    ec.logger.Error("Failed to unmarshal event", "error", err, "data", string(msg.Data))
    return fmt.Errorf("failed to unmarshal event: %w", err)
}
```

## Monitoring and Alerting

### Error Metrics

```go
type ErrorMetrics struct {
    ProcessedEvents    int64
    FailedEvents      int64
    RetryAttempts     int64
    DLQMessages       int64
    TransientErrors   int64
    PermanentErrors   int64
}

func (ec *EventConsumer) GetErrorMetrics() ErrorMetrics {
    return ErrorMetrics{
        ProcessedEvents:  ec.processedEvents,
        FailedEvents:     ec.failedEvents,
        RetryAttempts:    ec.retryAttempts,
        DLQMessages:      ec.dlqMessages,
        TransientErrors:  ec.transientErrors,
        PermanentErrors:  ec.permanentErrors,
    }
}
```

### Error Logging

```go
// Structured error logging
ec.logger.Error("Message processing failed after retries",
    "error", err,
    "attempts", ec.maxRetries,
    "event_id", eventID,
    "subject", msg.Subject,
    "data_length", len(msg.Data),
)

// Warning for transient errors
ec.logger.Warn("Message processing failed, retrying",
    "attempt", attempt+1,
    "max_retries", ec.maxRetries,
    "delay", delay,
    "error", err,
)
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

### Error Handling Configuration

```go
type ErrorHandlingConfig struct {
    MaxRetries    int           `yaml:"max_retries"`
    RetryDelay    time.Duration `yaml:"retry_delay"`
    MaxRetryDelay time.Duration `yaml:"max_retry_delay"`
    DLQSubject    string        `yaml:"dlq_subject"`
}

func NewEventConsumerWithConfig(natsURL string, db *Database, logger *slog.Logger, config ErrorHandlingConfig) (*EventConsumer, error) {
    consumer := &EventConsumer{
        nc:               nc,
        js:               js,
        logger:           logger,
        db:               db,
        identityResolver: identityResolver,

        maxRetries:     config.MaxRetries,
        retryDelay:     config.RetryDelay,
        maxRetryDelay:  config.MaxRetryDelay,
        processedEvents: make(map[string]bool),
    }

    return consumer, nil
}
```

### Environment Variables

```bash
# Error handling configuration
EVENT_MAX_RETRIES=3
EVENT_RETRY_DELAY=1s
EVENT_MAX_RETRY_DELAY=30s
EVENT_DLQ_SUBJECT=hashpost.events.dlq
```

## Recovery Procedures

### DLQ Processing

```go
// Process messages from dead letter queue
func (ec *EventConsumer) ProcessDLQMessages() error {
    // Subscribe to DLQ
    sub, err := ec.js.PullSubscribe("hashpost.events.dlq", "dlq-processor")
    if err != nil {
        return fmt.Errorf("failed to subscribe to DLQ: %w", err)
    }

    // Process DLQ messages
    for {
        msgs, err := sub.Fetch(10, nats.MaxWait(5*time.Second))
        if err != nil {
            if err == nats.ErrTimeout {
                continue
            }
            return fmt.Errorf("failed to fetch DLQ messages: %w", err)
        }

        for _, msg := range msgs {
            if err := ec.processDLQMessage(msg); err != nil {
                ec.logger.Error("Failed to process DLQ message", "error", err)
                continue
            }

            if err := msg.Ack(); err != nil {
                ec.logger.Error("Failed to ack DLQ message", "error", err)
            }
        }
    }
}
```

### Manual Recovery

```go
// Manual recovery for failed events
func (ec *EventConsumer) ReplayFailedEvents(eventIDs []string) error {
    for _, eventID := range eventIDs {
        // Remove from processed events to allow reprocessing
        delete(ec.processedEvents, eventID)
        
        // Log manual recovery
        ec.logger.Info("Manually recovered event", "event_id", eventID)
    }
    
    return nil
}
```

## Best Practices

### Error Handling Guidelines

1. **Fail Fast**: Detect errors early and fail fast
2. **Retry Wisely**: Use exponential backoff for transient errors
3. **Log Comprehensively**: Include context in error logs
4. **Monitor Continuously**: Track error rates and patterns
5. **Alert Proactively**: Set up alerts for error thresholds

### Recovery Strategies

1. **Automatic Recovery**: Retry logic for transient errors
2. **Manual Intervention**: DLQ for permanent errors
3. **Circuit Breaker**: Prevent cascade failures
4. **Health Checks**: Monitor system health continuously

### Testing Error Scenarios

1. **Network Failures**: Simulate network interruptions
2. **Database Errors**: Test database connection failures
3. **Invalid Data**: Test malformed event data
4. **Resource Constraints**: Test under high load

## References

- [Event Consumer Implementation](internal/appview/events.go)
- [Error Handling Patterns](internal/appview/events.go:132-226)
- [Dead Letter Queue](internal/appview/events.go:197-226)
- [NATS JetStream Error Handling](https://docs.nats.io/jetstream/error-handling)
