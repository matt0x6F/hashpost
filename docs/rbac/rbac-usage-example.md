# RBAC Usage Examples

This document shows how to use the Role-Based Access Control (RBAC) system with the centralized constants for consistent access control.

## Overview

The RBAC system provides:
- **PermissionDAO**: Data access layer for permission checking
- **PermissionMiddleware**: HTTP middleware for route protection
- **PermissionChecker**: Helper for checking permissions in handlers
- **Constants**: Centralized role, scope, and capability definitions

## Constants Usage

### Importing Constants

```go
import "github.com/matt0x6f/hashpost/internal/api/constants"
```

### Using Role Constants

```go
// Define roles using constants
roles := []string{constants.RoleUser, constants.RoleModerator}

// Check if role is valid
if constants.IsValidRole("platform_admin") {
    // Role is valid
}

// Get role capabilities
capabilities := constants.GetRoleCapabilities(constants.RoleModerator)
```

### Using Scope Constants

```go
// Use scope constants for IBE operations
pseudonyms, err := pseudonymDAO.GetPseudonymsByUserID(
    ctx, userID, role, constants.ScopeAuthentication)

// Verify ownership using self-correlation scope
ownsPseudonym, err := pseudonymDAO.VerifyPseudonymOwnership(
    ctx, pseudonymID, userID, constants.RoleUser, constants.ScopeSelfCorrelation)
```

### Using Capability Constants

```go
// Check moderation capability
canModerate, err := permissionDAO.HasUnifiedCapability(
    ctx, userID, activePseudonymID, constants.CapabilityModerateContent, &subforumID)

// Check ban capability
canBan, err := permissionDAO.HasUnifiedCapability(
    ctx, userID, activePseudonymID, constants.CapabilityBanUsers, &subforumID)
```

## Database Setup

### 1. Role Keys (Primary Permission Storage)

Role keys store all permissions in the `role_keys` table:

```sql
-- Example: Global user capabilities (no subforum context)
INSERT INTO role_keys (
    key_id,
    role_name,
    scope,
    key_data,
    key_version,
    capabilities,
    created_at,
    expires_at,
    is_active,
    created_by,
    pseudonym_id,
    subforum_id
) VALUES (
    gen_random_uuid(),
    'user',
    'authentication',
    E'\\x...', -- Encrypted IBE key data
    1,
    '["create_content", "vote", "message", "report"]'::jsonb,
    NOW(),
    NOW() + INTERVAL '1 year',
    TRUE,
    'pseudonym_123',
    'pseudonym_123',
    NULL  -- Global key
);

-- Example: Subforum-specific moderator capabilities
INSERT INTO role_keys (
    key_id,
    role_name,
    scope,
    key_data,
    key_version,
    capabilities,
    created_at,
    expires_at,
    is_active,
    created_by,
    pseudonym_id,
    subforum_id
) VALUES (
    gen_random_uuid(),
    'moderator',
    'correlation',
    E'\\x...', -- Encrypted IBE key data
    1,
    '["moderate_content", "ban_users", "remove_content"]'::jsonb,
    NOW(),
    NOW() + INTERVAL '1 year',
    TRUE,
    'admin_pseudonym_456',
    'moderator_pseudonym_789',
    123  -- Specific subforum ID
);
```

### 2. User Accounts

Users have no roles or capabilities:

```sql
-- Example user account (no roles or capabilities)
UPDATE users 
SET email = 'user@example.com',
    is_active = TRUE
WHERE user_id = 1;
```

### 3. Pseudonyms (No Direct Roles/Capabilities)

Pseudonyms do not store roles or capabilities directly:

```sql
-- Example pseudonym (no roles or capabilities columns)
UPDATE pseudonyms 
SET display_name = 'UserPseudonym',
    is_active = TRUE
WHERE pseudonym_id = 'pseudonym_123';
```

## Permission Checking Examples

### 1. Global Permission Check

```go
// Check if user can create content (global capability)
hasCapability, err := permissionDAO.HasUnifiedCapability(
    ctx, userID, activePseudonymID, constants.CapabilityCreateContent, nil)
if err != nil {
    return fmt.Errorf("failed to check permission: %w", err)
}

if !hasCapability {
    return huma.Error403Forbidden("insufficient permissions")
}
```

### 2. Subforum-Specific Permission Check

```go
// Check if user can moderate content in specific subforum
subforumID := int32(123)
hasCapability, err := permissionDAO.HasUnifiedCapability(
    ctx, userID, activePseudonymID, constants.CapabilityModerateContent, &subforumID)
if err != nil {
    return fmt.Errorf("failed to check permission: %w", err)
}

if !hasCapability {
    return huma.Error403Forbidden("insufficient permissions")
}
```

### 3. Getting Complete Permission Set

```go
// Get all capabilities for user in subforum context
roles, capabilities, err := permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
    ctx, userID, activePseudonymID, &subforumID)
if err != nil {
    return fmt.Errorf("failed to get permissions: %w", err)
}

// Check if user has specific capability
hasCapability := slices.Contains(capabilities, constants.CapabilityModerateContent)
```

## Handler Implementation Examples

### 1. Content Creation Handler

```go
func (h *ContentHandler) CreatePost(ctx context.Context, input *models.CreatePostInput) (*models.CreatePostResponse, error) {
    // Extract user context
    userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
    if err != nil {
        return nil, huma.Error401Unauthorized("authentication required")
    }
    
    // Check global permission
    hasCapability, err := h.permissionDAO.HasUnifiedCapability(
        ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityCreateContent, nil)
    if err != nil {
        return nil, fmt.Errorf("failed to check permissions: %w", err)
    }
    
    if !hasCapability {
        return nil, huma.Error403Forbidden("insufficient permissions")
    }
    
    // Create post logic...
    return &models.CreatePostResponse{}, nil
}
```

### 2. Moderation Handler

```go
func (h *ModerationHandler) RemovePost(ctx context.Context, input *models.RemovePostInput) (*models.RemovePostResponse, error) {
    // Extract user context
    userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
    if err != nil {
        return nil, huma.Error401Unauthorized("authentication required")
    }
    
    // Get subforum ID from post
    post, err := h.postDAO.GetPostByID(ctx, input.PostID)
    if err != nil {
        return nil, fmt.Errorf("failed to get post: %w", err)
    }
    
    // Check subforum-specific permission
    hasCapability, err := h.permissionDAO.HasUnifiedCapability(
        ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityRemoveContent, &post.SubforumID)
    if err != nil {
        return nil, fmt.Errorf("failed to check permissions: %w", err)
    }
    
    if !hasCapability {
        return nil, huma.Error403Forbidden("insufficient permissions")
    }
    
    // Remove post logic...
    return &models.RemovePostResponse{}, nil
}
```

## Middleware Usage

### 1. Route Protection

```go
// Protect route with global capability
app.Get("/api/posts", middleware.RequireCapability(constants.CapabilityCreateContent))

// Protect route with subforum-specific capability
app.Get("/api/subforums/{subforum}/moderation", middleware.RequireSubforumCapability(constants.CapabilityModerateContent))
```

### 2. Custom Middleware

```go
func RequireModerationCapability(capability string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract user context
            userCtx, err := middleware.ExtractUserFromContext(r.Context())
            if err != nil {
                http.Error(w, "authentication required", http.StatusUnauthorized)
                return
            }
            
            // Get subforum ID from URL
            subforumID := extractSubforumIDFromURL(r.URL.Path)
            
            // Check capability
            hasCapability, err := permissionDAO.HasUnifiedCapability(
                r.Context(), userCtx.UserID, userCtx.ActivePseudonymID, capability, &subforumID)
            if err != nil || !hasCapability {
                http.Error(w, "insufficient permissions", http.StatusForbidden)
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}
```

## Testing Examples

### 1. Unit Test

```go
func TestCreatePostHandler(t *testing.T) {
    tests := []struct {
        name           string
        userCtx        *middleware.UserContext
        hasCapability  bool
        wantErr        bool
        expectedStatus int
    }{
        {
            name: "user with create_content capability",
            userCtx: &middleware.UserContext{
                UserID: 1,
                ActivePseudonymID: "pseudonym_123",
            },
            hasCapability: true,
            wantErr: false,
        },
        {
            name: "user without create_content capability",
            userCtx: &middleware.UserContext{
                UserID: 2,
                ActivePseudonymID: "pseudonym_456",
            },
            hasCapability: false,
            wantErr: true,
            expectedStatus: 403,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Mock permission DAO
            mockPermissionDAO := &mocks.MockPermissionDAO{}
            mockPermissionDAO.On("HasUnifiedCapability", 
                mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, 
                constants.CapabilityCreateContent, (*int32)(nil)).
                Return(tt.hasCapability, nil)
            
            // Test handler...
        })
    }
}
```

### 2. Integration Test

```go
func TestModerationPermissions(t *testing.T) {
    suite := NewIntegrationTestSuite(t)
    defer suite.Cleanup()
    
    // Create test data
    user := suite.CreateTestUser("test@example.com")
    pseudonym := suite.CreateTestPseudonym(user.UserID, "TestUser")
    subforum := suite.CreateTestSubforum("test-subforum")
    
    // Create role key for moderation
    roleKey := &models.RoleKey{
        RoleName: "moderator",
        Scope: "correlation",
        Capabilities: types.JSON[json.RawMessage](`["moderate_content", "ban_users"]`),
        PseudonymID: pseudonym.PseudonymID,
        SubforumID: sql.Null[int32]{Int32: subforum.SubforumID, Valid: true},
        IsActive: sql.Null[bool]{Bool: true, Valid: true},
    }
    suite.CreateRoleKey(roleKey)
    
    // Test moderation capability
    hasCapability, err := suite.permissionDAO.HasUnifiedCapability(
        context.Background(), user.UserID, pseudonym.PseudonymID, 
        constants.CapabilityModerateContent, &subforum.SubforumID)
    
    assert.NoError(t, err)
    assert.True(t, hasCapability)
}
```

## Best Practices

1. **Always use constants** for roles, scopes, and capabilities
2. **Check permissions early** in handlers
3. **Use appropriate error handling** for permission failures
4. **Test permission scenarios** thoroughly
5. **Log permission checks** for audit purposes
6. **Use unified permission methods** for consistency 