package handlers_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// createTestAuthHandler creates an AuthHandler with mocked dependencies
func createTestAuthHandler() (*handlers.AuthHandler, *mocks.MockUserDAO, *mocks.MockSecurePseudonymDAO, *mocks.MockIdentityMappingDAO, *mocks.MockRoleKeyDAO, *mocks.MockSubforumDAO) {
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
	}

	mockUserDAO := &mocks.MockUserDAO{}
	mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()
	mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
	mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
	mockSubforumDAO := mocks.NewMockSubforumDAO()
	ibeSystem := ibe.NewIBESystem()

	handler := handlers.NewAuthHandlerWithDependencies(cfg, mockUserDAO, mockSecurePseudonymDAO, mockIdentityMappingDAO, mockRoleKeyDAO, ibeSystem, mockSubforumDAO)

	return handler, mockUserDAO, mockSecurePseudonymDAO, mockIdentityMappingDAO, mockRoleKeyDAO, mockSubforumDAO
}

// TestAuthHandler_Login tests the login functionality
func TestAuthHandler_Login(t *testing.T) {
	t.Run("LoginWithValidCredentials", func(t *testing.T) {
		handler, mockUserDAO, mockSecurePseudonymDAO, _, _, _ := createTestAuthHandler()

		// Test data
		testEmail := "test@example.com"
		testPassword := "TestPassword123!"
		testUserID := int64(1)
		testPseudonymID := "pseudonym-123"
		testDisplayName := "TestUser"

		// Mock user lookup - use the actual password hash that the handler expects
		hashedPassword := hashPasswordSHA256(testPassword)
		mockUser := &dbmodels.User{
			UserID:       testUserID,
			Email:        testEmail,
			PasswordHash: hashedPassword,
			IsActive:     sql.Null[bool]{V: true, Valid: true},
		}
		mockUserDAO.On("GetUserByEmail", mock.Anything, testEmail).Return(mockUser, nil)
		mockUserDAO.On("UpdateLastActive", mock.Anything, testUserID).Return(nil)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockSecurePseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil)
		mockSecurePseudonymDAO.On("GetDefaultPseudonymByUserID", mock.Anything, testUserID, "user", "authentication").Return(mockPseudonym, nil)

		// Create login input
		input := &apimodels.UserLoginInput{
			Body: apimodels.UserLoginBody{
				Email:    testEmail,
				Password: testPassword,
			},
		}

		// Call the handler
		response, err := handler.LoginUser(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, testEmail, response.Body.Email)
		assert.NotEmpty(t, response.Body.AccessToken)
		assert.NotEmpty(t, response.Body.RefreshToken)
		assert.Equal(t, testPseudonymID, response.Body.ActivePseudonymID)
		assert.Equal(t, testDisplayName, response.Body.DisplayName)

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
		mockSecurePseudonymDAO.AssertExpectations(t)
	})

	t.Run("LoginWithInvalidCredentials", func(t *testing.T) {
		handler, mockUserDAO, _, _, _, _ := createTestAuthHandler()

		// Mock user not found - return nil, nil to avoid panic
		mockUserDAO.On("GetUserByEmail", mock.Anything, "nonexistent@example.com").Return((*dbmodels.User)(nil), nil)

		// Create login input
		input := &apimodels.UserLoginInput{
			Body: apimodels.UserLoginBody{
				Email:    "nonexistent@example.com",
				Password: "WrongPassword",
			},
		}

		// Call the handler
		response, err := handler.LoginUser(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid credentials")

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
	})

	t.Run("LoginWithWrongPassword", func(t *testing.T) {
		handler, mockUserDAO, _, _, _, _ := createTestAuthHandler()

		// Test data
		testEmail := "test@example.com"
		testPassword := "CorrectPassword123!"
		wrongPassword := "WrongPassword"
		testUserID := int64(1)

		// Mock user lookup - use the actual password hash
		hashedPassword := hashPasswordSHA256(testPassword)
		mockUser := &dbmodels.User{
			UserID:       testUserID,
			Email:        testEmail,
			PasswordHash: hashedPassword,
			IsActive:     sql.Null[bool]{V: true, Valid: true},
		}
		mockUserDAO.On("GetUserByEmail", mock.Anything, testEmail).Return(mockUser, nil)

		// Create login input
		input := &apimodels.UserLoginInput{
			Body: apimodels.UserLoginBody{
				Email:    testEmail,
				Password: wrongPassword,
			},
		}

		// Call the handler
		response, err := handler.LoginUser(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "invalid credentials")

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
	})

	t.Run("LoginWithInactiveUser", func(t *testing.T) {
		handler, mockUserDAO, _, _, _, _ := createTestAuthHandler()

		// Test data
		testEmail := "inactive@example.com"
		testPassword := "TestPassword123!"
		testUserID := int64(1)

		// Mock inactive user
		hashedPassword := hashPasswordSHA256(testPassword)
		mockUser := &dbmodels.User{
			UserID:       testUserID,
			Email:        testEmail,
			PasswordHash: hashedPassword,
			IsActive:     sql.Null[bool]{V: false, Valid: true},
		}
		mockUserDAO.On("GetUserByEmail", mock.Anything, testEmail).Return(mockUser, nil)

		// Create login input
		input := &apimodels.UserLoginInput{
			Body: apimodels.UserLoginBody{
				Email:    testEmail,
				Password: testPassword,
			},
		}

		// Call the handler
		response, err := handler.LoginUser(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "account inactive")

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
	})

	t.Run("LoginWithAdminUser", func(t *testing.T) {
		handler, mockUserDAO, mockSecurePseudonymDAO, _, _, _ := createTestAuthHandler()

		// Test data
		testEmail := "admin@example.com"
		testPassword := "AdminPassword123!"
		testUserID := int64(1)
		testPseudonymID := "admin-pseudonym-123"
		testDisplayName := "AdminUser"

		// Mock admin user lookup
		hashedPassword := hashPasswordSHA256(testPassword)
		rolesJSON, _ := json.Marshal([]string{"platform_admin"})
		rolesNull := sql.Null[types.JSON[json.RawMessage]]{}
		rolesNull.Scan(rolesJSON)
		mockUser := &dbmodels.User{
			UserID:       testUserID,
			Email:        testEmail,
			PasswordHash: hashedPassword,
			IsActive:     sql.Null[bool]{V: true, Valid: true},
			Roles:        rolesNull,
		}
		mockUserDAO.On("GetUserByEmail", mock.Anything, testEmail).Return(mockUser, nil)
		mockUserDAO.On("UpdateLastActive", mock.Anything, testUserID).Return(nil)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockSecurePseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, "platform_admin", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil)
		mockSecurePseudonymDAO.On("GetDefaultPseudonymByUserID", mock.Anything, testUserID, "platform_admin", "authentication").Return(mockPseudonym, nil)

		// Create login input
		input := &apimodels.UserLoginInput{
			Body: apimodels.UserLoginBody{
				Email:    testEmail,
				Password: testPassword,
			},
		}

		// Call the handler
		response, err := handler.LoginUser(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, testEmail, response.Body.Email)
		assert.NotEmpty(t, response.Body.AccessToken)
		assert.NotEmpty(t, response.Body.RefreshToken)
		assert.Equal(t, testPseudonymID, response.Body.ActivePseudonymID)
		assert.Equal(t, testDisplayName, response.Body.DisplayName)

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
		mockSecurePseudonymDAO.AssertExpectations(t)
	})
}

// TestAuthHandler_Registration tests the registration functionality
func TestAuthHandler_Registration(t *testing.T) {
	t.Run("RegisterUserWithValidData", func(t *testing.T) {
		handler, mockUserDAO, mockSecurePseudonymDAO, _, mockRoleKeyDAO, _ := createTestAuthHandler()

		// Test data
		testEmail := "newuser@example.com"
		testPassword := "SecurePassword123!"
		testDisplayName := "NewUser123"
		testUserID := int64(1)
		testPseudonymID := "pseudonym-123"

		// Mock user doesn't exist
		mockUserDAO.On("GetUserByEmail", mock.Anything, testEmail).Return((*dbmodels.User)(nil), nil)

		// Mock user creation
		hashedPassword := hashPasswordSHA256(testPassword)
		mockUser := &dbmodels.User{
			UserID:       testUserID,
			Email:        testEmail,
			PasswordHash: hashedPassword,
		}
		mockUserDAO.On("CreateUser", mock.Anything, testEmail, hashedPassword).Return(mockUser, nil)

		// Mock pseudonym creation
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockSecurePseudonymDAO.On("CreatePseudonymWithIdentityMapping", mock.Anything, testUserID, testDisplayName).Return(mockPseudonym, nil)

		// Mock role key creation
		mockRoleKeyDAO.On("EnsureDefaultKeys", mock.Anything, mock.Anything, testUserID).Return(nil)

		// Create registration input
		input := &apimodels.UserRegistrationInput{
			Body: apimodels.UserRegistrationBody{
				Email:       testEmail,
				Password:    testPassword,
				DisplayName: testDisplayName,
			},
		}

		// Call the handler
		response, err := handler.RegisterUser(context.Background(), input)

		// Assert response
		require.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, testEmail, response.Body.Email)
		assert.Equal(t, int(testUserID), response.Body.UserID)
		assert.NotEmpty(t, response.Body.AccessToken)
		assert.NotEmpty(t, response.Body.RefreshToken)
		assert.Equal(t, testPseudonymID, response.Body.PseudonymID)
		assert.Equal(t, testDisplayName, response.Body.DisplayName)
		assert.Len(t, response.Body.Roles, 1)
		assert.Len(t, response.Body.Capabilities, 5) // Default capabilities

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
		mockSecurePseudonymDAO.AssertExpectations(t)
		mockRoleKeyDAO.AssertExpectations(t)
	})

	t.Run("RegisterUserWithDuplicateEmail", func(t *testing.T) {
		handler, mockUserDAO, _, _, _, _ := createTestAuthHandler()

		// Test data
		testEmail := "duplicate@example.com"
		testPassword := "SecurePassword123!"
		testDisplayName := "DuplicateUser"

		// Mock existing user
		existingUser := &dbmodels.User{
			UserID: 1,
			Email:  testEmail,
		}
		mockUserDAO.On("GetUserByEmail", mock.Anything, testEmail).Return(existingUser, nil)

		// Create registration input
		input := &apimodels.UserRegistrationInput{
			Body: apimodels.UserRegistrationBody{
				Email:       testEmail,
				Password:    testPassword,
				DisplayName: testDisplayName,
			},
		}

		// Call the handler
		response, err := handler.RegisterUser(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "already exists")

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
	})

	t.Run("RegisterUserWithInvalidEmail", func(t *testing.T) {
		handler, _, _, _, _, _ := createTestAuthHandler()

		// Create registration input with invalid email
		input := &apimodels.UserRegistrationInput{
			Body: apimodels.UserRegistrationBody{
				Email:       "invalid-email",
				Password:    "SecurePassword123!",
				DisplayName: "TestUser",
			},
		}

		// Call the handler
		response, err := handler.RegisterUser(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "email format is invalid")
	})

	t.Run("RegisterUserWithWeakPassword", func(t *testing.T) {
		handler, _, _, _, _, _ := createTestAuthHandler()

		// Create registration input with weak password
		input := &apimodels.UserRegistrationInput{
			Body: apimodels.UserRegistrationBody{
				Email:       "test@example.com",
				Password:    "weak",
				DisplayName: "TestUser",
			},
		}

		// Call the handler
		response, err := handler.RegisterUser(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "password")
	})

	t.Run("RegisterUserWithInvalidDisplayName", func(t *testing.T) {
		handler, _, _, _, _, _ := createTestAuthHandler()

		// Create registration input with invalid display name
		input := &apimodels.UserRegistrationInput{
			Body: apimodels.UserRegistrationBody{
				Email:       "test@example.com",
				Password:    "SecurePassword123!",
				DisplayName: "", // Empty display name
			},
		}

		// Call the handler
		response, err := handler.RegisterUser(context.Background(), input)

		// Assert response
		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "display_name")
	})
}

// TestAuthHandler_CurrentUserSession tests the current user session functionality
func TestAuthHandler_CurrentUserSession(t *testing.T) {
	t.Run("GetCurrentUserSession", func(t *testing.T) {
		handler, mockUserDAO, mockSecurePseudonymDAO, _, _, _ := createTestAuthHandler()

		// Initialize global auth middleware for testing
		middleware.SetGlobalAuthMiddleware(middleware.NewAuthMiddleware("test-secret", nil, &config.JWTConfig{
			Secret:      "test-secret",
			Expiration:  time.Hour,
			Development: true,
		}, &config.SecurityConfig{}))

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
		mockUserDAO.On("GetUserByID", mock.Anything, testUserID).Return(mockUser, nil)
		mockUserDAO.On("UpdateLastActive", mock.Anything, testUserID).Return(nil)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockSecurePseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil)

		// Create input with valid JWT token
		userCtx := &middleware.UserContext{
			UserID:            testUserID,
			Email:             testEmail,
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote", "message", "report", "create_subforum"},
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

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
		mockSecurePseudonymDAO.AssertExpectations(t)
	})

	t.Run("GetCurrentUserSessionWithInvalidToken", func(t *testing.T) {
		handler, _, _, _, _, _ := createTestAuthHandler()

		// Initialize global auth middleware for testing
		middleware.SetGlobalAuthMiddleware(middleware.NewAuthMiddleware("test-secret", nil, &config.JWTConfig{
			Secret:      "test-secret",
			Expiration:  time.Hour,
			Development: true,
		}, &config.SecurityConfig{}))

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

// TestAuthHandler_RefreshToken tests the token refresh functionality
func TestAuthHandler_RefreshToken(t *testing.T) {
	t.Run("RefreshTokenWithValidToken", func(t *testing.T) {
		handler, _, _, _, _, _ := createTestAuthHandler()

		// Create user context
		userCtx := &middleware.UserContext{
			UserID:            1,
			Email:             "test@example.com",
			ActivePseudonymID: "pseudonym-123",
			DisplayName:       "TestUser",
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote", "message", "report", "create_subforum"},
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
		handler, _, _, _, _, _ := createTestAuthHandler()

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

// TestAuthHandler_Logout tests the logout functionality
func TestAuthHandler_Logout(t *testing.T) {
	t.Run("LogoutUser", func(t *testing.T) {
		handler, _, _, _, _, _ := createTestAuthHandler()

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

// hashPasswordSHA256 is a helper function to hash passwords for testing using SHA-256
func hashPasswordSHA256(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
