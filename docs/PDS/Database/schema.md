# PDS Database Schema

## Overview

The PDS database stores canonical atproto records, user information, and forum-specific tables. It follows the atproto specification for data storage with forum extensions, including proper URI storage and relational integrity.

## Database Configuration

- **Database Name**: `hashpost_pds_dev` (development) / `hashpost_pds` (production)
- **Engine**: PostgreSQL
- **Extensions**: `uuid-ossp`, `pgcrypto`
- **Migration System**: SQLC with automated migrations

## Core Tables

### Users Table

**Table**: `users`  
**Purpose**: Store atproto user accounts with DID and handle information

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    did VARCHAR(255) UNIQUE NOT NULL, -- Decentralized Identifier
    handle VARCHAR(255) UNIQUE NOT NULL, -- Username/handle
    email VARCHAR(255) UNIQUE,
    password_hash TEXT, -- bcrypt hashed password
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Key Features**:
- **DID Storage**: Primary identifier for atproto compliance
- **Handle Uniqueness**: Ensures unique usernames
- **Password Hashing**: Secure password storage with bcrypt
- **Email Support**: Optional email for user identification

### User Sessions Table

**Table**: `user_sessions`  
**Purpose**: Manage user authentication sessions

```sql
CREATE TABLE user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    session_id VARCHAR(255) UNIQUE NOT NULL,
    user_did VARCHAR(255) NOT NULL,
    handle VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    last_accessed_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Key Features**:
- **Session Management**: Track active user sessions
- **Expiration**: Automatic session expiration
- **Access Tracking**: Last accessed timestamp
- **DID Binding**: Sessions bound to user DIDs

### Subforums Table

**Table**: `subforums`  
**Purpose**: Store HashPost-specific forum categories

```sql
CREATE TABLE subforums (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users(id),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Key Features**:
- **Slug Uniqueness**: Unique forum identifiers
- **Creator Tracking**: Links to user who created subforum
- **Metadata**: Name and description for forum display

### Posts Table

**Table**: `posts`  
**Purpose**: Store atproto post records with URI references

```sql
CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    subforum_id UUID REFERENCES subforums(id) ON DELETE CASCADE,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    atproto_uri VARCHAR(500) UNIQUE, -- atproto record URI
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Key Features**:
- **Atproto URI Storage**: Canonical atproto record URIs
- **User Association**: Links posts to users
- **Subforum Association**: Organizes posts by subforum
- **Content Storage**: Title and content for posts

### Comments Table

**Table**: `comments`  
**Purpose**: Store atproto comment records

```sql
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    atproto_uri VARCHAR(500) UNIQUE, -- atproto record URI
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

**Key Features**:
- **Hierarchical Comments**: Support for nested comment threads
- **Post Association**: Links comments to posts
- **Atproto URI Storage**: Canonical record URIs
- **Cascade Deletion**: Automatic cleanup on post deletion

### Subforum Subscriptions Table

**Table**: `subforum_subscriptions`  
**Purpose**: Track user subscriptions to subforums

```sql
CREATE TABLE subforum_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    subforum_id UUID REFERENCES subforums(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_id, subforum_id)
);
```

**Key Features**:
- **Unique Subscriptions**: Prevents duplicate subscriptions
- **Cascade Deletion**: Automatic cleanup on user/subforum deletion
- **Timestamp Tracking**: Subscription creation time

## OAuth Schema

### OAuth Clients Table

**Table**: `oauth_clients`  
**Purpose**: Store OAuth 2.0 client applications

```sql
CREATE TABLE oauth_clients (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    client_id VARCHAR(255) UNIQUE NOT NULL,
    client_secret VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    redirect_uri VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### OAuth Authorization Codes Table

**Table**: `oauth_authorization_codes`  
**Purpose**: Store OAuth authorization codes

```sql
CREATE TABLE oauth_authorization_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    user_did VARCHAR(255) NOT NULL,
    redirect_uri VARCHAR(500),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

### OAuth Access Tokens Table

**Table**: `oauth_access_tokens`  
**Purpose**: Store OAuth access tokens

```sql
CREATE TABLE oauth_access_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    access_token VARCHAR(255) UNIQUE NOT NULL,
    client_id VARCHAR(255) NOT NULL,
    user_did VARCHAR(255) NOT NULL,
    scope VARCHAR(500),
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
```

## Indexes

### Performance Indexes

```sql
-- User lookups
CREATE INDEX idx_users_did ON users(did);
CREATE INDEX idx_users_handle ON users(handle);
CREATE INDEX idx_users_email ON users(email);

-- Session management
CREATE INDEX idx_user_sessions_session_id ON user_sessions(session_id);
CREATE INDEX idx_user_sessions_user_did ON user_sessions(user_did);
CREATE INDEX idx_user_sessions_expires_at ON user_sessions(expires_at);

-- Content queries
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_subforum_id ON posts(subforum_id);
CREATE INDEX idx_posts_atproto_uri ON posts(atproto_uri);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);

-- Comment queries
CREATE INDEX idx_comments_post_id ON comments(post_id);
CREATE INDEX idx_comments_user_id ON comments(user_id);
CREATE INDEX idx_comments_parent_id ON comments(parent_id);
CREATE INDEX idx_comments_atproto_uri ON comments(atproto_uri);

-- Subforum queries
CREATE INDEX idx_subforums_slug ON subforums(slug);
CREATE INDEX idx_subforums_created_by ON subforums(created_by);

-- Subscription queries
CREATE INDEX idx_subforum_subscriptions_user_id ON subforum_subscriptions(user_id);
CREATE INDEX idx_subforum_subscriptions_subforum_id ON subforum_subscriptions(subforum_id);
```

## Data Relationships

### User Relationships
- **Users → Sessions**: One-to-many (user can have multiple sessions)
- **Users → Posts**: One-to-many (user can create multiple posts)
- **Users → Comments**: One-to-many (user can create multiple comments)
- **Users → Subscriptions**: One-to-many (user can subscribe to multiple subforums)

### Content Relationships
- **Subforums → Posts**: One-to-many (subforum contains multiple posts)
- **Posts → Comments**: One-to-many (post can have multiple comments)
- **Comments → Comments**: Self-referencing (nested comment threads)

### OAuth Relationships
- **Clients → Authorization Codes**: One-to-many
- **Clients → Access Tokens**: One-to-many
- **Users → Authorization Codes**: One-to-many
- **Users → Access Tokens**: One-to-many

## Atproto Compliance

### URI Storage
All atproto records are stored with their canonical URIs:
- **Format**: `at://{did}/{collection}/{rkey}`
- **Uniqueness**: URIs must be unique across the system
- **Validation**: URIs are validated against atproto specifications

### DID Integration
- **Primary Keys**: Users identified by DID, not internal IDs
- **Foreign Keys**: All relationships reference user DIDs
- **Resolution**: DIDs resolved through identity directory

### Record Management
- **CRUD Operations**: Full create, read, update, delete support
- **Versioning**: Record updates tracked with timestamps
- **Deletion**: Soft deletion with proper cleanup

## Migration Management

### Migration Files
- **Location**: `internal/database/migrations/pds/`
- **Naming**: `001_initial_schema.up.sql`, `002_oauth_schema.up.sql`, etc.
- **Rollback**: Corresponding `.down.sql` files for rollbacks

### Migration Commands
```bash
# Apply migrations
task migrate:up

# Rollback migrations
task migrate:down

# Generate SQLC code
task generate:sqlc
```

## References

- [Database Migrations](internal/database/migrations/pds/)
- [SQLC Queries](internal/database/queries/pds/)
- [Generated Models](internal/database/generated/pds/)
- [Atproto Data Model](https://atproto.com/specs/data-model)
