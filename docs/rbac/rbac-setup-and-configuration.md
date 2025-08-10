# RBAC Setup and Configuration

This document shows how to set up and configure the Role-Based Access Control (RBAC) system with centralized constants and database setup.

**Note**: For permission checking patterns and handler implementation examples, see [Permission Checking Patterns](permission-checking-patterns.md).

## Overview

The RBAC system provides:
- **PermissionDAO**: Data access layer for permission checking  
- **Constants**: Centralized role, scope, and capability definitions
- **Role Keys**: Cryptographic permission storage in the database

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
// Use constants for capability names
requiredCapability := constants.CapabilityModerateContent
globalCapability := constants.CapabilityCreateContent
adminCapability := constants.CapabilitySystemAdmin

// Always use constants instead of string literals
// Good: constants.CapabilityModerateContent
// Bad: "moderate_content"
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

## Permission Checking

For detailed permission checking patterns, handler implementation examples, and testing patterns, see:
- **[Permission Checking Patterns](permission-checking-patterns.md)** - Comprehensive guide for developers
- **[Unified Permission System](unified-permission-system.md)** - System architecture and real-world examples

## Quick Reference

### Basic Usage Patterns

```go
// Global capability check (use nil for subforumID)
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityCreateContent, nil)

// Subforum-specific capability check (use &subforumID)
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityModerateContent, &subforum.SubforumID)
```

### Common Constants

```go
// Roles
constants.RoleUser
constants.RoleModerator  
constants.RoleSubforumOwner
constants.RolePlatformAdmin

// Scopes
constants.ScopeAuthentication
constants.ScopeCorrelation
constants.ScopeSelfCorrelation

// Capabilities
constants.CapabilityCreateContent
constants.CapabilityModerateContent
constants.CapabilitySystemAdmin
```

## Setup Checklist

1. ✅ Import constants: `import "github.com/matt0x6f/hashpost/internal/api/constants"`
2. ✅ Set up role keys in database with proper capabilities
3. ✅ Use `HasUnifiedCapability()` for all permission checks
4. ✅ Use constants instead of string literals
5. ✅ Handle errors from permission checks
6. ✅ Use appropriate context (nil vs &subforumID)

## Related Documentation

- **[Permission Checking Patterns](permission-checking-patterns.md)** - Implementation guide
- **[Unified Permission System](unified-permission-system.md)** - System architecture
- **[User Roles](user-roles.md)** - Role definitions and capabilities
