# Unified Permission System

## Overview

The Unified Permission System resolves the incongruency between the global role system and subforum-specific moderator system by providing a single, consistent interface for permission checking that combines both sources of capabilities.

## Problem Solved

Previously, HashPost had two separate permission systems:

1. **Global System**: Pseudonym capabilities stored in `pseudonyms.capabilities` (JSONB) - **DEPRECATED**
2. **Subforum System**: Moderator capabilities stored in `subforum_moderators.permissions` (JSONB) + role-based capabilities - **DEPRECATED**

This created confusion and inconsistency in how permissions were checked and assigned.

## Solution

The unified system provides two new methods in the `PermissionDAO`:

### `GetUnifiedActivePseudonymRolesAndCapabilities(ctx, userID, activePseudonymID, subforumID)`

This method combines:
- **Global pseudonym capabilities** from role keys with `subforum_id = NULL`
- **Subforum-specific capabilities** from role keys with specific `subforum_id`
- **Automatic role assignment** - adds "moderator" role when subforum capabilities are present
- **Duplicate removal** - ensures no duplicate capabilities in the result

### `HasUnifiedCapability(ctx, userID, activePseudonymID, capability, subforumID)`

This method checks if a pseudonym has a specific capability by:
1. Getting unified roles and capabilities
2. Checking if the capability is present in the combined list
3. Supporting both global-only and subforum-context checks

## API Integration

### Global User Context (`/auth/me`)

```go
// Gets global pseudonym capabilities only (no subforum context)
roles, capabilities, err := permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
    ctx, userID, activePseudonymID, nil)
```

### Subforum-Specific Context (`/auth/me/subforum/{subforum_name}`)

```go
// Gets combined global + subforum-specific capabilities
subforumID := &subforum.SubforumID
roles, capabilities, err := permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
    ctx, userID, activePseudonymID, subforumID)
```

## Capability Sources

### Global Capabilities (from role_keys with subforum_id = NULL)
- `create_content` - Create posts and comments
- `vote` - Vote on content
- `message` - Send direct messages
- `report` - Report content or users
- Platform-wide admin capabilities (if assigned)

### Subforum-Specific Capabilities (from role_keys with specific subforum_id)
- Role-based capabilities from `role_keys.role_name`:
  - `moderate_content` - Moderate content
  - `ban_users` - Ban users from subforum
  - `remove_content` - Remove content
  - `manage_moderators` - Manage moderator assignments
- Specific permissions from `role_keys.capabilities` (JSONB):
  - Custom capabilities assigned to specific moderators
  - Granular permissions like `sticky_post`, `lock_post`

## Role Assignment Logic

### Automatic Role Assignment
- **"user"** - Default role for all pseudonyms (via role keys)
- **"moderator"** - Automatically added when subforum-specific capabilities are present
- **"owner"** - Manually added for subforum owners (not automatic)

### Role Hierarchy
```
Platform Admin (global)
├── Trust & Safety (global)
├── Legal Team (global)
└── Subforum Owner (subforum-specific)
    └── Moderator (subforum-specific, automatic)
        └── User (global, default)
```

## Database Schema

### Role Keys Table (Primary Permission Storage)
```sql
CREATE TABLE role_keys (
    key_id UUID PRIMARY KEY,
    role_name VARCHAR(100) NOT NULL,
    scope VARCHAR(100) NOT NULL,
    key_data BYTEA NOT NULL,
    key_version INTEGER NOT NULL DEFAULT 1,
    capabilities JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(64) NOT NULL,
    pseudonym_id VARCHAR(64) NOT NULL,
    subforum_id INTEGER NULL,
    
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id),
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id),
    FOREIGN KEY (created_by) REFERENCES pseudonyms(pseudonym_id)
);
```

### Users Table (No Roles or Capabilities)
```sql
-- Users do NOT have roles or capabilities columns
-- All permissions are managed through role_keys table
-- This ensures proper separation and cryptographic access control
```

### Pseudonyms Table (No Direct Roles/Capabilities)
```sql
-- Pseudonyms do NOT have roles or capabilities columns
-- All permissions are managed through role_keys table
-- This ensures proper separation and cryptographic access control
```

## Implementation Details

### Helper Methods

#### `getSubforumCapabilitiesForPseudonym(ctx, subforumID, pseudonymID)`
- Queries `role_keys` table for entries with specific `subforum_id` and `pseudonym_id`
- Combines role-based capabilities with specific permissions from `capabilities` JSONB
- Returns subforum-specific capabilities for a pseudonym

#### `removeDuplicateCapabilities(capabilities []string)`
- Removes duplicate capabilities from the combined list
- Ensures clean, unique capability sets

### Error Handling
- Graceful handling of missing role key records
- Proper error propagation for database issues
- Logging for debugging and monitoring

## Migration Strategy

### Phase 1: Add Unified Methods
- ✅ Added `GetUnifiedActivePseudonymRolesAndCapabilities`
- ✅ Added `HasUnifiedCapability`
- ✅ Updated interface definitions

### Phase 2: Update API Handlers
- ✅ Updated `/auth/me` to use unified system (global only)
- ✅ Updated `/auth/me/subforum/{subforum_name}` to use unified system (with subforum context)

### Phase 3: Complete Migration ✅
- ✅ Updated all handlers to use unified methods
- ✅ Deprecated legacy permission checking methods
- ✅ All endpoints now use `HasUnifiedCapability()` method

## Benefits

### Consistency
- Single source of truth for permission checking
- Consistent behavior across all API endpoints
- Unified logging and error handling

### Maintainability
- Clear separation of global vs. subforum-specific capabilities
- Easy to add new capabilities to either system
- Centralized permission logic

### Performance
- Reduced database queries (combines multiple sources)
- Efficient duplicate removal
- Cached role and capability lookups

### Flexibility
- Supports both global-only and subforum-context checks
- Maintains backward compatibility during migration
- Easy to extend with new capability types

## Testing

### Unit Tests
- `TestRemoveDuplicateCapabilities` - Tests duplicate removal logic
- `TestMockUnifiedSystem` - Tests unified system with mocks
- `TestUnifiedPermissionSystem` - Tests interface compliance

### Integration Tests (Future)
- Test with real database connections
- Test actual permission scenarios
- Test role assignment logic

## Usage Examples

### Checking Global Capabilities
```go
// Check if user can create content (global capability)
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityCreateContent, nil)
if err != nil {
    return fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    return huma.Error403Forbidden("insufficient permissions")
}
```

### Checking Subforum-Specific Capabilities
```go
// Check if user can moderate content in specific subforum
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityModerateContent, &subforum.SubforumID)
if err != nil {
    return fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    return huma.Error403Forbidden("insufficient permissions")
}
```

### Checking Platform-Wide Administrative Capabilities
```go
// Check if user has platform admin access
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilitySystemAdmin, nil)
if err != nil {
    return fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    return huma.Error403Forbidden("insufficient permissions")
}
```

### Getting Complete Permission Set
```go
// Get all capabilities for user in global context
roles, capabilities, err := h.permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, nil)
if err != nil {
    return fmt.Errorf("failed to get capabilities: %w", err)
}

// Get all capabilities for user in subforum context
roles, capabilities, err := h.permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, &subforum.SubforumID)
if err != nil {
    return fmt.Errorf("failed to get capabilities: %w", err)
}
```

## Real-World Implementation Examples

### Example 1: Subforum Creation (Global Capability)
From `subforums.go`:
```go
// Check global create_subforum capability via database
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityCreateSubforum, nil)
if err != nil {
    log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to check create_subforum capability")
    return nil, huma.Error500InternalServerError("failed to check permissions")
}
if !hasCapability {
    log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks create_subforum capability")
    return nil, huma.Error403Forbidden("insufficient permissions to create subforum")
}
```

### Example 2: Identity Correlation (Platform-Wide Administrative)
From `correlation.go`:
```go
// Validate admin permissions for platform-wide correlation
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityCorrelateIdentities, nil)
if err != nil {
    log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to check correlate_identities capability")
    return nil, fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    log.Warn().Int64("admin_id", adminID).Msg("User lacks correlate_identities capability")
    return nil, fmt.Errorf("insufficient permissions: correlate_identities capability required")
}
```

### Example 3: Report Forwarding (Specific Administrative Capability)
From `rules.go`:
```go
// Check specific capability for forwarding reports
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityForwardReports, nil)
if err != nil {
    log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to check forward_reports capability")
    return nil, fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    log.Error().Int("user_id", int(userCtx.UserID)).Msg("User lacks forward_reports capability")
    return nil, fmt.Errorf("insufficient permissions: forward_reports capability required")
}
```

### Example 4: Platform-Wide User Search (System Admin)
From `search.go`:
```go
// Check if user has platform admin capability via database
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, user.UserID, user.ActivePseudonymID, 
    constants.CapabilitySystemAdmin, nil)
if err != nil {
    log.Error().Err(err).Int64("user_id", user.UserID).Msg("Failed to check platform admin capability")
    return nil, fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    log.Warn().Int64("user_id", user.UserID).Msg("Platform admin capability required for user search")
    return nil, fmt.Errorf("insufficient permissions: platform admin capability required")
}
```

### Example 5: Getting Complete Permission Set for API Response
From `moderation.go`:
```go
// For global moderation endpoints, check platform-wide capabilities
// Get the unified roles and capabilities without a specific subforum
_, capabilities, err := h.permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
    context.Background(),
    userCtx.UserID,
    userCtx.ActivePseudonymID,
    nil, // No specific subforum for global moderation
)
if err != nil {
    log.Error().Err(err).Msg("Failed to get unified capabilities")
    return huma.Error500InternalServerError("Failed to validate permissions")
}

// Check for platform-wide moderation capabilities
hasModerateContent := false
hasSystemModeration := false

for _, cap := range capabilities {
    if cap == "moderate_content" {
        hasModerateContent = true
    }
    if cap == "system_moderation" {
        hasSystemModeration = true
    }
}
```

## Best Practices

### 1. Always Check Errors
```go
hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, capability, subforumID)
if err != nil {
    log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to check capability")
    return fmt.Errorf("failed to check permissions: %w", err)
}
```

### 2. Use Appropriate Context
- **Global capabilities**: Pass `nil` as `subforumID`
- **Subforum capabilities**: Pass `&subforum.SubforumID`
- **Never pass uninitialized pointers**

### 3. Log Permission Checks
```go
log.Info().
    Int64("user_id", userCtx.UserID).
    Str("capability", capability).
    Interface("subforum_id", subforumID).
    Bool("has_capability", hasCapability).
    Msg("Permission check completed")
```

### 4. Use Constants for Capabilities
```go
// Good
constants.CapabilityModerateContent

// Bad
"moderate_content"
```

### 5. Handle Permission Failures Gracefully
```go
if !hasCapability {
    log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks required capability")
    return huma.Error403Forbidden("insufficient permissions")
}
```

## Future Enhancements

### Planned Improvements
1. **Caching** - Cache permission results for better performance
2. **Audit Logging** - Track permission changes and usage
3. **Dynamic Capabilities** - Runtime capability assignment
4. **Permission Inheritance** - Hierarchical permission systems

### Potential Extensions
1. **Time-based Permissions** - Temporary capability assignments
2. **Conditional Permissions** - Context-dependent capabilities
3. **Permission Delegation** - Allow moderators to grant temporary permissions
4. **Permission Templates** - Predefined permission sets for common roles

## Conclusion

The Unified Permission System successfully resolves the incongruency between global and subforum-specific permission systems while maintaining flexibility and performance. It provides a clean, consistent interface for permission checking that combines both sources of capabilities in a logical and maintainable way. 