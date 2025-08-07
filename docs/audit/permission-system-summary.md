# HashPost Permission System Audit Summary

## Overview

This document provides a comprehensive summary of the HashPost permission system audit, including the relationships between domain keys, roles, role keys, scopes, capabilities, users, and pseudonyms.

## Audit Documents Created

### 1. Complete System Audit
**File**: `docs/audit/domain-keys-roles-capabilities-audit.md`

This document provides a detailed analysis of:
- **Users**: Base authentication layer (no roles/capabilities)
- **Pseudonyms**: Identity layer with roles/capabilities via role keys
- **Role Keys**: Primary permission storage with cryptographic domain separation
- **Identity Mappings**: Encrypted links between real identities and pseudonyms
- **Domain Keys**: Cryptographic separation for different privilege levels
- **Scopes**: Operational contexts for IBE operations
- **Capabilities**: Specific operations within each scope

### 2. Annotated Relationship Graphs
**File**: `docs/audit/permission-system-relationships.md`

This document contains comprehensive Mermaid diagrams showing:
- **System Architecture Graph**: Core entities and their relationships
- **Permission Flow Diagram**: How permission checking works
- **Role Key Structure**: Global vs subforum-specific role keys
- **Domain Key Separation**: Cryptographic domain architecture
- **Permission Checking Patterns**: Code examples for both systems
- **Role Hierarchy and Capabilities**: Role inheritance and capability flow
- **Database Schema Relationships**: Complete ER diagram

## Key Findings

### 1. **Two Permission Systems**
The system uses two distinct permission checking mechanisms:

#### Global Permissions (User Context)
- **Method**: `userCtx.HasCapability(capability)`
- **Source**: JWT token cached capabilities
- **Use Cases**: `create_content`, `vote`, `message`, `report`

#### Unified Permissions (Database Role Keys)
- **Method**: `permissionDAO.HasUnifiedCapability(ctx, userID, pseudonymID, capability, &subforumID)`
- **Source**: Database role assignments and specific permissions
- **Use Cases**: `moderate_content`, `ban_users`, `manage_subforum_settings`

### 2. **Cryptographic Domain Separation**
The IBE system uses separate cryptographic domains:
- `user_pseudonyms_v1`: Generate pseudonym IDs
- `user_self_correlation_v1`: User identity verification
- `moderator_correlation_v1`: Subforum moderation
- `admin_correlation_v1`: Platform administration
- `legal_correlation_v1`: Legal compliance

### 3. **Role Key Architecture**
- **Global Role Keys**: `subforum_id = NULL` (apply across all subforums)
- **Subforum-Specific Role Keys**: `subforum_id = specific_id` (apply only to specific subforum)
- **Automatic Role Assignment**: "moderator" role automatically added when subforum capabilities present

### 4. **Privacy Protection**
- Real identities never stored in plain text
- Fingerprints used for correlation instead of real identities
- Encrypted identity mappings require administrative keys to decrypt

## System Evolution

### Phase 1: Initial Schema
- Users had direct roles and capabilities in JSONB fields
- Simple role-based system

### Phase 2: Multiple Pseudonyms
- Introduced pseudonyms table
- Moved user profile data to pseudonyms
- Content now linked to pseudonyms instead of users

### Phase 3: Role Keys Migration
- Migrated from subforum_moderators table to role_keys
- Introduced unified permission system
- Added subforum-specific role keys

### Phase 4: Unified Permission System (Current)
- Single source of truth for all permissions via role keys
- Global and subforum-specific capabilities
- Cryptographic domain separation
- Time-bounded key derivation

## Security Features

### 1. **Cryptographic Domain Separation**
- Separate master keys for each privilege level
- Compromise of one domain doesn't affect others
- Role-based domain selection

### 2. **Time-Bounded Key Derivation**
- All correlation keys include time components
- Forward secrecy through time windows
- Automatic key expiration

### 3. **Unified Permission System**
- Single source of truth via role keys
- Combines global and subforum-specific capabilities
- Automatic role assignment
- Duplicate capability removal

### 4. **Privacy Protection**
- Real identities never stored in plain text
- Fingerprints used for correlation
- Encrypted identity mappings
- Administrative key requirements for decryption

## Database Schema Summary

### Core Tables
1. **users** - Authentication accounts (no roles/capabilities)
2. **pseudonyms** - User identities with profile data
3. **role_keys** - Primary permission storage
4. **identity_mappings** - Encrypted user-pseudonym links
5. **role_definitions** - Role templates and capabilities

### Content Tables (Linked to Pseudonyms)
- **posts** - `pseudonym_id` foreign key
- **comments** - `pseudonym_id` foreign key
- **votes** - `pseudonym_id` foreign key
- **reports** - `reporter_pseudonym_id` and `reported_pseudonym_id`

### Audit Tables
- **correlation_audit** - Logs all correlation activities
- **key_usage_audit** - Logs all key usage
- **moderation_actions** - Logs all moderation activities

## Role Hierarchy

```
Platform Admin (global)
├── Trust & Safety (global)
├── Legal Team (global)
└── Subforum Owner (subforum-specific)
    └── Moderator (subforum-specific, automatic)
        └── User (global, default)
```

## Key Relationships

### User → Pseudonym (1:Many)
- Each user can have multiple pseudonyms
- One pseudonym marked as `is_default`
- Pseudonyms carry roles and capabilities via role keys

### Pseudonym → Role Keys (1:Many)
- Each pseudonym can have multiple role keys
- Role keys can be global (`subforum_id = NULL`) or subforum-specific
- Role keys contain capabilities as JSONB array

### Role Keys → Scopes (Many:1)
- Each role key has a specific scope
- Scopes define the operational context
- Capabilities are validated against scopes

### Identity Mappings → Users (Many:1)
- Encrypted links between real identities and pseudonyms
- Fingerprints used for correlation instead of real identities
- Requires administrative keys to decrypt

## Recommendations

### 1. Documentation ✅
- Comprehensive audit completed
- Relationship diagrams created
- Permission patterns documented

### 2. Security ✅
- Cryptographic domain separation implemented
- Time-bounded key derivation active
- Unified permission system operational

### 3. Monitoring
- Implement real-time permission usage monitoring
- Add alerts for unusual permission patterns
- Monitor role key expiration and rotation

### 4. Testing
- Add comprehensive permission testing
- Test edge cases in unified permission system
- Validate cryptographic domain separation

## Conclusion

The HashPost permission system has evolved into a sophisticated, secure, and flexible architecture that provides:

1. **Strong Privacy Protection** through encrypted identity mappings and fingerprint-based correlation
2. **Granular Permission Control** through the unified role key system
3. **Cryptographic Security** through domain separation and time-bounded keys
4. **Scalable Architecture** that supports both global and subforum-specific permissions
5. **Audit Trail** through comprehensive logging of all permission-related activities

The system successfully balances user privacy with administrative accountability while providing the flexibility needed for a complex social platform.

## Related Documentation

- **Complete Audit**: `docs/audit/domain-keys-roles-capabilities-audit.md`
- **Relationship Graphs**: `docs/audit/permission-system-relationships.md`
- **RBAC Overview**: `docs/rbac/rbac-overview.md`
- **Unified Permission System**: `docs/rbac/unified-permission-system.md`
- **User Roles**: `docs/rbac/user-roles.md` 