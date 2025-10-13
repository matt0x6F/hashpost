# HashPost Architecture Design

## Project Overview

This document tracks the implementation of HashPost as a complete atproto application following Tangled's architecture pattern. HashPost consists of two main components: a **HashPost PDS** (Personal Data Server for atproto protocol compliance) and a **HashPost AppView** (stateful web application with business logic, persistence layer, and custom endpoints).

## Architecture Decisions

### HashPost PDS
- **Purpose**: Pure atproto protocol compliance only
- **Implementation**: Built using Bluesky Indigo Go libraries
- **API**: Standard atproto endpoints (`/xrpc/com.atproto.*`) only
- **Authentication**: DID-based authentication system using Indigo packages
- **Data Storage**: Canonical atproto records in separate database (`hashpost_pds_dev`)
- **Business Logic**: None - pure protocol compliance

### HashPost AppView
- **Purpose**: Full stateful web application with business logic and persistence
- **API**: Custom forum endpoints (`/api/v1/*`) + web UI routes
- **Authentication**: User session management and authorization
- **Data Storage**: Denormalized/aggregated data in separate database (`hashpost_appview_dev`)
- **Business Logic**: All forum logic, RBAC, moderation, voting, ownership models
- **Stateful Design**: Maintains application state and business logic

### Database Architecture
- **PDS Database**: `hashpost_pds_dev` - stores canonical atproto records only
- **AppView Database**: `hashpost_appview_dev` - stores denormalized/aggregated data + custom application data
- **Approach**: sqlc with PostgreSQL for both components
- **Rationale**: Clean separation of concerns with optimized data access patterns

### Integration
- **Event Streaming**: AppView consumes atproto events from PDS via NATS JetStream
- **Data Flow**: PDS stores canonical records → NATS JetStream events → AppView ingests and denormalizes
- **Authentication**: PDS handles DID resolution, AppView handles user sessions and business logic
- **Component Communication**: Event-driven architecture with proper separation of concerns

## Implementation Status

### ✅ Fully Implemented
- **Separate databases**: PDS and AppView databases exist with migrations
- **Event streaming**: PDS → NATS JetStream → AppView (working with retry/idempotency)
- **Database schemas**: Both PDS and AppView schemas complete
- **sqlc queries**: 27 PDS queries, 16 AppView queries generated
- **CLI structure**: Both `hashpost-pds` and `hashpost-appview` binaries working
- **Configuration system**: YAML configs for both components
- **Authentication system**: DID resolution, session management implemented
- **Event processing**: Infrastructure complete with enhanced error handling

### 🔄 Partially Implemented
- **PDS endpoints**: Core atproto endpoints implemented but 2 TODOs remain
- **AppView handlers**: Basic handlers exist but 9 TODOs for full implementation
- **Event processing**: Infrastructure complete but some handlers have placeholder logic
- **RBAC system**: Schema exists, handlers present, needs testing

### ❌ Not Implemented
- **Comprehensive testing**: Limited test coverage
- **Performance optimization**: No indexes, caching, or optimization
- **OAuth client registration**: OAuth code exists but incomplete
- **Monitoring/alerting**: No metrics or monitoring

## Progress Tracking

### Phase 1: Research and Planning ✅
- [x] Evaluate database ORM alternatives (chose sqlc)
- [x] Research Go OpenAPI spec servers
- [x] Research atproto protocol and data structures
- [x] Study Bluesky Indigo Go libraries
- [x] Set up sqlc configuration and workflows

### Phase 2: Core Infrastructure ✅
- [x] Set up cobra CLI framework for PDS and AppView binaries
- [x] Create Taskfile with development workflow targets
- [x] Add Bluesky Indigo dependencies to go.mod
- [x] Create Docker Compose configuration for development and testing
- [x] Build database layer with sqlc for both components
- [x] Create comprehensive database schema for HashPost features
- [x] Create custom atproto types for HashPost features
- [x] Implement HashPost PDS core using Indigo packages
- [x] Implement HashPost AppView core with OpenAPI spec server

### Phase 2.5: Implementation ✅
- [x] **Database Separation**: Separate PDS and AppView databases with proper schemas
- [x] **Event Streaming**: PDS → NATS JetStream → AppView data flow working
- [x] **AppView Data Storage**: AppView stores denormalized data from events
- [x] **Database Infrastructure**: Automated setup, migrations, and permissions
- [x] Implement real database integration for all PDS endpoints
- [x] **DID Authentication System**: Real atproto authentication using Bluesky Indigo libraries
- [x] **Session Management**: Token generation, validation, and refresh
- [x] **Environment-based Directory**: Mock directory for dev, real directory for production
- [x] **All atproto Endpoints**: createSession, getSession, refreshSession, deleteSession, resolveHandle

### Phase 3: Integration ✅
- [x] Integrate AppView with PDS for data aggregation
- [x] Implement authentication system (DID-based auth in PDS)
- [x] Create API endpoints (atproto in PDS, custom in AppView)
- [x] Fix mock identity directory integration for development
- [x] Complete end-to-end authentication flow (registration → login → session)

### Phase 4: Features and Testing
- [ ] Build core HashPost features (profiles, posts, social features)
- [ ] Implement comprehensive testing
- [ ] Optimize performance
- [ ] Security audit
- [ ] Documentation updates

## Key Components

### PDS Components
- **Atproto Server**: Built using Bluesky Indigo libraries
- **Authentication**: DID-based authentication system using Indigo packages
- **Data Management**: Atproto data structure handling with Indigo types
- **API Endpoints**: Atproto-compatible API using Indigo generated types
- **Database**: PostgreSQL with sqlc for canonical atproto records

### AppView Components
- **Web Application**: Full stateful application with business logic
- **API Endpoints**: Custom forum endpoints using OpenAPI spec
- **Event Processing**: Consumes atproto events from PDS via NATS JetStream
- **Database**: PostgreSQL with sqlc for denormalized data
- **RBAC System**: Role-based access control and permissions

### Event Processing
- **Event Generation**: PDS publishes atproto events to NATS JetStream
- **Event Streaming**: NATS JetStream with retry logic and idempotency
- **Event Consumption**: AppView processes events and updates denormalized data
- **Error Handling**: Enhanced error handling with dead letter queues

## Success Criteria

### Protocol Compliance
- Full atproto protocol compliance
- Successful authentication with atproto clients
- Proper data structure implementation
- API endpoint compatibility

### Performance
- API response times < 100ms for simple queries
- Database query optimization
- Efficient event processing
- Scalable architecture

### Maintainability
- Clear code organization
- Comprehensive test coverage
- Good documentation
- Clean separation of concerns

## Next Steps

1. **Complete AppView Implementation**: Finish remaining TODOs in AppView handlers
2. **Implement Core Features**: Build profiles, posts, social features using atproto types
3. **Add Comprehensive Testing**: Unit tests, integration tests, atproto compliance tests
4. **Performance Optimization**: Query optimization, API response times, memory usage
5. **Security Audit**: Comprehensive security validation

## Notes

- This architecture follows Tangled's atproto implementation pattern
- AppView is stateful (not stateless) and maintains its own database
- Both PDS and AppView have separate databases for proper separation of concerns
- Event-driven architecture ensures data consistency between components
- All agents should reference this document for context

---

*Last Updated: [Current Date]*
*Status: Phase 3 Complete - Phase 4 Implementation Needed*
