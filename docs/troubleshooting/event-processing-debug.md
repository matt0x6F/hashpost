# Event Processing Debug Notes

## Issue (RESOLVED ✅)
AppView was not receiving events from NATS Jetstream despite:
- PDS publishing events successfully
- AppView consumer running and fetching messages
- No error messages in logs

## Root Cause Analysis
The issue was caused by two main problems:

### 1. Build Errors Preventing Code Updates
- Air hot reload was failing to rebuild the PDS service due to compilation errors
- NATS async callback API was incompatible with the current NATS library version
- Error: `cannot use func(ack *nats.PubAck, err error) as nats.PubOpt value`

### 2. Subject Pattern Mismatch
- Stream was configured for `hashpost.events.*` (wildcard pattern)
- PDS was publishing to `hashpost.events.record.created` (specific subject)
- AppView was subscribing to `hashpost.events.*` (wildcard pattern)
- The wildcard pattern matching was not working as expected

## Solution Implemented

### 1. Fixed Build Errors
- Replaced async `PublishAsync()` with synchronous `Publish()`
- Removed problematic callback functions
- Ensured clean compilation and hot reload

### 2. Aligned Subject Patterns
- Changed stream configuration to use exact subject: `hashpost.events.record.created`
- Updated AppView subscription to match: `hashpost.events.record.created`
- Ensured perfect alignment between publisher, stream, and consumer

### 3. Added Comprehensive Debugging
- Added stream existence checks before publishing
- Added detailed logging for stream state and configuration
- Added event flow tracking from PDS → NATS → AppView

## Current Status (RESOLVED ✅)
- ✅ PDS successfully publishes events to NATS JetStream
- ✅ NATS stream stores events with sequence numbers
- ✅ AppView successfully consumes events and processes them
- ✅ End-to-end event delivery is working: PDS → NATS → AppView
- ✅ Enhanced error handling with retry logic and exponential backoff
- ✅ Idempotency handling prevents duplicate event processing
- ✅ Dead letter queue for permanently failed messages
- ✅ Comprehensive logging and error tracking

## Event Flow Verification
```
PDS: "Event published successfully" stream=HASHPOST_EVENTS sequence=1
NATS: Stream stores message with sequence number
AppView: "Received messages from NATS" count=1
AppView: "Processing atproto event" type=record.created
AppView: "Post data extracted" text="Test post content"
```

## Key Learnings
1. **Build errors can silently prevent hot reload** - Always check compilation status
2. **Subject pattern matching is critical** - Exact patterns work better than wildcards
3. **Debugging requires systematic approach** - Check each component in the chain
4. **NATS JetStream requires precise configuration** - Small mismatches cause failures

## Next Steps
- ✅ Event delivery is working
- ✅ Enhanced error handling and retry logic implemented
- ✅ Idempotency and deduplication added
- ✅ Dead letter queue for failed messages
- 🔄 Implement actual database storage in AppView
- 🔄 Add monitoring and alerting for event processing
- 🔄 Test under load and failure scenarios

