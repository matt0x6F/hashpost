# Authentication Guide

## Overview

HashPost uses a unified authentication system that supports both web sessions and API access through two primary methods:

1. **JWT Tokens** - For user sessions (stored in cookies)
2. **API Keys** - For programmatic access (passed in Authorization header)

Both authentication methods are handled by the auth middleware automatically and provide role-based access control.

## JWT Authentication

### Overview

JWT (JSON Web Token) authentication provides secure, stateless authentication for web applications and API clients. The system supports both cookie-based sessions and header-based API access.

### Authentication Flow

#### 1. User Registration
- User provides email, password, and display name
- System creates user account and initial pseudonym
- Returns access token and refresh token (both as JWTs)
- Tokens are included in response body and set as HTTP-only cookies

#### 2. User Login
- User provides email and password
- System validates credentials and retrieves user data
- Returns access token and refresh token with user information
- Tokens are included in response body and set as HTTP-only cookies

#### 3. Token Refresh
- Client sends refresh token to `/auth/refresh`
- System validates refresh token and generates new access token
- New access token is returned and set as cookie
- Refresh token remains valid for continued use

#### 4. User Logout
- Client sends logout request with refresh token
- System can invalidate refresh token (TODO: implement token blacklisting)
- Client should clear local tokens and cookies

### Token Types

#### Access Token (JWT)
- **Purpose**: Short-lived token for API access
- **Expiration**: 24 hours (configurable)
- **Claims**: User ID, email, roles, capabilities, active pseudonym
- **Usage**: Sent in Authorization header or access_token cookie

#### Refresh Token (JWT)
- **Purpose**: Long-lived token for obtaining new access tokens
- **Expiration**: 7 days
- **Claims**: Same as access token
- **Usage**: Sent in refresh_token cookie or request body

### JWT Claims Structure

```go
type JWTClaims struct {
    UserID            int64    `json:"user_id"`
    Email             string   `json:"email"`
    Roles             []string `json:"roles"`
    Capabilities      []string `json:"capabilities"`
    MFAEnabled        bool     `json:"mfa_enabled"`
    ActivePseudonymID string   `json:"active_pseudonym_id"`
    DisplayName       string   `json:"display_name"`
    jwt.RegisteredClaims
}
```

### Cookie Configuration

#### Access Token Cookie
- **Name**: `access_token`
- **HttpOnly**: true
- **Secure**: true (false in development)
- **SameSite**: Strict
- **Expiration**: 24 hours

#### Refresh Token Cookie
- **Name**: `refresh_token`
- **HttpOnly**: true
- **Secure**: true (false in development)
- **SameSite**: Strict
- **Expiration**: 7 days

### Configuration

#### Environment Variables
```bash
# JWT Secret (REQUIRED - change in production)
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# JWT Token Expiration (optional, default: 24h)
JWT_EXPIRATION=24h

# Development mode (optional, default: true)
JWT_DEVELOPMENT=true

# Enable MFA requirements (optional, default: false)
SECURITY_ENABLE_MFA=false
```

#### JWT Configuration Structure
```go
type JWTConfig struct {
    Secret      string        // JWT signing secret
    Expiration  time.Duration // Access token expiration
    Development bool          // Controls cookie security
}
```

### Security Features

#### Current Implementation
- ✅ JWT tokens are signed with HMAC-SHA256
- ✅ Tokens include expiration claims
- ✅ Cookies are HttpOnly and Secure (in production)
- ✅ SameSite cookie policy prevents CSRF
- ✅ Tokens include user roles and capabilities

#### Planned Improvements
- 🔄 Refresh token blacklisting for logout
- 🔄 Token rotation on refresh
- 🔄 Rate limiting for authentication endpoints
- 🔄 MFA support for sensitive operations
- 🔄 Token revocation for compromised accounts

### API Endpoints

#### Authentication Endpoints
- `POST /auth/register` - User registration
- `POST /auth/login` - User login
- `POST /auth/logout` - User logout
- `POST /auth/refresh` - Token refresh

#### Protected Endpoints
All other endpoints require valid authentication via:
- Authorization header: `Bearer <access_token>`
- Cookie: `access_token=<token>`

### Client Integration

#### Web Applications
```javascript
// Login request
const response = await fetch('/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
});

// Cookies are automatically set by the browser
// Response body contains additional token information
const data = await response.json();
console.log('Access token:', data.body.access_token);
console.log('Refresh token:', data.body.refresh_token);
```

#### API Clients
```javascript
// Send access token in Authorization header
const response = await fetch('/api/v1/posts', {
    headers: { 
        'Authorization': 'Bearer ' + accessToken,
        'Content-Type': 'application/json'
    }
});

// Handle 401 responses by refreshing token
if (response.status === 401) {
    // Refresh token logic
    const refreshResponse = await fetch('/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: currentRefreshToken })
    });
}
```

## API Key Authentication

### Overview

API keys provide static authentication for programmatic access. Each API key is associated with a specific pseudonym and has defined permissions.

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

### Usage Examples

#### Using API Keys with curl
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

#### Using API Keys in JavaScript
```javascript
const response = await fetch('/api/v1/posts', {
    headers: {
        'Authorization': 'Bearer your-api-key-here',
        'Content-Type': 'application/json'
    }
});
```

## Authentication Middleware

### Overview

The system uses a unified authentication middleware that:

1. **Extracts tokens** from either Authorization header or cookies
2. **Validates tokens** using the JWT secret or API key database
3. **Creates user context** with user information and permissions
4. **Adds context to request** for downstream handlers

### Priority Order
1. Authorization header (Bearer token) - for API access
2. access_token cookie - for web sessions

### User Context

The `UserContext` structure provides user information extracted from JWT tokens or API keys:

```go
type UserContext struct {
    UserID            int64    `json:"user_id"`
    Email             string   `json:"email"`
    Roles             []string `json:"roles"`
    Capabilities      []string `json:"capabilities"`
    MFAEnabled        bool     `json:"mfa_enabled"`
    ActivePseudonymID string   `json:"active_pseudonym_id"`
    DisplayName       string   `json:"display_name"`
    TokenType         string   `json:"token_type"` // "jwt" or "api_token"
}
```

### Helper Methods

```go
// Check if user has a specific capability
func (uc *UserContext) HasCapability(capability string) bool

// Check if user has a specific role
func (uc *UserContext) HasRole(role string) bool

// Check if an action requires MFA
func (uc *UserContext) RequiresMFA(action string) bool
```

### Usage Examples

#### Extracting User Context
```go
// Extract user from JWT token (from cookie or header)
userCtx, err := middleware.ExtractUserFromRequest(r)
if err != nil {
    // No valid authentication
    return
}

// Use user context
userID := userCtx.UserID
capabilities := userCtx.Capabilities
```

#### Checking Permissions
```go
// Check if user has required capability
if !userCtx.HasCapability("create_content") {
    return fmt.Errorf("insufficient permissions")
}

// Check if user has required role
if !userCtx.HasRole("moderator") {
    return fmt.Errorf("moderator role required")
}
```

## Multi-Factor Authentication (MFA)

### Overview

The system supports configurable MFA requirements for sensitive operations. MFA can be globally enabled or disabled using the `SECURITY_ENABLE_MFA` configuration setting.

### MFA Requirements

When MFA is enabled, the following actions require MFA validation:

- **System Administration**: `system_admin` actions
- **Legal Compliance**: `legal_compliance` operations  
- **Identity Correlation**: `correlate_identities` operations
- **Fingerprint Correlation**: `correlate_fingerprints` (for users with correlation capabilities)

### Configuration

```bash
# Enable MFA requirements globally (default: false)
SECURITY_ENABLE_MFA=true
```

### Implementation Status

**Note**: The MFA system is currently in development. While the configuration and middleware infrastructure is in place, the actual MFA token validation is not yet implemented.

### Future MFA Implementation

When MFA validation is implemented, it will include:

1. **MFA Token Generation**: TOTP (Time-based One-Time Password) or similar
2. **MFA Token Validation**: Server-side validation of MFA codes
3. **MFA Setup**: User enrollment in MFA systems
4. **MFA Recovery**: Backup codes and recovery procedures

## Error Handling

### Authentication Errors
- `401 Unauthorized`: Invalid or missing token
- `403 Forbidden`: Insufficient permissions
- `400 Bad Request`: Invalid token format

### Token Refresh Errors
- `401 Unauthorized`: Invalid refresh token
- `400 Bad Request`: Missing refresh token

### API Key Errors
- `API key not found` - Invalid or non-existent key
- `API key is inactive` - Key has been deactivated
- `API key has expired` - Key has passed its expiration date
- `API key is not associated with a pseudonym` - Key lacks pseudonym association

## Best Practices

### JWT Best Practices
1. **Use descriptive names** for API keys to identify their purpose
2. **Set expiration dates** for temporary keys
3. **Grant minimal permissions** - only what's needed
4. **Rotate keys regularly** for security
5. **Monitor usage** through the `last_used_at` field
6. **Use different keys** for different pseudonyms/services

### Security Best Practices
1. **Store tokens securely** - Use HttpOnly cookies for web applications
2. **Validate tokens** - Always validate tokens on the server side
3. **Handle token expiration** - Implement proper refresh logic
4. **Log authentication events** - Monitor for suspicious activity
5. **Use HTTPS** - Always use HTTPS in production
6. **Implement rate limiting** - Prevent brute force attacks

## Troubleshooting

### Common Issues

1. **"Authorization header is required"**
   - Ensure the client includes the `Authorization: Bearer <token>` header

2. **"Invalid authorization header format"**
   - Check that the header follows the format: `Bearer <token>`
   - Ensure there's a space between "Bearer" and the token

3. **"Invalid or expired token"**
   - Check that the JWT secret matches between token generation and validation
   - Verify the token hasn't expired
   - Ensure the token signature is valid

4. **"User context not found in request context"**
   - Verify that authentication middleware is properly configured
   - Check that the middleware is applied to the correct routes

5. **"Cookies not being set"**
   - Check that the client is properly handling the response
   - Verify cookie settings (HttpOnly, Secure, SameSite)
   - Ensure the domain and path are correct

### Debugging

Enable debug logging to see authentication processing details:

```go
log.SetLevel(log.DebugLevel)
```

This will show:
- Token extraction attempts
- User context creation
- Authentication success/failure
- Middleware processing steps
- Cookie setting operations

## Migration Guide

### From Mock Implementation

#### Before (Mock JWT)
```go
userCtx := &middleware.UserContext{
    UserID:            123,
    Email:             "user@example.com",
    Roles:             []string{"user"},
    Capabilities:      []string{"create_content", "vote", "message", "report"},
    MFAEnabled:        false,
    ActivePseudonymID: "abc123def456...",
    DisplayName:       "user_display_name",
}
```

#### After (Real JWT Implementation)
```go
// Extract user from context
userCtx, err := middleware.ExtractUserFromContext(ctx)
if err != nil {
    log.Warn().Err(err).Msg("User context not available")
    return nil, fmt.Errorf("authentication required")
}
```

### From Header-Only Authentication

#### Before (Header Only)
```go
// Client sends Authorization header
req.Header.Set("Authorization", "Bearer "+token)
```

#### After (Cookie Support)
```go
// Client can use either headers or cookies
// Headers (for API clients)
req.Header.Set("Authorization", "Bearer "+token)

// Cookies (for web applications)
// Cookies are automatically sent by the browser
```

## Future Enhancements

1. **Refresh Token Storage**: Store refresh tokens in database for revocation
2. **Token Rotation**: Generate new refresh tokens on refresh
3. **Session Management**: Track active sessions per user
4. **Audit Logging**: Log authentication events
5. **Rate Limiting**: Prevent brute force attacks
6. **MFA Integration**: Support for multi-factor authentication
7. **OAuth Integration**: Support for third-party authentication providers
8. **Single Sign-On (SSO)**: Enterprise SSO integration

## References

- [JWT RFC 7519](https://tools.ietf.org/html/rfc7519)
- [HTTP-Only Cookies](https://owasp.org/www-community/HttpOnly)
- [SameSite Cookie Attribute](https://developer.mozilla.org/en-US/docs/Web/HTTP/Cookies#SameSite_attribute)
- [Huma Cookie Documentation](https://huma.rocks/features/response-outputs/#cookies) 