# HashPost Rebuild Design Document

## Project Overview

This document tracks the progress of building HashPost from scratch on the atproto protocol. HashPost will use the standard atproto architecture: a **HashPost PDS** (Personal Data Server for atproto protocol compliance) and a **HashPost AppView** (web application with business logic, persistence layer, and custom endpoints). This architecture provides clean separation between protocol compliance and application logic.

**Related Design Documents:**
- `hashpost-appview.md` - HashPost AppView implementation details
- `hashpost-pds.md` - HashPost PDS implementation details
- `sqlc-patterns-workflows.md` - Database patterns for PDS component
- `go-command-structure.md` - CLI structure for both components

## Design Goals

### Primary Objectives
- **Atproto Compatibility**: Build on the atproto protocol foundation
- **Decentralized Architecture**: Leverage atproto's decentralized identity and data structures
- **Modern Architecture**: Build a maintainable and scalable design
- **Performance**: Optimize database queries and API responses
- **Developer Experience**: Clean code organization and comprehensive testing

### Technical Goals
- Build HashPost AppView with OpenAPI spec server
- Build HashPost PDS for atproto protocol compliance
- Implement atproto data structures and protocols
- Build with sqlc for better performance and type safety
- Create modern frontend-backend integration
- Design appropriate security measures

## Architecture Decisions

### HashPost PDS
- **Approach**: Pure atproto protocol compliance server
- **Responsibilities**: 
  - Standard atproto endpoints (`/xrpc/com.atproto.*`)
  - DID resolution and authentication
  - Atproto record storage and retrieval
  - Git operations (if needed)
- **Rationale**: Keeps PDS focused on protocol compliance only

### HashPost AppView
- **Approach**: Full web application with business logic and persistence
- **Responsibilities**:
  - Custom forum endpoints (`/api/v1/*`)
  - User authentication and authorization
  - Business logic for posts, comments, votes
  - Data aggregation and caching
  - Web UI routes
- **Rationale**: AppView handles all custom application logic

### Database Layer
- **PDS Database**: Stores canonical atproto records only
- **AppView Database**: Stores denormalized/aggregated data + custom application data
- **Approach**: sqlc with PostgreSQL for both components
- **Rationale**: Clean separation of concerns with optimized data access patterns

### Integration
- **Approach**: AppView consumes atproto events from PDS via Jetstream
- **Data Flow**: PDS → Jetstream → AppView Ingester → AppView Database
- **Rationale**: Event-driven architecture with proper separation of concerns

### Component Communication Architecture
- **Authentication Flow**: PDS handles DID-based auth, AppView handles session management
- **Data Flow**: PDS stores canonical records → Jetstream events → AppView ingests and denormalizes
- **Session Management**: AppView handles user sessions and authentication state
- **Error Handling**: AppView handles all business logic errors, PDS focuses on protocol errors
- **RBAC**: AppView handles all role-based access control and business logic
- **AppView**: Full web application with persistence layer and business logic

## Implementation Dependencies

### 🚀 START HERE - Independent Tasks (Can be done in parallel)
- [ ] **Research atproto protocol** - Understand data structures, authentication, API patterns
- [ ] **Study Bluesky Indigo libraries** - Understand package structure, PDS patterns
- [ ] **Set up sqlc configuration** - Create sqlc.yaml and query directory structure

### 📋 Foundation Phase (Depends on Research)
- [ ] **Set up cobra CLI framework** - Initialize CLI structure for separate binaries
- [ ] **Create custom atproto types** - Define lexicons for posts, profiles, social features

### 🏗️ Core Implementation Phase (Depends on Foundation)
- [ ] **Implement PDS core** - Set up DID resolution and authentication using Indigo
- [ ] **Build database layer** - Implement user, post, and permission queries with sqlc
- [ ] **Implement AppView core** - Set up oapi-codegen configuration for stateless aggregator

### 🔗 Integration Phase (Depends on Core Implementation)
- [ ] **Integrate AppView with PDS** - Implement data aggregation from PDS APIs
- [ ] **Implement authentication system** - DID-based auth in PDS, session management
- [ ] **Create API endpoints** - Atproto endpoints in PDS, custom endpoints in AppView

### 🎯 Features Phase (Depends on Integration)
- [ ] **Build core HashPost features** - Profiles, posts, social features using atproto types
- [ ] **Implement comprehensive testing** - Unit tests, integration tests, atproto compliance tests
- [ ] **Optimize performance** - Query optimization, API response times, memory usage

### 📊 Dependency Flow Diagram
```
🚀 START HERE (Parallel)
├── Research atproto protocol
├── Study Bluesky Indigo libraries  
└── Set up sqlc configuration
    ↓
📋 Foundation Phase
├── Set up cobra CLI framework
└── Create custom atproto types
    ↓
🏗️ Core Implementation Phase (Parallel)
├── Implement PDS core
├── Build database layer
└── Implement AppView core
    ↓
🔗 Integration Phase
├── Integrate AppView with PDS
├── Implement authentication system
└── Create API endpoints
    ↓
🎯 Features Phase
├── Build core HashPost features
├── Implement comprehensive testing
└── Optimize performance
```

## Progress Tracking

### Phase 1: Research and Planning
- [x] Evaluate database ORM alternatives (chose sqlc)
- [x] Research Go OpenAPI spec servers
- [x] Research atproto protocol and data structures
- [x] Study Bluesky Indigo Go libraries
- [x] Set up sqlc configuration and workflows

### Phase 2: Core Infrastructure
- [x] Set up cobra CLI framework for PDS and AppView binaries
- [x] Create Taskfile with development workflow targets
- [x] Add Bluesky Indigo dependencies to go.mod
- [x] Create Docker Compose configuration for development and testing
- [x] Build database layer with sqlc for PDS
- [x] Create comprehensive database schema for HashPost features
- [x] Create custom atproto types for HashPost features
- [x] Implement HashPost PDS core using Indigo packages
- [x] Implement HashPost AppView core with OpenAPI spec server

**Status: INFRASTRUCTURE COMPLETE, IMPLEMENTATION PENDING** - Infrastructure ready but endpoints return mock data

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
- [ ] Add comprehensive testing for real implementations

**Status: DID AUTHENTICATION COMPLETE** - Full atproto authentication system implemented

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

### Database Models
- TBD based on requirements

### API Endpoints
- TBD based on requirements

### Frontend Components
- TBD based on requirements

## Success Criteria

### Performance
- API response times < 100ms for simple queries
- Database query optimization
- Reduced memory usage

### Maintainability
- Clear code organization
- Comprehensive test coverage
- Good documentation

### Security
- Appropriate security measures
- Secure authentication (if needed)
- Security audit compliance

## Next Steps

1. **Research atproto Protocol**: Understand data structures, authentication, and API patterns
2. **Design atproto Architecture**: Plan how to implement atproto compatibility
3. **Design sqlc Patterns**: Define query organization and DAO patterns
4. **Create Implementation Plan**: Break down work into manageable tasks

## Research Findings

### Tangled Architecture Analysis
- **PDS (knotserver)**: Pure atproto protocol compliance - only handles `/xrpc/com.atproto.*` endpoints and git operations
- **AppView (appview)**: Full web application with:
  - Custom business logic and endpoints
  - Comprehensive persistence layer (SQLite database)
  - User authentication and authorization
  - Data aggregation from PDS via Jetstream events
  - Web UI routes and pages
- **Event Streaming**: AppView uses Jetstream to consume atproto events from PDS
- **Data Flow**: PDS stores canonical records → Jetstream → AppView Ingester → AppView Database
- **Key Insight**: AppView is NOT stateless - it's a full web application with its own database

### atproto Protocol Research
- **Core Components**: atproto consists of PDS (Personal Data Server), BGS (Big Graph Service), and various supporting services
- **Identity System**: Uses DIDs (Decentralized Identifiers) for user identity and authentication
- **Data Repositories**: Users maintain personal data repositories for data portability
- **Lexicon System**: Defines data structures and API endpoints through lexicon schemas
- **Key Lexicons**: `com.atproto.server`, `com.atproto.identity`, `app.bsky.feed`, `app.bsky.graph`

### Bluesky Indigo Libraries Research
- **Package Structure**: Indigo provides Go implementations of atproto components
- **Key Packages**: 
  - `atproto/identity` - DID and handle resolution
  - `atproto/crypto` - Cryptographic signing and key serialization
  - `atproto/syntax` - String types and parsers for identifiers
  - `atproto/lexicon` - Schema validation of data
  - `repo` - Account data storage
  - `mst` - Merkle Search Tree implementation
  - `xrpc` - HTTP API client
- **PDS Implementation**: Available in Indigo for building Personal Data Servers
- **Integration**: Can be used as foundation for HashPost PDS implementation

### sqlc Configuration
- **Setup Complete**: sqlc.yaml configured with PostgreSQL support
- **Directory Structure**: Created `internal/database/queries` and `internal/database/generated`
- **Comprehensive Schema**: Complete HashPost database schema with users, subforums, posts, comments, votes
- **Code Generation**: Successfully generating Go code from SQL queries
- **Type Safety**: Using pgx/v5 with proper type overrides for UUID and timestamps

### Phase 2 Implementation Status

#### ✅ Infrastructure Setup Complete
- **Project Structure**: Complete directory structure with cmd/, internal/, config/ directories
- **Cobra CLI**: PDS and AppView binaries created with cobra framework
- **Taskfile**: Comprehensive development workflow with targets for dev, test, migrate, generate, build
- **Docker Compose**: Development and test environments configured
- **Database Schema**: Complete HashPost schema supporting atproto records and HashPost-specific features
- **Dependencies**: All Bluesky Indigo packages added and working

#### ✅ Infrastructure Complete
- **Custom Atproto Types**: HashPost lexicon defined with profile, subforum, post, comment, vote, and subscription types
- **PDS Core**: Basic PDS server structure with endpoint handlers (returns mock data)
- **AppView Core**: AppView server structure with OpenAPI spec (calls PDS but gets mock data)
- **Configuration Management**: Complete configuration system for both components
- **Testing Infrastructure**: Unit tests for both PDS and AppView components
- **Database Schema**: Complete database schema with sqlc integration

#### ❌ Implementation Missing
- **Real Database Integration**: Most PDS endpoints return hardcoded JSON instead of database queries
- **Real Business Logic**: No actual CRUD operations implemented
- **Real AppView Aggregation**: AppView calls PDS but gets fake data
- **Error Handling**: Limited error handling and validation
- **Authentication**: No real authentication system implemented

**Phase 2 Status: INFRASTRUCTURE COMPLETE, IMPLEMENTATION PENDING** - Ready for Phase 2.5 implementation work

## Current Implementation Reality

### What We Actually Have
- **Working Infrastructure**: Docker Compose, database, sqlc, CLI commands all work
- **PDS Server**: Basic atproto protocol server with custom endpoints (needs refactoring)
- **AppView Server**: Basic OpenAPI server (needs full implementation)
- **Database Schema**: Complete schema with proper relationships and constraints
- **Configuration**: Full configuration system for both components
- **Testing**: Basic unit tests for server creation and configuration

### What We Need to Refactor (Phase 2.5)
- **PDS Refactoring**: Remove custom endpoints, focus on pure atproto protocol compliance
- **AppView Implementation**: Build full web application with business logic and persistence
- **Event Streaming**: Implement Jetstream integration for PDS → AppView data flow
- **Database Separation**: PDS for canonical records, AppView for denormalized data
- **Authentication**: Move user auth to AppView, keep DID resolution in PDS
- **Custom Endpoints**: Move all `/api/v1/*` endpoints from PDS to AppView

### Current Architecture (Needs Refactoring)
```bash
# Current: Custom endpoints on PDS (WRONG)
curl http://localhost:8080/api/v1/subforums
# Should be: Pure atproto protocol only

# Current: AppView calls PDS (WRONG)
curl http://localhost:8081/api/v1/subforums  
# Should be: AppView handles custom endpoints directly
```

### Target Architecture
```bash
# PDS: Pure atproto protocol compliance
curl http://localhost:8080/xrpc/com.atproto.server.getSession
# Returns: Standard atproto response

# AppView: Custom forum endpoints with business logic
curl http://localhost:8081/api/v1/subforums
# Returns: Real data from AppView database with business logic
```

## Notes

- This is a living document that will be updated as the project progresses
- All agents should reference this document for context
- Progress should be tracked in the checkboxes above
- New decisions and findings should be documented here
- Architecture inspired by Tangled's atproto implementation patterns

---

*Last Updated: [Current Date]*
*Status: Phase 2 Infrastructure Complete - Phase 2.5 Implementation Needed*
