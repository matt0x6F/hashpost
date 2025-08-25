# HashPost Cursor Rules

## Development Environment

### Always Running Services
- **Backend**: Always running with hot reload via `make dev`
- **Frontend**: Always running with hot reload via `cd ui && npm run dev`
- **Database**: Always running via Docker Compose
- **Never ask to start services** - they're always on and have hot reloading enabled

### Hot Reloading
- Backend automatically restarts on Go file changes
- Frontend automatically reloads on TypeScript/React changes
- No manual restarts or starts needed for development

## Unified Permission System

### Single Permission System
HashPost uses a **unified capability system** that leverages the `PermissionDAO.HasUnifiedCapability()` method for ALL permission checking. This system combines:

- **Global capabilities** from role keys with `subforum_id = NULL`
- **Subforum-specific capabilities** from role keys with specific `subforum_id`
- **Automatic role assignment** (e.g., "moderator" role when subforum capabilities are present)

### Standard Permission Check Pattern

```go
// Standard permission check pattern for ALL endpoints
func (h *Handler) SomeEndpoint(ctx context.Context, input *models.Input) (*models.Response, error) {
    // 1. Extract user context
    userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
    if err != nil {
        return nil, huma.Error401Unauthorized("authentication required")
    }
    
    // 2. Check permissions using unified capability system
    hasCapability, err := h.permissionDAO.HasUnifiedCapability(
        ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
        constants.CapabilityRequired, subforumID) // nil for global, &subforumID for subforum-specific
    if err != nil {
        log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to check capability")
        return nil, fmt.Errorf("failed to check permissions: %w", err)
    }
    if !hasCapability {
        log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks required capability")
        return nil, huma.Error403Forbidden("insufficient permissions")
    }
    
    // 3. Proceed with business logic
}
```

### Permission Context Usage

#### Global Permissions (use `nil`)
- Content creation (`create_content`)
- Voting (`vote`) 
- User management (`user_management`)
- System administration (`system_admin`)
- Platform-wide actions

```go
// Global capability check
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityCreateContent, nil) // nil = global scope
```

#### Subforum-Specific Permissions (use `&subforumID`)
- Subforum moderation (`moderate_content`, `ban_users`)
- Subforum settings (`manage_subforum_settings`)
- Subforum rules (`manage_subforum_rules`)
- Subforum-specific actions

```go
// Subforum-specific capability check
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityModerateContent, &subforum.SubforumID) // specific subforum
```

### Migration Status
- ✅ **Current**: All handlers use unified capability system
- ❌ **Deprecated**: `userCtx.HasCapability()` (JWT token cached capabilities)
- ❌ **Deprecated**: Direct database queries for permissions

## Handler Patterns

When making changing to the API follow these steps:
1. Make changes
2. Run `make ui-generate-api`

This updates all frontend SDK models and methods.

### Handler Structure
All handlers follow this pattern:
```go
type HandlerName struct {
    // DAO dependencies
    someDAO dao.SomeDAOInterface
    // ... other dependencies
}

func NewHandlerName(dependencies...) *HandlerName {
    return &HandlerName{
        // Initialize dependencies
    }
}

func (h *HandlerName) MethodName(ctx context.Context, input *models.InputType) (*models.ResponseType, error) {
    // 1. Extract user context
    userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
    if err != nil {
        return nil, huma.Error401Unauthorized("authentication required")
    }
    
    // 2. Permission check (global or unified)
    // 3. Business logic
    // 4. Return response
}
```

### Input/Output Models
- Use Huma struct tags for parameters: `path:"param"`, `query:"param"`, `header:"Header"`
- Embed `middleware.AuthInput` for authenticated endpoints
- Use `Body` field for request/response bodies

## Database Operations

### Migrations
```bash
make migrate-create name=descriptive_name
# Edit migration file in internal/database/migrations/
make migrate-up
make migrate-status
make generate  # Regenerate Bob models
```

### Bob Model Generation
- Always run `make generate` after migrations
- Models are in `internal/database/models/`
- DAOs are in `internal/database/dao/`

## Testing Patterns

### Unit Tests
- Use table-driven tests for comprehensive coverage
- Mock DAO interfaces for isolation
- Use fixtures for repeatable, stable information
- Test both success and error cases

Run `make test` to run the unit tests

### Integration Tests
- Run individual test suites: `make test-integration-local TESTS=./path/to/test.go`
- Each test creates its own `IntegrationTestSuite` with `defer suite.Cleanup()`
- This ensures isolation and prevents data collisions

### Test Structure
```go
func TestHandler_Method(t *testing.T) {
    tests := []struct {
        name           string
        input          *models.Input
        wantErr        bool
        expectedStatus int
    }{
        // Test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Common Commands

### Database
- `make migrate-up` - Apply pending migrations
- `make migrate-down` - Rollback last migration
- `make migrate-status` - Show migration status
- `make migrate-create name=migration_name` - Create new migration

### Testing
- `make test` - Run unit tests
- `make test-integration-local TESTS=./path/to/test.go` - Run specific integration tests

### Code Quality
- `make fmt` - Format code
- `make lint` - Lint code

## Database Debugging

### Connect to Database
```bash
docker-compose exec postgres psql -U hashpost -d hashpost
```

### Common Queries
```sql
-- Check user roles
SELECT * FROM subforum_moderators WHERE subforum_id = (SELECT subforum_id FROM subforums WHERE name = 'subforum_name');

-- Check user capabilities
SELECT * FROM users WHERE email = 'user@example.com';

-- Check role assignments
SELECT sm.role, sm.permissions, s.name as subforum_name 
FROM subforum_moderators sm 
JOIN subforums s ON sm.subforum_id = s.subforum_id;
```

## Error Handling

### Structured Error Wrapping
```go
if err != nil {
    return fmt.Errorf("failed to check permissions: %w", err)
}
```

### HTTP Error Responses
```go
return nil, huma.Error403Forbidden("insufficient permissions")
return nil, huma.Error404NotFound("subforum not found")
return nil, huma.Error401Unauthorized("authentication required")
```

### Logging
```go
log.Info().Str("endpoint", "subforums").Msg("Get subforums requested")
log.Error().Err(err).Msg("Failed to get subforums from database")
log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks capability")
```

## Role System

### Role Constants
- `RoleUser` = "user"
- `RoleModerator` = "moderator" 
- `RoleSubforumOwner` = "subforum_owner"
- `RolePlatformAdmin` = "platform_admin"

### Role Capabilities
- Moderator: `moderate_content`, `ban_users`, `remove_content`, `review_reports`, `manage_subforum_settings`
- SubforumOwner: All moderator capabilities + `manage_moderators`

### Database Role Names
- Use exact role constants in database: `"subforum_owner"` not `"owner"`
- Role names must match constants exactly for permission checking to work

## UI Development

### Technology Stack
- **Next.js** with App Router
- **TypeScript** for all components
- **ShadCN/UI** for components
- **Tailwind CSS** for styling

### UI Commands
- `cd ui && npm run dev` - Start UI development server
- `cd ui && npm run build` - Build UI
- `cd ui && npm run lint` - Lint UI
- `cd ui && npm run type-check` - Type check

### API Integration
- Use generated API client in `ui/lib/api-client.ts`
- Handle authentication via cookies
- Use optimistic updates where appropriate 