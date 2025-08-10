# RBAC Documentation

This directory contains documentation related to HashPost's Role-Based Access Control (RBAC) system.

## Overview

The RBAC system uses a **unified capability system** that combines traditional role-based permissions with Identity-Based Encryption (IBE) for secure, privacy-preserving operations. All permission checking uses the `PermissionDAO.HasUnifiedCapability()` method to ensure consistent access control across the platform.

## Documentation

### [RBAC Overview](rbac-overview.md)
Complete documentation of the RBAC system including:
- Architecture and core components
- Roles, scopes, and capabilities
- Role key system and IBE integration
- Identity management and pseudonyms
- Access control flow and security features
- Implementation examples and best practices

### [RBAC Setup and Configuration](rbac-setup-and-configuration.md)
Setup guide for the RBAC system:
- Constants usage and best practices
- Database setup examples
- Role key creation and management
- Quick reference for common patterns

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

### [Unified Permission System](unified-permission-system.md)
Comprehensive documentation of the unified capability system:
- System architecture and design
- Migration strategy and current status
- API integration patterns
- Real-world implementation examples
- Best practices and error handling

### [Permission Checking Patterns](permission-checking-patterns.md)
Developer guide for implementing permission checks:
- Standard permission check patterns
- Context usage (global vs subforum-specific)
- Error handling and logging patterns
- Handler structure patterns
- Testing patterns and best practices
- Migration checklist

### [System Architecture Diagrams](system-architecture-diagrams.md)
Visual diagrams of the RBAC system architecture:
- Complete system architecture with all component relationships
- Permission flow showing step-by-step authorization process
- Role hierarchy and domain key mapping
- Security features and cryptographic separation
- Implementation notes and context usage

## Quick Start

1. **Start with**: [RBAC Overview](rbac-overview.md) - Complete system understanding
2. **Visualize**: [System Architecture Diagrams](system-architecture-diagrams.md) - Visual architecture guide
3. **Implementation**: [Permission Checking Patterns](permission-checking-patterns.md) - Developer patterns
4. **Setup**: [RBAC Setup and Configuration](rbac-setup-and-configuration.md) - Configuration guide
5. **Deep dive**: [Unified Permission System](unified-permission-system.md) - System architecture
6. **Reference**: [User Roles](user-roles.md) - Role definitions
7. **Advanced**: [Role Keys and Site Roles](role-keys-and-site-roles.md) - Key management

## Related Documentation

- [Security Documentation](../security/) - IBE and cryptographic foundations
- [API Documentation](../api/) - API authentication and authorization
- [Database Documentation](../database/) - Database schema and operations 