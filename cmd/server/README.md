# HashPost Server CLI

This is the main server binary for HashPost with built-in CLI functionality using Huma's CLI framework.

## Usage

### Starting the Server

```bash
# Start with default settings (port 8888)
./server

# Start on a specific port
./server --port 9000

# Start on a specific host and port
./server --host localhost --port 9000

# Enable debug logging
./server --debug
```

### Environment Variables

You can also use environment variables with the `HASHPOST_` prefix:

```bash
# Set port via environment variable
export HASHPOST_PORT=9000
./server

# Set host via environment variable
export HASHPOST_HOST=0.0.0.0
./server
```

## Commands

### Create Admin User

Create a new admin user with specified role and capabilities:

```bash
./server create-admin
```

This command will interactively prompt for:
- Email address
- Password (minimum 8 characters)
- Display name (optional)
- Admin role (platform_admin, trust_safety, legal_team)
- Admin scope (optional)
- MFA enabled (y/n)

#### Admin Roles

- **platform_admin**: Full system administration
  - Capabilities: system_admin, user_management, correlate_identities, access_private_subforums, cross_platform_access, system_moderation

- **trust_safety**: Trust and safety operations
  - Capabilities: correlate_identities, cross_platform_access, system_moderation, harassment_investigation

- **legal_team**: Legal compliance operations
  - Capabilities: correlate_identities, legal_compliance, court_orders, cross_platform_access

### OpenAPI Specification

Generate the OpenAPI specification:

```bash
./server openapi
```

This outputs the complete OpenAPI 3.1 specification in YAML format.

## Development

### Building

```bash
go build ./cmd/server
```

### Running in Development

```bash
make dev
```

This will start the development environment with Docker Compose and run the server.

## Configuration

The server uses the configuration system defined in `internal/config/`. Make sure your configuration files are properly set up before running the server.

## Database Requirements

The server requires a PostgreSQL database to be running. In development, this is handled by Docker Compose. For production, ensure your database connection is properly configured.

## Security Notes

- Admin passwords are hashed using SHA-256 before storage
- MFA is enabled by default for admin users
- Admin usernames are automatically generated if not provided
- All admin operations are logged for audit purposes 