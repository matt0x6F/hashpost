# HashPost AppView Design

## Project Overview

This document tracks the implementation of the **HashPost AppView** as a stateless aggregator that presents a unified view of data from the HashPost PDS. The HashPost AppView aggregates and presents data to clients, while the HashPost PDS handles all data storage, business logic, and atproto protocol compliance.

## Design Goals

### Primary Objectives
- **Data Aggregation**: Aggregate data from HashPost PDS for unified presentation
- **Stateless Design**: No persistent storage, pure aggregation service
- **Type Safety**: Generate type-safe Go code from OpenAPI specs
- **Documentation**: Automatic API documentation generation
- **Client Generation**: Generate frontend clients from the same spec
- **PDS Integration**: AppView aggregates from HashPost PDS via APIs

### Technical Goals
- Implement HashPost AppView as stateless aggregator
- Aggregate data from HashPost PDS via APIs
- Present unified view to clients
- Improve API consistency and documentation
- Enable better frontend-backend integration
- Reduce manual API maintenance

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

### Phase 1: OpenAPI Specification
- [ ] Define core API requirements
- [ ] Design OpenAPI 3.0 specification
- [ ] Define request/response schemas
- [ ] Add authentication/authorization specs (if needed)
- [ ] Document error responses

### Phase 2: Server Generation
- [ ] Set up oapi-codegen configuration for net/http
- [ ] Generate server stubs with net/http handlers for HashPost AppView
- [ ] Implement middleware for authentication (if needed)
- [ ] Add request validation
- [ ] Set up error handling

### Phase 3: Integration
- [ ] Implement data aggregation in generated net/http handlers
- [ ] Integrate with PDS APIs for data retrieval
- [ ] Add comprehensive testing
- [ ] Set up net/http server for HashPost AppView

### Phase 4: Client Generation
- [ ] Generate TypeScript client for frontend
- [ ] Update frontend API calls
- [ ] Test end-to-end integration
- [ ] Update documentation

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

### PDS Dependency Management
- **Data Aggregation**: AppView aggregates data from PDS via APIs
- **No Business Logic**: AppView does not handle business logic - PDS handles all RBAC, analytics, moderation
- **No Database Access**: AppView has no persistent storage, only aggregates from PDS
- **Error Handling**: AppView handles PDS communication errors with appropriate fallbacks
- **User Context**: AppView retrieves complete user context and permissions from PDS
- **Stateless Design**: AppView is purely stateless aggregator

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
