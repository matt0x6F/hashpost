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
canModerate, err := permissionDAO.HasSubforumCapability(
    ctx, userID, subforumID, constants.CapabilityModerateContent)

// Check ban capability
canBan, err := permissionDAO.HasSubforumCapability(
    ctx, userID, subforumID, constants.CapabilityBanUsers)
```

## Database Setup

### 1. Pseudonym Roles and Capabilities

Pseudonyms store their own roles and capabilities in JSONB fields:

```sql
-- Example pseudonym with user role and basic capabilities
UPDATE pseudonyms 
SET roles = '["user"]'::jsonb,
    capabilities = '["create_content", "vote", "message", "report"]'::jsonb
WHERE pseudonym_id = 'pseudonym_123';

-- Example pseudonym with moderator capabilities for a specific subforum
UPDATE pseudonyms 
SET roles = '["user", "moderator"]'::jsonb,
    capabilities = '["create_content", "vote", "message", "report", "moderate_content", "ban_users"]'::jsonb
WHERE pseudonym_id = 'moderator_pseudonym_456';
```

### 2. User Accounts

Users only have basic account management - no roles or capabilities:

```sql
-- Example user account (no roles or capabilities)
UPDATE users 
SET email = 'user@example.com',
    is_active = TRUE
WHERE user_id = 1;
```

### 3. Subforum Moderators

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
// Check if user can moderate content in a subforum
canModerate, err := permissionDAO.HasSubforumCapabilityWithActivePseudonym(
    ctx, userID, subforumID, constants.CapabilityModerateContent, activePseudonymID)
if err != nil {
    return err
}

if !canModerate {
    return huma.Error403Forbidden("Insufficient permissions")
}
```

### 2. Subforum-Specific Session Endpoint

```go
// Get user session with subforum-specific capabilities
func (h *AuthHandler) GetCurrentUserSessionForSubforum(ctx context.Context, input *struct {
    middleware.AuthInput
    SubforumName string `path:"subforum_name"`
}) (*models.CurrentUserSessionResponse, error) {
    // Extract user context
    userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
    if err != nil {
        return nil, huma.Error401Unauthorized("Authentication required")
    }

    // Get subforum-specific capabilities
    subforumCapabilities := []string{}
    hasModerateContent, err := h.permissionDAO.HasSubforumCapabilityWithActivePseudonym(
        ctx, userCtx.UserID, subforumID, "moderate_content", userCtx.ActivePseudonymID)
    
    if err == nil && hasModerateContent {
        // Add moderator role dynamically
        if !contains(userCtx.Roles, "moderator") {
            userCtx.Roles = append(userCtx.Roles, "moderator")
        }
        subforumCapabilities = append(subforumCapabilities, "moderate_content")
    }

    // Return combined permissions
    return models.NewCurrentUserSessionResponse(
        userID,
        userCtx.Email,
        userCtx.Roles,        // Includes dynamically assigned roles
        append(userCtx.Capabilities, subforumCapabilities...), // Combined capabilities
        userCtx.ActivePseudonymID,
        userCtx.DisplayName,
        pseudonymInfos,
    ), nil
}
```

### 3. Pseudonym Management

```go
// Switch active pseudonym
func (h *AuthHandler) SwitchPseudonym(ctx context.Context, input *struct {
    middleware.AuthInput
    models.SwitchPseudonymInput
}) (*models.SwitchPseudonymResponse, error) {
    userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
    if err != nil {
        return nil, huma.Error401Unauthorized("Authentication required")
    }

    // Verify pseudonym ownership
    ownsPseudonym, err := h.pseudonymDAO.VerifyPseudonymOwnership(
        ctx, input.Body.PseudonymID, userCtx.UserID, "user", constants.ScopeAuthentication)
    
    if !ownsPseudonym {
        return nil, huma.Error403Forbidden("You do not own this pseudonym")
    }

    // Generate new JWT with updated pseudonym context
    newUserCtx := &middleware.UserContext{
        UserID:            userCtx.UserID,
        Email:             userCtx.Email,
        Roles:             userCtx.Roles,
        Capabilities:      userCtx.Capabilities,
        ActivePseudonymID: input.Body.PseudonymID,
        DisplayName:       targetPseudonym.DisplayName,
    }

    accessToken, err := middleware.GenerateJWT(newUserCtx, h.config.JWT.Secret, h.config.JWT.Expiration)
    if err != nil {
        return nil, huma.Error500InternalServerError("Failed to generate new token")
    }

    return models.NewSwitchPseudonymResponse(accessToken, h.config.JWT.Expiration, h.config.JWT.Development), nil
}
```

### 4. Content Creation with Pseudonym Context

```go
// Create a post with pseudonym-based permissions
func (h *ContentHandler) CreatePost(ctx context.Context, input *models.CreatePostInput) (*models.CreatePostResponse, error) {
    userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
    if err != nil {
        return nil, huma.Error401Unauthorized("Authentication required")
    }

    // Check if user can create content with their active pseudonym
    canCreateContent := false
    for _, capability := range userCtx.Capabilities {
        if capability == constants.CapabilityCreateContent {
            canCreateContent = true
            break
        }
    }

    if !canCreateContent {
        return nil, huma.Error403Forbidden("Insufficient permissions to create content")
    }

    // Create post with pseudonym context
    post, err := h.postDAO.CreatePost(ctx, &dao.CreatePostParams{
        SubforumID:   input.Body.SubforumID,
        PseudonymID:  userCtx.ActivePseudonymID, // Use active pseudonym
        Title:        input.Body.Title,
        Content:      input.Body.Content,
        Slug:         input.Body.Slug,
    })

    if err != nil {
        return nil, fmt.Errorf("failed to create post: %w", err)
    }

    return models.NewCreatePostResponse(post), nil
}
```

### 5. Moderation with Subforum Context

```go
// Moderate content with subforum-specific permissions
func (h *ModerationHandler) ModerateContent(ctx context.Context, input *models.ModerateContentInput) (*models.ModerateContentResponse, error) {
    userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
    if err != nil {
        return nil, huma.Error401Unauthorized("Authentication required")
    }

    // Check subforum-specific moderation capability
    canModerate, err := h.permissionDAO.HasSubforumCapabilityWithActivePseudonym(
        ctx, userCtx.UserID, input.Body.SubforumID, constants.CapabilityModerateContent, userCtx.ActivePseudonymID)
    
    if err != nil || !canModerate {
        return nil, huma.Error403Forbidden("Insufficient moderation permissions")
    }

    // Perform moderation action
    action, err := h.moderationDAO.CreateModerationAction(ctx, &dao.CreateModerationActionParams{
        SubforumID:        input.Body.SubforumID,
        ModeratorPseudonymID: userCtx.ActivePseudonymID,
        TargetPseudonymID:    input.Body.TargetPseudonymID,
        ActionType:           input.Body.ActionType,
        Reason:               input.Body.Reason,
    })

    if err != nil {
        return nil, fmt.Errorf("failed to create moderation action: %w", err)
    }

    return models.NewModerateContentResponse(action), nil
}
```

## Migration from User-Level to Pseudonym-Level Permissions

### Database Migration

```sql
-- Add roles and capabilities to pseudonyms
ALTER TABLE pseudonyms ADD COLUMN roles JSONB DEFAULT '["user"]';
ALTER TABLE pseudonyms ADD COLUMN capabilities JSONB DEFAULT '["create_content", "vote", "message", "report"]';

-- Remove capabilities from users table
ALTER TABLE users DROP COLUMN capabilities;
```

### Code Migration

```go
// Old way (user-level capabilities)
userCapabilities := user.Capabilities

// New way (pseudonym-level capabilities)
pseudonymCapabilities := pseudonym.Capabilities
activePseudonymCapabilities := userCtx.Capabilities // From JWT context
```

## Best Practices

### 1. Always Check Active Pseudonym Context
```go
// Good: Check capabilities with active pseudonym
canModerate, err := permissionDAO.HasSubforumCapabilityWithActivePseudonym(
    ctx, userID, subforumID, capability, userCtx.ActivePseudonymID)

// Avoid: Checking without pseudonym context
canModerate, err := permissionDAO.HasSubforumCapability(ctx, userID, subforumID, capability)
```

### 2. Use Dynamic Role Assignment
```go
// Good: Dynamically assign moderator role when appropriate
if hasModerateContent && !contains(userCtx.Roles, "moderator") {
    userCtx.Roles = append(userCtx.Roles, "moderator")
}
```

### 3. Combine Permissions Properly
```go
// Good: Combine pseudonym and subforum capabilities
allCapabilities := append(pseudonymCapabilities, subforumCapabilities...)
allRoles := append(pseudonymRoles, subforumRoles...)
```

### 4. Validate Pseudonym Ownership
```go
// Always verify pseudonym ownership before operations
ownsPseudonym, err := pseudonymDAO.VerifyPseudonymOwnership(
    ctx, pseudonymID, userID, role, constants.ScopeAuthentication)
```

## Related Documentation

- [RBAC Overview](rbac-overview.md)
- [User Roles](user-roles.md)
- [Role Keys and Site Roles](role-keys-and-site-roles.md) 