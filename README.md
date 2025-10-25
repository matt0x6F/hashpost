# HashPost

A decentralized forum platform built on the AT Protocol, supporting both hosted and external Personal Data Servers (PDS).

## Features

### Core Forum Functionality
- **Subforums**: Organized discussion categories
- **Posts and Comments**: Rich text discussions
- **Voting System**: Upvote/downvote posts and comments
- **User Management**: Profiles, roles, and permissions
- **Moderation Tools**: Content moderation and user management

### Federation Support (Partial Implementation)
- **Bring Your Own PDS**: Basic support for external PDS servers
- **Hosted PDS**: Easy onboarding with HashPost's PDS
- **Cross-PDS Authentication**: JWT token validation from any atproto PDS
- **Data Sovereignty**: Users control their primary data
- **AT Protocol Compliance**: Full compliance with atproto specifications

## Architecture

HashPost follows a dual-server architecture:

### PDS (Personal Data Server)
- **Purpose**: AT Protocol compliance and canonical data storage including forum tables
- **Authentication**: DID-based authentication with external PDS support
- **Data Storage**: Canonical atproto records plus forum-specific tables (posts, comments, votes, subforums)
- **API**: Standard atproto endpoints (`/xrpc/com.atproto.*`) plus OAuth/DPoP endpoints

### AppView (Application View)
- **Purpose**: Stateful web application providing forum functionality and business logic
- **Authentication**: Multi-PDS token validation
- **Data Storage**: Denormalized data optimized for forum queries
- **API**: Custom forum endpoints (`/api/v1/*`)

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Go 1.21+
- Node.js 18+

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/your-org/hashpost.git
   cd hashpost
   ```

2. **Start the development environment**
   ```bash
   docker-compose up -d
   ```

3. **Generate code**
   ```bash
   task generate:sqlc
   task generate:openapi
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   - PDS API: http://localhost:8080
   - AppView API: http://localhost:8081
   - API Documentation: http://localhost:8081/docs

## External PDS Support

HashPost supports "bring your own PDS" functionality, allowing users to authenticate with any atproto-compliant PDS while maintaining full forum functionality.

### For Users
- **Hosted PDS**: Register directly on HashPost for easy onboarding
- **External PDS**: Use your existing atproto identity to access HashPost
- **Data Control**: Your primary data stays on your chosen PDS
- **Forum Features**: Full access to HashPost's forum functionality

### For Developers
- **Multi-PDS Authentication**: Automatic support for external PDS tokens
- **Lightweight Records**: Minimal local storage for external users
- **Unified RBAC**: Same permission system for all user types
- **Event Processing**: Real-time updates from any PDS

## Configuration

### External PDS Support
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

## Documentation

- [External PDS Support](docs/PDS/Authentication/external-pds-support.md)
- [External Users in AppView](docs/AppView/Authentication/external-users.md)
- [Federation Architecture](docs/Shared/Architecture/federation.md)
- [Development Guide](docs/development/)
- [API Documentation](docs/AppView/README.md)

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Workflow
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `task test:unit`
5. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: Check the [docs](docs/) directory
- **Issues**: Open an issue on GitHub
- **Discussions**: Use GitHub Discussions for questions

## Roadmap

### Current Features
- ✅ Basic forum functionality
- ✅ External PDS support
- ✅ Multi-PDS authentication
- ✅ Federation architecture

### Upcoming Features
- 🔄 OAuth flow for external PDS
- 🔄 Advanced RBAC with PDS role inheritance
- 🔄 Data synchronization
- 🔄 Performance optimizations

## Acknowledgments

- Built on the [AT Protocol](https://atproto.com/)
- Uses [Bluesky Indigo](https://github.com/bluesky-social/indigo) libraries
- Inspired by [Tangled](https://tangled.org/)'s federation approach
