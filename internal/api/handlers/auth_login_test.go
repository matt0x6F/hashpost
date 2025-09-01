package handlers_test

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// hashPasswordSHA256 creates a SHA256 hash of the password
func hashPasswordSHA256(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// TestAuthHandler_Login tests the login functionality using gomock
func TestAuthHandler_Login(t *testing.T) {
	t.Run("LoginWithValidCredentials", func(t *testing.T) {
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
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, mockPseudonymDAO, nil, mockRoleKeyDAO, ibeSystem, nil, nil, nil, nil, nil)

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
		mockUserDAO.EXPECT().GetUserByEmail(gomock.Any(), testEmail).Return(mockUser, nil).Times(1)
		mockUserDAO.EXPECT().UpdateLastActive(gomock.Any(), testUserID).Return(nil).Times(1)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.EXPECT().GetPseudonymsByUserID(gomock.Any(), testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil).Times(2)
		mockPseudonymDAO.EXPECT().GetDefaultPseudonymByUserID(gomock.Any(), testUserID, "user", "authentication").Return(mockPseudonym, nil).Times(2)

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
	})

	t.Run("LoginWithInvalidCredentials", func(t *testing.T) {
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
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, nil, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

		// Mock user not found - return nil, nil to avoid panic
		mockUserDAO.EXPECT().GetUserByEmail(gomock.Any(), "nonexistent@example.com").Return((*dbmodels.User)(nil), nil).Times(1)

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
	})

	t.Run("LoginWithWrongPassword", func(t *testing.T) {
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
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, nil, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

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
		mockUserDAO.EXPECT().GetUserByEmail(gomock.Any(), testEmail).Return(mockUser, nil).Times(1)

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
	})

	t.Run("LoginWithInactiveUser", func(t *testing.T) {
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
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, nil, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

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
		mockUserDAO.EXPECT().GetUserByEmail(gomock.Any(), testEmail).Return(mockUser, nil).Times(1)

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
	})

	t.Run("LoginWithUnverifiedEmail", func(t *testing.T) {
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
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, nil, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

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
		mockUserDAO.EXPECT().GetUserByEmail(gomock.Any(), testEmail).Return(mockUser, nil).Times(1)

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
	})

	t.Run("LoginWithAdminUser", func(t *testing.T) {
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
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, mockUserDAO, mockPseudonymDAO, nil, mockRoleKeyDAO, ibeSystem, nil, nil, nil, nil, nil)

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
		mockUserDAO.EXPECT().GetUserByEmail(gomock.Any(), testEmail).Return(mockUser, nil).Times(1)
		mockUserDAO.EXPECT().UpdateLastActive(gomock.Any(), testUserID).Return(nil).Times(1)

		// Mock pseudonym retrieval
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: testPseudonymID,
			DisplayName: testDisplayName,
		}
		mockPseudonymDAO.EXPECT().GetPseudonymsByUserID(gomock.Any(), testUserID, testPseudonymID, "user", "authentication").Return([]*dbmodels.Pseudonym{mockPseudonym}, nil).Times(2)
		mockPseudonymDAO.EXPECT().GetDefaultPseudonymByUserID(gomock.Any(), testUserID, "user", "authentication").Return(mockPseudonym, nil).Times(2)

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
	})
}
