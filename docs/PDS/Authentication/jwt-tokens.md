# JWT Tokens

## Overview

The PDS implements JWT token-based authentication using ES256K cryptographic signing, following atproto specifications. Tokens are generated for session management and API access control.

## Implementation

### JWT Service

**File**: `internal/jwt/jwt_service_impl.go`  
**Service**: `ProductionJWTService`

The JWT service handles token generation and validation:

```go
type ProductionJWTService struct {
    algorithm          string
    expiration         time.Duration
    validateSignatures bool
}

func NewProductionJWTService(config JWTServiceConfig) *ProductionJWTService {
    return &ProductionJWTService{
        algorithm:          config.Algorithm,  // "ES256K"
        expiration:         config.Expiration, // 1 hour
        validateSignatures: config.ValidateSignatures,
    }
}
```

### Token Generation

**Method**: `GenerateTokens(session *Session) (string, string, error)`

The PDS generates two types of tokens:

#### Access Token (Short-lived)
- **Expiration**: 1 hour
- **Scope**: `com.atproto.access`
- **Purpose**: API access and authentication

```go
// Create access token (short-lived, 1 hour)
accessClaims := map[string]interface{}{
    "sub":    session.DID,
    "iss":    as.serverDID,
    "aud":    as.serverDID,
    "iat":    now.Unix(),
    "exp":    now.Add(time.Hour).Unix(),
    "jti":    session.ID,
    "scope":  "com.atproto.access",
    "handle": session.Handle,
}

accessHeader := map[string]interface{}{
    "alg": "ES256K",
    "typ": "JWT",
}
```

#### Refresh Token (Long-lived)
- **Expiration**: 30 days
- **Scope**: `com.atproto.refresh`
- **Purpose**: Token renewal without re-authentication

```go
// Create refresh token (long-lived, 30 days)
refreshClaims := map[string]interface{}{
    "sub":    session.DID,
    "iss":    as.serverDID,
    "aud":    as.serverDID,
    "iat":    now.Unix(),
    "exp":    now.Add(30 * 24 * time.Hour).Unix(),
    "jti":    session.ID + "_refresh",
    "scope":  "com.atproto.refresh",
    "handle": session.Handle,
}
```

### Cryptographic Signing

**Algorithm**: ES256K (Elliptic Curve Digital Signature Algorithm with secp256k1 curve)

The PDS uses ES256K signing for atproto compliance:

```go
// Generate signing key
signingKey, err := crypto.GeneratePrivateKeyK256()
if err != nil {
    logger.Error("Failed to generate signing key", "error", err)
    panic("Failed to generate signing key")
}

// Generate access token using JWT service
accessTokenString, err := as.jwtService.GenerateSignedToken(accessClaims, accessHeader, as.signingKey)
```

### Token Validation

**Method**: `ValidateToken(token string) (*Session, error)`

```go
func (as *AuthService) ValidateToken(token string) (*Session, error) {
    // Use JWT service to validate and parse the token
    claims, err := as.jwtService.ValidateAndParse(token, func(token interface{}) (interface{}, error) {
        // Get the public key from our signing key
        publicKey, err := as.signingKey.PublicKey()
        if err != nil {
            return nil, fmt.Errorf("failed to get public key: %w", err)
        }
        return publicKey, nil
    })

    if err != nil {
        return nil, fmt.Errorf("token validation failed: %w", err)
    }

    // Extract required fields and validate expiration
    // ...
}
```

## Token Structure

### Claims Structure

```json
{
  "sub": "did:plc:hashpost-binding-test",
  "iss": "did:plc:hashpost-server", 
  "aud": "did:plc:hashpost-server",
  "iat": 1759928110,
  "exp": 1759931710,
  "jti": "session-uuid",
  "scope": "com.atproto.access",
  "handle": "testuser.hashpost.local"
}
```

### Header Structure

```json
{
  "alg": "ES256K",
  "typ": "JWT"
}
```

## API Integration

### Session Creation Endpoint

**Endpoint**: `POST /xrpc/com.atproto.server.createSession`

```go
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
    // Authenticate user
    session, err := s.authService.AuthenticateSession(r.Context(), req.Identifier, req.Password)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // Generate tokens
    accessToken, refreshToken, err := s.authService.GenerateTokens(session)
    if err != nil {
        http.Error(w, "Failed to create session", http.StatusInternalServerError)
        return
    }

    response := map[string]interface{}{
        "accessJwt":  accessToken,
        "refreshJwt": refreshToken,
        "handle":     session.Handle,
        "did":        session.DID,
        "email":     s.getUserEmailFromDID(r.Context(), session.DID),
    }
}
```

### Token Refresh Endpoint

**Endpoint**: `POST /xrpc/com.atproto.server.refreshSession`

```go
func (s *Server) handleRefreshSession(w http.ResponseWriter, r *http.Request) {
    // Validate refresh token
    session, err := s.authService.ValidateToken(req.RefreshJwt)
    if err != nil {
        http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }

    // Generate new tokens
    accessToken, refreshToken, err := s.authService.GenerateTokens(session)
    // ...
}
```

## Security Features

### Cryptographic Security
- **ES256K Signing**: Real cryptographic signatures using Indigo crypto
- **Key Management**: Secure private key generation and storage
- **Signature Verification**: Proper signature verification and claim extraction

### Token Security
- **Expiration Handling**: Automatic token expiration and refresh flow
- **Scope Validation**: Proper scope checking for access control
- **Session Binding**: Tokens bound to specific sessions with JTI

### Environment Isolation
- **Development**: Mock identities for local testing
- **Production**: Real atproto identity resolution
- **No Credential Leakage**: Production credentials not exposed in development

## Configuration

### JWT Service Configuration

```go
jwtService := jwtservice.NewProductionJWTService(jwtservice.JWTServiceConfig{
    Algorithm:          "ES256K",
    Expiration:         time.Hour,
    ValidateSignatures: true,
})
```

### Server DID Configuration

```go
serverDID := os.Getenv("SERVER_DID")
if serverDID == "" {
    serverDID = "did:plc:hashpost-server"
}
```

## Error Handling

- **Invalid Token**: Returns validation error for malformed tokens
- **Expired Token**: Handles token expiration gracefully
- **Signature Verification**: Validates cryptographic signatures
- **Missing Claims**: Ensures required claims are present

## References

- [Atproto Authentication Specification](https://atproto.com/specs/authentication)
- [JWT Service Implementation](internal/jwt/jwt_service_impl.go)
- [Authentication Service](internal/pds/auth.go:254-365)
- [Session Management](internal/pds/auth.go:481-663)
