# IBE Key Management

This document describes how to manage Identity-Based Encryption (IBE) keys in HashPost.

## Overview

HashPost uses IBE for privacy-preserving correlation between users and pseudonyms. The IBE system requires a master secret key that must be consistent across all application instances.

## Key Components

- **Master Secret**: 32-byte cryptographic key used for all IBE operations
- **Fingerprint**: SHA256 hash of real identity + salt, used for correlation
- **Role Keys**: Derived from master secret + role + scope + expiration
- **Identity Mappings**: Encrypted `fingerprint:pseudonymID` pairs

## Setup

### Development Environment

1. **Generate IBE Keys**:
   ```bash
   make setup-ibe-keys
   ```
   This will create `./keys/master.key` with proper permissions.

2. **Start Application**:
   ```bash
   make dev
   ```
   The container will mount the key from the host filesystem.

### Production Environment

1. **Generate Production Key**:
   ```bash
   mkdir -p /opt/hashpost/keys
   openssl rand -hex 32 | tr -d '\n' > /opt/hashpost/keys/master.key
   chmod 600 /opt/hashpost/keys/master.key
   ```
   The key file will contain exactly 64 hex characters (representing 32 bytes) with no newline.

2. **Update Docker Compose**:
   ```yaml
   volumes:
     - /opt/hashpost/keys:/app/keys:ro
   ```

3. **Set Environment Variables**:
   ```bash
   export IBE_MASTER_KEY_PATH=/app/keys/master.key
   export IBE_KEY_VERSION=1
   export IBE_SALT=production_fingerprint_salt_v1
   ```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `IBE_MASTER_KEY_PATH` | Path to master key file | `./keys/master.key` |
| `IBE_KEY_VERSION` | Current key version | `1` |
| `IBE_SALT` | Salt for fingerprint generation | `hashpost_fingerprint_salt_v1` |
| `IBE_KEY_ROTATION_ENABLED` | Enable automatic key rotation | `false` |
| `IBE_KEY_ROTATION_INTERVAL` | How often to rotate keys | `8760h` (1 year) |
| `IBE_KEY_ROTATION_GRACE_PERIOD` | Grace period for rotation | `720h` (30 days) |

### Key Rotation

Key rotation is currently **not implemented** but planned for future releases. When implemented, it will:

1. **Check Rotation Schedule**: Determine if rotation is due
2. **Generate New Key**: Create a new master secret
3. **Re-encrypt Mappings**: Update all existing identity mappings
4. **Update Version**: Increment the key version
5. **Grace Period**: Allow both old and new keys during transition

### Security Considerations

1. **Key Storage**:
   - Store keys in secure, encrypted storage
   - Use proper file permissions (600)
   - Never commit keys to version control
   - Backup keys securely

2. **Key Distribution**:
   - Use different keys for different environments
   - Rotate keys regularly in production
   - Monitor key usage and access

3. **Access Control**:
   - Limit access to master keys
   - Use key management systems (AWS KMS, HashiCorp Vault)
   - Audit key access and usage

## Troubleshooting

### Common Issues

1. **"No master key found"**:
   ```bash
   make setup-ibe-keys
   ```

2. **"Key not readable"**:
   ```bash
   chmod 600 ./keys/master.key
   ```

3. **"Identity mapping decryption failed"**:
   - Check if key has changed
   - Verify key version matches
   - Ensure consistent salt across environments

### Key Recovery

If you lose your master key:

1. **Development**: Generate new key and reset database
2. **Production**: Restore from secure backup
3. **Partial Loss**: Use key rotation to migrate to new key

## Best Practices

1. **Environment Separation**:
   - Use different keys for dev/staging/production
   - Never use production keys in development

2. **Backup Strategy**:
   - Backup keys securely (encrypted)
   - Test key restoration procedures
   - Document key management procedures

3. **Monitoring**:
   - Monitor key usage and access
   - Alert on key rotation events
   - Log key-related operations

4. **Documentation**:
   - Document key generation procedures
   - Maintain key inventory
   - Update procedures when keys change

## Future Enhancements

1. **Automatic Key Rotation**: Implement scheduled key rotation
2. **Key Management Integration**: Support for AWS KMS, HashiCorp Vault
3. **Multi-Key Support**: Support for multiple active keys
4. **Key Recovery**: Automated key recovery procedures
5. **Audit Logging**: Comprehensive key usage logging

## TODO - IBE System Implementation

### Completed ✅
- [x] Basic IBE system implementation with master key generation
- [x] Identity mapping encryption/decryption
- [x] Fingerprint generation with salt
- [x] Role key derivation for admin operations
- [x] Integration with user-pseudonym correlation
- [x] Docker Compose configuration with persistent key mounting
- [x] Key management scripts and Makefile commands
- [x] Integration tests with deterministic IBE system
- [x] Environment variable configuration

### Remaining Tasks 🔄

#### High Priority
1. **Key Rotation Implementation**
   - [ ] Implement key rotation logic in `internal/ibe/ibe.go`
   - [ ] Add rotation scheduling and grace period handling
   - [ ] Create migration scripts for re-encrypting identity mappings
   - [ ] Add rotation status monitoring and alerts

2. **Production Key Management**
   - [ ] Implement secure key backup/restore procedures
   - [ ] Add key integrity verification (checksums)
   - [ ] Create production deployment scripts
   - [ ] Add key usage monitoring and logging

#### Medium Priority
3. **Enhanced Security**
   - [ ] Add key versioning support in identity mappings
   - [ ] Implement key derivation from external sources (KMS, Vault)
   - [ ] Add key access audit logging
   - [ ] Implement key escrow for admin recovery

4. **Operational Improvements**
   - [ ] Add health checks for IBE system
   - [ ] Implement key performance metrics
   - [ ] Add key rotation dry-run mode
   - [ ] Create key management CLI tools

#### Low Priority
5. **Advanced Features**
   - [ ] Multi-key support for different environments
   - [ ] Key recovery automation
   - [ ] Integration with cloud KMS services
   - [ ] Key usage analytics and reporting

### Technical Debt
- [ ] Add comprehensive unit tests for IBE operations
- [ ] Improve error handling and user feedback
- [ ] Add IBE system configuration validation
- [ ] Document IBE cryptographic details and security model 