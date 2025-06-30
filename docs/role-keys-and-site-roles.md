# Role Keys and Site Roles

This document explains the relationship between site roles (user permissions) and the `role_keys` table in the Identity-Based Encryption (IBE) system.

## Overview

The HashPost system uses a two-tier permission system:

1. **Site Roles**: Traditional user roles like "user", "platform_admin", "trust_safety", etc.
2. **Role Keys**: IBE cryptographic keys that enable secure operations for each role/scope combination

## Site Roles

Site roles define what a user can do within the system. Each role has associated capabilities:

### Standard Roles

- **`user`**: Basic user with content creation and voting rights
- **`platform_admin`**: Full system administration
- **`trust_safety`**: Content moderation and safety operations
- **`legal_team`**: Legal compliance and court order handling

### Role Capabilities

Each role has specific capabilities:

```go
// Example capabilities for different roles
"user": ["create_content", "vote", "message", "report", "create_subforum"]
"platform_admin": ["system_admin", "user_management", "correlate_identities", ...]
"trust_safety": ["correlate_identities", "cross_platform_access", "system_moderation", ...]
"legal_team": ["correlate_identities", "legal_compliance", "court_orders", ...]
```

## Role Keys

Role keys are IBE cryptographic keys stored in the `role_keys` table. Each key enables specific operations for a role within a particular scope.

### Key Structure

```sql
CREATE TABLE role_keys (
    key_id UUID PRIMARY KEY,
    role_name VARCHAR(50) NOT NULL,      -- e.g., "user", "platform_admin"
    scope VARCHAR(50) NOT NULL,          -- e.g., "authentication", "correlation"
    key_data BYTEA NOT NULL,             -- Encrypted IBE key
    capabilities JSONB NOT NULL,         -- What this key can do
    created_by BIGINT NOT NULL,          -- User who created this key
    created_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP,
    is_active BOOLEAN DEFAULT true
);
```

### Key Scopes

Each role key has a specific scope that defines its purpose:

- **`authentication`**: Used for login and session management
- **`self_correlation`**: Used for users to access their own pseudonyms
- **`correlation`**: Used for administrative identity correlation across users

### Default Keys

Every user gets default role keys created automatically:

```go
// Default keys created for each user
{
    role: "user",
    scope: "authentication",
    capabilities: ["access_own_pseudonyms", "login", "session_management"]
},
{
    role: "user", 
    scope: "self_correlation",
    capabilities: ["verify_own_pseudonym_ownership", "manage_own_profile"]
}
```

## Relationship Between Roles and Keys

### 1. Role → Key Mapping

Each site role can have multiple role keys for different scopes:

```
Site Role: "user"
├── Key 1: role="user", scope="authentication"
├── Key 2: role="user", scope="self_correlation"
└── Key 3: role="user", scope="correlation" (if admin)

Site Role: "platform_admin"  
├── Key 1: role="platform_admin", scope="authentication"
├── Key 2: role="platform_admin", scope="self_correlation"
└── Key 3: role="platform_admin", scope="correlation"
```

### 2. Key Creation Process

When a user is created (via registration or admin creation):

1. **User Creation**: User record is created with roles/capabilities
2. **Default Key Creation**: `EnsureDefaultKeys()` creates default role keys
3. **Pseudonym Creation**: User gets a pseudonym with IBE identity mapping

```go
// Example from registration handler
user, err := userDAO.CreateUser(ctx, email, passwordHash)
if err != nil {
    return nil, fmt.Errorf("failed to create user: %w", err)
}

// Create default role keys for the user
roleKeyDAO := dao.NewRoleKeyDAO(db)
if err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, user.UserID); err != nil {
    return nil, fmt.Errorf("failed to create default role keys: %w", err)
}
```

### 3. Key Usage in Operations

Role keys are used to authorize specific operations:

```go
// Example: Getting user pseudonyms
func (dao *SecurePseudonymDAO) GetPseudonymsByUserID(ctx context.Context, userID int64, roleName, scope string) ([]*models.Pseudonym, error) {
    // Validate that the key has the required capability
    hasCapability, err := dao.roleKeyDAO.ValidateKeyCapability(ctx, roleName, scope, "access_own_pseudonyms")
    if err != nil {
        return nil, fmt.Errorf("failed to validate key capability: %w", err)
    }

    if !hasCapability {
        return nil, fmt.Errorf("role key does not have permission to access own pseudonyms")
    }

    // Get the role key for this operation
    keyData, err := dao.roleKeyDAO.GetKeyData(ctx, roleName, scope)
    if err != nil {
        return nil, fmt.Errorf("failed to get role key: %w", err)
    }

    // Use the key to access pseudonyms
    return dao.getPseudonymsByUserIDWithKey(ctx, userID, keyData)
}
```

## Key Management

### Automatic Key Creation

Keys are created automatically for:
- New user registration
- Admin user creation
- Role changes (if needed)

### Key Validation

Before any operation, the system validates:
1. Key exists for the role/scope combination
2. Key has the required capability
3. Key is active and not expired
4. Key can be used by the requesting user

### Key Rotation

The system supports key rotation for security:
- Keys can be marked as expired
- New keys can be generated
- Grace periods allow for smooth transitions

## Database Queries

### Check User's Role Keys

```sql
SELECT role_name, scope, capabilities, is_active, expires_at
FROM role_keys 
WHERE created_by = $1
ORDER BY role_name, scope;
```

### Check Key Capabilities

```sql
SELECT capabilities 
FROM role_keys 
WHERE role_name = $1 
  AND scope = $2 
  AND is_active = true 
  AND (expires_at IS NULL OR expires_at > NOW());
```

### Find Users Without Keys

```sql
SELECT u.user_id, u.email, u.roles
FROM users u
LEFT JOIN role_keys rk ON u.user_id = rk.created_by
WHERE rk.key_id IS NULL;
```

## Troubleshooting

### Common Issues

1. **Missing Role Keys**: User can't login or access pseudonyms
   - **Solution**: Run `EnsureDefaultKeys()` for the user

2. **Invalid Key Capabilities**: Operation fails with permission error
   - **Solution**: Check key capabilities and update if needed

3. **Expired Keys**: Operations fail with key validation errors
   - **Solution**: Generate new keys or extend expiration

### Debugging Commands

```bash
# Check if user has role keys
docker-compose exec postgres psql -U hashpost -d hashpost -c "
SELECT u.email, COUNT(rk.key_id) as key_count
FROM users u
LEFT JOIN role_keys rk ON u.user_id = rk.created_by
GROUP BY u.user_id, u.email;"

# Check specific user's keys
docker-compose exec postgres psql -U hashpost -d hashpost -c "
SELECT role_name, scope, capabilities, is_active
FROM role_keys 
WHERE created_by = (SELECT user_id FROM users WHERE email = 'admin@example.com');"
```

## Security Considerations

1. **Key Isolation**: Each role/scope combination has its own key
2. **Capability Granularity**: Keys only grant specific capabilities
3. **User Association**: Keys are tied to specific users via `created_by`
4. **Expiration**: Keys can expire for security
5. **Audit Trail**: All key operations are logged

## Best Practices

1. **Always create default keys** when creating new users
2. **Validate key capabilities** before operations
3. **Use appropriate scopes** for different operations
4. **Monitor key expiration** and rotate as needed
5. **Audit key usage** regularly for security 