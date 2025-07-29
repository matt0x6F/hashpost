package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/stephenafamo/bob/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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
		name  string
		input *struct {
			middleware.AuthInput
			models.ReportInput
		}
		userContext   *middleware.UserContext
		expectedError bool
		errorContains string
	}{
		{
			name: "Successful Report",
			input: &struct {
				middleware.AuthInput
				models.ReportInput
			}{
				ReportInput: models.ReportInput{
					Body: models.ReportInputBody{
						ContentType:   "post",
						ContentID:     &[]int{123}[0],
						ReportReason:  "spam",
						ReportDetails: "This post violates community guidelines",
					},
				},
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Report Without Content ID",
			input: &struct {
				middleware.AuthInput
				models.ReportInput
			}{
				ReportInput: models.ReportInput{
					Body: models.ReportInputBody{
						ContentType:   "user",
						ReportReason:  "harassment",
						ReportDetails: "User is harassing others",
					},
				},
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Report With Reported Pseudonym",
			input: &struct {
				middleware.AuthInput
				models.ReportInput
			}{
				ReportInput: models.ReportInput{
					Body: models.ReportInputBody{
						ContentType:         "user",
						ReportReason:        "harassment",
						ReportDetails:       "User is harassing others",
						ReportedPseudonymID: "reported-pseudonym-456",
					},
				},
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Authentication Required",
			input: &struct {
				middleware.AuthInput
				models.ReportInput
			}{
				ReportInput: models.ReportInput{
					Body: models.ReportInputBody{
						ContentType:  "post",
						ContentID:    &[]int{123}[0],
						ReportReason: "spam",
					},
				},
			},
			userContext:   nil,
			expectedError: true,
			errorContains: "Authentication required",
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

			// Set up the input with proper authentication context
			var authInput middleware.AuthInput
			if tt.userContext != nil {
				// Create a valid JWT token for testing
				token, tokenErr := middleware.GenerateJWT(tt.userContext, "test-secret", 24*time.Hour)
				require.NoError(t, tokenErr)
				authInput = middleware.AuthInput{
					Authorization: "Bearer " + token,
					AccessToken:   "",
				}
			} else {
				authInput = middleware.AuthInput{
					Authorization: "",
					AccessToken:   "",
				}
			}

			inputWithAuth := &struct {
				middleware.AuthInput
				models.ReportInput
			}{
				AuthInput:   authInput,
				ReportInput: tt.input.ReportInput,
			}

			response, err := handler.ReportContent(ctx, inputWithAuth)

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

func TestModerationHandler_GetSubforumReports(t *testing.T) {
	tests := []struct {
		name  string
		input *struct {
			middleware.AuthInput
			SubforumPath string `path:"subforum_path" example:"b/hashpost"`
			Status       string `query:"status" example:"pending"`
			Page         int    `query:"page" example:"1"`
			Limit        int    `query:"limit" example:"25"`
		}
		userContext   *middleware.UserContext
		expectedError bool
		errorContains string
	}{
		{
			name: "Get Pending Reports",
			input: &struct {
				middleware.AuthInput
				SubforumPath string `path:"subforum_path" example:"b/hashpost"`
				Status       string `query:"status" example:"pending"`
				Page         int    `query:"page" example:"1"`
				Limit        int    `query:"limit" example:"25"`
			}{
				SubforumPath: "b/test-subforum",
				Status:       "pending",
				Page:         1,
				Limit:        25,
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Get Resolved Reports",
			input: &struct {
				middleware.AuthInput
				SubforumPath string `path:"subforum_path" example:"b/hashpost"`
				Status       string `query:"status" example:"pending"`
				Page         int    `query:"page" example:"1"`
				Limit        int    `query:"limit" example:"25"`
			}{
				SubforumPath: "b/test-subforum",
				Status:       "resolved",
				Page:         1,
				Limit:        10,
			},
			userContext:   createTestUserContext(),
			expectedError: false,
		},
		{
			name: "Authentication Required",
			input: &struct {
				middleware.AuthInput
				SubforumPath string `path:"subforum_path" example:"b/hashpost"`
				Status       string `query:"status" example:"pending"`
				Page         int    `query:"page" example:"1"`
				Limit        int    `query:"limit" example:"25"`
			}{
				SubforumPath: "b/test-subforum",
				Status:       "pending",
				Page:         1,
				Limit:        25,
			},
			userContext:   nil,
			expectedError: true,
			errorContains: "Authentication required",
		},
		{
			name: "Insufficient Permissions",
			input: &struct {
				middleware.AuthInput
				SubforumPath string `path:"subforum_path" example:"b/hashpost"`
				Status       string `query:"status" example:"pending"`
				Page         int    `query:"page" example:"1"`
				Limit        int    `query:"limit" example:"25"`
			}{
				SubforumPath: "b/test-subforum",
				Status:       "pending",
				Page:         1,
				Limit:        25,
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
			errorContains: "Insufficient permissions for this subforum",
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

			// Set up the input with proper authentication context
			var authInput middleware.AuthInput
			if tt.userContext != nil {
				// Create a valid JWT token for testing
				token, tokenErr := middleware.GenerateJWT(tt.userContext, "test-secret", 24*time.Hour)
				require.NoError(t, tokenErr)
				authInput = middleware.AuthInput{
					Authorization: "Bearer " + token,
					AccessToken:   "",
				}
			} else {
				authInput = middleware.AuthInput{
					Authorization: "",
					AccessToken:   "",
				}
			}

			inputWithAuth := &struct {
				middleware.AuthInput
				SubforumPath string `path:"subforum_path" example:"b/hashpost"`
				Status       string `query:"status" example:"pending"`
				Page         int    `query:"page" example:"1"`
				Limit        int    `query:"limit" example:"25"`
			}{
				AuthInput:    authInput,
				SubforumPath: tt.input.SubforumPath,
				Status:       tt.input.Status,
				Page:         tt.input.Page,
				Limit:        tt.input.Limit,
			}

			response, err := handler.GetSubforumReports(ctx, inputWithAuth)

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

// NewModerationHandlerWithMocks creates a new moderation handler with mock DAOs and fixture data
func NewModerationHandlerWithMocks() *ModerationHandler {
	// Create a mock auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	// Create mock DAOs
	mockReportDAO := mocks.NewMockReportDAO()
	mockModerationActionDAO := mocks.NewMockModerationActionDAO()
	mockUserBanDAO := mocks.NewMockUserBanDAO()
	mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()
	mockSubforumDAO := mocks.NewMockSubforumDAO()
	mockPostDAO := mocks.NewMockPostDAO()
	mockCommentDAO := mocks.NewMockCommentDAO()
	mockVoteDAO := mocks.NewMockVoteDAO()
	mockPermissionDAO := mocks.NewMockPermissionDAO()

	// Inject fixture data into mocks
	mockReportDAO.InjectReport(fixtures.CreateTestReport())
	mockReportDAO.InjectReport(fixtures.CreateTestReportWithResolution())
	mockReportDAO.InjectReportsByStatus("pending", []*dbmodels.Report{fixtures.CreateTestReport()})
	mockReportDAO.InjectReportsByStatus("resolved", []*dbmodels.Report{fixtures.CreateTestReportWithResolution()})
	mockReportDAO.InjectCount("pending", 1)
	mockReportDAO.InjectCount("resolved", 1)
	mockReportDAO.SetDefaultBehavior()

	mockModerationActionDAO.InjectAction(fixtures.CreateTestModerationAction())
	mockModerationActionDAO.InjectAction(fixtures.CreateTestModerationActionWithDetails())
	mockModerationActionDAO.InjectActionsByType("remove_post", []*dbmodels.ModerationAction{fixtures.CreateTestModerationAction()})
	mockModerationActionDAO.InjectCount("remove_post", 1)
	mockModerationActionDAO.SetDefaultBehavior()

	mockUserBanDAO.InjectBan(fixtures.CreateTestUserBan())
	mockUserBanDAO.InjectBan(fixtures.CreateTestPermanentUserBan())
	mockUserBanDAO.InjectBan(fixtures.CreateTestInactiveUserBan())
	mockUserBanDAO.InjectBansBySubforum(1, []*dbmodels.UserBan{fixtures.CreateTestUserBan()})
	mockUserBanDAO.InjectCount("subforum_1", 1)
	mockUserBanDAO.SetDefaultBehavior()

	// Set up mock secure pseudonym DAO with fixture data
	mockSecurePseudonymDAO.InjectPseudonym(fixtures.CreateTestPseudonym())
	// Inject the pseudonyms that the test reports are looking for
	mockSecurePseudonymDAO.InjectPseudonym(&dbmodels.Pseudonym{
		PseudonymID: "reporter-pseudonym-id",
		DisplayName: "ReporterUser",
	})
	mockSecurePseudonymDAO.InjectPseudonym(&dbmodels.Pseudonym{
		PseudonymID: "reported-pseudonym-id",
		DisplayName: "ReportedUser",
	})
	mockSecurePseudonymDAO.InjectPseudonym(&dbmodels.Pseudonym{
		PseudonymID: "moderator-pseudonym-id",
		DisplayName: "ModeratorUser",
	})
	mockSecurePseudonymDAO.SetDefaultBehavior()

	// Set up mock post DAO with fixture data
	mockPostDAO.InjectPost(fixtures.CreateTestPost())
	mockPostDAO.SetDefaultBehavior()

	// Set up mock comment DAO with fixture data
	mockCommentDAO.InjectComment(fixtures.CreateTestComment())
	mockCommentDAO.SetDefaultBehavior()

	// Set up mock vote DAO with fixture data
	mockVoteDAO.SetDefaultBehavior()

	// Set up mock subforum DAO with fixture data
	testSubforum := fixtures.CreateTestSubforum()
	mockSubforumDAO.InjectSubforum(testSubforum)
	mockSubforumDAO.InjectSubforumByCommunityTypeAndName("b", "test-subforum", testSubforum)
	mockSubforumDAO.SetDefaultBehavior()

	// Debug: Print the subforum name to verify it's set up correctly
	fmt.Printf("DEBUG: Test subforum name: %s\n", testSubforum.Name)

	// Set up mock permission DAO with fixture data
	// Inject capability for the test user (userID=1, subforumID=1, capability="moderate_content", activePseudonymID="moderator-pseudonym-123")
	mockPermissionDAO.InjectSubforumCapabilityWithActivePseudonym(1, 1, "moderate_content", "moderator-pseudonym-123", true)

	// Set up unified capabilities for global moderation endpoints
	mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities", mock.Anything, int64(1), "moderator-pseudonym-123", (*int32)(nil)).
		Return([]string{"user", "moderator"}, []string{"create_content", "vote", "message", "report", "moderate_content"}, nil)

	// Set up unified capabilities for regular user (userID=2)
	mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities", mock.Anything, int64(2), "user-pseudonym-456", (*int32)(nil)).
		Return([]string{"user"}, []string{"create_content", "vote"}, nil)

	mockPermissionDAO.SetDefaultBehavior()

	return &ModerationHandler{
		reportDAO:           mockReportDAO,
		moderationActionDAO: mockModerationActionDAO,
		userBanDAO:          mockUserBanDAO,
		securePseudonymDAO:  mockSecurePseudonymDAO,
		subforumDAO:         mockSubforumDAO,
		postDAO:             mockPostDAO,
		commentDAO:          mockCommentDAO,
		voteDAO:             mockVoteDAO,
		permissionDAO:       mockPermissionDAO,
	}
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
		mockVoteDAO := &mocks.MockVoteDAO{}
		mockPermissionDAO := &mocks.MockPermissionDAO{}

		// Create handler with dependencies
		handler := NewModerationHandler(
			mockReportDAO,
			mockModerationActionDAO,
			mockUserBanDAO,
			mockSecurePseudonymDAO,
			mockSubforumDAO,
			mockPostDAO,
			mockCommentDAO,
			mockVoteDAO,
			mockPermissionDAO,
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
		assert.Equal(t, mockVoteDAO, handler.voteDAO)
		assert.Equal(t, mockPermissionDAO, handler.permissionDAO)
	})
}
