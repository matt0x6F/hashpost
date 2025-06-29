# HashPost Scripts

This directory contains utility scripts for the HashPost project.

## API Testing with restish

### Setup

1. **Install restish** (if not already installed):
   ```bash
   go install github.com/danielgtaylor/restish@latest
   ```

2. **Configure the HashPost API**:
   ```bash
   restish api configure hashpost http://localhost:8888
   ```

3. **Start the HashPost server**:
   ```bash
   make dev
   ```

### Usage

#### Quick Start
```bash
# Run the comprehensive test script
./scripts/test-api.sh
```

#### Manual Testing
```bash
# Health check
restish hashpost health

# Get subforums
restish hashpost get-subforums

# Get specific subforum
restish hashpost get-subforum-details --name tech
```

#### Interactive Mode
```bash
# Interactive exploration
restish hashpost --interactive
```

#### Authentication Examples

**User Registration and Login:**
```bash
# Register a new user
restish hashpost register-user --username myuser --email my@example.com --password mypassword

# Login (restish handles JWT tokens automatically)
restish hashpost login-user --username myuser --password mypassword

# Get user profile (requires authentication)
restish hashpost get-user-profile
```

**Subforum Management:**
```bash
# Subscribe to a subforum
restish hashpost subscribe-to-subforum --name tech

# Unsubscribe from a subforum
restish hashpost unsubscribe-from-subforum --name tech
```

**Content Creation:**
```bash
# Create a post
restish hashpost create-post --subforum tech --title "My Post" --content "Post content"

# Create a comment
restish hashpost create-comment --post-id 123 --content "Great post!"
```

### Available Commands

restish automatically discovers all your API endpoints from the OpenAPI spec. Here are the main categories:

#### Authentication
- `register-user` - Register a new user account
- `login-user` - Authenticate a user
- `logout-user` - Logout a user
- `refresh-token` - Refresh an expired access token

#### Subforums
- `get-subforums` - Get a list of subforums
- `get-subforum-details` - Get detailed information about a specific subforum
- `subscribe-to-subforum` - Subscribe to a subforum
- `unsubscribe-from-subforum` - Unsubscribe from a subforum

#### Content
- `create-post` - Create a new post
- `get-post-details` - Get detailed information about a specific post
- `get-subforum-posts` - Get posts from a subforum
- `create-comment` - Create a comment on a post
- `vote-on-post` - Vote on a post
- `vote-on-comment` - Vote on a comment

#### Users
- `get-user-profile` - Get the current user's profile
- `get-user-preferences` - Get the current user's preferences
- `update-user-preferences` - Update the current user's preferences
- `block-user` - Block a pseudonym
- `unblock-user` - Unblock a user

#### Search
- `search-posts` - Search for posts across all subforums
- `search-users` - Search for users by display name

#### Moderation
- `report-content` - Report content or users
- `get-reports` - Get reports for moderation review
- `ban-user` - Ban a user from a subforum
- `remove-content` - Remove content as a moderator

### Tips

1. **Interactive Mode**: Use `--interactive` flag for guided API exploration
2. **Help**: Use `restish hashpost --help` to see all available commands
3. **Authentication**: restish handles JWT tokens automatically after login
4. **OpenAPI Integration**: Commands are generated from your OpenAPI spec
5. **Output Formats**: Use `-o` flag to change output format (json, table, etc.)

### Troubleshooting

- **Server not running**: Make sure to run `make dev` first
- **API not configured**: Run `restish api configure hashpost http://localhost:8888`
- **Authentication errors**: Make sure you're logged in with `login-user`
- **Command not found**: Check `restish hashpost --help` for available commands 