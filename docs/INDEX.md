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

### [API Documentation](api-documentation.md)
**Purpose**: Complete API reference with authentication flows
**Audience**: API developers, frontend developers
**Content**:
- All API endpoints with examples
- Authentication flows
- Request/response schemas
- Error handling
- Rate limiting
- WebSocket endpoints

### [Database Schema](database-schema.md)
**Purpose**: Complete database schema with role-based access control
**Audience**: Database administrators, backend developers
**Content**:
- All table definitions
- Relationships and constraints
- Indexes and performance considerations
- Role-based access patterns
- Data types and validation

### [Authentication Guide](authentication.md)
**Purpose**: JWT and API key authentication implementation
**Audience**: Developers, security engineers
**Content**:
- JWT authentication flow
- API key management
- Cookie configuration
- Security features
- MFA implementation
- Troubleshooting guide

### [Identity-Based Encryption](identity-based-encryption.md)
**Purpose**: IBE system and pseudonym generation
**Audience**: Cryptography experts, security engineers
**Content**:
- IBE implementation details
- Pseudonym generation process
- Key management
- Security considerations
- Mathematical foundations

### [IBE Key Management](ibe-key-management.md)
**Purpose**: Enhanced IBE key generation and management
**Audience**: DevOps engineers, security engineers, system administrators
**Content**:
- Enhanced architecture with domain separation
- Command-line key generation
- Production deployment procedures
- Security best practices
- Migration from legacy systems

## 🛠️ **Development & Operations**

### [Development Setup](development.md)
**Purpose**: Development environment setup and workflows
**Audience**: Developers, DevOps engineers
**Content**:
- Docker development environment
- Database management
- Testing strategies
- Common commands
- Troubleshooting guide
- Performance monitoring

### [Database Operations](database-operations.md)
**Purpose**: Common operations, maintenance, and best practices
**Audience**: Database administrators, DevOps engineers
**Content**:
- Migration workflows
- Backup and recovery
- Performance optimization
- Security operations
- Monitoring and alerting
- Troubleshooting

### [API Keys](api-keys.md)
**Purpose**: API key management and usage
**Audience**: API developers, system administrators
**Content**:
- API key structure and permissions
- Creation and validation
- Security features
- Best practices
- Integration examples

## 📊 **Visual Documentation**

### [Database ERD](database-erd.puml)
**Purpose**: Entity Relationship Diagram (PlantUML source)
**Audience**: Database designers, architects
**Content**:
- Visual representation of database schema
- Table relationships
- Primary and foreign keys
- Color-coded sections

### [README-ERD](README-ERD.md)
**Purpose**: Instructions for generating and viewing the ERD
**Audience**: Developers, documentation maintainers
**Content**:
- How to generate the ERD diagram
- Different viewing options
- Troubleshooting
- Diagram structure explanation

## 📝 **Changelog & History**

### [Changelog](changelog/README.md)
**Purpose**: Historical changes and feature updates
**Audience**: Developers, project managers
**Content**:
- Multiple pseudonyms support
- Moderator pseudonym support
- API changes
- Implementation considerations
- Migration guides

## 🔍 **Documentation by Use Case**

### For New Developers
1. **Start with**: [README.md](README.md) - Project overview
2. **Then read**: [Development Setup](development.md) - Environment setup
3. **Reference**: [Authentication Guide](authentication.md) - Auth implementation
4. **Explore**: [API Documentation](api-documentation.md) - API reference

### For API Integration
1. **Start with**: [API Documentation](api-documentation.md) - Complete API reference
2. **Then read**: [Authentication Guide](authentication.md) - Auth flows
3. **Reference**: [API Keys](api-keys.md) - API key management

### For Database Administration
1. **Start with**: [Database Schema](database-schema.md) - Schema reference
2. **Then read**: [Database Operations](database-operations.md) - Operations guide
3. **Reference**: [Database ERD](database-erd.puml) - Visual schema

### For Security Engineers
1. **Start with**: [Authentication Guide](authentication.md) - Auth implementation
2. **Then read**: [Identity-Based Encryption](identity-based-encryption.md) - IBE system
3. **Reference**: [Database Operations](database-operations.md) - Security operations

### For DevOps Engineers
1. **Start with**: [Development Setup](development.md) - Environment setup
2. **Then read**: [Database Operations](database-operations.md) - Operations guide
3. **Reference**: [README.md](README.md) - Deployment considerations

## 📋 **Documentation Maintenance**

### Recently Consolidated
The following documentation has been consolidated to reduce duplication:

#### Authentication Documentation
- **Before**: 3 separate files (jwt-authentication-strategy.md, jwt-implementation.md, huma-cookie-authentication.md)
- **After**: 1 comprehensive file ([authentication.md](authentication.md))

#### Development Documentation
- **Before**: 1 file (docker-development.md)
- **After**: 1 comprehensive file ([development.md](development.md))

#### Changelog Documentation
- **Before**: 3 separate files (api-multiple-pseudonyms-support.md, multiple-pseudonyms-support.md, moderator-pseudonym-update.md)
- **After**: 1 comprehensive file ([changelog/README.md](changelog/README.md))

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

## 🚀 **Quick Reference**

### Common Commands
```bash
# Start development environment
make dev

# Run tests
make test

# Apply database migrations
make migrate-up

# Generate models
make generate

# Build application
make build
```

### Key URLs
- **API Base**: `http://localhost:8888`
- **Health Check**: `http://localhost:8888/health`
- **Database**: `localhost:5432` (PostgreSQL)
- **Redis**: `localhost:6379`

### Important Files
- **Docker Compose**: `docker-compose.yml`
- **Database Config**: `dbconfig.yml`
- **Bob Config**: `bobgen.yaml`
- **Makefile**: `Makefile`

## 📞 **Getting Help**

### Documentation Issues
- Check this index for the right document
- Search for similar content in other files
- Create an issue for missing documentation

### Technical Issues
- Check the troubleshooting sections in relevant docs
- Review the changelog for recent changes
- Create a GitHub issue with detailed information

### Security Issues
- Report privately to security@hashpost.com
- Do not create public issues for security concerns

### New User Roles
- [User Roles & Permissions](user-roles.md)

---

This index should help you quickly find the documentation you need. If you can't find what you're looking for, please create an issue to help improve the documentation organization. 