package handlers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/stephenafamo/bob/types"
	gomock "go.uber.org/mock/gomock"
)

// TestNewRulesHandler_Gomock tests the rules handler constructor using gomock
func TestNewRulesHandler_Gomock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock DAOs
	mockReportDAO := dao.NewMockReportDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockSystemSettingsDAO := dao.NewMockSystemSettingsDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)

	// Create handler
	handler := handlers.NewRulesHandler(mockReportDAO, mockSubforumDAO, mockSystemSettingsDAO, mockPermissionDAO, mockPseudonymDAO, nil)

	// Assertions
	assert.NotNil(t, handler)
	// Note: Fields are unexported, so we can't access them directly in tests
	// The constructor test verifies the handler was created successfully
}

// TestRulesHandler_GetPlatformRules_Gomock tests the platform rules retrieval using gomock
func TestRulesHandler_GetPlatformRules_Gomock(t *testing.T) {
	t.Run("GetPlatformRulesSuccess", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock DAOs
		mockSystemSettingsDAO := dao.NewMockSystemSettingsDAOInterface(ctrl)

		// Set up mock expectations
		mockSetting := &models.SystemSetting{
			SettingKey:   "platform_rules",
			SettingValue: `[{"code":"no_spam","name":"No Spam","description":"No spam allowed","category":"content","severity":"high","active":true}]`,
			SettingType:  "json",
		}

		mockSystemSettingsDAO.EXPECT().
			GetSetting(gomock.Any(), "platform_rules").
			Return(mockSetting, nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, nil, mockSystemSettingsDAO, nil, nil, nil)

		// Test input
		input := &apimodels.PlatformRulesInput{
			ActiveOnly: false,
		}

		// Execute test
		result, err := handler.GetPlatformRules(context.Background(), input)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 200, result.Status)
	})

	t.Run("GetPlatformRulesNoRules", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock DAOs
		mockSystemSettingsDAO := dao.NewMockSystemSettingsDAOInterface(ctrl)

		// Set up mock expectations
		mockSystemSettingsDAO.EXPECT().
			GetSetting(gomock.Any(), "platform_rules").
			Return(nil, nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, nil, mockSystemSettingsDAO, nil, nil, nil)

		// Test input
		input := &apimodels.PlatformRulesInput{}

		// Execute test
		result, err := handler.GetPlatformRules(context.Background(), input)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 200, result.Status)
	})

	t.Run("GetPlatformRulesActiveOnly", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock DAOs
		mockSystemSettingsDAO := dao.NewMockSystemSettingsDAOInterface(ctrl)

		// Set up mock expectations
		mockSetting := &models.SystemSetting{
			SettingKey:   "platform_rules",
			SettingValue: `[{"code":"no_spam","name":"No Spam","description":"No spam allowed","category":"content","severity":"high","active":true},{"code":"old_rule","name":"Old Rule","description":"Inactive rule","category":"content","severity":"low","active":false}]`,
			SettingType:  "json",
		}

		mockSystemSettingsDAO.EXPECT().
			GetSetting(gomock.Any(), "platform_rules").
			Return(mockSetting, nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, nil, mockSystemSettingsDAO, nil, nil, nil)

		// Test input
		input := &apimodels.PlatformRulesInput{
			ActiveOnly: true,
		}

		// Execute test
		result, err := handler.GetPlatformRules(context.Background(), input)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 200, result.Status)
	})
}

// TestRulesHandler_GetSubforumRules_Gomock tests the subforum rules retrieval using gomock
func TestRulesHandler_GetSubforumRules_Gomock(t *testing.T) {
	t.Run("GetSubforumRulesSuccess", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock DAOs
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

		// Set up mock expectations
		mockSubforum := &models.Subforum{
			SubforumID: 1,
			Name:       "golang",
			SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
				Valid: true,
				V:     types.NewJSON[json.RawMessage]([]byte(`[{"code":"no_politics","name":"No Politics","description":"No political discussion","category":"content","severity":"medium","active":true}]`)),
			},
		}

		mockSubforumDAO.EXPECT().
			GetSubforumByName(gomock.Any(), "golang").
			Return(mockSubforum, nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, mockSubforumDAO, nil, nil, nil, nil)

		// Test input
		input := &apimodels.SubforumRulesInput{
			CommunityType: constants.CommunityTypeTopical,
			SubforumName:  "golang",
			ActiveOnly:    false,
		}

		// Execute test
		result, err := handler.GetSubforumRules(context.Background(), input)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 200, result.Status)
	})

	t.Run("GetSubforumRulesNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock DAOs
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

		// Set up mock expectations
		mockSubforumDAO.EXPECT().
			GetSubforumByName(gomock.Any(), "nonexistent").
			Return(nil, nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, mockSubforumDAO, nil, nil, nil, nil)

		// Test input
		input := &apimodels.SubforumRulesInput{
			CommunityType: constants.CommunityTypeTopical,
			SubforumName:  "nonexistent",
			ActiveOnly:    false,
		}

		// Execute test
		result, err := handler.GetSubforumRules(context.Background(), input)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("GetSubforumRulesNoRules", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create mock DAOs
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)

		// Set up mock expectations
		mockSubforum := &models.Subforum{
			SubforumID: 1,
			Name:       "golang",
			SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
				Valid: false,
				V:     types.NewJSON[json.RawMessage]([]byte(`{}`)),
			},
		}

		mockSubforumDAO.EXPECT().
			GetSubforumByName(gomock.Any(), "golang").
			Return(mockSubforum, nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, mockSubforumDAO, nil, nil, nil, nil)

		// Test input
		input := &apimodels.SubforumRulesInput{
			CommunityType: constants.CommunityTypeTopical,
			SubforumName:  "golang",
			ActiveOnly:    false,
		}

		// Execute test
		result, err := handler.GetSubforumRules(context.Background(), input)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 200, result.Status)
	})
}

// TestRulesHandler_CreateSubforumRule_Gomock tests the subforum rule creation using gomock
func TestRulesHandler_CreateSubforumRule_Gomock(t *testing.T) {
	t.Run("CreateSubforumRuleSuccess", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Initialize global auth middleware for testing
		authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
		middleware.SetGlobalAuthMiddleware(authMiddleware)

		// Create mock DAOs
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)

		// Test data
		userID := int64(1)
		pseudonymID := "test-pseudonym-id"
		subforumID := int32(1)

		// Set up mock expectations
		mockSubforumDAO.EXPECT().
			GetSubforumByCommunityTypeAndName(gomock.Any(), constants.CommunityTypeTopical, "golang").
			Return(&models.Subforum{
				SubforumID: subforumID,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: false,
					V:     types.NewJSON[json.RawMessage]([]byte(`{}`)),
				},
			}, nil).
			Times(2) // Called once in validateModeratorPermissionsForSubforum and once in main method

		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(gomock.Any(), userID, pseudonymID, constants.CapabilityManageSubforumRules, &subforumID).
			Return(true, nil).
			Times(1)

		mockSubforumDAO.EXPECT().
			UpdateRules(gomock.Any(), subforumID, gomock.Any()).
			Return(nil).
			Times(1)

		mockPseudonymDAO.EXPECT().
			UpdateLastActive(gomock.Any(), pseudonymID).
			Return(nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, mockSubforumDAO, nil, mockPermissionDAO, mockPseudonymDAO, nil)

		// Set up proper authentication
		userCtx := fixtures.CreateTestUserContext()
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, tokenErr)

		// Test input
		input := &apimodels.RuleCreateInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
				AccessToken:   "",
			},
			CommunityType: constants.CommunityTypeTopical,
			SubforumName:  "golang",
			Body: apimodels.RuleCreateInputBody{
				Code:        "no_politics",
				Name:        "No Politics",
				Description: "No political discussion allowed",
				Category:    "content",
				Severity:    "medium",
				Active:      true,
			},
		}

		// Execute test
		result, err := handler.CreateSubforumRule(context.Background(), input)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("CreateSubforumRuleInsufficientPermissions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Initialize global auth middleware for testing
		authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
		middleware.SetGlobalAuthMiddleware(authMiddleware)

		// Create mock DAOs
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)

		// Test data
		userID := int64(1)
		pseudonymID := "test-pseudonym-id"
		subforumID := int32(1)

		// Set up mock expectations
		mockSubforumDAO.EXPECT().
			GetSubforumByCommunityTypeAndName(gomock.Any(), constants.CommunityTypeTopical, "golang").
			Return(&models.Subforum{
				SubforumID: subforumID,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: false,
					V:     types.NewJSON[json.RawMessage]([]byte(`{}`)),
				},
			}, nil).
			Times(1) // Called once in validateModeratorPermissionsForSubforum, main method never reached

		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(gomock.Any(), userID, pseudonymID, constants.CapabilityManageSubforumRules, &subforumID).
			Return(false, nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, mockSubforumDAO, nil, mockPermissionDAO, mockPseudonymDAO, nil)

		// Set up proper authentication
		userCtx := fixtures.CreateTestUserContext()
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, tokenErr)

		// Test input
		input := &apimodels.RuleCreateInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
				AccessToken:   "",
			},
			CommunityType: constants.CommunityTypeTopical,
			SubforumName:  "golang",
			Body: apimodels.RuleCreateInputBody{
				Code:        "no_politics",
				Name:        "No Politics",
				Description: "No political discussion allowed",
				Category:    "content",
				Severity:    "medium",
				Active:      true,
			},
		}

		// Execute test
		result, err := handler.CreateSubforumRule(context.Background(), input)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestRulesHandler_UpdateSubforumRule_Gomock tests the subforum rule update using gomock
func TestRulesHandler_UpdateSubforumRule_Gomock(t *testing.T) {
	t.Run("UpdateSubforumRuleSuccess", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Initialize global auth middleware for testing
		authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
		middleware.SetGlobalAuthMiddleware(authMiddleware)

		// Create mock DAOs
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)

		// Test data
		userID := int64(1)
		pseudonymID := "test-pseudonym-id"
		subforumID := int32(1)

		// Set up mock expectations
		mockSubforumDAO.EXPECT().
			GetSubforumByCommunityTypeAndName(gomock.Any(), constants.CommunityTypeTopical, "golang").
			Return(&models.Subforum{
				SubforumID: subforumID,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: true,
					V:     types.NewJSON[json.RawMessage]([]byte(`[{"code":"no_politics","name":"No Politics","description":"No political discussion","category":"content","severity":"medium","active":true}]`)),
				},
			}, nil).
			Times(2) // Called once in validateModeratorPermissionsForSubforum and once in main method

		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(gomock.Any(), userID, pseudonymID, constants.CapabilityManageSubforumRules, &subforumID).
			Return(true, nil).
			Times(1)

		mockSubforumDAO.EXPECT().
			UpdateRules(gomock.Any(), subforumID, gomock.Any()).
			Return(nil).
			Times(1)

		mockPseudonymDAO.EXPECT().
			UpdateLastActive(gomock.Any(), pseudonymID).
			Return(nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, mockSubforumDAO, nil, mockPermissionDAO, mockPseudonymDAO, nil)

		// Set up proper authentication
		userCtx := fixtures.CreateTestUserContext()
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, tokenErr)

		// Test input
		input := &apimodels.RuleUpdateInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
				AccessToken:   "",
			},
			CommunityType: constants.CommunityTypeTopical,
			SubforumName:  "golang",
			RuleCode:      "no_politics",
			Body: apimodels.RuleUpdateInputBody{
				Name:        stringPtr("Updated No Politics"),
				Description: stringPtr("Updated description"),
				Active:      boolPtr(false),
			},
		}

		// Execute test
		result, err := handler.UpdateSubforumRule(context.Background(), input)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("UpdateSubforumRuleRuleNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Initialize global auth middleware for testing
		authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
		middleware.SetGlobalAuthMiddleware(authMiddleware)

		// Create mock DAOs
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)

		// Test data
		userID := int64(1)
		pseudonymID := "test-pseudonym-id"
		subforumID := int32(1)

		// Set up mock expectations
		mockSubforumDAO.EXPECT().
			GetSubforumByCommunityTypeAndName(gomock.Any(), constants.CommunityTypeTopical, "golang").
			Return(&models.Subforum{
				SubforumID: subforumID,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: true,
					V:     types.NewJSON[json.RawMessage]([]byte(`[{"code":"no_politics","name":"No Politics","description":"No political discussion","category":"content","severity":"medium","active":true}]`)),
				},
			}, nil).
			Times(2) // Called once in validateModeratorPermissionsForSubforum and once in main method

		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(gomock.Any(), userID, pseudonymID, constants.CapabilityManageSubforumRules, &subforumID).
			Return(true, nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, mockSubforumDAO, nil, mockPermissionDAO, mockPseudonymDAO, nil)

		// Set up proper authentication
		userCtx := fixtures.CreateTestUserContext()
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, tokenErr)

		// Test input
		input := &apimodels.RuleUpdateInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
				AccessToken:   "",
			},
			CommunityType: constants.CommunityTypeTopical,
			SubforumName:  "golang",
			RuleCode:      "nonexistent_rule",
			Body: apimodels.RuleUpdateInputBody{
				Name: stringPtr("Updated Name"),
			},
		}

		// Execute test
		result, err := handler.UpdateSubforumRule(context.Background(), input)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// TestRulesHandler_DeleteSubforumRule_Gomock tests the subforum rule deletion using gomock
func TestRulesHandler_DeleteSubforumRule_Gomock(t *testing.T) {
	t.Run("DeleteSubforumRuleSuccess", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Initialize global auth middleware for testing
		authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
		middleware.SetGlobalAuthMiddleware(authMiddleware)

		// Create mock DAOs
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)

		// Test data
		userID := int64(1)
		pseudonymID := "test-pseudonym-id"
		subforumID := int32(1)

		// Set up mock expectations
		mockSubforumDAO.EXPECT().
			GetSubforumByCommunityTypeAndName(gomock.Any(), constants.CommunityTypeTopical, "golang").
			Return(&models.Subforum{
				SubforumID: subforumID,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: true,
					V:     types.NewJSON[json.RawMessage]([]byte(`[{"code":"no_politics","name":"No Politics","description":"No political discussion","category":"content","severity":"medium","active":true}]`)),
				},
			}, nil).
			Times(2) // Called once in validateModeratorPermissionsForSubforum and once in main method

		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(gomock.Any(), userID, pseudonymID, constants.CapabilityManageSubforumRules, &subforumID).
			Return(true, nil).
			Times(1)

		mockSubforumDAO.EXPECT().
			UpdateRules(gomock.Any(), subforumID, gomock.Any()).
			Return(nil).
			Times(1)

		mockPseudonymDAO.EXPECT().
			UpdateLastActive(gomock.Any(), pseudonymID).
			Return(nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, mockSubforumDAO, nil, mockPermissionDAO, mockPseudonymDAO, nil)

		// Set up proper authentication
		userCtx := fixtures.CreateTestUserContext()
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, tokenErr)

		// Test input
		input := &apimodels.RuleDeleteInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
				AccessToken:   "",
			},
			CommunityType: constants.CommunityTypeTopical,
			SubforumName:  "golang",
			RuleCode:      "no_politics",
		}

		// Execute test
		result, err := handler.DeleteSubforumRule(context.Background(), input)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("DeleteSubforumRuleRuleNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Initialize global auth middleware for testing
		authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
		middleware.SetGlobalAuthMiddleware(authMiddleware)

		// Create mock DAOs
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)

		// Test data
		userID := int64(1)
		pseudonymID := "test-pseudonym-id"
		subforumID := int32(1)

		// Set up mock expectations
		mockSubforumDAO.EXPECT().
			GetSubforumByCommunityTypeAndName(gomock.Any(), constants.CommunityTypeTopical, "golang").
			Return(&models.Subforum{
				SubforumID: subforumID,
				Name:       "golang",
				SubforumRules: sql.Null[types.JSON[json.RawMessage]]{
					Valid: true,
					V:     types.NewJSON[json.RawMessage]([]byte(`[{"code":"no_politics","name":"No Politics","description":"No political discussion","category":"content","severity":"medium","active":true}]`)),
				},
			}, nil).
			Times(2) // Called once in validateModeratorPermissionsForSubforum and once in main method

		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(gomock.Any(), userID, pseudonymID, constants.CapabilityManageSubforumRules, &subforumID).
			Return(true, nil).
			Times(1)

		// Create handler
		handler := handlers.NewRulesHandler(nil, mockSubforumDAO, nil, mockPermissionDAO, mockPseudonymDAO, nil)

		// Set up proper authentication
		userCtx := fixtures.CreateTestUserContext()
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, tokenErr)

		// Test input
		input := &apimodels.RuleDeleteInput{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
				AccessToken:   "",
			},
			CommunityType: constants.CommunityTypeTopical,
			SubforumName:  "golang",
			RuleCode:      "nonexistent_rule",
		}

		// Execute test
		result, err := handler.DeleteSubforumRule(context.Background(), input)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

// Helper functions for pointer types
func stringPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
