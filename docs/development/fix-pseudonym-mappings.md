# Fix Pseudonym Mappings

This document explains how to fix missing identity mappings for existing pseudonyms using the new `update-admin` command.

## Problem

If your pseudonym was created before the IBE system was properly configured, it may be missing the required identity mappings for:
- **Authentication** (for login/session management)
- **Self-correlation** (for user self-verification)
- **Correlation** (for admin roles - user, moderator, platform_admin)

## Solution

Use the new `update-admin` command to fix your pseudonym mappings:

### Basic Usage

```bash
# Update your admin user and fix missing mappings
./bin/hashpost update-admin --email your-email@example.com --role platform_admin --fix-mappings
```

### Interactive Mode

```bash
# Run interactively (will prompt for inputs)
./bin/hashpost update-admin
```

### Non-Interactive Mode

```bash
# Run with all flags specified
./bin/hashpost update-admin --email your-email@example.com --role platform_admin --fix-mappings --non-interactive
```

## Available Roles

- `platform_admin` - Full platform administration access
- `trust_safety` - Trust and safety team access
- `legal_team` - Legal team access

## What the Command Does

1. **Updates User Role**: Sets your user's role in the database
2. **Fixes Identity Mappings**: Creates missing identity mappings for your pseudonyms
3. **Uses Correct Domains**: Applies the proper IBE domains based on your role
4. **Idempotent**: Safe to run multiple times

## Example Output

```
Enter admin email: your-email@example.com
Enter admin role (platform_admin, trust_safety, legal_team): platform_admin
Fix missing identity mappings for pseudonyms? (y/N): y

✅ Admin user updated successfully
✅ Pseudonym mappings fixed successfully
```

## Manual Verification

After running the command, you can verify the fix by checking your identity mappings:

```sql
-- Connect to the database
docker-compose exec postgres psql -U hashpost -d hashpost

-- Check your identity mappings
SELECT * FROM identity_mappings WHERE user_id = YOUR_USER_ID;
```

## Troubleshooting

### If the command fails:

1. **Check your email**: Make sure you're using the correct email address
2. **Verify database connection**: Ensure the database is running
3. **Check role keys**: Run `make roles-setup` if role keys aren't configured
4. **Review logs**: Check the application logs for detailed error messages

### If mappings are still missing:

The command will log which mappings were created. If you see warnings about failed mapping creation, it may indicate:
- Missing role keys for your role
- IBE system configuration issues
- Database permission problems

## Related Commands

- `make roles-setup` - Set up role keys if missing
- `./bin/hashpost roles list` - List available roles
- `./bin/hashpost roles keys` - List active role keys

## Security Notes

- The command requires database access
- Identity mappings are encrypted with role-specific keys
- Only run this command on your own account
- The command is idempotent and safe to run multiple times 