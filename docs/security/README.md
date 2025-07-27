# Security Documentation

This directory contains security-related documentation for the HashPost system.

## Key Management

- [Key Management Architecture](key-management.md) - Overview of cryptographic key management
- [IBE System Documentation](ibe.md) - Identity-Based Encryption system details
- [Key Migration Operations Guide](key-migration-operations.md) - **Complete operational procedures for key migrations**
- [Key Migration Quick Reference](key-migration-quick-reference.md) - **Quick reference card for engineers**

## Security Analysis

- [Domain Separation Security Analysis](domain-separation-security-analysis.md) - Analysis of privilege separation
- [Enhancement Summary](enhancement-summary.md) - Security enhancements overview

## Key Migration

Key migrations are critical operations that require careful planning and execution. The system supports **multi-version key management** during migration, allowing both old and new keys to be used simultaneously.

### Key Migration Features

1. **Multi-Version Support**: System can decrypt with old keys and encrypt with new keys
2. **Resumable Migrations**: Migrations can be paused and resumed without data loss
3. **Progress Tracking**: Real-time monitoring of migration progress
4. **Rollback Capability**: Emergency rollback procedures for failed migrations
5. **Batch Processing**: Efficient processing of large datasets with rate limiting

### Migration Process Overview

1. **Preparation**: Generate new keys, backup current system
2. **Deployment**: Deploy new keys while keeping old keys available
3. **Migration Mode**: Enable system to use both key versions
4. **Data Migration**: Re-encrypt existing data with new keys
5. **Verification**: Ensure all data has been successfully migrated
6. **Cleanup**: Remove old keys and disable migration mode

### Operational Documentation

- **[Key Migration Operations Guide](key-migration-operations.md)** - Complete step-by-step procedures
- **[Key Migration Quick Reference](key-migration-quick-reference.md)** - Essential commands and checkpoints

### Testing

The multi-version key system is thoroughly tested:

```bash
# Run IBE system tests
go test ./internal/ibe -v -run TestMultiVersionKeyMigration

# Run migration service tests
go test ./internal/database/services -v -run TestResumableMigrationService
```

## Security Considerations

### Key Storage
- Keys are stored in secure, encrypted locations
- Production environments should use hardware security modules (HSMs)
- Key access is restricted to authorized personnel only

### Migration Safety
- All migrations are resumable and can be rolled back
- System maintains data integrity during migration
- Comprehensive monitoring and alerting during migration
- Backup procedures before any migration

### Access Control
- Migration operations require administrative privileges
- All migration activities are logged for audit purposes
- Approval workflows for production migrations

## Emergency Procedures

In case of migration issues:

1. **Immediate Rollback**: Use emergency rollback procedures
2. **Contact Security Team**: Notify security team of any issues
3. **Investigate Root Cause**: Analyze logs and system state
4. **Document Incident**: Record all actions taken and lessons learned

## References

- [Database Schema Documentation](../../database/schema.md)
- [API Documentation](../../api/README.md)
- [Development Setup](../../development/setup.md) 