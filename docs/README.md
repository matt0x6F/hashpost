# HashPost Documentation

## Overview

HashPost is a Reddit-like social media platform that uses Identity-Based Encryption (IBE) to provide pseudonymous user profiles while maintaining administrative accountability. This platform balances user privacy with the need for effective moderation and compliance through a simplified single-user system with comprehensive role-based access control.

## Architecture Overview

### Core Design Principles

1. **Privacy by Design**: Users interact through pseudonymous profiles, with real identities encrypted and only accessible to authorized personnel
2. **Single-User System**: All users exist in a unified system with role-based capabilities rather than separate administrative accounts
3. **Pseudonym-Based Access Control**: Permissions are managed at multiple levels: user, pseudonym, subforum, and platform
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

### Permission Hierarchy

The permission system operates at multiple levels:

| Level | Description | Examples |
|-------|-------------|----------|
| **User Level** | Basic account management and authentication | Login, profile updates, session management |
| **Pseudonym Level** | Content creation and personal pseudonym management | `create_content`, `vote`, `message`, `report`, `manage_own_pseudonyms` |
| **Subforum Level** | Moderation capabilities specific to subforums | `moderate_content`, `ban_users`, `manage_moderators` (dynamically assigned) |
| **Platform Level** | Administrative capabilities across all subforums | `correlate_identities`, `access_all_pseudonyms`, `system_admin` |

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
- **Multiple Pseudonyms**: Users can have multiple distinct personas under a single account, each with their own roles and capabilities
- **Pseudonym Management**: Create, switch, and deactivate pseudonyms as needed
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
- **Dynamic Role Assignment**: Moderator role is automatically assigned when accessing subforums with moderation capabilities

### For Administrators
- **Identity Correlation**: Full identity correlation for platform-wide investigations
- **System Administration**: User management and platform configuration
- **Legal Compliance**: Handle court orders and legal requests
- **Cross-Platform Access**: Investigate issues across multiple subforums

## 📚 **Documentation Structure**

This documentation is organized into logical categories:

### 🔐 **Security & Access Control**
- **[RBAC Documentation](rbac/)**: Complete Role-Based Access Control system with pseudonym-based permissions
- **[Security Documentation](security/)**: IBE, cryptography, and security analysis

### 🗄️ **Database & Data**
- **[Database Schema](database/schema.md)**: Complete database schema with pseudonym-based role and capability management
- **[Database Operations](database/operations.md)**: Common operations, maintenance, and best practices
- **[Database ERD](database/erd.puml)**: Entity Relationship Diagram

### 🛠️ **Development & Operations**
- **[Development Setup](development/setup.md)**: Docker development environment and setup
- **[CORS Configuration](development/cors.md)**: Cross-Origin Resource Sharing setup

### 🎯 **Features & Functionality**
- **[Comment Workflow](features/comments.md)**: Comment system implementation 