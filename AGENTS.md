# Agent Guidelines for HashPost

## Design Document System

The `docs/designs/` directory contains design documents that track progress toward project goals. These documents serve as coordination points for agents working on the project.

### Design Document Workflow

1. **Always reference design documents** when working on related features
2. **Update progress checkboxes** as work is completed
3. **Document decisions** in the Architecture Decisions section
4. **Add notes** for important findings or changes
5. **Keep documents current** - update last modified dates

## Agent Coordination System

### Multi-Agent Workflow
1. **Check Design Documents**: Always review relevant design documents first
2. **Update Progress**: Mark completed work in design documents
3. **Coordinate Changes**: Use design document notes for coordination
4. **Avoid Duplication**: Check existing work before starting

### Change Management
- **Document Decisions**: Explain why technical choices were made
- **Update Progress**: Keep design documents current
- **Coordinate Updates**: Use notes section for important changes
- **Maintain Context**: Ensure all agents have current information

## Quality Standards

### Code Quality
- **Readability**: Code should be self-documenting
- **Maintainability**: Easy to modify and extend
- **Performance**: Optimize for the use case
- **Security**: Follow security best practices

### Documentation Quality
- **Clarity**: Clear and concise explanations
- **Completeness**: Cover all necessary information
- **Accuracy**: Keep documentation current with code
- **Accessibility**: Easy to find and understand

## Best Practices

### Development Workflow
- **Read before writing**: Always check existing design documents
- **Update as you go**: Don't wait until the end to update progress
- **Document decisions**: Explain why choices were made
- **Be specific**: Use concrete examples in notes
- **Stay current**: Update documents as the project evolves

### Communication
- **Clear Context**: Provide sufficient context in all communications
- **Specific Examples**: Use concrete examples when explaining concepts
- **Current Information**: Ensure all information is up-to-date
- **Coordinated Updates**: Keep all agents informed of changes

## Architecture Patterns

### Dual-Server Architecture
- **PDS (Personal Data Server)**: Handles atproto protocol compliance and stores canonical data including forum tables
- **AppView**: Stateful web application with business logic and persistence that provides forum functionality
- **Database Separation**: PDS uses `hashpost_pds_dev`, AppView uses `hashpost_appview_dev`
- **Event-Driven Communication**: PDS publishes events → NATS JetStream → AppView consumes (limited current usage)
- **Reference**: See `docs/designs/hashpost-architecture.md` for detailed architecture

### Database Access Patterns
- **sqlc Only**: Never write raw SQL queries - always use sqlc for type-safe database access
- **Generated Code**: Use `task generate:sqlc` when modifying database schemas or queries
- **Two Databases**: PDS uses `hashpost_pds_dev`, AppView uses `hashpost_appview_dev`
- **Data Flow**: PDS stores canonical records including forum tables → AppView stores denormalized data via direct access and limited events
- **Reference**: See `docs/development/sqlc-patterns.md` for detailed patterns

## Development Workflows

### Code Generation
- **Database Changes**: Run `task generate:sqlc` after modifying SQL queries or schema
- **API Changes**: Run `task generate:openapi` after modifying OpenAPI specifications
- **Mock Generation**: Use `make generate` for gomock mocks, never write hand-written mocks
- **Reference**: See `docs/development/command-structure.md` for complete workflow

### Testing Strategy
- **Unit Tests**: Use mocks from `internal/testutil/` for isolated testing
- **Integration Tests**: Use real database with proper setup/teardown
- **Mock Management**: Use gomock for all mocks, avoid hand-written mock implementations
- **Test Fixtures**: Use `internal/testutil/fixtures.go` for shared test data
- **Reference**: See `.cursor/rules/test-files.mdc` for testing patterns

## Project-Specific Patterns

### Authentication Flow
- **PDS**: Handles DID-based authentication using Bluesky Indigo libraries
- **AppView**: Validates tokens and manages user sessions
- **Middleware Chain**: Authentication → Permission checks → Handler execution
- **Token Flow**: PDS creates sessions → AppView validates tokens for API access

### RBAC System
- **Role-Based Access**: Use `internal/testutil/` for RBAC testing utilities
- **Permission Checks**: Middleware validates permissions before handler execution
- **Database Queries**: All RBAC queries use sqlc-generated code
- **Testing**: Use fixtures for role and permission test data

### Handler Patterns
- **AppView Handlers**: Query AppView database for denormalized data, proxy auth to PDS
- **PDS Handlers**: Direct database access using sqlc queries for canonical records
- **Error Handling**: Use structured logging with slog, wrap errors with context
- **Response Format**: Consistent JSON responses with proper HTTP status codes

### Error Handling Standards
- **Structured Logging**: Use slog for all logging, never printf debugging
- **Error Wrapping**: Use `fmt.Errorf("context: %w", err)` for error context
- **Never Ignore**: Always handle errors explicitly, never use `_` to ignore
- **Logging Levels**: Use appropriate levels (Debug, Info, Warn, Error)

## Common Pitfalls to Avoid

### Database Access
- ❌ **Never use raw SQL** - Always use sqlc-generated queries
- ❌ **Don't access PDS database from AppView** - Use event processing for data sync
- ❌ **Don't ignore errors** - Always handle errors explicitly

### Development Workflow
- ❌ **Don't start/stop services** - Environment runs automatically
- ❌ **Don't create hand-written mocks** - Use gomock instead
- ❌ **Don't modify generated code** - Regenerate when schemas change

### Testing
- ❌ **Don't rely on live database** - Use mocks for unit tests
- ❌ **Don't create random test data** - Use fixtures from `internal/testutil/`
- ❌ **Don't skip error path testing** - Test both success and failure cases

## Cross-References

### File-Specific Rules
- **Go Files**: See `.cursor/rules/go-files.mdc` for Go coding standards
- **Test Files**: See `.cursor/rules/test-files.mdc` for testing patterns

### Detailed Documentation
- **Architecture**: See `docs/designs/` for architecture decisions
- **Development**: See `docs/development/` for detailed workflows
- **Implementation**: See `docs/pds-server-implementation.md` for PDS details

## Development Environment Guidelines

### HashPost Development Workflow

**⚠️ CRITICAL**: HashPost runs in Docker Compose for development. The development environment is typically already running. Do not attempt to run the applications directly on your host machine.

**🚫 DO NOT RUN**: `task dev`, `task dev:down`, `go run ./cmd/appview`, `go run ./cmd/pds`, or any commands that would start/stop/restart services unless you are specifically setting up the environment for the first time.

#### When the Environment is Already Running

If the development environment is already running (which is typical), you should:

1. **Make Code Changes**: Edit files in your IDE - changes are automatically reloaded by Air
2. **View Logs**: Use `task dev:logs` to monitor application behavior (read-only)
3. **Test Changes**: The applications will automatically restart when you save changes
4. **Generate Code**: Use `task generate:sqlc` or `task generate:openapi` when you modify schemas
5. **Run Tests**: Use `task test:unit` or `task lint` for testing (these don't interfere with running services)

#### Available Services When Running

- **Database**: PostgreSQL running on `localhost:5432`
- **PDS Server**: HashPost PDS on `localhost:8080` 
- **AppView Server**: HashPost AppView on `localhost:8081`
- **Frontend**: Next.js UI on `localhost:3000`
- **Swagger UI**: API documentation on `localhost:8081/docs`

#### Safe Commands for AI Assistants

- ✅ `task generate:sqlc` - Regenerate database code
- ✅ `task generate:openapi` - Regenerate API client
- ✅ `task test:unit` - Run unit tests
- ✅ `task test:integration` - Run integration tests
- ✅ `task lint` - Run linter
- ✅ `task dev:logs` - View logs (read-only)
- ✅ Edit files in IDE (Air handles auto-reload)

#### Commands to Avoid

- ❌ `task dev` - Would restart services
- ❌ `task dev:down` - Would stop services  
- ❌ `go run ./cmd/appview` - Wrong environment
- ❌ `go run ./cmd/pds` - Wrong environment
- ❌ Any commands that start/stop/restart services

### Key Understanding

- Applications expect to connect to `postgres:5432` (Docker service name), not `localhost:5432`
- Docker Compose provides proper environment variables
- Services can find each other using Docker service names
- Air handles automatic rebuilding and restarting in the Docker environment