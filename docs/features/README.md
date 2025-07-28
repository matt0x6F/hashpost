# Features Documentation

This directory contains documentation for specific HashPost features and functionality.

## Overview

HashPost provides a comprehensive forum platform with advanced features for content creation, moderation, and user interaction. The platform emphasizes privacy, security, and user experience with pseudonymous interactions and robust community management tools.

## Implementation Status

### ✅ **Implemented Features**
- **Key Rotation Migration System** - Resumable, fault-tolerant IBE key rotation
- **Direct Messaging System** - Private messaging with blocking controls
- **User Blocking System** - Pseudonym and fingerprint-level blocking
- **Authentication & User Management** - Complete user authentication system
- **Content Creation** - Posts, comments, and voting system
- **Communities & Subforums** - Community management and participation with four community types (t/, g/, b/, c/)
- **Moderation & Safety** - Content moderation and user safety features
- **Comments & Discussions** - Comprehensive comment system

### 📋 **Planned Features**
- **Search & Discovery System** - Content and user search capabilities
- **Multi-Factor Authentication** - Enhanced account security
- **Email Verification** - Account verification system
- **Notifications System** - Real-time user notifications
- **Moderator Dashboard** - Advanced moderation tools
- **Audit System** - Comprehensive audit logging
- **Revenue & Monetization** - Advertising and subscription systems

## Documentation

### [Feature Roadmap](roadmap.md)
Comprehensive roadmap of planned features:
- Security and authentication features
- User experience enhancements
- Revenue and monetization systems
- Platform governance improvements
- Implementation timelines and priorities
- Technical considerations and dependencies

### [Roadmap Summary](roadmap-summary.md)
Quick overview of feature status and priorities:
- Implementation status table
- Feature dependencies and relationships
- Priority-based implementation order
- Technical focus areas

### [Authentication & User Management](authentication.md)
User-focused authentication system:
- Account creation and login
- Pseudonym management and switching
- Privacy and security features
- Session management

### [Content Creation](content-creation.md)
Content creation and management:
- Post creation and editing
- Comment system with threading
- Voting and engagement
- Content moderation

### [Communities & Subforums](communities.md)
Community features and management:
- Subforum discovery and subscription
- Community participation
- Privacy-focused community building

### [Moderation & Safety](moderation.md)
Moderation and safety features:
- Content reporting system
- Moderation tools and processes
- User safety and privacy protection

### [Comment Workflow](comments.md)
Comment system implementation and usage:
- Comment creation and management
- Voting and moderation systems
- API endpoints and integration
- UI components and interactions
- Moderation workflows
- Thread management

### [Key Rotation Migration](key-rotation-migration.md)
Advanced key rotation system:
- Resumable migration capabilities
- Fault tolerance and recovery
- Progress tracking and monitoring
- CLI management tools
- Production-ready implementation

## Feature Categories

### Security & Privacy
- IBE-based key rotation system
- Pseudonymous user interactions
- User blocking and privacy controls
- Authentication and session management
- Audit logging and compliance

### Content Creation
- Posts and comments
- Rich text editing
- Media uploads
- Markdown support
- Draft management

### Interaction & Engagement
- Voting and rating
- User subscriptions
- Direct messaging
- User blocking system
- Community participation

### Moderation & Safety
- Content moderation
- User management
- Report handling
- Safety tools
- Blocking and privacy controls

### Platform Management
- Subforum creation
- User administration
- System configuration
- Performance monitoring
- Compliance reporting

## Related Documentation

- [API Documentation](../api/) - Feature API endpoints
- [RBAC Documentation](../rbac/) - Feature access control
- [Database Documentation](../database/) - Feature data models
- [Security Documentation](../security/) - Security features and implementation 