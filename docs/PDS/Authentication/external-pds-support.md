# External PDS Support

## Overview

HashPost now supports "bring your own PDS" functionality, allowing users to authenticate with any atproto-compliant Personal Data Server while maintaining HashPost's forum features. This follows Tangled's architecture pattern of providing both hosted and external PDS support.

## Architecture

### Dual PDS Support

HashPost maintains two types of PDS support:

1. **Hosted PDS**: HashPost's own PDS server for easy user onboarding
2. **External PDS**: Support for users with their own PDS servers

### Authentication Flow

#### Local Users (HashPost PDS)
- Users register and authenticate directly with HashPost's PDS
- Standard password-based authentication
- All user data stored locally in HashPost's database

#### External Users (Bring Your Own PDS)
- Users authenticate with their home PDS server
- HashPost validates tokens issued by external PDS servers
- Lightweight user records created in HashPost's AppView database
- Only AppView-specific data (roles, permissions, forum activity) stored locally

## Implementation Details

### PDS Authentication Layer

The PDS authentication system (`internal/pds/auth.go`) has been updated to:

1. **Detect User Type**: Check if user exists in local database
2. **Local Authentication**: Use password validation for local users
3. **External Authentication**: Use external PDS client for external users
4. **Session Creation**: Create sessions for both user types

### External PDS Client

The `ExternalPDSClient` (`internal/pds/external_pds.go`) handles:

- **PDS Discovery**: Resolve PDS endpoints from DID documents
- **Cross-PDS Authentication**: Authenticate users against their home PDS
- **Token Validation**: Validate JWT tokens from external PDS servers
- **Public Key Management**: Fetch and cache PDS public keys

### Multi-PDS Token Validation

The `MultiPDSTokenValidator` (`internal/pds/token_validator.go`) provides:

- **Token Parsing**: Extract issuer, user info, and claims
- **Public Key Fetching**: Retrieve and cache PDS public keys
- **Signature Validation**: Validate tokens using appropriate PDS public keys
- **Claims Verification**: Verify expiration, audience, and scope

### AppView Integration

The AppView system (`internal/appview/rbac.go`) has been updated to:

- **Automatic User Creation**: Create lightweight records for external users
- **RBAC Support**: Assign default roles and permissions to external users
- **Session Management**: Handle sessions from any PDS source
- **Data Separation**: Store only AppView-specific data locally

## Database Schema

### PDS Database

New table for external PDS OAuth registrations:

```sql
CREATE TABLE external_pds_clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pds_endpoint VARCHAR(500) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(pds_endpoint)
);
```

### AppView Database

Enhanced user table with external PDS support:

```sql
ALTER TABLE appview_users ADD COLUMN pds_source VARCHAR(500);
ALTER TABLE appview_users ADD COLUMN is_local BOOLEAN DEFAULT TRUE;
ALTER TABLE appview_users ADD COLUMN last_seen_at TIMESTAMPTZ;
```

## Configuration

### Development Configuration

```yaml
pds:
  external_support:
    enabled: true
    token_cache_ttl: "1h"
    public_key_cache_ttl: "24h"
  oauth:
    client_name: "HashPost"
    redirect_uris:
      - "http://localhost:3000/auth/callback"
```

### Environment Variables

```bash
ENABLE_EXTERNAL_PDS=true
PDS_PUBLIC_KEY_CACHE_TTL=24h
OAUTH_CLIENT_NAME=HashPost
```

## Usage

### For Users

1. **Hosted PDS Users**: Register normally on HashPost
2. **External PDS Users**: 
   - Provide DID or handle during registration
   - Authenticate with their home PDS
   - Access HashPost features with external identity

### For Developers

The system automatically handles both user types:

- **Authentication**: Transparent to application code
- **User Management**: Automatic lightweight record creation
- **RBAC**: Works with both local and external users
- **Session Handling**: Unified session management

## Benefits

1. **User Choice**: Users can choose their PDS provider
2. **Data Sovereignty**: Users control their primary data
3. **Federation**: True atproto federation support
4. **Scalability**: Reduced load on HashPost's PDS
5. **Compatibility**: Works with any atproto-compliant PDS

## Future Enhancements

1. **OAuth Flow**: Complete OAuth implementation for external PDS
2. **PDS Discovery**: Automatic PDS endpoint resolution
3. **Data Sync**: Bidirectional data synchronization
4. **Advanced RBAC**: PDS-specific role inheritance
5. **Performance**: Enhanced caching and optimization

## Security Considerations

1. **Token Validation**: Proper signature verification for all tokens
2. **Public Key Caching**: Secure caching of PDS public keys
3. **Session Management**: Secure session handling across PDS sources
4. **Data Isolation**: Proper separation of local and external user data
5. **Audit Logging**: Comprehensive logging of external PDS interactions
