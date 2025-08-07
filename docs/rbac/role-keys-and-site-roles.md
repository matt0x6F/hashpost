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
    role_name VARCHAR(100) NOT NULL,      -- e.g., "user", "platform_admin"
    scope VARCHAR(100) NOT NULL,          -- e.g., "authentication", "correlation"
    key_data BYTEA NOT NULL,              -- Encrypted IBE key
    key_version INTEGER NOT NULL DEFAULT 1,
    capabilities JSONB NOT NULL,          -- What this key can do
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by VARCHAR(64) NOT NULL,      -- Pseudonym ID of creator
    pseudonym_id VARCHAR(64) NOT NULL,    -- Pseudonym this key is for
    subforum_id INTEGER NULL,             -- NULL for global, specific ID for subforum-specific
    
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id),
    FOREIGN KEY (subforum_id) REFERENCES subforums(subforum_id),
    FOREIGN KEY (created_by) REFERENCES pseudonyms(pseudonym_id)
);
```

### Key Scopes

Each role key has a specific scope that defines its purpose:

- **`authentication`**: Used for login and session management
- **`self_correlation`**: Used for users to access their own pseudonyms
- **`correlation`**: Used for administrative identity correlation across users

### Default Keys

Every pseudonym gets default role keys created automatically:

```go
// Default keys created for each pseudonym
{
    role: "user",
    scope: "authentication",
    capabilities: ["access_own_pseudonyms", "login", "session_management"],
    subforum_id: NULL  // Global key
},
{
    role: "user", 
    scope: "self_correlation",
    capabilities: ["verify_own_pseudonym_ownership", "manage_own_profile"],
    subforum_id: NULL  // Global key
}
```

## Relationship Between Roles and Keys

### 1. Role → Key Mapping

Each site role can have multiple role keys for different scopes and contexts:

```
Site Role: "user"
├── Global Key 1: role="user", scope="authentication", subforum_id=NULL
├── Global Key 2: role="user", scope="self_correlation", subforum_id=NULL
└── Subforum Key: role="user", scope="correlation", subforum_id=123 (if moderator)

Site Role: "platform_admin"  
├── Global Key 1: role="platform_admin", scope="authentication", subforum_id=NULL
├── Global Key 2: role="platform_admin", scope="self_correlation", subforum_id=NULL
└── Global Key 3: role="platform_admin", scope="correlation", subforum_id=NULL
```

### 2. Key Creation Process

When a pseudonym is created:

1. **Pseudonym Creation**: Pseudonym record is created
2. **Default Key Creation**: `EnsureDefaultKeys()` creates default role keys for the pseudonym
3. **Identity Mapping**: Pseudonym gets IBE identity mapping

```go
// Example from pseudonym creation
pseudonym, err := pseudonymDAO.CreatePseudonym(ctx, userID, displayName)
if err != nil {
    return nil, fmt.Errorf("failed to create pseudonym: %w", err)
}

// Create default role keys for the pseudonym
roleKeyDAO := dao.NewRoleKeyDAO(db)
if err := roleKeyDAO.EnsureDefaultKeys(ctx, ibeSystem, pseudonym.PseudonymID); err != nil {
    return nil, fmt.Errorf("failed to create default role keys: %w", err)
}
```

### 3. Key Usage in Operations

Role keys are used to authorize specific operations:

```go
// Example: Getting user pseudonyms
func (dao *PseudonymDAO) GetPseudonymsByUserID(ctx context.Context, userID int64, roleName, scope string) ([]*models.Pseudonym, error) {
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
- New pseudonym creation
- Admin pseudonym creation
- Role changes (if needed)

### Key Validation

Before any operation, the system validates:
1. Key exists for the role/scope combination
2. Key has the required capability
3. Key is active and not expired
4. Key can be used by the requesting pseudonym

### Key Rotation

The system supports key rotation for security:
- Keys can be marked as expired
- New keys can be generated
- Grace periods allow for smooth transitions

## Database Queries

### Check Pseudonym's Role Keys

```sql
SELECT role_name, scope, capabilities, is_active, expires_at, subforum_id
FROM role_keys 
WHERE pseudonym_id = $1
ORDER BY role_name, scope;
```

### Check Key Capabilities

```sql
SELECT capabilities 
FROM role_keys 
WHERE role_name = $1 
  AND scope = $2 
  AND pseudonym_id = $3
  AND is_active = true 
  AND (expires_at IS NULL OR expires_at > NOW());
```

### Find Pseudonyms Without Keys

```sql
SELECT p.pseudonym_id, p.display_name
FROM pseudonyms p
LEFT JOIN role_keys rk ON p.pseudonym_id = rk.pseudonym_id
WHERE rk.key_id IS NULL;
```

### Check Subforum-Specific Keys

```sql
SELECT role_name, scope, capabilities
FROM role_keys 
WHERE pseudonym_id = $1 
  AND subforum_id = $2
  AND is_active = true;
```

## Troubleshooting

### Common Issues

1. **Missing Role Keys**: Pseudonym can't login or access content
   - **Solution**: Run `EnsureDefaultKeys()` for the pseudonym

2. **Invalid Key Capabilities**: Operation fails with permission error
   - **Solution**: Check key capabilities and update if needed

3. **Expired Keys**: Operations fail with key validation errors
   - **Solution**: Generate new keys or extend expiration

### Debugging Commands

```bash
# Check if pseudonym has role keys
docker-compose exec postgres psql -U hashpost -d hashpost -c "
SELECT p.display_name, COUNT(rk.key_id) as key_count
FROM pseudonyms p
LEFT JOIN role_keys rk ON p.pseudonym_id = rk.pseudonym_id
GROUP BY p.pseudonym_id, p.display_name;"

# Check specific pseudonym's keys
docker-compose exec postgres psql -U hashpost -d hashpost -c "
SELECT role_name, scope, capabilities, is_active, subforum_id
FROM role_keys 
WHERE pseudonym_id = 'pseudonym_id_here';"

# Check subforum-specific keys
docker-compose exec postgres psql -U hashpost -d hashpost -c "
SELECT rk.role_name, rk.scope, rk.capabilities, p.display_name
FROM role_keys rk
JOIN pseudonyms p ON rk.pseudonym_id = p.pseudonym_id
WHERE rk.subforum_id = 123;"
```

## Security Considerations

1. **Key Isolation**: Each role/scope/subforum combination has its own key
2. **Capability Granularity**: Keys only grant specific capabilities
3. **Pseudonym Association**: Keys are tied to specific pseudonyms via `pseudonym_id`
4. **Subforum Isolation**: Subforum-specific keys only work within that subforum
5. **Expiration**: Keys can expire for security
6. **Audit Trail**: All key operations are logged

## Best Practices

1. **Always create default keys** when creating new pseudonyms
2. **Validate key capabilities** before operations
3. **Use appropriate scopes** for different operations
4. **Monitor key expiration** and rotate as needed
5. **Audit key usage** regularly for security
6. **Separate global and subforum-specific keys** for proper isolation 