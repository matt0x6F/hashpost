# IBE Key Management

## Overview

HashPost's Identity-Based Encryption (IBE) system uses an enhanced architecture with cryptographic domain separation and time-bounded key derivation. This document describes how to generate and manage IBE keys using the command-line interface.

## Enhanced Architecture

### Cryptographic Domain Separation

The enhanced IBE system separates cryptographic operations into distinct domains to prevent privilege escalation:

- **User Pseudonyms Domain** (`user_pseudonyms_v1`): For generating user pseudonyms
- **User Self-Correlation Domain** (`user_self_correlation_v1`): For user self-correlation operations
- **Moderator Correlation Domain** (`moderator_correlation_v1`): For moderator fingerprint correlation
- **Admin Correlation Domain** (`admin_correlation_v1`): For platform-wide identity correlation
- **Legal Correlation Domain** (`legal_correlation_v1`): For legal compliance operations

### Time-Bounded Key Derivation

All correlation keys include time components for forward secrecy:

- **1 Hour Windows**: For short-term operations
- **24 Hour Windows**: For daily operations
- **7 Day Windows**: For weekly operations
- **30 Day Windows**: For monthly operations

## Command Line Interface

### Generate IBE Keys

The `generate-ibe-keys` command creates all necessary keys for the enhanced IBE architecture:

```bash
./hashpost-server generate-ibe-keys [flags]
```

#### Basic Usage

```bash
# Generate keys with default settings
./hashpost-server generate-ibe-keys --output-dir ./keys --generate-new

# Use existing master key
./hashpost-server generate-ibe-keys --output-dir ./keys --master-key-path ./existing-master.key

# Custom configuration
./hashpost-server generate-ibe-keys \
  --output-dir ./production-keys \
  --key-version 2 \
  --salt "production_salt_v2" \
  --generate-new \
  --time-windows "1h,24h,7d,30d" \
  --roles "user,moderator,platform_admin" \
  --scopes "authentication,correlation"
```

#### Command Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--output-dir` | string | `./keys` | Output directory for generated keys |
| `--key-version` | int | `1` | Key version to generate |
| `--salt` | string | `fingerprint_salt_v1` | Salt for fingerprint generation |
| `--master-key-path` | string | `""` | Path to existing master key file |
| `--generate-new` | bool | `false` | Generate new master key |
| `--domains` | string | `""` | Comma-separated list of domains |
| `--time-windows` | string | `""` | Comma-separated time windows (1h,24h,7d,30d) |
| `--roles` | string | `""` | Comma-separated list of roles |
| `--scopes` | string | `""` | Comma-separated list of scopes |
| `--non-interactive` | bool | `false` | Non-interactive mode |

#### Default Values

When not specified, the command uses these defaults:

- **Domains**: All five cryptographic domains
- **Time Windows**: 1h, 24h, 7d, 30d
- **Roles**: user, moderator, subforum_owner, platform_admin, trust_safety, legal_team
- **Scopes**: authentication, self_correlation, correlation

## Generated File Structure

The command creates a hierarchical directory structure:

```
output-dir/
├── master.key                    # Master secret key
├── ibe_config.json              # Configuration metadata
├── domains/                      # Domain-specific keys
│   ├── user_pseudonyms_v1.key
│   ├── user_self_correlation_v1.key
│   ├── moderator_correlation_v1.key
│   ├── admin_correlation_v1.key
│   └── legal_correlation_v1.key
├── roles/                        # Role-specific keys by time window
│   ├── user/
│   │   ├── authentication/
│   │   │   ├── 1h.key
│   │   │   ├── 1d.key
│   │   │   ├── 1w.key
│   │   │   └── 1m.key
│   │   ├── self_correlation/
│   │   └── correlation/
│   ├── moderator/
│   ├── subforum_owner/
│   ├── platform_admin/
│   ├── trust_safety/
│   └── legal_team/
└── test/                         # Test keys for development
    ├── pseudonym_1_v1.txt
    ├── pseudonym_2_v1.txt
    ├── test_user_authentication.key
    ├── test_user_correlation.key
    └── ...
```

## Configuration File

The `ibe_config.json` file contains metadata about the generated keys:

```json
{
  "key_version": 1,
  "salt": "fingerprint_salt_v1",
  "domains": {
    "user_pseudonyms": "user_pseudonyms_v1",
    "user_correlation": "user_self_correlation_v1",
    "mod_correlation": "moderator_correlation_v1",
    "admin_correlation": "admin_correlation_v1",
    "legal_correlation": "legal_correlation_v1"
  },
  "generated_at": "2025-07-02T03:10:54Z"
}
```

## Security Considerations

### Master Key Management

- **Generate securely**: Use cryptographically secure random number generation
- **Store securely**: Keep master keys in secure storage (HashiCorp Vault, AWS KMS, etc.)
- **Backup securely**: Encrypt backups of master keys
- **Rotate regularly**: Generate new master keys periodically

### Domain Separation Benefits

- **Privilege isolation**: Compromise of one domain doesn't affect others
- **Granular recovery**: Individual domain key rotation possible
- **Audit separation**: Clear cryptographic boundaries for compliance

### Time-Bounded Keys

- **Forward secrecy**: Historical compromise doesn't affect current operations
- **Limited exposure**: Key compromise limited to time window
- **Automatic rotation**: Keys automatically rotate based on time epochs

## Consequences of Master Key Rotation

Rotating a domain master key for IBE has significant and irreversible consequences:

- **Loss of Access to Old Data:** All data encrypted or pseudonymized with the old master key for that domain becomes permanently inaccessible. This includes all derived keys for all time windows, roles, and scopes under that domain.
- **User Pseudonym Loss:** If you rotate the user pseudonyms domain master key, all existing user pseudonyms become invalid. Users would lose their pseudonyms and would need to generate new ones, breaking all existing content attribution and user identity continuity.
- **Forward/Backward Secrecy:** Forward secrecy is provided by time-bounded key derivation, not by rotating the master key. Rotating the master key does not retroactively protect old data; it simply makes all old data unrecoverable unless you have a migration/unsealing process.
- **Migration/Unsealing Required:** If you need to retain access to old data, you must decrypt it with the old master key and re-encrypt it with the new one before rotation. This process must be carefully planned and executed.
- **High-Impact, Rare Operation:** Master key rotation is not part of normal operational hygiene. It should only be performed in response to a suspected compromise or as part of a rare, planned rekeying event.
- **Operational Recommendations:**
  - Only rotate master keys if absolutely necessary.
  - Always back up current master keys securely before rotation.
  - Plan for downtime or read-only mode during rotation if data migration is required.
  - Document and test your migration/unsealing process in advance.
  - Communicate the impact to all stakeholders, as data loss is irreversible without proper migration.

**Summary:**
Master key rotation is a destructive operation for all data encrypted under the old key. For regular forward secrecy, rely on the system's time-bounded key derivation, not on master key rotation.

## Future Work: Online Migration Strategy

In the event of a master key compromise, we need to develop a strategy for migrating keys while keeping services online. This would involve:

- **Dual-key support**: Ability to use both old and new master keys simultaneously
- **Gradual migration**: Migrate data in batches without service interruption
- **Rollback capability**: Ability to revert to old keys if migration fails
- **Zero-downtime deployment**: Service updates that don't require stopping the application
- **Data consistency**: Ensuring all data is properly migrated before switching over

This is a complex operational challenge that requires careful design and testing before implementation.

## Production Deployment

### Step 1: Generate Production Keys

```bash
# Create production key directory
mkdir -p /opt/hashpost/keys

# Generate production keys
./hashpost-server generate-ibe-keys \
  --output-dir /opt/hashpost/keys \
  --key-version 1 \
  --salt "production_salt_v1" \
  --generate-new \
  --non-interactive
```

### Step 2: Secure Key Storage

```bash
# Set proper permissions
chmod 600 /opt/hashpost/keys/master.key
chmod 600 /opt/hashpost/keys/domains/*.key
chmod 600 /opt/hashpost/keys/roles/**/*.key

# Set ownership
chown hashpost:hashpost /opt/hashpost/keys -R
```

### Step 3: Environment Configuration

```bash
# Set environment variables
export IBE_MASTER_KEY_PATH="/opt/hashpost/keys/master.key"
export IBE_SALT="production_salt_v1"
export IBE_KEY_VERSION="1"
```

### Step 4: Key Rotation

```bash
# Generate new keys with incremented version
./hashpost-server generate-ibe-keys \
  --output-dir /opt/hashpost/keys-v2 \
  --key-version 2 \
  --salt "production_salt_v2" \
  --generate-new \
  --non-interactive

# Update environment and restart services
export IBE_MASTER_KEY_PATH="/opt/hashpost/keys-v2/master.key"
export IBE_SALT="production_salt_v2"
export IBE_KEY_VERSION="2"
```

## Development and Testing

### Test Key Generation

The command automatically generates test keys for development:

```bash
# Generate test keys
./hashpost-server generate-ibe-keys --output-dir ./test-keys --generate-new

# Use test keys in development
export IBE_MASTER_KEY_PATH="./test-keys/master.key"
```

### Integration Testing

Test keys are used in integration tests to verify the enhanced architecture:

```go
// Test domain separation
func TestIBESystem_DomainSeparation(t *testing.T) {
    ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
        MasterSecret: testMasterSecret,
        KeyVersion:   1,
        Salt:         "test_salt",
    })
    
    // Verify different domains generate different keys
    userKey := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
    modKey := ibeSystem.GenerateTimeBoundedKey("moderator", "correlation", time.Hour)
    
    if bytes.Equal(userKey, modKey) {
        t.Fatal("Domain separation failed: user and moderator keys are identical")
    }
}
```

## Troubleshooting

### Common Issues

#### Permission Denied
```bash
# Fix key file permissions
chmod 600 /path/to/keys/*.key
chown hashpost:hashpost /path/to/keys -R
```

#### Invalid Master Key
```bash
# Regenerate master key
./hashpost-server generate-ibe-keys --generate-new --output-dir ./new-keys
```

#### Missing Dependencies
```bash
# Ensure all required packages are installed
go mod tidy
go build ./cmd/server
```

### Debug Mode

Enable debug logging to see detailed key generation process:

```bash
./hashpost-server --debug generate-ibe-keys --output-dir ./debug-keys --generate-new
```

## Migration from Legacy System

### Step 1: Backup Existing Keys

```bash
# Backup existing master key
cp /path/to/existing/master.key /backup/master.key.backup
```

### Step 2: Generate Enhanced Keys

```bash
# Generate new enhanced keys
./hashpost-server generate-ibe-keys \
  --output-dir /path/to/enhanced-keys \
  --key-version 2 \
  --salt "enhanced_salt_v2" \
  --generate-new
```

### Step 3: Update Configuration

```bash
# Update environment variables
export IBE_MASTER_KEY_PATH="/path/to/enhanced-keys/master.key"
export IBE_SALT="enhanced_salt_v2"
export IBE_KEY_VERSION="2"
```

### Step 4: Verify Migration

```bash
# Run integration tests
make test-integration-local

# Verify key functionality
./hashpost-server generate-ibe-keys --master-key-path /path/to/enhanced-keys/master.key
```

## Best Practices

### Key Management

1. **Use separate keys for different environments** (dev, staging, production)
2. **Rotate keys regularly** (quarterly or annually)
3. **Monitor key usage** for unusual patterns
4. **Backup keys securely** with encryption
5. **Document key versions** and migration procedures

### Security

1. **Never commit keys to version control**
2. **Use secure random generation** for all keys
3. **Implement proper access controls** for key files
4. **Monitor for key compromise** indicators
5. **Have incident response procedures** for key compromise

### Operations

1. **Test key generation** in staging environments
2. **Validate key functionality** before production deployment
3. **Monitor application logs** for key-related errors
4. **Have rollback procedures** for key changes
5. **Document all key management procedures**

## API Integration

The enhanced IBE system is automatically used by the application when the environment variables are set:

```go
// The application automatically uses enhanced IBE system
ibeSystem := ibe.NewIBESystemFromEnv()

// Generate pseudonyms with domain separation
pseudonym := ibeSystem.CreateEnhancedPseudonym(userID, context)

// Generate time-bounded correlation keys
correlationKey := ibeSystem.GenerateTimeBoundedKey(role, scope, timeWindow)
```

## Conclusion

The enhanced IBE key management system provides:

- **Cryptographic domain separation** for privilege isolation
- **Time-bounded key derivation** for forward secrecy
- **Comprehensive key generation** for all system components
- **Secure key management** procedures for production deployment
- **Backward compatibility** with existing systems

This architecture transforms HashPost into a platform with industry-leading privacy and security capabilities, suitable for security-conscious users and premium advertisers. 