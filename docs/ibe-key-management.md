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
    └── test_user_correlation.key
```

## Quick Start

### Development Setup

For development environments, use the Makefile target:

```bash
# Build the application first
make build

# Generate IBE keys
make setup-ibe-keys
```

This will:
1. Create the `./keys/` directory
2. Generate a new master key
3. Create all domain-specific keys
4. Generate role-based keys with time windows
5. Create test keys for development
6. Save configuration metadata

### Manual Setup

For more control, use the command directly:

```bash
# Build the application
go build -o bin/hashpost ./cmd/server

# Generate keys with custom settings
./bin/hashpost generate-ibe-keys \
  --output-dir ./keys \
  --key-version 1 \
  --salt "dev_salt_v1" \
  --generate-new \
  --non-interactive
```

### Container Setup

The application automatically generates IBE keys on container startup if they don't exist. The entrypoint script will:

1. Check if keys exist in `/app/keys/`
2. If not, run `./main generate-ibe-keys --output-dir /app/keys --generate-new --non-interactive`
3. Continue with application startup

## Environment Configuration

Set these environment variables to configure IBE key usage:

```bash
# Key file paths
export IBE_MASTER_KEY_PATH="./keys/master.key"
export IBE_DOMAIN_KEYS_DIR="./keys/domains"

# Configuration
export IBE_KEY_VERSION="1"
export IBE_SALT="fingerprint_salt_v1"

# Optional: Enable key rotation
export IBE_KEY_ROTATION_ENABLED="true"
export IBE_KEY_ROTATION_INTERVAL="30d"
export IBE_KEY_ROTATION_GRACE_PERIOD="7d"
```

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
./hashpost-server generate-ibe-keys --output-dir ./test-migration --generate-new
```

## Security Considerations

### Key Storage

- Store keys in secure, encrypted storage in production
- Use proper file permissions (600) for key files
- Implement key rotation policies
- Monitor key usage and access

### Domain Separation

- Each cryptographic domain has its own master key
- Keys from different domains cannot be used interchangeably
- This prevents privilege escalation attacks

### Time-Bounded Keys

- Keys automatically expire based on time windows
- Provides forward secrecy for correlation operations
- Reduces impact of key compromise

### Key Rotation

- Implement regular key rotation schedules
- Use grace periods to allow for migration
- Maintain backward compatibility during transitions
- Test rotation procedures in staging environments

## Advanced Configuration

### Custom Domains

```bash
./hashpost-server generate-ibe-keys \
  --domains "custom_domain_v1,another_domain_v1" \
  --generate-new
```

### Custom Time Windows

```bash
./hashpost-server generate-ibe-keys \
  --time-windows "15m,1h,6h,1d,1w" \
  --generate-new
```

### Custom Roles and Scopes

```bash
./hashpost-server generate-ibe-keys \
  --roles "user,moderator,admin" \
  --scopes "auth,correlation,audit" \
  --generate-new
```

## Monitoring and Logging

### Key Usage Monitoring

Monitor key usage through application logs:

```bash
# Enable debug logging for IBE operations
export LOG_LEVEL=debug

# Monitor key generation and usage
tail -f /var/log/hashpost/application.log | grep -i ibe
```

### Key Health Checks

Implement health checks for key availability:

```bash
# Check if keys exist and are readable
ls -la /opt/hashpost/keys/
./hashpost-server generate-ibe-keys --output-dir ./health-check --generate-new
```

## Best Practices

1. **Key Generation**: Always use the `generate-ibe-keys` command for consistent key generation
2. **Environment Separation**: Use different keys for development, staging, and production
3. **Backup Strategy**: Implement regular backups of key files and configurations
4. **Access Control**: Limit access to key files to only necessary personnel
5. **Monitoring**: Monitor key usage and implement alerts for unusual patterns
6. **Documentation**: Document key generation procedures and emergency procedures
7. **Testing**: Regularly test key generation and rotation procedures
8. **Compliance**: Ensure key management meets regulatory requirements

## Emergency Procedures

### Key Compromise

If keys are compromised:

1. **Immediate Response**:
   - Generate new keys with incremented version
   - Update environment configuration
   - Restart affected services

2. **Investigation**:
   - Audit key access logs
   - Identify compromise vector
   - Implement additional security measures

3. **Recovery**:
   - Migrate existing data to new keys
   - Update all dependent systems
   - Verify system functionality

### Key Loss

If keys are lost:

1. **Assessment**:
   - Determine scope of data affected
   - Identify backup availability
   - Assess recovery options

2. **Recovery**:
   - Restore from secure backups
   - Regenerate keys if necessary
   - Verify data integrity

3. **Prevention**:
   - Implement additional backup procedures
   - Review key management processes
   - Update disaster recovery plans 