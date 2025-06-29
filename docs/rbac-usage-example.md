# RBAC Usage Examples

This document shows how to use the Role-Based Access Control (RBAC) system for private subforum access control.

## Overview

The RBAC system provides:
- **PermissionDAO**: Data access layer for permission checking
- **PermissionMiddleware**: HTTP middleware for route protection
- **PermissionChecker**: Helper for checking permissions in handlers

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
    user_id, 
    role, 
    permissions,
    added_by_user_id,
    created_at
) VALUES (
    1,           -- subforum_id
    2,           -- user_id  
    'moderator', -- role
    '["moderate_content", "ban_users"]'::jsonb, -- specific permissions
    1,           -- added_by_user_id
    NOW()
);
```

## Usage Examples

### 1. Using PermissionDAO Directly

```go
package main

import (
    "context"
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
    
    // Check specific capabilities
    canModerate, err := permissionDAO.CanModerateSubforum(ctx, userID, subforumID)
    if err != nil {
        log.Error().Err(err).Msg("Failed to check moderation capability")
        return
    }
    
    if canModerate {
        log.Info().Msg("User can moderate this subforum")
    }
}
```

### 2. Using PermissionMiddleware

```go
package main

import (
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
    
    // Protect moderation routes
    http.HandleFunc("/subforums/moderate", 
        permissionMiddleware.RequireModerationCapability()(
            http.HandlerFunc(handleModeration),
        ),
    )
    
    // Protect ban routes
    http.HandleFunc("/subforums/ban", 
        permissionMiddleware.RequireBanCapability()(
            http.HandlerFunc(handleBanUser),
        ),
    )
}
```

### 3. Using PermissionChecker in Handlers

```go
package handlers

import (
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

## Role Hierarchy

### Platform-Wide Roles
- `platform_admin`: Full system access
- `trust_safety`: Trust and safety operations
- `legal_team`: Legal compliance operations

### Subforum-Specific Roles
- `owner`: Full subforum control
  - Capabilities: `moderate_content`, `ban_users`, `remove_content`, `correlate_fingerprints`, `manage_moderators`, `access_private_subforums`
  
- `moderator`: Standard moderation
  - Capabilities: `moderate_content`, `ban_users`, `remove_content`, `correlate_fingerprints`
  
- `junior_moderator`: Limited moderation
  - Capabilities: `moderate_content`, `remove_content`

## Capabilities

### Content Moderation
- `moderate_content`: Can approve/remove posts and comments
- `remove_content`: Can remove posts and comments
- `ban_users`: Can ban users from subforum

### Administrative
- `manage_moderators`: Can add/remove moderators
- `correlate_fingerprints`: Can perform identity correlation
- `access_private_subforums`: Can access private subforums

### System
- `system_admin`: Full system access
- `cross_platform_access`: Access across multiple platforms

## Best Practices

1. **Always check permissions at the handler level** for fine-grained control
2. **Use middleware for route-level protection** to prevent unauthorized access
3. **Log permission checks** for audit trails
4. **Cache permission results** for performance in high-traffic scenarios
5. **Use specific capabilities** rather than broad roles when possible

## Error Handling

The RBAC system provides detailed error messages and logging:

```go
// Check access with proper error handling
canAccess, err := permissionDAO.CanAccessPrivateSubforum(ctx, userID, subforumID)
if err != nil {
    log.Error().Err(err).
        Int64("user_id", userID).
        Int32("subforum_id", subforumID).
        Msg("Failed to check private subforum access")
    return fmt.Errorf("failed to verify subforum access")
}

if !canAccess {
    log.Warn().
        Int64("user_id", userID).
        Int32("subforum_id", subforumID).
        Msg("User denied access to private subforum")
    return fmt.Errorf("access denied to private subforum")
}
``` 