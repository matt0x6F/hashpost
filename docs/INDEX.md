# HashPost Documentation Index

## Overview

This index provides a comprehensive overview of all HashPost documentation, organized by category and purpose. Use this guide to quickly find the information you need.

## 📚 **Core Documentation**

### [README.md](README.md)
**Purpose**: Main project overview and getting started guide
**Audience**: New users, developers, stakeholders
**Content**: 
- Project overview and architecture
- Key features and user roles
- Quick start guide
- Documentation structure
- Deployment considerations

## 🔐 **Security & Access Control**

### [RBAC Documentation](rbac/)
**Purpose**: Complete Role-Based Access Control system documentation
**Audience**: Security engineers, developers, system administrators
**Content**:
- [RBAC Overview](rbac/rbac-overview.md) - Complete RBAC system documentation
- [RBAC Usage Examples](rbac/rbac-usage-example.md) - Implementation patterns
- [Role Keys and Site Roles](rbac/role-keys-and-site-roles.md) - Role key management
- [User Roles](rbac/user-roles.md) - Role definitions and permissions

### [Security Documentation](security/)
**Purpose**: Security systems, cryptography, and security analysis
**Audience**: Security engineers, cryptography experts
**Content**:
- [Identity-Based Encryption](security/identity-based-encryption.md) - IBE system and cryptographic foundations
- [IBE Key Management](security/ibe-key-management.md) - Enhanced key generation and management
- [IBE Security Enhancement Summary](security/ibe-security-enhancement-summary.md) - Security enhancement overview
- [Domain Separation Security Analysis](security/domain-separation-security-analysis.md) - Security analysis
- [Key Rotation Infrastructure](security/key-rotation-infrastructure.md) - Key rotation strategies

## 📡 **API & Integration**

### [API Documentation](api/)
**Purpose**: Complete API reference and authentication implementation
**Audience**: API developers, frontend developers, system administrators
**Content**:
- [API Documentation](api/documentation.md) - Complete API reference with examples
- [Authentication Guide](api/authentication.md) - JWT and API key authentication
- [API Keys](api/keys.md) - API key management and usage

## 🗄️ **Database & Data**

### [Database Documentation](database/)
**Purpose**: Database schema, operations, and administration
**Audience**: Database administrators, backend developers, DevOps engineers
**Content**:
- [Database Schema](database/schema.md) - Complete database schema with role-based access control
- [Database Operations](database/operations.md) - Common operations, maintenance, and best practices
- [Database ERD](database/erd.puml) - Entity Relationship Diagram (PlantUML source)
- [ERD](database/ERD.md) - Instructions for generating and viewing the ERD

## 🛠️ **Development & Operations**

### [Development Documentation](development/)
**Purpose**: Development environment setup and operational procedures
**Audience**: Developers, DevOps engineers
**Content**:
- [Development Setup](development/setup.md) - Development environment setup and workflows
- [CORS Configuration](development/cors.md) - Cross-Origin Resource Sharing configuration

## 🎯 **Features & Functionality**

### [Features Documentation](features/)
**Purpose**: Specific HashPost features and functionality
**Audience**: Developers, product managers, users
**Content**:
- [Comment Workflow](features/comments.md) - Comment system implementation and usage

## 🔍 **Documentation by Use Case**

### For Security Engineers
1. **Start with**: [RBAC Overview](rbac/rbac-overview.md) - Complete access control system
2. **Then read**: [Identity-Based Encryption](security/ibe.md) - Cryptographic foundations
3. **Reference**: [Authentication Guide](api/authentication.md) - Auth implementation
4. **Explore**: [Domain Separation Security Analysis](security/domain-separation-security-analysis.md) - Security analysis

### For New Developers
1. **Start with**: [README.md](README.md) - Project overview
2. **Then read**: [Development Setup](development/setup.md) - Environment setup
3. **Reference**: [RBAC Overview](rbac/rbac-overview.md) - Access control system
4. **Explore**: [API Documentation](api/documentation.md) - API reference

### For API Integration
1. **Start with**: [API Documentation](api/documentation.md) - Complete API reference
2. **Then read**: [Authentication Guide](api/authentication.md) - Auth flows
3. **Reference**: [RBAC Usage Examples](rbac/rbac-usage-example.md) - Implementation examples

### For Database Administration
1. **Start with**: [Database Schema](database/schema.md) - Schema reference
2. **Then read**: [Database Operations](database/operations.md) - Operations guide
3. **Reference**: [Database ERD](database/erd.puml) - Visual schema

### For DevOps Engineers
1. **Start with**: [Development Setup](development/setup.md) - Environment setup
2. **Then read**: [Database Operations](database/operations.md) - Operations guide
3. **Reference**: [IBE Key Management](security/key-management.md) - Key management

### For System Administrators
1. **Start with**: [RBAC Overview](rbac/rbac-overview.md) - Access control system
2. **Then read**: [Role Keys and Site Roles](rbac/role-keys-and-site-roles.md) - Role management
3. **Reference**: [User Roles](rbac/user-roles.md) - Role definitions

### For Product Managers
1. **Start with**: [README.md](README.md) - Project overview
2. **Then read**: [Features Documentation](features/) - Feature functionality
3. **Reference**: [User Roles](rbac/user-roles.md) - Role definitions

## 📋 **Documentation Organization**

### Directory Structure

```
docs/
├── README.md                    # Main project overview
├── INDEX.md                     # This documentation index
├── rbac/                        # Role-Based Access Control
│   ├── README.md               # RBAC documentation overview
│   ├── rbac-overview.md        # Complete RBAC system
│   ├── rbac-usage-example.md   # Implementation examples
│   ├── role-keys-and-site-roles.md
│   └── user-roles.md
├── security/                    # Security and cryptography
│   ├── README.md               # Security documentation overview
│   ├── ibe.md                  # Identity-Based Encryption
│   ├── key-management.md       # IBE key management
│   ├── enhancement-summary.md  # Security enhancement summary
│   ├── domain-separation-security-analysis.md
│   └── key-rotation-infrastructure.md
├── api/                         # API and integration
│   ├── README.md               # API documentation overview
│   ├── documentation.md        # Complete API reference
│   ├── authentication.md       # Auth implementation
│   └── keys.md                # API key management
├── database/                    # Database and data
│   ├── README.md               # Database documentation overview
│   ├── schema.md               # Complete schema
│   ├── operations.md           # Operations guide
│   ├── erd.puml               # ERD diagram
│   └── ERD.md          # ERD instructions
├── development/                 # Development and operations
│   ├── README.md               # Development documentation overview
│   ├── setup.md                # Environment setup
│   └── cors.md                # CORS setup
└── features/                    # Feature documentation
    ├── README.md               # Features documentation overview
    └── comments.md             # Comment system
```

### Benefits of Directory Organization

1. **Logical Grouping**: Related documentation is grouped together
2. **Easy Navigation**: Clear directory structure for finding information
3. **Scalable**: Easy to add new documentation in appropriate categories
4. **Contextual**: Each directory has its own README explaining the content
5. **Cross-References**: Clear relationships between different documentation areas

### Documentation Standards

#### File Naming
- Use kebab-case for file names
- Include descriptive names that indicate content
- Use consistent extensions (.md for markdown, .puml for PlantUML)

#### Content Structure
- Start with a clear purpose statement
- Include audience information
- Use consistent heading levels
- Include code examples where appropriate
- Provide troubleshooting sections

#### Maintenance
- Update documentation when code changes
- Review for accuracy quarterly
- Remove outdated information
- Consolidate duplicate content 