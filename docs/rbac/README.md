# RBAC Documentation

This directory contains documentation related to HashPost's Role-Based Access Control (RBAC) system.

## Overview

The RBAC system combines traditional role-based permissions with Identity-Based Encryption (IBE) for secure, privacy-preserving operations. This ensures users can only access data and perform operations appropriate to their role and scope.

## Documentation

### [RBAC Overview](rbac-overview.md)
Complete documentation of the RBAC system including:
- Architecture and core components
- Roles, scopes, and capabilities
- Role key system and IBE integration
- Identity management and pseudonyms
- Access control flow and security features
- Implementation examples and best practices

### [RBAC Usage Examples](rbac-usage-example.md)
Practical examples of using the RBAC system:
- Constants usage and best practices
- Database setup examples
- Permission checking patterns
- Middleware integration
- Handler implementation examples
- Role key creation and management

### [Role Keys and Site Roles](role-keys-and-site-roles.md)
Detailed documentation of role key management:
- Role key creation and management
- Site role definitions and relationships
- Key scope separation
- Administrative workflows
- Key rotation strategies

### [User Roles](user-roles.md)
User role definitions and permissions:
- Role definitions and hierarchies
- Permission matrices
- Role assignment workflows
- Platform-wide vs subforum-specific roles

## Quick Start

1. **Start with**: [RBAC Overview](rbac-overview.md) - Complete system understanding
2. **Then read**: [RBAC Usage Examples](rbac-usage-example.md) - Implementation patterns
3. **Reference**: [User Roles](user-roles.md) - Role definitions
4. **Advanced**: [Role Keys and Site Roles](role-keys-and-site-roles.md) - Key management

## Related Documentation

- [Security Documentation](../security/) - IBE and cryptographic foundations
- [API Documentation](../api/) - API authentication and authorization
- [Database Documentation](../database/) - Database schema and operations 