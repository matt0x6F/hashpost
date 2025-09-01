package handlers_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// TestAuthHandler_GetCurrentUserSessionForSubforum tests the subforum-specific session functionality using gomock
func TestAuthHandler_GetCurrentUserSessionForSubforum(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, &config.JWTConfig{
		Secret:      "test-secret",
		Expiration:  time.Hour,
		Development: true,
	}, &config.SecurityConfig{})
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("SuccessWithSubforumCapabilities", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler with mocks
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
			Security: config.SecurityConfig{
				PasswordValidation: config.PasswordValidationConfig{
					MinLength:          8,
					RequireUppercase:   true,
					RequireLowercase:   true,
					RequireDigit:       true,
					RequireSpecialChar: true,
				},
			},
			Email: config.EmailConfig{
				Validation: config.EmailValidationConfig{
					Enabled:       true,
					VerifierEmail: "noreply@example.com",
				},
			},
		}

		mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, mockPseudonymDAO, nil, mockRoleKeyDAO, ibeSystem, mockSubforumDAO, mockPermissionDAO, nil, nil, nil)

		// Test data
		testUserID := int64(1)
		testEmail := "test@example.com"
		testPseudonymID := "pseudonym-123"
		testDisplayName := "TestUser"
		testSubforumName := "test-subforum"
		testSubforumID := int64(100)

		// Mock user lookup
		mockUser := &dbmodels.User{
			UserID:   testUserID,
			Email:    testEmail,
			IsActive: sql.Null[bool]{V: true, Valid: true},
		}
		mockUserDAO.EXPECT().GetUserByID(gomock.Any(), testUserID).Return(mockUser, nil).Times(1)
		mockUserDAO.EXPECT().UpdateLastActive(gomock.Any(), testUserID).Return(nil).Times(1)

		// Mock subforum lookup
		mockSubforum := &dbmodels.Subforum{
			SubforumID: int32(testSubforumID),
			Name:       testSubforumName,
		}
		mockSubforumDAO.EXPECT().GetSubforumByCommunityTypeAndName(gomock.Any(), "h", testSubforumName).Return(mockSubforum, nil).Times(1)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.EXPECT().GetPseudonymsByUserID(gomock.Any(), testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil).Times(1)

		// Mock role keys for the active pseudonym with platform capabilities
		mockRoleKey := &dbmodels.RoleKey{
			PseudonymID:  testPseudonymID,
			RoleName:     "user",
			Capabilities: types.NewJSON[json.RawMessage]([]byte(`["create_content", "vote", "message", "report", "create_subforum"]`)),
		}
		mockRoleKeyDAO.EXPECT().ListRoleKeysByPseudonym(gomock.Any(), testPseudonymID).Return([]*dbmodels.RoleKey{mockRoleKey}, nil).Times(1)

		// Mock unified roles and capabilities for active pseudonym with subforum context
		subforumID := int32(testSubforumID)
		mockPermissionDAO.EXPECT().GetUnifiedActivePseudonymRolesAndCapabilities(gomock.Any(), testUserID, testPseudonymID, &subforumID).Return(
			[]string{constants.RoleUser, constants.RoleModerator},
			[]string{"create_content", "vote", "message", "report", "create_subforum", "moderate_content", "ban_users"},
			nil).Times(1)

		// Create input with valid JWT token
		userCtx := &middleware.UserContext{
			UserID:            testUserID,
			Email:             testEmail,
			MFAEnabled:        false, // roles and capabilities deprecated
			ActivePseudonymID: testPseudonymID,
			DisplayName:       testDisplayName,
		}

		// Generate a valid JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, err)

		input := &struct {
			middleware.AuthInput
			SubforumName string `path:"subforum_name"`
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SubforumName: testSubforumName,
		}

		// Call the handler
		response, err := handler.GetCurrentUserSessionForSubforum(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, testEmail, response.Body.Email)
		assert.Equal(t, int(testUserID), response.Body.UserID)
		assert.Equal(t, testPseudonymID, response.Body.ActivePseudonymID)
		assert.Equal(t, testDisplayName, response.Body.DisplayName)
		assert.Len(t, response.Body.Capabilities, 7) // Platform capabilities + subforum capabilities (moderate_content + ban_users)
		assert.Contains(t, response.Body.Capabilities, "moderate_content")
		assert.Contains(t, response.Body.Capabilities, "ban_users")
		// Note: manage_moderators is not added because the mock returns false for it
	})

	t.Run("SuccessWithSubforumModeratorCapabilities", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler with mocks
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
			Security: config.SecurityConfig{
				PasswordValidation: config.PasswordValidationConfig{
					MinLength:          8,
					RequireUppercase:   true,
					RequireLowercase:   true,
					RequireDigit:       true,
					RequireSpecialChar: true,
				},
			},
			Email: config.EmailConfig{
				Validation: config.EmailValidationConfig{
					Enabled:       true,
					VerifierEmail: "noreply@example.com",
				},
			},
		}

		mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, mockPseudonymDAO, nil, mockRoleKeyDAO, ibeSystem, mockSubforumDAO, mockPermissionDAO, nil, nil, nil)

		// Test data
		testUserID := int64(1)
		testEmail := "moderator@example.com"
		testPseudonymID := "moderator-pseudonym-123"
		testDisplayName := "ModeratorUser"
		testSubforumName := "moderated-subforum"
		testSubforumID := int64(200)

		// Mock user lookup
		mockUser := &dbmodels.User{
			UserID:   testUserID,
			Email:    testEmail,
			IsActive: sql.Null[bool]{V: true, Valid: true},
		}
		mockUserDAO.EXPECT().GetUserByID(gomock.Any(), testUserID).Return(mockUser, nil).Times(1)
		mockUserDAO.EXPECT().UpdateLastActive(gomock.Any(), testUserID).Return(nil).Times(1)

		// Mock subforum lookup
		mockSubforum := &dbmodels.Subforum{
			SubforumID: int32(testSubforumID),
			Name:       testSubforumName,
		}
		mockSubforumDAO.EXPECT().GetSubforumByCommunityTypeAndName(gomock.Any(), "h", testSubforumName).Return(mockSubforum, nil).Times(1)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.EXPECT().GetPseudonymsByUserID(gomock.Any(), testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil).Times(1)

		// Mock role keys for the active pseudonym with platform capabilities
		mockRoleKey := &dbmodels.RoleKey{
			PseudonymID:  testPseudonymID,
			RoleName:     "user",
			Capabilities: types.NewJSON[json.RawMessage]([]byte(`["create_content", "vote", "message", "report", "create_subforum"]`)),
		}
		mockRoleKeyDAO.EXPECT().ListRoleKeysByPseudonym(gomock.Any(), testPseudonymID).Return([]*dbmodels.RoleKey{mockRoleKey}, nil).Times(1)

		// Mock unified roles and capabilities for active pseudonym with subforum context
		subforumID := int32(testSubforumID)
		mockPermissionDAO.EXPECT().GetUnifiedActivePseudonymRolesAndCapabilities(gomock.Any(), testUserID, testPseudonymID, &subforumID).Return(
			[]string{constants.RoleUser, constants.RoleModerator},
			[]string{"create_content", "vote", "message", "report", "create_subforum", "moderate_content", "ban_users"},
			nil).Times(1)

		// Create input with valid JWT token
		userCtx := &middleware.UserContext{
			UserID:            testUserID,
			Email:             testEmail,
			MFAEnabled:        false, // roles and capabilities deprecated
			ActivePseudonymID: testPseudonymID,
			DisplayName:       testDisplayName,
		}

		// Generate a valid JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, err)

		input := &struct {
			middleware.AuthInput
			SubforumName string `path:"subforum_name"`
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SubforumName: testSubforumName,
		}

		// Call the handler
		response, err := handler.GetCurrentUserSessionForSubforum(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, testEmail, response.Body.Email)
		assert.Equal(t, int(testUserID), response.Body.UserID)
		assert.Equal(t, testPseudonymID, response.Body.ActivePseudonymID)
		assert.Equal(t, testDisplayName, response.Body.DisplayName)
		assert.Len(t, response.Body.Capabilities, 7) // Platform capabilities + subforum capabilities (moderate_content + ban_users)
		assert.Contains(t, response.Body.Capabilities, "moderate_content")
		assert.Contains(t, response.Body.Capabilities, "ban_users")
		// Note: manage_moderators is not added because the mock returns false for it
	})

	t.Run("UnauthorizedAccess", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler with mocks
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
			Security: config.SecurityConfig{
				PasswordValidation: config.PasswordValidationConfig{
					MinLength:          8,
					RequireUppercase:   true,
					RequireLowercase:   true,
					RequireDigit:       true,
					RequireSpecialChar: true,
				},
			},
			Email: config.EmailConfig{
				Validation: config.EmailValidationConfig{
					Enabled:       true,
					VerifierEmail: "noreply@example.com",
				},
			},
		}

		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})
		handler := handlers.NewAuthHandler(cfg, nil, nil, nil, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

		// Create input without JWT token
		input := &struct {
			middleware.AuthInput
			SubforumName string `path:"subforum_name"`
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "",
			},
			SubforumName: "test-subforum",
		}

		// Call the handler
		response, err := handler.GetCurrentUserSessionForSubforum(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "Authentication required")
	})

	t.Run("InvalidToken", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler with mocks
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
			Security: config.SecurityConfig{
				PasswordValidation: config.PasswordValidationConfig{
					MinLength:          8,
					RequireUppercase:   true,
					RequireLowercase:   true,
					RequireDigit:       true,
					RequireSpecialChar: true,
				},
			},
			Email: config.EmailConfig{
				Validation: config.EmailValidationConfig{
					Enabled:       true,
					VerifierEmail: "noreply@example.com",
				},
			},
		}

		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})
		handler := handlers.NewAuthHandler(cfg, nil, nil, nil, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

		// Create input with invalid JWT token
		input := &struct {
			middleware.AuthInput
			SubforumName string `path:"subforum_name"`
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer invalid_token",
			},
			SubforumName: "test-subforum",
		}

		// Call the handler
		response, err := handler.GetCurrentUserSessionForSubforum(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "Authentication required")
	})

	t.Run("UserNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler with mocks
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
			Security: config.SecurityConfig{
				PasswordValidation: config.PasswordValidationConfig{
					MinLength:          8,
					RequireUppercase:   true,
					RequireLowercase:   true,
					RequireDigit:       true,
					RequireSpecialChar: true,
				},
			},
			Email: config.EmailConfig{
				Validation: config.EmailValidationConfig{
					Enabled:       true,
					VerifierEmail: "noreply@example.com",
				},
			},
		}

		mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, mockPseudonymDAO, nil, mockRoleKeyDAO, ibeSystem, mockSubforumDAO, mockPermissionDAO, nil, nil, nil)

		// Test data
		testUserID := int64(999)
		testEmail := "nonexistent@example.com"
		testPseudonymID := "pseudonym-123"
		testDisplayName := "TestUser"

		// Mock user not found
		mockUserDAO.EXPECT().GetUserByID(gomock.Any(), testUserID).Return((*dbmodels.User)(nil), nil).Times(1)

		// Create input with valid JWT token
		userCtx := &middleware.UserContext{
			UserID:            testUserID,
			Email:             testEmail,
			MFAEnabled:        false, // roles and capabilities deprecated
			ActivePseudonymID: testPseudonymID,
			DisplayName:       testDisplayName,
		}

		// Generate a valid JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, err)

		input := &struct {
			middleware.AuthInput
			SubforumName string `path:"subforum_name"`
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SubforumName: "test-subforum",
		}

		// Call the handler
		response, err := handler.GetCurrentUserSessionForSubforum(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "User not found")
	})

	t.Run("InactiveUser", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler with mocks
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
			Security: config.SecurityConfig{
				PasswordValidation: config.PasswordValidationConfig{
					MinLength:          8,
					RequireUppercase:   true,
					RequireLowercase:   true,
					RequireDigit:       true,
					RequireSpecialChar: true,
				},
			},
			Email: config.EmailConfig{
				Validation: config.EmailValidationConfig{
					Enabled:       true,
					VerifierEmail: "noreply@example.com",
				},
			},
		}

		mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, mockPseudonymDAO, nil, mockRoleKeyDAO, ibeSystem, mockSubforumDAO, mockPermissionDAO, nil, nil, nil)

		// Test data
		testUserID := int64(1)
		testEmail := "inactive@example.com"
		testPseudonymID := "pseudonym-123"
		testDisplayName := "InactiveUser"

		// Mock inactive user
		mockUser := &dbmodels.User{
			UserID:   testUserID,
			Email:    testEmail,
			IsActive: sql.Null[bool]{V: false, Valid: true},
		}
		mockUserDAO.EXPECT().GetUserByID(gomock.Any(), testUserID).Return(mockUser, nil).Times(1)

		// Create input with valid JWT token
		userCtx := &middleware.UserContext{
			UserID:            testUserID,
			Email:             testEmail,
			MFAEnabled:        false, // roles and capabilities deprecated
			ActivePseudonymID: testPseudonymID,
			DisplayName:       testDisplayName,
		}

		// Generate a valid JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, err)

		input := &struct {
			middleware.AuthInput
			SubforumName string `path:"subforum_name"`
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SubforumName: "test-subforum",
		}

		// Call the handler
		response, err := handler.GetCurrentUserSessionForSubforum(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "Account inactive")
	})

	t.Run("SuspendedUser", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler with mocks
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
			Security: config.SecurityConfig{
				PasswordValidation: config.PasswordValidationConfig{
					MinLength:          8,
					RequireUppercase:   true,
					RequireLowercase:   true,
					RequireDigit:       true,
					RequireSpecialChar: true,
				},
			},
			Email: config.EmailConfig{
				Validation: config.EmailValidationConfig{
					Enabled:       true,
					VerifierEmail: "noreply@example.com",
				},
			},
		}

		mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, mockPseudonymDAO, nil, mockRoleKeyDAO, ibeSystem, mockSubforumDAO, mockPermissionDAO, nil, nil, nil)

		// Test data
		testUserID := int64(1)
		testEmail := "suspended@example.com"
		testPseudonymID := "pseudonym-123"
		testDisplayName := "SuspendedUser"

		// Mock suspended user
		mockUser := &dbmodels.User{
			UserID:      testUserID,
			Email:       testEmail,
			IsActive:    sql.Null[bool]{V: true, Valid: true},
			IsSuspended: sql.Null[bool]{V: true, Valid: true},
		}
		mockUserDAO.EXPECT().GetUserByID(gomock.Any(), testUserID).Return(mockUser, nil).Times(1)

		// Create input with valid JWT token
		userCtx := &middleware.UserContext{
			UserID:            testUserID,
			Email:             testEmail,
			MFAEnabled:        false, // roles and capabilities deprecated
			ActivePseudonymID: testPseudonymID,
			DisplayName:       testDisplayName,
		}

		// Generate a valid JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, err)

		input := &struct {
			middleware.AuthInput
			SubforumName string `path:"subforum_name"`
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SubforumName: "test-subforum",
		}

		// Call the handler
		response, err := handler.GetCurrentUserSessionForSubforum(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "Account suspended")
	})

	t.Run("SubforumNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler with mocks
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
			Security: config.SecurityConfig{
				PasswordValidation: config.PasswordValidationConfig{
					MinLength:          8,
					RequireUppercase:   true,
					RequireLowercase:   true,
					RequireDigit:       true,
					RequireSpecialChar: true,
				},
			},
			Email: config.EmailConfig{
				Validation: config.EmailValidationConfig{
					Enabled:       true,
					VerifierEmail: "noreply@example.com",
				},
			},
		}

		mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
		mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, mockPseudonymDAO, nil, mockRoleKeyDAO, ibeSystem, mockSubforumDAO, mockPermissionDAO, nil, nil, nil)

		// Test data
		testUserID := int64(1)
		testEmail := "test@example.com"
		testPseudonymID := "pseudonym-123"
		testDisplayName := "TestUser"
		testSubforumName := "nonexistent-subforum"

		// Mock user lookup
		mockUser := &dbmodels.User{
			UserID:   testUserID,
			Email:    testEmail,
			IsActive: sql.Null[bool]{V: true, Valid: true},
		}
		mockUserDAO.EXPECT().GetUserByID(gomock.Any(), testUserID).Return(mockUser, nil).Times(1)
		mockUserDAO.EXPECT().UpdateLastActive(gomock.Any(), testUserID).Return(nil).Times(1)

		// Mock subforum not found
		mockSubforumDAO.EXPECT().GetSubforumByCommunityTypeAndName(gomock.Any(), "h", testSubforumName).Return(nil, nil).Times(1)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.EXPECT().GetPseudonymsByUserID(gomock.Any(), testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil).Times(1)

		// Mock role keys for the active pseudonym with platform capabilities
		mockRoleKey := &dbmodels.RoleKey{
			PseudonymID:  testPseudonymID,
			RoleName:     "user",
			Capabilities: types.NewJSON[json.RawMessage]([]byte(`["create_content", "vote", "message", "report", "create_subforum"]`)),
		}
		mockRoleKeyDAO.EXPECT().ListRoleKeysByPseudonym(gomock.Any(), testPseudonymID).Return([]*dbmodels.RoleKey{mockRoleKey}, nil).Times(1)

		// Create input with valid JWT token
		userCtx := &middleware.UserContext{
			UserID:            testUserID,
			Email:             testEmail,
			MFAEnabled:        false, // roles and capabilities deprecated
			ActivePseudonymID: testPseudonymID,
			DisplayName:       testDisplayName,
		}

		// Generate a valid JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, err)

		input := &struct {
			middleware.AuthInput
			SubforumName string `path:"subforum_name"`
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SubforumName: testSubforumName,
		}

		// Call the handler
		response, err := handler.GetCurrentUserSessionForSubforum(context.Background(), input)

		// Assert response - should succeed but with no subforum capabilities
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, testEmail, response.Body.Email)
		assert.Equal(t, int(testUserID), response.Body.UserID)
		assert.Equal(t, testPseudonymID, response.Body.ActivePseudonymID)
		assert.Equal(t, testDisplayName, response.Body.DisplayName)
		// Should have only global capabilities when subforum is not found
		assert.ElementsMatch(t, []string{"create_content", "vote", "message", "report", "create_subforum"}, response.Body.Capabilities)
	})
}
