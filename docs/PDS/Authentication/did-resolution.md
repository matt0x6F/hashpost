# DID Resolution

## Overview

The PDS implements DID (Decentralized Identifier) resolution using the Bluesky Indigo Go libraries. This provides the foundation for atproto authentication by resolving handles to DIDs and vice versa.

## Implementation

### Core Service

**File**: `internal/pds/auth.go`  
**Service**: `AuthService`

The `AuthService` handles DID resolution with environment-aware directory switching:

```go
// NewAuthService creates a new authentication service
func NewAuthService(db *generated.Queries, logger *slog.Logger, jwtService jwtservice.JWTService) *AuthService {
    var directory identity.Directory

    // Check environment to determine which directory to use
    environment := os.Getenv("ENVIRONMENT")
    if environment == EnvironmentProduction {
        // Production: Use real atproto identity resolution
        directory = identity.DefaultDirectory()
        logger.Info("Using production identity directory (plc.directory + DNS)")
    } else {
        // Development/Testing: Use mock directory with test identities
        mockDir := identity.NewMockDirectory()
        // Add test identities...
        directory = &mockDir
        logger.Info("Using development identity directory (mock)")
    }
}
```

### Environment Configuration

#### Development Mode
- **Directory**: `identity.NewMockDirectory()`
- **Test Users**: `testuser.hashpost.local`, `admin.hashpost.local`
- **DIDs**: `did:plc:hashpost-binding-test`, `did:plc:hashpost-admin-test`
- **Purpose**: Local testing without external dependencies

#### Production Mode
- **Directory**: `identity.DefaultDirectory()`
- **Resolution**: Connects to `plc.directory` and DNS
- **Purpose**: Full atproto protocol compliance

### Handle Resolution

**Method**: `ResolveHandle(ctx context.Context, handle string) (string, error)`

```go
func (as *AuthService) ResolveHandle(ctx context.Context, handle string) (string, error) {
    // First check if this is a local user in our database
    if as.db != nil {
        user, err := as.db.GetUserByHandle(ctx, handle)
        if err == nil && user != nil {
            return user.Did, nil
        }
    }

    // If not found in database, try the identity directory
    identity, err := as.directory.LookupHandle(ctx, syntax.Handle(handle))
    if err != nil {
        return "", fmt.Errorf("handle resolution failed: %w", err)
    }

    return identity.DID.String(), nil
}
```

### DID Resolution

**Method**: `ResolveDID(ctx context.Context, did string) (*identity.Identity, error)`

```go
func (as *AuthService) ResolveDID(ctx context.Context, did string) (*identity.Identity, error) {
    identity, err := as.directory.LookupDID(ctx, syntax.DID(did))
    if err != nil {
        return nil, fmt.Errorf("DID resolution failed: %w", err)
    }
    return identity, nil
}
```

### Authentication Flow

The authentication process uses DID resolution in the session creation flow:

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
    
    // Continue with password validation...
}
```

## Database Integration

The PDS maintains a local database cache of resolved identities:

- **Table**: `users`
- **Fields**: `did`, `handle`, `email`
- **Purpose**: Local caching and session management
- **Fallback**: Directory resolution for unknown identities

## Configuration

### Environment Variables

```bash
ENVIRONMENT=development  # or production
SERVER_DID=did:plc:hashpost-server  # Optional, defaults to did:plc:hashpost-server
```

### Mock Directory Setup

For development, the mock directory includes test identities:

```go
testUser := identity.Identity{
    DID:    syntax.DID("did:plc:hashpost-binding-test"),
    Handle: syntax.Handle("testuser.hashpost.local"),
}
adminUser := identity.Identity{
    DID:    syntax.DID("did:plc:hashpost-admin-test"),
    Handle: syntax.Handle("admin.hashpost.local"),
}
```

## Error Handling

- **Invalid Identifier**: Returns parsing error for malformed handles/DIDs
- **Resolution Failure**: Returns error with context for failed lookups
- **Database Errors**: Logs errors and falls back to directory resolution
- **Network Issues**: Handles timeouts and connection failures gracefully

## References

- [Atproto Identity Specification](https://atproto.com/specs/identity)
- [Bluesky Indigo Identity Package](https://github.com/bluesky-social/indigo/tree/main/atproto/identity)
- [DID Resolution Implementation](internal/pds/auth.go:367-404)
