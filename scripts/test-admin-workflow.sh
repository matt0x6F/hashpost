#!/bin/bash

# Comprehensive admin workflow test script using restish
# Assumes admin user already exists from test-admin-creation.sh
set -e

# Configuration
ADMIN_EMAIL="testadmin@example.com"
ADMIN_PASSWORD="TestPassword123!"
SUBFORUM_NAME="test-subforum-$(date +%s)"
SUBFORUM_DESCRIPTION="A test subforum created by admin workflow"
POST_TITLE="Test Post by Admin"
POST_CONTENT="This is a test post created by the admin workflow script."
COMMENT_CONTENT="This is a test comment by the admin."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧪 HashPost Admin Workflow Test (using restish)${NC}"
echo "=================================================="
echo -e "${YELLOW}Note: Assumes admin user already exists from test-admin-creation.sh${NC}"

# Step 1: Get authentication token
echo -e "\n${YELLOW}Step 1: Getting authentication token...${NC}"
LOGIN_RESPONSE=$(restish hashpost login-user email:$ADMIN_EMAIL, password:$ADMIN_PASSWORD)

# Extract token from response (the token is directly in access_token)
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token // empty')

echo "Token: $TOKEN"

if [ -z "$TOKEN" ] || [ "$TOKEN" = "null" ]; then
    echo -e "${RED}❌ Failed to get authentication token${NC}"
    echo "Login response: $LOGIN_RESPONSE"
    echo "Make sure you've run test-admin-creation.sh first"
    exit 1
fi

echo -e "${GREEN}✅ Authentication token obtained${NC}"

# Step 2: Fetch or create subforum
echo -e "\n${YELLOW}Step 2: Fetching or creating subforum...${NC}"

# Try to fetch the subforum first
set +e
FETCH_SUBFORUM_RESPONSE=$(restish hashpost get-subforum-details $SUBFORUM_NAME -H "Authorization: Bearer $TOKEN")
FETCH_EXIT_CODE=$?
set -e
echo "Fetch response: $FETCH_SUBFORUM_RESPONSE"
EXISTING_SUBFORUM_SLUG=$(echo "$FETCH_SUBFORUM_RESPONSE" | jq -r '.slug // empty')
echo "Existing subforum slug: $EXISTING_SUBFORUM_SLUG"

if [ $FETCH_EXIT_CODE -eq 0 ] && [ -n "$EXISTING_SUBFORUM_SLUG" ] && [ "$EXISTING_SUBFORUM_SLUG" != "null" ]; then
    echo "Subforum already exists: $EXISTING_SUBFORUM_SLUG"
    ACTUAL_SUBFORUM_NAME="$EXISTING_SUBFORUM_SLUG"
else
    # Create the subforum
    CREATE_SUBFORUM_RESPONSE=$(restish hashpost create-subforum slug:$SUBFORUM_NAME, name:$SUBFORUM_NAME, description:$SUBFORUM_DESCRIPTION --authorization "Bearer $TOKEN")
    echo "Subforum created: $CREATE_SUBFORUM_RESPONSE"
    ACTUAL_SUBFORUM_NAME=$(echo "$CREATE_SUBFORUM_RESPONSE" | jq -r '.slug // empty')
    if [ -z "$ACTUAL_SUBFORUM_NAME" ] || [ "$ACTUAL_SUBFORUM_NAME" = "null" ]; then
        ACTUAL_SUBFORUM_NAME="$SUBFORUM_NAME"
    fi
fi

echo "Using subforum: $ACTUAL_SUBFORUM_NAME"

# Step 3: Create a post in the subforum
echo -e "\n${YELLOW}Step 3: Creating post...${NC}"

POST_RESPONSE=$(restish hashpost create-post $ACTUAL_SUBFORUM_NAME content:"$POST_CONTENT", is_nsfw:false, is_spoiler:false, post_type:text, title:"$POST_TITLE", url:"" --authorization "Bearer $TOKEN")
echo "Post created: $POST_RESPONSE"

# Extract post ID from response
POST_ID=$(echo "$POST_RESPONSE" | jq -r '.post_id // empty')
if [ -z "$POST_ID" ] || [ "$POST_ID" = "null" ]; then
    echo -e "${RED}❌ Failed to extract post ID${NC}"
    echo "Post response: $POST_RESPONSE"
    exit 1
fi
echo "Post ID: $POST_ID"

# Step 4: Fetch the post
echo -e "\n${YELLOW}Step 4: Fetching post...${NC}"
POST_DETAILS=$(restish hashpost get-post-details $POST_ID -H "Authorization: Bearer $TOKEN")
echo "Post details: $POST_DETAILS"

# Step 5: Create a comment on the post
echo -e "\n${YELLOW}Step 5: Creating comment...${NC}"

COMMENT_RESPONSE=$(restish hashpost create-comment $POST_ID content:"$COMMENT_CONTENT",parent_comment_id:null -H "Authorization: Bearer $TOKEN")
echo "Comment created: $COMMENT_RESPONSE"

# Extract comment ID from response
COMMENT_ID=$(echo "$COMMENT_RESPONSE" | jq -r '.comment_id // empty')
if [ -z "$COMMENT_ID" ] || [ "$COMMENT_ID" = "null" ]; then
    echo -e "${RED}❌ Failed to extract comment ID${NC}"
    echo "Comment response: $COMMENT_RESPONSE"
    exit 1
fi
echo "Comment ID: $COMMENT_ID"

# Step 6: Fetch the post again to see the comment
echo -e "\n${YELLOW}Step 6: Fetching post with comments...${NC}"
POST_WITH_COMMENTS=$(restish hashpost get-post-details $POST_ID -H "Authorization: Bearer $TOKEN")
echo "Post with comments: $POST_WITH_COMMENTS"

# Summary
echo -e "\n${BLUE}🎉 Admin Workflow Test Complete!${NC}"
echo "======================================"
echo -e "${GREEN}✅ Admin user: $ADMIN_EMAIL${NC}"
echo -e "${GREEN}✅ Subforum: $SUBFORUM_NAME${NC}"
echo -e "${GREEN}✅ Post: $POST_TITLE (ID: $POST_ID)${NC}"
echo -e "${GREEN}✅ Comment: $COMMENT_ID${NC}"
echo ""
echo "You can now test the web interface at: http://localhost:8888"
echo "Login with: $ADMIN_EMAIL / $ADMIN_PASSWORD" 