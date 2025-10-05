# HashPost AppView Design

## Project Overview

This document tracks the implementation of the **HashPost AppView** for custom application logic and business rules. The HashPost AppView provides custom HashPost features through an OpenAPI spec server, while the HashPost PDS handles atproto protocol compliance.

## Design Goals

### Primary Objectives
- **AppView API**: Custom HashPost application logic and business rules
- **Type Safety**: Generate type-safe Go code from OpenAPI specs
- **Documentation**: Automatic API documentation generation
- **Client Generation**: Generate frontend clients from the same spec
- **PDS Integration**: AppView calls HashPost PDS for data operations

### Technical Goals
- Implement HashPost AppView with OpenAPI spec server
- Integrate AppView with HashPost PDS for data operations
- Improve API consistency and documentation
- Enable better frontend-backend integration
- Reduce manual API maintenance

## Research Findings

### Go OpenAPI Tools Comparison

#### 1. **oapi-codegen** (Recommended)
**Pros:**
- Modern OpenAPI 3.0 support
- Excellent type safety
- Good performance
- Active development
- Clean generated code

**Cons:**
- Requires OpenAPI spec maintenance
- Learning curve for spec-first development

**Best for:** Type-safe APIs with complex schemas

#### 2. **gin-swagger**
**Pros:**
- Integrates well with Gin framework
- Good documentation generation
- Familiar to Gin users

**Cons:**
- Limited to Gin framework
- Less type safety than oapi-codegen
- Swagger 2.0 focused

**Best for:** Existing Gin applications

#### 3. **go-swagger**
**Pros:**
- Comprehensive toolkit
- Good documentation
- Mature project

**Cons:**
- Primarily Swagger 2.0
- More complex setup
- Less modern than oapi-codegen

**Best for:** Legacy Swagger 2.0 projects

## Architecture Decision

**Recommended Tool: oapi-codegen with net/http server**

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
- [ ] Implement business logic in generated net/http handlers
- [ ] Integrate with database layer (sqlc)
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
