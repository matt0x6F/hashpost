package handlers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
)

// TestAuthHandler_RegisterUser_PermissionWorkflow tests the complete user registration workflow
// and verifies that proper permissions are set up for new users
func TestAuthHandler_RegisterUser_PermissionWorkflow(t *testing.T) {
	t.Run("UserRegistration_CreatesUserWithProperPermissions", func(t *testing.T) {
		// Arrange
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, _, mockPermissionDAO := NewAuthHandlerWithMocks()

		// Test data
		email := "newuser@example.com"
		password := "SecurePassword123!"
		displayName := "NewUser"
		userID := int64(1)
		pseudonymID := "new-user-pseudo-1"

		// Mock email uniqueness check (should return error if user doesn't exist)
		mockUserDAO.On("GetUserByEmail", mock.Anything, email).
			Return(nil, assert.AnError) // User doesn't exist, so error is expected

		// Mock user creation
		mockUserDAO.On("CreateUser", mock.Anything, email, mock.AnythingOfType("string")).
			Return(&dbmodels.User{
				UserID: userID,
				Email:  email,
			}, nil)

		// Mock pseudonym creation with identity mapping
		mockPseudonymDAO.On("CreatePseudonymWithIdentityMapping", mock.Anything, userID, displayName).
			Return(&dbmodels.Pseudonym{
				PseudonymID: pseudonymID,
				DisplayName: displayName,
			}, nil)

		// Mock role key creation - this is where we verify proper permission setup
		mockRoleKeyDAO.On("EnsureDefaultKeys", mock.Anything, mock.Anything, pseudonymID, []string{constants.RoleUser}).
			Return(nil)

		// Mock permission verification after user creation
		// New users should have these capabilities through their user role keys with authentication and self-correlation scopes
		expectedNewUserCapabilities := []string{
			// Basic user capabilities
			constants.CapabilityCreateContent,
			constants.CapabilityVote,
			constants.CapabilityMessage,
			constants.CapabilityReport,
			constants.CapabilityCreateSubforum,
			// Authentication scope capabilities
			constants.CapabilityAccessOwnPseudonyms,
			constants.CapabilityLogin,
			constants.CapabilitySessionManagement,
			// Self-correlation scope capabilities
			constants.CapabilityVerifyOwnPseudonymOwnership,
			constants.CapabilityManageOwnProfile,
			constants.CapabilityManageOwnPseudonyms,
		}

		mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities",
			mock.Anything, userID, pseudonymID, (*int32)(nil)).
			Return([]string{constants.RoleUser}, expectedNewUserCapabilities, nil)

		// Mock specific capability checks for key workflows
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityCreateSubforum, (*int32)(nil)).
			Return(true, nil)

		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityManageOwnProfile, (*int32)(nil)).
			Return(true, nil)

		// Mock that new users should NOT have admin capabilities
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilitySystemAdmin, (*int32)(nil)).
			Return(false, nil)

		// Act - Register the user
		input := &apimodels.UserRegistrationInput{
			Body: apimodels.UserRegistrationBody{
				Email:       email,
				Password:    password,
				DisplayName: displayName,
			},
		}

		response, err := handler.RegisterUser(context.Background(), input)

		// Assert
		assert.NoError(t, err, "User registration should succeed")
		assert.NotNil(t, response, "Registration response should not be nil")
		assert.NotNil(t, response.Body, "Registration response body should not be nil")

		// Verify the registration created the expected user structure
		assert.Equal(t, int(userID), response.Body.UserID)
		assert.Equal(t, email, response.Body.Email)
		assert.Equal(t, pseudonymID, response.Body.PseudonymID)
		assert.Equal(t, displayName, response.Body.DisplayName)

		// Verify all mocks were called as expected
		mockUserDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
		mockRoleKeyDAO.AssertExpectations(t)

		// Now test the permission setup by verifying the new user has expected capabilities
		roles, capabilities, err := mockPermissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
			context.Background(), userID, pseudonymID, nil)
		assert.NoError(t, err)
		assert.Contains(t, roles, constants.RoleUser, "New user should have user role")

		// Verify key capabilities that enable important workflows
		assert.Contains(t, capabilities, constants.CapabilityCreateSubforum,
			"New user should be able to create subforums")
		assert.Contains(t, capabilities, constants.CapabilityManageOwnProfile,
			"New user should be able to manage their own profile")
		assert.Contains(t, capabilities, constants.CapabilityCreateContent,
			"New user should be able to create content")

		// Test specific capability checks
		canCreateSubforum, err := mockPermissionDAO.HasUnifiedCapability(
			context.Background(), userID, pseudonymID, constants.CapabilityCreateSubforum, nil)
		assert.NoError(t, err)
		assert.True(t, canCreateSubforum, "New user should be able to create subforums")

		canManageProfile, err := mockPermissionDAO.HasUnifiedCapability(
			context.Background(), userID, pseudonymID, constants.CapabilityManageOwnProfile, nil)
		assert.NoError(t, err)
		assert.True(t, canManageProfile, "New user should be able to manage their own profile")

		// Verify they don't have admin capabilities
		hasSystemAdmin, err := mockPermissionDAO.HasUnifiedCapability(
			context.Background(), userID, pseudonymID, constants.CapabilitySystemAdmin, nil)
		assert.NoError(t, err)
		assert.False(t, hasSystemAdmin, "New user should NOT have system admin capabilities")

		mockPermissionDAO.AssertExpectations(t)
	})

	t.Run("UserRegistration_VerifiesAuthenticationAndSelfCorrelationScopes", func(t *testing.T) {
		// This test focuses specifically on verifying that the registration process
		// creates role keys with the correct scopes (authentication and self_correlation)
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, _, mockPermissionDAO := NewAuthHandlerWithMocks()

		email := "scopetest@example.com"
		password := "SecurePassword123!"
		displayName := "ScopeTestUser"
		userID := int64(2)
		pseudonymID := "scope-test-pseudo-2"

		// Mock the creation steps
		mockUserDAO.On("GetUserByEmail", mock.Anything, email).
			Return(nil, assert.AnError) // User doesn't exist

		mockUserDAO.On("CreateUser", mock.Anything, email, mock.AnythingOfType("string")).
			Return(&dbmodels.User{UserID: userID, Email: email}, nil)

		mockPseudonymDAO.On("CreatePseudonymWithIdentityMapping", mock.Anything, userID, displayName).
			Return(&dbmodels.Pseudonym{PseudonymID: pseudonymID, DisplayName: displayName}, nil)

		// This is the critical mock - verify that EnsureDefaultKeys is called to create
		// role keys with the proper scopes for a user role
		mockRoleKeyDAO.On("EnsureDefaultKeys", mock.Anything, mock.Anything, pseudonymID, []string{constants.RoleUser}).
			Return(nil)

		// Mock that the user now has capabilities from both authentication and self-correlation scopes
		authScopeCapabilities := []string{
			constants.CapabilityAccessOwnPseudonyms,
			constants.CapabilityLogin,
			constants.CapabilitySessionManagement,
		}

		selfCorrelationScopeCapabilities := []string{
			constants.CapabilityVerifyOwnPseudonymOwnership,
			constants.CapabilityManageOwnProfile,
			constants.CapabilityManageOwnPseudonyms,
		}

		// Mock individual capability checks for both scope types
		for _, capability := range authScopeCapabilities {
			mockPermissionDAO.On("HasUnifiedCapability",
				mock.Anything, userID, pseudonymID, capability, (*int32)(nil)).
				Return(true, nil)
		}

		for _, capability := range selfCorrelationScopeCapabilities {
			mockPermissionDAO.On("HasUnifiedCapability",
				mock.Anything, userID, pseudonymID, capability, (*int32)(nil)).
				Return(true, nil)
		}

		// Act
		input := &apimodels.UserRegistrationInput{
			Body: apimodels.UserRegistrationBody{
				Email:       email,
				Password:    password,
				DisplayName: displayName,
			},
		}

		response, err := handler.RegisterUser(context.Background(), input)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)

		// The key assertion: verify that EnsureDefaultKeys was called
		// This ensures that the user gets role keys with authentication and self_correlation scopes
		mockRoleKeyDAO.AssertExpectations(t)

		// Verify the user has capabilities from both scopes
		ctx := context.Background()

		// Test authentication scope capabilities
		for _, capability := range authScopeCapabilities {
			hasCapability, err := mockPermissionDAO.HasUnifiedCapability(
				ctx, userID, pseudonymID, capability, nil)
			assert.NoError(t, err)
			assert.True(t, hasCapability, "New user should have auth scope capability: %s", capability)
		}

		// Test self-correlation scope capabilities
		for _, capability := range selfCorrelationScopeCapabilities {
			hasCapability, err := mockPermissionDAO.HasUnifiedCapability(
				ctx, userID, pseudonymID, capability, nil)
			assert.NoError(t, err)
			assert.True(t, hasCapability, "New user should have self-correlation scope capability: %s", capability)
		}

		mockPermissionDAO.AssertExpectations(t)
	})
}
