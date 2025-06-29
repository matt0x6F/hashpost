#!/bin/bash

# Test script for admin user creation inside Docker
set -e
set -x

ADMIN_EMAIL="testadmin@example.com"
ADMIN_PASSWORD="TestPassword123!"
ADMIN_ROLE="platform_admin"
ADMIN_DISPLAY_NAME="Test Admin"

echo "🧪 Testing HashPost Admin User Creation (Docker)"
echo "==============================================="

# Run the create-admin command inside the app container
CREATE_OUTPUT=$(docker compose exec -it app ./tmp/main create-admin \
  --non-interactive \
  --email "$ADMIN_EMAIL" \
  --password "$ADMIN_PASSWORD" \
  --role "$ADMIN_ROLE" \
  --display-name "$ADMIN_DISPLAY_NAME" \
  --mfa-enabled true 2>&1)

# Check for success message
if echo "$CREATE_OUTPUT" | grep -q "Admin user created successfully"; then
  echo "✅ Admin user creation succeeded inside Docker"
  echo "$CREATE_OUTPUT"
else
  echo "❌ Admin user creation failed inside Docker"
  echo "$CREATE_OUTPUT"
  exit 1
fi

echo ""
echo "🎉 Test complete! The CLI create-admin command works inside Docker."
echo "You can now log in as: $ADMIN_EMAIL" 