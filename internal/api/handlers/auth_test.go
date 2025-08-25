package handlers_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// hashPasswordSHA256 creates a SHA256 hash of the password
func hashPasswordSHA256(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// NewAuthHandlerWithMocks creates a new auth handler with mock DAOs and fixture data
func NewAuthHandlerWithMocks() (*handlers.AuthHandler, *mocks.MockUserDAO, *mocks.MockPseudonymDAO, *mocks.MockIdentityMappingDAO, *mocks.MockRoleKeyDAO, *mocks.MockSubforumDAO, *mocks.MockPermissionDAO) {
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
				Enabled:       true,                  // Enable email validation for tests
				VerifierEmail: "noreply@example.com", // Default verifier email for tests
			},
		},
	}

	mockUserDAO := &mocks.MockUserDAO{}
	mockPseudonymDAO := mocks.NewMockPseudonymDAO()
	mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
	mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
	mockSubforumDAO := mocks.NewMockSubforumDAO()
	mockPermissionDAO := mocks.NewMockPermissionDAO()
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

	// Create handler with the SAME mock instances that we return
	// Note: Email service and token DAOs are nil for tests since we're not testing email functionality
	handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, mockPseudonymDAO, mockIdentityMappingDAO, mockRoleKeyDAO, ibeSystem, mockSubforumDAO, mockPermissionDAO, nil, nil, nil)

	// Return the SAME mock instances that the handler is using
	return handler, mockUserDAO, mockPseudonymDAO, mockIdentityMappingDAO, mockRoleKeyDAO, mockSubforumDAO, mockPermissionDAO
}

// TestAuthHandler_Login tests the login functionality
func TestAuthHandler_Login(t *testing.T) {
	t.Run("LoginWithValidCredentials", func(t *testing.T) {
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, _, _ := NewAuthHandlerWithMocks()

		// Test data
		testEmail := "test@example.com"
		testPassword := "TestPassword123!"
		testUserID := int64(1)
		testPseudonymID := "pseudonym-123"
		testDisplayName := "TestUser"

		// Mock user lookup - use the actual password hash that the handler expects
		hashedPassword := hashPasswordSHA256(testPassword)
		mockUser := &dbmodels.User{
			UserID:        testUserID,
			Email:         testEmail,
			PasswordHash:  hashedPassword,
			IsActive:      sql.Null[bool]{V: true, Valid: true},
			EmailVerified: sql.Null[bool]{V: true, Valid: true},
		}
		mockUserDAO.On("GetUserByEmail", mock.Anything, testEmail).Return(mockUser, nil)
		mockUserDAO.On("UpdateLastActive", mock.Anything, testUserID).Return(nil)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil)
		mockPseudonymDAO.On("GetDefaultPseudonymByUserID", mock.Anything, testUserID, "user", "authentication").Return(mockPseudonym, nil)

		// Mock role keys for the default pseudonym
		mockRoleKeyDAO.On("ListRoleKeysByPseudonym", mock.Anything, testPseudonymID).Return([]*dbmodels.RoleKey{}, nil)

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
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("LoginWithInvalidCredentials", func(t *testing.T) {
		handler, mockUserDAO, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		handler, mockUserDAO, _, _, _, _, _ := NewAuthHandlerWithMocks()

		// Test data
		testEmail := "test@example.com"
		testPassword := "CorrectPassword123!"
		wrongPassword := "WrongPassword"
		testUserID := int64(1)

		// Mock user lookup - use the actual password hash
		hashedPassword := hashPasswordSHA256(testPassword)
		mockUser := &dbmodels.User{
			UserID:        testUserID,
			Email:         testEmail,
			PasswordHash:  hashedPassword,
			IsActive:      sql.Null[bool]{V: true, Valid: true},
			EmailVerified: sql.Null[bool]{V: true, Valid: true},
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
		handler, mockUserDAO, _, _, _, _, _ := NewAuthHandlerWithMocks()

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

	t.Run("LoginWithUnverifiedEmail", func(t *testing.T) {
		handler, mockUserDAO, _, _, _, _, _ := NewAuthHandlerWithMocks()

		// Test data
		testEmail := "unverified@example.com"
		testPassword := "TestPassword123!"
		testUserID := int64(1)

		// Mock user with unverified email
		hashedPassword := hashPasswordSHA256(testPassword)
		mockUser := &dbmodels.User{
			UserID:        testUserID,
			Email:         testEmail,
			PasswordHash:  hashedPassword,
			IsActive:      sql.Null[bool]{V: true, Valid: true},
			EmailVerified: sql.Null[bool]{V: false, Valid: true},
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
		assert.Contains(t, err.Error(), "email not verified")

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
	})

	t.Run("LoginWithAdminUser", func(t *testing.T) {
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, _, _ := NewAuthHandlerWithMocks()

		// Test data
		testEmail := "admin@example.com"
		testPassword := "AdminPassword123!"
		testUserID := int64(1)
		testPseudonymID := "admin-pseudonym-123"
		testDisplayName := "AdminUser"

		// Mock admin user lookup
		hashedPassword := hashPasswordSHA256(testPassword)
		mockUser := &dbmodels.User{
			UserID:        testUserID,
			Email:         testEmail,
			PasswordHash:  hashedPassword,
			IsActive:      sql.Null[bool]{V: true, Valid: true},
			EmailVerified: sql.Null[bool]{V: true, Valid: true},
		}
		mockUserDAO.On("GetUserByEmail", mock.Anything, testEmail).Return(mockUser, nil)
		mockUserDAO.On("UpdateLastActive", mock.Anything, testUserID).Return(nil)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil)
		mockPseudonymDAO.On("GetDefaultPseudonymByUserID", mock.Anything, testUserID, "user", "authentication").Return(mockPseudonym, nil)

		// Mock role keys for the default pseudonym
		mockRoleKeyDAO.On("ListRoleKeysByPseudonym", mock.Anything, testPseudonymID).Return([]*dbmodels.RoleKey{}, nil)

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
		mockPseudonymDAO.AssertExpectations(t)
	})
}

// TestAuthHandler_Registration tests the registration functionality
func TestAuthHandler_Registration(t *testing.T) {
	t.Run("RegisterUserWithValidData", func(t *testing.T) {
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, _, _ := NewAuthHandlerWithMocks()

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

		// Mock user pseudonym creation (this will be the default pseudonym since it's the first one)
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.On("CreatePseudonymWithIdentityMapping", mock.Anything, testUserID, testDisplayName).Return(mockPseudonym, nil)

		// Mock role key creation for the user's pseudonym
		mockRoleKeyDAO.On("EnsureDefaultKeys", mock.Anything, mock.Anything, testPseudonymID, []string{"user"}).Return(nil)

		// Note: ListRoleKeysByPseudonym is no longer called during registration since we don't return roles/capabilities

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
		// Registration no longer returns tokens - user must verify email first
		assert.Empty(t, response.Body.AccessToken)
		assert.Empty(t, response.Body.RefreshToken)
		assert.Equal(t, testPseudonymID, response.Body.PseudonymID)
		assert.Equal(t, testDisplayName, response.Body.DisplayName)
		assert.Len(t, response.Body.Roles, 0)        // Roles are now checked dynamically, not returned in registration
		assert.Len(t, response.Body.Capabilities, 0) // Capabilities are now checked dynamically, not returned in registration

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
		// Note: mockRoleKeyDAO is no longer called during registration since we don't return roles/capabilities
		// mockRoleKeyDAO.AssertExpectations(t)
	})

	t.Run("RegisterUserWithDuplicateEmail", func(t *testing.T) {
		handler, mockUserDAO, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		handler, mockUserDAO, _, _, _, _, _ := NewAuthHandlerWithMocks()

		// Mock that no user exists with this email (since it's invalid, it shouldn't exist)
		mockUserDAO.On("GetUserByEmail", mock.Anything, "invalid-email").Return((*dbmodels.User)(nil), nil)

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
		assert.Contains(t, err.Error(), "email validation failed: email does not match the regular expression")
	})

	t.Run("RegisterUserWithWeakPassword", func(t *testing.T) {
		handler, _, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		handler, _, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, _, mockPermissionDAO := NewAuthHandlerWithMocks()

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
		mockPseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil)

		// Mock role keys for the default pseudonym
		mockRoleKeyDAO.On("ListRoleKeysByPseudonym", mock.Anything, testPseudonymID).Return([]*dbmodels.RoleKey{}, nil)

		// Mock roles and capabilities for active pseudonym
		mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities", mock.Anything, testUserID, testPseudonymID, (*int32)(nil)).Return([]string{"user"}, []string{"create_content", "vote", "message", "report", "create_subforum"}, nil)

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

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
		mockRoleKeyDAO.AssertExpectations(t)
		mockPermissionDAO.AssertExpectations(t)
	})

	t.Run("GetCurrentUserSessionWithInvalidToken", func(t *testing.T) {
		handler, _, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		handler, _, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		handler, _, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		handler, _, _, _, _, _, _ := NewAuthHandlerWithMocks()

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

// TestAuthHandler_SwitchPseudonym tests the pseudonym switching functionality
func TestAuthHandler_SwitchPseudonym(t *testing.T) {
	handler, _, mockPseudonymDAO, _, _, _, _ := NewAuthHandlerWithMocks()

	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()
		userCtx.ActivePseudonymID = "current-pseudonym-123"
		userCtx.DisplayName = "Current User"

		// Mock target pseudonym data
		targetPseudonym := fixtures.CreateTestPseudonym()
		targetPseudonym.PseudonymID = "target-pseudonym-456"
		targetPseudonym.DisplayName = "Target User"

		// Set up mock expectations
		mockPseudonymDAO.On("GetPseudonymByID", ctx, "target-pseudonym-456").Return(targetPseudonym, nil)
		mockPseudonymDAO.On("VerifyPseudonymOwnership", ctx, "target-pseudonym-456", int64(1), "current-pseudonym-123", "user", "authentication").Return(true, nil)
		mockPseudonymDAO.On("UpdateLastActive", ctx, "target-pseudonym-456").Return(nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.SwitchPseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SwitchPseudonymInput: apimodels.SwitchPseudonymInput{
				Body: apimodels.SwitchPseudonymBody{
					PseudonymID: "target-pseudonym-456",
				},
			},
		}

		// Call the method
		result, err := handler.SwitchPseudonym(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 200, result.Status)
		require.Equal(t, "Pseudonym switched successfully", result.Body.Message)
		// Verify that cookies are set (the token is now in HTTP-only cookies)
		require.NotEmpty(t, result.Cookies)

		// Verify mock expectations
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("SwitchToSamePseudonym", func(t *testing.T) {
		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()
		userCtx.ActivePseudonymID = "current-pseudonym-123"

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.SwitchPseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SwitchPseudonymInput: apimodels.SwitchPseudonymInput{
				Body: apimodels.SwitchPseudonymBody{
					PseudonymID: "current-pseudonym-123", // Same as current
				},
			},
		}

		// Call the method
		result, err := handler.SwitchPseudonym(ctx, input)

		// Assertions
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "Already using this pseudonym")
	})

	t.Run("PseudonymNotFound", func(t *testing.T) {
		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()

		// Set up mock expectations
		mockPseudonymDAO.On("GetPseudonymByID", ctx, "nonexistent-pseudonym").Return(nil, nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.SwitchPseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SwitchPseudonymInput: apimodels.SwitchPseudonymInput{
				Body: apimodels.SwitchPseudonymBody{
					PseudonymID: "nonexistent-pseudonym",
				},
			},
		}

		// Call the method
		result, err := handler.SwitchPseudonym(ctx, input)

		// Assertions
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "Target pseudonym not found")

		// Verify mock expectations
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("UnauthorizedPseudonym", func(t *testing.T) {
		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()

		// Mock target pseudonym data
		targetPseudonym := fixtures.CreateTestPseudonym()
		targetPseudonym.PseudonymID = "unauthorized-pseudonym-789"

		// Set up mock expectations
		mockPseudonymDAO.On("GetPseudonymByID", ctx, "unauthorized-pseudonym-789").Return(targetPseudonym, nil)
		mockPseudonymDAO.On("VerifyPseudonymOwnership", ctx, "unauthorized-pseudonym-789", int64(1), "test-pseudonym-id", "user", "authentication").Return(false, nil)
		mockPseudonymDAO.On("VerifyPseudonymOwnership", ctx, "unauthorized-pseudonym-789", int64(1), "test-pseudonym-id", "user", "self_correlation").Return(false, nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.SwitchPseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SwitchPseudonymInput: apimodels.SwitchPseudonymInput{
				Body: apimodels.SwitchPseudonymBody{
					PseudonymID: "unauthorized-pseudonym-789",
				},
			},
		}

		// Call the method
		result, err := handler.SwitchPseudonym(ctx, input)

		// Assertions
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "You do not own this pseudonym")

		// Verify mock expectations
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("MultiScopeFallbackStrategy", func(t *testing.T) {
		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()
		// Roles deprecated - permissions now checked dynamically
		userCtx.ActivePseudonymID = "current-pseudonym-123"
		userCtx.DisplayName = "Current User"

		// Mock target pseudonym data
		targetPseudonym := fixtures.CreateTestPseudonym()
		targetPseudonym.PseudonymID = "target-pseudonym-456"
		targetPseudonym.DisplayName = "Target User"

		// Set up mock expectations for multi-scope fallback
		mockPseudonymDAO.On("GetPseudonymByID", ctx, "target-pseudonym-456").Return(targetPseudonym, nil)

		// Authentication scope with user role should succeed
		mockPseudonymDAO.On("VerifyPseudonymOwnership", ctx, "target-pseudonym-456", int64(1), "current-pseudonym-123", "user", "authentication").Return(true, nil)

		mockPseudonymDAO.On("UpdateLastActive", ctx, "target-pseudonym-456").Return(nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.SwitchPseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SwitchPseudonymInput: apimodels.SwitchPseudonymInput{
				Body: apimodels.SwitchPseudonymBody{
					PseudonymID: "target-pseudonym-456",
				},
			},
		}

		// Call the method
		result, err := handler.SwitchPseudonym(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 200, result.Status)
		require.Equal(t, "Pseudonym switched successfully", result.Body.Message)

		// Verify mock expectations
		mockPseudonymDAO.AssertExpectations(t)
	})

}

func TestAuthHandler_DeactivatePseudonym(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()

		// Create handler and mocks for this specific test
		handler, _, mockPseudonymDAO, _, _, _, _ := NewAuthHandlerWithMocks()

		// Set up mock expectations
		mockPseudonymDAO.On("DeactivatePseudonym", ctx, "test-pseudonym-123", int64(1), "active-pseudonym-456", "user", "self_correlation").Return(nil)

		// Create test request
		input := &struct {
			middleware.AuthInput
			apimodels.DeactivatePseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "active-pseudonym-456"),
			},
			DeactivatePseudonymInput: apimodels.DeactivatePseudonymInput{
				Body: apimodels.DeactivatePseudonymBody{
					PseudonymID: "test-pseudonym-123",
				},
			},
		}

		// Call the handler
		response, err := handler.DeactivatePseudonym(ctx, input)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, "Pseudonym deactivated successfully", response.Body.Message)

		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("MissingPseudonymID", func(t *testing.T) {
		ctx := context.Background()

		// Create handler and mocks for this specific test
		handler, _, _, _, _, _, _ := NewAuthHandlerWithMocks()

		// Create test request with missing pseudonym ID
		input := &struct {
			middleware.AuthInput
			apimodels.DeactivatePseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "active-pseudonym-456"),
			},
			DeactivatePseudonymInput: apimodels.DeactivatePseudonymInput{
				Body: apimodels.DeactivatePseudonymBody{
					PseudonymID: "",
				},
			},
		}

		// Call the handler
		response, err := handler.DeactivatePseudonym(ctx, input)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, response)
	})

	t.Run("PseudonymNotOwned", func(t *testing.T) {
		ctx := context.Background()

		// Create handler and mocks for this specific test
		handler, _, mockPseudonymDAO, _, _, _, _ := NewAuthHandlerWithMocks()

		// Set up mock expectations to return ownership error
		mockPseudonymDAO.On("DeactivatePseudonym", mock.Anything, "test-pseudonym-123", int64(1), "active-pseudonym-456", "user", "self_correlation").Return(fmt.Errorf("does not own"))

		// Create test request
		input := &struct {
			middleware.AuthInput
			apimodels.DeactivatePseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "active-pseudonym-456"),
			},
			DeactivatePseudonymInput: apimodels.DeactivatePseudonymInput{
				Body: apimodels.DeactivatePseudonymBody{
					PseudonymID: "test-pseudonym-123",
				},
			},
		}

		// Call the handler
		response, err := handler.DeactivatePseudonym(ctx, input)

		// Debug output
		t.Logf("Error: %v", err)
		t.Logf("Response: %+v", response)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, response)

		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("PseudonymNotFound", func(t *testing.T) {
		ctx := context.Background()

		// Create handler and mocks for this specific test
		handler, _, mockPseudonymDAO, _, _, _, _ := NewAuthHandlerWithMocks()

		// Set up mock expectations to return not found error
		mockPseudonymDAO.On("DeactivatePseudonym", mock.Anything, "non-existent-pseudonym", int64(1), "test-pseudonym-123", "user", "self_correlation").Return(fmt.Errorf("not found"))

		// Create test request
		input := &struct {
			middleware.AuthInput
			apimodels.DeactivatePseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "test-pseudonym-123"),
			},
			DeactivatePseudonymInput: apimodels.DeactivatePseudonymInput{
				Body: apimodels.DeactivatePseudonymBody{
					PseudonymID: "non-existent-pseudonym",
				},
			},
		}

		// Call the handler
		response, err := handler.DeactivatePseudonym(ctx, input)

		// Debug output
		t.Logf("Error: %v", err)
		t.Logf("Response: %+v", response)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, response)

		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("DatabaseError", func(t *testing.T) {
		ctx := context.Background()

		// Create handler and mocks for this specific test
		handler, _, mockPseudonymDAO, _, _, _, _ := NewAuthHandlerWithMocks()

		// Set up mock expectations to return database error
		mockPseudonymDAO.On("DeactivatePseudonym", mock.Anything, "test-pseudonym-123", int64(1), "active-pseudonym-456", "user", "self_correlation").Return(fmt.Errorf("database connection failed"))

		// Create test request
		input := &struct {
			middleware.AuthInput
			apimodels.DeactivatePseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "active-pseudonym-456"),
			},
			DeactivatePseudonymInput: apimodels.DeactivatePseudonymInput{
				Body: apimodels.DeactivatePseudonymBody{
					PseudonymID: "test-pseudonym-123",
				},
			},
		}

		// Call the handler
		response, err := handler.DeactivatePseudonym(ctx, input)

		// Debug output
		t.Logf("Error: %v", err)
		t.Logf("Response: %+v", response)

		// Assertions
		assert.Error(t, err)
		assert.Nil(t, response)

		mockPseudonymDAO.AssertExpectations(t)
	})
}

func TestAuthHandler_DeactivatePseudonym_Simple(t *testing.T) {
	handler, _, mockPseudonymDAO, _, _, _, _ := NewAuthHandlerWithMocks()

	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("SimpleSuccess", func(t *testing.T) {
		ctx := context.Background()

		// Set up mock expectations
		mockPseudonymDAO.On("DeactivatePseudonym", ctx, "test-pseudonym-123", int64(1), "active-pseudonym-456", "user", "self_correlation").Return(nil)

		// Create test request
		input := &struct {
			middleware.AuthInput
			apimodels.DeactivatePseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(1, "active-pseudonym-456"),
			},
			DeactivatePseudonymInput: apimodels.DeactivatePseudonymInput{
				Body: apimodels.DeactivatePseudonymBody{
					PseudonymID: "test-pseudonym-123",
				},
			},
		}

		// Call the handler
		response, err := handler.DeactivatePseudonym(ctx, input)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, "Pseudonym deactivated successfully", response.Body.Message)

		// Verify mock was called
		mockPseudonymDAO.AssertExpectations(t)
	})
}

// TestAuthHandler_GetCurrentUserSessionForSubforum tests the subforum-specific session functionality
func TestAuthHandler_GetCurrentUserSessionForSubforum(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, &config.JWTConfig{
		Secret:      "test-secret",
		Expiration:  time.Hour,
		Development: true,
	}, &config.SecurityConfig{})
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("SuccessWithSubforumCapabilities", func(t *testing.T) {
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, mockSubforumDAO, mockPermissionDAO := NewAuthHandlerWithMocks()

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
		mockUserDAO.On("GetUserByID", mock.Anything, testUserID).Return(mockUser, nil)
		mockUserDAO.On("UpdateLastActive", mock.Anything, testUserID).Return(nil)

		// Mock subforum lookup
		mockSubforum := &dbmodels.Subforum{
			SubforumID: int32(testSubforumID),
			Name:       testSubforumName,
		}
		mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, "h", testSubforumName).Return(
			func(ctx context.Context, communityType string, name string) (*dbmodels.Subforum, error) {
				return mockSubforum, nil
			},
		)

		// Mock permission DAO - user has moderator capabilities
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(testUserID), testPseudonymID, "moderate_content", mock.AnythingOfType("*int32")).Return(true, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(testUserID), testPseudonymID, "manage_moderators", mock.AnythingOfType("*int32")).Return(false, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(testUserID), testPseudonymID, "ban_users", mock.AnythingOfType("*int32")).Return(true, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(testUserID), testPseudonymID, "sticky_post", mock.AnythingOfType("*int32")).Return(false, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(testUserID), testPseudonymID, "lock_post", mock.AnythingOfType("*int32")).Return(false, nil)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil)

		// Mock role keys for the active pseudonym with platform capabilities
		mockRoleKey := &dbmodels.RoleKey{
			PseudonymID:  testPseudonymID,
			RoleName:     "user",
			Capabilities: types.NewJSON[json.RawMessage]([]byte(`["create_content", "vote", "message", "report", "create_subforum"]`)),
		}
		mockRoleKeyDAO.On("ListRoleKeysByPseudonym", mock.Anything, testPseudonymID).Return([]*dbmodels.RoleKey{mockRoleKey}, nil)

		// Mock unified roles and capabilities for active pseudonym with subforum context
		subforumID := int32(testSubforumID)
		mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities", mock.Anything, testUserID, testPseudonymID, &subforumID).Return(
			[]string{constants.RoleUser, constants.RoleModerator},
			[]string{"create_content", "vote", "message", "report", "create_subforum", "moderate_content", "ban_users"},
			nil)

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

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
		mockRoleKeyDAO.AssertExpectations(t)
		mockSubforumDAO.AssertExpectations(t)
	})

	t.Run("SuccessWithSubforumModeratorCapabilities", func(t *testing.T) {
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, mockSubforumDAO, mockPermissionDAO := NewAuthHandlerWithMocks()

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
		mockUserDAO.On("GetUserByID", mock.Anything, testUserID).Return(mockUser, nil)
		mockUserDAO.On("UpdateLastActive", mock.Anything, testUserID).Return(nil)

		// Mock subforum lookup
		mockSubforum := &dbmodels.Subforum{
			SubforumID: int32(testSubforumID),
			Name:       testSubforumName,
		}
		mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, "h", testSubforumName).Return(
			func(ctx context.Context, communityType string, name string) (*dbmodels.Subforum, error) {
				return mockSubforum, nil
			},
		)

		// Mock permission DAO - user has moderator capabilities
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(testUserID), testPseudonymID, "moderate_content", mock.AnythingOfType("*int32")).Return(true, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(testUserID), testPseudonymID, "manage_moderators", mock.AnythingOfType("*int32")).Return(false, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(testUserID), testPseudonymID, "ban_users", mock.AnythingOfType("*int32")).Return(true, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(testUserID), testPseudonymID, "sticky_post", mock.AnythingOfType("*int32")).Return(false, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(testUserID), testPseudonymID, "lock_post", mock.AnythingOfType("*int32")).Return(false, nil)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil)

		// Mock role keys for the active pseudonym with platform capabilities
		mockRoleKey := &dbmodels.RoleKey{
			PseudonymID:  testPseudonymID,
			RoleName:     "user",
			Capabilities: types.NewJSON[json.RawMessage]([]byte(`["create_content", "vote", "message", "report", "create_subforum"]`)),
		}
		mockRoleKeyDAO.On("ListRoleKeysByPseudonym", mock.Anything, testPseudonymID).Return([]*dbmodels.RoleKey{mockRoleKey}, nil)

		// Mock unified roles and capabilities for active pseudonym with subforum context
		subforumID := int32(testSubforumID)
		mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities", mock.Anything, testUserID, testPseudonymID, &subforumID).Return(
			[]string{constants.RoleUser, constants.RoleModerator},
			[]string{"create_content", "vote", "message", "report", "create_subforum", "moderate_content", "ban_users"},
			nil)

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

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
		mockRoleKeyDAO.AssertExpectations(t)
		mockSubforumDAO.AssertExpectations(t)
	})

	t.Run("UnauthorizedAccess", func(t *testing.T) {
		handler, _, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		handler, _, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		handler, mockUserDAO, _, _, _, _, _ := NewAuthHandlerWithMocks()

		// Test data
		testUserID := int64(999)
		testEmail := "nonexistent@example.com"
		testPseudonymID := "pseudonym-123"
		testDisplayName := "TestUser"

		// Mock user not found
		mockUserDAO.On("GetUserByID", mock.Anything, testUserID).Return((*dbmodels.User)(nil), nil)

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

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
	})

	t.Run("InactiveUser", func(t *testing.T) {
		handler, mockUserDAO, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		mockUserDAO.On("GetUserByID", mock.Anything, testUserID).Return(mockUser, nil)

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

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
	})

	t.Run("SuspendedUser", func(t *testing.T) {
		handler, mockUserDAO, _, _, _, _, _ := NewAuthHandlerWithMocks()

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
		mockUserDAO.On("GetUserByID", mock.Anything, testUserID).Return(mockUser, nil)

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

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
	})

	t.Run("SubforumNotFound", func(t *testing.T) {
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, mockSubforumDAO, mockPermissionDAO := NewAuthHandlerWithMocks()

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
		mockUserDAO.On("GetUserByID", mock.Anything, testUserID).Return(mockUser, nil)
		mockUserDAO.On("UpdateLastActive", mock.Anything, testUserID).Return(nil)

		// Mock subforum not found
		mockSubforumDAO.On("GetSubforumByCommunityTypeAndName", mock.Anything, "h", testSubforumName).Return(
			func(ctx context.Context, communityType string, name string) (*dbmodels.Subforum, error) {
				return nil, nil
			},
		)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil)

		// Mock role keys for the active pseudonym with platform capabilities
		mockRoleKey := &dbmodels.RoleKey{
			PseudonymID:  testPseudonymID,
			RoleName:     "user",
			Capabilities: types.NewJSON[json.RawMessage]([]byte(`["create_content", "vote", "message", "report", "create_subforum"]`)),
		}
		mockRoleKeyDAO.On("ListRoleKeysByPseudonym", mock.Anything, mock.Anything).Return([]*dbmodels.RoleKey{mockRoleKey}, nil).Maybe()

		// Mock unified roles and capabilities for active pseudonym with subforum context (nil since subforum not found)
		mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities", mock.Anything, testUserID, testPseudonymID, (*int32)(nil)).Return([]string{"user"}, []string{"create_content", "vote", "message", "report", "create_subforum"}, nil)

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

		// Verify mocks were called
		mockUserDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
		mockSubforumDAO.AssertExpectations(t)
		mockRoleKeyDAO.AssertExpectations(t)
	})
}
