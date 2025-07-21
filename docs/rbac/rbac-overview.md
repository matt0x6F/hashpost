# Role-Based Access Control (RBAC) System

## Overview

HashPost implements a sophisticated Role-Based Access Control (RBAC) system that combines traditional role-based permissions with Identity-Based Encryption (IBE) for secure, privacy-preserving operations. This system ensures that users can only access data and perform operations appropriate to their role and scope.

**Important**: The permission system operates at multiple levels:
- **User Level**: Basic authentication and account management (no roles or capabilities)
- **Pseudonym Level**: Content creation, voting, messaging, reporting, and pseudonym management
- **Subforum Level**: Moderation capabilities specific to subforums
- **Platform Level**: Administrative capabilities across all subforums

## Architecture

### Core Components

1. **Roles**: Define pseudonym types and their basic permissions
2. **Scopes**: Define operational contexts for cryptographic operations
3. **Capabilities**: Define specific operations within each scope
4. **Role Keys**: IBE cryptographic keys that enable secure operations
5. **Identity Mappings**: Encrypted mappings between real identities and pseudonyms
6. **Pseudonym Permissions**: Role and capability assignments per pseudonym

### System Flow

```
User Authentication → Pseudonym Selection → Role Resolution → Scope Selection → Capability Check → Operation Execution
```

## Permission Hierarchy

### User Level
- Account management and authentication
- Basic user profile operations
- Session management
- **Note**: Users do not have roles or capabilities - these are managed at the pseudonym level

### Pseudonym Level
- Content creation and interaction
- Personal pseudonym management
- Default capabilities: `create_content`, `vote`, `message`, `report`, `manage_own_pseudonyms`
- Default roles: `["user"]`

### Subforum Level
- Moderation capabilities specific to subforums
- Dynamic role assignment (e.g., "moderator" role added when user has subforum moderation capabilities)
- Subforum-specific capabilities: `moderate_content`, `ban_users`, `manage_moderators`

### Platform Level
- Administrative capabilities across all subforums
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

**Note**: The "moderator" role is dynamically assigned when a user has subforum-specific moderation capabilities, rather than being a permanent user role.

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

## Pseudonym-Based Permissions

### Default Pseudonym Capabilities
Each pseudonym starts with these default capabilities:
```json
{
  "roles": ["user"],
  "capabilities": ["create_content", "vote", "message", "report"]
}
```

### Subforum-Specific Permissions
When a user accesses a subforum with moderation capabilities, the system:
1. Checks if the active pseudonym has subforum-specific capabilities
2. Dynamically assigns the "moderator" role if appropriate
3. Combines pseudonym capabilities with subforum-specific capabilities
4. Returns the combined permission set in the session response

### Permission Inheritance
- **User permissions**: Apply to all pseudonyms owned by the user (basic account management only)
- **Pseudonym permissions**: Specific to individual pseudonyms
- **Subforum permissions**: Apply only within specific subforums
- **Platform permissions**: Apply across the entire platform

## Database Schema

### Pseudonyms Table
```sql
ALTER TABLE pseudonyms ADD COLUMN roles JSONB DEFAULT '["user"]';
ALTER TABLE pseudonyms ADD COLUMN capabilities JSONB DEFAULT '["create_content", "vote", "message", "report"]';
```

### Users Table
```sql
-- Users no longer have roles or capabilities - these are managed at pseudonym level
-- Capabilities moved from users to pseudonyms
ALTER TABLE users DROP COLUMN capabilities;
```

This architecture provides granular control over permissions while maintaining the privacy and security benefits of the IBE system. 