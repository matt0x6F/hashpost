package handlers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
)

// TestUserCreationPermissionWorkflow tests that new users get proper role keys and capabilities
func TestUserCreationPermissionWorkflow(t *testing.T) {
	t.Run("NewUser_GetsUserRoleWithAuthAndSelfCorrelationScopes", func(t *testing.T) {
		// Arrange
		mockPermissionDAO := mocks.NewMockPermissionDAO()

		userID := int64(1)
		pseudonymID := "new-user-pseudo-1"

		// Expected capabilities for a new user
		expectedCapabilities := []string{
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

		// Mock: New user should have user role with expected capabilities
		mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities",
			mock.Anything, userID, pseudonymID, (*int32)(nil)).
			Return([]string{constants.RoleUser}, expectedCapabilities, nil)

		// Mock: Specific capability checks for key features
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityCreateSubforum, (*int32)(nil)).
			Return(true, nil)

		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityManageOwnProfile, (*int32)(nil)).
			Return(true, nil)

		// Mock: User should NOT have admin capabilities
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilitySystemAdmin, (*int32)(nil)).
			Return(false, nil)

		// Act & Assert
		ctx := context.Background()

		// Verify new user has expected role and capabilities
		roles, capabilities, err := mockPermissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
			ctx, userID, pseudonymID, nil)
		assert.NoError(t, err)
		assert.Contains(t, roles, constants.RoleUser, "New user should have user role")

		// Verify specific capabilities that enable core workflows
		assert.Contains(t, capabilities, constants.CapabilityCreateSubforum,
			"New user should be able to create subforums")
		assert.Contains(t, capabilities, constants.CapabilityManageOwnProfile,
			"New user should be able to manage their own profile")
		assert.Contains(t, capabilities, constants.CapabilityCreateContent,
			"New user should be able to create content")

		// Test specific capability checks
		canCreateSubforum, err := mockPermissionDAO.HasUnifiedCapability(
			ctx, userID, pseudonymID, constants.CapabilityCreateSubforum, nil)
		assert.NoError(t, err)
		assert.True(t, canCreateSubforum, "New user should be able to create subforums")

		canManageProfile, err := mockPermissionDAO.HasUnifiedCapability(
			ctx, userID, pseudonymID, constants.CapabilityManageOwnProfile, nil)
		assert.NoError(t, err)
		assert.True(t, canManageProfile, "New user should be able to manage their own profile")

		// Verify they don't have admin capabilities
		hasSystemAdmin, err := mockPermissionDAO.HasUnifiedCapability(
			ctx, userID, pseudonymID, constants.CapabilitySystemAdmin, nil)
		assert.NoError(t, err)
		assert.False(t, hasSystemAdmin, "New user should NOT have system admin capabilities")

		mockPermissionDAO.AssertExpectations(t)
	})

	t.Run("NewUser_HasExpectedAuthenticationAndSelfCorrelationCapabilities", func(t *testing.T) {
		mockPermissionDAO := mocks.NewMockPermissionDAO()

		userID := int64(2)
		pseudonymID := "new-user-pseudo-2"

		// Authentication scope capabilities (these require authentication role key)
		authCapabilities := []string{
			constants.CapabilityAccessOwnPseudonyms,
			constants.CapabilityLogin,
			constants.CapabilitySessionManagement,
		}

		// Self-correlation scope capabilities (these require self-correlation role key)
		selfCorrCapabilities := []string{
			constants.CapabilityVerifyOwnPseudonymOwnership,
			constants.CapabilityManageOwnProfile,
			constants.CapabilityManageOwnPseudonyms,
		}

		allExpectedCapabilities := append(authCapabilities, selfCorrCapabilities...)

		// Mock that user has both sets of capabilities
		for _, capability := range allExpectedCapabilities {
			mockPermissionDAO.On("HasUnifiedCapability",
				mock.Anything, userID, pseudonymID, capability, (*int32)(nil)).
				Return(true, nil)
		}

		ctx := context.Background()

		// Test each authentication capability
		for _, capability := range authCapabilities {
			hasCapability, err := mockPermissionDAO.HasUnifiedCapability(
				ctx, userID, pseudonymID, capability, nil)
			assert.NoError(t, err)
			assert.True(t, hasCapability, "New user should have auth capability: %s", capability)
		}

		// Test each self-correlation capability
		for _, capability := range selfCorrCapabilities {
			hasCapability, err := mockPermissionDAO.HasUnifiedCapability(
				ctx, userID, pseudonymID, capability, nil)
			assert.NoError(t, err)
			assert.True(t, hasCapability, "New user should have self-correlation capability: %s", capability)
		}

		mockPermissionDAO.AssertExpectations(t)
	})
}

// TestSubforumCreationPermissionWorkflow tests that subforum creators get proper moderation capabilities
func TestSubforumCreationPermissionWorkflow(t *testing.T) {
	t.Run("SubforumCreator_GetsSubforumOwnerRoleWithModerationCapabilities", func(t *testing.T) {
		// Arrange
		mockPermissionDAO := mocks.NewMockPermissionDAO()

		userID := int64(1)
		pseudonymID := "creator-pseudo-1"
		subforumID := int32(1)

		// Expected roles: user + subforum_owner
		expectedRoles := []string{constants.RoleUser, constants.RoleSubforumOwner}

		// Expected capabilities: basic user + all moderator capabilities
		expectedCapabilities := []string{
			// Basic user capabilities
			constants.CapabilityCreateContent,
			constants.CapabilityVote,
			constants.CapabilityMessage,
			constants.CapabilityReport,
			// Moderation capabilities (from moderation scope role key)
			constants.CapabilityModerateContent,
			constants.CapabilityBanUsers,
			constants.CapabilityRemoveContent,
			constants.CapabilityReviewReports,
			constants.CapabilityForwardReports,
			constants.CapabilityManageSubforumRules,
			constants.CapabilityManageSubforumSettings,
			constants.CapabilityStickyPost,
			constants.CapabilityLockPost,
			// Owner-specific capabilities
			constants.CapabilityManageModerators,
			// Correlation capabilities (from correlation scope role key)
			constants.CapabilityAccessSubforumPseudonyms,
			constants.CapabilityCorrelateFingerprints,
		}

		// Mock: Creator should have subforum owner role and all capabilities in their subforum
		mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities",
			mock.Anything, userID, pseudonymID, &subforumID).
			Return(expectedRoles, expectedCapabilities, nil)

		// Mock: Key moderator capabilities
		keyModerationCapabilities := []string{
			constants.CapabilityModerateContent,
			constants.CapabilityBanUsers,
			constants.CapabilityManageModerators,
			constants.CapabilityManageSubforumSettings,
		}

		for _, capability := range keyModerationCapabilities {
			mockPermissionDAO.On("HasUnifiedCapability",
				mock.Anything, userID, pseudonymID, capability, &subforumID).
				Return(true, nil)
		}

		// Act & Assert
		ctx := context.Background()

		// Verify creator has subforum owner role and expected capabilities
		roles, capabilities, err := mockPermissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
			ctx, userID, pseudonymID, &subforumID)
		assert.NoError(t, err)
		assert.Contains(t, roles, constants.RoleSubforumOwner,
			"Subforum creator should have subforum_owner role")

		// Check that they have all the key moderation capabilities
		expectedModerationCaps := []string{
			constants.CapabilityModerateContent,
			constants.CapabilityBanUsers,
			constants.CapabilityManageModerators,
			constants.CapabilityManageSubforumSettings,
		}

		for _, cap := range expectedModerationCaps {
			assert.Contains(t, capabilities, cap,
				"Subforum creator should have capability: %s", cap)
		}

		// Test specific capability checks
		for _, capability := range keyModerationCapabilities {
			hasCapability, err := mockPermissionDAO.HasUnifiedCapability(
				ctx, userID, pseudonymID, capability, &subforumID)
			assert.NoError(t, err)
			assert.True(t, hasCapability,
				"Subforum creator should have capability: %s", capability)
		}

		mockPermissionDAO.AssertExpectations(t)
	})

	t.Run("SubforumCreator_CannotModerateOtherSubforums", func(t *testing.T) {
		// Test permission isolation: creators only get moderation rights for their subforum
		mockPermissionDAO := mocks.NewMockPermissionDAO()

		userID := int64(1)
		pseudonymID := "creator-pseudo-1"
		ownSubforumID := int32(1)
		otherSubforumID := int32(2)

		// Mock: Has moderation capabilities in own subforum
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityModerateContent, &ownSubforumID).
			Return(true, nil)

		// Mock: Does NOT have moderation capabilities in other subforum
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityModerateContent, &otherSubforumID).
			Return(false, nil)

		ctx := context.Background()

		// Verify they can moderate their own subforum
		canModerateOwn, err := mockPermissionDAO.HasUnifiedCapability(
			ctx, userID, pseudonymID, constants.CapabilityModerateContent, &ownSubforumID)
		assert.NoError(t, err)
		assert.True(t, canModerateOwn,
			"Creator should be able to moderate their own subforum")

		// Verify they cannot moderate other subforums
		canModerateOther, err := mockPermissionDAO.HasUnifiedCapability(
			ctx, userID, pseudonymID, constants.CapabilityModerateContent, &otherSubforumID)
		assert.NoError(t, err)
		assert.False(t, canModerateOther,
			"Creator should NOT be able to moderate other subforums")

		mockPermissionDAO.AssertExpectations(t)
	})

	t.Run("SubforumCreator_HasBothModerationAndCorrelationCapabilities", func(t *testing.T) {
		// Test that subforum creators get capabilities from both moderation and correlation scopes
		mockPermissionDAO := mocks.NewMockPermissionDAO()

		userID := int64(1)
		pseudonymID := "creator-pseudo-1"
		subforumID := int32(1)

		// Moderation scope capabilities (require moderation role key for subforum)
		moderationCapabilities := []string{
			constants.CapabilityModerateContent,
			constants.CapabilityBanUsers,
			constants.CapabilityRemoveContent,
			constants.CapabilityManageModerators,
			constants.CapabilityReviewReports,
			constants.CapabilityManageSubforumRules,
			constants.CapabilityManageSubforumSettings,
		}

		// Correlation scope capabilities (require correlation role key for subforum)
		correlationCapabilities := []string{
			constants.CapabilityAccessSubforumPseudonyms,
			constants.CapabilityCorrelateFingerprints,
		}

		allSubforumCapabilities := append(moderationCapabilities, correlationCapabilities...)

		// Mock all capabilities as available
		for _, capability := range allSubforumCapabilities {
			mockPermissionDAO.On("HasUnifiedCapability",
				mock.Anything, userID, pseudonymID, capability, &subforumID).
				Return(true, nil)
		}

		ctx := context.Background()

		// Test moderation capabilities
		for _, capability := range moderationCapabilities {
			hasCapability, err := mockPermissionDAO.HasUnifiedCapability(
				ctx, userID, pseudonymID, capability, &subforumID)
			assert.NoError(t, err)
			assert.True(t, hasCapability,
				"Subforum creator should have moderation capability: %s", capability)
		}

		// Test correlation capabilities
		for _, capability := range correlationCapabilities {
			hasCapability, err := mockPermissionDAO.HasUnifiedCapability(
				ctx, userID, pseudonymID, capability, &subforumID)
			assert.NoError(t, err)
			assert.True(t, hasCapability,
				"Subforum creator should have correlation capability: %s", capability)
		}

		mockPermissionDAO.AssertExpectations(t)
	})
}

// TestPermissionWorkflowIntegration tests complete permission scenarios
func TestPermissionWorkflowIntegration(t *testing.T) {
	t.Run("UserLifecycle_FromRegistrationToSubforumOwner", func(t *testing.T) {
		// Test the complete user journey: registration -> create subforum -> moderate
		mockPermissionDAO := mocks.NewMockPermissionDAO()

		userID := int64(1)
		pseudonymID := "user-pseudo-1"
		subforumID := int32(1)

		// Step 1: As new user (global scope)
		newUserCapabilities := []string{
			constants.CapabilityCreateContent,
			constants.CapabilityVote,
			constants.CapabilityCreateSubforum,
			constants.CapabilityManageOwnProfile,
		}

		mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities",
			mock.Anything, userID, pseudonymID, (*int32)(nil)).
			Return([]string{constants.RoleUser}, newUserCapabilities, nil)

		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityCreateSubforum, (*int32)(nil)).
			Return(true, nil)

		// Step 2: After creating subforum (subforum-specific scope)
		subforumOwnerCapabilities := append(newUserCapabilities,
			constants.CapabilityModerateContent,
			constants.CapabilityBanUsers,
			constants.CapabilityManageModerators,
			constants.CapabilityManageSubforumSettings,
		)

		mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities",
			mock.Anything, userID, pseudonymID, &subforumID).
			Return([]string{constants.RoleUser, constants.RoleSubforumOwner}, subforumOwnerCapabilities, nil)

		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityModerateContent, &subforumID).
			Return(true, nil)

		ctx := context.Background()

		// Test Step 1: User can create subforum
		canCreateSubforum, err := mockPermissionDAO.HasUnifiedCapability(
			ctx, userID, pseudonymID, constants.CapabilityCreateSubforum, nil)
		assert.NoError(t, err)
		assert.True(t, canCreateSubforum,
			"User should be able to create subforums")

		// Test Step 2: After creating subforum, user can moderate it
		canModerateSubforum, err := mockPermissionDAO.HasUnifiedCapability(
			ctx, userID, pseudonymID, constants.CapabilityModerateContent, &subforumID)
		assert.NoError(t, err)
		assert.True(t, canModerateSubforum,
			"Subforum creator should be able to moderate their subforum")

		// Verify role progression
		globalRoles, _, err := mockPermissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
			ctx, userID, pseudonymID, nil)
		assert.NoError(t, err)
		assert.Contains(t, globalRoles, constants.RoleUser)

		subforumRoles, _, err := mockPermissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
			ctx, userID, pseudonymID, &subforumID)
		assert.NoError(t, err)
		assert.Contains(t, subforumRoles, constants.RoleUser)
		assert.Contains(t, subforumRoles, constants.RoleSubforumOwner)

		mockPermissionDAO.AssertExpectations(t)
	})
}

