package fixtures

import (
	"database/sql"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
)

// CreateTestPost creates a test post
func CreateTestPost() *dbmodels.Post {
	return &dbmodels.Post{
		PostID:      123,
		SubforumID:  1,
		PseudonymID: "test-pseudonym-id",
		Title:       "Test Post",
		Content:     sql.Null[string]{V: "Test content", Valid: true},
		PostType:    "text",
		Score:       sql.Null[int32]{V: 10, Valid: true},
		Upvotes:     sql.Null[int32]{V: 15, Valid: true},
		Downvotes:   sql.Null[int32]{V: 5, Valid: true},
		IsDeleted:   sql.Null[bool]{V: false, Valid: true},
		CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// CreateTestComment creates a test comment
func CreateTestComment() *dbmodels.Comment {
	return &dbmodels.Comment{
		CommentID:   456,
		PostID:      123,
		PseudonymID: "test-pseudonym-id",
		Content:     "Test comment",
		Score:       sql.Null[int32]{V: 5, Valid: true},
		Upvotes:     sql.Null[int32]{V: 8, Valid: true},
		Downvotes:   sql.Null[int32]{V: 3, Valid: true},
		IsDeleted:   sql.Null[bool]{V: false, Valid: true},
		CreatedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// CreateTestSubforum creates a test subforum
func CreateTestSubforum() *dbmodels.Subforum {
	return &dbmodels.Subforum{
		SubforumID:   1,
		Name:         "test-subforum",
		DisplayName:  "Test Subforum",
		Description:  sql.Null[string]{V: "A test subforum", Valid: true},
		IsPrivate:    sql.Null[bool]{V: false, Valid: true},
		IsNSFW:       sql.Null[bool]{V: false, Valid: true},
		IsRestricted: sql.Null[bool]{V: false, Valid: true},
	}
}

// CreateAuthenticatedContentInput creates a PostCreateInput with a valid JWT token for testing
func CreateAuthenticatedContentInput(userID int64, activePseudonymID string, displayName string, subforumName string, title string, content string) *apimodels.PostCreateInput {
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

	return &apimodels.PostCreateInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
		SubforumName: subforumName,
		Body: apimodels.PostCreateBody{
			Title:    title,
			Content:  content,
			PostType: "text",
		},
	}
}

// CreateAuthenticatedVoteInput creates a PostVoteInput with a valid JWT token for testing
func CreateAuthenticatedVoteInput(userID int64, activePseudonymID string, displayName string, postID int64, voteValue int) *apimodels.PostVoteInput {
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

	return &apimodels.PostVoteInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
		PostID: postID,
		Body: apimodels.VoteInputBody{
			VoteValue: voteValue,
		},
	}
}

// CreateAuthenticatedDeleteInput creates a PostDeleteInput with a valid JWT token for testing
func CreateAuthenticatedDeleteInput(userID int64, activePseudonymID string, displayName string, postID int64, reason string) *apimodels.PostDeleteInput {
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

	return &apimodels.PostDeleteInput{
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
