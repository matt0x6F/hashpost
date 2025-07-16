# HashPost API Documentation

## Overview

This document provides comprehensive API documentation for HashPost, a Reddit-like social media platform that uses Identity-Based Encryption (IBE) to provide pseudonymous user profiles while maintaining administrative accountability. The API uses a single-user system with role-based access control.

## API Architecture

### Base URL
```
https://api.hashpost.com/v1
```

### Authentication
HashPost uses a single authentication system with role-based access control:
- **Single Authentication**: All users authenticate through the same endpoint
- **Role-Based Capabilities**: User capabilities are determined by their roles
- **MFA Requirements**: Sensitive operations require MFA based on user roles and actions

### Response Format
All API responses follow this structure:
```json
{
  "success": true,
  "data": {},
  "error": null,
  "timestamp": "2024-01-01T12:00:00Z"
}
```

## Authentication Endpoints

### User Registration

#### POST /auth/register
Register a new user account.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "secure_password",
  "display_name": "user_display_name",
  "bio": "Optional user bio",
  "website_url": "https://example.com",
  "timezone": "UTC",
  "language": "en"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": 123,
    "pseudonym_id": "abc123def456...",
    "display_name": "user_display_name",
    "email": "user@example.com",
    "karma_score": 0,
    "created_at": "2024-01-01T12:00:00Z",
    "roles": ["user"],
    "capabilities": ["create_content", "vote", "message", "report"],
    "access_token": "jwt_token_here",
    "refresh_token": "refresh_token_here"
  }
}
```

### User Login

#### POST /auth/login
Authenticate a user and receive access tokens with role-based capabilities.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "user_id": 123,
    "pseudonym_id": "abc123def456...",
    "display_name": "user_display_name",
    "roles": ["user", "moderator"],
    "capabilities": ["create_content", "vote", "message", "report", "moderate_content", "ban_users", "remove_content", "correlate_fingerprints"],
    "access_token": "jwt_token_here",
    "refresh_token": "refresh_token_here",
    "expires_in": 3600
  }
}
```

**Note:** The response includes all roles and capabilities for the user. Users with administrative roles (moderator, trust_safety, etc.) will have additional capabilities in their response.

### Token Refresh

#### POST /auth/refresh
Refresh an expired access token.

**Request Body:**
```json
{
  "refresh_token": "refresh_token_here"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "access_token": "new_jwt_token_here",
    "expires_in": 3600
  }
}
```

## User Management Endpoints

### Get User Profile

#### GET /users/{pseudonym_id}
Get a user's public profile.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "pseudonym_id": "abc123def456...",
    "display_name": "user_display_name",
    "karma_score": 1250,
    "created_at": "2024-01-01T12:00:00Z",
    "bio": "User bio text",
    "website_url": "https://example.com",
    "show_karma": true,
    "allow_direct_messages": true,
    "post_count": 45,
    "comment_count": 230
  }
}
```

### Update User Profile

#### PUT /users/profile
Update the current user's profile.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "display_name": "new_display_name",
  "bio": "Updated bio text",
  "website_url": "https://newwebsite.com",
  "show_karma": false,
  "allow_direct_messages": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "pseudonym_id": "abc123def456...",
    "display_name": "new_display_name",
    "bio": "Updated bio text",
    "website_url": "https://newwebsite.com",
    "show_karma": false,
    "allow_direct_messages": false,
    "updated_at": "2024-01-01T13:00:00Z"
  }
}
```

### Get User Preferences

#### GET /users/preferences
Get the current user's preferences.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "timezone": "UTC",
    "language": "en",
    "theme": "light",
    "email_notifications": true,
    "push_notifications": true,
    "auto_hide_nsfw": true,
    "auto_hide_spoilers": true
  }
}
```

### Update User Preferences

#### PUT /users/preferences
Update the current user's preferences.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "timezone": "America/New_York",
  "theme": "dark",
  "email_notifications": false,
  "auto_hide_nsfw": false
}
```

## Subforum Endpoints

### Get Subforums

#### GET /subforums
Get a list of subforums.

**Query Parameters:**
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 25)
- `sort` (string): Sort by 'subscribers', 'posts', 'new' (default: 'subscribers')

**Response:**
```json
{
  "success": true,
  "data": {
    "subforums": [
      {
        "subforum_id": 1,
        "name": "golang",
        "display_name": "Golang",
        "description": "The Go programming language",
        "subscriber_count": 125000,
        "post_count": 45000,
        "is_nsfw": false,
        "is_private": false,
        "created_at": "2020-01-01T00:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 25,
      "total": 1500,
      "pages": 60
    }
  }
}
```

### Get Subforum Details

#### GET /subforums/{name}
Get detailed information about a specific subforum.

**Response:**
```json
{
  "success": true,
  "data": {
    "subforum_id": 1,
    "name": "golang",
    "display_name": "Golang",
    "description": "The Go programming language",
    "sidebar_text": "Welcome to r/golang...",
    "rules_text": "1. Be respectful...",
    "subscriber_count": 125000,
    "post_count": 45000,
    "is_nsfw": false,
    "is_private": false,
    "is_restricted": false,
    "created_at": "2020-01-01T00:00:00Z",
    "moderators": [
      {
        "pseudonym_id": "mod1_pseudonym_id",
        "display_name": "moderator1",
        "role": "owner"
      }
    ],
    "is_subscribed": true,
    "is_favorite": false
  }
}
```

### Subscribe to Subforum

#### POST /subforums/{name}/subscribe
Subscribe to a subforum.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "subforum_id": 1,
    "name": "golang",
    "subscribed": true,
    "subscriber_count": 125001
  }
}
```

### Unsubscribe from Subforum

#### DELETE /subforums/{name}/subscribe
Unsubscribe from a subforum.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "subforum_id": 1,
    "name": "golang",
    "subscribed": false,
    "subscriber_count": 125000
  }
}
```

## Content Endpoints

### Get Posts

#### GET /subforums/{name}/posts
Get posts from a subforum.

**Query Parameters:**
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 25)
- `sort` (string): Sort by 'hot', 'new', 'top', 'rising' (default: 'hot')
- `time` (string): Time filter for 'top' sort: 'hour', 'day', 'week', 'month', 'year', 'all' (default: 'day')

**Response:**
```json
{
  "success": true,
  "data": {
    "posts": [
      {
        "post_id": 123,
        "title": "Post Title",
        "content": "Post content...",
        "post_type": "text",
        "url": null,
        "is_self_post": true,
        "is_nsfw": false,
        "is_spoiler": false,
        "score": 1250,
        "upvotes": 1300,
        "downvotes": 50,
        "comment_count": 45,
        "view_count": 5000,
        "created_at": "2024-01-01T12:00:00Z",
        "author": {
          "pseudonym_id": "abc123def456...",
          "display_name": "user_display_name"
        },
        "user_vote": 1, // 1 for upvote, -1 for downvote, 0 for no vote
        "is_saved": false
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 25,
      "total": 45000,
      "pages": 1800
    }
  }
}
```

### Create Post

#### POST /subforums/{name}/posts
Create a new post.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "title": "Post Title",
  "content": "Post content text...",
  "post_type": "text", // "text", "link", "image", "video", "poll"
  "url": "https://example.com", // Required for link posts
  "is_nsfw": false,
  "is_spoiler": false
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "post_id": 124,
    "title": "Post Title",
    "content": "Post content text...",
    "post_type": "text",
    "score": 0,
    "comment_count": 0,
    "created_at": "2024-01-01T14:00:00Z",
    "author": {
      "pseudonym_id": "abc123def456...",
      "display_name": "user_display_name"
    }
  }
}
```

### Get Post Details

#### GET /posts/{post_id}
Get detailed information about a specific post.

**Query Parameters:**
- `sort` (string): Comment sort order: 'best', 'top', 'new', 'controversial', 'old', 'qa' (default: 'best')

**Response:**
```json
{
  "success": true,
  "data": {
    "post_id": 123,
    "title": "Post Title",
    "content": "Post content...",
    "post_type": "text",
    "url": null,
    "is_self_post": true,
    "is_nsfw": false,
    "is_spoiler": false,
    "score": 1250,
    "upvotes": 1300,
    "downvotes": 50,
    "comment_count": 45,
    "view_count": 5000,
    "created_at": "2024-01-01T12:00:00Z",
    "author": {
      "pseudonym_id": "abc123def456...",
      "display_name": "user_display_name"
    },
    "subforum": {
      "subforum_id": 1,
      "name": "golang",
      "display_name": "Golang"
    },
    "user_vote": 1,
    "is_saved": false,
    "comments": [
      {
        "comment_id": 456,
        "content": "Comment text...",
        "score": 25,
        "created_at": "2024-01-01T12:30:00Z",
        "author": {
          "pseudonym_id": "def789ghi012...",
          "display_name": "commenter_name"
        },
        "user_vote": 0,
        "replies": []
      }
    ]
  }
}
```

### Vote on Post

#### POST /posts/{post_id}/vote
Vote on a post.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "vote_value": 1 // 1 for upvote, -1 for downvote, 0 to remove vote
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "post_id": 123,
    "vote_value": 1,
    "score": 1251,
    "upvotes": 1301,
    "downvotes": 50
  }
}
```

### Create Comment

#### POST /posts/{post_id}/comments
Create a comment on a post.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "content": "Comment text...",
  "parent_comment_id": 456 // Optional, for replies
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "comment_id": 789,
    "content": "Comment text...",
    "parent_comment_id": 456,
    "score": 0,
    "created_at": "2024-01-01T15:00:00Z",
    "author": {
      "pseudonym_id": "abc123def456...",
      "display_name": "user_display_name"
    }
  }
}
```

### Vote on Comment

#### POST /comments/{comment_id}/vote
Vote on a comment.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "vote_value": 1 // 1 for upvote, -1 for downvote, 0 to remove vote
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "comment_id": 789,
    "vote_value": 1,
    "score": 1,
    "upvotes": 1,
    "downvotes": 0
  }
}
```

## Moderation Endpoints

### Report Content

#### POST /reports
Report content or users.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "content_type": "post", // "post", "comment", "user", "subforum"
  "content_id": 123, // Required for post/comment reports
  "reported_pseudonym_id": "def789ghi012...", // Required for user reports
  "report_reason": "spam", // "spam", "harassment", "violence", "misinformation", etc.
  "report_details": "This post violates community guidelines..."
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "report_id": 789,
    "status": "pending",
    "created_at": "2024-01-01T16:00:00Z"
  }
}
```

### Get Reports (Moderators)

#### GET /moderation/reports
Get reports for moderation review.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `subforum_id` (integer): Filter by subforum
- `status` (string): Filter by status: 'pending', 'investigating', 'resolved', 'dismissed'
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 25)

**Response:**
```json
{
  "success": true,
  "data": {
    "reports": [
      {
        "report_id": 789,
        "content_type": "post",
        "content_id": 123,
        "reported_pseudonym_id": "def789ghi012...",
        "report_reason": "spam",
        "report_details": "This post violates community guidelines...",
        "status": "pending",
        "created_at": "2024-01-01T16:00:00Z",
        "resolved_by": {
          "pseudonym_id": "mod_pseudonym_id",
          "display_name": "moderator_name"
        },
        "resolved_at": "2024-01-01T17:00:00Z",
        "resolution_notes": "Post removed for violation of community guidelines",
        "reporter": {
          "pseudonym_id": "reporter_pseudonym_id",
          "display_name": "reporter_name"
        },
        "reported_user": {
          "pseudonym_id": "reported_pseudonym_id",
          "display_name": "reported_user_name"
        },
        "content": {
          "title": "Reported Post Title",
          "content": "Reported post content..."
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 25,
      "total": 150,
      "pages": 6
    }
  }
}
```

### Remove Content (Moderators)

#### POST /moderation/content/{content_type}/{content_id}/remove
Remove content as a moderator.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "removal_reason": "violates community guidelines",
  "send_notification": true
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "content_id": 123,
    "content_type": "post",
    "removed": true,
    "removal_reason": "violates community guidelines",
    "removed_at": "2024-01-01T17:00:00Z",
    "removed_by": {
      "pseudonym_id": "mod_pseudonym_id",
      "display_name": "moderator_name"
    }
  }
}
```

### Ban User (Moderators)

#### POST /moderation/users/{pseudonym_id}/ban
Ban a user from a subforum. (The client only knows pseudonym_id, never user_id.)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "subforum_id": 1,
  "ban_reason": "Repeated violations of community guidelines",
  "is_permanent": false,
  "duration_days": 30,
  "send_notification": true
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "ban_id": 123,
    "banned_fingerprint": "a1b2c3d4e5f6...",
    "subforum_id": 1,
    "ban_reason": "Repeated violations of community guidelines",
    "is_permanent": false,
    "expires_at": "2024-02-01T17:00:00Z",
    "created_at": "2024-01-01T17:00:00Z",
    "banned_by": {
      "pseudonym_id": "mod_pseudonym_id",
      "display_name": "moderator_name"
    }
  }
}
```

### Get Moderation History (Moderators)

#### GET /moderation/history
Get moderation action history for the authenticated moderator.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `subforum_id` (integer): Filter by subforum
- `action_type` (string): Filter by action type: 'remove_post', 'remove_comment', 'ban_user', 'unban_user'
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 25)

**Response:**
```json
{
  "success": true,
  "data": {
    "actions": [
      {
        "action_id": 123,
        "action_type": "remove_post",
        "target_content_type": "post",
        "target_content_id": 456,
        "action_details": {
          "removal_reason": "violates community guidelines"
        },
        "created_at": "2024-01-01T17:00:00Z",
        "moderator": {
          "pseudonym_id": "mod_pseudonym_id",
          "display_name": "moderator_name"
        },
        "subforum": {
          "subforum_id": 1,
          "name": "golang",
          "display_name": "Golang"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 25,
      "total": 150,
      "pages": 6
    }
  }
}
```

## Administrative Correlation Endpoints

### Request Fingerprint Correlation (Moderators)

#### POST /admin/correlation/fingerprint
Request fingerprint-based correlation for moderation purposes.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "requested_pseudonym": "abc123def456...",
  "requested_fingerprint": "a1b2c3d4e5f6...",
  "justification": "Investigation of ban evasion in r/golang",
  "subforum_id": 1,
  "incident_id": "ban_evasion_123"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "correlation_id": "uuid_here",
    "correlation_type": "fingerprint",
    "scope": "subforum_specific",
    "time_window": "30_days",
    "status": "completed",
    "results": [
      {
        "pseudonym_id": "def789ghi012...",
        "display_name": "suspected_user",
        "created_at": "2024-01-01T10:00:00Z",
        "posts_in_subforum": 5,
        "comments_in_subforum": 12
      }
    ],
    "audit_id": "audit_uuid_here"
  }
}
```

### Request Identity Correlation (Admins)

#### POST /admin/correlation/identity
Request identity-based correlation for platform-wide investigations.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "requested_pseudonym": "abc123def456...",
  "requested_fingerprint": "a1b2c3d4e5f6...",
  "justification": "Investigation of reported harassment across subforums",
  "legal_basis": "Platform Terms of Service",
  "incident_id": "harassment_case_123",
  "scope": "platform_wide"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "correlation_id": "uuid_here",
    "correlation_type": "identity",
    "scope": "platform_wide",
    "time_window": "unlimited",
    "status": "completed",
    "results": [
      {
        "pseudonym_id": "def789ghi012...",
        "display_name": "suspected_user",
        "encrypted_real_identity": "encrypted_data_here",
        "created_at": "2024-01-01T10:00:00Z",
        "total_posts": 45,
        "total_comments": 230,
        "subforums_active": ["golang", "programming", "tech"]
      }
    ],
    "audit_id": "audit_uuid_here"
  }
}
```

### Get Correlation History

#### GET /admin/correlation/history
Get correlation request history for the authenticated user.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `correlation_type` (string): Filter by type: 'fingerprint', 'identity'
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 25)

**Response:**
```json
{
  "success": true,
  "data": {
    "correlations": [
      {
        "correlation_id": "uuid_here",
        "correlation_type": "fingerprint",
        "requested_pseudonym": "abc123def456...",
        "justification": "Investigation of ban evasion",
        "status": "completed",
        "timestamp": "2024-01-01T16:00:00Z",
        "results_count": 2
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 25,
      "total": 45,
      "pages": 2
    }
  }
}
```

## User Interaction Endpoints

### Block User

#### POST /users/{pseudonym_id}/block
Block a pseudonym. (Client-side: always blocks by pseudonym. To block all personas, the backend will correlate and create the appropriate block records, but the client never submits or receives a user_id.)

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  // No user_id here! Only pseudonym_id is used.
  // Optionally, the client can request to block all personas:
  "block_all_personas": true // (optional, handled by backend)
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "blocked_pseudonym_id": "def789ghi012...",
    "blocked_user_fingerprint": "a1b2c3d4e5f6...", // User fingerprint (from real identity) when block_all_personas=true
    "blocked_at": "2024-01-01T18:00:00Z"
  }
}
```

-- Note: If `block_all_personas` is true, the backend will correlate and block all pseudonyms for the user, but the client never sees or submits a user_id.

### Unblock User

#### DELETE /users/{pseudonym_id}/block
Unblock a user.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Response:**
```json
{
  "success": true,
  "data": {
    "blocked_user_id": 456,
    "blocked_pseudonym_id": "def789ghi012...",
    "unblocked_at": "2024-01-01T19:00:00Z"
  }
}
```

### Send Direct Message

#### POST /messages
Send a direct message to another user.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Request Body:**
```json
{
  "recipient_pseudonym_id": "def789ghi012...",
  "content": "Hello! I wanted to discuss..."
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "message_id": 123,
    "recipient_pseudonym_id": "def789ghi012...",
    "content": "Hello! I wanted to discuss...",
    "created_at": "2024-01-01T20:00:00Z"
  }
}
```

### Get Direct Messages

#### GET /messages
Get direct messages for the current user.

**Headers:**
```
Authorization: Bearer <access_token>
```

**Query Parameters:**
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 25)

**Response:**
```json
{
  "success": true,
  "data": {
    "messages": [
      {
        "message_id": 123,
        "sender_pseudonym_id": "def789ghi012...",
        "sender_display_name": "sender_name",
        "content": "Hello! I wanted to discuss...",
        "is_read": false,
        "created_at": "2024-01-01T20:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 25,
      "total": 50,
      "pages": 2
    }
  }
}
```

## Search Endpoints

### Search Posts

#### GET /search/posts
Search for posts across all subforums.

**Query Parameters:**
- `q` (string): Search query (required)
- `subforum` (string): Filter by subforum name
- `author` (string): Filter by author pseudonym
- `sort` (string): Sort by 'relevance', 'hot', 'top', 'new', 'comments' (default: 'relevance')
- `time` (string): Time filter: 'hour', 'day', 'week', 'month', 'year', 'all' (default: 'all')
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 25)

**Response:**
```json
{
  "success": true,
  "data": {
    "query": "golang concurrency",
    "posts": [
      {
        "post_id": 123,
        "title": "Understanding Golang Concurrency",
        "content": "Post content about golang concurrency...",
        "score": 1250,
        "comment_count": 45,
        "created_at": "2024-01-01T12:00:00Z",
        "author": {
          "pseudonym_id": "abc123def456...",
          "display_name": "user_display_name"
        },
        "subforum": {
          "name": "golang",
          "display_name": "Golang"
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 25,
      "total": 150,
      "pages": 6
    }
  }
}
```

### Search Users

#### GET /search/users
Search for users by display name.

**Query Parameters:**
- `q` (string): Search query (required)
- `page` (integer): Page number (default: 1)
- `limit` (integer): Items per page (default: 25)

**Response:**
```json
{
  "success": true,
  "data": {
    "query": "john",
    "users": [
      {
        "pseudonym_id": "abc123def456...",
        "display_name": "john_doe",
        "karma_score": 1250,
        "created_at": "2024-01-01T12:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 25,
      "total": 45,
      "pages": 2
    }
  }
}
```

## Error Handling

### Error Response Format
```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "field": "email",
      "issue": "Email format is invalid"
    }
  },
  "timestamp": "2024-01-01T12:00:00Z"
}
```

### Common Error Codes

| Code | Description | HTTP Status |
|------|-------------|-------------|
| `AUTHENTICATION_REQUIRED` | Authentication token is required | 401 |
| `INVALID_TOKEN` | Authentication token is invalid or expired | 401 |
| `INSUFFICIENT_PERMISSIONS` | User lacks required permissions | 403 |
| `MFA_REQUIRED` | Multi-factor authentication required | 403 |
| `VALIDATION_ERROR` | Request validation failed | 400 |
| `RESOURCE_NOT_FOUND` | Requested resource not found | 404 |
| `RATE_LIMIT_EXCEEDED` | Rate limit exceeded | 429 |
| `INTERNAL_ERROR` | Internal server error | 500 |

### Rate Limiting

API endpoints are rate-limited to prevent abuse:

- **Regular endpoints**: 1000 requests per hour per user
- **Authentication endpoints**: 10 requests per hour per IP
- **Moderation endpoints**: 100 requests per hour per moderator
- **Correlation endpoints**: 50 requests per day per admin

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## WebSocket Endpoints

### Real-time Notifications

#### WebSocket Connection
```
wss://api.hashpost.com/v1/ws/notifications
```

**Authentication:**
Send authentication message after connection:
```json
{
  "type": "auth",
  "token": "access_token_here"
}
```

**Message Types:**
- `notification`: New notification
- `message`: New direct message
- `vote_update`: Vote count update
- `comment`: New comment on user's post

**Example Message:**
```json
{
  "type": "notification",
  "data": {
    "notification_id": 123,
    "type": "comment",
    "title": "New comment on your post",
    "body": "user_display_name commented on your post",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

## API Versioning

The API uses semantic versioning. The current version is v1. Future versions will be available at:
- `https://api.hashpost.com/v2`
- `https://api.hashpost.com/v3`

Breaking changes will only be introduced in major version updates.

## SDKs and Libraries

Official SDKs are available for:
- JavaScript/TypeScript
- Python
- Go
- Java
- C#

Example usage with JavaScript SDK:
```javascript
import { HashPostAPI } from '@hashpost/sdk';

const api = new HashPostAPI({
  baseURL: 'https://api.hashpost.com/v1',
  accessToken: 'your_access_token'
});

// Get posts from a subforum
const posts = await api.subforums.getPosts('golang', {
  sort: 'hot',
  limit: 25
});

// Create a post
const post = await api.subforums.createPost('golang', {
  title: 'My Post Title',
  content: 'Post content...',
  post_type: 'text'
});
```

This API documentation provides comprehensive coverage of all endpoints and functionality for the HashPost platform, ensuring developers can effectively integrate with the single-user system with role-based access control.

## Role-Based Security Model

### User Roles and Capabilities

HashPost uses a role-based access control system where users can have multiple roles, each granting specific capabilities:

| Role | Capabilities | MFA Required For |
|------|-------------|------------------|
| **User** | create_content, vote, message, report | None |
| **Moderator** | All user + moderate_content, ban_users, remove_content, correlate_fingerprints | correlate_fingerprints |
| **Subforum Owner** | All moderator + manage_moderators | correlate_fingerprints |
| **Trust & Safety** | correlate_identities, cross_platform_access, system_moderation | All correlation actions |
| **Legal Team** | correlate_identities, legal_compliance, court_orders | All correlation actions |
| **Platform Admin** | system_admin, user_management, correlate_identities | All administrative actions |

### MFA Requirements

Multi-factor authentication is required for sensitive operations based on:

1. **User Roles**: Users with correlation capabilities require MFA for those operations
2. **Action Type**: Certain actions (system_admin, legal_compliance) always require MFA
3. **Scope**: Platform-wide operations require MFA regardless of role

### Authentication Flow

1. **Login**: User authenticates with email/password
2. **Role Resolution**: System determines user's roles and capabilities
3. **Token Generation**: JWT token includes roles and capabilities
4. **API Access**: Each endpoint checks required capabilities
5. **MFA Validation**: Sensitive operations validate MFA if required

### Example: Moderator Login

```json
// POST /auth/login
{
  "email": "moderator@example.com",
  "password": "secure_password"
}

// Response includes all capabilities
{
  "success": true,
  "data": {
    "user_id": 123,
    "roles": ["user", "moderator"],
    "capabilities": [
      "create_content", "vote", "message", "report",
      "moderate_content", "ban_users", "remove_content", "correlate_fingerprints"
    ],
    "access_token": "jwt_token_here"
  }
}
```

When this moderator tries to access correlation endpoints, the system will:
1. Check if they have the `correlate_fingerprints` capability ✓
2. Check if the action requires MFA ✓
3. Validate MFA token if provided
4. Allow or deny access accordingly

## User Management Endpoints