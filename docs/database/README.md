# Database Documentation

This directory contains documentation related to HashPost's database system, schema, and operations.

## Overview

HashPost uses PostgreSQL as its primary database with a sophisticated schema supporting role-based access control, identity-based encryption, and comprehensive audit logging.

## Documentation

### [Database Schema](schema.md)
Complete database schema documentation:
- All table definitions and relationships
- Indexes and performance considerations
- Role-based access patterns
- Data types and validation
- Foreign key constraints
- Audit and logging tables

### [Database Operations](operations.md)
Database administration and operations:
- Migration workflows
- Backup and recovery procedures
- Performance optimization
- Security operations
- Monitoring and alerting
- Troubleshooting guide

### [Database ERD](erd.puml)
Entity Relationship Diagram (PlantUML source):
- Visual representation of database schema
- Table relationships and dependencies
- Primary and foreign keys
- Color-coded sections by functionality

### [ERD](ERD.md)
Instructions for ERD generation and viewing:
- How to generate the ERD diagram
- Different viewing options
- Troubleshooting
- Diagram structure explanation

## Quick Start

1. **Start with**: [Database Schema](schema.md) - Complete schema reference
2. **Then read**: [Database Operations](operations.md) - Operations guide
3. **Reference**: [Database ERD](erd.puml) - Visual schema

## Database Categories

### Core Tables
- Users and authentication
- Posts and comments
- Subforums and subscriptions
- Voting and moderation

### Security & Access Control
- Role keys and permissions
- Identity mappings
- Pseudonym management
- Audit logging

### Content Management
- Media attachments
- Polls and votes
- Reports and moderation
- System events

### Administrative
- System settings
- Performance metrics
- Compliance reports
- Key usage audit

## Related Documentation

- [RBAC Documentation](../rbac/) - Access control implementation
- [Security Documentation](../security/) - Cryptographic foundations
- [API Documentation](../api/) - Data access patterns 