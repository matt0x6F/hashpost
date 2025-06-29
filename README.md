# HashPost - Identity-Based Encryption Proof of Concept

This repository contains a proof of concept implementation of Identity-Based Encryption (IBE) for a pseudonymous social media platform. The system demonstrates how to balance user privacy with administrative accountability through cryptographic controls.

## Overview

The proof of concept implements the IBE strategy outlined in `docs/identity-based-encryption.md`, showing how to:

- Generate pseudonymous user profiles that cannot be correlated by regular users
- Implement role-based administrative access for correlation
- Maintain comprehensive audit trails for all correlation activities
- Support different administrative roles with varying scopes and capabilities

## Architecture

### Core Components

- **IBE System** (`internal/ibe/`): Cryptographic foundation for pseudonym generation and identity mapping
- **Data Models** (`internal/models/`): Structures for users, posts, administrative roles, and audit logs
- **Storage Layer** (`internal/storage/`): In-memory database for demonstration purposes
- **Demonstration** (`cmd/ibe-demo/`): Interactive proof of concept showing the system in action

### Key Features

1. **Pseudonym Generation**: Users register with real identities but receive pseudonymous profiles
2. **Role-Based Access**: Different administrative roles have different correlation capabilities
3. **Audit Trail**: All correlation activities are logged with timestamps and justifications
4. **Scope Limitations**: Administrative access is limited by role, time, and community boundaries

## Running the Proof of Concept

### Prerequisites

- Go 1.24.3 or later
- Git

### Installation

1. Clone the repository:
```bash
git clone https://github.com/matt0x6f/hashpost.git
cd hashpost
```

2. Run the demonstration:
```bash
go run cmd/ibe-demo/main.go
```

### What the Demo Shows

The demonstration walks through:

1. **Administrative Role Setup**: Creating different admin roles with varying permissions
2. **User Registration**: Users register with real identities but receive pseudonymous profiles
3. **Content Creation**: Users create posts using their pseudonyms
4. **Administrative Correlation**: Different admin roles demonstrate correlation capabilities
5. **Audit Trail**: Complete logging of all correlation activities

## Example Output

```
=== Identity-Based Encryption (IBE) Proof of Concept ===
Demonstrating pseudonymous social platform with administrative correlation

🚀 Starting IBE Demonstration...

📋 Setting up Administrative Roles...
  ✅ Created role: Site Administrator (full_correlation)
  ✅ Created role: Trust & Safety (harassment_investigation)
  ✅ Created role: Subforum Moderator (golang:local_correlation)
  ✅ Created role: Anti-Spam Team (network_analysis)
  ✅ Created admin: admin_sarah
  ✅ Created admin: trust_alex
  ✅ Created admin: mod_john
  ✅ Created admin: spam_bot

👥 Registering Users with Pseudonyms...
  ✅ User alice@example.com registered with pseudonym: a1b2c3d4...
  ✅ User bob@example.com registered with pseudonym: e5f6g7h8...
  ✅ User charlie@example.com registered with pseudonym: i9j0k1l2...
  ✅ User diana@example.com registered with pseudonym: m3n4o5p6...

🏛️  Creating Subforums...
  ✅ Created subforum: r/golang
  ✅ Created subforum: r/privacy

📝 Users Creating Posts...
  ✅ Post 1: How to implement IBE in Go? (by user_1)
  ✅ Post 2: Best practices for pseudonymous systems (by user_2)
  ✅ Post 3: Privacy concerns with social media (by user_3)
  ✅ Post 4: Go crypto libraries recommendation (by user_4)
  ✅ Post 5: Identity-based encryption explained (by user_1)

🔍 Demonstrating Administrative Correlation...

  🔍 Scenario 1: Site Administrator
     Scope: full_correlation
     Justification: Platform-wide investigation of coordinated activity
     ✅ Correlated: user@example.com -> user_1_1
     ✅ Correlated: user@example.com -> user_1_2
     ✅ Correlated: user@example.com -> user_1_3
     ✅ Correlated: user@example.com -> user_1_4

  🔍 Scenario 2: Trust & Safety
     Scope: harassment_investigation
     Justification: Investigation of reported harassment across subforums
     ✅ Correlated: user@example.com -> user_2_1
     ✅ Correlated: user@example.com -> user_2_2

  🔍 Scenario 3: Subforum Moderator
     Scope: golang:local_correlation
     Justification: Local rule enforcement in r/golang
     ✅ Correlated: user@example.com -> user_3_1

  🔍 Scenario 4: Anti-Spam Team
     Scope: network_analysis
     Justification: Automated detection of spam ring activity
     ✅ Correlated: user@example.com -> user_4_1
     ✅ Correlated: user@example.com -> user_4_2

📊 Correlation Audit Trail...
  Total correlation requests: 9

  Audit Entry 1:
    Admin: admin_1 (Site Administrator)
    Target: user_1_1
    Justification: Platform-wide investigation of coordinated activity
    Result: user@example.com -> user_1_1
    Timestamp: 2024-01-15 10:30:00
    Legal Basis: Platform Terms of Service

✅ IBE Demonstration Complete!
```

## Security Features

### Privacy Protections

- **Pseudonym Generation**: Real identities are never stored in plain text
- **Encrypted Mappings**: Identity correlations are encrypted with role-based keys
- **Scope Limitations**: Administrative access is restricted by role and time windows
- **Audit Requirements**: All correlation activities require justification and are logged

### Administrative Controls

- **Role Hierarchy**: Different admin roles have different capabilities
- **Key Rotation**: Administrative keys expire and must be rotated
- **Multi-Signature**: Critical operations require multiple approvals
- **Legal Compliance**: All activities are documented for legal review

## Implementation Notes

This proof of concept uses simplified cryptographic operations for demonstration purposes. A production implementation would require:

- **Hardware Security Modules (HSMs)** for key storage
- **Production-grade IBE libraries** (e.g., Boneh-Franklin IBE)
- **Database encryption at rest** for all sensitive data
- **Network isolation** between administrative and public systems
- **Multi-factor authentication** for administrative access
- **Regular security audits** and penetration testing

## Contributing

This is a proof of concept for educational and demonstration purposes. For production use, please ensure proper security review and implementation of all recommended security measures.

## License

This project is provided as-is for educational purposes. Please review the security implications before using in any production environment. 