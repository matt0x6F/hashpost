# Dual Server Pattern

## Overview

HashPost implements a dual-server architecture following the atproto specification pattern. This design separates concerns between protocol compliance (PDS) and application logic (AppView), enabling scalability and maintainability.

## Architecture Components

### PDS (Personal Data Server)

**Purpose**: Pure atproto protocol compliance only

**Responsibilities**:
- **Protocol Compliance**: Implements all atproto endpoints (`/xrpc/com.atproto.*`)
- **Data Storage**: Stores canonical atproto records
- **Authentication**: DID-based authentication using Bluesky Indigo libraries
- **Event Publishing**: Publishes events to NATS JetStream for AppView consumption

**Database**: `hashpost_pds_dev`
- **Schema**: Canonical atproto records with proper URI storage
- **Purpose**: Source of truth for all atproto data
- **Access Pattern**: Direct database access using SQLC queries

### AppView (Application View)

**Purpose**: Full stateful web application with business logic and persistence

**Responsibilities**:
- **Business Logic**: All forum logic, RBAC, moderation, voting, ownership models
- **API Endpoints**: Custom forum endpoints (`/api/v1/*`) + web UI routes
- **Data Storage**: Denormalized/aggregated data optimized for user-facing queries
- **Event Processing**: Consumes atproto events from PDS via NATS JetStream

**Database**: `hashpost_appview_dev`
- **Schema**: Denormalized tables with statistics and aggregated data
- **Purpose**: Optimized for read-heavy operations and user-facing queries
- **Access Pattern**: Event-driven updates from PDS events

## Separation Rationale

### Protocol vs Application Logic

**PDS Focus**:
- **Atproto Compliance**: Ensures full compatibility with atproto clients
- **Data Integrity**: Maintains canonical record structure
- **Authentication**: Handles DID resolution and JWT token management
- **Event Generation**: Publishes events for downstream consumption

**AppView Focus**:
- **User Experience**: Provides rich forum functionality
- **Business Logic**: Implements forum-specific features
- **Performance**: Optimized for user-facing queries
- **Scalability**: Can scale independently from PDS

### Database Separation

**PDS Database**:
- **Canonical Records**: Stores original atproto records
- **Data Integrity**: Ensures protocol compliance
- **Write-Heavy**: Optimized for record creation and updates
- **Schema**: Follows atproto data model

**AppView Database**:
- **Denormalized Data**: Pre-computed statistics and aggregated data
- **Read-Heavy**: Optimized for user-facing queries
- **Performance**: Indexes and denormalization for fast queries
- **Schema**: Forum-specific data model

## Communication Pattern

### Event-Driven Architecture

```
PDS → NATS JetStream → AppView
```

**Flow**:
1. **PDS**: User creates/updates/deletes atproto record
2. **PDS**: Publishes event to NATS JetStream
3. **AppView**: Consumes event from NATS JetStream
4. **AppView**: Updates denormalized data in AppView database

### Event Types

**Record Events**:
- `record.created` - New atproto record created
- `record.updated` - Existing record updated
- `record.deleted` - Record deleted

**Identity Events**:
- `identity.resolved` - DID resolved to handle
- `session.created` - User session created

### Event Payload Structure

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

## Benefits

### Scalability

**Independent Scaling**:
- **PDS**: Can scale based on atproto protocol load
- **AppView**: Can scale based on user-facing application load
- **Database**: Each service can optimize its database independently

**Load Distribution**:
- **PDS**: Handles protocol compliance and data storage
- **AppView**: Handles user interactions and business logic
- **NATS**: Provides reliable event streaming between services

### Maintainability

**Separation of Concerns**:
- **PDS**: Focus on atproto protocol compliance
- **AppView**: Focus on forum business logic
- **Clear Boundaries**: Well-defined interfaces between services

**Independent Development**:
- **PDS**: Can be updated for protocol changes
- **AppView**: Can be updated for feature changes
- **Minimal Coupling**: Services communicate only through events

### Reliability

**Fault Isolation**:
- **PDS Failure**: AppView can continue with cached data
- **AppView Failure**: PDS continues to serve atproto clients
- **Database Failure**: Each service has its own database

**Event Reliability**:
- **NATS JetStream**: Provides at-least-once delivery guarantees
- **Retry Logic**: Automatic retry for failed event processing
- **Dead Letter Queue**: Manual intervention for permanent failures

## Implementation Details

### PDS Implementation

**File Structure**:
```
internal/pds/
├── auth.go              # Authentication and session management
├── events.go            # Event publishing to NATS
├── repo.go              # Repository operations
├── server.go            # HTTP server and endpoint registration
└── identity_handlers.go # Identity resolution
```

**Key Components**:
- **AuthService**: DID resolution, JWT tokens, session management
- **EventStreamer**: Publishes events to NATS JetStream
- **Repository Handlers**: CRUD operations for atproto records

### AppView Implementation

**File Structure**:
```
internal/appview/
├── events.go            # Event consumption from NATS
├── handlers.go          # HTTP handlers for API endpoints
├── rbac.go              # Role-based access control
├── identity.go          # Identity resolution with caching
└── database.go          # Database operations
```

**Key Components**:
- **EventConsumer**: Consumes events from NATS JetStream
- **RBACService**: Role and permission management
- **IdentityResolver**: DID-to-handle resolution with caching

### Database Implementation

**PDS Database**:
```sql
-- Canonical atproto records
CREATE TABLE users (
    id UUID PRIMARY KEY,
    did VARCHAR(255) UNIQUE NOT NULL,
    handle VARCHAR(255) UNIQUE NOT NULL,
    -- ... other fields
);

CREATE TABLE posts (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    atproto_uri VARCHAR(500) UNIQUE,
    -- ... other fields
);
```

**AppView Database**:
```sql
-- Denormalized data for performance
CREATE TABLE appview_users (
    id UUID PRIMARY KEY,
    did VARCHAR(255) UNIQUE NOT NULL,
    handle VARCHAR(255) UNIQUE NOT NULL,
    post_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    reputation INTEGER DEFAULT 0
);

CREATE TABLE appview_posts (
    id UUID PRIMARY KEY,
    atproto_uri VARCHAR(500) UNIQUE NOT NULL,
    author_did VARCHAR(255) NOT NULL,
    author_handle VARCHAR(255) NOT NULL,
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    score INTEGER DEFAULT 0
);
```

## Configuration

### Environment Variables

**PDS Configuration**:
```bash
ENVIRONMENT=development
NATS_URL=nats://nats:4222
DATABASE_URL=postgres://hashpost:password@postgres:5432/hashpost_pds_dev
```

**AppView Configuration**:
```bash
ENVIRONMENT=development
NATS_URL=nats://nats:4222
DATABASE_URL=postgres://hashpost:password@postgres:5432/hashpost_appview_dev
PDS_URL=http://hashpost-pds:8080
```

### Docker Compose

```yaml
services:
  hashpost-pds:
    build: .
    ports:
      - "8080:8080"
    environment:
      - NATS_URL=nats://nats:4222
      - DATABASE_URL=postgres://hashpost:password@postgres:5432/hashpost_pds_dev
    depends_on:
      - nats
      - postgres

  hashpost-appview:
    build: .
    ports:
      - "8081:8081"
    environment:
      - NATS_URL=nats://nats:4222
      - DATABASE_URL=postgres://hashpost:password@postgres:5432/hashpost_appview_dev
      - PDS_URL=http://hashpost-pds:8080
    depends_on:
      - nats
      - postgres
```

## Monitoring and Observability

### Health Checks

**PDS Health Check**:
- **Database**: Check database connectivity
- **NATS**: Check NATS connection
- **Authentication**: Check identity directory

**AppView Health Check**:
- **Database**: Check database connectivity
- **NATS**: Check NATS connection
- **Event Processing**: Check event consumer status

### Metrics

**PDS Metrics**:
- **Request Rate**: Atproto endpoint requests per second
- **Response Time**: Average response time for endpoints
- **Event Publishing**: Events published to NATS per second
- **Database Operations**: Database query performance

**AppView Metrics**:
- **Request Rate**: API endpoint requests per second
- **Event Processing**: Events processed per second
- **Database Operations**: Database query performance
- **Error Rate**: Failed event processing rate

## Best Practices

### Development

1. **Clear Boundaries**: Maintain clear separation between PDS and AppView
2. **Event Design**: Design events for loose coupling
3. **Database Design**: Optimize each database for its use case
4. **Error Handling**: Implement comprehensive error handling

### Deployment

1. **Independent Deployment**: Deploy PDS and AppView independently
2. **Database Management**: Manage databases separately
3. **Monitoring**: Monitor each service independently
4. **Scaling**: Scale services based on their specific load patterns

### Maintenance

1. **Protocol Updates**: Update PDS for atproto protocol changes
2. **Feature Updates**: Update AppView for new features
3. **Database Maintenance**: Maintain each database independently
4. **Event Schema**: Version event schemas for compatibility

## References

- [Atproto Specification](https://atproto.com/specs/atp)
- [PDS Implementation](../PDS/README.md)
- [AppView Implementation](../AppView/README.md)
- [Event Processing Architecture](../designs/archive/event-processing-architecture.md)
