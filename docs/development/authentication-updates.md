# Authentication System Updates

## Overview
The authentication system has been updated to provide more granular control and better activity tracking.

## Changes

### Middleware Updates
- Removed global authentication middleware from server initialization
- Authentication is now handled per-route using AuthInput structs
- Added pseudonym activity tracking across all authenticated operations

### Per-Route Authentication
Instead of applying authentication globally, each protected route now uses the `AuthInput` struct to handle authentication consistently:

```go
type AuthInput struct {
    Authorization string `header:"Authorization" required:"true"`
}
```

### Activity Tracking
All authenticated operations now update the pseudonym's `last_active` timestamp to track user engagement and activity patterns. This includes:
- Content creation and modification
- Rule management operations
- Message sending
- Search operations
- Moderation actions

## Migration Notes
- Existing authentication flows remain unchanged
- New routes use AuthInput structs for consistent authentication handling
- Activity tracking is automatically applied to all authenticated endpoints
- No changes required for existing client applications

## Benefits
- **Better Security**: Authentication can be applied selectively to routes
- **Activity Monitoring**: Track user engagement patterns
- **Audit Trail**: Maintain logs of user activity for compliance
- **Performance**: Avoid unnecessary authentication checks on public routes

## Implementation Details
The `UpdateLastActive` method is called on the pseudonym DAO after successful operations. If this update fails, the operation continues but logs the error for monitoring purposes.
