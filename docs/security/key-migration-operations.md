# Key Migration Operations Guide

This guide provides step-by-step procedures for performing key migrations in the HashPost system. Key migrations are critical operations that require careful planning, execution, and monitoring.

## Overview

Key migrations involve:
1. **Preparing new keys** - Generating and deploying new cryptographic keys
2. **Enabling migration mode** - Allowing the system to use both old and new keys simultaneously
3. **Migrating data** - Re-encrypting existing data with new keys
4. **Verifying completion** - Ensuring all data has been successfully migrated
5. **Cleaning up** - Removing old keys and disabling migration mode

## Prerequisites

Before starting a key migration, ensure you have:

- **Access to production database** with appropriate permissions
- **Backup of current keys** stored securely
- **Maintenance window** scheduled (migrations can take hours)
- **Monitoring tools** configured to track migration progress
- **Rollback plan** prepared in case of issues

## Pre-Migration Checklist

### 1. System Health Check
```bash
# Check system status
curl -f http://localhost:8080/health

# Verify database connectivity
make migrate-status

# Check current key version
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/keys/version
```

### 2. Backup Current Keys
```bash
# Backup current domain keys
cp -r keys/domains keys/domains.backup.$(date +%Y%m%d_%H%M%S)

# Backup IBE configuration
cp keys/ibe_config.json keys/ibe_config.backup.$(date +%Y%m%d_%H%M%S)

# Verify backup integrity
sha256sum keys/domains.backup.*/user_self_correlation_v1.key
```

### 3. Database Backup
```bash
# Create database backup
pg_dump -h localhost -U hashpost -d hashpost > backup_$(date +%Y%m%d_%H%M%S).sql

# Verify backup size and integrity
ls -lh backup_*.sql
```

### 4. Estimate Migration Scope
```sql
-- Check how many records need migration
SELECT 
    key_version,
    COUNT(*) as record_count,
    MIN(created_at) as oldest_record,
    MAX(created_at) as newest_record
FROM identity_mappings 
GROUP BY key_version;

-- Check distribution by domain
SELECT 
    key_scope,
    key_version,
    COUNT(*) as record_count
FROM identity_mappings 
GROUP BY key_scope, key_version
ORDER BY key_scope, key_version;
```

## Migration Procedure

### Step 1: Generate New Keys

```bash
# Generate new domain keys (version 2)
./cmd/server/server ibe generate-keys --version 2 --output-dir keys/domains.v2

# Verify new keys were generated
ls -la keys/domains.v2/

# Generate new IBE configuration
./cmd/server/server ibe generate-config --version 2 --output keys/ibe_config.v2.json
```

### Step 2: Deploy New Keys

```bash
# Stop the application
docker-compose stop server

# Backup current keys
mv keys/domains keys/domains.v1
mv keys/ibe_config.json keys/ibe_config.v1.json

# Deploy new keys
cp -r keys/domains.v2 keys/domains
cp keys/ibe_config.v2.json keys/ibe_config.json

# Verify key permissions
chmod 600 keys/domains/*/*.key
chmod 600 keys/ibe_config.json

# Start application with new keys
docker-compose up -d server

# Verify application starts successfully
docker-compose logs server | tail -20
```

### Step 3: Enable Migration Mode

```bash
# Enable migration mode via API
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "old_key_version": 1,
    "new_key_version": 2,
    "domains": ["user_self_correlation_v1", "admin_correlation_v1"]
  }' \
  http://localhost:8080/api/v1/admin/keys/migration/enable

# Verify migration mode is active
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/keys/migration/status
```

### Step 4: Start Migration

```bash
# Start migration for user correlation domain
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "user_correlation",
    "old_key_version": 1,
    "new_key_version": 2,
    "batch_size": 1000,
    "rate_limit_ms": 100
  }' \
  http://localhost:8080/api/v1/admin/migrations/start

# Get migration ID from response
MIGRATION_ID="migration-uuid-from-response"

# Monitor migration progress
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/migrations/$MIGRATION_ID/progress
```

### Step 5: Monitor Migration Progress

```bash
# Set up monitoring loop
while true; do
  echo "$(date): Checking migration progress..."
  
  PROGRESS=$(curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
    http://localhost:8080/api/v1/admin/migrations/$MIGRATION_ID/progress)
  
  echo "$PROGRESS" | jq '.'
  
  STATUS=$(echo "$PROGRESS" | jq -r '.status')
  if [ "$STATUS" = "completed" ]; then
    echo "Migration completed successfully!"
    break
  elif [ "$STATUS" = "failed" ]; then
    echo "Migration failed! Check logs and consider rollback."
    break
  fi
  
  sleep 30
done
```

### Step 6: Verify Migration Completion

```sql
-- Verify all records have been migrated
SELECT 
    key_version,
    COUNT(*) as record_count
FROM identity_mappings 
GROUP BY key_version;

-- Should show only version 2 records
-- If version 1 records remain, migration is incomplete

-- Check for any failed migrations
SELECT 
    migration_id,
    status,
    processed_records,
    failed_records,
    total_records
FROM key_rotation_migrations 
WHERE status != 'completed';
```

### Step 7: Test System Functionality

```bash
# Test encryption/decryption with new keys
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "real_identity": "test@example.com",
    "pseudonym_id": "test-pseudonym-123"
  }' \
  http://localhost:8080/api/v1/admin/test/encrypt

# Test correlation functionality
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "admin",
    "scope": "correlation",
    "duration_hours": 24
  }' \
  http://localhost:8080/api/v1/admin/test/correlation-key
```

### Step 8: Disable Migration Mode

```bash
# Disable migration mode
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/keys/migration/disable

# Verify migration mode is disabled
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/keys/migration/status
```

### Step 9: Clean Up Old Keys

```bash
# Archive old keys (keep for potential rollback)
mv keys/domains.v1 keys/domains.archive.v1.$(date +%Y%m%d)
mv keys/ibe_config.v1.json keys/ibe_config.archive.v1.$(date +%Y%m%d)

# Remove old key backups after 30 days
find keys/ -name "*.backup.*" -mtime +30 -delete
```

## Rollback Procedures

### Emergency Rollback (If Migration Fails)

```bash
# 1. Stop the application
docker-compose stop server

# 2. Restore old keys
rm -rf keys/domains
rm keys/ibe_config.json
cp -r keys/domains.archive.v1.* keys/domains
cp keys/ibe_config.archive.v1.* keys/ibe_config.json

# 3. Restart application
docker-compose up -d server

# 4. Verify system is working with old keys
curl -f http://localhost:8080/health
```

### Partial Rollback (If Some Data Corrupted)

```bash
# 1. Identify corrupted records
SELECT 
    mapping_id,
    key_version,
    created_at
FROM identity_mappings 
WHERE key_version = 2 
  AND encrypted_real_identity IS NULL;

# 2. Restore from backup for specific records
# (This requires manual intervention based on the specific corruption)

# 3. Re-run migration for specific records
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "user_correlation",
    "old_key_version": 1,
    "new_key_version": 2,
    "record_ids": ["specific-mapping-id-1", "specific-mapping-id-2"]
  }' \
  http://localhost:8080/api/v1/admin/migrations/retry
```

## Monitoring and Alerting

### Key Metrics to Monitor

1. **Migration Progress**
   - Records processed per minute
   - Failed records count
   - Overall completion percentage

2. **System Performance**
   - Database connection pool usage
   - CPU and memory usage
   - API response times

3. **Error Rates**
   - Decryption failures
   - Encryption failures
   - Database errors

### Alerting Rules

```yaml
# Example Prometheus alerting rules
groups:
  - name: key_migration
    rules:
      - alert: MigrationStalled
        expr: migration_records_processed_per_minute == 0
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "Key migration appears to be stalled"
          
      - alert: MigrationFailed
        expr: migration_status == "failed"
        labels:
          severity: critical
        annotations:
          summary: "Key migration has failed"
          
      - alert: HighFailureRate
        expr: rate(migration_failed_records_total[5m]) > 0.1
        labels:
          severity: warning
        annotations:
          summary: "High failure rate in key migration"
```

## Troubleshooting

### Common Issues

1. **Migration Stalls**
   ```bash
   # Check for stuck records
   curl -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:8080/api/v1/admin/migrations/$MIGRATION_ID/stuck-records
   
   # Reset stuck records
   curl -X POST \
     -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:8080/api/v1/admin/migrations/$MIGRATION_ID/reset-stuck-records
   ```

2. **High Failure Rate**
   ```bash
   # Check failure reasons
   curl -H "Authorization: Bearer $ADMIN_TOKEN" \
     http://localhost:8080/api/v1/admin/migrations/$MIGRATION_ID/failures
   
   # Check system logs
   docker-compose logs server | grep -i "migration\|decrypt\|encrypt"
   ```

3. **Memory Issues**
   ```bash
   # Check memory usage
   docker stats hashpost_server_1
   
   # Restart with increased memory
   docker-compose down
   docker-compose up -d server --memory=2g
   ```

### Log Analysis

```bash
# Monitor migration logs in real-time
docker-compose logs -f server | grep -E "(migration|key|encrypt|decrypt)"

# Search for errors
docker-compose logs server | grep -i "error\|failed\|exception"

# Check specific migration ID
docker-compose logs server | grep "$MIGRATION_ID"
```

## Post-Migration Verification

### 1. Data Integrity Check

```sql
-- Verify no data loss
SELECT 
    COUNT(*) as total_records,
    COUNT(DISTINCT mapping_id) as unique_mappings,
    COUNT(DISTINCT encrypted_real_identity) as unique_encrypted
FROM identity_mappings;

-- Should have same counts before and after migration
```

### 2. Functional Testing

```bash
# Test user authentication
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "testpass"}' \
  http://localhost:8080/api/v1/auth/login

# Test pseudonym creation
curl -X POST \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"display_name": "TestPseudonym", "bio": "Test bio"}' \
  http://localhost:8080/api/v1/pseudonyms

# Test content creation
curl -X POST \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test Post", "content": "Test content"}' \
  http://localhost:8080/api/v1/posts
```

### 3. Performance Testing

```bash
# Load test with new keys
ab -n 1000 -c 10 -H "Authorization: Bearer $USER_TOKEN" \
  http://localhost:8080/api/v1/posts

# Compare response times with previous baseline
```

## Security Considerations

### Key Storage
- Store keys in secure, encrypted locations
- Use hardware security modules (HSMs) in production
- Implement key rotation policies
- Monitor key access and usage

### Access Control
- Limit migration operations to authorized personnel
- Use audit logging for all migration operations
- Implement approval workflows for production migrations

### Data Protection
- Ensure no plaintext data is logged during migration
- Use secure communication channels for key distribution
- Implement data retention policies for old keys

## Documentation Updates

After successful migration:

1. Update system documentation with new key versions
2. Update operational procedures
3. Update disaster recovery plans
4. Update monitoring and alerting configurations
5. Archive old documentation for reference

## Emergency Contacts

- **Primary Engineer**: [Contact Info]
- **Backup Engineer**: [Contact Info]
- **Security Team**: [Contact Info]
- **Database Administrator**: [Contact Info]

## References

- [Key Management Architecture](../key-management.md)
- [IBE System Documentation](../ibe.md)
- [Database Schema Documentation](../../database/schema.md)
- [API Documentation](../../api/README.md) 