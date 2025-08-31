package handlers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	gomock "go.uber.org/mock/gomock"
)

// TestAuthHandler_RegisterUser_PermissionWorkflow tests the complete user registration workflow
// and verifies that proper permissions are set up for new users using gomock
func TestAuthHandler_RegisterUser_PermissionWorkflow(t *testing.T) {
	t.Run("UserRegistration_CreatesUserWithProperPermissions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Arrange
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, _, mockPermissionDAO := NewAuthHandlerWithGomocks(ctrl)

		// Test data
		email := "newuser@example.com"
		password := "SecurePassword123!"
		displayName := "NewUser"
		userID := int64(1)
		pseudonymID := "new-user-pseudo-1"

		// Mock email uniqueness check (should return error if user doesn't exist)
		mockUserDAO.EXPECT().
			GetUserByEmail(gomock.Any(), email).
			Return(nil, assert.AnError). // User doesn't exist, so error is expected
			Times(1)

		// Mock user creation
		mockUserDAO.EXPECT().
			CreateUser(gomock.Any(), email, gomock.Any()).
			Return(&dbmodels.User{
				UserID: userID,
				Email:  email,
			}, nil).
			Times(1)

		// Mock pseudonym creation with identity mapping
		mockPseudonymDAO.EXPECT().
			CreatePseudonymWithIdentityMapping(gomock.Any(), userID, displayName).
			Return(&dbmodels.Pseudonym{
				PseudonymID: pseudonymID,
				DisplayName: displayName,
			}, nil).
			Times(1)

		// Mock role key creation - this is where we verify proper permission setup
		mockRoleKeyDAO.EXPECT().
			EnsureDefaultKeys(gomock.Any(), gomock.Any(), pseudonymID, []string{constants.RoleUser}).
			Return(nil).
			Times(1)

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

		mockPermissionDAO.EXPECT().
			GetUnifiedActivePseudonymRolesAndCapabilities(
				gomock.Any(), userID, pseudonymID, (*int32)(nil)).
			Return([]string{constants.RoleUser}, expectedNewUserCapabilities, nil).
			Times(1)

		// Mock specific capability checks for key workflows
		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(
				gomock.Any(), userID, pseudonymID, constants.CapabilityCreateSubforum, (*int32)(nil)).
			Return(true, nil).
			Times(1)

		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(
				gomock.Any(), userID, pseudonymID, constants.CapabilityManageOwnProfile, (*int32)(nil)).
			Return(true, nil).
			Times(1)

		// Mock that new users should NOT have admin capabilities
		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(
				gomock.Any(), userID, pseudonymID, constants.CapabilitySystemAdmin, (*int32)(nil)).
			Return(false, nil).
			Times(1)

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
	})

	t.Run("UserRegistration_VerifiesAuthenticationAndSelfCorrelationScopes", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// This test focuses specifically on verifying that the registration process
		// creates role keys with the correct scopes (authentication and self_correlation)
		handler, mockUserDAO, mockPseudonymDAO, _, mockRoleKeyDAO, _, mockPermissionDAO := NewAuthHandlerWithGomocks(ctrl)

		email := "scopetest@example.com"
		password := "SecurePassword123!"
		displayName := "ScopeTestUser"
		userID := int64(2)
		pseudonymID := "scope-test-pseudo-2"

		// Mock the creation steps
		mockUserDAO.EXPECT().
			GetUserByEmail(gomock.Any(), email).
			Return(nil, assert.AnError). // User doesn't exist
			Times(1)

		mockUserDAO.EXPECT().
			CreateUser(gomock.Any(), email, gomock.Any()).
			Return(&dbmodels.User{UserID: userID, Email: email}, nil).
			Times(1)

		mockPseudonymDAO.EXPECT().
			CreatePseudonymWithIdentityMapping(gomock.Any(), userID, displayName).
			Return(&dbmodels.Pseudonym{PseudonymID: pseudonymID, DisplayName: displayName}, nil).
			Times(1)

		// This is the critical mock - verify that EnsureDefaultKeys is called to create
		// role keys with the proper scopes for a user role
		mockRoleKeyDAO.EXPECT().
			EnsureDefaultKeys(gomock.Any(), gomock.Any(), pseudonymID, []string{constants.RoleUser}).
			Return(nil).
			Times(1)

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
			mockPermissionDAO.EXPECT().
				HasUnifiedCapability(
					gomock.Any(), userID, pseudonymID, capability, (*int32)(nil)).
				Return(true, nil).
				Times(1)
		}

		for _, capability := range selfCorrelationScopeCapabilities {
			mockPermissionDAO.EXPECT().
				HasUnifiedCapability(
					gomock.Any(), userID, pseudonymID, capability, (*int32)(nil)).
				Return(true, nil).
				Times(1)
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
	})
}
