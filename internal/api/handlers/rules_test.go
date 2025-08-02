package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/stephenafamo/bob/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewRulesHandler(t *testing.T) {
	// Create mock DAOs
	mockReportDAO := mocks.NewMockReportDAO()
	mockSubforumDAO := mocks.NewMockSubforumDAO()
	mockSystemSettingsDAO := mocks.NewMockSystemSettingsDAO()
	mockPermissionDAO := mocks.NewMockPermissionDAO()

	// Create handler
	handler := NewRulesHandler(mockReportDAO, mockSubforumDAO, mockSystemSettingsDAO, mockPermissionDAO, nil)

	// Assertions
	assert.NotNil(t, handler)
	assert.Equal(t, mockReportDAO, handler.reportDAO)
	assert.Equal(t, mockSubforumDAO, handler.subforumDAO)
	assert.Equal(t, mockSystemSettingsDAO, handler.systemSettingsDAO)
	assert.Equal(t, mockPermissionDAO, handler.permissionDAO)
	assert.Nil(t, handler.db)
}

func TestRulesHandler_GetPlatformRules(t *testing.T) {
	tests := []struct {
		name           string
		input          *apimodels.PlatformRulesInput
		mockSetting    *dbmodels.SystemSetting
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "GetPlatformRulesSuccess",
			input: &apimodels.PlatformRulesInput{
				ActiveOnly: false,
			},
			mockSetting: &dbmodels.SystemSetting{
				SettingKey:   "platform_rules",
				SettingValue: `[{"code":"no_spam","name":"No Spam","description":"No spam allowed","category":"content","severity":"high","active":true}]`,
				SettingType:  "json",
			},
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name:           "GetPlatformRulesNoRules",
			input:          &apimodels.PlatformRulesInput{},
			mockSetting:    nil,
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "GetPlatformRulesActiveOnly",
			input: &apimodels.PlatformRulesInput{
				ActiveOnly: true,
			},
			mockSetting: &dbmodels.SystemSetting{
				SettingKey:   "platform_rules",
				SettingValue: `[{"code":"no_spam","name":"No Spam","description":"No spam allowed","category":"content","severity":"high","active":true},{"code":"old_rule","name":"Old Rule","description":"Inactive rule","category":"content","severity":"low","active":false}]`,
				SettingType:  "json",
			},
			wantErr:        false,
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSystemSettingsDAO := mocks.NewMockSystemSettingsDAO()

			// Set up mock expectations
			mockSystemSettingsDAO.On("GetSetting", mock.Anything, "platform_rules").Return(tt.mockSetting, nil)

			// Create handler
			handler := &RulesHandler{
				systemSettingsDAO: mockSystemSettingsDAO,
			}

			// Execute test
			result, err := handler.GetPlatformRules(context.Background(), tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedStatus, result.Status)
			}

			// Verify mocks
			mockSystemSettingsDAO.AssertExpectations(t)
		})
	}
}

func TestRulesHandler_GetSubforumRules(t *testing.T) {
	tests := []struct {
		name           string
		input          *apimodels.SubforumRulesInput
		mockSubforum   *dbmodels.Subforum
		wantErr        bool
		expectedStatus int
	}{
		{
			name: "GetSubforumRulesSuccess",
			input: &apimodels.SubforumRulesInput{
				CommunityType: "t",
				SubforumName:  "golang",
				ActiveOnly:    false,
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: true,
					V:     types.NewJSON[json.RawMessage]([]byte(`[{"code":"no_politics","name":"No Politics","description":"No political discussion","category":"content","severity":"medium","active":true}]`)),
				},
			},
			wantErr:        false,
			expectedStatus: 200,
		},
		{
			name: "GetSubforumRulesNotFound",
			input: &apimodels.SubforumRulesInput{
				CommunityType: "t",
				SubforumName:  "nonexistent",
				ActiveOnly:    false,
			},
			mockSubforum:   nil,
			wantErr:        true,
			expectedStatus: 404,
		},
		{
			name: "GetSubforumRulesNoRules",
			input: &apimodels.SubforumRulesInput{
				CommunityType: "t",
				SubforumName:  "golang",
				ActiveOnly:    false,
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: false,
					V:     types.NewJSON[json.RawMessage]([]byte(`{}`)),
				},
			},
			wantErr:        false,
			expectedStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()

			// Set up mock expectations
			if tt.mockSubforum != nil {
				mockSubforumDAO.On("GetSubforumByName", mock.Anything, tt.input.SubforumName).Return(tt.mockSubforum, nil)
			} else {
				mockSubforumDAO.On("GetSubforumByName", mock.Anything, tt.input.SubforumName).Return(nil, nil)
			}

			// Create handler
			handler := &RulesHandler{
				subforumDAO: mockSubforumDAO,
			}

			// Execute test
			result, err := handler.GetSubforumRules(context.Background(), tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.expectedStatus, result.Status)
			}

			// Verify mocks
			mockSubforumDAO.AssertExpectations(t)
		})
	}
}

func TestRulesHandler_CreateSubforumRule(t *testing.T) {
	t.Skip("TODO: Mock database operations for CreateSubforumRule test")
	tests := []struct {
		name           string
		input          *apimodels.RuleCreateInput
		mockSubforum   *dbmodels.Subforum
		mockPermission bool
		wantErr        bool
	}{
		{
			name: "CreateSubforumRuleSuccess",
			input: &apimodels.RuleCreateInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
				CommunityType: "t",
				SubforumName:  "golang",
				Body: apimodels.RuleCreateInputBody{
					Code:        "no_politics",
					Name:        "No Politics",
					Description: "No political discussion allowed",
					Category:    "content",
					Severity:    "medium",
					Active:      true,
				},
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: false,
					V:     types.NewJSON[json.RawMessage]([]byte(`{}`)),
				},
			},
			mockPermission: true,
			wantErr:        false,
		},
		{
			name: "CreateSubforumRuleInsufficientPermissions",
			input: &apimodels.RuleCreateInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
				CommunityType: "t",
				SubforumName:  "golang",
				Body: apimodels.RuleCreateInputBody{
					Code:        "no_politics",
					Name:        "No Politics",
					Description: "No political discussion allowed",
					Category:    "content",
					Severity:    "medium",
					Active:      true,
				},
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: false,
					V:     types.NewJSON[json.RawMessage]([]byte(`{}`)),
				},
			},
			mockPermission: false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize global auth middleware for testing
			authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
			middleware.SetGlobalAuthMiddleware(authMiddleware)

			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock expectations only if authentication should succeed
			if tt.mockSubforum != nil {
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.CommunityType, tt.input.SubforumName).Return(tt.mockSubforum, nil)
				subforumID := int32(1)
				mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "moderator-pseudonym-123", "manage_subforum_rules", &subforumID).Return(tt.mockPermission, nil)
			}

			// Create handler
			handler := &RulesHandler{
				subforumDAO:   mockSubforumDAO,
				permissionDAO: mockPermissionDAO,
				db:            nil, // Mock database - the method will fail if it tries to use it
			}

			// Set up proper authentication
			userCtx := fixtures.CreateTestUserContext()
			token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
			require.NoError(t, tokenErr)

			// Update input with proper authentication
			inputWithAuth := &apimodels.RuleCreateInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + token,
					AccessToken:   "",
				},
				CommunityType: tt.input.CommunityType,
				SubforumName:  tt.input.SubforumName,
				Body:          tt.input.Body,
			}

			// Execute test
			result, err := handler.CreateSubforumRule(context.Background(), inputWithAuth)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			// Verify mocks only if they were set up
			if tt.mockSubforum != nil {
				mockSubforumDAO.AssertExpectations(t)
				mockPermissionDAO.AssertExpectations(t)
			}
		})
	}
}

func TestRulesHandler_UpdateSubforumRule(t *testing.T) {
	t.Skip("TODO: Mock database operations for UpdateSubforumRule test")
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *apimodels.RuleUpdateInput
		mockSubforum   *dbmodels.Subforum
		mockPermission bool
		wantErr        bool
	}{
		{
			name: "UpdateSubforumRuleSuccess",
			input: &apimodels.RuleUpdateInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
				CommunityType: "t",
				SubforumName:  "golang",
				RuleCode:      "no_politics",
				Body: apimodels.RuleUpdateInputBody{
					Name:        stringPtr("Updated No Politics"),
					Description: stringPtr("Updated description"),
					Active:      boolPtr(false),
				},
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: true,
					V:     types.NewJSON[json.RawMessage]([]byte(`[{"code":"no_politics","name":"No Politics","description":"No political discussion","category":"content","severity":"medium","active":true}]`)),
				},
			},
			mockPermission: true,
			wantErr:        false,
		},
		{
			name: "UpdateSubforumRuleRuleNotFound",
			input: &apimodels.RuleUpdateInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
				CommunityType: "t",
				SubforumName:  "golang",
				RuleCode:      "nonexistent_rule",
				Body: apimodels.RuleUpdateInputBody{
					Name: stringPtr("Updated Name"),
				},
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: true,
					V:     types.NewJSON[json.RawMessage]([]byte(`[{"code":"no_politics","name":"No Politics","description":"No political discussion","category":"content","severity":"medium","active":true}]`)),
				},
			},
			mockPermission: true,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock expectations
			if tt.mockSubforum != nil {
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.CommunityType, tt.input.SubforumName).Return(tt.mockSubforum, nil)
				subforumID := int32(1)
				mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "moderator-pseudonym-123", "manage_subforum_rules", &subforumID).Return(tt.mockPermission, nil)
			}

			// Create handler
			handler := &RulesHandler{
				subforumDAO:   mockSubforumDAO,
				permissionDAO: mockPermissionDAO,
			}

			// Execute test
			result, err := handler.UpdateSubforumRule(context.Background(), tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			// Verify mocks
			mockSubforumDAO.AssertExpectations(t)
			mockPermissionDAO.AssertExpectations(t)
		})
	}
}

func TestRulesHandler_DeleteSubforumRule(t *testing.T) {
	t.Skip("TODO: Mock database operations for DeleteSubforumRule test")
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name           string
		input          *apimodels.RuleDeleteInput
		mockSubforum   *dbmodels.Subforum
		mockPermission bool
		wantErr        bool
	}{
		{
			name: "DeleteSubforumRuleSuccess",
			input: &apimodels.RuleDeleteInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
				CommunityType: "t",
				SubforumName:  "golang",
				RuleCode:      "no_politics",
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: true,
					V:     types.NewJSON[json.RawMessage]([]byte(`[{"code":"no_politics","name":"No Politics","description":"No political discussion","category":"content","severity":"medium","active":true}]`)),
				},
			},
			mockPermission: true,
			wantErr:        false,
		},
		{
			name: "DeleteSubforumRuleRuleNotFound",
			input: &apimodels.RuleDeleteInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
				CommunityType: "t",
				SubforumName:  "golang",
				RuleCode:      "nonexistent_rule",
			},
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: true,
					V:     types.NewJSON[json.RawMessage]([]byte(`[{"code":"no_politics","name":"No Politics","description":"No political discussion","category":"content","severity":"medium","active":true}]`)),
				},
			},
			mockPermission: true,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock expectations
			if tt.mockSubforum != nil {
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.input.CommunityType, tt.input.SubforumName).Return(tt.mockSubforum, nil)
				subforumID := int32(1)
				mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "moderator-pseudonym-123", "manage_subforum_rules", &subforumID).Return(tt.mockPermission, nil)
			}

			// Create handler
			handler := &RulesHandler{
				subforumDAO:   mockSubforumDAO,
				permissionDAO: mockPermissionDAO,
			}

			// Execute test
			result, err := handler.DeleteSubforumRule(context.Background(), tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}

			// Verify mocks
			mockSubforumDAO.AssertExpectations(t)
			mockPermissionDAO.AssertExpectations(t)
		})
	}
}

func TestRulesHandler_ReportRuleViolation(t *testing.T) {
	t.Skip("TODO: Mock database operations for ReportRuleViolation test")
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	tests := []struct {
		name         string
		input        *apimodels.RuleViolationInput
		mockReportID int
		wantErr      bool
	}{
		{
			name: "ReportRuleViolationSuccess",
			input: &apimodels.RuleViolationInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
				Body: apimodels.RuleViolationInputBody{
					ContentType:         "post",
					ContentID:           intPtr(123),
					ReportedPseudonymID: "def789ghi012",
					RuleCode:            "harassment",
					RuleType:            "platform",
					ReportDetails:       "This post violates the harassment rule",
				},
			},
			mockReportID: 789,
			wantErr:      false,
		},
		{
			name: "ReportRuleViolationSubforumRule",
			input: &apimodels.RuleViolationInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-id"),
				},
				Body: apimodels.RuleViolationInputBody{
					ContentType:         "comment",
					ContentID:           intPtr(456),
					ReportedPseudonymID: "abc123def456",
					RuleCode:            "no_politics",
					RuleType:            "subforum",
					ReportDetails:       "This comment violates the no politics rule",
				},
			},
			mockReportID: 790,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockReportDAO := mocks.NewMockReportDAO()

			// Set up mock expectations
			mockReportDAO.On("CreateReport", mock.Anything, mock.AnythingOfType("*apimodels.Report")).Return(tt.mockReportID, nil)

			// Create handler
			handler := &RulesHandler{
				reportDAO: mockReportDAO,
			}

			// Execute test
			result, err := handler.ReportRuleViolation(context.Background(), tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.mockReportID, result.ReportID)
			}

			// Verify mocks
			mockReportDAO.AssertExpectations(t)
		})
	}
}

func TestRulesHandler_ForwardReportToPlatform(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	// Create a test user context with forward_reports capability
	testUserCtx := &middleware.UserContext{
		UserID:            1,
		Email:             "moderator@example.com",
		ActivePseudonymID: "test-pseudonym-id",
		DisplayName:       "TestModerator",
		Roles:             []string{"user", "moderator"},
		Capabilities:      []string{"create_content", "vote", "message", "report", "forward_reports"},
		TokenType:         "jwt",
	}

	// Generate a valid JWT token
	token, err := middleware.GenerateJWT(testUserCtx, "test-secret", 24*time.Hour)
	require.NoError(t, err)

	tests := []struct {
		name       string
		input      *apimodels.ReportForwardInput
		mockReport interface{}
		wantErr    bool
	}{
		{
			name: "ForwardReportToPlatformSuccess",
			input: &apimodels.ReportForwardInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + token,
				},
				ReportID: 789,
				Body: apimodels.ReportForwardInputBody{
					ForwardingNotes: "This appears to be a systemic issue",
				},
			},
			mockReport: &dbmodels.Report{
				ReportID:              789,
				ReporterPseudonymID:   "reporter-pseudonym-id",
				ContentType:           "post",
				ContentID:             sql.Null[int64]{V: 123, Valid: true},
				ReportedPseudonymID:   sql.Null[string]{V: "reported-pseudonym-id", Valid: true},
				ReportReason:          "spam",
				ReportDetails:         sql.Null[string]{V: "This post violates community guidelines...", Valid: true},
				Status:                sql.Null[string]{V: "pending", Valid: true},
				CreatedAt:             sql.Null[time.Time]{V: time.Now(), Valid: true},
				ResolvedByUserID:      sql.Null[int64]{Valid: false},
				ResolvedByPseudonymID: sql.Null[string]{Valid: false},
				ResolutionNotes:       sql.Null[string]{Valid: false},
				ResolvedAt:            sql.Null[time.Time]{Valid: false},
				ForwardedToPlatform:   sql.Null[bool]{V: false, Valid: true},
				ForwardingNotes:       sql.Null[string]{Valid: false},
				ForwardedByUserID:     sql.Null[int64]{Valid: false},
				ForwardedAt:           sql.Null[time.Time]{Valid: false},
			},
			wantErr: false,
		},
		{
			name: "ForwardReportToPlatformReportNotFound",
			input: &apimodels.ReportForwardInput{
				AuthInput: middleware.AuthInput{
					Authorization: "Bearer " + token,
				},
				ReportID: 999,
				Body: apimodels.ReportForwardInputBody{
					ForwardingNotes: "This appears to be a systemic issue",
				},
			},
			mockReport: nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockReportDAO := mocks.NewMockReportDAO()

			// Set up mock expectations
			if tt.mockReport != nil {
				mockReportDAO.On("GetReportByID", mock.Anything, int64(tt.input.ReportID)).Return(tt.mockReport, nil)
				mockReportDAO.On("UpdateReport", mock.Anything, int64(tt.input.ReportID), mock.AnythingOfType("*models.ReportSetter")).Return(nil)
			} else {
				mockReportDAO.On("GetReportByID", mock.Anything, int64(tt.input.ReportID)).Return(nil, nil)
			}

			// Create handler
			handler := &RulesHandler{
				reportDAO: mockReportDAO,
			}

			// Execute test
			result, err := handler.ForwardReportToPlatform(context.Background(), tt.input)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.input.ReportID, result.ReportID)
			}

			// Verify mocks
			mockReportDAO.AssertExpectations(t)
		})
	}
}

func TestRulesHandler_validateModeratorPermissionsForSubforum(t *testing.T) {
	tests := []struct {
		name           string
		userCtx        *middleware.UserContext
		communityType  string
		subforumName   string
		mockSubforum   *dbmodels.Subforum
		mockPermission bool
		wantErr        bool
	}{
		{
			name:          "ValidatePermissionsSuccess",
			userCtx:       fixtures.CreateTestUserContext(),
			communityType: "b",
			subforumName:  "hashpost",
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "hashpost",
			},
			mockPermission: true,
			wantErr:        false,
		},
		{
			name:           "ValidatePermissionsSubforumNotFound",
			userCtx:        fixtures.CreateTestUserContext(),
			communityType:  "b",
			subforumName:   "nonexistent",
			mockSubforum:   nil,
			mockPermission: false,
			wantErr:        true,
		},
		{
			name:          "ValidatePermissionsInsufficientPermissions",
			userCtx:       fixtures.CreateTestUserContext(),
			communityType: "b",
			subforumName:  "hashpost",
			mockSubforum: &dbmodels.Subforum{
				SubforumID: 1,
				Name:       "hashpost",
			},
			mockPermission: false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock DAOs
			mockSubforumDAO := mocks.NewMockSubforumDAO()
			mockPermissionDAO := mocks.NewMockPermissionDAO()

			// Set up mock expectations
			if tt.mockSubforum != nil {
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.communityType, tt.subforumName).Return(tt.mockSubforum, nil)
				subforumID := int32(1)
				mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, tt.userCtx.UserID, tt.userCtx.ActivePseudonymID, "manage_subforum_rules", &subforumID).Return(tt.mockPermission, nil)
			} else {
				mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, tt.communityType, tt.subforumName).Return(nil, nil)
			}

			// Create handler
			handler := &RulesHandler{
				subforumDAO:   mockSubforumDAO,
				permissionDAO: mockPermissionDAO,
			}

			// Execute test
			err := handler.validateModeratorPermissionsForSubforum(context.Background(), tt.userCtx, tt.communityType, tt.subforumName)

			// Assertions
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			// Verify mocks
			mockSubforumDAO.AssertExpectations(t)
			mockPermissionDAO.AssertExpectations(t)
		})
	}
}

// Helper functions for creating pointers to primitive types
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
