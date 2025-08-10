# Permission Checking Patterns Guide

## Overview

This guide provides comprehensive patterns for implementing permission checks in HashPost handlers using the unified capability system. All permission checking should use the `PermissionDAO.HasUnifiedCapability()` method.

## Core Pattern

### Standard Permission Check
```go
func (h *Handler) SomeEndpoint(ctx context.Context, input *models.Input) (*models.Response, error) {
    // 1. Extract user context
    userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
    if err != nil {
        return nil, huma.Error401Unauthorized("authentication required")
    }
    
    // 2. Check permissions using unified capability system
    hasCapability, err := h.permissionDAO.HasUnifiedCapability(
        ctx, 
        userCtx.UserID, 
        userCtx.ActivePseudonymID, 
        constants.CapabilityRequired, 
        subforumID) // nil for global, &subforumID for subforum-specific
    if err != nil {
        log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to check capability")
        return nil, fmt.Errorf("failed to check permissions: %w", err)
    }
    if !hasCapability {
        log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks required capability")
        return nil, huma.Error403Forbidden("insufficient permissions")
    }
    
    // 3. Proceed with business logic
    // ...
}
```

## Permission Context Patterns

### Global Permissions (nil context)
Use for platform-wide capabilities that apply across all subforums:

```go
// Content creation (applies everywhere)
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityCreateContent, nil)

// Subforum creation (platform-wide privilege)
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityCreateSubforum, nil)

// System administration
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilitySystemAdmin, nil)
```

### Subforum-Specific Permissions (&subforumID context)
Use for capabilities that are specific to individual subforums:

```go
// Moderation within a specific subforum
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityModerateContent, &subforum.SubforumID)

// Managing subforum settings
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityManageSubforumSettings, &subforum.SubforumID)

// Banning users from a subforum
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityBanUsers, &subforum.SubforumID)
```

## Common Capability Categories

### Content Management
```go
// Basic content operations (global)
constants.CapabilityCreateContent    // Create posts/comments
constants.CapabilityVote            // Vote on content
constants.CapabilityReport          // Report content

// Content moderation (subforum-specific)
constants.CapabilityModerateContent // Moderate content
constants.CapabilityRemoveContent   // Remove content
constants.CapabilityLockContent     // Lock posts/comments
```

### User Management
```go
// Basic user operations (global)
constants.CapabilityMessage         // Send direct messages

// User moderation (subforum-specific)
constants.CapabilityBanUsers        // Ban users from subforum
constants.CapabilityManageUsers     // Manage user accounts
```

### Administrative Operations
```go
// Subforum administration (subforum-specific)
constants.CapabilityManageSubforumSettings  // Manage subforum settings
constants.CapabilityManageModerators        // Manage moderator assignments
constants.CapabilityManageSubforumRules     // Manage subforum rules

// Platform administration (global)
constants.CapabilitySystemAdmin             // Full system access
constants.CapabilityCorrelateIdentities     // Identity correlation
constants.CapabilityForwardReports          // Forward reports to platform
```

## Error Handling Patterns

### Standard Error Handling
```go
hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, capability, subforumID)
if err != nil {
    log.Error().
        Err(err).
        Int64("user_id", userCtx.UserID).
        Str("capability", capability).
        Interface("subforum_id", subforumID).
        Msg("Failed to check capability")
    return nil, fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    log.Warn().
        Int64("user_id", userCtx.UserID).
        Str("capability", capability).
        Interface("subforum_id", subforumID).
        Msg("User lacks required capability")
    return nil, huma.Error403Forbidden("insufficient permissions")
}
```

### Multiple Capability Checks
```go
// Check multiple capabilities (OR logic)
hasModerate, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityModerateContent, nil)
if err != nil {
    return nil, fmt.Errorf("failed to check moderate capability: %w", err)
}

hasSystemAdmin, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilitySystemAdmin, nil)
if err != nil {
    return nil, fmt.Errorf("failed to check system admin capability: %w", err)
}

if !hasModerate && !hasSystemAdmin {
    return nil, huma.Error403Forbidden("insufficient permissions")
}
```

## Logging Patterns

### Comprehensive Logging
```go
log.Info().
    Str("endpoint", "endpoint_name").
    Str("component", "handler").
    Int64("user_id", userCtx.UserID).
    Str("active_pseudonym_id", userCtx.ActivePseudonymID).
    Str("capability", capability).
    Interface("subforum_id", subforumID).
    Msg("Permission check requested")

hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, capability, subforumID)
if err != nil {
    log.Error().
        Err(err).
        Int64("user_id", userCtx.UserID).
        Str("capability", capability).
        Msg("Failed to check capability")
    return nil, fmt.Errorf("failed to check permissions: %w", err)
}

log.Debug().
    Int64("user_id", userCtx.UserID).
    Str("capability", capability).
    Interface("subforum_id", subforumID).
    Bool("has_capability", hasCapability).
    Msg("Permission check completed")
```

## Handler Structure Patterns

### Standard Handler Structure
```go
type HandlerName struct {
    permissionDAO dao.PermissionDAOInterface
    // ... other DAOs
}

func NewHandlerName(permissionDAO dao.PermissionDAOInterface /* other DAOs */) *HandlerName {
    return &HandlerName{
        permissionDAO: permissionDAO,
        // ... other DAOs
    }
}

func (h *HandlerName) EndpointMethod(ctx context.Context, input *models.InputType) (*models.ResponseType, error) {
    // 1. Extract user context
    userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
    if err != nil {
        return nil, huma.Error401Unauthorized("authentication required")
    }
    
    // 2. Permission check
    hasCapability, err := h.permissionDAO.HasUnifiedCapability(
        ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
        constants.CapabilityRequired, subforumContext)
    if err != nil {
        return nil, fmt.Errorf("failed to check permissions: %w", err)
    }
    if !hasCapability {
        return nil, huma.Error403Forbidden("insufficient permissions")
    }
    
    // 3. Business logic
    // ...
    
    return response, nil
}
```

## Advanced Patterns

### Getting Complete Capability Set
```go
// Get all capabilities for decision making
roles, capabilities, err := h.permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, subforumID)
if err != nil {
    return nil, fmt.Errorf("failed to get capabilities: %w", err)
}

// Check multiple capabilities from the set
hasModerate := false
hasSystemAdmin := false
for _, cap := range capabilities {
    switch cap {
    case constants.CapabilityModerateContent:
        hasModerate = true
    case constants.CapabilitySystemAdmin:
        hasSystemAdmin = true
    }
}
```

### Conditional Permission Checks
```go
// Different permission requirements based on context
var requiredCapability string
var subforumContext *int32

if isGlobalOperation {
    requiredCapability = constants.CapabilitySystemAdmin
    subforumContext = nil
} else {
    requiredCapability = constants.CapabilityModerateContent
    subforumContext = &subforum.SubforumID
}

hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    requiredCapability, subforumContext)
```

### Resource-Specific Checks
```go
// Check permission for specific resource
func (h *Handler) validateResourceAccess(ctx context.Context, userCtx *middleware.UserContext, resourceID int64) error {
    // Get resource to determine subforum context
    resource, err := h.resourceDAO.GetByID(ctx, resourceID)
    if err != nil {
        return fmt.Errorf("failed to get resource: %w", err)
    }
    
    // Check permission in resource's subforum context
    hasCapability, err := h.permissionDAO.HasUnifiedCapability(
        ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
        constants.CapabilityModerateContent, &resource.SubforumID)
    if err != nil {
        return fmt.Errorf("failed to check permissions: %w", err)
    }
    if !hasCapability {
        return huma.Error403Forbidden("insufficient permissions for this resource")
    }
    
    return nil
}
```

## Testing Patterns

### Unit Test Pattern
```go
func TestHandler_EndpointMethod_Success(t *testing.T) {
    // Setup mocks
    mockPermissionDAO := &mocks.PermissionDAOInterface{}
    handler := NewHandler(mockPermissionDAO)
    
    // Mock permission check to return true
    mockPermissionDAO.On("HasUnifiedCapability", 
        mock.Anything, int64(1), "pseudonym-1", 
        constants.CapabilityRequired, (*int32)(nil)).
        Return(true, nil)
    
    // Test the handler
    input := &models.Input{
        AuthInput: middleware.AuthInput{/* auth data */},
        // ... other input fields
    }
    
    response, err := handler.EndpointMethod(context.Background(), input)
    
    assert.NoError(t, err)
    assert.NotNil(t, response)
    mockPermissionDAO.AssertExpectations(t)
}

func TestHandler_EndpointMethod_InsufficientPermissions(t *testing.T) {
    // Setup mocks
    mockPermissionDAO := &mocks.PermissionDAOInterface{}
    handler := NewHandler(mockPermissionDAO)
    
    // Mock permission check to return false
    mockPermissionDAO.On("HasUnifiedCapability", 
        mock.Anything, int64(1), "pseudonym-1", 
        constants.CapabilityRequired, (*int32)(nil)).
        Return(false, nil)
    
    // Test the handler
    input := &models.Input{
        AuthInput: middleware.AuthInput{/* auth data */},
        // ... other input fields
    }
    
    response, err := handler.EndpointMethod(context.Background(), input)
    
    assert.Error(t, err)
    assert.Nil(t, response)
    assert.Contains(t, err.Error(), "insufficient permissions")
    mockPermissionDAO.AssertExpectations(t)
}
```

## Anti-Patterns

### ❌ Don't Use Deprecated Methods
```go
// DON'T DO THIS - deprecated JWT-based checking
if !userCtx.HasCapability(constants.CapabilityCreateContent) {
    return huma.Error403Forbidden("insufficient permissions")
}
```

### ❌ Don't Use String Literals
```go
// DON'T DO THIS - use constants
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    "moderate_content", nil) // BAD: string literal
```

### ❌ Don't Skip Error Handling
```go
// DON'T DO THIS - always handle errors
hasCapability, _ := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityModerateContent, nil) // BAD: ignoring errors
```

### ❌ Don't Use Uninitialized Pointers
```go
// DON'T DO THIS - use nil or properly initialized pointer
var subforumID *int32 // uninitialized pointer
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityModerateContent, subforumID) // BAD
```

## Best Practices Summary

1. **Always use the unified capability system** via `HasUnifiedCapability()`
2. **Use constants** for capability names from `constants` package
3. **Handle all errors** from permission checks
4. **Log permission checks** for debugging and auditing
5. **Use appropriate context** (nil for global, &subforumID for subforum-specific)
6. **Follow consistent error handling** patterns
7. **Write comprehensive tests** for permission scenarios
8. **Extract user context** before permission checks
9. **Use structured logging** with relevant fields
10. **Return appropriate HTTP status codes** (401 for auth, 403 for permissions)

## Migration Checklist

When updating existing handlers to use the unified capability system:

- [ ] Replace `userCtx.HasCapability()` calls with `permissionDAO.HasUnifiedCapability()`
- [ ] Add `permissionDAO` dependency to handler struct
- [ ] Update constructor to accept `permissionDAO` parameter
- [ ] Add proper error handling for permission checks
- [ ] Add logging for permission checks
- [ ] Use constants instead of string literals
- [ ] Update tests to mock `PermissionDAO` instead of user context capabilities
- [ ] Verify correct context usage (nil vs &subforumID)
- [ ] Test both success and failure scenarios
- [ ] Update integration tests if needed
