# HashPost Changelog

## Overview

This document tracks major changes and feature updates to the HashPost platform. Each entry includes the problem statement, changes made, and benefits of the updates.

## Multiple Pseudonyms Support (Latest)

### Overview

This major update fundamentally improved the HashPost platform by properly supporting multiple pseudonyms per real user. The original schema incorrectly assumed a 1:1 relationship between users and pseudonyms, which has been corrected to support true pseudonymous user experiences.

### Problem Statement

The original database schema had a fundamental design flaw: it assumed each user could only have one pseudonym. This was reflected in several ways:

1. **`users` table**: Had a UNIQUE constraint on `pseudonym_id`, implying one pseudonym per user
2. **Mixed identity concepts**: The `users` table mixed real user identity (email, password) with pseudonym-specific data (display_name, karma_score)
3. **Incorrect foreign key references**: Many tables referenced `users(pseudonym_id)` directly
4. **Limited pseudonymity**: Users couldn't have multiple distinct personas

### Changes Made

#### 1. User Identity Separation

**Before:**
```sql
CREATE TABLE users (
    user_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    pseudonym_id VARCHAR(64) UNIQUE NOT NULL, -- One pseudonym per user
    display_name VARCHAR(50) NOT NULL,        -- Pseudonym data mixed with user data
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    karma_score INTEGER DEFAULT 0,            -- Pseudonym data mixed with user data
    -- ... other mixed fields
);
```

**After:**
```sql
-- Real user identity (one per actual person)
CREATE TABLE users (
    user_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    roles JSON DEFAULT '["user"]',
    capabilities JSON DEFAULT '["create_content", "vote", "message", "report"]',
    -- ... administrative fields only
);

-- Pseudonym-specific information (multiple per user)
CREATE TABLE pseudonyms (
    pseudonym_id VARCHAR(64) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    display_name VARCHAR(50) NOT NULL,
    karma_score INTEGER DEFAULT 0,
    bio TEXT,
    avatar_url VARCHAR(255),
    -- ... pseudonym-specific fields
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE
);
```

#### 2. Content Creation Updates

**Before:**
```sql
CREATE TABLE posts (
    post_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id BIGINT NOT NULL, -- Content created by real user
    -- ...
    FOREIGN KEY (user_id) REFERENCES users(user_id)
);
```

**After:**
```sql
CREATE TABLE posts (
    post_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    pseudonym_id VARCHAR(64) NOT NULL, -- Content created by pseudonym
    -- ...
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE
);
```

#### 3. Identity Management Updates

**Before:**
```sql
CREATE TABLE identity_mappings (
    mapping_id UUID PRIMARY KEY,
    fingerprint VARCHAR(32) NOT NULL,
    pseudonym_id VARCHAR(64) NOT NULL, -- Only pseudonym reference
    -- ...
);
```

**After:**
```sql
CREATE TABLE identity_mappings (
    mapping_id UUID PRIMARY KEY,
    fingerprint VARCHAR(32) NOT NULL,
    user_id BIGINT NOT NULL,           -- Real user identity
    pseudonym_id VARCHAR(64) NOT NULL, -- Specific pseudonym
    -- ...
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE
);
```

#### 4. Foreign Key Reference Updates

Updated all tables that previously referenced `users(pseudonym_id)` to now reference `pseudonyms(pseudonym_id)`:

- `subforum_subscriptions`
- `votes`
- `poll_votes`
- `user_blocks`
- `direct_messages`
- `reports`
- `user_bans`
- `moderation_actions`
- `correlation_audit`

### API Changes

#### Updated User Models

**Before:**
```go
type UserProfile struct {
    PseudonymID         string `json:"pseudonym_id"`
    DisplayName         string `json:"display_name"`
    KarmaScore          int    `json:"karma_score"`
    // ... pseudonym-specific fields mixed with user data
}

type UserProfileInput struct {
    DisplayName string `json:"display_name"`
    Bio         string `json:"bio"`
    // ... pseudonym-specific fields
}
```

**After:**
```go
type PseudonymProfile struct {
    PseudonymID         string `json:"pseudonym_id"`
    DisplayName         string `json:"display_name"`
    KarmaScore          int    `json:"karma_score"`
    LastActiveAt        string `json:"last_active_at"`
    IsActive            bool   `json:"is_active"`
    // ... pseudonym-specific fields only
}

type UserProfile struct {
    UserID       int                `json:"user_id"`
    Email        string             `json:"email"`
    Roles        []string           `json:"roles"`
    Capabilities []string           `json:"capabilities"`
    Pseudonyms   []PseudonymProfile `json:"pseudonyms"`
}

type PseudonymProfileInput struct {
    DisplayName         string `json:"display_name"`
    Bio                 string `json:"bio"`
    // ... pseudonym-specific fields only
}

type CreatePseudonymInput struct {
    DisplayName         string `json:"display_name" required:"true"`
    Bio                 string `json:"bio"`
    // ... pseudonym creation fields
}
```

#### New API Endpoints

- `GET /pseudonyms/{pseudonym_id}/profile` - Get public pseudonym profile
- `PUT /pseudonyms/{pseudonym_id}/profile` - Update pseudonym profile
- `POST /pseudonyms` - Create new pseudonym
- `GET /users/profile` - Get current user profile with all pseudonyms

#### Updated API Endpoints

- Profile management moved from `/users/profile` to `/pseudonyms/{pseudonym_id}/profile`
- User profile now shows all pseudonyms at `/users/profile`

### Benefits of These Changes

#### 1. True Multiple Pseudonyms
- Users can now have multiple distinct personas
- Each pseudonym can have its own display name, karma score, bio, etc.
- Pseudonyms are properly isolated from each other

#### 2. Clean Identity Separation
- Real user identity (email, password, roles) is separate from pseudonym identity
- User preferences (timezone, language) remain at the user level
- Pseudonym-specific data (display_name, karma_score) is at the pseudonym level

#### 3. Administrative Accountability
- Real identities are preserved for moderation and compliance
- Administrative actions can reference both real users and pseudonyms
- Audit trails maintain proper identity correlation

#### 4. Flexible Correlation
- `identity_mappings` table supports correlation at both fingerprint and identity levels
- Different administrative roles can correlate at different levels
- Regular users cannot correlate pseudonyms across different personas

#### 5. Backward Compatibility
- Existing administrative workflows remain intact
- User-level operations (authentication, roles) are unchanged
- Migration path is clear and manageable

## Moderator Pseudonym Support

### Overview

This update addressed a critical privacy and usability issue: moderators were only tracked by their real `user_id` in moderation-related tables, but moderators would naturally want to moderate under their pseudonymous identity.

### Problem Statement

The original database schema had a significant oversight: moderators were only tracked by their real `user_id` in moderation-related tables, but moderators would naturally want to moderate under their pseudonymous identity. This created several issues:

1. **Privacy Violation**: Users would see moderator actions tied to real identities instead of pseudonyms
2. **Inconsistent User Experience**: Moderators couldn't maintain their pseudonymous persona while moderating
3. **Trust Issues**: Users might be less likely to trust moderation from "real" accounts vs. established pseudonymous community members

### Changes Made

#### 1. `subforum_moderators` Table

**Before:**
```sql
CREATE TABLE subforum_moderators (
    moderator_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    subforum_id INTEGER NOT NULL,
    user_id BIGINT NOT NULL, -- Only real identity
    role VARCHAR(20) NOT NULL DEFAULT 'moderator',
    added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    added_by_user_id BIGINT,
    permissions JSON,
    -- ...
);
```

**After:**
```sql
CREATE TABLE subforum_moderators (
    moderator_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    subforum_id INTEGER NOT NULL,
    user_id BIGINT NOT NULL, -- Real identity for administrative purposes
    pseudonym_id VARCHAR(64) NOT NULL, -- Pseudonym under which they moderate
    role VARCHAR(20) NOT NULL DEFAULT 'moderator',
    added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    added_by_user_id BIGINT, -- Real identity of who added them
    permissions JSON,
    -- ...
    UNIQUE KEY unique_moderator_pseudonym_subforum (subforum_id, pseudonym_id),
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE,
);
```

#### 2. `moderation_actions` Table

**Before:**
```sql
CREATE TABLE moderation_actions (
    action_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    moderator_user_id BIGINT NOT NULL, -- Only real identity
    -- ...
);
```

**After:**
```sql
CREATE TABLE moderation_actions (
    action_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    moderator_user_id BIGINT NOT NULL, -- Real identity for administrative purposes
    moderator_pseudonym_id VARCHAR(64) NOT NULL, -- Pseudonym under which action was performed
    -- ...
    FOREIGN KEY (moderator_pseudonym_id) REFERENCES pseudonyms(pseudonym_id),
);
```

#### 3. Content Tables Updates

Updated the following tables to include both real and pseudonymous moderator identities:

- `posts` - Added `removed_by_pseudonym_id`
- `comments` - Added `removed_by_pseudonym_id`
- `user_bans` - Added `banned_by_pseudonym_id`
- `reports` - Added `resolved_by_pseudonym_id`

### Benefits of These Changes

#### 1. Privacy Protection
- Moderators can maintain their pseudonymous identity while performing administrative functions
- Users see moderation actions from familiar pseudonymous community members
- Real identities remain protected for administrative oversight only

#### 2. Improved User Experience
- Users are more likely to trust moderation from established pseudonymous community members
- Moderators can build reputation and trust under their pseudonym
- Consistent pseudonymous interaction across all platform activities

#### 3. Administrative Oversight
- Real identities are still tracked for compliance and audit purposes
- Administrative functions can still correlate real identities when needed
- Audit trails maintain both real and pseudonymous identities for complete transparency

#### 4. Community Trust
- Moderators appear as community members rather than "faceless admins"
- Users can recognize and trust moderators they've interacted with
- Moderation feels more organic and community-driven

## API Multiple Pseudonyms Support

### Overview

This update focused on the API layer changes needed to support multiple pseudonyms, ensuring proper separation between user identity and pseudonym identity in API responses and endpoints.

### Key API Changes

#### 1. Updated Auth Models

**Before:**
```go
type UserInfo struct {
    UserID       int      `json:"user_id"`
    PseudonymID  string   `json:"pseudonym_id"`
    DisplayName  string   `json:"display_name"`
    Email        string   `json:"email"`
    KarmaScore   int      `json:"karma_score"`
    // ... mixed user and pseudonym data
}

type UserLoginResponse struct {
    Body struct {
        UserInfo
        TokenInfo
    } `json:"body"`
}
```

**After:**
```go
type UserInfo struct {
    UserID       int      `json:"user_id"`
    Email        string   `json:"email"`
    LastActiveAt string   `json:"last_active_at"`
    IsActive     bool     `json:"is_active"`
    IsSuspended  bool     `json:"is_suspended"`
    Roles        []string `json:"roles"`
    Capabilities []string `json:"capabilities"`
}

type PseudonymInfo struct {
    PseudonymID  string `json:"pseudonym_id"`
    DisplayName  string `json:"display_name"`
    KarmaScore   int    `json:"karma_score"`
    CreatedAt    string `json:"created_at"`
    LastActiveAt string `json:"last_active_at"`
    IsActive     bool   `json:"is_active"`
}

type UserLoginResponse struct {
    Body struct {
        UserInfo
        Pseudonyms []PseudonymInfo `json:"pseudonyms"`
        TokenInfo
    } `json:"body"`
}
```

#### 2. New User Handler Endpoints

- `GetPseudonymProfile` - Get public profile of a specific pseudonym
- `UpdatePseudonymProfile` - Update current user's pseudonym profile
- `CreatePseudonym` - Create a new pseudonym for the current user
- `GetUserProfile` - Get current user's profile with all pseudonyms

#### 3. Updated Auth Handler

- `RegisterUser` now creates both user and initial pseudonym
- `LoginUser` returns all pseudonyms for the user
- Updated function signatures to match new model structure

### Example API Usage

#### Creating a New Pseudonym
```http
POST /pseudonyms
Content-Type: application/json

{
  "display_name": "tech_guru",
  "bio": "Programming enthusiast",
  "website_url": "https://techguru.dev",
  "show_karma": true,
  "allow_direct_messages": true
}
```

#### Getting User Profile with All Pseudonyms
```http
GET /users/profile
Authorization: Bearer <jwt_token>
```

Response:
```json
{
  "body": {
    "user_id": 123,
    "email": "user@example.com",
    "roles": ["user"],
    "capabilities": ["create_content", "vote", "message", "report"],
    "pseudonyms": [
      {
        "pseudonym_id": "abc123def456...",
        "display_name": "user_display_name",
        "karma_score": 1250,
        "created_at": "2024-01-01T12:00:00Z",
        "is_active": true
      },
      {
        "pseudonym_id": "def789ghi012...",
        "display_name": "tech_guru",
        "karma_score": 500,
        "created_at": "2024-01-15T10:00:00Z",
        "is_active": true
      }
    ]
  }
}
```

## Implementation Considerations

### Data Migration

#### Existing User Records
- Existing user records need to be split into `users` and `pseudonyms` tables
- Default pseudonym can be created from existing display_name and karma_score
- Foreign key references need to be updated across all tables

#### Moderation Records
- Existing moderation records will need to be updated to include pseudonym IDs
- Default pseudonym IDs can be derived from existing user records
- Migration scripts should be tested thoroughly in staging environments

### Application Logic

#### Authentication
- Authentication remains at the user level
- Content creation and interaction switches to pseudonym level
- User management interfaces need to handle multiple pseudonyms per user

#### API Changes
- Content creation endpoints need to specify which pseudonym is creating content
- User profile endpoints need to handle multiple pseudonyms
- Administrative endpoints need to correlate pseudonyms to users

#### Moderation Interfaces
- Moderation interfaces should default to showing the moderator's pseudonym
- Administrative interfaces should have access to both real and pseudonymous identities
- API responses should include appropriate pseudonym information based on user permissions

### Security Considerations

#### Pseudonym Ownership
- Ensure that pseudonym IDs are validated against the user's actual pseudonyms
- Prevent moderators from using pseudonyms they don't own
- Maintain audit trails that link real identities to pseudonymous actions

#### Performance Impact
- Additional foreign key constraints may have minor performance impact
- New indexes on pseudonym fields should be monitored
- Query patterns may need optimization for the dual-identity system

## Example Usage Scenarios

### Scenario 1: Multiple Personas
```
User "alice@example.com" creates three pseudonyms:
- "tech_guru" (for programming discussions)
- "crypto_trader" (for cryptocurrency content)
- "bookworm" (for literature discussions)

Each pseudonym has its own:
- Display name and bio
- Karma score
- Post and comment history
- Subscriptions and preferences
```

### Scenario 2: Content Creation
```
1. User logs in with their real identity
2. User selects which pseudonym to post under
3. Post is created with pseudonym_id, not user_id
4. Other users see content from the pseudonym, not the real user
5. Administrative tools can correlate pseudonym to real user when needed
```

### Scenario 3: Moderation
```
1. Moderator "community_mod" removes a post
2. Action is logged with:
   - moderator_user_id: 123 (real identity for admin oversight)
   - moderator_pseudonym_id: "community_mod" (what users see)
3. Users see: "Post removed by community_mod"
4. Admins can see: "Post removed by community_mod (User ID: 123)"
```

## Summary

These changes fundamentally improve the HashPost platform by properly supporting multiple pseudonyms per user. The separation of user identity from pseudonym identity creates a clean, maintainable architecture that supports:

- True pseudonymous user experiences
- Administrative oversight and compliance
- Flexible correlation capabilities
- Scalable user management

The updated system aligns with the core principles of HashPost: providing robust pseudonymous user experiences while maintaining the administrative capabilities necessary for platform governance and legal compliance.

## Next Steps

### 1. Database Integration
- Update handlers to use the new database schema
- Implement proper joins between `users` and `pseudonyms` tables
- Add database queries for pseudonym management

### 2. Authentication Updates
- Update JWT token structure to include both `user_id` and `pseudonym_id`
- Implement pseudonym ownership validation
- Add context extraction for both user and pseudonym IDs

### 3. Testing
- Create unit tests for new pseudonym endpoints
- Test pseudonym creation and management flows
- Verify proper isolation between pseudonyms

### 4. Documentation
- Update API documentation with new endpoints
- Create migration guide for existing clients
- Document authentication changes

### 5. Client Updates
- Update frontend to handle multiple pseudonyms
- Implement pseudonym switching functionality
- Update profile management UI 