# HashPost Federation Architecture

## Overview

HashPost implements a federated architecture that supports both hosted and external Personal Data Servers (PDS), following the AT Protocol specification. This enables true decentralization while maintaining HashPost's forum functionality.

## Architecture Components

### PDS Layer (Personal Data Server)

#### HashPost Hosted PDS
- **Purpose**: Easy onboarding for new users
- **Authentication**: Password-based authentication
- **Data Storage**: Complete user profiles and atproto records
- **API**: Standard atproto endpoints (`/xrpc/com.atproto.*`)

#### External PDS Support
- **Purpose**: Support for users with their own PDS
- **Authentication**: JWT token validation from external PDS
- **Data Storage**: Lightweight records for AppView integration
- **API**: Cross-PDS authentication and token validation

### AppView Layer (Application View)

#### User Management
- **Local Users**: Complete user profiles from HashPost PDS
- **External Users**: Lightweight records with AppView-specific data
- **RBAC**: Unified role-based access control for all user types
- **Session Management**: Transparent session handling

#### Data Separation
- **Canonical Data**: Stored on user's home PDS
- **AppView Data**: Forum-specific data (roles, permissions, activity)
- **Event Processing**: Real-time updates from PDS events
- **Caching**: Optimized data access patterns

## Federation Model

### Data Sovereignty

```
User's Home PDS (Canonical Data)
├── Profile Information
├── Posts and Comments
├── Social Graph
└── Authentication

HashPost AppView (Forum Data)
├── Forum Roles and Permissions
├── Subforum Subscriptions
├── Moderation Actions
└── Forum-Specific Activity
```

### Authentication Flow

#### Local Users
```
User → HashPost PDS → Password Auth → HashPost AppView → Forum Access
```

#### External Users
```
User → External PDS → JWT Token → HashPost AppView → Forum Access
```

## Implementation Details

### PDS Federation

#### External PDS Client
- **PDS Discovery**: Resolve PDS endpoints from DID documents
- **Cross-PDS Auth**: Authenticate users against their home PDS
- **Token Validation**: Validate JWT tokens from any PDS
- **Public Key Management**: Fetch and cache PDS public keys

#### Multi-PDS Token Validation
- **Token Parsing**: Extract issuer, user info, and claims
- **Public Key Fetching**: Retrieve and cache PDS public keys
- **Signature Validation**: Validate tokens using appropriate keys
- **Claims Verification**: Verify expiration, audience, scope

### AppView Federation

#### User Record Management
- **Automatic Creation**: Create lightweight records for external users
- **Data Separation**: Store only AppView-specific data
- **Session Tracking**: Update last seen time for external users
- **Role Assignment**: Assign default roles and permissions

#### RBAC Federation
- **Unified Permissions**: Same permission system for all users
- **Role Inheritance**: External users can inherit roles from PDS
- **Forum-Specific Roles**: HashPost-specific roles and permissions
- **Cross-PDS RBAC**: Support for PDS-specific role inheritance

## Database Architecture

### PDS Database
```sql
-- External PDS OAuth registrations
CREATE TABLE external_pds_clients (
    id UUID PRIMARY KEY,
    pds_endpoint VARCHAR(500) NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    client_secret TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(pds_endpoint)
);
```

### AppView Database
```sql
-- Enhanced user table with federation support
ALTER TABLE appview_users ADD COLUMN pds_source VARCHAR(500);
ALTER TABLE appview_users ADD COLUMN is_local BOOLEAN DEFAULT TRUE;
ALTER TABLE appview_users ADD COLUMN last_seen_at TIMESTAMPTZ;
```

## Configuration

### Federation Settings
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

## Security Model

### Authentication Security
1. **Token Validation**: Proper signature verification for all tokens
2. **Public Key Caching**: Secure caching of PDS public keys
3. **Session Management**: Secure session handling across PDS sources
4. **Audit Logging**: Comprehensive logging of federation activities

### Data Security
1. **Data Isolation**: Proper separation of local and external user data
2. **Minimal Storage**: Store only necessary data locally
3. **Access Control**: Proper RBAC for all user types
4. **Encryption**: Secure data transmission and storage

## Performance Considerations

### Caching Strategy
1. **Public Key Cache**: Cache PDS public keys for 24 hours
2. **User Record Cache**: Cache user records in memory
3. **Token Validation**: Optimize token validation process
4. **Database Queries**: Efficient queries for federated users

### Optimization
1. **Lazy Loading**: Load user data on demand
2. **Batch Operations**: Batch user record updates
3. **Connection Pooling**: Efficient database connections
4. **Query Optimization**: Optimized SQL queries

## Monitoring and Observability

### Metrics
- Federation authentication rate
- Token validation success/failure rate
- Public key cache hit rate
- User record creation rate
- Cross-PDS API response times

### Logging
- Federation authentication attempts
- Token validation details
- User record operations
- Security events and anomalies
- Performance metrics

### Alerting
- High authentication failure rates
- Public key cache misses
- Database performance issues
- Security anomalies

## Benefits

### For Users
1. **Data Sovereignty**: Users control their primary data
2. **PDS Choice**: Users can choose their PDS provider
3. **Portability**: Users can migrate between PDS providers
4. **Federation**: Access multiple applications with one identity

### For HashPost
1. **Reduced Load**: Less load on HashPost's PDS
2. **Scalability**: Better scalability through federation
3. **Compatibility**: Works with any atproto-compliant PDS
4. **Innovation**: Focus on forum features, not PDS infrastructure

### For the Ecosystem
1. **Decentralization**: True decentralized social networking
2. **Interoperability**: Standard atproto protocol compliance
3. **Innovation**: Encourages PDS provider innovation
4. **Resilience**: Distributed architecture reduces single points of failure

## Future Enhancements

### Advanced Federation
1. **OAuth Flow**: Complete OAuth implementation for external PDS
2. **PDS Discovery**: Automatic PDS endpoint resolution
3. **Data Sync**: Bidirectional data synchronization
4. **Advanced RBAC**: PDS-specific role inheritance

### Performance Improvements
1. **Enhanced Caching**: More sophisticated caching strategies
2. **Connection Pooling**: Optimized database connections
3. **Query Optimization**: Advanced SQL query optimization
4. **Load Balancing**: Intelligent load balancing for federation

### Security Enhancements
1. **Advanced Token Validation**: More sophisticated token validation
2. **Public Key Rotation**: Support for public key rotation
3. **Audit Logging**: Enhanced audit logging and monitoring
4. **Security Hardening**: Additional security measures

## Conclusion

HashPost's federation architecture provides a solid foundation for decentralized social networking while maintaining the rich forum functionality that users expect. By supporting both hosted and external PDS servers, HashPost enables true user choice and data sovereignty while providing a seamless user experience.

The architecture is designed to be:
- **Scalable**: Efficient handling of federated users
- **Secure**: Proper authentication and data protection
- **Performant**: Optimized for federation workloads
- **Extensible**: Ready for future enhancements
- **Compliant**: Full AT Protocol specification compliance
