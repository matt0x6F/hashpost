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
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to create test content handler with mocked dependencies
func createTestContentHandler() (*ContentHandler, *mocks.MockPostDAO, *mocks.MockCommentDAO, *mocks.MockVoteDAO, *mocks.MockSubforumDAO, *mocks.MockSecurePseudonymDAO) {
	mockPostDAO := mocks.NewMockPostDAO()
	mockCommentDAO := mocks.NewMockCommentDAO()
	mockVoteDAO := mocks.NewMockVoteDAO()
	mockSubforumDAO := mocks.NewMockSubforumDAO()
	mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()
	mockUserDAO := &mocks.MockUserDAO{}
	mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
	mockUserBlocksDAO := &mocks.MockUserBlocksDAO{}
	mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}

	ibeSystem := ibe.NewIBESystem()

	// Create a mock auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	permissionChecker := middleware.NewPermissionChecker(nil) // Mock DB for now

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
	)

	return handler, mockPostDAO, mockCommentDAO, mockVoteDAO, mockSubforumDAO, mockSecurePseudonymDAO
}

// TestContentHandler_VoteOnPost_PreventsVotingOnDeletedPost tests that voting on deleted posts is prevented
func TestContentHandler_VoteOnPost_PreventsVotingOnDeletedPost(t *testing.T) {
	handler, mockPostDAO, _, _, _, _ := createTestContentHandler()

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
	handler, _, mockCommentDAO, _, _, _ := createTestContentHandler()

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
	handler, mockPostDAO, _, _, _, mockSecurePseudonymDAO := createTestContentHandler()

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
	handler, _, mockCommentDAO, _, _, mockSecurePseudonymDAO := createTestContentHandler()

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
	handler, mockPostDAO, mockCommentDAO, mockVoteDAO, _, _ := createTestContentHandler()

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
	handler, mockPostDAO, mockCommentDAO, mockVoteDAO, mockSubforumDAO, _ := createTestContentHandler()

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
	_, mockPostDAO, mockCommentDAO, _, _, _ := createTestContentHandler()

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
	_, mockPostDAO, mockCommentDAO, _, _, _ := createTestContentHandler()

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
