# User Roles & Permissions Overview

HashPost uses a role-based access control (RBAC) system to manage what users can do on the platform. Each pseudonym is assigned one or more roles, which grant specific permissions and capabilities.

**Important**: Permissions are now managed at multiple levels:
- **User Level**: Basic account management and authentication (no roles or capabilities)
- **Pseudonym Level**: Content creation, voting, messaging, and personal pseudonym management
- **Subforum Level**: Moderation capabilities specific to individual subforums
- **Platform Level**: Administrative capabilities across the entire platform

---

## Permission Levels

### User Level Permissions
These apply to the user account itself:
- Account management and profile updates
- Authentication and session management
- Basic user operations
- **Note**: Users do not have roles or capabilities - these are managed at the pseudonym level

### Pseudonym Level Permissions
Each pseudonym has its own set of roles and capabilities:
- **Default Capabilities**: `create_content`, `vote`, `message`, `report`
- **Management Capabilities**: `manage_own_pseudonyms` (create, switch, deactivate pseudonyms)
- **Roles**: Typically start with `["user"]`

### Subforum Level Permissions
These are specific to individual subforums:
- **Dynamic Role Assignment**: The "moderator" role is automatically assigned when a user has subforum moderation capabilities
- **Subforum Capabilities**: `moderate_content`, `ban_users`, `manage_moderators`
- **Scope**: Only apply within the specific subforum

### Platform Level Permissions
These apply across the entire platform:
- **Administrative Roles**: Platform Admin, Trust & Safety, Legal Team
- **Platform Capabilities**: `correlate_identities`, `access_all_pseudonyms`, `system_admin`
- **Scope**: Apply to all subforums and users

---

## Platform-Wide Roles

| Role Name         | Description                                                                 |
|-------------------|-----------------------------------------------------------------------------|
| **Platform Admin**   | Full system access. Can manage all users, settings, and content.             |
| **Trust & Safety**   | Handles abuse, harassment, and safety issues across the platform.            |
| **Legal Team**       | Manages legal compliance, court orders, and privacy requests.                |

---

## Subforum Roles

| Role Name            | Description                                                        |
|----------------------|--------------------------------------------------------------------|
| **Owner**            | Full control over a subforum, including moderator management.       |
| **Moderator**        | Can moderate content and users within a subforum. **Dynamically assigned** when user has subforum moderation capabilities. |
| **Junior Moderator** | Limited moderation capabilities in a subforum.                     |

---

## Regular User

| Role Name         | Description                                      |
|-------------------|--------------------------------------------------|
| **User**          | Standard platform user. Can post, comment, vote, and message. |

---

## How Roles Affect Permissions

### Permission Inheritance
- **Platform-wide roles** grant permissions everywhere on HashPost
- **Subforum roles** only apply within specific communities
- **Pseudonym capabilities** are specific to individual pseudonyms
- **Dynamic role assignment** occurs when users access subforums with moderation capabilities

### Session Context
When a user accesses a subforum, the system:
1. Checks the user's active pseudonym
2. Evaluates subforum-specific capabilities for that pseudonym
3. Dynamically assigns the "moderator" role if appropriate
4. Returns combined permissions in the session response

### Example Session Response
```json
{
  "user_id": 123,
  "email": "user@example.com",
  "roles": ["user", "moderator"],  // "moderator" added dynamically
  "capabilities": [
    "create_content", 
    "vote", 
    "message", 
    "report",
    "moderate_content",  // From subforum capabilities
    "ban_users"          // From subforum capabilities
  ],
  "active_pseudonym_id": "pseudonym_123",
  "display_name": "UserPseudonym"
}
```

---

## Pseudonym Management

### Creating Pseudonyms
Users can create multiple pseudonyms, each with its own:
- Display name
- Roles and capabilities
- Activity history
- Karma score

### Switching Pseudonyms
Users can switch between their pseudonyms:
- Each pseudonym maintains its own context
- Capabilities may differ between pseudonyms
- Session context updates to reflect the active pseudonym

### Deactivating Pseudonyms
Users can deactivate pseudonyms they no longer need:
- Cannot deactivate the currently active pseudonym
- Deactivated pseudonyms retain their history
- Can be reactivated if needed

---

## How to Become an Admin or Moderator

- **Admins** are created by system operators
- **Moderators** are assigned by subforum owners or admins
- **Subforum moderation** is automatically detected based on pseudonym capabilities
- **Platform roles** require explicit assignment by administrators

---

## Security Notes

- All admin actions are logged for audit purposes
- Admins and moderators should use strong passwords and enable MFA (multi-factor authentication)
- Pseudonym permissions are cryptographically secured using IBE
- Subforum-specific permissions are validated on each request

---

For more detailed information about implementing roles and permissions, contact your platform administrator or see the [RBAC Usage Example](rbac-usage-example.md). 