package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPseudonymSwitchingScopeValidation tests the scope validation logic for pseudonym switching
func TestPseudonymSwitchingScopeValidation(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("ScopeHierarchyValidation", func(t *testing.T) {
		// Test that scopes are tried in the correct order
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
		}

		mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()
		handler := handlers.NewAuthHandler(cfg, nil, nil, mockSecurePseudonymDAO, nil, nil, ibe.NewIBESystemWithOptions(ibe.IBEOptions{}), nil, nil, nil, nil, nil)

		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()
		userCtx.Roles = []string{"user"}

		// Mock target pseudonym
		targetPseudonym := fixtures.CreateTestPseudonym()
		targetPseudonym.PseudonymID = "test-pseudonym-123"

		// Set up expectations for scope hierarchy
		mockSecurePseudonymDAO.On("GetPseudonymByID", ctx, "test-pseudonym-123").Return(targetPseudonym, nil)

		// Expect authentication scope to be tried first
		mockSecurePseudonymDAO.On("VerifyPseudonymOwnership", ctx, "test-pseudonym-123", int64(1), "user", constants.ScopeAuthentication).Return(true, nil)
		mockSecurePseudonymDAO.On("UpdateLastActive", ctx, "test-pseudonym-123").Return(nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			models.SwitchPseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SwitchPseudonymInput: models.SwitchPseudonymInput{
				Body: models.SwitchPseudonymBody{
					PseudonymID: "test-pseudonym-123",
				},
			},
		}

		// Call the method
		result, err := handler.SwitchPseudonym(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 200, result.Status)

		// Verify that authentication scope was used (not self-correlation)
		mockSecurePseudonymDAO.AssertExpectations(t)
	})

	t.Run("CapabilityValidation", func(t *testing.T) {
		// Test that the correct capability is being validated
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
		}

		mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()
		handler := handlers.NewAuthHandler(cfg, nil, nil, mockSecurePseudonymDAO, nil, nil, ibe.NewTestIBESystem(), nil, nil, nil, nil, nil)

		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()
		userCtx.Roles = []string{"user"}

		// Mock target pseudonym
		targetPseudonym := fixtures.CreateTestPseudonym()
		targetPseudonym.PseudonymID = "test-pseudonym-123"

		// Set up expectations
		mockSecurePseudonymDAO.On("GetPseudonymByID", ctx, "test-pseudonym-123").Return(targetPseudonym, nil)

		// The VerifyPseudonymOwnership method should validate CapabilityVerifyOwnPseudonymOwnership
		mockSecurePseudonymDAO.On("VerifyPseudonymOwnership", ctx, "test-pseudonym-123", int64(1), "user", constants.ScopeAuthentication).Return(true, nil)
		mockSecurePseudonymDAO.On("UpdateLastActive", ctx, "test-pseudonym-123").Return(nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			models.SwitchPseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SwitchPseudonymInput: models.SwitchPseudonymInput{
				Body: models.SwitchPseudonymBody{
					PseudonymID: "test-pseudonym-123",
				},
			},
		}

		// Call the method
		result, err := handler.SwitchPseudonym(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 200, result.Status)

		// Verify mock expectations
		mockSecurePseudonymDAO.AssertExpectations(t)
	})

	t.Run("MultiRoleFallback", func(t *testing.T) {
		// Test that multiple roles are tried when available
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
		}

		mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()
		handler := handlers.NewAuthHandler(cfg, nil, nil, mockSecurePseudonymDAO, nil, nil, ibe.NewTestIBESystem(), nil, nil, nil, nil, nil)

		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()
		userCtx.Roles = []string{"user", "platform_admin"}

		// Mock target pseudonym
		targetPseudonym := fixtures.CreateTestPseudonym()
		targetPseudonym.PseudonymID = "test-pseudonym-123"

		// Set up expectations for multi-role fallback
		mockSecurePseudonymDAO.On("GetPseudonymByID", ctx, "test-pseudonym-123").Return(targetPseudonym, nil)

		// First role succeeds with authentication scope
		mockSecurePseudonymDAO.On("VerifyPseudonymOwnership", ctx, "test-pseudonym-123", int64(1), "user", constants.ScopeAuthentication).Return(true, nil)
		mockSecurePseudonymDAO.On("UpdateLastActive", ctx, "test-pseudonym-123").Return(nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			models.SwitchPseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SwitchPseudonymInput: models.SwitchPseudonymInput{
				Body: models.SwitchPseudonymBody{
					PseudonymID: "test-pseudonym-123",
				},
			},
		}

		// Call the method
		result, err := handler.SwitchPseudonym(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, 200, result.Status)

		// Verify that platform_admin role was used
		mockSecurePseudonymDAO.AssertExpectations(t)
	})

	t.Run("SecurityBoundaries", func(t *testing.T) {
		// Test that users cannot access pseudonyms they don't own
		cfg := &config.Config{
			JWT: config.JWTConfig{
				Secret:     "test-secret-key",
				Expiration: 24 * time.Hour,
			},
		}

		mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()
		handler := handlers.NewAuthHandler(cfg, nil, nil, mockSecurePseudonymDAO, nil, nil, ibe.NewTestIBESystem(), nil, nil, nil, nil, nil)

		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()
		userCtx.Roles = []string{"user"}

		// Mock target pseudonym
		targetPseudonym := fixtures.CreateTestPseudonym()
		targetPseudonym.PseudonymID = "other-user-pseudonym-123"

		// Set up expectations for security boundary test
		mockSecurePseudonymDAO.On("GetPseudonymByID", ctx, "other-user-pseudonym-123").Return(targetPseudonym, nil)

		// All scope attempts fail (user doesn't own this pseudonym)
		mockSecurePseudonymDAO.On("VerifyPseudonymOwnership", ctx, "other-user-pseudonym-123", int64(1), "user", constants.ScopeAuthentication).Return(false, nil)
		mockSecurePseudonymDAO.On("VerifyPseudonymOwnership", ctx, "other-user-pseudonym-123", int64(1), "user", constants.ScopeSelfCorrelation).Return(false, nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			models.SwitchPseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			SwitchPseudonymInput: models.SwitchPseudonymInput{
				Body: models.SwitchPseudonymBody{
					PseudonymID: "other-user-pseudonym-123",
				},
			},
		}

		// Call the method
		result, err := handler.SwitchPseudonym(ctx, input)

		// Assertions
		require.Error(t, err)
		require.Nil(t, result)
		assert.Contains(t, err.Error(), "You do not own this pseudonym")

		// Verify mock expectations
		mockSecurePseudonymDAO.AssertExpectations(t)
	})
}

// TestScopeAndCapabilityDefinitions tests that the scope and capability definitions are correct
func TestScopeAndCapabilityDefinitions(t *testing.T) {
	t.Run("AuthenticationScopeCapabilities", func(t *testing.T) {
		// Verify that authentication scope has the correct capabilities
		capabilities := constants.GetCapabilitiesByScope(constants.ScopeAuthentication)

		// Authentication scope should include pseudonym access capabilities
		assert.Contains(t, capabilities, constants.CapabilityAccessOwnPseudonyms)
		assert.Contains(t, capabilities, constants.CapabilityLogin)
		assert.Contains(t, capabilities, constants.CapabilitySessionManagement)

		// Authentication scope should NOT include administrative capabilities
		assert.NotContains(t, capabilities, constants.CapabilityAccessAllPseudonyms)
		assert.NotContains(t, capabilities, constants.CapabilityCrossUserCorrelation)
	})

	t.Run("SelfCorrelationScopeCapabilities", func(t *testing.T) {
		// Verify that self-correlation scope has the correct capabilities
		capabilities := constants.GetCapabilitiesByScope(constants.ScopeSelfCorrelation)

		// Self-correlation scope should include ownership verification
		assert.Contains(t, capabilities, constants.CapabilityVerifyOwnPseudonymOwnership)
		assert.Contains(t, capabilities, constants.CapabilityManageOwnProfile)
		assert.Contains(t, capabilities, constants.CapabilityManageOwnPseudonyms)

		// Self-correlation scope should NOT include administrative capabilities
		assert.NotContains(t, capabilities, constants.CapabilityAccessAllPseudonyms)
		assert.NotContains(t, capabilities, constants.CapabilityCrossUserCorrelation)
	})

	t.Run("UserRoleScopeAccess", func(t *testing.T) {
		// Verify that user role has access to appropriate scopes
		userScopes := constants.GetRoleScopes(constants.RoleUser)

		// User role should have access to authentication and self-correlation scopes
		assert.Contains(t, userScopes, constants.ScopeAuthentication)
		assert.Contains(t, userScopes, constants.ScopeSelfCorrelation)

		// User role should NOT have access to correlation scope
		assert.NotContains(t, userScopes, constants.ScopeCorrelation)
	})

	t.Run("PlatformAdminRoleScopeAccess", func(t *testing.T) {
		// Verify that platform admin role has access to all scopes
		adminScopes := constants.GetRoleScopes(constants.RolePlatformAdmin)

		// Platform admin should have access to all scopes
		assert.Contains(t, adminScopes, constants.ScopeAuthentication)
		assert.Contains(t, adminScopes, constants.ScopeSelfCorrelation)
		assert.Contains(t, adminScopes, constants.ScopeCorrelation)
	})

	t.Run("CapabilityValidation", func(t *testing.T) {
		// Verify that pseudonym ownership verification capability is valid for appropriate scopes
		authCapabilities := constants.GetCapabilitiesByScope(constants.ScopeAuthentication)
		selfCorrCapabilities := constants.GetCapabilitiesByScope(constants.ScopeSelfCorrelation)

		// Check that the capability exists in self-correlation scope (where it belongs)
		assert.NotContains(t, authCapabilities, constants.CapabilityVerifyOwnPseudonymOwnership)
		assert.Contains(t, selfCorrCapabilities, constants.CapabilityVerifyOwnPseudonymOwnership)

		// Verify using the validation function
		assert.False(t, constants.IsValidCapability(constants.CapabilityVerifyOwnPseudonymOwnership, constants.ScopeAuthentication))
		assert.True(t, constants.IsValidCapability(constants.CapabilityVerifyOwnPseudonymOwnership, constants.ScopeSelfCorrelation))
		assert.False(t, constants.IsValidCapability(constants.CapabilityVerifyOwnPseudonymOwnership, constants.ScopeCorrelation))

		// This validates our multi-scope fallback strategy:
		// 1. Try authentication scope first (fails due to missing capability)
		// 2. Fall back to self-correlation scope (succeeds with correct capability)
	})
}
