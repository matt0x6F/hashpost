# Role-Based Access Control (RBAC) System

## Overview

HashPost implements a sophisticated Role-Based Access Control (RBAC) system that combines traditional role-based permissions with Identity-Based Encryption (IBE) for secure, privacy-preserving operations. This system ensures that users can only access data and perform operations appropriate to their role and scope.

## Architecture

### Core Components

1. **Roles**: Define user types and their basic permissions
2. **Scopes**: Define operational contexts for cryptographic operations
3. **Capabilities**: Define specific operations within each scope
4. **Role Keys**: IBE cryptographic keys that enable secure operations
5. **Identity Mappings**: Encrypted mappings between real identities and pseudonyms

### System Flow

```
User Authentication → Role Resolution → Scope Selection → Capability Check → Operation Execution
```

## Roles

### User Roles

| Role | Description | Access Level |
|------|-------------|--------------|
| **User** | Standard platform user | Own data only |
| **Moderator** | Subforum content moderator | Subforum-specific |
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
    └── Moderator
        └── User
```

## Scopes

Scopes define the operational context for cryptographic operations and access control:

### Authentication Scope
- **Purpose**: Basic user authentication and session management
- **Access**: User's own pseudonyms and profile data
- **Operations**: Login, session management, profile access

### Self-Correlation Scope
- **Purpose**: Users managing their own identity and pseudonyms
- **Access**: User's own identity mappings and pseudonym verification
- **Operations**: Pseudonym ownership verification, profile management

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
- `create_subforum`: Create new subforums

## Role Key System

### Overview

Role keys are IBE cryptographic keys that enable secure operations for each role/scope combination. Each key is associated with specific capabilities and has an expiration date.

### Key Structure

```sql
CREATE TABLE role_keys (
    key_id UUID PRIMARY KEY,
    role_name VARCHAR(100) NOT NULL,
    scope VARCHAR(100) NOT NULL,
    key_data BYTEA NOT NULL,
    capabilities JSONB NOT NULL,
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true,
    created_by BIGINT NOT NULL
);
```

### Key Creation

Role keys are created automatically when:
1. A new user is registered
2. A user's roles are updated
3. Keys are rotated for security

### Key Usage

Keys are used to:
1. **Encrypt/Decrypt Identity Mappings**: Secure the relationship between real identities and pseudonyms
2. **Validate Operations**: Ensure users can only perform operations appropriate to their role
3. **Enable Correlation**: Allow administrative roles to correlate identities when necessary

## Identity Management

### Pseudonym System

Users can create multiple pseudonyms, each representing a different online identity:

- **Default Pseudonym**: Automatically created for each user
- **Additional Pseudonyms**: Users can create additional identities
- **Identity Separation**: Pseudonyms are cryptographically separated by scope

### Identity Mappings

Identity mappings are encrypted using IBE keys and stored with scope information:

```sql
CREATE TABLE identity_mappings (
    mapping_id UUID PRIMARY KEY,
    fingerprint VARCHAR(32) NOT NULL,
    pseudonym_id VARCHAR(64) NOT NULL,
    encrypted_real_identity BYTEA NOT NULL,
    key_scope VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT true
);
```

## Access Control Flow

### 1. Authentication
```
User Login → JWT Token Generation → Role Resolution → Capability Assignment
```

### 2. Operation Request
```
API Request → Token Validation → Role Extraction → Scope Determination → Capability Check
```

### 3. Authorization
```
Capability Check → Role Key Validation → IBE Key Usage → Operation Execution
```

## Security Features

### Multi-Factor Authentication (MFA)
- Required for sensitive operations (correlation, admin actions)
- Configurable per role and capability
- Time-based one-time passwords (TOTP)

### Key Rotation
- Automatic key rotation for security
- Graceful transition between key versions
- Audit logging for key usage

### Audit Logging
- All capability usage is logged
- Identity correlation attempts are tracked
- Administrative actions are recorded

## Implementation

### Constants Organization

All RBAC constants are centralized in `internal/api/constants/`:

- `scopes.go`: Scope definitions and validation
- `capabilities.go`: Capability definitions organized by scope
- `roles.go`: Role definitions with associated scopes and capabilities

### Database Integration

RBAC is integrated with the database through:

- **Role Keys DAO**: Manages cryptographic keys
- **Permissions DAO**: Handles capability checking
- **Secure Pseudonym DAO**: Enforces access control on pseudonym operations

### API Integration

RBAC is enforced at the API level through:

- **Authentication Middleware**: Validates tokens and extracts user context
- **Permission Middleware**: Checks capabilities before allowing operations
- **Handler Integration**: Each handler validates required capabilities

## Usage Examples

### Creating a Moderator

```go
// Create user with moderator role
user, err := userDAO.CreateUser(ctx, email, passwordHash)
if err != nil {
    return err
}

// Assign moderator role
roles := []string{constants.RoleModerator}
userDAO.UpdateUserRoles(ctx, user.UserID, roles)

// Create role keys for moderator
roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, user.UserID)
```

### Checking Capabilities

```go
// Check if user can moderate content
canModerate, err := permissionDAO.HasSubforumCapability(
    ctx, userID, subforumID, constants.CapabilityModerateContent)
if err != nil {
    return err
}

if !canModerate {
    return huma.Error403Forbidden("Insufficient permissions")
}
```

### Secure Pseudonym Access

```go
// Get user's pseudonyms using authentication scope
pseudonyms, err := securePseudonymDAO.GetPseudonymsByUserID(
    ctx, userID, role, constants.ScopeAuthentication)
if err != nil {
    return err
}
```

## Best Practices

### 1. Principle of Least Privilege
- Always use the minimum required capabilities
- Avoid granting broad permissions unnecessarily
- Regularly review and audit permissions

### 2. Scope Separation
- Use appropriate scopes for different operations
- Avoid mixing authentication and correlation scopes
- Maintain clear boundaries between user and admin operations

### 3. Key Management
- Rotate keys regularly
- Monitor key usage and access patterns
- Implement proper key backup and recovery

### 4. Audit and Monitoring
- Log all capability usage
- Monitor for unusual access patterns
- Regular security reviews

## Related Documentation

- [Identity-Based Encryption](../security/ibe.md)
- [Role Keys and Site Roles](role-keys-and-site-roles.md)
- [RBAC Usage Examples](rbac-usage-example.md)
- [User Roles](user-roles.md)
- [Security Analysis](../security/domain-separation-security-analysis.md) 