# HashPost Domain Keys, Roles, and Capabilities Audit

## Executive Summary

This audit provides a comprehensive analysis of HashPost's permission system, covering domain keys, roles, role keys, scopes, capabilities, users, and pseudonyms. The system has evolved from a simple role-based system to a sophisticated unified permission system with cryptographic domain separation and pseudonym-based role management.

## System Architecture Overview

### Core Components

1. **Users** - Base authentication accounts (no roles/capabilities)
2. **Pseudonyms** - User identities with roles and capabilities via role keys
3. **Role Keys** - Primary permission storage with cryptographic domain separation
4. **Identity Mappings** - Encrypted links between real identities and pseudonyms
5. **Domain Keys** - Cryptographic separation for different privilege levels
6. **Scopes** - Operational contexts for IBE operations
7. **Capabilities** - Specific operations within each scope

### Completed Migration

**Migration Applied**: Remove deprecated admin fields from users table
- **File**: `internal/database/migrations/20250805000100-remove_admin_fields_from_users.sql`
- **Status**: ✅ Successfully applied
- **Fields Removed**: `admin_username`, `admin_password_hash`, `admin_scope`
- **Reason**: These fields were no longer used in the current unified permission system
- **Impact**: User model now only contains authentication-related fields, with all admin functionality handled through the unified permission system

## Detailed Component Analysis

### 1. Users (Base Authentication Layer)

**Purpose**: Provide basic authentication and account management
**Location**: `internal/database/models/users.bob.go`

**Key Fields**:
- `user_id` (BIGSERIAL PRIMARY KEY)
- `email` (VARCHAR(255) UNIQUE)
- `password_hash` (VARCHAR(255))
- `is_active` (BOOLEAN)
- `is_suspended` (BOOLEAN)
- `admin_username` (VARCHAR(100) UNIQUE) - For platform admins
- `admin_password_hash` (VARCHAR(255)) - For platform admins
- `admin_scope` (VARCHAR(100)) - 'trust_safety', 'legal', 'platform_admin'

**Important**: Users do NOT have roles or capabilities directly. These are managed at the pseudonym level via role keys.

**Removed Fields**:
- `admin_username` (VARCHAR(100) UNIQUE) - ✅ Removed via migration
- `admin_password_hash` (VARCHAR(255)) - ✅ Removed via migration  
- `admin_scope` (VARCHAR(100)) - ✅ Removed via migration

These admin fields were part of the old permission system and have been successfully removed. All admin functionality is now handled through the unified permission system with role keys.

**Relationships**:
- One-to-many with Pseudonyms
- One-to-many with IdentityMappings
- One-to-one with UserPreferences

### 2. Pseudonyms (Identity Layer)

**Purpose**: User identities with roles and capabilities
**Location**: `internal/database/models/pseudonyms.bob.go`

**Key Fields**:
- `pseudonym_id` (VARCHAR(64) PRIMARY KEY)
- `display_name` (VARCHAR(50))
- `karma_score` (INTEGER)
- `is_default` (BOOLEAN) - Marks the default pseudonym
- `slug` (VARCHAR) - URL-friendly identifier

**Important**: Pseudonyms are the primary carriers of roles and capabilities via role keys.

**Relationships**:
- Many-to-one with Users
- One-to-many with RoleKeys
- One-to-many with Posts, Comments, Votes, etc.

### 3. Role Keys (Permission Storage)

**Purpose**: Primary storage for all permissions with cryptographic domain separation
**Location**: `internal/database/models/role_keys.bob.go`

**Key Fields**:
- `key_id` (UUID PRIMARY KEY)
- `role_name` (VARCHAR(100)) - e.g., "user", "moderator", "platform_admin"
- `scope` (VARCHAR(100)) - e.g., "authentication", "moderation", "correlation"
- `key_data` (BYTEA) - Encrypted key material
- `key_version` (INTEGER) - For key rotation
- `capabilities` (JSONB) - Array of specific capabilities
- `expires_at` (TIMESTAMP) - Time-bounded access
- `is_active` (BOOLEAN)
- `created_by` (VARCHAR(64)) - Pseudonym ID of creator
- `pseudonym_id` (VARCHAR(64)) - Target pseudonym
- `subforum_id` (INTEGER NULL) - NULL for global, specific ID for subforum-specific

**Critical Design**:
- **Global Role Keys**: `subforum_id = NULL` - Apply across all subforums
- **Subforum-Specific Role Keys**: `subforum_id = specific_id` - Apply only to specific subforum
- **Automatic Role Assignment**: "moderator" role automatically added when subforum capabilities present

### 4. Identity Mappings (Privacy Layer)

**Purpose**: Encrypted links between real identities and pseudonyms
**Location**: `internal/database/models/identity_mappings.bob.go`

**Key Fields**:
- `mapping_id` (UUID PRIMARY KEY)
- `fingerprint` (VARCHAR(32)) - SHA-256 hash of real identity + salt
- `pseudonym_id` (VARCHAR(64))
- `encrypted_real_identity` (BYTEA) - Encrypted email/phone
- `encrypted_pseudonym_mapping` (BYTEA) - Encrypted mapping data
- `key_version` (INTEGER) - For key rotation
- `user_id` (BIGINT) - Links to user account
- `key_scope` (VARCHAR) - Scope for this mapping

**Privacy Protection**:
- Real identities never stored in plain text
- Fingerprints used for correlation instead of real identities
- Encrypted mappings require administrative keys to decrypt

### 5. Domain Keys (Cryptographic Separation)

**Purpose**: Separate cryptographic domains for different privilege levels
**Location**: `internal/ibe/ibe.go`

**Domain Constants**:
```go
const (
    DOMAIN_USER_PSEUDONYMS   = "user_pseudonyms_v1"
    DOMAIN_USER_CORRELATION  = "user_self_correlation_v1"
    DOMAIN_MOD_CORRELATION   = "moderator_correlation_v1"
    DOMAIN_ADMIN_CORRELATION = "admin_correlation_v1"
    DOMAIN_LEGAL_CORRELATION = "legal_correlation_v1"
)
```

**Security Benefits**:
- **Privilege Isolation**: Moderator key compromise doesn't affect user pseudonyms
- **Administrative Separation**: Admin key compromise doesn't affect legal operations
- **Cryptographic Boundaries**: Each domain mathematically isolated

### 6. Scopes (Operational Contexts)

**Purpose**: Define operational contexts for IBE operations and access control
**Location**: `internal/api/constants/scopes.go`

**Scope Definitions**:
- **Authentication** (`authentication`): Basic user authentication and session management
- **Self-Correlation** (`self_correlation`): Users managing their own identity and pseudonyms
- **Moderation** (`moderation`): Content moderation, report reviews, and rule management
- **Administration** (`administration`): Platform administration and user management
- **Correlation** (`correlation`): Cross-user identity correlation (admin only)

### 7. Capabilities (Specific Operations)

**Purpose**: Define specific operations within each scope
**Location**: `internal/api/constants/capabilities.go`

**Capability Categories**:

#### Authentication Capabilities
- `access_own_pseudonyms` - Access user's own pseudonyms
- `login` - Authenticate and create sessions
- `session_management` - Manage active sessions

#### Self-Correlation Capabilities
- `verify_own_pseudonym_ownership` - Verify pseudonym ownership
- `manage_own_profile` - Manage user's own profile
- `manage_own_pseudonyms` - Create, switch, deactivate pseudonyms

#### Correlation Capabilities
- `access_all_pseudonyms` - Access all pseudonyms (admin only)
- `access_subforum_pseudonyms` - Access pseudonyms within a subforum
- `cross_user_correlation` - Correlate identities across users
- `correlate_fingerprints` - Correlate fingerprints within a subforum

#### Moderation Capabilities
- `moderate_content` - Moderate content within a subforum
- `ban_users` - Ban users from a subforum
- `remove_content` - Remove content from a subforum
- `manage_moderators` - Manage moderator assignments
- `review_reports` - Review and resolve reports
- `forward_reports` - Forward reports to platform-level moderators
- `manage_subforum_rules` - Create, update, delete subforum rules
- `manage_subforum_settings` - Manage subforum settings

#### Platform-Wide Capabilities
- `moderation` - Platform-wide moderation operations
- `compliance` - Compliance-related operations
- `legal_requests` - Handle legal requests
- `system_admin` - Full system administration
- `user_management` - Manage user accounts

#### Basic User Capabilities
- `create_content` - Create posts and comments
- `vote` - Vote on posts and comments
- `message` - Send direct messages
- `report` - Report content or users
- `create_subforum` - Create new subforums

## Permission System Evolution

### Phase 1: Initial Schema (001_initial_schema.sql)
- Users had direct roles and capabilities in JSONB fields
- Simple role-based system
- No pseudonym separation

### Phase 2: Multiple Pseudonyms (20250622092738-multiple_pseudonyms_support.sql)
- Introduced pseudonyms table
- Moved user profile data to pseudonyms
- Content now linked to pseudonyms instead of users
- Users became pure authentication layer

### Phase 3: Role Keys Migration (20250803000000-migrate-role-keys-to-pseudonym-based.sql)
- Migrated from subforum_moderators table to role_keys
- Introduced unified permission system
- Added subforum-specific role keys
- Implemented automatic role assignment

### Phase 4: Unified Permission System (Current)
- Single source of truth for all permissions via role keys
- Global and subforum-specific capabilities
- Cryptographic domain separation
- Time-bounded key derivation

## Permission Checking Patterns

### Global Permission Check
```go
// Check global user capabilities (JWT token cached)
if !userCtx.HasCapability(constants.CapabilityCreateContent) {
    return huma.Error403Forbidden("insufficient permissions")
}
```

### Subforum-Specific Permission Check
```go
// Check unified capabilities (database role keys)
hasCapability, err := h.permissionDAO.HasUnifiedCapability(
    ctx, userCtx.UserID, userCtx.ActivePseudonymID, 
    constants.CapabilityModerateContent, &subforumID)
if err != nil {
    return fmt.Errorf("failed to check permissions: %w", err)
}
if !hasCapability {
    return huma.Error403Forbidden("insufficient permissions")
}
```

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

## Security Features

### 1. Cryptographic Domain Separation
- Separate master keys for each privilege level
- Compromise of one domain doesn't affect others
- Role-based domain selection

### 2. Time-Bounded Key Derivation
- All correlation keys include time components
- Forward secrecy through time windows
- Automatic key expiration

### 3. Unified Permission System
- Single source of truth via role keys
- Combines global and subforum-specific capabilities
- Automatic role assignment
- Duplicate capability removal

### 4. Privacy Protection
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

## Recommendations

### 1. Documentation
- ✅ Comprehensive audit completed
- ✅ Relationship diagrams created
- ✅ Permission patterns documented

### 2. Security
- ✅ Cryptographic domain separation implemented
- ✅ Time-bounded key derivation active
- ✅ Unified permission system operational

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