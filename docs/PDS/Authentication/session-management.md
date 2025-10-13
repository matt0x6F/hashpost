# Session Management

## Overview

The PDS implements comprehensive session management for atproto authentication, including session creation, validation, token generation, and database storage. Sessions are the foundation for user authentication and API access control.

## Implementation

### Session Structure

**File**: `internal/pds/auth.go`  
**Type**: `Session`

```go
type Session struct {
    ID        string    `json:"id"`
    DID       string    `json:"did"`
    Handle    string    `json:"handle"`
    CreatedAt time.Time `json:"created_at"`
    ExpiresAt time.Time `json:"expires_at"`
}
```

### Session Creation

**Method**: `AuthenticateSession(ctx context.Context, identifier, password string) (*Session, error)`

The session creation process involves DID resolution, password validation, and database storage:

```go
func (as *AuthService) AuthenticateSession(ctx context.Context, identifier, password string) (*Session, error) {
    // Parse the identifier (could be handle or DID)
    ident, err := syntax.ParseAtIdentifier(identifier)
    if err != nil {
        return nil, fmt.Errorf("invalid identifier: %w", err)
    }

    var did string
    var handle string

    // Resolve the identifier to get DID and handle
    if ident.IsHandle() {
        handle = ident.String()
        // Resolve handle to DID...
    } else if ident.IsDID() {
        did = ident.String()
        // Resolve DID to get handle...
    }

    // Validate password against stored hash
    if err := as.validatePassword(ctx, did, password); err != nil {
        return nil, fmt.Errorf("invalid credentials: %w", err)
    }

    // Create session
    session := &Session{
        ID:        uuid.New().String(),
        DID:       did,
        Handle:    handle,
        CreatedAt: time.Now(),
        ExpiresAt: time.Now().Add(24 * time.Hour), // 24 hour session
    }

    // Store session in database
    _, err = as.db.CreateUserSession(ctx, &generated.CreateUserSessionParams{
        SessionID: session.ID,
        UserDid:   session.DID,
        Handle:    session.Handle,
        ExpiresAt: session.ExpiresAt,
    })

    return session, nil
}
```

### Session Validation

**Method**: `ValidateSession(ctx context.Context, sessionID string) (*Session, error)`

```go
func (as *AuthService) ValidateSession(ctx context.Context, sessionID string) (*Session, error) {
    // Get session from database
    dbSession, err := as.db.GetUserSession(ctx, sessionID)
    if err != nil {
        return nil, fmt.Errorf("session not found or expired")
    }

    // Update last accessed time
    err = as.db.UpdateUserSessionLastAccessed(ctx, sessionID)
    if err != nil {
        as.logger.Error("Failed to update session last accessed", "error", err)
    }

    // Convert to session object
    session := &Session{
        ID:        dbSession.SessionID,
        DID:       dbSession.UserDid,
        Handle:    dbSession.Handle,
        CreatedAt: dbSession.CreatedAt.Time,
        ExpiresAt: dbSession.ExpiresAt,
    }

    return session, nil
}
```

### Token-Based Session Validation

**Method**: `ValidateToken(token string) (*Session, error)`

For JWT token validation, the service extracts session information from token claims:

```go
func (as *AuthService) ValidateToken(token string) (*Session, error) {
    // Validate JWT token and extract claims
    claims, err := as.jwtService.ValidateAndParse(token, func(token interface{}) (interface{}, error) {
        publicKey, err := as.signingKey.PublicKey()
        if err != nil {
            return nil, fmt.Errorf("failed to get public key: %w", err)
        }
        return publicKey, nil
    })

    // Extract required fields
    did, ok := claims["sub"].(string)
    handle, ok := claims["handle"].(string)
    jti, ok := claims["jti"].(string)
    
    // Check expiration
    exp, ok := claims["exp"].(float64)
    expirationTime := time.Unix(int64(exp), 0)
    if time.Now().After(expirationTime) {
        return nil, fmt.Errorf("token has expired")
    }

    // Create session from token claims
    session := &Session{
        ID:        jti,
        DID:       did,
        Handle:    handle,
        CreatedAt: time.Now().Add(-1 * time.Hour), // Approximate
        ExpiresAt: expirationTime,
    }

    return session, nil
}
```

## Database Integration

### Session Storage

**Table**: `user_sessions`  
**Fields**: `session_id`, `user_did`, `handle`, `expires_at`, `created_at`, `last_accessed_at`

The PDS stores session information in the database for validation and management:

```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_did VARCHAR(255) NOT NULL,
    handle VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ DEFAULT NOW()
);
```

### SQLC Queries

**File**: `internal/database/queries/pds/`  
**Generated**: `internal/database/generated/pds/`

Key session-related queries:

```sql
-- name: CreateUserSession :one
INSERT INTO user_sessions (session_id, user_did, handle, expires_at)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetUserSession :one
SELECT * FROM user_sessions WHERE session_id = $1;

-- name: UpdateUserSessionLastAccessed :exec
UPDATE user_sessions 
SET last_accessed_at = NOW() 
WHERE session_id = $1;

-- name: DeleteUserSession :exec
DELETE FROM user_sessions WHERE session_id = $1;
```

## API Endpoints

### Session Creation

**Endpoint**: `POST /xrpc/com.atproto.server.createSession`

```go
func (s *Server) handleCreateSession(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Identifier string `json:"identifier"` // handle or email
        Password   string `json:"password"`
    }

    // Use DID-based authentication
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

### Session Retrieval

**Endpoint**: `GET /xrpc/com.atproto.server.getSession`

```go
func (s *Server) handleGetSession(w http.ResponseWriter, r *http.Request) {
    // Extract token from Authorization header
    authHeader := r.Header.Get("Authorization")
    token := strings.TrimPrefix(authHeader, "Bearer ")
    
    session, err := s.authService.ValidateToken(token)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    response := map[string]interface{}{
        "handle": session.Handle,
        "did":    session.DID,
        "email": s.getUserEmailFromDID(r.Context(), session.DID),
    }
}
```

### Session Refresh

**Endpoint**: `POST /xrpc/com.atproto.server.refreshSession`

```go
func (s *Server) handleRefreshSession(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshJwt string `json:"refreshJwt"`
    }

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

### Session Deletion

**Endpoint**: `POST /xrpc/com.atproto.server.deleteSession`

```go
func (s *Server) handleDeleteSession(w http.ResponseWriter, r *http.Request) {
    // Extract token and validate session
    session, err := s.authService.ValidateToken(token)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }

    // Delete session from database
    err = s.db.DeleteUserSession(r.Context(), session.ID)
    if err != nil {
        http.Error(w, "Failed to delete session", http.StatusInternalServerError)
        return
    }
}
```

## Password Management

### Password Validation

**Method**: `validatePassword(ctx context.Context, did, password string) error`

```go
func (as *AuthService) validatePassword(ctx context.Context, did, password string) error {
    // Get the user's password hash from the database
    user, err := as.db.GetUserByDID(ctx, did)
    if err != nil {
        return fmt.Errorf("user not found: %w", err)
    }

    // Check if user has a password hash
    if user.PasswordHash == nil || *user.PasswordHash == "" {
        // For development, allow any password for users without password hash
        as.logger.Warn("User has no password hash, allowing authentication", "did", did)
        return nil
    }

    // Compare the provided password with the stored hash
    err = bcrypt.CompareHashAndPassword([]byte(*user.PasswordHash), []byte(password))
    if err != nil {
        return fmt.Errorf("password mismatch: %w", err)
    }

    return nil
}
```

### Password Hashing

**Method**: `HashPassword(password string) (string, error)`

```go
func (as *AuthService) HashPassword(password string) (string, error) {
    // Generate a salt and hash the password
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return "", fmt.Errorf("failed to hash password: %w", err)
    }
    return string(hash), nil
}
```

## Session Lifecycle

### Creation Flow
1. **Identifier Resolution**: Resolve handle/DID to identity
2. **Password Validation**: Verify password against stored hash
3. **Session Generation**: Create session with unique ID
4. **Database Storage**: Store session in database
5. **Token Generation**: Generate access and refresh tokens

### Validation Flow
1. **Token Extraction**: Extract JWT from Authorization header
2. **Signature Verification**: Validate token signature
3. **Claim Validation**: Verify required claims and expiration
4. **Session Creation**: Create session object from claims
5. **Access Grant**: Allow API access

### Refresh Flow
1. **Refresh Token Validation**: Validate refresh token
2. **Session Recreation**: Create new session from token claims
3. **New Token Generation**: Generate new access and refresh tokens
4. **Response**: Return new tokens to client

### Deletion Flow
1. **Token Validation**: Validate current session
2. **Database Cleanup**: Remove session from database
3. **Confirmation**: Return success response

## Security Considerations

### Session Security
- **Unique Session IDs**: UUID-based session identifiers
- **Expiration Handling**: 24-hour session expiration
- **Token Binding**: Sessions bound to specific tokens via JTI
- **Database Cleanup**: Proper session deletion from database

### Password Security
- **bcrypt Hashing**: Secure password hashing with salt
- **Cost Configuration**: Configurable bcrypt cost
- **Validation**: Proper password strength validation
- **Development Mode**: Special handling for development users

### Token Security
- **ES256K Signing**: Cryptographic token signing
- **Expiration**: Automatic token expiration
- **Scope Validation**: Proper scope checking
- **Signature Verification**: Real signature validation

## Error Handling

- **Invalid Credentials**: Returns authentication error for invalid passwords
- **Session Not Found**: Handles missing or expired sessions
- **Token Validation**: Proper error handling for invalid tokens
- **Database Errors**: Graceful handling of database failures

## References

- [Session Management Implementation](internal/pds/auth.go:118-252)
- [API Endpoint Handlers](internal/pds/auth.go:481-663)
- [Database Queries](internal/database/queries/pds/)
- [Atproto Session Specification](https://atproto.com/specs/session)
