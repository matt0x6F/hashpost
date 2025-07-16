# Identity-Based Encryption Strategy for Pseudonymous Social Platform

## Overview

This document outlines the Identity-Based Encryption (IBE) strategy for implementing pseudonymous user profiles on a Reddit-like social media platform. The system balances user privacy with administrative accountability by allowing correlation of pseudonymous accounts only by authorized personnel with appropriate access levels.

## Core Objectives

- **User Privacy**: Regular users cannot correlate pseudonymous profiles to real identities or link multiple pseudonyms
- **Administrative Accountability**: Authorized staff can correlate accounts for moderation and legal compliance
- **Role-Based Access**: Different administrative roles have different correlation capabilities
- **Forward Secrecy**: Historical data remains protected even if current keys are compromised
- **Audit Trail**: All correlation activities are logged for compliance and oversight

## System Architecture

### Public Layer
- Users interact through pseudonymous profiles only
- Posts, comments, and votes are tied to pseudonym IDs
- No correlation information visible to regular users
- Standard Reddit-like features (subforums, voting, commenting)

### Administrative Layer
- Encrypted identity mappings stored separately from public data
- Role-based decryption keys enable selective correlation
- Audit logging for all correlation requests
- Multi-signature requirements for sensitive operations

## IBE Key Structure

### Master Authority
- **Root IBE Authority**: Generates all other keys, air-gapped storage
- **Key Escrow Service**: Manages active administrative keys
- **Audit Authority**: Independent monitoring of key usage

### Administrative Key Hierarchy

#### Site-Wide Administrator Keys
```
site_admin:full_correlation
├── Capability: Correlate any user across entire platform
├── Use Cases: Platform-wide investigations, legal compliance
├── Access Control: Requires 2-of-3 senior admin signatures
└── Audit Level: Maximum logging, external oversight
```

#### Legal Compliance Keys
```
legal_team:court_orders
├── Capability: Full correlation for specific users under court order
├── Use Cases: Subpoenas, law enforcement requests
├── Access Control: Legal team + senior admin approval
└── Audit Level: Court order documentation required
```

#### Trust & Safety Keys
```
trust_safety:harassment_investigation
├── Capability: Correlate reported users and their associated accounts
├── Use Cases: Harassment, doxxing, coordinated attacks
├── Access Control: Trust & Safety team lead approval
└── Audit Level: Incident report documentation
```

#### Anti-Spam Keys
```
spam_team:network_analysis
├── Capability: Correlate accounts within 48-hour windows
├── Use Cases: Spam rings, vote manipulation, bot networks
├── Access Control: Anti-spam team, automated systems
└── Audit Level: Standard logging
```

#### subforum Moderator Keys
```
subforum_mod:{subforum_name}:local_correlation
├── Capability: Correlate users active within their subforum only
├── Use Cases: Local rule enforcement, ban evasion
├── Access Control: subforum owner delegation
├── Scope Limitation: 30-day window, subforum-specific activity only
└── Audit Level: subforum action logs
```

#### subforum Owner Keys
```
subforum_owner:{subforum_name}:enhanced_correlation
├── Capability: Full correlation within subforum, extended time windows
├── Use Cases: Community management, moderator oversight
├── Access Control: Platform verification + community size thresholds
├── Scope Limitation: subforum-specific, 90-day windows
└── Audit Level: Community governance logs
```

## Technical Implementation

### User Registration & Pseudonym Generation
```
1. User registers with email/phone verification
2. System generates master user secret: user_master_secret
3. IBE generates pseudonym keypair: pseudonym_id = IBE.KeyGen(user_master_secret)
4. Public pseudonym profile created with no linkage information
5. Encrypted mapping stored: E(user_identity → pseudonym_id, admin_keys)
```

### Correlation Process
```
1. Administrative user requests correlation with justification
2. System validates role-based permissions
3. Appropriate IBE private key retrieved based on request scope
4. Correlation performed: user_identity = IBE.Decrypt(pseudonym_id, role_key)
5. Action logged with timestamp, requester, justification, results
6. Results provided through secure interface
```

### Key Rotation Schedule
- **Spam Detection Keys**: Rotated weekly
- **subforum Moderator Keys**: Rotated monthly
- **Trust & Safety Keys**: Rotated quarterly
- **Legal/Admin Keys**: Rotated annually
- **Master Authority**: Rotated every 2 years with ceremony

## Fingerprint-Based Correlation

To preserve user privacy while enabling administrative correlation, the system uses a cryptographic fingerprint derived from the user's real identity (e.g., email or phone number). This fingerprint is used in all encrypted identity mappings and for administrative lookups, instead of storing or revealing the real identity.

### Fingerprint Generation
- The fingerprint is generated as the first 16 bytes (128 bits) of a SHA-256 hash of the real identity concatenated with a system-wide salt.
- This value is hex-encoded (32 characters) and is deterministic for a given identity and salt.
- Example: `fingerprint = hex(SHA256(real_identity || salt)[:16])`

### Collision Risk Analysis
- A 128-bit fingerprint provides extremely strong collision resistance:
    - The probability of a collision (two users having the same fingerprint) is negligible for any realistic number of users.
    - The birthday bound for a 128-bit hash is about 2^64 entries before a 50% chance of a collision.
    - For 1 million users, the probability of a collision is less than 10^-20.
    - For 1 billion users, the probability is still negligible.
- If even stronger guarantees are desired, the full 32-byte (256-bit) SHA-256 output can be used, but 128 bits is sufficient for almost all applications.

### Privacy and Security Rationale
- The fingerprint allows correlation of all pseudonyms belonging to the same real user without revealing the actual identity to administrators or in the database.
- Only authorized administrative processes can use the fingerprint for lookups; regular users cannot reverse or correlate fingerprints.
- The system-wide salt ensures that fingerprints are unique to this deployment and cannot be precomputed from public data.

### Usage in the System
- All encrypted identity mappings store the fingerprint and pseudonym ID, not the real identity.
- Administrative correlation and audit logs reference the fingerprint, not the real identity.
- This approach provides strong privacy guarantees while enabling necessary moderation and compliance operations.

## Database Schema

### Public Database
```sql
-- Pseudonymous user profiles
users_pseudonymous (
    pseudonym_id VARCHAR(64) PRIMARY KEY,
    display_name VARCHAR(50),
    karma_score INTEGER,
    created_at TIMESTAMP,
    -- No correlation fields
)

-- Posts tied to pseudonyms only
posts (
    post_id BIGINT PRIMARY KEY,
    pseudonym_id VARCHAR(64) REFERENCES users_pseudonymous,
    subforum_id INTEGER,
    title TEXT,
    content TEXT,
    created_at TIMESTAMP
)
```

### Administrative Database (Encrypted)
```sql
-- Identity mappings (encrypted at rest)
identity_mappings (
    mapping_id UUID PRIMARY KEY,
    encrypted_real_identity BYTEA, -- E(email/phone, admin_keys)
    encrypted_pseudonym_mapping BYTEA, -- E(pseudonym_id, admin_keys)
    key_version INTEGER,
    created_at TIMESTAMP
)

-- Correlation audit log
correlation_audit (
    audit_id UUID PRIMARY KEY,
    admin_user_id UUID,
    admin_role VARCHAR(50),
    requested_pseudonym VARCHAR(64),
    justification TEXT,
    approved_by UUID,
    correlation_result BYTEA, -- Encrypted result
    timestamp TIMESTAMP,
    legal_basis VARCHAR(100)
)
```

## Security Considerations

### Threat Model
- **Internal Threats**: Rogue administrators, compromised admin accounts
- **External Threats**: Database breaches, key theft, state-level attacks
- **User Threats**: Sophisticated correlation attacks, timing analysis

### Mitigation Strategies
- **Key Splitting**: Critical operations require multiple key holders
- **Hardware Security**: Admin keys stored in HSMs where possible
- **Network Isolation**: Administrative systems air-gapped from public platform
- **Regular Audits**: Independent security reviews of correlation activities
- **Legal Framework**: Clear policies governing administrative access

### Privacy Protections
- **Data Minimization**: Only necessary correlation data stored
- **Purpose Limitation**: Keys restricted to specific use cases
- **Retention Limits**: Automatic deletion of old correlation records
- **User Notification**: Users informed when their accounts are correlated (where legally permissible)

## Operational Procedures

### Routine Moderation
1. subforum moderators use limited correlation for local enforcement
2. Automated spam detection uses time-limited correlation windows
3. Regular rotation of detection keys to limit exposure

### Escalated Incidents
1. Trust & Safety team handles cross-community violations
2. Multi-signature approval for serious investigations
3. Legal team involvement for law enforcement requests

### Emergency Procedures
1. Incident response team can request emergency correlation access
2. Time-limited emergency keys with enhanced audit requirements
3. Post-incident review and justification documentation

## Compliance & Legal Framework

### Jurisdiction Considerations
- Key escrow policies aligned with local data protection laws
- Legal basis documentation for all correlation activities
- Cross-border data transfer protections

### User Rights
- Right to know when correlation occurs (where legally permitted)
- Right to deletion with cryptographic erasure
- Right to audit correlation history for their accounts

### Transparency Reporting
- Quarterly reports on correlation requests and approvals
- Statistics on administrative access patterns
- Public documentation of governance policies

## Implementation Roadmap

### Phase 1: Core IBE Infrastructure
- IBE library integration and testing
- Basic administrative key generation
- Simple correlation interface for site admins

### Phase 2: Role-Based Access
- Implement hierarchical key structure
- subforum moderator key distribution
- Audit logging system

### Phase 3: Advanced Features
- Automated spam detection integration
- Legal compliance workflows
- User notification systems

### Phase 4: Optimization
- Performance tuning for large-scale operations
- Advanced audit analytics
- Mobile moderator interfaces

## Success Metrics

- **User Trust**: Survey metrics on perceived privacy protection
- **Moderation Effectiveness**: Reduction in ban evasion and coordinated attacks
- **Compliance**: 100% audit trail coverage for correlation activities
- **Performance**: Sub-second correlation times for administrative queries
- **Security**: Zero unauthorized correlation incidents

## Risk Assessment

### High Risk
- Master key compromise: Complete system trust failure
- Legal overreach: Inappropriate government surveillance
- Internal misuse: Administrators abusing correlation capabilities

### Medium Risk
- Key rotation failures: Temporary loss of correlation abilities
- Database breaches: Exposure of encrypted mapping data
- Performance degradation: Slow correlation affecting moderation

### Low Risk
- Library vulnerabilities: IBE implementation flaws
- Audit log manipulation: Administrative activity concealment
- User correlation attacks: Sophisticated timing/behavioral analysis

## Conclusion

This IBE-based approach provides strong privacy guarantees for users while maintaining the administrative oversight necessary for effective moderation at scale. The role-based key hierarchy ensures appropriate access controls while the audit framework provides accountability and compliance capabilities.

The system balances the competing needs of user privacy, platform safety, and legal compliance through cryptographic controls rather than policy alone, providing technical enforcement of administrative boundaries.