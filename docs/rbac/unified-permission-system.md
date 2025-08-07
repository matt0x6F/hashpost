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

### Phase 3: Gradual Migration (Future)
- Update other handlers to use unified methods
- Deprecate old methods once migration is complete
- Remove legacy methods in future versions

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
hasCapability, err := permissionDAO.HasUnifiedCapability(
    ctx, userID, activePseudonymID, "create_content", nil)
```

### Checking Subforum-Specific Capabilities
```go
// Check if user can moderate content in specific subforum
subforumID := int32(123)
hasCapability, err := permissionDAO.HasUnifiedCapability(
    ctx, userID, activePseudonymID, "moderate_content", &subforumID)
```

### Getting Complete Permission Set
```go
// Get all capabilities for user in subforum context
roles, capabilities, err := permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
    ctx, userID, activePseudonymID, &subforumID)
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