# HashPost Database ERD (Entity Relationship Diagram)

This document contains the PlantUML ERD for the HashPost database schema, providing a visual representation of all tables and their relationships.

## Files

- `database-erd.puml` - The PlantUML source code for the ERD
- `database-schema.md` - The complete database schema documentation

## How to Generate the ERD

### Option 1: Online PlantUML Editor

1. Go to [PlantUML Online Editor](http://www.plantuml.com/plantuml/uml/)
2. Copy the contents of `database-erd.puml`
3. Paste into the editor
4. The diagram will be generated automatically

### Option 2: VS Code Extension

1. Install the "PlantUML" extension in VS Code
2. Open `database-erd.puml`
3. Press `Alt+Shift+D` to preview the diagram
4. Or right-click and select "Preview Current Diagram"

### Option 3: Command Line

If you have PlantUML installed locally:

```bash
# Install PlantUML (if not already installed)
# On Ubuntu/Debian:
sudo apt-get install plantuml

# On macOS with Homebrew:
brew install plantuml

# Generate PNG
plantuml database-erd.puml

# Generate SVG
plantuml -tsvg database-erd.puml
```

### Option 4: Docker

```bash
docker run -d -p 8080:8080 plantuml/plantuml-server
# Then visit http://localhost:8080 and paste the PlantUML code
```

## Diagram Structure

The ERD is organized into logical sections:

### Core User Tables
- `users` - Main user accounts with role-based capabilities
- `user_preferences` - User-specific settings and preferences

### Identity Management Tables
- `identity_mappings` - Encrypted mappings between real identities and pseudonyms
- `role_keys` - Role-based keys for correlation and administrative access

### Community Tables
- `subforums` - Community spaces (equivalent to Reddit's subreddits)
- `subforum_subscriptions` - User subscriptions to subforums
- `subforum_moderators` - Moderator relationships

### Content Tables
- `posts` - All user posts
- `comments` - Comments on posts (with hierarchical structure)
- `votes` - User votes on posts and comments

### Media and Attachments
- `media_attachments` - Media files attached to posts
- `polls` - Poll data for poll-type posts
- `poll_votes` - Individual poll votes

### User Interaction Tables
- `user_blocks` - User blocking relationships
- `direct_messages` - Direct messages between users

### Moderation Tables
- `reports` - User reports of content or users
- `user_bans` - User bans from subforums
- `moderation_actions` - Logs of all moderation actions

### Audit and Compliance Tables
- `correlation_audit` - Logs of correlation activities
- `key_usage_audit` - Tracks usage of role-based keys
- `compliance_reports` - Legal request documentation
- `compliance_correlations` - Links compliance reports to correlations

### System Tables
- `system_settings` - Global system configuration
- `api_keys` - API keys for external integrations
- `system_events` - System-level event logs
- `performance_metrics` - Performance monitoring data
- `role_definitions` - Role definitions and capabilities

## Key Features of the ERD

### Color Coding
- All tables are displayed with a light red background (`#FFAAAA`)
- Primary keys are marked with `PK`
- Unique keys are marked with `UK`

### Relationship Types
- `||--o{` - One-to-many relationship
- `||--||` - One-to-one relationship
- Self-referencing relationships (e.g., comments replying to comments)

### Security Features Highlighted
- Identity-based encryption tables
- Role-based access control
- Audit and compliance tracking
- Moderation and reporting systems

## Notes

1. **Encrypted Fields**: Fields marked as `BYTEA` contain encrypted data
2. **JSON Fields**: Many fields use JSON for flexible data storage
3. **UUID Primary Keys**: Audit and compliance tables use UUIDs for better security
4. **Self-Referencing**: Comments can have parent comments, creating a hierarchical structure
5. **Role-Based Access**: The system uses a sophisticated RBAC system with correlation capabilities

## Updating the ERD

When the database schema changes:

1. Update the `database-schema.md` file
2. Update the corresponding table definitions in `database-erd.puml`
3. Add or modify relationships as needed
4. Regenerate the diagram using one of the methods above

## Troubleshooting

### Common Issues

1. **Diagram too large**: The ERD is comprehensive and may be large. Consider generating it as SVG for better zoom capabilities.

2. **PlantUML syntax errors**: Ensure all table definitions are properly closed with `}` and relationships use correct syntax.

3. **Missing relationships**: If you notice missing relationships, add them to the appropriate section in the PlantUML file.

### Performance

For large diagrams, consider:
- Using SVG format for better performance
- Breaking the diagram into smaller, focused sections
- Using PlantUML's `!pragma` directives for optimization

## Related Documentation

- [Database Schema Documentation](database-schema.md)
- [API Documentation](api-documentation.md)
- [Identity-Based Encryption Documentation](identity-based-encryption.md)
- [Database Operations Guide](database-operations.md) 