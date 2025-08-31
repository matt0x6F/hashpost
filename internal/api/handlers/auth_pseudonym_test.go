package handlers_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

// TestAuthHandler_DeactivatePseudonym_Gomock tests the pseudonym deactivation functionality using gomock
func TestAuthHandler_DeactivatePseudonym_Gomock(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler and mocks for this specific test
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

		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, nil, mockPseudonymDAO, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

		ctx := context.Background()

		// Set up mock expectations
		mockPseudonymDAO.EXPECT().DeactivatePseudonym(ctx, "test-pseudonym-123", int64(1), "active-pseudonym-456", "user", "self_correlation").Return(nil).Times(1)

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
	})

	t.Run("MissingPseudonymID", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler and mocks for this specific test
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

		ctx := context.Background()

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
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler and mocks for this specific test
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

		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, nil, mockPseudonymDAO, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

		ctx := context.Background()

		// Set up mock expectations to return ownership error
		mockPseudonymDAO.EXPECT().DeactivatePseudonym(gomock.Any(), "test-pseudonym-123", int64(1), "active-pseudonym-456", "user", "self_correlation").Return(fmt.Errorf("does not own")).Times(1)

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

		
	})

	t.Run("PseudonymNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler and mocks for this specific test
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

		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, nil, mockPseudonymDAO, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

		ctx := context.Background()

		// Set up mock expectations to return not found error
		mockPseudonymDAO.EXPECT().DeactivatePseudonym(gomock.Any(), "non-existent-pseudonym", int64(1), "test-pseudonym-123", "user", "self_correlation").Return(fmt.Errorf("not found")).Times(1)

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

		
	})

	t.Run("DatabaseError", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler and mocks for this specific test
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

		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, nil, mockPseudonymDAO, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

		ctx := context.Background()

		// Set up mock expectations to return database error
		mockPseudonymDAO.EXPECT().DeactivatePseudonym(gomock.Any(), "test-pseudonym-123", int64(1), "active-pseudonym-456", "user", "self_correlation").Return(fmt.Errorf("database connection failed")).Times(1)

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

		
	})
}

// TestAuthHandler_DeactivatePseudonym_Simple_Gomock tests the simple pseudonym deactivation functionality using gomock
func TestAuthHandler_DeactivatePseudonym_Simple_Gomock(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("SimpleSuccess", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Create handler and mocks for this specific test
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

		mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

		handler := handlers.NewAuthHandler(cfg, nil, nil, mockPseudonymDAO, nil, nil, ibeSystem, nil, nil, nil, nil, nil)

		ctx := context.Background()

		// Set up mock expectations
		mockPseudonymDAO.EXPECT().DeactivatePseudonym(ctx, "test-pseudonym-123", int64(1), "active-pseudonym-456", "user", "self_correlation").Return(nil).Times(1)

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
		
	})
}
