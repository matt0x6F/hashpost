#!/bin/bash

# HashPost API Testing Script using restish
# Make sure your server is running: make dev

set -e

echo "🚀 HashPost API Testing with restish"
echo "====================================="

# Function to print section headers
print_section() {
    echo ""
    echo "📋 $1"
    echo "-------------------------------------"
}

# 1. Health Check
print_section "Health Check"
restish hashpost health

# 2. Get Subforums (public endpoint)
print_section "Get Subforums"
restish hashpost get-subforums

# 3. Get specific subforum details
print_section "Get Subforum Details (tech)"
restish hashpost get-subforum-details tech

# 4. User Registration
print_section "User Registration"
restish hashpost register-user testuser test@example.com testpassword123

# 5. User Login
print_section "User Login"
echo "Logging in with testuser..."
LOGIN_RESPONSE=$(restish hashpost login-user testuser testpassword123)

echo "Login response: $LOGIN_RESPONSE"

echo ""
echo "🔑 The JWT token should be automatically handled by restish"
echo "You can now use authenticated endpoints!"

# 6. Get User Profile (requires auth)
print_section "Get User Profile"
restish hashpost get-user-profile

# 7. Subscribe to Subforum (requires auth)
print_section "Subscribe to Subforum"
restish hashpost subscribe-to-subforum tech

# 8. Create API Key (requires auth)
print_section "Create API Key"
restish hashpost post /auth/api-keys <<EOF
{
  "name": "Test API Key",
  "permissions": {
    "roles": ["user"],
    "capabilities": ["read", "write"]
  }
}
EOF

echo ""
echo "✅ Testing complete!"
echo ""
echo "💡 Tips:"
echo "- Use 'restish hashpost --help' to see all available commands"
echo "- Use 'restish hashpost --interactive' for interactive mode"
echo "- Check the OpenAPI docs at http://localhost:8888/docs"
echo "- Authentication is handled automatically by restish" 