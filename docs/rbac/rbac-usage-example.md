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
pseudonyms, err := securePseudonymDAO.GetPseudonymsByUserID(
    ctx, userID, role, constants.ScopeAuthentication)

// Verify ownership using self-correlation scope
ownsPseudonym, err := securePseudonymDAO.VerifyPseudonymOwnership(
    ctx, pseudonymID, userID, constants.RoleUser, constants.ScopeSelfCorrelation)
```

### Using Capability Constants

```go
// Check moderation capability
canModerate, err := permissionDAO.HasSubforumCapability(
    ctx, userID, subforumID, constants.CapabilityModerateContent)

// Check ban capability
canBan, err := permissionDAO.HasSubforumCapability(
    ctx, userID, subforumID, constants.CapabilityBanUsers)
```

## Database Setup

### 1. User Roles and Capabilities

Users can have platform-wide roles and capabilities stored in JSON fields:

```sql
-- Example user with platform admin role
UPDATE users 
SET roles = '["platform_admin"]'::jsonb,
    capabilities = '["access_private_subforums", "system_admin"]'::jsonb
WHERE user_id = 1;
```

### 2. Subforum Moderators

Subforum-specific permissions are managed through the `subforum_moderators` table:

```sql
-- Add a moderator to a subforum
INSERT INTO subforum_moderators (
    subforum_id, 
    pseudonym_id, 
    role, 
    permissions,
    added_by_pseudonym_id,
    created_at
) VALUES (
    1,           -- subforum_id
    'pseudonym_123', -- pseudonym_id  
    'moderator', -- role
    '["moderate_content", "ban_users"]'::jsonb, -- specific permissions
    'admin_pseudonym', -- added_by_pseudonym_id
    NOW()
);
```

## Usage Examples

### 1. Using PermissionDAO with Constants

```go
package main

import (
    "context"
    "github.com/matt0x6f/hashpost/internal/api/constants"
    "github.com/matt0x6f/hashpost/internal/database/dao"
)

func checkUserAccess(db bob.Executor, userID int64, subforumID int32) {
    permissionDAO := dao.NewPermissionDAO(db)
    
    // Check if user can access private subforum
    canAccess, err := permissionDAO.CanAccessPrivateSubforum(ctx, userID, subforumID)
    if err != nil {
        log.Error().Err(err).Msg("Failed to check access")
        return
    }
    
    if !canAccess {
        log.Warn().Msg("Access denied to private subforum")
        return
    }
    
    // Check specific capabilities using constants
    canModerate, err := permissionDAO.HasSubforumCapability(
        ctx, userID, subforumID, constants.CapabilityModerateContent)
    if err != nil {
        log.Error().Err(err).Msg("Failed to check moderation capability")
        return
    }
    
    if canModerate {
        log.Info().Msg("User can moderate this subforum")
    }
}
```

### 2. Using PermissionMiddleware with Constants

```go
package main

import (
    "github.com/matt0x6f/hashpost/internal/api/constants"
    "github.com/matt0x6f/hashpost/internal/api/middleware"
)

func setupRoutes(db bob.Executor) {
    permissionMiddleware := middleware.NewPermissionMiddleware(db)
    
    // Protect routes that require private subforum access
    http.HandleFunc("/subforums/private", 
        permissionMiddleware.RequirePrivateSubforumAccess()(
            http.HandlerFunc(handlePrivateSubforum),
        ),
    )
    
    // Protect moderation routes using constants
    http.HandleFunc("/subforums/moderate", 
        permissionMiddleware.RequireSubforumCapability(constants.CapabilityModerateContent)(
            http.HandlerFunc(handleModeration),
        ),
    )
    
    // Protect ban routes using constants
    http.HandleFunc("/subforums/ban", 
        permissionMiddleware.RequireSubforumCapability(constants.CapabilityBanUsers)(
            http.HandlerFunc(handleBanUser),
        ),
    )
}
```

### 3. Using PermissionChecker in Handlers with Constants

```go
package handlers

import (
    "github.com/matt0x6f/hashpost/internal/api/constants"
    "github.com/matt0x6f/hashpost/internal/api/middleware"
)

type ContentHandler struct {
    permissionChecker *middleware.PermissionChecker
    // ... other fields
}

func (h *ContentHandler) GetPosts(ctx context.Context, input *models.PostListInput) (*models.PostListResponse, error) {
    // Get user context
    userCtx, err := middleware.ExtractUserFromContext(ctx)
    if err != nil {
        return nil, fmt.Errorf("authentication required")
    }
    
    // Get subforum
    subforum, err := h.subforumDAO.GetSubforumByName(ctx, input.SubforumName)
    if err != nil {
        return nil, err
    }
    
    // Check private subforum access
    if subforum.IsPrivate.Valid && subforum.IsPrivate.V {
        canAccess, err := h.permissionChecker.CheckPrivateSubforumAccess(
            ctx, userCtx.UserID, subforum.SubforumID,
        )
        if err != nil {
            return nil, fmt.Errorf("failed to verify access")
        }
        
        if !canAccess {
            return nil, fmt.Errorf("access denied to private subforum")
        }
    }
    
    // Continue with normal post retrieval...
    return h.getPostsFromSubforum(ctx, subforum.SubforumID)
}
```

### 4. Creating Role Keys with Constants

```go
package commands

import (
    "github.com/matt0x6f/hashpost/internal/api/constants"
)

func createRoleKeys() error {
    // Get all role definitions using constants
    roleDefinitions := constants.GetRoleDefinitions()
    
    for _, roleDef := range roleDefinitions {
        log.Info().Str("role", roleDef.RoleName).Msg("Creating role keys")
        
        for _, scope := range roleDef.Scopes {
            capabilities := roleDef.Capabilities[scope]
            
            // Create role key with proper scope and capabilities
            keyData := ibeSystem.GenerateTestRoleKey(roleDef.RoleName, scope)
            
            _, err := roleKeyDAO.CreateRoleKey(
                ctx, 
                roleDef.RoleName, 
                scope, 
                keyData, 
                capabilities, 
                expiresAt, 
                creatorUserID,
            )
            if err != nil {
                log.Error().Err(err).Msg("Failed to create role key")
                continue
            }
            
            log.Info().
                Str("role", roleDef.RoleName).
                Str("scope", scope).
                Strs("capabilities", capabilities).
                Msg("Role key created successfully")
        }
    }
    
    return nil
}
```

## Role Hierarchy

### Platform-Wide Roles (using constants)
- `constants.RolePlatformAdmin`: Full system access
- `constants.RoleTrustSafety`: Trust and safety operations
- `constants.RoleLegalTeam`: Legal compliance operations

### Subforum-Specific Roles
- `constants.RoleSubforumOwner`: Full subforum control
  - Capabilities: `moderate_content`, `ban_users`, `remove_content`, `correlate_fingerprints`, `manage_moderators`, `access_private_subforums`
  
- `constants.RoleModerator`: Standard moderation
  - Capabilities: `moderate_content`, `ban_users`, `remove_content`, `correlate_fingerprints`
  
- `constants.RoleUser`: Basic user
  - Capabilities: `create_content`, `vote`, `message`, `report`

## Capabilities by Category

### Content Moderation
- `constants.CapabilityModerateContent`: Can approve/remove posts and comments
- `constants.CapabilityRemoveContent`: Can remove posts and comments
- `constants.CapabilityBanUsers`: Can ban users from subforum

### Administrative
- `constants.CapabilityManageModerators`: Can add/remove moderators
- `constants.CapabilityCorrelateFingerprints`: Can perform identity correlation
- `constants.CapabilityAccessSubforumPseudonyms`: Can access pseudonyms within subforum

### System
- `constants.CapabilitySystemAdmin`: Full system access
- `constants.CapabilityUserManagement`: Manage user accounts
- `constants.CapabilityCompliance`: Compliance-related operations

## Best Practices

### 1. Always Use Constants
```go
// ✅ Good - Using constants
canModerate, err := permissionDAO.HasSubforumCapability(
    ctx, userID, subforumID, constants.CapabilityModerateContent)

// ❌ Bad - Using string literals
canModerate, err := permissionDAO.HasSubforumCapability(
    ctx, userID, subforumID, "moderate_content")
```

### 2. Validate Roles and Capabilities
```go
// Check if role is valid before using
if !constants.IsValidRole(roleName) {
    return fmt.Errorf("invalid role: %s", roleName)
}

// Check if capability is valid for scope
if !constants.IsValidCapability(capability, scope) {
    return fmt.Errorf("invalid capability %s for scope %s", capability, scope)
}
```

### 3. Use Role Definitions
```go
// Get role definition for comprehensive information
roleDef := constants.GetRoleDefinition(constants.RoleModerator)
if roleDef != nil {
    // Use roleDef.Scopes and roleDef.Capabilities
    for _, scope := range roleDef.Scopes {
        capabilities := roleDef.Capabilities[scope]
        // Process capabilities for this scope
    }
}
```

## Related Documentation

- [RBAC Overview](rbac-overview.md) - Complete RBAC system documentation
- [Role Keys and Site Roles](role-keys-and-site-roles.md) - Role key management
- [User Roles](user-roles.md) - Role definitions and permissions 