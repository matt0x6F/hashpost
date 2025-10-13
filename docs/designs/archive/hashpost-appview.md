# HashPost AppView Design

## Project Overview

This document tracks the implementation of the **HashPost AppView** as a full web application with business logic, persistence layer, and custom endpoints. The HashPost AppView handles all forum functionality, while the HashPost PDS provides pure atproto protocol compliance.

## Design Goals

### Primary Objectives
- **Business Logic**: Handle all forum functionality, user management, and content operations
- **Persistence Layer**: Maintain denormalized data and application state
- **Custom Endpoints**: Provide forum-specific API endpoints
- **Web Application**: Serve web UI and handle user interactions
- **Event Processing**: Consume atproto events from PDS via Jetstream
- **Data Management**: Aggregate and cache data for optimal performance

### Technical Goals
- Implement HashPost AppView as full web application
- Process atproto events from PDS via Jetstream
- Maintain application state and business logic
- Provide custom forum API endpoints
- Serve web UI and handle user interactions
- Optimize data access and caching

## Architecture Decision

**Chosen Tool: oapi-codegen with net/http server**

**Rationale:**
- Modern OpenAPI 3.0 support aligns with current standards
- Excellent type safety with generated Go code
- Good performance with standard net/http server
- Active development ensures long-term support
- Clean separation between generated code and business logic
- Uses net/http codegen for HashPost AppView implementation

## Implementation Plan

### Phase 1: OpenAPI Specification ✅
- [x] Define core API requirements
- [x] Design OpenAPI 3.0 specification
- [x] Define request/response schemas
- [x] Add authentication/authorization specs
- [x] Document error responses

### Phase 2: Server Generation ✅
- [x] Set up oapi-codegen configuration for net/http
- [x] Generate server stubs with net/http handlers for HashPost AppView
- [x] Implement middleware for authentication
- [x] Add request validation
- [x] Set up error handling

### Phase 3: Integration ✅
- [x] Implement data aggregation in generated net/http handlers
- [x] Integrate with PDS APIs for data retrieval
- [x] Add comprehensive testing
- [x] Set up net/http server for HashPost AppView
- [x] Implement authentication endpoints (login, register, logout, session)
- [x] Add PDS proxy functionality for authentication

### Phase 4: Client Generation ✅
- [x] Generate TypeScript client for frontend
- [x] Update frontend API calls
- [x] Test end-to-end integration
- [x] Update documentation

## Key Components

### API Endpoints
- TBD based on requirements

### Schema Definitions
- TBD based on requirements

### Security
- **Authentication**: TBD based on requirements
- **Authorization**: TBD based on requirements
- **Validation**: Request/response validation
- **Rate Limiting**: API endpoint protection

### Database and State Management ✅
- **✅ Persistence Layer**: AppView maintains separate database (`hashpost_appview_dev`) for denormalized data
- **✅ Business Logic**: AppView handles all forum business logic, RBAC, and user management
- **✅ Event Processing**: AppView consumes atproto events from PDS via NATS JetStream
- **✅ Data Storage**: AppView stores aggregated data, user sessions, and application state
- **✅ Database Separation**: Complete separation from PDS database with proper schemas
- **✅ Event-Driven Updates**: Real-time data synchronization via event streaming
- **✅ Error Handling**: AppView handles all business logic errors and user interactions
- **User Context**: AppView manages user sessions, authentication, and permissions (pending)
- **Stateful Design**: AppView maintains application state and business logic

## Success Criteria

### Performance
- API response times < 100ms for simple queries
- Generated code performance equivalent to hand-written
- Minimal overhead from validation

### Developer Experience
- Clear API documentation
- Type-safe client generation
- Easy endpoint testing
- Good error messages

### Maintainability
- Single source of truth for API spec
- Automatic client updates
- Clear separation of concerns
- Comprehensive test coverage

## Next Steps

1. **Define API Requirements**: What endpoints do we need?
2. **Create OpenAPI Specification**: Start with core endpoints
3. **Set up oapi-codegen**: Configure for net/http server
4. **Generate Initial Server**: Create basic server structure
5. **Test Integration**: Ensure everything works together

## Notes

- Focus on OpenAPI 3.0 for modern standards
- Use oapi-codegen with net/http server approach for HashPost AppView
- Plan for future API versioning
- Keep generated code separate from business logic
- AppView uses net/http codegen for custom HashPost features

---

*Last Updated: [Current Date]*
*Status: Research Phase*
