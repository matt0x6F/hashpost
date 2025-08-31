package handlers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// TestAuthHandler_CurrentUserSession_Gomock tests the current user session functionality using gomock
func TestAuthHandler_CurrentUserSession_Gomock(t *testing.T) {
	t.Run("GetCurrentUserSession", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Initialize global auth middleware for testing
		middleware.SetGlobalAuthMiddleware(middleware.NewAuthMiddleware("test-secret", nil, &config.JWTConfig{
			Secret:      "test-secret",
			Expiration:  time.Hour,
			Development: true,
		}, &config.SecurityConfig{}))

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
		mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, mockPseudonymDAO, nil, mockRoleKeyDAO, ibeSystem, nil, mockPermissionDAO, nil, nil, nil)

		// Test data
		testUserID := int64(1)
		testEmail := "test@example.com"
		testPseudonymID := "pseudonym-123"
		testDisplayName := "TestUser"

		// Mock user lookup
		mockUser := &dbmodels.User{
			UserID:   testUserID,
			Email:    testEmail,
			IsActive: sql.Null[bool]{V: true, Valid: true},
		}
		mockUserDAO.EXPECT().GetUserByID(gomock.Any(), testUserID).Return(mockUser, nil).Times(1)
		mockUserDAO.EXPECT().UpdateLastActive(gomock.Any(), testUserID).Return(nil).Times(1)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.EXPECT().GetPseudonymsByUserID(gomock.Any(), testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil).Times(1)

		// Mock role keys for the default pseudonym
		mockRoleKeyDAO.EXPECT().ListRoleKeysByPseudonym(gomock.Any(), testPseudonymID).Return([]*dbmodels.RoleKey{}, nil).Times(1)

		// Mock roles and capabilities for active pseudonym
		mockPermissionDAO.EXPECT().GetUnifiedActivePseudonymRolesAndCapabilities(gomock.Any(), testUserID, testPseudonymID, (*int32)(nil)).Return([]string{"user"}, []string{"create_content", "vote", "message", "report", "create_subforum"}, nil).Times(1)

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

		input := &middleware.AuthInput{
			Authorization: "Bearer " + token,
		}

		// Call the handler
		response, err := handler.GetCurrentUserSession(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, testEmail, response.Body.Email)
		assert.Equal(t, int(testUserID), response.Body.UserID)
		assert.Equal(t, testPseudonymID, response.Body.ActivePseudonymID)
		assert.Equal(t, testDisplayName, response.Body.DisplayName)
	})

	t.Run("GetCurrentUserSessionWithInvalidToken", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Initialize global auth middleware for testing
		middleware.SetGlobalAuthMiddleware(middleware.NewAuthMiddleware("test-secret", nil, &config.JWTConfig{
			Secret:      "test-secret",
			Expiration:  time.Hour,
			Development: true,
		}, &config.SecurityConfig{}))

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

		// Create input with invalid token
		input := &middleware.AuthInput{
			Authorization: "Bearer invalid_token",
		}

		// Call the handler
		response, err := handler.GetCurrentUserSession(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

// TestAuthHandler_RefreshToken_Gomock tests the token refresh functionality using gomock
func TestAuthHandler_RefreshToken_Gomock(t *testing.T) {
	t.Run("RefreshTokenWithValidToken", func(t *testing.T) {
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

		// Create user context
		userCtx := &middleware.UserContext{
			UserID:            1,
			Email:             "test@example.com",
			ActivePseudonymID: "pseudonym-123",
			DisplayName:       "TestUser",
			MFAEnabled:        false, // roles and capabilities deprecated
		}

		// Generate refresh token
		refreshToken, err := middleware.GenerateJWT(userCtx, "test-secret-key", 7*24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			RefreshToken string `cookie:"refresh_token"`
			Body         apimodels.RefreshTokenBody
		}{
			RefreshToken: refreshToken,
		}

		// Call the handler
		response, err := handler.RefreshToken(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Body.AccessToken)
		assert.Greater(t, response.Body.ExpiresIn, 0)
	})

	t.Run("RefreshTokenWithInvalidToken", func(t *testing.T) {
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

		// Create input with invalid token
		input := &struct {
			RefreshToken string `cookie:"refresh_token"`
			Body         apimodels.RefreshTokenBody
		}{
			RefreshToken: "invalid_token",
		}

		// Call the handler
		response, err := handler.RefreshToken(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}

// TestAuthHandler_Logout_Gomock tests the logout functionality using gomock
func TestAuthHandler_Logout_Gomock(t *testing.T) {
	t.Run("LogoutUser", func(t *testing.T) {
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

		// Create input
		input := &apimodels.UserLogoutInput{
			Body: apimodels.UserLogoutBody{},
		}

		// Call the handler
		response, err := handler.LogoutUser(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotEmpty(t, response.Message)
		assert.Len(t, response.Cookies, 2) // Should have expired cookies for clearing
	})
}
