//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommentWorkflow(t *testing.T) {
	t.Run("create post and comment, then fetch post with comments", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		defer suite.Cleanup()

		// Create test user with pseudonym
		testUser := suite.CreateTestUser(t, "test1@example.com", "password123", []string{"user"})

		// Create test subforum
		testSubforum := suite.CreateTestSubforum(t, "test-sub", "Test subforum", testUser.UserID, false)

		// Create test post
		testPost := suite.CreateTestPost(t, "Test Post", "Test post content", testSubforum.SubforumID, testUser.UserID, testUser.PseudonymID)

		// Create test comment
		testComment := suite.CreateTestComment(t, "Test comment content", testPost.PostID, testUser.UserID, testUser.PseudonymID, nil)

		// Verify comment was created
		require.NotNil(t, testComment)
		assert.Equal(t, "Test comment content", testComment.Content)
		assert.Equal(t, testPost.PostID, testComment.PostID)
		assert.Nil(t, testComment.ParentID) // Should be a root comment

		// Create a handler to test the API
		handler := handlers.NewContentHandler(suite.DB, suite.DB.DB, suite.IBESystem, suite.IdentityMappingDAO, suite.UserDAO)

		// Test: Get post details and verify comment appears
		ctx := context.Background()
		input := &models.PostDetailsInput{
			PostID: testPost.PostID,
			Sort:   "best",
		}

		response, err := handler.GetPostDetails(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, response)

		// Verify post details
		assert.Equal(t, testPost.PostID, int64(response.Body.PostID))
		assert.Equal(t, "Test Post", response.Body.Title)
		assert.Equal(t, "Test post content", response.Body.Content)

		// Verify comments are present
		require.Len(t, response.Body.Comments, 1, "Should have exactly one comment")

		comment := response.Body.Comments[0]
		assert.Equal(t, int(testComment.CommentID), comment.CommentID)
		assert.Equal(t, "Test comment content", comment.Content)
		assert.Nil(t, comment.ParentCommentID) // Should be a root comment
		assert.Equal(t, testUser.PseudonymID, comment.Author.PseudonymID)
		assert.Equal(t, testUser.DisplayName, comment.Author.DisplayName)
		assert.Empty(t, comment.Replies, "Root comment should have no replies")
	})

	t.Run("create post with nested comments", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		defer suite.Cleanup()

		// Create test user with pseudonym
		testUser := suite.CreateTestUser(t, "test2@example.com", "password123", []string{"user"})

		// Create test subforum
		testSubforum := suite.CreateTestSubforum(t, "test-sub", "Test subforum", testUser.UserID, false)

		// Create test post
		testPost := suite.CreateTestPost(t, "Test Post", "Test post content", testSubforum.SubforumID, testUser.UserID, testUser.PseudonymID)

		// Create root comment
		rootComment := suite.CreateTestComment(t, "Root comment", testPost.PostID, testUser.UserID, testUser.PseudonymID, nil)

		// Create reply to root comment
		replyComment := suite.CreateTestComment(t, "Reply to root", testPost.PostID, testUser.UserID, testUser.PseudonymID, &rootComment.CommentID)

		// Create handler
		handler := handlers.NewContentHandler(suite.DB, suite.DB.DB, suite.IBESystem, suite.IdentityMappingDAO, suite.UserDAO)

		// Get post details
		ctx := context.Background()
		input := &models.PostDetailsInput{
			PostID: testPost.PostID,
			Sort:   "best",
		}

		response, err := handler.GetPostDetails(ctx, input)
		require.NoError(t, err)
		require.NotNil(t, response)

		// Verify post details
		assert.Equal(t, testPost.PostID, int64(response.Body.PostID))

		// Verify comments structure
		require.Len(t, response.Body.Comments, 1, "Should have one root comment")

		rootCommentResponse := response.Body.Comments[0]
		assert.Equal(t, int(rootComment.CommentID), rootCommentResponse.CommentID)
		assert.Equal(t, "Root comment", rootCommentResponse.Content)
		assert.Nil(t, rootCommentResponse.ParentCommentID)

		// Verify reply is nested under root comment
		require.Len(t, rootCommentResponse.Replies, 1, "Root comment should have one reply")

		replyResponse := rootCommentResponse.Replies[0]
		assert.Equal(t, int(replyComment.CommentID), replyResponse.CommentID)
		assert.Equal(t, "Reply to root", replyResponse.Content)
		assert.Equal(t, int(rootComment.CommentID), *replyResponse.ParentCommentID)
	})

	t.Run("comment creation via API handler", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		defer suite.Cleanup()

		// Create test user with pseudonym
		testUser := suite.CreateTestUser(t, "test3@example.com", "password123", []string{"user"})

		// Create test subforum
		testSubforum := suite.CreateTestSubforum(t, "test-sub", "Test subforum", testUser.UserID, false)

		// Create test post
		testPost := suite.CreateTestPost(t, "Test Post", "Test post content", testSubforum.SubforumID, testUser.UserID, testUser.PseudonymID)

		// Create handler
		handler := handlers.NewContentHandler(suite.DB, suite.DB.DB, suite.IBESystem, suite.IdentityMappingDAO, suite.UserDAO)

		// Test comment creation via handler
		ctx := context.Background()
		commentInput := &models.CommentInput{
			PostID: testPost.PostID,
			Body: models.CommentInputBody{
				Content:         "API created comment",
				ParentCommentID: nil,
			},
		}

		// Add user context to AuthInput field by generating a JWT
		userCtx := &middleware.UserContext{
			UserID:            testUser.UserID,
			Email:             testUser.Email,
			ActivePseudonymID: testUser.PseudonymID,
			DisplayName:       testUser.DisplayName,
			Roles:             testUser.Roles,
			Capabilities:      testUser.Capabilities,
		}
		jwt, err := middleware.GenerateJWT(userCtx, suite.Config.JWT.Secret, 24*time.Hour)
		require.NoError(t, err)
		commentInput.AuthInput.AccessToken = jwt

		commentResponse, err := handler.CreateComment(ctx, commentInput)
		require.NoError(t, err)
		require.NotNil(t, commentResponse)

		// Verify comment was created
		assert.Equal(t, "API created comment", commentResponse.Body.Content)
		assert.Nil(t, commentResponse.Body.ParentCommentID)
		assert.Equal(t, testUser.PseudonymID, commentResponse.Body.Author.PseudonymID)
		assert.Equal(t, testUser.DisplayName, commentResponse.Body.Author.DisplayName)

		// Now fetch post details and verify comment appears
		postInput := &models.PostDetailsInput{
			PostID: testPost.PostID,
			Sort:   "best",
		}

		postResponse, err := handler.GetPostDetails(ctx, postInput)
		require.NoError(t, err)
		require.NotNil(t, postResponse)

		// Verify comment appears in post details
		require.Len(t, postResponse.Body.Comments, 1, "Should have one comment")

		comment := postResponse.Body.Comments[0]
		assert.Equal(t, commentResponse.Body.CommentID, comment.CommentID)
		assert.Equal(t, "API created comment", comment.Content)
		assert.Nil(t, comment.ParentCommentID)
		assert.Equal(t, testUser.PseudonymID, comment.Author.PseudonymID)
	})
}
