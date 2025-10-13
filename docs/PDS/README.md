# HashPost PDS (Personal Data Server)

## Overview

The HashPost PDS is a fully compliant atproto Personal Data Server implementation using the Bluesky Indigo Go libraries. It provides core data storage and authentication services for the HashPost platform, handling all atproto protocol compliance.

## Architecture

The PDS follows the atproto specification for Personal Data Servers:
- **Purpose**: Pure atproto protocol compliance only
- **API**: Standard atproto endpoints (`/xrpc/com.atproto.*`) only
- **Authentication**: DID-based authentication system using Indigo packages
- **Data Storage**: Canonical atproto records in separate database (`hashpost_pds_dev`)
- **Business Logic**: None - pure protocol compliance

## Key Components

### Authentication System
- **DID Resolution**: Resolves DIDs to identity documents via PLC directory
- **Handle Resolution**: Resolves handles to DIDs via DNS TXT records
- **Session Management**: Generates and validates access/refresh tokens
- **Environment Switching**: Mock directory for development, real directory for production

### Database Layer
- **Schema**: Canonical atproto records with proper URI storage
- **Queries**: SQLC-generated type-safe database operations
- **Migrations**: Automated schema management

### Event Publishing
- **Event Generation**: Publishes atproto events to NATS JetStream
- **Event Types**: Record created/updated/deleted, identity resolved, session created
- **Integration**: Feeds events to AppView for denormalized data storage

## API Endpoints

### Repository Endpoints
- `com.atproto.repo.createRecord` - Create new records
- `com.atproto.repo.getRecord` - Retrieve records by URI
- `com.atproto.repo.listRecords` - List records in repository
- `com.atproto.repo.putRecord` - Update existing records
- `com.atproto.repo.deleteRecord` - Delete records

### Server Endpoints
- `com.atproto.server.createSession` - Authenticate and create session
- `com.atproto.server.getSession` - Get current session info
- `com.atproto.server.refreshSession` - Refresh access tokens
- `com.atproto.server.deleteSession` - End session
- `com.atproto.server.createAccount` - User registration

### Identity Endpoints
- `com.atproto.identity.resolveHandle` - Resolve handle to DID

## Documentation Structure

### Authentication
- [DID Resolution](Authentication/did-resolution.md) - DID resolution using Bluesky Indigo
- [JWT Tokens](Authentication/jwt-tokens.md) - ES256K signing and token validation
- [Session Management](Authentication/session-management.md) - Session lifecycle and storage

### Database
- [Schema](Database/schema.md) - PDS database tables and relationships
- [Queries](Database/queries.md) - SQLC queries and operations

### Events
- [Event Publishing](Events/event-publishing.md) - Event generation and publishing
- [NATS Configuration](Events/nats-configuration.md) - Stream setup and configuration

## Quick Start

1. **Environment Setup**: Configure `ENVIRONMENT` variable (development/production)
2. **Database**: Ensure PostgreSQL is running with PDS database
3. **NATS**: Start NATS JetStream for event publishing
4. **Start Server**: Run `hashpost-pds` binary

## References

- [Atproto Protocol Specification](https://atproto.com/specs/atp)
- [Repository Protocol](https://atproto.com/specs/repository)
- [Server Protocol](https://atproto.com/specs/server)
- [Identity Protocol](https://atproto.com/specs/identity)
