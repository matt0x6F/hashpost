# HashPost Documentation

## Overview

HashPost is a Reddit-like social media platform that uses Identity-Based Encryption (IBE) to provide pseudonymous user profiles while maintaining administrative accountability. This platform balances user privacy with the need for effective moderation and compliance through a simplified single-user system with comprehensive role-based access control.

## Architecture Overview

### Core Design Principles

1. **Privacy by Design**: Users interact through pseudonymous profiles, with real identities encrypted and only accessible to authorized personnel
2. **Single-User System**: All users exist in a unified system with role-based capabilities rather than separate administrative accounts
3. **Role-Based Access Control**: Different user roles have different capabilities and access levels for correlation and administrative functions
4. **Cryptographic Privacy**: Real identities are encrypted and only accessible through role-based keys
5. **Comprehensive Audit Trail**: All administrative activities are logged for compliance and oversight

### System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    HashPost Platform                        │
├─────────────────────────────────────────────────────────────┤
│  Client Applications (Web, Mobile, API)                    │
├─────────────────────────────────────────────────────────────┤
│  API Gateway & Authentication                              │
│  ├─ JWT Authentication (Web & API)                         │
│  └─ API Key Authentication (Programmatic Access)           │
├─────────────────────────────────────────────────────────────┤
│  Application Layer                                         │
│  ├─ User Management                                        │
│  ├─ Content Management                                     │
│  ├─ Moderation System                                      │
│  └─ Correlation Engine                                     │
├─────────────────────────────────────────────────────────────┤
│  Database Layer (PostgreSQL)                               │
│  ├─ User Data & Content                                    │
│  ├─ Encrypted Identity Mappings                            │
│  ├─ Role-Based Keys                                        │
│  └─ Audit Logs                                             │
└─────────────────────────────────────────────────────────────┘
```

### User Roles and Capabilities

| Role | Capabilities | Correlation Access | Scope | Time Window |
|------|-------------|-------------------|-------|-------------|
| **User** | create_content, vote, message, report | none | none | none |
| **Moderator** | moderate_content, ban_users, remove_content, correlate_fingerprints | fingerprint | subforum_specific | 30 days |
| **Subforum Owner** | All moderator + manage_moderators | fingerprint | subforum_specific | 90 days |
| **Trust & Safety** | correlate_identities, cross_platform_access, system_moderation | identity | platform_wide | unlimited |
| **Legal Team** | correlate_identities, legal_compliance, court_orders | identity | platform_wide | unlimited |
| **Platform Admin** | system_admin, user_management, correlate_identities | identity | platform_wide | unlimited |

## Key Features

### For Regular Users
- **Pseudonymous Profiles**: Users interact through display names without revealing real identities
- **Multiple Pseudonyms**: Users can have multiple distinct personas under a single account
- **Content Creation**: Create posts, comments, and engage with community content
- **Voting System**: Upvote/downvote content to influence visibility
- **Direct Messaging**: Private communication between users
- **Subforum Subscriptions**: Follow and participate in communities
- **Privacy Controls**: Manage visibility of karma and messaging preferences

### For Moderators
- **Content Moderation**: Remove inappropriate content and ban users
- **Fingerprint Correlation**: Identify ban evaders within their subforums
- **Report Management**: Review and resolve user reports
- **Community Management**: Manage subforum rules and settings
- **Pseudonymous Moderation**: Moderate under pseudonymous identities

### For Administrators
- **Identity Correlation**: Full identity correlation for platform-wide investigations
- **System Administration**: User management and platform configuration
- **Legal Compliance**: Handle court orders and legal requests
- **Cross-Platform Access**: Investigate issues across multiple subforums

## Documentation Structure

### 📚 **Core Documentation**
- **[API Reference](api-documentation.md)**: Complete API documentation with authentication flows
- **[Database Schema](database-schema.md)**: Complete database schema with role-based access control
- **[Authentication Guide](authentication.md)**: JWT and API key authentication implementation
- **[Identity-Based Encryption](identity-based-encryption.md)**: IBE system and pseudonym generation

### 🛠️ **Development & Operations**
- **[Development Setup](development.md)**: Docker development environment and setup
- **[Database Operations](database-operations.md)**: Common operations, maintenance, and best practices
- **[API Keys](api-keys.md)**: API key management and usage

### 📊 **Visual Documentation**
- **[Database ERD](database-erd.puml)**: Entity Relationship Diagram (PlantUML)
- **[README-ERD](README-ERD.md)**: Instructions for generating and viewing the ERD

### 📝 **Changelog & History**
- **[Changelog](changelog/)**: Historical changes and feature updates

## Getting Started

### Prerequisites
- PostgreSQL 14+ with JSON support
- Go 1.21+ for backend services
- Node.js 18+ for frontend applications
- Redis for caching and session management

### Quick Start
1. **Development Environment**: Run `make dev` to start the Docker development environment
2. **Database Setup**: Database migrations run automatically on container startup
3. **API Access**: API is available at `http://localhost:8888`
4. **Documentation**: Review the authentication and API documentation

### Development Environment
```bash
# Clone the repository
git clone https://github.com/hashpost/hashpost.git
cd hashpost

# Start development environment
make dev

# Access the API
curl http://localhost:8888/health
```

## API Integration

### Authentication Flow
1. **User Registration**: Users register with email/password and receive pseudonym
2. **JWT Authentication**: Standard JWT authentication for user operations
3. **API Key Authentication**: Static API keys for programmatic access
4. **Role-Based Access**: Different roles have different capabilities

### Example API Usage
```javascript
// User authentication
const response = await fetch('/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    email: 'user@example.com',
    password: 'password'
  })
});

// API key authentication
const apiResponse = await fetch('/api/v1/posts', {
  headers: { 'Authorization': 'Bearer your-api-key-here' }
});
```

## Deployment

### Production Considerations
- **Database Security**: Use encrypted connections and secure key management
- **Network Security**: Implement proper firewall rules and access controls
- **Monitoring**: Set up comprehensive logging and monitoring for correlation activities
- **Backup Strategy**: Regular encrypted backups with disaster recovery procedures
- **Compliance**: Ensure GDPR, CCPA, and other privacy regulation compliance

### Environment Variables
```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_NAME=hashpost
DB_USER=hashpost_user
DB_PASSWORD=secure_password

# Security Configuration
JWT_SECRET=your_jwt_secret
IBE_MASTER_KEY=your_ibe_master_key
ENCRYPTION_KEY=your_encryption_key

# API Configuration
API_PORT=8080
API_HOST=0.0.0.0
CORS_ORIGINS=https://hashpost.com,https://www.hashpost.com
```

## Contributing

### Development Guidelines
1. **Privacy First**: Always consider privacy implications of new features
2. **Role-Based Design**: Implement features with appropriate role-based access
3. **Audit Trail**: Ensure all administrative actions are properly logged
4. **Security Review**: All changes must pass security review before deployment

### Code Standards
- Follow Go best practices for backend development
- Use TypeScript for frontend development
- Implement comprehensive testing for all correlation features
- Document all API changes and database modifications

## Support and Community

### Getting Help
- **Documentation**: Start with this README and linked documentation
- **API Reference**: Use the comprehensive API documentation
- **Issues**: Report bugs and feature requests through GitHub issues
- **Discussions**: Join community discussions for questions and ideas

### Security Reporting
For security vulnerabilities, please report them privately to security@hashpost.com. Do not create public issues for security concerns.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

---

This documentation provides a comprehensive overview of the HashPost platform, its architecture, and how to get started with development and deployment. The single-user system with role-based access control provides a balance between simplicity and security while maintaining the privacy protections that make HashPost unique. 