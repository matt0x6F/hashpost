# API Keys Documentation

## Overview

The HashPost system supports two types of authentication:

1. **JWT Tokens** - Stateless tokens stored in cookies, used for user sessions
2. **API Keys** - Static tokens passed in Authorization headers, used for programmatic access by pseudonyms

## API Key Authentication

### API Key Structure

API keys are stored in the `api_keys` table with the following structure:

- `key_id` - Unique identifier
- `key_name` - Human-readable name for the key
- `key_hash` - SHA-256 hash of the raw key (for security)
- `pseudonym_id` - The pseudonym this API key belongs to
- `permissions` - JSON object containing roles and capabilities
- `created_at` - When the key was created
- `expires_at` - Optional expiration date
- `is_active` - Whether the key is currently active
- `last_used_at` - Last time the key was used

### Permissions Structure

API key permissions are stored as JSON with the following structure:

```json
{
  "roles": ["admin", "moderator"],
  "capabilities": ["read", "write", "delete"]
}
```

- `roles` - Array of role names the key has
- `capabilities` - Array of specific capabilities the key has

**Note**: API keys are now directly associated with pseudonyms via the `pseudonym_id` field, so there's no need for a `user_id` in the permissions.

### Using the API Key DAO

#### Creating an API Key

```go
import "github.com/matt0x6f/hashpost/internal/database/dao"

// Create API Key DAO
apiKeyDAO := dao.NewAPIKeyDAO(db)

// Define permissions
permissions := &dao.APIKeyPermissions{
    Roles:        []string{"admin"},
    Capabilities: []string{"read", "write"},
}

// Create the API key for a specific pseudonym
apiKey, err := apiKeyDAO.CreateAPIKey(
    ctx,
    "My API Key",
    "raw-api-key-string",
    "pseudonym_123", // The pseudonym this key belongs to
    permissions,
    nil, // No expiration
)
```

#### Validating an API Key

```go
// Validate an API key from a request
permissions, pseudonymID, err := apiKeyDAO.ValidateAPIKey(ctx, "raw-api-key-string")
if err != nil {
    // Handle invalid key
    return err
}

// Use the permissions and pseudonym ID
if contains(permissions.Roles, "admin") {
    // User has admin role
}
fmt.Printf("API key belongs to pseudonym: %s\n", pseudonymID)
```

#### Getting API Keys for a Pseudonym

```go
// Get all API keys for a specific pseudonym
apiKeys, err := apiKeyDAO.GetAPIKeysByPseudonymID(ctx, "pseudonym_123")
if err != nil {
    return err
}

for _, key := range apiKeys {
    fmt.Printf("Key: %s, Active: %v\n", key.KeyName, key.IsActive.V)
}
```

#### Getting API Key with Pseudonym Information

```go
// Get an API key with its associated pseudonym details
apiKey, err := apiKeyDAO.GetAPIKeyWithPseudonym(ctx, keyID)
if err != nil {
    return err
}

if apiKey != nil && apiKey.R.Pseudonym != nil {
    fmt.Printf("Key belongs to pseudonym: %s\n", apiKey.R.Pseudonym.DisplayName)
}
```

### Authentication Flow

1. **Client sends request** with `Authorization: Bearer <api-key>` header
2. **Auth middleware extracts** the API key from the header
3. **API Key DAO validates** the key against the database
4. **User context created** from the key's permissions and pseudonym ID
5. **Request continues** with the user context

### Security Features

- **Key Hashing**: Raw API keys are hashed with SHA-256 before storage
- **Pseudonym Association**: Each API key is tied to a specific pseudonym
- **Expiration**: Keys can have optional expiration dates
- **Active Status**: Keys can be deactivated without deletion
- **Usage Tracking**: Last used timestamp is updated on each validation
- **Permission Granularity**: Fine-grained control via roles and capabilities

### Management Operations

#### List API Keys

```go
apiKeys, err := apiKeyDAO.ListAPIKeys(ctx, 10, 0) // limit=10, offset=0
```

#### Deactivate API Key

```go
err := apiKeyDAO.DeactivateAPIKey(ctx, keyID)
```

#### Activate API Key

```go
err := apiKeyDAO.ActivateAPIKey(ctx, keyID)
```

#### Update API Key

```go
updates := &models.APIKeySetter{
    KeyName: &newName,
    // ... other fields
}
err := apiKeyDAO.UpdateAPIKey(ctx, keyID, updates)
```

#### Cleanup Expired Keys

```go
cleanedCount, err := apiKeyDAO.CleanupExpiredAPIKeys(ctx)
```

### Integration with Auth Middleware

The auth middleware automatically handles both JWT and API key authentication:

```go
// Create auth middleware with API Key DAO
authMiddleware := middleware.NewAuthMiddleware(jwtSecret, apiKeyDAO)

// Use in your routes
router.Use(authMiddleware.AuthenticateUser)
```

When an API key is validated, the middleware creates a user context with:
- `ActivePseudonymID` set to the pseudonym ID from the API key
- `Roles` and `Capabilities` from the key's permissions
- `TokenType` set to "api_token"

### Best Practices

1. **Use descriptive names** for API keys to identify their purpose
2. **Set expiration dates** for temporary keys
3. **Grant minimal permissions** - only what's needed
4. **Rotate keys regularly** for security
5. **Monitor usage** through the `last_used_at` field
6. **Use different keys** for different pseudonyms/services
7. **Associate keys with specific pseudonyms** for accountability

### Example Usage in HTTP Client

```bash
# Using curl with API key
curl -H "Authorization: Bearer your-api-key-here" \
     https://api.hashpost.com/v1/posts

# Using with specific permissions
curl -H "Authorization: Bearer admin-api-key" \
     -X POST \
     -H "Content-Type: application/json" \
     -d '{"title":"New Post","content":"Hello World"}' \
     https://api.hashpost.com/v1/posts
```

### Error Handling

Common API key errors:

- `API key not found` - Invalid or non-existent key
- `API key is inactive` - Key has been deactivated
- `API key has expired` - Key has passed its expiration date
- `API key is not associated with a pseudonym` - Key lacks pseudonym association
- `failed to parse API key permissions` - Corrupted permissions data

### Migration from JWT-only

If you're migrating from JWT-only authentication:

1. **Create API keys** for existing pseudonyms
2. **Update client code** to use `Authorization: Bearer <api-key>` headers
3. **Test thoroughly** to ensure permissions work correctly
4. **Monitor logs** for any authentication issues
5. **Gradually migrate** services to use API keys

### Pseudonym-Based Design

The new API key system is designed around pseudonyms:

- **Direct Association**: Each API key belongs to exactly one pseudonym
- **Accountability**: All actions performed with an API key are attributed to the pseudonym
- **Isolation**: Different pseudonyms can have different API keys with different permissions
- **Flexibility**: A user can have multiple pseudonyms, each with their own API keys 