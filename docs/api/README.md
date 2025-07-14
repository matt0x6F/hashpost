# API Documentation

This directory contains documentation related to HashPost's API system, authentication, and integration.

## Overview

The API provides RESTful endpoints for all HashPost functionality, with comprehensive authentication and authorization systems supporting both JWT tokens and API keys.

## Documentation

### [API Documentation](documentation.md)
Complete API reference including:
- All API endpoints with examples
- Request/response schemas
- Error handling and status codes
- Rate limiting and pagination
- WebSocket endpoints
- Integration patterns

### [Authentication Guide](authentication.md)
Comprehensive authentication implementation:
- JWT authentication flow
- API key management
- Cookie configuration
- Security features and MFA
- Troubleshooting guide
- Token refresh mechanisms

### [API Keys](keys.md)
API key management and usage:
- API key structure and permissions
- Creation and validation
- Security features
- Best practices
- Integration examples
- Key rotation strategies

## Quick Start

1. **Start with**: [API Documentation](documentation.md) - Complete API reference
2. **Then read**: [Authentication Guide](authentication.md) - Auth implementation
3. **Reference**: [API Keys](keys.md) - API key management

## API Categories

### Authentication & Authorization
- User registration and login
- JWT token management
- API key authentication
- Role-based access control

### Content Management
- Posts and comments
- Voting and moderation
- File uploads and media
- Search and discovery

### User Management
- User profiles and preferences
- Pseudonym management
- Subforum subscriptions
- Direct messaging

### Administrative
- Subforum management
- User moderation
- System configuration
- Audit and logging

## Related Documentation

- [RBAC Documentation](../rbac/) - Access control and permissions
- [Security Documentation](../security/) - Cryptographic foundations
- [Database Documentation](../database/) - Data models and operations 