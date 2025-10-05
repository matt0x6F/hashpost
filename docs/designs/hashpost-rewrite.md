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

## Progress Tracking

### Phase 1: Research and Planning
- [x] Evaluate database ORM alternatives (chose sqlc)
- [x] Research Go OpenAPI spec servers
- [ ] Research atproto protocol and data structures
- [ ] Design atproto-compatible architecture
- [ ] Design sqlc patterns and workflows
- [ ] Create implementation plan

### Phase 2: Core Infrastructure
- [ ] Set up HashPost PDS with database and atproto protocol compliance
- [ ] Set up HashPost AppView as stateless aggregator
- [ ] Set up sqlc configuration and workflows for PDS
- [ ] Create atproto data structures in PDS
- [ ] Implement RBAC and business logic in PDS
- [ ] Set up AppView aggregation from PDS
- [ ] Implement atproto-compliant architecture patterns

### Phase 3: Feature Implementation
- [ ] Define core features and requirements
- [ ] Build core functionality
- [ ] Build authentication (if needed)
- [ ] Build admin interface (if needed)

### Phase 4: Testing and Optimization
- [ ] Comprehensive testing
- [ ] Performance optimization
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

## Notes

- This is a living document that will be updated as the project progresses
- All agents should reference this document for context
- Progress should be tracked in the checkboxes above
- New decisions and findings should be documented here
- Architecture inspired by Tangled's atproto implementation patterns

---

*Last Updated: [Current Date]*
*Status: Planning Phase*
