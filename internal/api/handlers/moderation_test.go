package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	"github.com/stephenafamo/bob/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestUserContext creates a test user context with moderation capabilities
func createTestUserContext() *middleware.UserContext {
	return &middleware.UserContext{
		UserID:            1,
		Email:             "moderator@example.com",
		ActivePseudonymID: "moderator-pseudonym-123",
		DisplayName:       "TestModerator",
		Roles:             []string{"user", "moderator"},
		Capabilities:      []string{"create_content", "vote", "message", "report", "create_subforum", "moderate_content"},
		MFAEnabled:        false,
	}
}

// createTestContextWithUser creates a context with user information
func createTestContextWithUser(userCtx *middleware.UserContext) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, middleware.UserContextKeyValue, userCtx)
}

func TestNewModerationHandlerWithMocks(t *testing.T) {
	// Test that the mock handler can be created successfully
	handler := NewModerationHandlerWithMocks()
	require.NotNil(t, handler)

	// Test that all DAOs are properly initialized
	assert.NotNil(t, handler.reportDAO)
	assert.NotNil(t, handler.moderationActionDAO)
	assert.NotNil(t, handler.userBanDAO)
	assert.NotNil(t, handler.securePseudonymDAO)
	assert.NotNil(t, handler.subforumDAO)
	assert.NotNil(t, handler.postDAO)
	assert.NotNil(t, handler.commentDAO)
}

func TestModerationHandler_ReportContent(t *testing.T) {
	tests := []struct {
		name          string
		input         *models.ReportInput
		userContext   *middleware.UserContext
		expectedError bool
		errorContains string
	}{
		{
			name: "Successful Report",
			input: &models.ReportInput{
				Body: models.ReportInputBody{
					ContentType:   "post",
					ContentID:     &[]int{123}[0],
					ReportReason:  "spam",
					ReportDetails: "This post violates community guidelines",
				},
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Report Without Content ID",
			input: &models.ReportInput{
				Body: models.ReportInputBody{
					ContentType:   "user",
					ReportReason:  "harassment",
					ReportDetails: "User is harassing others",
				},
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Report With Reported Pseudonym",
			input: &models.ReportInput{
				Body: models.ReportInputBody{
					ContentType:         "user",
					ReportReason:        "harassment",
					ReportDetails:       "User is harassing others",
					ReportedPseudonymID: "reported-pseudonym-456",
				},
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Authentication Required",
			input: &models.ReportInput{
				Body: models.ReportInputBody{
					ContentType:  "post",
					ContentID:    &[]int{123}[0],
					ReportReason: "spam",
				},
			},
			userContext:   nil,
			expectedError: true,
			errorContains: "authentication required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewModerationHandlerWithMocks()

			var ctx context.Context
			if tt.userContext != nil {
				ctx = createTestContextWithUser(tt.userContext)
			} else {
				ctx = context.Background()
			}

			response, err := handler.ReportContent(ctx, tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotZero(t, response.Body.ReportID)
			}
		})
	}
}

func TestModerationHandler_GetReports(t *testing.T) {
	tests := []struct {
		name          string
		input         *models.ReportsListInput
		userContext   *middleware.UserContext
		expectedError bool
		errorContains string
	}{
		{
			name: "Get Pending Reports",
			input: &models.ReportsListInput{
				Status: "pending",
				Page:   1,
				Limit:  25,
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Get Resolved Reports",
			input: &models.ReportsListInput{
				Status: "resolved",
				Page:   1,
				Limit:  10,
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Authentication Required",
			input: &models.ReportsListInput{
				Status: "pending",
				Page:   1,
				Limit:  25,
			},
			userContext:   nil,
			expectedError: true,
			errorContains: "authentication required",
		},
		{
			name: "Insufficient Permissions",
			input: &models.ReportsListInput{
				Status: "pending",
				Page:   1,
				Limit:  25,
			},
			userContext: &middleware.UserContext{
				UserID:            2,
				Email:             "user@example.com",
				ActivePseudonymID: "user-pseudonym-456",
				DisplayName:       "RegularUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"create_content", "vote"},
				MFAEnabled:        false,
			},
			expectedError: true,
			errorContains: "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewModerationHandlerWithMocks()

			var ctx context.Context
			if tt.userContext != nil {
				ctx = createTestContextWithUser(tt.userContext)
			} else {
				ctx = context.Background()
			}

			response, err := handler.GetReports(ctx, tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotNil(t, response.Body.Reports)
				assert.Equal(t, tt.input.Page, response.Body.Pagination.Page)
				assert.Equal(t, tt.input.Limit, response.Body.Pagination.Limit)
			}
		})
	}
}

func TestModerationHandler_RemoveContent(t *testing.T) {
	tests := []struct {
		name          string
		input         *models.ContentRemovalInput
		userContext   *middleware.UserContext
		expectedError bool
		errorContains string
	}{
		{
			name: "Remove Post",
			input: &models.ContentRemovalInput{
				ContentType: "post",
				ContentID:   123,
				Body: models.ContentRemovalInputBody{
					RemovalReason: "violates community guidelines",
				},
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Remove Comment",
			input: &models.ContentRemovalInput{
				ContentType: "comment",
				ContentID:   456,
				Body: models.ContentRemovalInputBody{
					RemovalReason: "harassment",
				},
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Authentication Required",
			input: &models.ContentRemovalInput{
				ContentType: "post",
				ContentID:   123,
				Body: models.ContentRemovalInputBody{
					RemovalReason: "violates community guidelines",
				},
			},
			userContext:   nil,
			expectedError: true,
			errorContains: "authentication required",
		},
		{
			name: "Insufficient Permissions",
			input: &models.ContentRemovalInput{
				ContentType: "post",
				ContentID:   123,
				Body: models.ContentRemovalInputBody{
					RemovalReason: "violates community guidelines",
				},
			},
			userContext: &middleware.UserContext{
				UserID:            2,
				Email:             "user@example.com",
				ActivePseudonymID: "user-pseudonym-456",
				DisplayName:       "RegularUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"create_content", "vote"},
				MFAEnabled:        false,
			},
			expectedError: true,
			errorContains: "insufficient permissions",
		},
		{
			name: "Unsupported Content Type",
			input: &models.ContentRemovalInput{
				ContentType: "invalid",
				ContentID:   123,
				Body: models.ContentRemovalInputBody{
					RemovalReason: "test",
				},
			},
			userContext:   createTestUserContext(),
			expectedError: true,
			errorContains: "unsupported content type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewModerationHandlerWithMocks()

			var ctx context.Context
			if tt.userContext != nil {
				ctx = createTestContextWithUser(tt.userContext)
			} else {
				ctx = context.Background()
			}

			response, err := handler.RemoveContent(ctx, tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.input.ContentID, response.Body.ContentID)
				assert.Equal(t, tt.input.ContentType, response.Body.ContentType)
				assert.Equal(t, tt.input.Body.RemovalReason, response.Body.RemovalReason)
				assert.NotEmpty(t, response.Body.RemovedBy.PseudonymID)
				assert.NotEmpty(t, response.Body.RemovedBy.DisplayName)
			}
		})
	}
}

func TestModerationHandler_BanUser(t *testing.T) {
	durationDays := 30
	tests := []struct {
		name          string
		input         *models.UserBanInput
		userContext   *middleware.UserContext
		expectedError bool
		errorContains string
	}{
		{
			name: "Temporary Ban",
			input: &models.UserBanInput{
				PseudonymID: "test-pseudonym-id",
				Body: models.UserBanInputBody{
					SubforumID:   1,
					BanReason:    "repeated violations",
					IsPermanent:  false,
					DurationDays: &durationDays,
				},
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Permanent Ban",
			input: &models.UserBanInput{
				PseudonymID: "test-pseudonym-id",
				Body: models.UserBanInputBody{
					SubforumID:  1,
					BanReason:   "severe violations",
					IsPermanent: true,
				},
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Authentication Required",
			input: &models.UserBanInput{
				PseudonymID: "test-pseudonym-id",
				Body: models.UserBanInputBody{
					SubforumID:  1,
					BanReason:   "violations",
					IsPermanent: false,
				},
			},
			userContext:   nil,
			expectedError: true,
			errorContains: "authentication required",
		},
		{
			name: "Insufficient Permissions",
			input: &models.UserBanInput{
				PseudonymID: "test-pseudonym-id",
				Body: models.UserBanInputBody{
					SubforumID:  1,
					BanReason:   "violations",
					IsPermanent: false,
				},
			},
			userContext: &middleware.UserContext{
				UserID:            2,
				Email:             "user@example.com",
				ActivePseudonymID: "user-pseudonym-456",
				DisplayName:       "RegularUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"create_content", "vote"},
				MFAEnabled:        false,
			},
			expectedError: true,
			errorContains: "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewModerationHandlerWithMocks()

			var ctx context.Context
			if tt.userContext != nil {
				ctx = createTestContextWithUser(tt.userContext)
			} else {
				ctx = context.Background()
			}

			response, err := handler.BanUser(ctx, tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotZero(t, response.Body.BanID)
				assert.Equal(t, tt.input.Body.SubforumID, response.Body.SubforumID)
				assert.Equal(t, tt.input.Body.BanReason, response.Body.BanReason)
				assert.Equal(t, tt.input.Body.IsPermanent, response.Body.IsPermanent)
				assert.NotEmpty(t, response.Body.BannedFingerprint)
				assert.NotEmpty(t, response.Body.BannedBy.PseudonymID)
				assert.NotEmpty(t, response.Body.BannedBy.DisplayName)
			}
		})
	}
}

func TestModerationHandler_GetModerationHistory(t *testing.T) {
	tests := []struct {
		name          string
		input         *models.ModerationHistoryInput
		userContext   *middleware.UserContext
		expectedError bool
		errorContains string
	}{
		{
			name: "Get Remove Post Actions",
			input: &models.ModerationHistoryInput{
				ActionType: "remove_post",
				Page:       1,
				Limit:      25,
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Get All Actions",
			input: &models.ModerationHistoryInput{
				ActionType: "",
				Page:       1,
				Limit:      10,
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Authentication Required",
			input: &models.ModerationHistoryInput{
				ActionType: "remove_post",
				Page:       1,
				Limit:      25,
			},
			userContext:   nil,
			expectedError: true,
			errorContains: "authentication required",
		},
		{
			name: "Insufficient Permissions",
			input: &models.ModerationHistoryInput{
				ActionType: "remove_post",
				Page:       1,
				Limit:      25,
			},
			userContext: &middleware.UserContext{
				UserID:            2,
				Email:             "user@example.com",
				ActivePseudonymID: "user-pseudonym-456",
				DisplayName:       "RegularUser",
				Roles:             []string{"user"},
				Capabilities:      []string{"create_content", "vote"},
				MFAEnabled:        false,
			},
			expectedError: true,
			errorContains: "insufficient permissions",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewModerationHandlerWithMocks()

			var ctx context.Context
			if tt.userContext != nil {
				ctx = createTestContextWithUser(tt.userContext)
			} else {
				ctx = context.Background()
			}

			response, err := handler.GetModerationHistory(ctx, tt.input)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotNil(t, response.Body.Actions)
				assert.Equal(t, tt.input.Page, response.Body.Pagination.Page)
				assert.Equal(t, tt.input.Limit, response.Body.Pagination.Limit)
			}
		})
	}
}

func TestModerationHandler_HelperMethods(t *testing.T) {
	handler := NewModerationHandlerWithMocks()
	userCtx := createTestUserContext()

	t.Run("ExtractUserFromContext", func(t *testing.T) {
		ctx := createTestContextWithUser(userCtx)
		extractedUser, err := handler.extractUserFromContext(ctx)

		assert.NoError(t, err)
		assert.NotNil(t, extractedUser)
		assert.Equal(t, userCtx.UserID, extractedUser.UserID)
		assert.Equal(t, userCtx.ActivePseudonymID, extractedUser.ActivePseudonymID)
	})

	t.Run("ValidateModeratorPermissions", func(t *testing.T) {
		// Test with moderator permissions
		err := handler.validateModeratorPermissions(userCtx)
		assert.NoError(t, err)

		// Test with insufficient permissions
		regularUser := &middleware.UserContext{
			UserID:            2,
			Email:             "user@example.com",
			ActivePseudonymID: "user-pseudonym-456",
			DisplayName:       "RegularUser",
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote"},
			MFAEnabled:        false,
		}
		err = handler.validateModeratorPermissions(regularUser)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "moderation permissions")
	})

	t.Run("ParseActionDetails", func(t *testing.T) {
		// Test with invalid JSON (nil)
		details := handler.parseActionDetails(sql.Null[types.JSON[json.RawMessage]]{Valid: false})
		assert.Nil(t, details)
	})
}

// TestNewModerationHandler tests the main constructor function
func TestNewModerationHandler(t *testing.T) {
	t.Run("NewModerationHandlerSuccess", func(t *testing.T) {
		// Create mock DAOs
		mockReportDAO := &mocks.MockReportDAO{}
		mockModerationActionDAO := &mocks.MockModerationActionDAO{}
		mockUserBanDAO := &mocks.MockUserBanDAO{}
		mockSecurePseudonymDAO := &mocks.MockSecurePseudonymDAO{}
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockPostDAO := &mocks.MockPostDAO{}
		mockCommentDAO := &mocks.MockCommentDAO{}

		// Create handler with dependencies
		handler := NewModerationHandler(
			mockReportDAO,
			mockModerationActionDAO,
			mockUserBanDAO,
			mockSecurePseudonymDAO,
			mockSubforumDAO,
			mockPostDAO,
			mockCommentDAO,
		)

		// Verify handler is created
		assert.NotNil(t, handler)
		assert.Equal(t, mockReportDAO, handler.reportDAO)
		assert.Equal(t, mockModerationActionDAO, handler.moderationActionDAO)
		assert.Equal(t, mockUserBanDAO, handler.userBanDAO)
		assert.Equal(t, mockSecurePseudonymDAO, handler.securePseudonymDAO)
		assert.Equal(t, mockSubforumDAO, handler.subforumDAO)
		assert.Equal(t, mockPostDAO, handler.postDAO)
		assert.Equal(t, mockCommentDAO, handler.commentDAO)
	})
}
