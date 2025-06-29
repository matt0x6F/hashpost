#!/bin/bash

# HashPost API Testing Script using restish
# Make sure your server is running: make dev

set -e

BASE_URL="http://localhost:8888"
RESTISH_CONFIG="restish-config.yaml"

echo "🚀 HashPost API Testing with restish"
echo "====================================="

# Function to print section headers
print_section() {
    echo ""
    echo "📋 $1"
    echo "-------------------------------------"
}

# Function to run restish command with config
run_restish() {
    restish --config $RESTISH_CONFIG "$@"
}

# 1. Health Check
print_section "Health Check"
run_restish get /health

# 2. Get Subforums (public endpoint)
print_section "Get Subforums"
run_restish get /subforums

# 3. Get specific subforum details
print_section "Get Subforum Details (tech)"
run_restish get /subforums/tech

# 4. User Registration
print_section "User Registration"
run_restish post /auth/register <<EOF
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "testpassword123"
}
EOF

# 5. User Login
print_section "User Login"
LOGIN_RESPONSE=$(run_restish post /auth/login <<EOF
{
  "username": "testuser",
  "password": "testpassword123"
}
EOF
)

echo "Login response: $LOGIN_RESPONSE"

# Extract JWT token (you'll need to manually copy this for subsequent requests)
echo ""
echo "🔑 Copy the JWT token from above and use it in subsequent requests"
echo "Example: restish --config $RESTISH_CONFIG --auth-bearer YOUR_TOKEN get /users/me"

# 6. Get User Profile (requires auth)
print_section "Get User Profile (requires JWT token)"
echo "Run this after getting your JWT token:"
echo "restish --config $RESTISH_CONFIG --auth-bearer YOUR_TOKEN get /users/me"

# 7. Subscribe to Subforum (requires auth)
print_section "Subscribe to Subforum (requires JWT token)"
echo "Run this after getting your JWT token:"
echo "restish --config $RESTISH_CONFIG --auth-bearer YOUR_TOKEN post /subforums/tech/subscribe"

# 8. Create API Key (requires auth)
print_section "Create API Key (requires JWT token)"
echo "Run this after getting your JWT token:"
echo 'restish --config $RESTISH_CONFIG --auth-bearer YOUR_TOKEN post /auth/api-keys <<EOF'
echo '{'
echo '  "name": "Test API Key",'
echo '  "permissions": {'
echo '    "roles": ["user"],'
echo '    "capabilities": ["read", "write"]'
echo '  }'
echo '}'
echo 'EOF'

# 9. Test with API Key
print_section "Test with API Key"
echo "After creating an API key, test it with:"
echo "restish --config $RESTISH_CONFIG --auth-bearer YOUR_API_KEY get /users/me"

echo ""
echo "✅ Testing complete!"
echo ""
echo "💡 Tips:"
echo "- Use 'restish --help' for more options"
echo "- Use 'restish --config $RESTISH_CONFIG --interactive' for interactive mode"
echo "- Check the OpenAPI docs at http://localhost:8888/docs" 