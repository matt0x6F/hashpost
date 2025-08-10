package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/mock"
)

// createBenchmarkAuthHandler creates a fully mocked auth handler for benchmarking
func createBenchmarkAuthHandler() *AuthHandler {
	// Create mock DAOs with realistic behavior
	mockUserDAO := &mocks.MockUserDAO{}
	mockPseudonymDAO := &mocks.MockPseudonymDAO{}
	mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
	mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
	mockSubforumDAO := &mocks.MockSubforumDAO{}
	mockPermissionDAO := &mocks.MockPermissionDAO{}
	mockEmailVerificationTokenDAO := &mocks.MockEmailVerificationTokenDAO{}
	mockPasswordResetTokenDAO := &mocks.MockPasswordResetTokenDAO{}

	// Set up realistic mock responses for registration
	mockUserDAO.On("CreateUser", mock.Anything, mock.Anything, mock.Anything).Return(&dbmodels.User{
		UserID: 1,
		Email:  "test@example.com",
	}, nil)

	// For registration: return nil (no existing user) for any email
	mockUserDAO.On("GetUserByEmail", mock.Anything, mock.Anything).Return(nil, nil)

	mockPseudonymDAO.On("CreatePseudonymWithIdentityMapping", mock.Anything, mock.Anything, mock.Anything).Return(&dbmodels.Pseudonym{
		PseudonymID: "pseudonym_123",
		DisplayName: "Test User",
	}, nil)

	mockRoleKeyDAO.On("EnsureDefaultKeys", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRoleKeyDAO.On("ListRoleKeysByPseudonym", mock.Anything, mock.Anything).Return([]*dbmodels.RoleKey{}, nil)

	// Mock GetUserByID for various operations
	mockUserDAO.On("GetUserByID", mock.Anything, int64(1)).Return(&dbmodels.User{
		UserID:        1,
		Email:         "test@example.com",
		PasswordHash:  "486f0a9839674a9dfcf52efbffe2469b01b87f5e5c967ec8f0ee3b7d9a308a8c",
		IsActive:      sql.Null[bool]{V: true, Valid: true},
		EmailVerified: sql.Null[bool]{V: true, Valid: true},
	}, nil)

	// Mock pseudonym operations
	mockPseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, int64(1), "user", "authentication").Return([]*dbmodels.Pseudonym{
		{
			PseudonymID: "pseudonym_123",
			DisplayName: "Test User",
			IsDefault:   true,
		},
	}, nil)
	mockPseudonymDAO.On("GetDefaultPseudonymByUserID", mock.Anything, int64(1), "user", "authentication").Return(&dbmodels.Pseudonym{
		PseudonymID: "pseudonym_123",
		DisplayName: "Test User",
		IsDefault:   true,
	}, nil)

	mockEmailVerificationTokenDAO.On("CreateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockEmailVerificationTokenDAO.On("GetToken", mock.Anything, mock.Anything).Return(&dbmodels.EmailVerificationToken{
		Token:     "valid_token",
		UserID:    1,
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil)
	mockEmailVerificationTokenDAO.On("MarkTokenAsUsed", mock.Anything, mock.Anything).Return(nil)

	// Mock UpdateUser for email verification
	mockUserDAO.On("UpdateUser", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	mockPasswordResetTokenDAO.On("CreateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockPasswordResetTokenDAO.On("GetToken", mock.Anything, mock.Anything).Return(&dbmodels.PasswordResetToken{
		Token:     "valid_reset_token",
		UserID:    1,
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil)
	mockPasswordResetTokenDAO.On("MarkTokenAsUsed", mock.Anything, mock.Anything).Return(nil)

	// Generate realistic 32-byte domain master keys (AES-256 equivalent)
	domainMasters := make(map[string][]byte)
	domains := []string{
		ibe.DOMAIN_USER_PSEUDONYMS,
		ibe.DOMAIN_USER_CORRELATION,
		ibe.DOMAIN_MOD_CORRELATION,
		ibe.DOMAIN_ADMIN_CORRELATION,
		ibe.DOMAIN_LEGAL_CORRELATION,
	}

	for _, domain := range domains {
		key := make([]byte, 32) // 256-bit keys for production
		rand.Read(key)
		domainMasters[domain] = key
	}

	// Create IBE system with production key sizes
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
		DomainMasters: domainMasters,
		KeyVersion:    1,
		Salt:          "production_fingerprint_salt_v1_secure_random_string",
	})

	// Create config with production JWT secret and disabled email validation
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "production_jwt_secret_256_bit_key_for_benchmarking_secure_random_string_32_bytes_long",
			Expiration: 24 * time.Hour,
		},
		Email: config.EmailConfig{
			Provider: "", // Disable email service for benchmarking
			Validation: config.EmailValidationConfig{
				Enabled: false, // Disable email validation for benchmarking
			},
		},
	}

	// Create config with email disabled for benchmarking
	cfg.Email.Provider = "" // Disable email service for benchmarking

	return NewAuthHandler(
		cfg,
		nil, // No real DB
		mockUserDAO,
		mockPseudonymDAO,
		mockIdentityMappingDAO,
		mockRoleKeyDAO,
		ibeSystem,
		mockSubforumDAO,
		mockPermissionDAO,
		nil, // No email service for benchmarking
		mockEmailVerificationTokenDAO,
		mockPasswordResetTokenDAO,
	)
}

// Benchmark complete user registration flow
func BenchmarkAuthHandler_RegisterUser(b *testing.B) {
	handler := createBenchmarkAuthHandler()

	input := &models.UserRegistrationInput{
		Body: models.UserRegistrationBody{
			Email:       "test@example.com",
			Password:    "secure_password_123",
			DisplayName: "Test User",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.RegisterUser(context.Background(), input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// createBenchmarkAuthHandlerForLogin creates a handler configured for login benchmarks
func createBenchmarkAuthHandlerForLogin() *AuthHandler {
	// Create mock DAOs with realistic behavior
	mockUserDAO := &mocks.MockUserDAO{}
	mockPseudonymDAO := &mocks.MockPseudonymDAO{}
	mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}
	mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
	mockSubforumDAO := &mocks.MockSubforumDAO{}
	mockPermissionDAO := &mocks.MockPermissionDAO{}
	mockEmailVerificationTokenDAO := &mocks.MockEmailVerificationTokenDAO{}
	mockPasswordResetTokenDAO := &mocks.MockPasswordResetTokenDAO{}

	// Set up realistic mock responses for login
	// Generate a proper SHA-256 hash for "secure_password_123"
	mockUserDAO.On("GetUserByEmail", mock.Anything, "test@example.com").Return(&dbmodels.User{
		UserID:        1,
		Email:         "test@example.com",
		PasswordHash:  "486f0a9839674a9dfcf52efbffe2469b01b87f5e5c967ec8f0ee3b7d9a308a8c",
		IsActive:      sql.Null[bool]{V: true, Valid: true},
		EmailVerified: sql.Null[bool]{V: true, Valid: true},
	}, nil)

	mockUserDAO.On("UpdateLastActive", mock.Anything, int64(1)).Return(nil)

	// Mock for email verification
	mockUserDAO.On("UpdateUser", mock.Anything, int64(1), mock.Anything).Return(nil)

	// Mock for password reset
	mockUserDAO.On("GetUserByID", mock.Anything, int64(1)).Return(&dbmodels.User{
		UserID:        1,
		Email:         "test@example.com",
		PasswordHash:  "486f0a9839674a9dfcf52efbffe2469b01b87f5e5c967ec8f0ee3b7d9a308a8c",
		IsActive:      sql.Null[bool]{V: true, Valid: true},
		EmailVerified: sql.Null[bool]{V: true, Valid: true},
	}, nil)

	mockPseudonymDAO.On("CreatePseudonymWithIdentityMapping", mock.Anything, mock.Anything, mock.Anything).Return(&dbmodels.Pseudonym{
		PseudonymID: "pseudonym_123",
		DisplayName: "Test User",
	}, nil)

	// Mock pseudonym retrieval for login
	mockPseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, int64(1), "user", "authentication").Return([]*dbmodels.Pseudonym{
		{
			PseudonymID: "pseudonym_123",
			DisplayName: "Test User",
			IsDefault:   true,
		},
	}, nil)

	mockPseudonymDAO.On("GetDefaultPseudonymByUserID", mock.Anything, int64(1), "user", "authentication").Return(&dbmodels.Pseudonym{
		PseudonymID: "pseudonym_123",
		DisplayName: "Test User",
		IsDefault:   true,
	}, nil)

	mockRoleKeyDAO.On("EnsureDefaultKeys", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockRoleKeyDAO.On("ListRoleKeysByPseudonym", mock.Anything, mock.Anything).Return([]*dbmodels.RoleKey{}, nil)

	mockEmailVerificationTokenDAO.On("CreateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockEmailVerificationTokenDAO.On("GetToken", mock.Anything, mock.Anything).Return(&dbmodels.EmailVerificationToken{
		Token:     "valid_token",
		UserID:    1,
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil)
	mockEmailVerificationTokenDAO.On("MarkTokenAsUsed", mock.Anything, mock.Anything).Return(nil)

	mockPasswordResetTokenDAO.On("CreateToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
	mockPasswordResetTokenDAO.On("GetToken", mock.Anything, mock.Anything).Return(&dbmodels.PasswordResetToken{
		Token:     "valid_reset_token",
		UserID:    1,
		ExpiresAt: time.Now().Add(time.Hour),
	}, nil)
	mockPasswordResetTokenDAO.On("MarkTokenAsUsed", mock.Anything, mock.Anything).Return(nil)

	// Generate realistic 32-byte domain master keys (AES-256 equivalent)
	domainMasters := make(map[string][]byte)
	domains := []string{
		ibe.DOMAIN_USER_PSEUDONYMS,
		ibe.DOMAIN_USER_CORRELATION,
		ibe.DOMAIN_MOD_CORRELATION,
		ibe.DOMAIN_ADMIN_CORRELATION,
		ibe.DOMAIN_LEGAL_CORRELATION,
	}

	for _, domain := range domains {
		key := make([]byte, 32) // 256-bit keys for production
		rand.Read(key)
		domainMasters[domain] = key
	}

	// Create IBE system with production key sizes
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
		DomainMasters: domainMasters,
		KeyVersion:    1,
		Salt:          "production_fingerprint_salt_v1_secure_random_string",
	})

	// Create config with production JWT secret and disabled email validation
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "production_jwt_secret_256_bit_key_for_benchmarking_secure_random_string_32_bytes_long",
			Expiration: 24 * time.Hour,
		},
		Email: config.EmailConfig{
			Provider: "", // Disable email service for benchmarking
			Validation: config.EmailValidationConfig{
				Enabled: false, // Disable email validation for benchmarking
			},
		},
	}

	return NewAuthHandler(
		cfg,
		nil, // No real DB
		mockUserDAO,
		mockPseudonymDAO,
		mockIdentityMappingDAO,
		mockRoleKeyDAO,
		ibeSystem,
		mockSubforumDAO,
		mockPermissionDAO,
		nil, // No email service for benchmarking
		mockEmailVerificationTokenDAO,
		mockPasswordResetTokenDAO,
	)
}

// Benchmark complete user login flow
func BenchmarkAuthHandler_LoginUser(b *testing.B) {
	handler := createBenchmarkAuthHandlerForLogin()

	input := &models.UserLoginInput{
		Body: models.UserLoginBody{
			Email:    "test@example.com",
			Password: "secure_password_123",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.LoginUser(context.Background(), input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark email verification flow
func BenchmarkAuthHandler_VerifyEmail(b *testing.B) {
	handler := createBenchmarkAuthHandler()

	input := &models.EmailVerificationInput{
		Body: models.EmailVerificationBody{
			Token: "valid_token",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.VerifyEmail(context.Background(), input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark password reset request flow
func BenchmarkAuthHandler_RequestPasswordReset(b *testing.B) {
	handler := createBenchmarkAuthHandler()

	input := &models.PasswordResetRequestInput{
		Body: models.PasswordResetRequestBody{
			Email: "test@example.com",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.RequestPasswordReset(context.Background(), input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark password reset flow
func BenchmarkAuthHandler_ResetPassword(b *testing.B) {
	handler := createBenchmarkAuthHandler()

	input := &models.PasswordResetInput{
		Body: models.PasswordResetBody{
			Token:    "valid_reset_token",
			Password: "new_secure_password_123",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.ResetPassword(context.Background(), input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark token refresh flow
func BenchmarkAuthHandler_RefreshToken(b *testing.B) {
	handler := createBenchmarkAuthHandler()

	// Create a valid refresh token
	userCtx := &middleware.UserContext{
		UserID:            1,
		Email:             "test@example.com",
		ActivePseudonymID: "pseudonym_123",
		MFAEnabled:        false,
	}

	refreshToken, err := middleware.GenerateJWT(userCtx, handler.config.JWT.Secret, 7*24*time.Hour)
	if err != nil {
		b.Fatal(err)
	}

	input := &struct {
		RefreshToken string `cookie:"refresh_token"`
		Body         models.RefreshTokenBody
	}{
		RefreshToken: refreshToken,
		Body:         models.RefreshTokenBody{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.RefreshToken(context.Background(), input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark current user session retrieval
func BenchmarkAuthHandler_GetCurrentUserSession(b *testing.B) {
	handler := createBenchmarkAuthHandler()

	// Create a valid access token
	userCtx := &middleware.UserContext{
		UserID:            1,
		Email:             "test@example.com",
		ActivePseudonymID: "pseudonym_123",
		MFAEnabled:        false,
	}

	accessToken, err := middleware.GenerateJWT(userCtx, handler.config.JWT.Secret, time.Hour)
	if err != nil {
		b.Fatal(err)
	}

	input := &middleware.AuthInput{
		AccessToken: accessToken,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.GetCurrentUserSession(context.Background(), input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark pseudonym switching
func BenchmarkAuthHandler_SwitchPseudonym(b *testing.B) {
	handler := createBenchmarkAuthHandler()

	// Create a valid access token
	userCtx := &middleware.UserContext{
		UserID:            1,
		Email:             "test@example.com",
		ActivePseudonymID: "pseudonym_123",
		MFAEnabled:        false,
	}

	accessToken, err := middleware.GenerateJWT(userCtx, handler.config.JWT.Secret, time.Hour)
	if err != nil {
		b.Fatal(err)
	}

	input := &struct {
		middleware.AuthInput
		models.SwitchPseudonymInput
	}{
		AuthInput: middleware.AuthInput{
			AccessToken: accessToken,
		},
		SwitchPseudonymInput: models.SwitchPseudonymInput{
			Body: models.SwitchPseudonymBody{
				PseudonymID: "new_pseudonym_456",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.SwitchPseudonym(context.Background(), input)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark different password strengths
func BenchmarkAuthHandler_RegisterUser_DifferentPasswordStrengths(b *testing.B) {
	handler := createBenchmarkAuthHandler()

	passwords := []string{
		"weak_password",
		"medium_strength_password_123",
		"very_strong_password_with_many_characters_and_numbers_1234567890!@#$%^&*()",
	}

	for _, password := range passwords {
		b.Run(fmt.Sprintf("%d_chars", len(password)), func(b *testing.B) {
			input := &models.UserRegistrationInput{
				Body: models.UserRegistrationBody{
					Email:       "test@example.com",
					Password:    password,
					DisplayName: "Test User",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := handler.RegisterUser(context.Background(), input)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark different email addresses
func BenchmarkAuthHandler_RegisterUser_DifferentEmails(b *testing.B) {
	handler := createBenchmarkAuthHandler()

	emails := []string{
		"short@test.com",
		"medium_length_email@example.com",
		"very_long_email_address_with_many_characters@very_long_domain_name.com",
	}

	for _, email := range emails {
		b.Run(fmt.Sprintf("%d_chars", len(email)), func(b *testing.B) {
			input := &models.UserRegistrationInput{
				Body: models.UserRegistrationBody{
					Email:       email,
					Password:    "secure_password_123",
					DisplayName: "Test User",
				},
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := handler.RegisterUser(context.Background(), input)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
