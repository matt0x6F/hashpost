package handlers_test

import (
	"database/sql"
	"testing"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// TestContentHandler_VoteOnPost_Success tests successful post voting
func TestContentHandler_VoteOnPost_Success(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("VoteOnPostSuccess", func(t *testing.T) {
		handler, mockPostDAO, _, _, mockPseudonymDAO, mockPermissionDAO, mockVoteDAO := NewContentHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		postID := int64(1)

		// Create context with user
		ctx := createTestContentContext(t, userID, activePseudonymID, displayName)

		// Mock permission check
		mockPermissionDAO.EXPECT().HasUnifiedCapability(gomock.Any(), userID, activePseudonymID, constants.CapabilityVote, (*int32)(nil)).Return(true, nil).Times(1)

		// Mock post
		mockPost := &dbmodels.Post{
			PostID:      postID,
			Title:       "Test Post",
			SubforumID:  1,
			PseudonymID: "other-user",
		}
		mockPostDAO.EXPECT().GetPostByID(gomock.Any(), postID).Return(mockPost, nil).Times(1)

		// Mock vote upsert
		mockVoteDAO.EXPECT().UpsertVote(gomock.Any(), activePseudonymID, "post", postID, int32(1)).Return(&dbmodels.Vote{
			VoteID:      1,
			PseudonymID: activePseudonymID,
			ContentType: "post",
			ContentID:   postID,
			VoteValue:   1,
		}, nil).Times(1)

		// Mock vote summary retrieval
		mockVoteDAO.EXPECT().GetVoteSummaryByContent(gomock.Any(), "post", postID).Return(1, 0, 1, nil).Times(1)

		// Mock pseudonym last active update
		mockPseudonymDAO.EXPECT().UpdateLastActive(gomock.Any(), activePseudonymID).Return(nil).Times(1)

		// Mock post score update
		mockPostDAO.EXPECT().UpdatePostScore(gomock.Any(), postID, int32(1), int32(1), int32(0)).Return(nil).Times(1)

		// Mock karma update for the post author
		mockPseudonymDAO.EXPECT().UpdateKarmaForPseudonym(gomock.Any(), "other-user").Return(nil).Times(1)

		// Create authenticated input
		input := createAuthenticatedVoteInput(userID, activePseudonymID, displayName, postID, 1)

		// Call handler
		response, err := handler.VoteOnPost(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
	})
}

// TestContentHandler_VoteOnPost_PreventsVotingOnDeletedPost tests that voting on deleted posts is prevented
func TestContentHandler_VoteOnPost_PreventsVotingOnDeletedPost(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("VoteOnDeletedPost", func(t *testing.T) {
		handler, mockPostDAO, _, _, _, _, _ := NewContentHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		postID := int64(123)

		// Create context with user
		ctx := createTestContentContext(t, userID, activePseudonymID, displayName)

		// Mock deleted post
		mockPost := &dbmodels.Post{
			PostID:      postID,
			Title:       "Deleted Post",
			SubforumID:  1,
			PseudonymID: "other-user",
			IsDeleted:   sql.Null[bool]{V: true, Valid: true},
		}
		mockPostDAO.EXPECT().GetPostByID(gomock.Any(), postID).Return(mockPost, nil).Times(1)

		// Create authenticated input
		input := createAuthenticatedVoteInput(userID, activePseudonymID, displayName, postID, 1)

		// Call handler - should fail because post is deleted
		response, err := handler.VoteOnPost(ctx, input)

		// Assertions - should fail with appropriate error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleted")
		assert.Nil(t, response)
	})
}

// TestContentHandler_VoteOnComment_PreventsVotingOnDeletedComment tests that voting on deleted comments is prevented
func TestContentHandler_VoteOnComment_PreventsVotingOnDeletedComment(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("VoteOnDeletedComment", func(t *testing.T) {
		handler, _, mockCommentDAO, _, _, _, _ := NewContentHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		commentID := int64(123)

		// Create context with user
		ctx := createTestContentContext(t, userID, activePseudonymID, displayName)

		// Mock deleted comment
		mockComment := &dbmodels.Comment{
			CommentID:   commentID,
			Content:     "Deleted comment",
			PostID:      1,
			PseudonymID: "other-user",
			IsDeleted:   sql.Null[bool]{V: true, Valid: true},
		}
		mockCommentDAO.EXPECT().GetCommentByID(gomock.Any(), commentID).Return(mockComment, nil).Times(1)

		// Create authenticated input for comment voting
		input := &models.CommentVoteInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWT(userID, activePseudonymID, displayName),
			},
			CommentID: commentID,
			Body: models.VoteInputBody{
				VoteValue: 1,
			},
		}

		// Call handler - should fail because comment is deleted
		response, err := handler.VoteOnComment(ctx, input)

		// Assertions - should fail with appropriate error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deleted")
		assert.Nil(t, response)
	})
}

// TestContentHandler_VoteOnPost_InsufficientPermissions tests post voting without required permissions
func TestContentHandler_VoteOnPost_InsufficientPermissions(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("VoteOnPostInsufficientPermissions", func(t *testing.T) {
		handler, mockPostDAO, _, _, mockPseudonymDAO, _, mockVoteDAO := NewContentHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		postID := int64(1)

		// Create context with user
		ctx := createTestContentContext(t, userID, activePseudonymID, displayName)

		// Mock post retrieval
		mockPost := &dbmodels.Post{
			PostID:      postID,
			Title:       "Test Post",
			SubforumID:  1,
			PseudonymID: "other-user",
			IsRemoved:   sql.Null[bool]{V: false, Valid: true},
			IsDeleted:   sql.Null[bool]{V: false, Valid: true},
		}
		mockPostDAO.EXPECT().GetPostByID(gomock.Any(), postID).Return(mockPost, nil).Times(1)

		// Mock vote operations
		mockVoteDAO.EXPECT().UpsertVote(gomock.Any(), activePseudonymID, "post", postID, int32(1)).Return(&dbmodels.Vote{
			VoteID:      1,
			PseudonymID: activePseudonymID,
			ContentType: "post",
			ContentID:   postID,
			VoteValue:   1,
		}, nil).Times(1)

		mockVoteDAO.EXPECT().GetVoteSummaryByContent(gomock.Any(), "post", postID).Return(int(1), int(0), int(1), nil).Times(1)

		// Mock pseudonym operations
		mockPseudonymDAO.EXPECT().UpdateLastActive(gomock.Any(), activePseudonymID).Return(nil).Times(1)
		mockPseudonymDAO.EXPECT().UpdateKarmaForPseudonym(gomock.Any(), "other-user").Return(nil).Times(1)

		// Mock post score update
		mockPostDAO.EXPECT().UpdatePostScore(gomock.Any(), postID, int32(1), int32(1), int32(0)).Return(nil).Times(1)

		// Create authenticated input
		input := createAuthenticatedVoteInput(userID, activePseudonymID, displayName, postID, 1)

		// Call handler - should succeed since there's no permission check for voting
		response, err := handler.VoteOnPost(ctx, input)

		// Assertions - should succeed
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
	})
}

// TestContentHandler_VoteOnComment_RemoveVote tests removing a vote on a comment
func TestContentHandler_VoteOnComment_RemoveVote(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("VoteOnCommentRemoveVote", func(t *testing.T) {
		handler, _, mockCommentDAO, _, mockPseudonymDAO, mockPermissionDAO, mockVoteDAO := NewContentHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		commentID := int64(1)

		// Create context with user
		ctx := createTestContentContext(t, userID, activePseudonymID, displayName)

		// Mock permission check
		mockPermissionDAO.EXPECT().HasUnifiedCapability(gomock.Any(), userID, activePseudonymID, constants.CapabilityVote, (*int32)(nil)).Return(true, nil).Times(1)

		// Mock comment
		mockComment := &dbmodels.Comment{
			CommentID:   commentID,
			Content:     "Test comment",
			PostID:      1,
			PseudonymID: "user-pseudonym-123", // Set the author's pseudonym ID
			IsDeleted:   sql.Null[bool]{V: false, Valid: true},
		}
		mockCommentDAO.EXPECT().GetCommentByID(gomock.Any(), commentID).Return(mockComment, nil).Times(1)

		// Mock vote removal (vote value 0 means remove vote)
		// First check if vote exists
		mockVoteDAO.EXPECT().GetVoteByPseudonymAndContent(gomock.Any(), activePseudonymID, "comment", commentID).Return(&dbmodels.Vote{
			VoteID:      1,
			PseudonymID: activePseudonymID,
			ContentType: "comment",
			ContentID:   commentID,
			VoteValue:   1,
		}, nil).Times(1)

		// Then delete the existing vote
		mockVoteDAO.EXPECT().DeleteVote(gomock.Any(), int64(1)).Return(nil).Times(1)

		// Mock vote summary
		mockVoteDAO.EXPECT().GetVoteSummaryByContent(gomock.Any(), "comment", commentID).Return(int(0), int(0), int(0), nil).Times(1)

		// Mock pseudonym operations
		mockPseudonymDAO.EXPECT().UpdateLastActive(gomock.Any(), activePseudonymID).Return(nil).Times(1)

		// Mock comment score update
		mockCommentDAO.EXPECT().UpdateCommentScore(gomock.Any(), commentID, int32(0), int32(0), int32(0)).Return(nil).Times(1)

		// Mock karma update for comment author
		mockPseudonymDAO.EXPECT().UpdateKarmaForPseudonym(gomock.Any(), "user-pseudonym-123").Return(nil).Times(1)

		// Create authenticated input for comment vote removal
		input := &models.CommentVoteInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWT(userID, activePseudonymID, displayName),
			},
			CommentID: commentID,
			Body: models.VoteInputBody{
				VoteValue: 0,
			},
		}

		// Call handler
		response, err := handler.VoteOnComment(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
	})
}

// TestContentHandler_VoteOnComment_InvalidVoteValue tests comment voting with invalid vote values
func TestContentHandler_VoteOnComment_InvalidVoteValue(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("VoteOnCommentInvalidVoteValue", func(t *testing.T) {
		handler, _, _, _, _, _, _ := NewContentHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		commentID := int64(1)

		// Create context with user
		ctx := createTestContentContext(t, userID, activePseudonymID, displayName)

		// Create authenticated input with invalid vote value (should be -1, 0, or 1)
		input := &models.CommentVoteInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWT(userID, activePseudonymID, displayName),
			},
			CommentID: commentID,
			Body: models.VoteInputBody{
				VoteValue: 5,
			},
		}

		// Call handler - should fail due to invalid vote value
		response, err := handler.VoteOnComment(ctx, input)

		// Assertions - should fail with validation error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "vote value")
		assert.Nil(t, response)
	})
}

// TestContentHandler_VoteOnComment_NotFound tests voting on non-existent comment
func TestContentHandler_VoteOnComment_NotFound(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("VoteOnCommentNotFound", func(t *testing.T) {
		handler, _, mockCommentDAO, _, _, mockPermissionDAO, _ := NewContentHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		commentID := int64(999)

		// Create context with user
		ctx := createTestContentContext(t, userID, activePseudonymID, displayName)

		// Mock permission check
		mockPermissionDAO.EXPECT().HasUnifiedCapability(gomock.Any(), userID, activePseudonymID, constants.CapabilityVote, (*int32)(nil)).Return(true, nil).Times(1)

		// Mock comment not found
		mockCommentDAO.EXPECT().GetCommentByID(gomock.Any(), commentID).Return(nil, sql.ErrNoRows).Times(1)

		// Create authenticated input for comment voting
		input := &models.CommentVoteInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWT(userID, activePseudonymID, displayName),
			},
			CommentID: commentID,
			Body: models.VoteInputBody{
				VoteValue: 1,
			},
		}

		// Call handler - should fail because comment doesn't exist
		response, err := handler.VoteOnComment(ctx, input)

		// Assertions - should fail with appropriate error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "sql: no rows in result set")
		assert.Nil(t, response)
	})
}

// TestContentHandler_VoteOnPost_NotFound tests voting on non-existent post
func TestContentHandler_VoteOnPost_NotFound(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("VoteOnPostNotFound", func(t *testing.T) {
		handler, mockPostDAO, _, _, _, mockPermissionDAO, _ := NewContentHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		postID := int64(999)

		// Create context with user
		ctx := createTestContentContext(t, userID, activePseudonymID, displayName)

		// Mock permission check
		mockPermissionDAO.EXPECT().HasUnifiedCapability(gomock.Any(), userID, activePseudonymID, constants.CapabilityVote, (*int32)(nil)).Return(true, nil).Times(1)

		// Mock post not found
		mockPostDAO.EXPECT().GetPostByID(gomock.Any(), postID).Return(nil, sql.ErrNoRows).Times(1)

		// Create authenticated input
		input := createAuthenticatedVoteInput(userID, activePseudonymID, displayName, postID, 1)

		// Call handler - should fail because post doesn't exist
		response, err := handler.VoteOnPost(ctx, input)

		// Assertions - should fail with appropriate error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "post not found")
		assert.Nil(t, response)
	})
}

// TestContentHandler_VoteOnComment_Success tests successful comment voting
func TestContentHandler_VoteOnComment_Success(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("VoteOnCommentSuccess", func(t *testing.T) {
		handler, _, mockCommentDAO, _, mockPseudonymDAO, mockPermissionDAO, mockVoteDAO := NewContentHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		commentID := int64(1)

		// Create context with user
		ctx := createTestContentContext(t, userID, activePseudonymID, displayName)

		// Mock permission check
		mockPermissionDAO.EXPECT().HasUnifiedCapability(gomock.Any(), userID, activePseudonymID, constants.CapabilityVote, (*int32)(nil)).Return(true, nil).Times(1)

		// Mock comment
		mockComment := &dbmodels.Comment{
			CommentID:   commentID,
			Content:     "Test comment",
			PostID:      1,
			PseudonymID: "other-user",
			IsDeleted:   sql.Null[bool]{V: false, Valid: true},
		}
		mockCommentDAO.EXPECT().GetCommentByID(gomock.Any(), commentID).Return(mockComment, nil).Times(1)

		// Mock vote upsert
		mockVoteDAO.EXPECT().UpsertVote(gomock.Any(), activePseudonymID, "comment", commentID, int32(1)).Return(&dbmodels.Vote{
			VoteID:      1,
			PseudonymID: activePseudonymID,
			ContentType: "comment",
			ContentID:   commentID,
			VoteValue:   1,
		}, nil).Times(1)

		// Mock vote summary retrieval
		mockVoteDAO.EXPECT().GetVoteSummaryByContent(gomock.Any(), "comment", commentID).Return(1, 0, 1, nil).Times(1)

		// Mock pseudonym last active update
		mockPseudonymDAO.EXPECT().UpdateLastActive(gomock.Any(), activePseudonymID).Return(nil).Times(1)

		// Mock comment score update
		mockCommentDAO.EXPECT().UpdateCommentScore(gomock.Any(), commentID, int32(1), int32(1), int32(0)).Return(nil).Times(1)

		// Mock karma update for the comment author
		mockPseudonymDAO.EXPECT().UpdateKarmaForPseudonym(gomock.Any(), "other-user").Return(nil).Times(1)

		// Create authenticated input for comment voting
		input := &models.CommentVoteInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + generateTestJWT(userID, activePseudonymID, displayName),
			},
			CommentID: commentID,
			Body: models.VoteInputBody{
				VoteValue: 1,
			},
		}

		// Call handler
		response, err := handler.VoteOnComment(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
	})
}
