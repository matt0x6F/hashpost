# Comment Creation Workflow

This document outlines the complete workflow for creating, managing, and moderating comments in the HashPost platform.

## Overview

The comment system supports:
- **Creating comments** on posts with optional parent comments for threaded discussions
- **Voting** on comments (upvote/downvote)
- **Editing** comments (planned feature)
- **Moderation** capabilities for removing comments
- **Nested replies** for threaded conversations
- **Authentication** using pseudonyms for privacy

## Database Schema

### Comments Table

```sql
CREATE TABLE comments (
    comment_id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL,
    parent_comment_id BIGINT,
    pseudonym_id VARCHAR(64) NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    score INTEGER DEFAULT 0,
    upvotes INTEGER DEFAULT 0,
    downvotes INTEGER DEFAULT 0,
    is_edited BOOLEAN DEFAULT FALSE,
    edited_at TIMESTAMP WITH TIME ZONE,
    edit_reason VARCHAR(100),
    
    -- Moderation fields
    is_removed BOOLEAN DEFAULT FALSE,
    removed_by_user_id BIGINT,
    removed_by_pseudonym_id VARCHAR(64),
    removal_reason VARCHAR(100),
    removed_at TIMESTAMP WITH TIME ZONE,
    
    FOREIGN KEY (post_id) REFERENCES posts(post_id) ON DELETE CASCADE,
    FOREIGN KEY (parent_comment_id) REFERENCES comments(comment_id) ON DELETE CASCADE,
    FOREIGN KEY (pseudonym_id) REFERENCES pseudonyms(pseudonym_id) ON DELETE CASCADE,
    FOREIGN KEY (removed_by_user_id) REFERENCES users(user_id),
    FOREIGN KEY (removed_by_pseudonym_id) REFERENCES pseudonyms(pseudonym_id)
);
```

## API Endpoints

### 1. Create Comment

**Endpoint:** `POST /posts/{post_id}/comments`

**Authentication:** Required (JWT token)

**Request Body:**
```json
{
  "content": "This is my comment text",
  "parent_comment_id": 123  // Optional, for replies
}
```

**Response:**
```json
{
  "comment_id": 456,
  "content": "This is my comment text",
  "parent_comment_id": 123,
  "score": 0,
  "created_at": "2024-01-01T15:00:00Z",
  "author": {
    "pseudonym_id": "abc123def456...",
    "display_name": "user_display_name"
  }
}
```

**Validation Rules:**
- Content cannot be empty
- Post must exist and not be removed/locked
- Parent comment (if provided) must exist and belong to the same post
- User must be authenticated with an active pseudonym

### 2. Vote on Comment

**Endpoint:** `POST /comments/{comment_id}/vote`

**Authentication:** Required (JWT token)

**Request Body:**
```json
{
  "vote_value": 1  // 1 for upvote, -1 for downvote, 0 to remove vote
}
```

**Response:**
```json
{
  "comment_id": 456,
  "vote_value": 1,
  "score": 1,
  "upvotes": 1,
  "downvotes": 0
}
```

### 3. Get Post with Comments

**Endpoint:** `GET /posts/{post_id}`

**Query Parameters:**
- `sort`: "best", "top", "new", "controversial", "old", "qa"

**Response:**
```json
{
  "post": {
    "post_id": 123,
    "title": "Post Title",
    "content": "Post content...",
    "score": 25,
    "comment_count": 5,
    "author": {
      "pseudonym_id": "abc123def456...",
      "display_name": "author_name"
    }
  },
  "comments": [
    {
      "comment_id": 456,
      "content": "Comment text...",
      "parent_comment_id": null,
      "score": 5,
      "created_at": "2024-01-01T15:00:00Z",
      "author": {
        "pseudonym_id": "def789ghi012...",
        "display_name": "commenter_name"
      },
      "user_vote": 1,
      "replies": [
        {
          "comment_id": 789,
          "content": "Reply to comment...",
          "parent_comment_id": 456,
          "score": 2,
          "created_at": "2024-01-01T16:00:00Z",
          "author": {
            "pseudonym_id": "ghi345jkl678...",
            "display_name": "replier_name"
          },
          "user_vote": 0,
          "replies": []
        }
      ]
    }
  ]
}
```

## Implementation Details

### Handler Structure

The comment workflow is implemented in `internal/api/handlers/content.go`:

```go
type ContentHandler struct {
    db                 bob.Executor
    rawDB              *sql.DB
    ibeSystem          *ibe.IBESystem
    identityMappingDAO *dao.IdentityMappingDAO
    userDAO            *dao.UserDAO
    postDAO            *dao.PostDAO
    commentDAO         *dao.CommentDAO
    subforumDAO        *dao.SubforumDAO
    securePseudonymDAO *dao.SecurePseudonymDAO
    voteDAO            *dao.VoteDAO
    permissionChecker  *middleware.PermissionChecker
}
```

### Key Methods

#### CreateComment
```go
func (h *ContentHandler) CreateComment(ctx context.Context, input *models.CommentInput) (*models.CommentResponse, error)
```

**Workflow:**
1. Extract user context from JWT token
2. Validate input (content not empty)
3. Check if post exists and is not removed/locked
4. Validate parent comment if provided
5. Create comment in database
6. Update post comment count
7. Return comment response

#### VoteOnComment
```go
func (h *ContentHandler) VoteOnComment(ctx context.Context, input *models.CommentVoteInput) (*models.CommentVoteResponse, error)
```

**Workflow:**
1. Extract user context from JWT token
2. Validate vote value (-1, 0, or 1)
3. Check if comment exists and is not removed
4. Handle vote (create/update/delete)
5. Update comment score and vote counts
6. Return updated vote response

### Data Access Layer

The comment DAO (`internal/database/dao/comments.go`) provides:

```go
type CommentDAO struct {
    db bob.Executor
}

// Key methods:
func (dao *CommentDAO) CreateComment(ctx context.Context, postID int64, pseudonymID, content string, parentCommentID *int64) (*models.Comment, error)
func (dao *CommentDAO) GetCommentByID(ctx context.Context, commentID int64) (*models.Comment, error)
func (dao *CommentDAO) GetCommentsByPost(ctx context.Context, postID int64, sort string) ([]*models.Comment, error)
func (dao *CommentDAO) UpdateCommentScore(ctx context.Context, commentID int64, score, upvotes, downvotes int32) error
func (dao *CommentDAO) CountCommentsByPost(ctx context.Context, postID int64) (int64, error)
```

## Security Features

### Authentication & Privacy
- Comments are created using pseudonyms, not real user identities
- JWT tokens provide authentication while preserving privacy
- IBE (Identity-Based Encryption) system secures identity mappings

### Moderation Capabilities
- Comments can be removed by moderators
- Removal tracking includes moderator pseudonym and reason
- Removed comments are hidden from public view but preserved for audit

### Permission System
- RBAC (Role-Based Access Control) controls moderation actions
- Different permission levels: user, moderator, admin
- Subforum-specific moderation permissions

## Testing

### Integration Tests

The comment workflow is thoroughly tested in `internal/api/integration/comment_integration_test.go`:

```go
func TestCommentWorkflow(t *testing.T) {
    t.Run("create post and comment, then fetch post with comments", func(t *testing.T) {
        // Test basic comment creation and retrieval
    })
    
    t.Run("create post with nested comments", func(t *testing.T) {
        // Test threaded comment functionality
    })
    
    t.Run("comment creation via API handler", func(t *testing.T) {
        // Test full API workflow
    })
}
```

### Test Utilities

The test suite provides helper methods:

```go
// Create test comment
testComment := suite.CreateTestComment(t, "Test comment content", testPost.PostID, testUser.UserID, testUser.PseudonymID, nil)

// Create test post
testPost := suite.CreateTestPost(t, "Test Post", "Test post content", testSubforum.SubforumID, testUser.UserID, testUser.PseudonymID)
```

## Error Handling

### Common Error Scenarios

1. **Authentication Required (401)**
   - No JWT token provided
   - Invalid or expired token

2. **Bad Request (400)**
   - Empty comment content
   - Invalid vote value
   - Invalid parent comment ID

3. **Not Found (404)**
   - Post doesn't exist
   - Parent comment doesn't exist
   - Comment doesn't exist

4. **Forbidden (403)**
   - Post is locked
   - Post is removed
   - Comment is removed (for voting)

### Error Response Format

```json
{
  "error": {
    "code": "validation_error",
    "message": "content is required",
    "details": {
      "field": "content",
      "value": ""
    }
  }
}
```

## Performance Considerations

### Database Indexes
```sql
-- Optimized indexes for comment queries
CREATE INDEX idx_comments_post ON comments(post_id);
CREATE INDEX idx_comments_parent ON comments(parent_comment_id);
CREATE INDEX idx_comments_pseudonym ON comments(pseudonym_id);
CREATE INDEX idx_comments_created_at ON comments(created_at);
CREATE INDEX idx_comments_score ON comments(score);
CREATE INDEX idx_comments_post_score ON comments(post_id, score);
```

### Caching Strategy
- Comment counts are cached at the post level
- Vote summaries are calculated on-demand
- Comment trees are built efficiently using parent-child relationships

## Future Enhancements

### Planned Features

1. **Comment Editing**
   - Allow users to edit their own comments
   - Track edit history and reasons
   - Show "edited" indicator

2. **Comment Moderation**
   - Implement comment removal by moderators
   - Add comment reporting system
   - Create moderation queue

3. **Advanced Features**
   - Comment sorting options
   - Comment search functionality
   - Comment notifications
   - Comment bookmarks

### API Extensions

```go
// Planned endpoints:
PATCH /comments/{comment_id}          // Edit comment
DELETE /comments/{comment_id}          // Delete comment (owner only)
POST /comments/{comment_id}/report     // Report comment
GET /comments/{comment_id}/history     // Get edit history
```

## Usage Examples

### Creating a Comment

```bash
curl -X POST "http://localhost:8080/posts/123/comments" \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "Great post! Thanks for sharing.",
    "parent_comment_id": null
  }'
```

### Creating a Reply

```bash
curl -X POST "http://localhost:8080/posts/123/comments" \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "content": "I agree with your point.",
    "parent_comment_id": 456
  }'
```

### Voting on a Comment

```bash
curl -X POST "http://localhost:8080/comments/456/vote" \
  -H "Authorization: Bearer <jwt_token>" \
  -H "Content-Type: application/json" \
  -d '{
    "vote_value": 1
  }'
```

## Integration with Other Systems

### Post System
- Comments automatically update post comment count
- Post locking prevents new comments
- Post removal affects comment visibility

### User System
- Comments are tied to pseudonyms, not real identities
- User blocking affects comment visibility
- User roles determine moderation capabilities

### Moderation System
- Comments can be reported for review
- Moderators can remove comments
- Removal actions are tracked for audit

This comprehensive comment workflow provides a robust foundation for user discussions while maintaining privacy and enabling effective moderation. 