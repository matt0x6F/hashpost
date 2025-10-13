# AppView Database Schema

## Overview

The AppView database stores denormalized data optimized for read-heavy operations and user-facing queries. It maintains denormalized statistics and aggregated data for efficient API responses.

## Database Configuration

- **Database Name**: `hashpost_appview_dev` (development) / `hashpost_appview` (production)
- **Engine**: PostgreSQL
- **Extensions**: `uuid-ossp`, `pgcrypto`
- **Migration System**: SQLC with automated migrations

## Core Tables

### Users Table

**Table**: `appview_users`  
**Purpose**: Store denormalized user information for AppView

```sql
CREATE TABLE appview_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    did VARCHAR(255) UNIQUE NOT NULL,
    handle VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255),
    bio TEXT,
    avatar_url VARCHAR(500),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized stats
    post_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    reputation INTEGER DEFAULT 0
);
```

**Key Features**:
- **DID Storage**: Primary identifier linking to PDS
- **Handle Uniqueness**: Ensures unique usernames
- **Denormalized Stats**: Post count, comment count, reputation
- **Profile Data**: Display name, bio, avatar URL

### Subforums Table

**Table**: `appview_subforums`  
**Purpose**: Store denormalized subforum information

```sql
CREATE TABLE appview_subforums (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) UNIQUE NOT NULL,
    description TEXT,
    created_by_did VARCHAR(255) NOT NULL,
    created_by_handle VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized stats
    subscriber_count INTEGER DEFAULT 0,
    post_count INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0
);
```

**Key Features**:
- **Slug Uniqueness**: Unique forum identifiers
- **Creator Tracking**: Links to creator DID and handle
- **Denormalized Stats**: Subscriber count, post count, comment count
- **Metadata**: Name and description for forum display

### Posts Table

**Table**: `appview_posts`  
**Purpose**: Store denormalized post information

```sql
CREATE TABLE appview_posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    atproto_uri VARCHAR(500) UNIQUE NOT NULL,
    author_did VARCHAR(255) NOT NULL,
    author_handle VARCHAR(255) NOT NULL,
    subforum_slug VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized stats
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    comment_count INTEGER DEFAULT 0,
    score INTEGER DEFAULT 0 -- calculated: upvotes - downvotes
);
```

**Key Features**:
- **Atproto URI Storage**: Links to canonical PDS records
- **Author Information**: DID and handle for display
- **Subforum Association**: Links posts to subforums
- **Denormalized Stats**: Vote counts, comment count, calculated score

### Comments Table

**Table**: `appview_comments`  
**Purpose**: Store denormalized comment information

```sql
CREATE TABLE appview_comments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    atproto_uri VARCHAR(500) UNIQUE NOT NULL,
    author_did VARCHAR(255) NOT NULL,
    author_handle VARCHAR(255) NOT NULL,
    post_id UUID REFERENCES appview_posts(id) ON DELETE CASCADE,
    parent_id UUID REFERENCES appview_comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    -- Denormalized stats
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    score INTEGER DEFAULT 0
);
```

**Key Features**:
- **Hierarchical Comments**: Support for nested comment threads
- **Post Association**: Links comments to posts
- **Atproto URI Storage**: Links to canonical PDS records
- **Denormalized Stats**: Vote counts and calculated score

### Subscriptions Table

**Table**: `appview_subscriptions`  
**Purpose**: Track user subscriptions to subforums

```sql
CREATE TABLE appview_subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_did VARCHAR(255) NOT NULL,
    user_handle VARCHAR(255) NOT NULL,
    subforum_slug VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_did, subforum_slug)
);
```

**Key Features**:
- **Unique Subscriptions**: Prevents duplicate subscriptions
- **User Tracking**: Links to user DID and handle
- **Subforum Association**: Links to subforum slug

### Votes Table

**Table**: `appview_votes`  
**Purpose**: Store user votes on posts and comments

```sql
CREATE TABLE appview_votes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_did VARCHAR(255) NOT NULL,
    post_id UUID REFERENCES appview_posts(id) ON DELETE CASCADE,
    comment_id UUID REFERENCES appview_comments(id) ON DELETE CASCADE,
    vote_type VARCHAR(10) NOT NULL CHECK (vote_type IN ('up', 'down')),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(user_did, post_id),
    UNIQUE(user_did, comment_id),
    CHECK (
        (post_id IS NOT NULL AND comment_id IS NULL) OR 
        (post_id IS NULL AND comment_id IS NOT NULL)
    )
);
```

**Key Features**:
- **Vote Types**: Upvote or downvote
- **Content Association**: Links to posts or comments
- **User Tracking**: Links to user DID
- **Constraints**: Ensures votes are on either posts or comments, not both

## RBAC Tables

### Roles Table

**Table**: `roles`  
**Purpose**: Define available roles in the system

```sql
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    is_platform_role BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Permissions Table

**Table**: `permissions`  
**Purpose**: Define available permissions

```sql
CREATE TABLE permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    resource_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
```

### Role Permissions Table

**Table**: `role_permissions`  
**Purpose**: Define which permissions each role has

```sql
CREATE TABLE role_permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);
```

### User Roles Table

**Table**: `user_roles`  
**Purpose**: Assign roles to users

```sql
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_did VARCHAR(255) NOT NULL,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    subforum_id UUID REFERENCES appview_subforums(id) ON DELETE CASCADE,
    granted_by VARCHAR(255) NOT NULL,
    granted_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(user_did, role_id, subforum_id)
);
```

## Indexes

### Performance Indexes

```sql
-- User lookups
CREATE INDEX idx_appview_users_did ON appview_users(did);
CREATE INDEX idx_appview_users_handle ON appview_users(handle);

-- Subforum lookups
CREATE INDEX idx_appview_subforums_slug ON appview_subforums(slug);
CREATE INDEX idx_appview_subforums_created_by_did ON appview_subforums(created_by_did);

-- Post lookups
CREATE INDEX idx_appview_posts_atproto_uri ON appview_posts(atproto_uri);
CREATE INDEX idx_appview_posts_author_did ON appview_posts(author_did);
CREATE INDEX idx_appview_posts_subforum_slug ON appview_posts(subforum_slug);
CREATE INDEX idx_appview_posts_created_at ON appview_posts(created_at DESC);
CREATE INDEX idx_appview_posts_score ON appview_posts(score DESC);

-- Comment lookups
CREATE INDEX idx_appview_comments_atproto_uri ON appview_comments(atproto_uri);
CREATE INDEX idx_appview_comments_author_did ON appview_comments(author_did);
CREATE INDEX idx_appview_comments_post_id ON appview_comments(post_id);
CREATE INDEX idx_appview_comments_parent_id ON appview_comments(parent_id);
CREATE INDEX idx_appview_comments_created_at ON appview_comments(created_at DESC);

-- Subscription lookups
CREATE INDEX idx_appview_subscriptions_user_did ON appview_subscriptions(user_did);
CREATE INDEX idx_appview_subscriptions_subforum_slug ON appview_subscriptions(subforum_slug);

-- Vote lookups
CREATE INDEX idx_appview_votes_user_did ON appview_votes(user_did);
CREATE INDEX idx_appview_votes_post_id ON appview_votes(post_id);
CREATE INDEX idx_appview_votes_comment_id ON appview_votes(comment_id);

-- RBAC lookups
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_permissions_name ON permissions(name);
CREATE INDEX idx_user_roles_user_did ON user_roles(user_did);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
CREATE INDEX idx_user_roles_subforum_id ON user_roles(subforum_id);
```

## Data Relationships

### User Relationships
- **Users → Posts**: One-to-many (user can create multiple posts)
- **Users → Comments**: One-to-many (user can create multiple comments)
- **Users → Subscriptions**: One-to-many (user can subscribe to multiple subforums)
- **Users → Votes**: One-to-many (user can vote on multiple items)
- **Users → Roles**: Many-to-many (user can have multiple roles)

### Content Relationships
- **Subforums → Posts**: One-to-many (subforum contains multiple posts)
- **Posts → Comments**: One-to-many (post can have multiple comments)
- **Comments → Comments**: Self-referencing (nested comment threads)
- **Posts → Votes**: One-to-many (post can have multiple votes)
- **Comments → Votes**: One-to-many (comment can have multiple votes)

### RBAC Relationships
- **Roles → Permissions**: Many-to-many (role can have multiple permissions)
- **Users → Roles**: Many-to-many (user can have multiple roles)
- **Subforums → User Roles**: One-to-many (subforum can have multiple user roles)

## Denormalized Data Strategy

### Statistics Columns
- **User Stats**: `post_count`, `comment_count`, `reputation`
- **Subforum Stats**: `subscriber_count`, `post_count`, `comment_count`
- **Post Stats**: `upvotes`, `downvotes`, `comment_count`, `score`
- **Comment Stats**: `upvotes`, `downvotes`, `score`

### Update Triggers
Statistics are updated through:
- **Event Processing**: AppView updates stats when processing PDS events
- **Application Logic**: Handlers update stats when creating/updating content
- **Batch Updates**: Periodic batch updates for consistency

### Consistency Considerations
- **Eventual Consistency**: Stats may be slightly out of sync
- **Batch Reconciliation**: Periodic reconciliation of stats
- **Error Handling**: Graceful handling of stat update failures

## Migration Management

### Migration Files
- **Location**: `internal/database/migrations/appview/`
- **Naming**: `001_appview_schema.up.sql`, `002_rbac_schema.up.sql`, etc.
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

## Performance Considerations

### Query Optimization
- **Indexes**: Comprehensive indexing for common query patterns
- **Denormalization**: Pre-computed stats for fast queries
- **Pagination**: LIMIT/OFFSET for large result sets

### Data Consistency
- **Event Processing**: Maintain consistency through event processing
- **Batch Updates**: Periodic batch updates for stats
- **Error Handling**: Graceful handling of consistency issues

### Scalability
- **Read Replicas**: Consider read replicas for heavy read workloads
- **Partitioning**: Consider partitioning for large tables
- **Caching**: Application-level caching for frequently accessed data

## References

- [Database Migrations](internal/database/migrations/appview/)
- [SQLC Queries](internal/database/queries/appview/)
- [Generated Models](internal/database/generated/appview/)
- [Event Processing](internal/appview/events.go)
