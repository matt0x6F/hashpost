# Key Migration Quick Reference

## Pre-Migration Checklist

- [ ] System health check passed
- [ ] Database backup completed
- [ ] Key backup completed
- [ ] Maintenance window scheduled
- [ ] Monitoring configured
- [ ] Rollback plan ready

## Essential Commands

### Health Checks
```bash
# System status
curl -f http://localhost:8080/health

# Database status
make migrate-status

# Current key version
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/keys/version
```

### Backups
```bash
# Database backup
pg_dump -h localhost -U hashpost -d hashpost > backup_$(date +%Y%m%d_%H%M%S).sql

# Key backup
cp -r keys/domains keys/domains.backup.$(date +%Y%m%d_%H%M%S)
cp keys/ibe_config.json keys/ibe_config.backup.$(date +%Y%m%d_%H%M%S)
```

### Migration Scope
```sql
-- Check migration scope
SELECT key_version, COUNT(*) as count 
FROM identity_mappings 
GROUP BY key_version;
```

## Migration Steps

### 1. Generate & Deploy New Keys
```bash
# Generate new keys
./cmd/server/server ibe generate-keys --version 2 --output-dir keys/domains.v2

# Deploy keys
docker-compose stop server
mv keys/domains keys/domains.v1
cp -r keys/domains.v2 keys/domains
docker-compose up -d server
```

### 2. Enable Migration Mode
```bash
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"old_key_version": 1, "new_key_version": 2}' \
  http://localhost:8080/api/v1/admin/keys/migration/enable
```

### 3. Start Migration
```bash
# Start migration
RESPONSE=$(curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"domain": "user_correlation", "old_key_version": 1, "new_key_version": 2}' \
  http://localhost:8080/api/v1/admin/migrations/start)

# Extract migration ID
MIGRATION_ID=$(echo $RESPONSE | jq -r '.migration_id')
```

### 4. Monitor Progress
```bash
# Check progress
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/migrations/$MIGRATION_ID/progress

# Monitor continuously
watch -n 30 'curl -s -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/migrations/$MIGRATION_ID/progress | jq'
```

### 5. Verify Completion
```sql
-- Verify all records migrated
SELECT key_version, COUNT(*) as count 
FROM identity_mappings 
GROUP BY key_version;
-- Should show only version 2
```

### 6. Disable Migration Mode
```bash
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/keys/migration/disable
```

## Emergency Rollback

```bash
# Stop application
docker-compose stop server

# Restore old keys
rm -rf keys/domains
cp -r keys/domains.v1 keys/domains
cp keys/ibe_config.v1.json keys/ibe_config.json

# Restart
docker-compose up -d server

# Verify
curl -f http://localhost:8080/health
```

## Troubleshooting

### Migration Stalled
```bash
# Check stuck records
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/migrations/$MIGRATION_ID/stuck-records

# Reset stuck records
curl -X POST \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/migrations/$MIGRATION_ID/reset-stuck-records
```

### High Failure Rate
```bash
# Check failures
curl -H "Authorization: Bearer $ADMIN_TOKEN" \
  http://localhost:8080/api/v1/admin/migrations/$MIGRATION_ID/failures

# Check logs
docker-compose logs server | grep -i "migration\|decrypt\|encrypt"
```

### System Issues
```bash
# Check system resources
docker stats

# Check logs
docker-compose logs -f server | grep -E "(error|failed|exception)"
```

## Verification Commands

### Data Integrity
```sql
-- Check record counts
SELECT COUNT(*) as total_records,
       COUNT(DISTINCT mapping_id) as unique_mappings
FROM identity_mappings;
```

### Functional Testing
```bash
# Test authentication
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "password": "testpass"}' \
  http://localhost:8080/api/v1/auth/login

# Test content creation
curl -X POST \
  -H "Authorization: Bearer $USER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "Test", "content": "Test"}' \
  http://localhost:8080/api/v1/posts
```

## Key Metrics to Monitor

- **Progress**: Records processed per minute
- **Failures**: Failed records count
- **Performance**: API response times
- **Resources**: CPU, memory, database connections

## Emergency Contacts

- **Primary**: [Contact]
- **Backup**: [Contact]
- **Security**: [Contact]
- **DBA**: [Contact]

## Important Notes

- **Never disable migration mode until verification is complete**
- **Keep old keys for at least 30 days after successful migration**
- **Monitor system performance during migration**
- **Have rollback plan ready at all times**
- **Test in staging environment first** 