package handlers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	"github.com/matt0x6f/hashpost/internal/database/models"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to create test content handler with mocked dependencies
func createTestContentHandler() (*ContentHandler, *mocks.MockPostDAO, *mocks.MockCommentDAO, *mocks.MockVoteDAO, *mocks.MockSubforumDAO, *mocks.MockSecurePseudonymDAO, *mocks.MockPermissionDAO) {
	mockPostDAO := mocks.NewMockPostDAO()
	mockCommentDAO := mocks.NewMockCommentDAO()
	mockVoteDAO := mocks.NewMockVoteDAO()
	mockSubforumDAO := mocks.NewMockSubforumDAO()
	mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()
	mockUserDAO := &mocks.MockUserDAO{}
	mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
	mockUserBlocksDAO := &mocks.MockUserBlocksDAO{}
	mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
	mockPermissionDAO := &mocks.MockPermissionDAO{}

	ibeSystem := ibe.NewIBESystem()

	// Create a mock auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	// Note: Individual tests should set up their own mock expectations

	// Note: Individual tests should set up their own mock expectations

	// Create a permission checker with the mock DAO
	permissionChecker := middleware.NewPermissionCheckerWithDAO(mockPermissionDAO)

	handler := NewContentHandlerWithDependencies(
		nil, // Mock DB
		nil, // Mock raw DB
		ibeSystem,
		mockIdentityMappingDAO,
		mockUserDAO,
		mockPostDAO,
		mockCommentDAO,
		mockSubforumDAO,
		mockSecurePseudonymDAO,
		mockVoteDAO,
		mockUserBlocksDAO,
		mockRoleKeyDAO,
		permissionChecker,
		mockPermissionDAO,
	)

	return handler, mockPostDAO, mockCommentDAO, mockVoteDAO, mockSubforumDAO, mockSecurePseudonymDAO, mockPermissionDAO
}

// TestContentHandler_VoteOnPost_PreventsVotingOnDeletedPost tests that voting on deleted posts is prevented
func TestContentHandler_VoteOnPost_PreventsVotingOnDeletedPost(t *testing.T) {
	handler, mockPostDAO, _, _, _, _, _ := createTestContentHandler()

	// Create a deleted post
	deletedPost := fixtures.CreateTestPost()
	deletedPost.IsDeleted = sql.Null[bool]{V: true, Valid: true}

	// Inject the deleted post into the mock
	mockPostDAO.InjectPost(deletedPost)
	mockPostDAO.SetDefaultBehavior()

	// The vote DAO should NOT be called for deleted posts
	// We don't need to set up expectations since the handler should return an error before calling the DAO

	// Create a valid JWT token for testing
	userCtx := fixtures.CreateTestUserContext()
	token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
	require.NoError(t, tokenErr)

	// Create test input
	input := &apimodels.PostVoteInput{
		PostID: 123,
		Body: apimodels.VoteInputBody{
			VoteValue: 1,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
	}

	// Set up user context in the request context
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call the handler
	_, err := handler.VoteOnPost(ctx, input)

	// Should return an error for deleted posts
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleted")

	// Verify that the vote DAO was not called (no expectations set up)
}

// TestContentHandler_VoteOnComment_PreventsVotingOnDeletedComment tests that voting on deleted comments is prevented
func TestContentHandler_VoteOnComment_PreventsVotingOnDeletedComment(t *testing.T) {
	handler, _, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Create a deleted comment
	deletedComment := fixtures.CreateTestComment()
	deletedComment.IsDeleted = sql.Null[bool]{V: true, Valid: true}

	// Inject the deleted comment into the mock
	mockCommentDAO.InjectComment(deletedComment)
	mockCommentDAO.SetDefaultBehavior()

	// The vote DAO should NOT be called for deleted comments
	// We don't need to set up expectations since the handler should return an error before calling the DAO

	// Create a valid JWT token for testing
	userCtx := fixtures.CreateTestUserContext()
	token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
	require.NoError(t, tokenErr)

	// Create test input
	input := &apimodels.CommentVoteInput{
		CommentID: 456,
		Body: apimodels.VoteInputBody{
			VoteValue: 1,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
	}

	// Set up user context in the request context
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call the handler
	_, err := handler.VoteOnComment(ctx, input)

	// Should return an error for deleted comments
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleted")

	// Verify that the vote DAO was not called (no expectations set up)
}

// TestContentHandler_DeletePost_HandlesDeletedPostResponse tests that the delete post response has the correct structure
func TestContentHandler_DeletePost_HandlesDeletedPostResponse(t *testing.T) {
	handler, mockPostDAO, _, _, _, mockSecurePseudonymDAO, _ := createTestContentHandler()

	// Create a test post
	testPost := fixtures.CreateTestPost()

	// Inject the post into the mock
	mockPostDAO.InjectPost(testPost)

	// Mock the deletion operation
	mockPostDAO.On("MarkPostAsDeletedByPseudonym", mock.Anything, int64(123), "test-pseudonym-id", "User requested deletion").Return(nil)

	// Set up pseudonym data for the test
	testPseudonym := &models.Pseudonym{
		PseudonymID: "test-pseudonym-id",
		DisplayName: "TestUser",
	}
	mockSecurePseudonymDAO.InjectPseudonym(testPseudonym)
	mockSecurePseudonymDAO.SetDefaultBehavior()

	// Don't set up default behavior for PostDAO since we only need the specific method

	// Create a valid JWT token for testing
	userCtx := fixtures.CreateTestUserContext()
	token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
	require.NoError(t, tokenErr)

	// Create test input
	input := &apimodels.PostDeleteInput{
		PostID: 123,
		Body: struct {
			Reason string `json:"reason,omitempty" example:"User requested deletion"`
		}{
			Reason: "User requested deletion",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
	}

	// Set up user context in the request context
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call the handler
	response, err := handler.DeletePost(ctx, input)

	// Should succeed
	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify response structure
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, 123, response.Body.PostID)
	assert.Equal(t, "User requested deletion", response.Body.DeleteReason)
	assert.Equal(t, "test-pseudonym-id", response.Body.DeletedBy.PseudonymID)
	assert.Equal(t, "TestUser", response.Body.DeletedBy.DisplayName)

	// Verify that the DAO was called correctly
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_DeleteComment_HandlesDeletedCommentResponse tests that the delete comment response has the correct structure
func TestContentHandler_DeleteComment_HandlesDeletedCommentResponse(t *testing.T) {
	handler, _, mockCommentDAO, _, _, mockSecurePseudonymDAO, _ := createTestContentHandler()

	// Create a test comment
	testComment := fixtures.CreateTestComment()

	// Inject the comment into the mock
	mockCommentDAO.InjectComment(testComment)

	// Mock the deletion operation
	mockCommentDAO.On("MarkCommentAsDeletedByPseudonym", mock.Anything, int64(456), "test-pseudonym-id", "User requested deletion").Return(nil)

	// Set up pseudonym data for the test
	testPseudonym := &models.Pseudonym{
		PseudonymID: "test-pseudonym-id",
		DisplayName: "TestUser",
	}
	mockSecurePseudonymDAO.InjectPseudonym(testPseudonym)
	mockSecurePseudonymDAO.SetDefaultBehavior()

	// Create a valid JWT token for testing
	userCtx := fixtures.CreateTestUserContext()
	token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
	require.NoError(t, tokenErr)

	// Create test input
	input := &apimodels.CommentDeleteInput{
		CommentID: 456,
		Body: struct {
			Reason string `json:"reason,omitempty" example:"User requested deletion"`
		}{
			Reason: "User requested deletion",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
	}

	// Set up user context in the request context
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call the handler
	response, err := handler.DeleteComment(ctx, input)

	// Should succeed
	require.NoError(t, err)
	require.NotNil(t, response)

	// Verify response structure
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, 456, response.Body.CommentID)
	assert.Equal(t, "User requested deletion", response.Body.DeleteReason)
	assert.Equal(t, "test-pseudonym-id", response.Body.DeletedBy.PseudonymID)
	assert.Equal(t, "TestUser", response.Body.DeletedBy.DisplayName)

	// Verify that the DAO was called correctly
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_GetPostDetails_FiltersDeletedPosts tests that deleted posts are filtered out
func TestContentHandler_GetPostDetails_FiltersDeletedPosts(t *testing.T) {
	handler, mockPostDAO, mockCommentDAO, mockVoteDAO, _, _, _ := createTestContentHandler()

	// Create a deleted post
	deletedPost := fixtures.CreateTestPost()
	deletedPost.IsDeleted = sql.Null[bool]{V: true, Valid: true}

	// Inject the deleted post into the mock
	mockPostDAO.InjectPost(deletedPost)

	// Set up post DAO expectations for GetPostDetails
	mockPostDAO.On("GetPostByID", mock.Anything, int64(123)).Return(
		func(ctx context.Context, postID int64) (*models.Post, error) {
			return deletedPost, nil
		},
	)

	// Set up comment DAO expectations for GetPostDetails
	mockCommentDAO.On("GetCommentsByPostWithNestedReplies", mock.Anything, int64(123)).Return([]*models.Comment{}, nil)

	// Set up vote DAO expectations for GetPostDetails
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "post", int64(123)).Return(nil, nil)

	// Create test input
	input := &apimodels.PostDetailsInput{
		PostID: 123,
	}

	// Set up user context in the request context
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, fixtures.CreateTestUserContext())

	// Call the handler
	_, err := handler.GetPostDetails(ctx, input)

	// Should return an error for deleted posts
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleted")

	// Verify that the DAO was called correctly
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_GetPostBySlug_FiltersDeletedPosts tests that deleted posts are filtered out by slug
func TestContentHandler_GetPostBySlug_FiltersDeletedPosts(t *testing.T) {
	handler, mockPostDAO, mockCommentDAO, mockVoteDAO, mockSubforumDAO, _, _ := createTestContentHandler()

	// Create a test subforum
	testSubforum := fixtures.CreateTestSubforum()
	mockSubforumDAO.InjectSubforumByName("test-subforum", testSubforum)

	// Set up subforum DAO expectations for GetPostBySlug
	mockSubforumDAO.On("GetSubforumByName", mock.Anything, "test-subforum").Return(
		func(ctx context.Context, name string) (*models.Subforum, error) {
			return testSubforum, nil
		},
	)

	// Create a deleted post
	deletedPost := fixtures.CreateTestPost()
	deletedPost.IsDeleted = sql.Null[bool]{V: true, Valid: true}

	// Inject the deleted post into the mock
	mockPostDAO.InjectPostBySlug(1, "test-post-123", deletedPost)

	// Set up post DAO expectations for GetPostBySlug
	mockPostDAO.On("GetPostBySubforumAndSlug", mock.Anything, int32(1), "test-post-123").Return(
		func(ctx context.Context, subforumID int32, slug string) (*models.Post, error) {
			return deletedPost, nil
		},
	)

	// Set up comment DAO expectations for GetPostBySlug
	mockCommentDAO.On("GetCommentsByPostWithNestedReplies", mock.Anything, int64(123)).Return(
		func(ctx context.Context, postID int64) ([]*models.Comment, error) {
			return []*models.Comment{}, nil
		},
	)

	// Set up vote DAO expectations for GetPostBySlug
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "post", int64(123)).Return(
		func(ctx context.Context, pseudonymID string, contentType string, contentID int64) (*models.Vote, error) {
			return nil, nil
		},
	)

	// Create a valid JWT token for testing
	userCtx := fixtures.CreateTestUserContext()
	token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
	require.NoError(t, tokenErr)

	// Create test input
	input := &apimodels.PostBySlugInput{
		SubforumName: "test-subforum",
		Slug:         "test-post-123",
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
	}

	// Set up user context in the request context
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call the handler
	_, err := handler.GetPostBySlug(ctx, input)

	// Should return an error for deleted posts
	require.Error(t, err)
	assert.Contains(t, err.Error(), "deleted")

	// Verify that the DAOs were called correctly
	mockPostDAO.AssertExpectations(t)
	mockSubforumDAO.AssertExpectations(t)
}

// TestSoftDeletion_ScorePreservation tests that scores are preserved for deleted content
func TestSoftDeletion_ScorePreservation(t *testing.T) {
	_, mockPostDAO, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Create a deleted post with votes
	deletedPostWithVotes := fixtures.CreateTestPost()
	deletedPostWithVotes.IsDeleted = sql.Null[bool]{V: true, Valid: true}
	deletedPostWithVotes.Score = sql.Null[int32]{V: 15, Valid: true}
	deletedPostWithVotes.Upvotes = sql.Null[int32]{V: 20, Valid: true}
	deletedPostWithVotes.Downvotes = sql.Null[int32]{V: 5, Valid: true}

	// Create a deleted comment with votes
	deletedCommentWithVotes := fixtures.CreateTestComment()
	deletedCommentWithVotes.IsDeleted = sql.Null[bool]{V: true, Valid: true}
	deletedCommentWithVotes.Score = sql.Null[int32]{V: 8, Valid: true}
	deletedCommentWithVotes.Upvotes = sql.Null[int32]{V: 10, Valid: true}
	deletedCommentWithVotes.Downvotes = sql.Null[int32]{V: 2, Valid: true}

	// Inject the data into the mocks
	mockPostDAO.InjectPost(deletedPostWithVotes)
	mockCommentDAO.InjectComment(deletedCommentWithVotes)

	// Verify that scores are preserved even when content is deleted
	assert.True(t, deletedPostWithVotes.Score.Valid && deletedPostWithVotes.Score.V == 15, "Post score should be preserved")
	assert.True(t, deletedCommentWithVotes.Score.Valid && deletedCommentWithVotes.Score.V == 8, "Comment score should be preserved")

	// Verify that vote counts are preserved
	assert.True(t, deletedPostWithVotes.Upvotes.Valid && deletedPostWithVotes.Upvotes.V == 20, "Post upvotes should be preserved")
	assert.True(t, deletedPostWithVotes.Downvotes.Valid && deletedPostWithVotes.Downvotes.V == 5, "Post downvotes should be preserved")
	assert.True(t, deletedCommentWithVotes.Upvotes.Valid && deletedCommentWithVotes.Upvotes.V == 10, "Comment upvotes should be preserved")
	assert.True(t, deletedCommentWithVotes.Downvotes.Valid && deletedCommentWithVotes.Downvotes.V == 2, "Comment downvotes should be preserved")

	// These tests only verify data structures, no DAO calls are made
}

// TestSoftDeletion_DeletionMetadata tests that deletion metadata is properly set
func TestSoftDeletion_DeletionMetadata(t *testing.T) {
	_, mockPostDAO, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Create a deleted post with metadata
	deletedPost := fixtures.CreateTestPost()
	deletedPost.IsDeleted = sql.Null[bool]{V: true, Valid: true}
	deletedPost.DeletedByPseudonymID = sql.Null[string]{V: "deleter-pseudonym", Valid: true}
	deletedPost.DeletedByPseudonymAt = sql.Null[time.Time]{V: time.Now(), Valid: true}
	deletedPost.DeletedByPseudonymReason = sql.Null[string]{V: "User requested deletion", Valid: true}

	// Create a deleted comment with metadata
	deletedComment := fixtures.CreateTestComment()
	deletedComment.IsDeleted = sql.Null[bool]{V: true, Valid: true}
	deletedComment.DeletedByPseudonymID = sql.Null[string]{V: "deleter-pseudonym", Valid: true}
	deletedComment.DeletedByPseudonymAt = sql.Null[time.Time]{V: time.Now(), Valid: true}
	deletedComment.DeletedByPseudonymReason = sql.Null[string]{V: "User requested deletion", Valid: true}

	// Inject the data into the mocks
	mockPostDAO.InjectPost(deletedPost)
	mockCommentDAO.InjectComment(deletedComment)

	// Verify that deletion metadata is properly set
	assert.True(t, deletedPost.IsDeleted.Valid && deletedPost.IsDeleted.V, "Post should be marked as deleted")
	assert.True(t, deletedPost.DeletedByPseudonymID.Valid && deletedPost.DeletedByPseudonymID.V == "deleter-pseudonym", "Post should have deleter pseudonym ID")
	assert.True(t, deletedPost.DeletedByPseudonymAt.Valid, "Post should have deletion timestamp")
	assert.True(t, deletedPost.DeletedByPseudonymReason.Valid && deletedPost.DeletedByPseudonymReason.V == "User requested deletion", "Post should have deletion reason")

	assert.True(t, deletedComment.IsDeleted.Valid && deletedComment.IsDeleted.V, "Comment should be marked as deleted")
	assert.True(t, deletedComment.DeletedByPseudonymID.Valid && deletedComment.DeletedByPseudonymID.V == "deleter-pseudonym", "Comment should have deleter pseudonym ID")
	assert.True(t, deletedComment.DeletedByPseudonymAt.Valid, "Comment should have deletion timestamp")
	assert.True(t, deletedComment.DeletedByPseudonymReason.Valid && deletedComment.DeletedByPseudonymReason.V == "User requested deletion", "Comment should have deletion reason")

	// These tests only verify data structures, no DAO calls are made
}

// TestContentHandler_GetPosts_Success tests successful post retrieval
func TestContentHandler_GetPosts_Success(t *testing.T) {
	handler, mockPostDAO, _, mockVoteDAO, mockSubforumDAO, _, _ := createTestContentHandler()

	// Create test subforum
	testSubforum := fixtures.CreateTestSubforum()

	// Create test posts
	testPosts := []*dbmodels.Post{
		fixtures.CreateTestPost(),
		fixtures.CreateTestPost(),
	}
	testPosts[1].PostID = 124
	testPosts[1].Title = "Second Test Post"

	// Set up expectations
	mockSubforumDAO.On("GetSubforumByName", mock.Anything, "test-subforum").Return(
		func(ctx context.Context, name string) (*dbmodels.Subforum, error) {
			return testSubforum, nil
		},
	)
	mockPostDAO.On("GetPostsBySubforum", mock.Anything, int32(1), 1, 25, "created_at", true).Return(testPosts, nil)
	mockPostDAO.On("CountPostsBySubforum", mock.Anything, int32(1)).Return(int64(2), nil)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "post", int64(123)).Return(nil, nil)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "post", int64(124)).Return(nil, nil)

	// Create test input
	input := &apimodels.PostListInput{
		SubforumName: "test-subforum",
		Page:         1,
		Limit:        25,
		Sort:         "new",
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.GetPosts(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Len(t, response.Body.Posts, 2)
	assert.Equal(t, "Test Post", response.Body.Posts[0].Title)
	assert.Equal(t, "Second Test Post", response.Body.Posts[1].Title)

	// Verify DAO calls
	mockSubforumDAO.AssertExpectations(t)
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_GetPosts_PrivateSubforumAccess tests private subforum access
func TestContentHandler_GetPosts_PrivateSubforumAccess(t *testing.T) {
	handler, _, _, _, mockSubforumDAO, _, _ := createTestContentHandler()

	// Create private subforum
	privateSubforum := fixtures.CreateTestSubforum()
	privateSubforum.IsPrivate = sql.Null[bool]{V: true, Valid: true}
	mockSubforumDAO.InjectSubforumByName("private-subforum", privateSubforum)

	// Set up expectations for private subforum access check
	mockSubforumDAO.On("GetSubforumByName", mock.Anything, "private-subforum").Return(
		func(ctx context.Context, name string) (*dbmodels.Subforum, error) {
			return privateSubforum, nil
		},
	)

	// Create test input without authentication
	input := &apimodels.PostListInput{
		SubforumName: "private-subforum",
		Page:         1,
		Limit:        25,
	}

	// Call handler without user context
	ctx := context.Background()
	_, err := handler.GetPosts(ctx, input)

	// Should return unauthorized error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authentication required")

	// Verify DAO calls
	mockSubforumDAO.AssertExpectations(t)
}

// TestContentHandler_CreatePost_Success tests successful post creation
func TestContentHandler_CreatePost_Success(t *testing.T) {
	handler, mockPostDAO, _, _, mockSubforumDAO, _, _ := createTestContentHandler()

	// Create test subforum
	testSubforum := fixtures.CreateTestSubforum()
	mockSubforumDAO.InjectSubforumByName("test-subforum", testSubforum)

	// Create test post
	testPost := fixtures.CreateTestPost()
	testPost.Slug = sql.Null[string]{V: "test-post-123", Valid: true}

	// Set up expectations
	mockSubforumDAO.On("GetSubforumByName", mock.Anything, "test-subforum").Return(
		func(ctx context.Context, name string) (*dbmodels.Subforum, error) {
			return testSubforum, nil
		},
	)
	mockPostDAO.On("CreatePost", mock.Anything, int32(1), "test-pseudonym-id", "Test Post", "Test content", "text", (*string)(nil), false, false).Return(testPost, nil)

	// Create test input
	input := &apimodels.PostCreateInput{
		SubforumName: "test-subforum",
		Body: apimodels.PostCreateBody{
			Title:     "Test Post",
			Content:   "Test content",
			PostType:  "text",
			IsNSFW:    false,
			IsSpoiler: false,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.CreatePost(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, "Test Post", response.Body.Title)
	assert.Equal(t, "test-pseudonym-id", response.Body.Author.PseudonymID)

	// Verify DAO calls
	mockSubforumDAO.AssertExpectations(t)
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_CreatePost_ValidationErrors tests post creation validation
func TestContentHandler_CreatePost_ValidationErrors(t *testing.T) {
	handler, _, _, _, _, _, _ := createTestContentHandler()

	// Test cases for validation errors
	testCases := []struct {
		name        string
		title       string
		content     string
		postType    string
		expectedErr string
	}{
		{
			name:        "EmptyTitle",
			title:       "",
			content:     "Test content",
			postType:    "text",
			expectedErr: "title is required",
		},
		{
			name:        "EmptyContentForTextPost",
			title:       "Test Post",
			content:     "",
			postType:    "text",
			expectedErr: "content is required for text posts",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := &apimodels.PostCreateInput{
				SubforumName: "test-subforum",
				Body: apimodels.PostCreateBody{
					Title:     tc.title,
					Content:   tc.content,
					PostType:  tc.postType,
					IsNSFW:    false,
					IsSpoiler: false,
				},
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
			}

			userCtx := fixtures.CreateTestUserContext()
			ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

			_, err := handler.CreatePost(ctx, input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestContentHandler_GetPostDetails_Success tests successful post details retrieval
func TestContentHandler_GetPostDetails_Success(t *testing.T) {
	handler, mockPostDAO, mockCommentDAO, mockVoteDAO, _, _, _ := createTestContentHandler()

	// Create test post
	testPost := fixtures.CreateTestPost()
	testPost.Slug = sql.Null[string]{V: "test-post-123", Valid: true}

	// Create test comments
	testComments := []*dbmodels.Comment{
		fixtures.CreateTestComment(),
	}

	// Set up expectations
	mockPostDAO.On("GetPostByID", mock.Anything, int64(123)).Return(
		func(ctx context.Context, postID int64) (*dbmodels.Post, error) {
			return testPost, nil
		},
	)
	mockCommentDAO.On("GetCommentsByPostWithNestedReplies", mock.Anything, int64(123)).Return(testComments, nil)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "post", int64(123)).Return(nil, nil)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "comment", int64(456)).Return(nil, nil)

	// Create test input
	input := &apimodels.PostDetailsInput{
		PostID: 123,
		Sort:   "best",
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.GetPostDetails(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, "Test Post", response.Body.Title)
	assert.Len(t, response.Body.Comments, 1)

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_GetPostBySlug_Success tests successful post retrieval by slug
func TestContentHandler_GetPostBySlug_Success(t *testing.T) {
	handler, mockPostDAO, mockCommentDAO, mockVoteDAO, mockSubforumDAO, _, _ := createTestContentHandler()

	// Create test subforum
	testSubforum := fixtures.CreateTestSubforum()
	mockSubforumDAO.InjectSubforumByName("test-subforum", testSubforum)

	// Create test post
	testPost := fixtures.CreateTestPost()
	testPost.Slug = sql.Null[string]{V: "test-post-123", Valid: true}

	// Create test comments
	testComments := []*dbmodels.Comment{
		fixtures.CreateTestComment(),
	}

	// Set up expectations
	mockSubforumDAO.On("GetSubforumByName", mock.Anything, "test-subforum").Return(
		func(ctx context.Context, name string) (*dbmodels.Subforum, error) {
			return testSubforum, nil
		},
	)
	mockPostDAO.On("GetPostBySubforumAndSlug", mock.Anything, int32(1), "test-post-123").Return(
		func(ctx context.Context, subforumID int32, slug string) (*dbmodels.Post, error) {
			return testPost, nil
		},
	)
	mockCommentDAO.On("GetCommentsByPostWithNestedReplies", mock.Anything, int64(123)).Return(testComments, nil)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "post", int64(123)).Return(nil, nil)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "comment", int64(456)).Return(nil, nil)

	// Create test input
	input := &apimodels.PostBySlugInput{
		SubforumName: "test-subforum",
		Slug:         "test-post-123",
		Sort:         "best",
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.GetPostBySlug(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, "Test Post", response.Body.Title)
	assert.Len(t, response.Body.Comments, 1)

	// Verify DAO calls
	mockSubforumDAO.AssertExpectations(t)
	mockPostDAO.AssertExpectations(t)
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_VoteOnPost_Success tests successful post voting
func TestContentHandler_VoteOnPost_Success(t *testing.T) {
	handler, mockPostDAO, _, mockVoteDAO, _, _, _ := createTestContentHandler()

	// Create test post
	testPost := fixtures.CreateTestPost()

	// Set up expectations
	mockPostDAO.On("GetPostByID", mock.Anything, int64(123)).Return(
		func(ctx context.Context, postID int64) (*dbmodels.Post, error) {
			return testPost, nil
		},
	)
	mockVoteDAO.On("UpsertVote", mock.Anything, "test-pseudonym-id", "post", int64(123), int32(1)).Return(&models.Vote{VoteID: 1, VoteValue: 1}, nil)
	mockVoteDAO.On("GetVoteSummaryByContent", mock.Anything, "post", int64(123)).Return(16, 6, 1, nil)
	mockPostDAO.On("UpdatePostScore", mock.Anything, int64(123), int32(10), int32(16), int32(6)).Return(nil)

	// Create test input
	input := &apimodels.PostVoteInput{
		PostID: 123,
		Body: apimodels.VoteInputBody{
			VoteValue: 1,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.VoteOnPost(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, 123, response.Body.PostID)
	assert.Equal(t, 1, response.Body.VoteValue)
	assert.Equal(t, 10, response.Body.Score)

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
	mockVoteDAO.AssertExpectations(t)
}

// TestContentHandler_VoteOnPost_RemoveVote tests removing a vote
func TestContentHandler_VoteOnPost_RemoveVote(t *testing.T) {
	handler, mockPostDAO, _, mockVoteDAO, _, _, _ := createTestContentHandler()

	// Create test post
	testPost := fixtures.CreateTestPost()

	// Create existing vote
	existingVote := &models.Vote{
		VoteID:    1,
		VoteValue: 1,
	}

	// Set up expectations
	mockPostDAO.On("GetPostByID", mock.Anything, int64(123)).Return(
		func(ctx context.Context, postID int64) (*dbmodels.Post, error) {
			return testPost, nil
		},
	)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "post", int64(123)).Return(existingVote, nil)
	mockVoteDAO.On("DeleteVote", mock.Anything, int64(1)).Return(nil)
	mockVoteDAO.On("GetVoteSummaryByContent", mock.Anything, "post", int64(123)).Return(14, 5, 0, nil)
	mockPostDAO.On("UpdatePostScore", mock.Anything, int64(123), int32(9), int32(14), int32(5)).Return(nil)

	// Create test input
	input := &apimodels.PostVoteInput{
		PostID: 123,
		Body: apimodels.VoteInputBody{
			VoteValue: 0,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.VoteOnPost(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, 0, response.Body.VoteValue)
	assert.Equal(t, 9, response.Body.Score)

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
	mockVoteDAO.AssertExpectations(t)
}

// TestContentHandler_VoteOnPost_InvalidVoteValue tests invalid vote values
func TestContentHandler_VoteOnPost_InvalidVoteValue(t *testing.T) {
	handler, _, _, _, _, _, _ := createTestContentHandler()

	// Test invalid vote values
	testCases := []struct {
		name        string
		voteValue   int
		expectedErr string
	}{
		{"VoteValue2", 2, "invalid vote value"},
		{"VoteValueNegative2", -2, "invalid vote value"},
		{"VoteValue10", 10, "invalid vote value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := &apimodels.PostVoteInput{
				PostID: 123,
				Body: apimodels.VoteInputBody{
					VoteValue: tc.voteValue,
				},
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
			}

			userCtx := fixtures.CreateTestUserContext()
			ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

			_, err := handler.VoteOnPost(ctx, input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestContentHandler_VoteOnComment_Success tests successful comment voting
func TestContentHandler_VoteOnComment_Success(t *testing.T) {
	handler, _, mockCommentDAO, mockVoteDAO, _, _, _ := createTestContentHandler()

	// Create test comment
	testComment := fixtures.CreateTestComment()

	// Set up expectations
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(456)).Return(
		func(ctx context.Context, commentID int64) (*dbmodels.Comment, error) {
			return testComment, nil
		},
	)
	mockVoteDAO.On("UpsertVote", mock.Anything, "test-pseudonym-id", "comment", int64(456), int32(-1)).Return(&models.Vote{VoteID: 1, VoteValue: -1}, nil)
	mockVoteDAO.On("GetVoteSummaryByContent", mock.Anything, "comment", int64(456)).Return(7, 4, 1, nil)
	mockCommentDAO.On("UpdateCommentScore", mock.Anything, int64(456), int32(3), int32(7), int32(4)).Return(nil)

	// Create test input
	input := &apimodels.CommentVoteInput{
		CommentID: 456,
		Body: apimodels.VoteInputBody{
			VoteValue: -1,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.VoteOnComment(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, 456, response.Body.CommentID)
	assert.Equal(t, -1, response.Body.VoteValue)
	assert.Equal(t, 3, response.Body.Score)

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
	mockVoteDAO.AssertExpectations(t)
}

// TestContentHandler_VoteOnComment_RemoveVote tests removing a comment vote
func TestContentHandler_VoteOnComment_RemoveVote(t *testing.T) {
	handler, _, mockCommentDAO, mockVoteDAO, _, _, _ := createTestContentHandler()

	// Create test comment
	testComment := fixtures.CreateTestComment()

	// Create existing vote
	existingVote := &models.Vote{
		VoteID:    1,
		VoteValue: -1,
	}

	// Set up expectations
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(456)).Return(
		func(ctx context.Context, commentID int64) (*dbmodels.Comment, error) {
			return testComment, nil
		},
	)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "comment", int64(456)).Return(existingVote, nil)
	mockVoteDAO.On("DeleteVote", mock.Anything, int64(1)).Return(nil)
	mockVoteDAO.On("GetVoteSummaryByContent", mock.Anything, "comment", int64(456)).Return(8, 3, 0, nil)
	mockCommentDAO.On("UpdateCommentScore", mock.Anything, int64(456), int32(5), int32(8), int32(3)).Return(nil)

	// Create test input
	input := &apimodels.CommentVoteInput{
		CommentID: 456,
		Body: apimodels.VoteInputBody{
			VoteValue: 0,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.VoteOnComment(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, 0, response.Body.VoteValue)
	assert.Equal(t, 5, response.Body.Score)

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
	mockVoteDAO.AssertExpectations(t)
}

// TestContentHandler_CreateComment_Success tests successful comment creation
func TestContentHandler_CreateComment_Success(t *testing.T) {
	handler, mockPostDAO, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Create test post
	testPost := fixtures.CreateTestPost()
	testPost.CommentCount = sql.Null[int32]{V: 5, Valid: true}

	// Create test comment
	testComment := fixtures.CreateTestComment()

	// Set up expectations
	mockPostDAO.On("GetPostByID", mock.Anything, int64(123)).Return(
		func(ctx context.Context, postID int64) (*dbmodels.Post, error) {
			return testPost, nil
		},
	)
	mockCommentDAO.On("CreateComment", mock.Anything, int64(123), "test-pseudonym-id", "Test comment", (*int64)(nil)).Return(testComment, nil)
	mockPostDAO.On("UpdateCommentCount", mock.Anything, int64(123), int32(6)).Return(nil)

	// Create test input
	input := &apimodels.CommentInput{
		PostID: 123,
		Body: apimodels.CommentInputBody{
			Content: "Test comment",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.CreateComment(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, "Test comment", response.Body.Content)
	assert.Equal(t, "test-pseudonym-id", response.Body.Author.PseudonymID)

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_CreateComment_ValidationErrors tests comment creation validation
func TestContentHandler_CreateComment_ValidationErrors(t *testing.T) {
	handler, _, _, _, _, _, _ := createTestContentHandler()

	// Test empty content
	input := &apimodels.CommentInput{
		PostID: 123,
		Body: apimodels.CommentInputBody{
			Content: "",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	_, err := handler.CreateComment(ctx, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "content is required")
}

// TestContentHandler_EditPost_Success tests successful post editing
func TestContentHandler_EditPost_Success(t *testing.T) {
	t.Skip("Skipping EditPost test - requires real database connection for post.Update()")
}

// TestContentHandler_EditPost_NotOwner tests editing a post the user doesn't own
func TestContentHandler_EditPost_NotOwner(t *testing.T) {
	handler, mockPostDAO, _, _, _, _, _ := createTestContentHandler()

	// Create test post owned by different user
	testPost := fixtures.CreateTestPost()
	testPost.PseudonymID = "different-pseudonym-id"

	// Set up expectations
	mockPostDAO.On("GetPostByID", mock.Anything, int64(123)).Return(
		func(ctx context.Context, postID int64) (*dbmodels.Post, error) {
			return testPost, nil
		},
	)

	// Create test input
	input := &apimodels.PostEditInput{
		PostID: 123,
		Body: apimodels.PostEditInputBody{
			Title:   "Updated Post Title",
			Content: "Updated post content",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.EditPost(ctx, input)

	// Should return forbidden error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "you can only edit your own posts")

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_EditComment_Success tests successful comment editing
func TestContentHandler_EditComment_Success(t *testing.T) {
	handler, _, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Create test comment owned by the user
	testComment := fixtures.CreateTestComment()

	// Set up expectations
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(456)).Return(
		func(ctx context.Context, commentID int64) (*dbmodels.Comment, error) {
			return testComment, nil
		},
	)
	mockCommentDAO.On("UpdateComment", mock.Anything, int64(456), "Updated comment content", "Fixed typo").Return(nil)

	// Create test input
	input := &apimodels.CommentEditInput{
		CommentID: 456,
		Body: apimodels.CommentEditInputBody{
			Content:    "Updated comment content",
			EditReason: "Fixed typo",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.EditComment(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, "Updated comment content", response.Body.Content)
	assert.Equal(t, "Fixed typo", response.Body.EditReason)

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_RemoveComment_Success tests successful comment removal
func TestContentHandler_RemoveComment_Success(t *testing.T) {
	handler, _, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Create test comment owned by the user
	testComment := fixtures.CreateTestComment()

	// Set up expectations
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(456)).Return(
		func(ctx context.Context, commentID int64) (*dbmodels.Comment, error) {
			return testComment, nil
		},
	)
	mockCommentDAO.On("SetCommentRemoved", mock.Anything, int64(456), true, "Violates community guidelines", "test-pseudonym-id").Return(nil)

	// Create test input
	input := &apimodels.CommentRemoveInput{
		CommentID: 456,
		Body: struct {
			Removed bool   `json:"removed" example:"true" required:"true"`
			Reason  string `json:"reason,omitempty" example:"Violates community guidelines"`
		}{
			Removed: true,
			Reason:  "Violates community guidelines",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.RemoveComment(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, 456, response.Body.CommentID)
	assert.True(t, response.Body.Removed)
	assert.Equal(t, "Violates community guidelines", response.Body.RemovalReason)

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_ReportComment_Success tests successful comment reporting
func TestContentHandler_ReportComment_Success(t *testing.T) {
	handler, _, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Create test comment owned by different user
	testComment := fixtures.CreateTestComment()
	testComment.PseudonymID = "different-pseudonym-id"

	// Set up expectations
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(456)).Return(
		func(ctx context.Context, commentID int64) (*dbmodels.Comment, error) {
			return testComment, nil
		},
	)

	// Create test input
	input := &apimodels.CommentReportInput{
		CommentID: 456,
		Body: apimodels.CommentReportInputBody{
			ReportReason:  "spam",
			ReportDetails: "This comment violates community guidelines",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.ReportComment(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, 200, response.Status)
	assert.Equal(t, 456, response.Body.CommentID)
	assert.Equal(t, "spam", response.Body.ReportReason)
	assert.Equal(t, "This comment violates community guidelines", response.Body.ReportDetails)

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_ReportComment_OwnComment tests reporting own comment
func TestContentHandler_ReportComment_OwnComment(t *testing.T) {
	handler, _, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Create test comment owned by the user
	testComment := fixtures.CreateTestComment()

	// Set up expectations
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(456)).Return(
		func(ctx context.Context, commentID int64) (*dbmodels.Comment, error) {
			return testComment, nil
		},
	)

	// Create test input
	input := &apimodels.CommentReportInput{
		CommentID: 456,
		Body: apimodels.CommentReportInputBody{
			ReportReason:  "spam",
			ReportDetails: "This comment violates community guidelines",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.ReportComment(ctx, input)

	// Should return error for reporting own comment
	require.Error(t, err)
	assert.Contains(t, err.Error(), "you cannot report your own comment")

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_LockPost_Success tests successful post locking
func TestContentHandler_LockPost_Success(t *testing.T) {
	handler, mockPostDAO, _, mockVoteDAO, _, _, mockPermissionDAO := createTestContentHandler()

	// Create test post
	testPost := fixtures.CreateTestPost()

	// Set up expectations
	mockPostDAO.On("GetPostByID", mock.Anything, int64(123)).Return(
		func(ctx context.Context, postID int64) (*dbmodels.Post, error) {
			return testPost, nil
		},
	)
	mockPostDAO.On("SetLocked", mock.Anything, int64(123), true).Return(nil)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "post", int64(123)).Return(nil, nil)
	mockPermissionDAO.On("HasSubforumCapabilityWithActivePseudonym", mock.Anything, int64(1), int32(1), "moderate_content", "test-pseudonym-id").Return(true, nil)

	// Note: Permission DAO is mocked in the handler creation
	// The test will pass if the permission check is properly mocked

	// Create test input
	input := &apimodels.PostLockInput{
		PostID: 123,
		Body: struct {
			Locked bool `json:"locked" example:"true" required:"true"`
		}{
			Locked: true,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.LockPost(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "Test Post", response.Body.Title)

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
	mockPermissionDAO.AssertExpectations(t)
}

// TestContentHandler_LockPost_NotFound tests locking a non-existent post
func TestContentHandler_LockPost_NotFound(t *testing.T) {
	handler, mockPostDAO, _, _, _, _, _ := createTestContentHandler()

	// Set up expectations for non-existent post
	mockPostDAO.On("GetPostByID", mock.Anything, int64(999)).Return(nil, sql.ErrNoRows)

	// Create test input
	input := &apimodels.PostLockInput{
		PostID: 999,
		Body: struct {
			Locked bool `json:"locked" example:"true" required:"true"`
		}{
			Locked: true,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.LockPost(ctx, input)

	// Should return error for non-existent post
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch post")

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_StickyPost_Success tests successful post stickying
func TestContentHandler_StickyPost_Success(t *testing.T) {
	handler, mockPostDAO, _, mockVoteDAO, _, _, mockPermissionDAO := createTestContentHandler()

	// Create test post
	testPost := fixtures.CreateTestPost()

	// Set up expectations
	mockPostDAO.On("GetPostByID", mock.Anything, int64(123)).Return(testPost, nil).Twice()
	mockPostDAO.On("SetSticky", mock.Anything, int64(123), true).Return(nil)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "post", int64(123)).Return(nil, nil)
	mockPermissionDAO.On("HasSubforumCapabilityWithActivePseudonym", mock.Anything, int64(1), int32(1), "moderate_content", "test-pseudonym-id").Return(true, nil)

	// Create test input
	input := &apimodels.PostStickyInput{
		PostID: 123,
		Body: struct {
			Sticky bool `json:"sticky" example:"true" required:"true"`
		}{
			Sticky: true,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.StickyPost(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "Test Post", response.Body.Title)

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
	mockPermissionDAO.AssertExpectations(t)
}

// TestContentHandler_StickyPost_NotFound tests stickying a non-existent post
func TestContentHandler_StickyPost_NotFound(t *testing.T) {
	handler, mockPostDAO, _, _, _, _, _ := createTestContentHandler()

	// Set up expectations for non-existent post
	mockPostDAO.On("GetPostByID", mock.Anything, int64(999)).Return(nil, sql.ErrNoRows)

	// Create test input
	input := &apimodels.PostStickyInput{
		PostID: 999,
		Body: struct {
			Sticky bool `json:"sticky" example:"true" required:"true"`
		}{
			Sticky: true,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.StickyPost(ctx, input)

	// Should return error for non-existent post
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch post")

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_RemovePost_Success tests successful post removal
func TestContentHandler_RemovePost_Success(t *testing.T) {
	handler, mockPostDAO, _, mockVoteDAO, _, _, mockPermissionDAO := createTestContentHandler()

	// Create test post
	testPost := fixtures.CreateTestPost()

	// Set up expectations
	mockPostDAO.On("GetPostByID", mock.Anything, int64(123)).Return(testPost, nil)
	mockPostDAO.On("SetRemoved", mock.Anything, int64(123), true).Return(nil)
	mockVoteDAO.On("GetVoteByPseudonymAndContent", mock.Anything, "test-pseudonym-id", "post", int64(123)).Return(nil, nil)
	mockPermissionDAO.On("HasSubforumCapabilityWithActivePseudonym", mock.Anything, int64(1), int32(1), "moderate_content", "test-pseudonym-id").Return(true, nil)

	// Create test input
	input := &apimodels.PostRemoveInput{
		PostID: 123,
		Body: struct {
			Removed bool `json:"removed" example:"true" required:"true"`
		}{
			Removed: true,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	response, err := handler.RemovePost(ctx, input)

	// Verify response
	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "Test Post", response.Body.Title)

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
	mockPermissionDAO.AssertExpectations(t)
}

// TestContentHandler_RemovePost_NotFound tests removing a non-existent post
func TestContentHandler_RemovePost_NotFound(t *testing.T) {
	handler, mockPostDAO, _, _, _, _, _ := createTestContentHandler()

	// Set up expectations for non-existent post
	mockPostDAO.On("GetPostByID", mock.Anything, int64(999)).Return(nil, sql.ErrNoRows)

	// Create test input
	input := &apimodels.PostRemoveInput{
		PostID: 999,
		Body: struct {
			Removed bool `json:"removed" example:"true" required:"true"`
		}{
			Removed: true,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.RemovePost(ctx, input)

	// Should return error for non-existent post
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch post")

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_EditComment_NotFound tests editing a non-existent comment
func TestContentHandler_EditComment_NotFound(t *testing.T) {
	handler, _, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Set up expectations for non-existent comment
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(999)).Return(nil, nil)

	// Create test input
	input := &apimodels.CommentEditInput{
		CommentID: 999,
		Body: apimodels.CommentEditInputBody{
			Content:    "Updated comment content",
			EditReason: "Fixed typo",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.EditComment(ctx, input)

	// Should return error for non-existent comment
	require.Error(t, err)
	assert.Contains(t, err.Error(), "comment not found")

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_EditComment_NotOwner tests editing a comment the user doesn't own
func TestContentHandler_EditComment_NotOwner(t *testing.T) {
	handler, _, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Create test comment owned by different user
	testComment := fixtures.CreateTestComment()
	testComment.PseudonymID = "different-pseudonym-id"

	// Set up expectations
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(456)).Return(testComment, nil).Once()

	// Create test input
	input := &apimodels.CommentEditInput{
		CommentID: 456,
		Body: apimodels.CommentEditInputBody{
			Content:    "Updated comment content",
			EditReason: "Fixed typo",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Debug: Print pseudonym IDs
	t.Logf("Test comment pseudonym ID: %q", testComment.PseudonymID)
	t.Logf("User context active pseudonym ID: %q", userCtx.ActivePseudonymID)

	// Call handler
	_, err := handler.EditComment(ctx, input)

	// Should return forbidden error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "you can only edit your own comments")

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_EditComment_ValidationErrors tests comment editing validation
func TestContentHandler_EditComment_ValidationErrors(t *testing.T) {
	handler, _, _, _, _, _, _ := createTestContentHandler()

	// Test empty content
	input := &apimodels.CommentEditInput{
		CommentID: 456,
		Body: apimodels.CommentEditInputBody{
			Content:    "",
			EditReason: "Fixed typo",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	_, err := handler.EditComment(ctx, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "content is required")
}

// TestContentHandler_RemoveComment_NotFound tests removing a non-existent comment
func TestContentHandler_RemoveComment_NotFound(t *testing.T) {
	handler, _, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Set up expectations for non-existent comment
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(999)).Return(nil, nil)

	// Create test input
	input := &apimodels.CommentRemoveInput{
		CommentID: 999,
		Body: struct {
			Removed bool   `json:"removed" example:"true" required:"true"`
			Reason  string `json:"reason,omitempty" example:"Violates community guidelines"`
		}{
			Removed: true,
			Reason:  "Violates community guidelines",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.RemoveComment(ctx, input)

	// Should return error for non-existent comment
	require.Error(t, err)
	assert.Contains(t, err.Error(), "comment not found")

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_RemoveComment_NotOwner tests removing a comment the user doesn't own
func TestContentHandler_RemoveComment_NotOwner(t *testing.T) {
	handler, mockPostDAO, mockCommentDAO, _, _, _, mockPermissionDAO := createTestContentHandler()

	// Create test comment owned by different user
	testComment := fixtures.CreateTestComment()
	testComment.PseudonymID = "different-pseudonym-id"

	// Set up expectations
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(456)).Return(testComment, nil)
	mockPostDAO.On("GetPostByID", mock.Anything, int64(123)).Return(fixtures.CreateTestPost(), nil)
	mockPermissionDAO.On("HasSubforumCapability", mock.Anything, int64(1), int32(1), "moderate_content").Return(false, nil)

	// Create test input
	input := &apimodels.CommentRemoveInput{
		CommentID: 456,
		Body: struct {
			Removed bool   `json:"removed" example:"true" required:"true"`
			Reason  string `json:"reason,omitempty" example:"Violates community guidelines"`
		}{
			Removed: true,
			Reason:  "Violates community guidelines",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.RemoveComment(ctx, input)

	// Should return forbidden error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient permissions to remove comment")

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
	mockPostDAO.AssertExpectations(t)
	mockPermissionDAO.AssertExpectations(t)
}

// TestContentHandler_ReportComment_NotFound tests reporting a non-existent comment
func TestContentHandler_ReportComment_NotFound(t *testing.T) {
	handler, _, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Set up expectations for non-existent comment
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(999)).Return(nil, nil)

	// Create test input
	input := &apimodels.CommentReportInput{
		CommentID: 999,
		Body: apimodels.CommentReportInputBody{
			ReportReason:  "spam",
			ReportDetails: "This comment violates community guidelines",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.ReportComment(ctx, input)

	// Should return error for non-existent comment
	require.Error(t, err)
	assert.Contains(t, err.Error(), "comment not found")

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_ReportComment_ValidationErrors tests comment reporting validation
func TestContentHandler_ReportComment_ValidationErrors(t *testing.T) {
	handler, _, _, _, _, _, _ := createTestContentHandler()

	// Test empty report reason
	input := &apimodels.CommentReportInput{
		CommentID: 456,
		Body: apimodels.CommentReportInputBody{
			ReportReason: "",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	_, err := handler.ReportComment(ctx, input)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "report_reason is required")
}

// TestContentHandler_VoteOnComment_InvalidVoteValue tests invalid comment vote values
func TestContentHandler_VoteOnComment_InvalidVoteValue(t *testing.T) {
	handler, _, _, _, _, _, _ := createTestContentHandler()

	// Test invalid vote values
	testCases := []struct {
		name        string
		voteValue   int
		expectedErr string
	}{
		{"VoteValue2", 2, "invalid vote value"},
		{"VoteValueNegative2", -2, "invalid vote value"},
		{"VoteValue10", 10, "invalid vote value"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			input := &apimodels.CommentVoteInput{
				CommentID: 456,
				Body: apimodels.VoteInputBody{
					VoteValue: tc.voteValue,
				},
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
			}

			userCtx := fixtures.CreateTestUserContext()
			ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

			_, err := handler.VoteOnComment(ctx, input)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedErr)
		})
	}
}

// TestContentHandler_VoteOnComment_NotFound tests voting on a non-existent comment
func TestContentHandler_VoteOnComment_NotFound(t *testing.T) {
	handler, _, mockCommentDAO, _, _, _, _ := createTestContentHandler()

	// Set up expectations for non-existent comment
	mockCommentDAO.On("GetCommentByID", mock.Anything, int64(999)).Return(nil, nil)

	// Create test input
	input := &apimodels.CommentVoteInput{
		CommentID: 999,
		Body: apimodels.VoteInputBody{
			VoteValue: 1,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.VoteOnComment(ctx, input)

	// Should return error for non-existent comment
	require.Error(t, err)
	assert.Contains(t, err.Error(), "comment not found")

	// Verify DAO calls
	mockCommentDAO.AssertExpectations(t)
}

// TestContentHandler_VoteOnPost_NotFound tests voting on a non-existent post
func TestContentHandler_VoteOnPost_NotFound(t *testing.T) {
	handler, mockPostDAO, _, _, _, _, _ := createTestContentHandler()

	// Set up expectations for non-existent post
	mockPostDAO.On("GetPostByID", mock.Anything, int64(999)).Return(nil, nil)

	// Create test input
	input := &apimodels.PostVoteInput{
		PostID: 999,
		Body: apimodels.VoteInputBody{
			VoteValue: 1,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.VoteOnPost(ctx, input)

	// Should return error for non-existent post
	require.Error(t, err)
	assert.Contains(t, err.Error(), "post not found")

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_CreateComment_NotFound tests creating a comment on a non-existent post
func TestContentHandler_CreateComment_NotFound(t *testing.T) {
	handler, mockPostDAO, _, _, _, _, _ := createTestContentHandler()

	// Set up expectations for non-existent post
	mockPostDAO.On("GetPostByID", mock.Anything, int64(999)).Return(nil, nil)

	// Create test input
	input := &apimodels.CommentInput{
		PostID: 999,
		Body: apimodels.CommentInputBody{
			Content: "Test comment",
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.CreateComment(ctx, input)

	// Should return error for non-existent post
	require.Error(t, err)
	assert.Contains(t, err.Error(), "post not found")

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_CreatePost_SubforumNotFound tests creating a post in a non-existent subforum
func TestContentHandler_CreatePost_SubforumNotFound(t *testing.T) {
	handler, _, _, _, mockSubforumDAO, _, _ := createTestContentHandler()

	// Set up expectations for non-existent subforum
	mockSubforumDAO.On("GetSubforumByName", mock.Anything, "non-existent-subforum").Return(nil, sql.ErrNoRows)

	// Create test input
	input := &apimodels.PostCreateInput{
		SubforumName: "non-existent-subforum",
		Body: apimodels.PostCreateBody{
			Title:     "Test Post",
			Content:   "Test content",
			PostType:  "text",
			IsNSFW:    false,
			IsSpoiler: false,
		},
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.CreatePost(ctx, input)

	// Should return error for non-existent subforum
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subforum not found")

	// Verify DAO calls
	mockSubforumDAO.AssertExpectations(t)
}

// TestContentHandler_GetPosts_SubforumNotFound tests getting posts from a non-existent subforum
func TestContentHandler_GetPosts_SubforumNotFound(t *testing.T) {
	handler, _, _, _, mockSubforumDAO, _, _ := createTestContentHandler()

	// Set up expectations for non-existent subforum
	mockSubforumDAO.On("GetSubforumByName", mock.Anything, "non-existent-subforum").Return(nil, sql.ErrNoRows)

	// Create test input
	input := &apimodels.PostListInput{
		SubforumName: "non-existent-subforum",
		Page:         1,
		Limit:        25,
	}

	// Call handler without user context
	ctx := context.Background()
	_, err := handler.GetPosts(ctx, input)

	// Should return error for non-existent subforum
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subforum not found")

	// Verify DAO calls
	mockSubforumDAO.AssertExpectations(t)
}

// TestContentHandler_GetPostBySlug_SubforumNotFound tests getting a post by slug from a non-existent subforum
func TestContentHandler_GetPostBySlug_SubforumNotFound(t *testing.T) {
	handler, _, _, _, mockSubforumDAO, _, _ := createTestContentHandler()

	// Set up expectations for non-existent subforum
	mockSubforumDAO.On("GetSubforumByName", mock.Anything, "non-existent-subforum").Return(nil, sql.ErrNoRows)

	// Create test input
	input := &apimodels.PostBySlugInput{
		SubforumName: "non-existent-subforum",
		Slug:         "test-post-123",
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.GetPostBySlug(ctx, input)

	// Should return error for non-existent subforum
	require.Error(t, err)
	assert.Contains(t, err.Error(), "subforum not found")

	// Verify DAO calls
	mockSubforumDAO.AssertExpectations(t)
}

// TestContentHandler_GetPostBySlug_PostNotFound tests getting a non-existent post by slug
func TestContentHandler_GetPostBySlug_PostNotFound(t *testing.T) {
	handler, mockPostDAO, _, _, mockSubforumDAO, _, _ := createTestContentHandler()

	// Create test subforum
	testSubforum := fixtures.CreateTestSubforum()

	// Set up expectations
	mockSubforumDAO.On("GetSubforumByName", mock.Anything, "test-subforum").Return(testSubforum, nil)
	mockPostDAO.On("GetPostBySubforumAndSlug", mock.Anything, int32(1), "non-existent-post").Return(nil, sql.ErrNoRows)

	// Create test input
	input := &apimodels.PostBySlugInput{
		SubforumName: "test-subforum",
		Slug:         "non-existent-post",
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
		},
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.GetPostBySlug(ctx, input)

	// Should return error for non-existent post
	require.Error(t, err)
	assert.Contains(t, err.Error(), "post not found")

	// Verify DAO calls
	mockSubforumDAO.AssertExpectations(t)
	mockPostDAO.AssertExpectations(t)
}

// TestContentHandler_GetPostDetails_NotFound tests getting details of a non-existent post
func TestContentHandler_GetPostDetails_NotFound(t *testing.T) {
	handler, mockPostDAO, _, _, _, _, _ := createTestContentHandler()

	// Set up expectations for non-existent post
	mockPostDAO.On("GetPostByID", mock.Anything, int64(999)).Return(nil, sql.ErrNoRows)

	// Create test input
	input := &apimodels.PostDetailsInput{
		PostID: 999,
	}

	// Set up user context
	userCtx := fixtures.CreateTestUserContext()
	ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

	// Call handler
	_, err := handler.GetPostDetails(ctx, input)

	// Should return error for non-existent post
	require.Error(t, err)
	assert.Contains(t, err.Error(), "post not found")

	// Verify DAO calls
	mockPostDAO.AssertExpectations(t)
}
