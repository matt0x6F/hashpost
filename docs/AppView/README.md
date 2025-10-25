# HashPost AppView

## Overview

The HashPost AppView is a stateful web application that provides business logic, persistence layer, and custom endpoints for the HashPost platform. It consumes events from the PDS and maintains denormalized data optimized for user-facing queries.

## Architecture

The AppView follows the atproto AppView pattern:
- **Purpose**: Full stateful web application with business logic and persistence
- **API**: Custom forum endpoints (`/api/v1/*`) + web UI routes
- **Authentication**: User session management and authorization
- **Data Storage**: Denormalized/aggregated data in separate database (`hashpost_appview_dev`)
- **Business Logic**: All forum logic, RBAC, moderation, voting, ownership models
- **Stateful Design**: Maintains application state and business logic

## Key Components

### Authentication System
- **Identity Resolution**: DID-to-handle resolution with caching
- **RBAC System**: Role-based access control and permissions
- **Session Management**: User session validation and authorization

### Database Layer
- **Schema**: Denormalized tables optimized for read-heavy operations
- **Queries**: SQLC-generated type-safe database operations
- **Migrations**: Automated schema management

### Event Processing
- **Event Consumption**: Consumes atproto events from PDS via NATS JetStream (limited current usage)
- **Data Denormalization**: Stores denormalized data from events
- **Error Handling**: Retry logic, exponential backoff, idempotency, dead letter queue

## API Endpoints

### Authentication Endpoints
- `POST /api/v1/auth/login` - User authentication
- `POST /api/v1/auth/register` - User registration
- `GET /api/v1/auth/me` - Get current user session
- `POST /api/v1/auth/logout` - User logout

### Content Endpoints
- `GET /api/v1/posts` - List posts
- `POST /api/v1/posts` - Create post
- `GET /api/v1/posts/{id}` - Get post by ID
- `PUT /api/v1/posts/{id}` - Update post
- `DELETE /api/v1/posts/{id}` - Delete post
- `GET /api/v1/posts/{id}/vote` - Vote on post
- `GET /api/v1/posts/{id}/user-vote` - Get user's vote on post
- `GET /api/v1/posts/{id}/metrics` - Get post metrics
- `GET /api/v1/posts/{id}/moderation` - Get post moderation state
- `GET /api/v1/comments` - List comments
- `POST /api/v1/comments` - Create comment
- `GET /api/v1/comments/{id}` - Get comment by ID
- `PUT /api/v1/comments/{id}` - Update comment
- `DELETE /api/v1/comments/{id}` - Delete comment
- `GET /api/v1/comments/{id}/vote` - Vote on comment
- `GET /api/v1/subforums` - List subforums
- `POST /api/v1/subforums` - Create subforum
- `GET /api/v1/subforums/{slug}` - Get subforum by slug

### User Management Endpoints
- `GET /api/v1/me/roles` - Get my roles
- `GET /api/v1/me/permissions` - Get my permissions
- `GET /api/v1/me/subscriptions` - Get my subscriptions
- `GET /api/v1/me/moderated-subforums` - Get subforums I moderate

### Subscription Endpoints
- `POST /api/v1/subforums/{slug}/subscribe` - Subscribe to subforum
- `DELETE /api/v1/subforums/{slug}/subscribe` - Unsubscribe from subforum
- `GET /api/v1/subforums/{slug}/subscribe` - Get subscription status
- `GET /api/v1/subforums/{slug}/subscribers` - List subforum subscribers

### RBAC Management Endpoints
- `POST /api/v1/admin/assign-role` - Assign role to user
- `GET /api/v1/admin/roles` - List available roles
- `GET /api/v1/admin/user-roles` - Get user roles
- `POST /api/v1/admin/revoke-role` - Revoke role from user
- `GET /api/v1/admin/permissions` - List permissions
- `GET /api/v1/admin/users` - List all users
- `GET /api/v1/subforums/{slug}/members` - List subforum members
- `POST /api/v1/subforums/{slug}/members/{user_did}/roles` - Assign subforum role
- `DELETE /api/v1/subforums/{slug}/members/{user_did}/roles` - Revoke subforum role

### PDS Management Endpoints
- `GET /api/v1/admin/pds/servers` - List PDS servers
- `GET /api/v1/admin/pds/{endpoint}` - Get PDS server details
- `GET /api/v1/admin/pds/{endpoint}/users` - List PDS server users

## Documentation Structure

### Authentication
- [Identity Resolution](Authentication/identity-resolution.md) - DID-to-handle resolution with caching
- [RBAC System](Authentication/rbac.md) - Role-based access control and permissions

### Database
- [Schema](Database/schema.md) - AppView denormalized tables and relationships
- [Queries](Database/queries.md) - SQLC queries for posts, subforums, comments, RBAC

### Events
- [Event Consumption](Events/event-consumption.md) - NATS subscription and message processing
- [Error Handling](Events/error-handling.md) - Retry logic, exponential backoff, idempotency

## Quick Start

1. **Environment Setup**: Configure `ENVIRONMENT` variable (development/production)
2. **Database**: Ensure PostgreSQL is running with AppView database
3. **NATS**: Start NATS JetStream for event consumption
4. **Start Server**: Run `hashpost-appview` binary

## References

- [AppView Implementation](internal/appview/)
- [OpenAPI Specification](internal/appview/openapi.yaml)
- [Database Schema](internal/database/migrations/appview/)
- [Event Processing](internal/appview/events.go)
