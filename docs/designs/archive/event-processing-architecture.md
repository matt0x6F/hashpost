# Event Processing Architecture

## Overview

This document outlines the event processing architecture for HashPost's atproto implementation, covering the flow from PDS event generation through NATS JetStream to AppView consumption and processing.

## Current Architecture

```
PDS (Personal Data Server)
    ↓ (publishes events)
NATS JetStream
    ↓ (streams events)
AppView (Application View)
    ↓ (stores denormalized data)
AppView Database
```

## Event Flow Components

### 1. Event Generation (PDS)
- **Location**: `internal/pds/events.go`
- **Responsibility**: Generate atproto events when records are created/updated/deleted
- **Event Types**: `record.created`, `record.updated`, `record.deleted`, `identity.resolved`, `session.created`

### 2. Event Streaming (NATS JetStream)
- **Stream**: `HASHPOST_EVENTS`
- **Subject**: `hashpost.events.record.created`
- **Retention**: 24 hours with limits policy
- **Storage**: File-based storage

### 3. Event Consumption (AppView)
- **Location**: `internal/appview/events.go`
- **Responsibility**: Consume events and update denormalized data
- **Processing**: Parse events, extract data, store in AppView database

## Current Implementation Status

### ✅ Working Components
- PDS event publishing to NATS JetStream
- NATS stream configuration and message storage
- AppView event consumption from NATS
- Basic event processing and logging

### ✅ Recently Improved
- ✅ Enhanced error handling with retry logic and exponential backoff
- ✅ Idempotency handling prevents duplicate event processing
- ✅ Dead letter queue for permanently failed messages
- ✅ Comprehensive logging and error tracking

### ✅ Recently Completed
- ✅ Database storage implementation - AppView now stores denormalized data
- ✅ Database separation - PDS and AppView use separate databases
- ✅ Event processing with retry logic and idempotency
- ✅ Dead letter queue for failed messages

### 🔄 Still Needs Improvement
- Complete AppView event handler implementations (9 TODOs remain)
- Monitoring and alerting
- Performance optimization under load
- Event processing under failure scenarios

## Event Processing Patterns

### 1. At-Least-Once Delivery
- NATS JetStream provides at-least-once delivery guarantees
- Messages are acknowledged only after successful processing
- Failed messages remain in the stream for retry

### 2. Event Ordering
- Events are processed in the order they arrive at the stream
- Each event has a sequence number for ordering
- Consumer processes events sequentially

### 3. Error Handling
- Failed message processing doesn't acknowledge the message
- Messages remain available for retry
- Error logging for debugging and monitoring

## Current Event Types

### Record Events
```go
type AtprotoEvent struct {
    Type       string                 `json:"type"`        // record.created, record.updated, record.deleted
    Repo       string                 `json:"repo"`       // DID of the repository
    Collection string                 `json:"collection"`  // com.hashpost.feed.post, com.hashpost.forum.subforum
    Record     map[string]interface{} `json:"record"`     // The actual record data
    URI        string                 `json:"uri"`        // at:// URI of the record
    CID        string                 `json:"cid"`        // Content identifier
    Timestamp  time.Time              `json:"timestamp"`  // When the event occurred
    Metadata   map[string]interface{} `json:"metadata"`   // Additional metadata
}
```

### Supported Collections
- `com.hashpost.feed.post` - HashPost feed posts
- `com.hashpost.forum.subforum` - HashPost subforums
- `com.hashpost.forum.comment` - HashPost comments (future)
- `com.hashpost.forum.vote` - HashPost votes (future)

## Event Processing Logic

### 1. Event Routing
Events are routed based on their type and collection:

```go
switch event.Type {
case "record.created":
    return ec.handleRecordCreated(event)
case "record.updated":
    return ec.handleRecordUpdated(event)
case "record.deleted":
    return ec.handleRecordDeleted(event)
}
```

### 2. Collection-Specific Handling
Within each event type, events are further routed by collection:

```go
switch event.Collection {
case "com.hashpost.feed.post":
    return ec.handlePostCreated(event)
case "com.hashpost.forum.subforum":
    return ec.handleSubforumCreated(event)
}
```

## Current Limitations

### 1. Error Handling
- Basic error logging only
- No retry logic for transient failures
- No dead letter queue for permanently failed messages
- No circuit breaker pattern

### 2. Performance
- Single-threaded event processing
- No batching of events
- No backpressure handling
- No rate limiting

### 3. Monitoring
- Basic logging only
- No metrics collection
- No alerting for failures
- No performance monitoring

### 4. Reliability
- No duplicate event detection
- No event ordering guarantees beyond NATS
- No transaction support
- No rollback capability

## Recommended Improvements

### 1. Enhanced Error Handling
- Implement exponential backoff retry logic
- Add dead letter queue for failed messages
- Implement circuit breaker pattern
- Add error classification (transient vs permanent)

### 2. Performance Optimization
- Implement event batching
- Add concurrent processing with worker pools
- Implement backpressure handling
- Add rate limiting and throttling

### 3. Monitoring and Observability
- Add Prometheus metrics
- Implement structured logging
- Add distributed tracing
- Create alerting rules

### 4. Reliability Improvements
- Implement idempotent event processing
- Add duplicate event detection
- Implement event ordering guarantees
- Add transaction support

## Next Steps

1. **Review current implementation** - Document existing patterns and limitations
2. **Implement error handling** - Add retry logic and dead letter queues
3. **Add monitoring** - Implement metrics and alerting
4. **Performance testing** - Test under load and failure scenarios
5. **Documentation** - Create operational runbooks and troubleshooting guides

## Files to Modify

- `internal/appview/events.go` - Enhance event processing logic
- `internal/pds/events.go` - Improve event publishing reliability
- `docker-compose.yml` - Add monitoring services
- `docs/designs/` - Add operational documentation
