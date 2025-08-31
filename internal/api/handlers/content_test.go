package handlers_test

import (
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
)

// createAuthenticatedContentInput creates a PostCreateInput with a valid JWT token for testing
func createAuthenticatedContentInput(userID int64, activePseudonymID string, displayName string, subforumName string, title string, content string) *models.PostCreateInput {
	// Create a user context
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false,
	}

	// Generate a JWT token
	token, _ := middleware.GenerateJWT(user, "test-secret", time.Hour)

	return &models.PostCreateInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
		SubforumName: subforumName,
		Body: models.PostCreateBody{
			Title:    title,
			Content:  content,
			PostType: "text",
		},
	}
}

// createAuthenticatedVoteInput creates a PostVoteInput with a valid JWT token for testing
func createAuthenticatedVoteInput(userID int64, activePseudonymID string, displayName string, postID int64, voteValue int) *models.PostVoteInput {
	// Create a user context
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false,
	}

	// Generate a JWT token
	token, _ := middleware.GenerateJWT(user, "test-secret", time.Hour)

	return &models.PostVoteInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
		PostID: postID,
		Body: models.VoteInputBody{
			VoteValue: voteValue,
		},
	}
}

// createAuthenticatedDeleteInput creates a PostDeleteInput with a valid JWT token for testing
func createAuthenticatedDeleteInput(userID int64, activePseudonymID string, displayName string, postID int64, reason string) *models.PostDeleteInput {
	// Create a user context
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false,
	}

	// Generate a JWT token
	token, _ := middleware.GenerateJWT(user, "test-secret", time.Hour)

	return &models.PostDeleteInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
		PostID: postID,
		Body: struct {
			Reason string `json:"reason,omitempty" example:"User requested deletion"`
		}{
			Reason: reason,
		},
	}
}
