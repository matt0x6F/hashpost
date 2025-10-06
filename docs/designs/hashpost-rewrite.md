# HashPost Rebuild Design Document

## Project Overview

This document tracks the progress of building HashPost from scratch on the atproto protocol. HashPost will consist of two main components: a **HashPost AppView** (stateless aggregator with OpenAPI API) and a **HashPost PDS** (Personal Data Server for atproto protocol compliance and business logic). This is a fresh start with modern architecture, improved performance, and enhanced maintainability.

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

### HashPost AppView
- **Approach**: Stateless aggregator with OpenAPI spec server
- **Rationale**: Provides unified data presentation with type-safe API and documentation

### HashPost PDS
- **Approach**: Atproto Personal Data Server for protocol compliance
- **Rationale**: Handles atproto protocol requirements, identity, and data storage

### Database Layer
- **Approach**: Fresh implementation with sqlc and PostgreSQL
- **Rationale**: sqlc provides zero runtime overhead, compile-time query validation, and excellent performance for complex queries

### Integration
- **Approach**: AppView aggregates data from PDS via APIs
- **Rationale**: Clean separation between custom logic and protocol compliance

### Component Communication Architecture
- **Authentication Flow**: PDS handles DID-based auth and session management
- **Data Flow**: PDS handles all data storage and business logic, AppView aggregates from PDS
- **Session Management**: PDS issues and validates session tokens
- **Error Handling**: AppView handles PDS communication errors gracefully with fallbacks
- **RBAC**: PDS handles all role-based access control and business logic
- **AppView**: Stateless aggregator that presents unified view to clients

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
- [ ] Create custom atproto types for HashPost features
- [ ] Implement HashPost PDS core using Indigo packages
- [ ] Implement HashPost AppView core with OpenAPI spec server

### Phase 3: Integration
- [ ] Integrate AppView with PDS for data aggregation
- [ ] Implement authentication system (DID-based auth in PDS)
- [ ] Create API endpoints (atproto in PDS, custom in AppView)

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

### Phase 2 Implementation Complete
- **Project Structure**: Complete directory structure with cmd/, internal/, config/ directories
- **Cobra CLI**: PDS and AppView binaries created with cobra framework
- **Taskfile**: Comprehensive development workflow with targets for dev, test, migrate, generate, build
- **Docker Compose**: Development and test environments configured
- **Database Schema**: Complete HashPost schema supporting atproto records and HashPost-specific features
- **Dependencies**: All Bluesky Indigo packages added and working

## Notes

- This is a living document that will be updated as the project progresses
- All agents should reference this document for context
- Progress should be tracked in the checkboxes above
- New decisions and findings should be documented here
- Architecture inspired by Tangled's atproto implementation patterns

---

*Last Updated: [Current Date]*
*Status: Phase 2 Complete - Ready for Phase 3 Implementation*
