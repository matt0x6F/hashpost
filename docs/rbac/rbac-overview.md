# Role-Based Access Control (RBAC) System

## Overview

HashPost implements a sophisticated **Unified Permission System** that combines traditional role-based permissions with Identity-Based Encryption (IBE) for secure, privacy-preserving operations. This system has evolved from a simple role-based system to a sophisticated architecture with cryptographic domain separation and pseudonym-based role management.

**Important**: The permission system operates at multiple levels:
- **User Level**: Basic authentication and account management (no roles or capabilities)
- **Pseudonym Level**: Content creation, voting, messaging, reporting, and pseudonym management via role keys
- **Subforum Level**: Moderation capabilities specific to subforums via role keys
- **Platform Level**: Administrative capabilities across all subforums via role keys

### Unified Permission System

The system uses a **unified capability system** that leverages the `PermissionDAO.HasUnifiedCapability()` method for all permission checking. This system combines:

- **Global capabilities** from role keys with `subforum_id = NULL`
- **Subforum-specific capabilities** from role keys with specific `subforum_id`
- **Automatic role assignment** (e.g., "moderator" role when subforum capabilities are present)

#### Permission Checking Method
- **Method**: `permissionDAO.HasUnifiedCapability(ctx, userID, pseudonymID, capability, subforumID)`
- **Source**: Database role keys with cryptographic domain separation
- **Context**: Use `nil` for global checks, `&subforumID` for subforum-specific checks

#### Migration Status
- ✅ **Current**: All handlers now use unified capability system
- ❌ **Deprecated**: `userCtx.HasCapability()` (JWT token cached capabilities)
- ❌ **Deprecated**: Direct database queries for permissions

## Architecture

### Core Components

1. **Users**: Base authentication accounts (no roles/capabilities)
2. **Pseudonyms**: User identities with roles and capabilities via role keys
3. **Role Keys**: Primary permission storage with cryptographic domain separation
4. **Identity Mappings**: Encrypted links between real identities and pseudonyms
5. **Domain Keys**: Cryptographic separation for different privilege levels
6. **Scopes**: Operational contexts for IBE operations
7. **Capabilities**: Specific operations within each scope

### System Flow

```
User Authentication → Active Pseudonym → Role Key Resolution → Unified Capability Check → Operation Execution
```

### Permission Checking Patterns

All permission checking now uses the unified capability system. Here are the standard patterns:

#### Global Permission Check
```go
// Check global capabilities (e.g., create content, system admin)
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityCreateContent, nil) // nil = global scope
if err != nil {
    return fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    return huma.Error403Forbidden("insufficient permissions")
}
```

#### Subforum-Specific Permission Check
```go
// Check subforum-specific capabilities (e.g., moderation)
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityModerateContent, &subforumID) // specific subforum
if err != nil {
    return fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    return huma.Error403Forbidden("insufficient permissions")
}
```

#### Platform-Wide Administrative Check
```go
// Check platform-wide administrative capabilities
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilitySystemAdmin, nil) // nil = platform-wide
if err != nil {
    return fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    return huma.Error403Forbidden("insufficient permissions")
}
```

## Permission Hierarchy

### User Level
- Account management and authentication
- Basic user profile operations
- Session management
- **Note**: Users do not have roles or capabilities - these are managed at the pseudonym level via role keys

### Pseudonym Level
- Content creation and interaction
- Personal pseudonym management
- **Default capabilities**: Managed through role keys, not stored in pseudonyms table
- **Default roles**: Managed through role keys, not stored in pseudonyms table

### Subforum Level
- Moderation capabilities specific to subforums via role keys
- Dynamic role assignment via role keys
- Subforum-specific capabilities: `moderate_content`, `ban_users`, `manage_moderators`

### Platform Level
- Administrative capabilities across all subforums via role keys
- System-wide operations
- Platform-wide capabilities: `correlate_identities`, `access_all_pseudonyms`, `system_admin`

## Roles

### Pseudonym Roles

| Role | Description | Access Level |
|------|-------------|--------------|
| **User** | Standard platform user | Own data only |
| **Moderator** | Subforum content moderator | Subforum-specific (dynamically assigned) |
| **Subforum Owner** | Subforum administrator | Subforum-wide |
| **Trust & Safety** | Platform safety team | Platform-wide |
| **Legal Team** | Legal compliance team | Platform-wide |
| **Platform Admin** | System administrator | Full system access |

### Role Hierarchy

```
Platform Admin
├── Trust & Safety
├── Legal Team
└── Subforum Owner
    └── Moderator (dynamically assigned)
        └── User
```

**Note**: The "moderator" role is dynamically assigned when a user has subforum-specific moderation capabilities via role keys, rather than being a permanent user role.

## Scopes

Scopes define the operational context for cryptographic operations and access control:

### Authentication Scope
- **Purpose**: Basic user authentication and session management
- **Access**: User's own pseudonyms and profile data
- **Operations**: Login, session management, profile access

### Self-Correlation Scope
- **Purpose**: Users managing their own identity and pseudonyms
- **Access**: User's own identity mappings and pseudonym verification
- **Operations**: Pseudonym ownership verification, profile management, pseudonym management

### Correlation Scope
- **Purpose**: Administrative identity correlation and moderation
- **Access**: Cross-user identity correlation and administrative data
- **Operations**: Content moderation, user management, compliance

## Capabilities

Capabilities are organized by scope and define specific operations:

### Authentication Capabilities
- `access_own_pseudonyms`: Access user's own pseudonyms
- `login`: Authenticate and create sessions
- `session_management`: Manage active sessions

### Self-Correlation Capabilities
- `verify_own_pseudonym_ownership`: Verify pseudonym ownership
- `manage_own_profile`: Manage user's own profile
- `manage_own_pseudonyms`: Create, switch, and deactivate pseudonyms

### Correlation Capabilities
- `access_all_pseudonyms`: Access all pseudonyms (admin only)
- `access_subforum_pseudonyms`: Access pseudonyms within a subforum
- `cross_user_correlation`: Correlate identities across users
- `correlate_fingerprints`: Correlate fingerprints within a subforum

### Moderation Capabilities
- `moderate_content`: Moderate content within a subforum
- `ban_users`: Ban users from a subforum
- `remove_content`: Remove content from a subforum
- `manage_moderators`: Manage moderator assignments

### Platform-Wide Capabilities
- `moderation`: Platform-wide moderation operations
- `compliance`: Compliance-related operations
- `legal_requests`: Handle legal requests
- `system_admin`: Full system administration
- `user_management`: Manage user accounts

### Basic User Capabilities
- `create_content`: Create posts and comments
- `vote`: Vote on posts and comments
- `message`: Send direct messages
- `report`: Report content or users

## Role Key-Based Permissions

### Default Pseudonym Capabilities
Each pseudonym starts with these default capabilities via role keys:
```json
{
  "roles": ["user"],
  "capabilities": ["create_content", "vote", "message", "report"]
}
```

### Subforum-Specific Permissions
When a user accesses a subforum with moderation capabilities, the system:
1. Checks if the active pseudonym has subforum-specific role keys
2. Dynamically assigns the "moderator" role if appropriate
3. Combines pseudonym capabilities with subforum-specific capabilities
4. Returns the combined permission set in the session response

### Permission Inheritance
- **User permissions**: Apply to all pseudonyms owned by the user (basic account management only)
- **Pseudonym permissions**: Specific to individual pseudonyms via role keys
- **Subforum permissions**: Apply only within specific subforums via role keys
- **Platform permissions**: Apply across the entire platform via role keys

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

This architecture provides granular control over permissions while maintaining the privacy and security benefits of the IBE system. 