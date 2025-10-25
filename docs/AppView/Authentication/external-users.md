# External Users in AppView

## Overview

HashPost's AppView system now supports users from external PDS servers while maintaining full forum functionality. External users have lightweight local records that store only AppView-specific data.

## User Data Model

### Local Users (HashPost PDS)
- Complete user profile stored locally
- Password authentication
- Full data sovereignty within HashPost

### External Users (External PDS)
- Lightweight records with minimal data
- JWT token authentication from home PDS
- AppView-specific data only (roles, permissions, forum activity)

## AppViewUser Structure

```go
type AppViewUser struct {
    ID           uuid.UUID  `json:"id"`
    DID          string     `json:"did"`
    Handle       string     `json:"handle"`
    DisplayName  string     `json:"display_name"`
    Bio          string     `json:"bio"`
    AvatarURL    string     `json:"avatar_url"`
    CreatedAt    time.Time  `json:"created_at"`
    UpdatedAt    time.Time  `json:"updated_at"`
    PostCount    int        `json:"post_count"`
    CommentCount int        `json:"comment_count"`
    Reputation   int        `json:"reputation"`
    PDSSource    *string    `json:"pds_source,omitempty"`    // PDS endpoint for external users
    IsLocal      bool       `json:"is_local"`               // Whether user is on HashPost PDS
    LastSeenAt   *time.Time `json:"last_seen_at,omitempty"` // Last authentication time
}
```

## Authentication Flow

### Token Validation

1. **Multi-PDS Support**: Validate tokens from any atproto PDS
2. **Public Key Fetching**: Retrieve and cache PDS public keys
3. **Signature Verification**: Validate token signatures
4. **Claims Verification**: Verify expiration, audience, scope

### User Record Management

1. **Automatic Creation**: Create lightweight records for external users
2. **Data Separation**: Store only AppView-specific data
3. **Session Tracking**: Update last seen time for external users
4. **Role Assignment**: Assign default roles and permissions

## RBAC Integration

### Role-Based Access Control

External users receive default roles:

- **Basic User**: Standard forum permissions
- **Subscriber**: Access to subscribed subforums
- **Moderator**: Moderation permissions (if assigned)

### Permission System

The RBAC system works transparently with external users:

- **Permission Checks**: Same logic for local and external users
- **Role Inheritance**: External users can inherit roles from their PDS
- **Forum-Specific Roles**: HashPost-specific roles and permissions

## Event Processing

### Identity Resolution

When external users authenticate:

1. **Event Generation**: PDS publishes identity resolution events
2. **User Creation**: AppView creates lightweight user records
3. **Role Assignment**: Assign default roles and permissions
4. **Session Tracking**: Update last seen time

### Data Flow

```
External PDS → JWT Token → HashPost AppView → Lightweight Record → Forum Access
```

## Database Operations

### User Creation

```go
// Automatic lightweight record creation
user := &AppViewUser{
    ID:          uuid.New(),
    DID:         did,
    Handle:      handle,
    DisplayName: handle,
    CreatedAt:   time.Now(),
    UpdatedAt:   time.Now(),
    IsLocal:     false,
    PDSSource:   &issuer,
    LastSeenAt:  &now,
}
```

### User Updates

- **Last Seen**: Update authentication time
- **Profile Data**: Sync basic profile information
- **Role Changes**: Update permissions and roles

## Configuration

### AppView Settings

```yaml
appview:
  external_users:
    enabled: true
    auto_create: true
    default_roles:
      - "user"
      - "subscriber"
```

### Environment Variables

```bash
APPVIEW_EXTERNAL_USERS_ENABLED=true
APPVIEW_AUTO_CREATE_USERS=true
```

## API Endpoints

### User Management

- `GET /api/v1/auth/me` - Get current user info
- `POST /api/v1/auth/external` - External PDS authentication
- `GET /api/v1/users/{did}` - Get user profile

### PDS Discovery

- `GET /api/v1/pds/discover?identifier={handle_or_did}` - Discover user's PDS
- `GET /api/v1/pds/info?endpoint={url}` - Get PDS server info

## Security Considerations

### Token Security

1. **Signature Validation**: Verify all JWT signatures
2. **Public Key Caching**: Secure caching of PDS public keys
3. **Token Expiration**: Respect token expiration times
4. **Audience Verification**: Verify token audience claims

### Data Security

1. **Minimal Storage**: Store only necessary data locally
2. **Data Isolation**: Separate local and external user data
3. **Access Control**: Proper RBAC for all user types
4. **Audit Logging**: Log all external user activities

## Performance Considerations

### Caching

1. **Public Key Cache**: Cache PDS public keys for 24 hours
2. **User Record Cache**: Cache user records in memory
3. **Token Validation**: Optimize token validation process
4. **Database Queries**: Efficient queries for external users

### Optimization

1. **Lazy Loading**: Load user data on demand
2. **Batch Operations**: Batch user record updates
3. **Connection Pooling**: Efficient database connections
4. **Query Optimization**: Optimized SQL queries

## Monitoring and Logging

### Metrics

- External user authentication rate
- Token validation success/failure rate
- Public key cache hit rate
- User record creation rate

### Logging

- Authentication attempts and results
- Token validation details
- User record operations
- Security events and anomalies

## Troubleshooting

### Common Issues

1. **Token Validation Failures**: Check PDS public key availability
2. **User Creation Errors**: Verify database permissions
3. **Role Assignment Issues**: Check RBAC configuration
4. **Performance Problems**: Monitor cache hit rates

### Debug Information

- Enable debug logging for external user operations
- Monitor token validation metrics
- Check public key cache status
- Verify user record creation

## Future Enhancements

1. **Advanced RBAC**: PDS-specific role inheritance
2. **Data Synchronization**: Bidirectional data sync
3. **Performance Optimization**: Enhanced caching strategies
4. **Security Hardening**: Additional security measures
5. **Monitoring**: Advanced metrics and alerting
