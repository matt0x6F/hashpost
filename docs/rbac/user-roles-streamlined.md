# User Roles & Permissions Overview

HashPost uses a role-based access control (RBAC) system to manage what users can do on the platform. Each pseudonym is assigned roles and capabilities via role keys.

**Important**: Permissions are managed at multiple levels:
- **User Level**: Basic account management and authentication (no roles or capabilities)
- **Pseudonym Level**: Content creation, voting, messaging, and personal pseudonym management via role keys
- **Subforum Level**: Moderation capabilities specific to individual subforums via role keys
- **Platform Level**: Administrative capabilities across the entire platform via role keys

---

## Platform-Wide Roles

| Role Name         | Description                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| **Platform Admin**   | Full system access. Can manage all users, settings, and content.             |
| **Trust & Safety**   | Handles abuse, harassment, and safety issues across the platform.            |
| **Legal Team**       | Manages legal compliance, court orders, and privacy requests.                |

---

## Subforum Roles

| Role Name            | Description                                                               |
|----------------------|---------------------------------------------------------------------------|
| **Subforum Owner**      | Created the subforum. Has all moderation powers plus ownership rights.      |
| **Moderator**           | Can moderate content, ban users, and manage subforum settings.              |

**Note**: The "moderator" role is dynamically assigned when a pseudonym has subforum-specific moderation capabilities via role keys.

---

## Pseudonym Roles

| Role Name | Description                                                                   |
|-----------|-------------------------------------------------------------------------------|
| **User**     | Standard platform user. Can create content, vote, message, and report.          |

**Note**: All pseudonyms start with the "user" role via role keys. Additional capabilities are granted through role key assignments.

---

## Role Capabilities Matrix

### Platform-Wide Capabilities

| Capability | Platform Admin | Trust & Safety | Legal Team | Description |
|------------|----------------|----------------|------------|-------------|
| `system_admin` | ✅ | ❌ | ❌ | Full system administration |
| `correlate_identities` | ✅ | ✅ | ✅ | Cross-platform identity correlation |
| `access_all_pseudonyms` | ✅ | ✅ | ✅ | Access all user pseudonyms |
| `user_management` | ✅ | ❌ | ❌ | Manage user accounts |
| `legal_compliance` | ❌ | ❌ | ✅ | Handle legal requests |

### Subforum-Specific Capabilities

| Capability | Subforum Owner | Moderator | Description |
|------------|----------------|-----------|-------------|
| `moderate_content` | ✅ | ✅ | Moderate posts and comments |
| `ban_users` | ✅ | ✅ | Ban users from subforum |
| `remove_content` | ✅ | ✅ | Remove posts and comments |
| `manage_moderators` | ✅ | ❌ | Add/remove moderators |
| `manage_subforum_settings` | ✅ | ❌ | Change subforum settings |
| `manage_subforum_rules` | ✅ | ❌ | Update subforum rules |

### Basic User Capabilities

| Capability | User | Description |
|------------|------|-------------|
| `create_content` | ✅ | Create posts and comments |
| `vote` | ✅ | Vote on content |
| `message` | ✅ | Send direct messages |
| `report` | ✅ | Report content or users |
| `create_subforum` | ✅* | Create new subforums |

*Subject to platform policies and may require approval.

---

## Role Assignment

### Platform Roles
- Assigned by existing Platform Admins
- Require manual approval and verification
- Apply across the entire platform

### Subforum Roles
- **Subforum Owner**: Automatically assigned when creating a subforum
- **Moderator**: Assigned by Subforum Owners or Platform Admins
- Apply only within the specific subforum

### User Role
- Automatically assigned to all pseudonyms via role keys
- Default role for all platform users
- Cannot be removed (but can be supplemented with additional roles)

---

## How to Become an Admin or Moderator

### Platform Admin
1. Must be nominated by an existing Platform Admin
2. Requires verification of identity and trustworthiness
3. Typically reserved for core team members

### Trust & Safety Team
1. Must be nominated by Platform Admin
2. Requires background in content moderation or safety
3. Subject to regular review and approval

### Subforum Moderator
1. **Apply**: Contact the Subforum Owner directly
2. **Demonstrate**: Show positive contribution to the community
3. **Approval**: Subforum Owner can assign moderator role via role keys

### Subforum Owner
1. **Create**: Create a new subforum (subject to platform policies)
2. **Request**: Request ownership of an abandoned subforum
3. **Transfer**: Receive ownership transfer from current owner

---

## Permission Checking

For information on how permissions are checked in the system, see:
- **[Permission Checking Patterns](permission-checking-patterns.md)** - Developer implementation guide
- **[Unified Permission System](unified-permission-system.md)** - System architecture

## Related Documentation

- **[RBAC Overview](rbac-overview.md)** - Complete system architecture
- **[Role Keys and Site Roles](role-keys-and-site-roles.md)** - Cryptographic key management
- **[RBAC Setup and Configuration](rbac-setup-and-configuration.md)** - Setup guide
