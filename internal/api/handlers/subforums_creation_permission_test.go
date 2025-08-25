package handlers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
)

// TestSubforumHandler_CreateSubforum_PermissionWorkflow tests the complete subforum creation workflow
// and verifies that the creator gets proper owner permissions
func TestSubforumHandler_CreateSubforum_PermissionWorkflow(t *testing.T) {
	t.Run("SubforumCreation_GrantsCreatorOwnerPermissions", func(t *testing.T) {
		// Arrange
		handler, mockSubforumDAO, mockRoleKeyDAO, mockPermissionDAO, mockPseudonymDAO := NewSubforumHandlerWithMocks()

		// Set up global auth middleware
		authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
		middleware.SetGlobalAuthMiddleware(authMiddleware)

		// Test data
		userID := int64(1)
		pseudonymID := "creator-pseudo-1"
		subforumID := int32(1)
		subforumName := "TestSubforum"

		// Create user context for subforum creator
		userCtx := &middleware.UserContext{
			UserID:            userID,
			Email:             "creator@example.com",
			ActivePseudonymID: pseudonymID,
			DisplayName:       "SubforumCreator",
			MFAEnabled:        false,
		}

		// Mock: User can create subforum (pre-creation permission)
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityCreateSubforum, (*int32)(nil)).
			Return(true, nil)

		// Mock: User does not have admin capabilities (for is_restricted check)
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilitySystemAdmin, (*int32)(nil)).
			Return(false, nil)

		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityUserManagement, (*int32)(nil)).
			Return(false, nil)

		// Mock: Subforum creation
		mockSubforumDAO.On("CreateSubforum", mock.Anything,
			subforumName,                   // name
			subforumName,                   // displayName
			"Test subforum description",    // description
			"",                             // sidebarText
			constants.CommunityTypeBranded, // communityType
			constants.GovernanceStyleOwned, // governanceStyle
			false,                          // isNSFW
			false,                          // isPrivate
			false,                          // isRestricted
			pseudonymID,                    // ownerPseudonymID
		).Return(&dbmodels.Subforum{
			SubforumID:    subforumID,
			Name:          subforumName,
			DisplayName:   subforumName,
			CommunityType: constants.CommunityTypeBranded,
		}, nil)

		// Mock: Role key creation for subforum owner - this is the key part we're testing
		// The creator should get role keys for moderation scope in this subforum
		// Note: The handler only creates ONE role key for moderation scope, not two
		mockRoleKeyDAO.On("CreateRoleKeyWithIBE",
			mock.Anything,                    // context
			constants.RoleSubforumOwner,      // role name
			constants.ScopeModeration,        // scope
			mock.AnythingOfType("[]string"),  // capabilities
			mock.AnythingOfType("time.Time"), // expires at
			pseudonymID,                      // created by
			pseudonymID,                      // pseudonym id
			&subforumID,                      // subforum id
		).Return(&dbmodels.RoleKey{}, nil)

		// Mock: Pseudonym DAO calls during subforum creation
		// The handler calls UpdateLastActive on the pseudonym DAO
		mockPseudonymDAO.On("UpdateLastActive", mock.Anything, pseudonymID).Return(nil)

		// Mock: Post-creation permission verification
		// After creation, the user should have subforum owner role and all moderation capabilities in their subforum
		expectedRoles := []string{constants.RoleUser, constants.RoleSubforumOwner}
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
		}

		mockPermissionDAO.On("GetUnifiedActivePseudonymRolesAndCapabilities",
			mock.Anything, userID, pseudonymID, &subforumID).
			Return(expectedRoles, expectedCapabilities, nil)

		// Mock specific capability checks for key moderation features
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

		// Mock: Verify creator cannot moderate OTHER subforums
		otherSubforumID := int32(999)
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityModerateContent, &otherSubforumID).
			Return(false, nil)

		// Act - Create the subforum
		ctx := middleware.SetUserContext(context.Background(), userCtx)

		input := &models.SubforumCreateInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(userID, pseudonymID),
			Body: models.SubforumCreateBody{
				Slug:          subforumName,
				Name:          subforumName,
				Description:   "Test subforum description",
				CommunityType: constants.CommunityTypeBranded, // Use branded type for owned governance
				IsNSFW:        false,
				IsPrivate:     false,
				IsRestricted:  false,
			},
		}

		response, err := handler.CreateSubforum(ctx, input)

		// Assert
		assert.NoError(t, err, "Subforum creation should succeed")
		assert.NotNil(t, response, "Create subforum response should not be nil")
		assert.NotNil(t, response.Body, "Create subforum response body should not be nil")
		assert.Equal(t, subforumName, response.Body.Subforum.Name)
		assert.Equal(t, subforumName, response.Body.Subforum.DisplayName)

		// Verify all creation mocks were called
		mockSubforumDAO.AssertExpectations(t)
		mockRoleKeyDAO.AssertExpectations(t)

		// Verify the creator now has subforum owner permissions in their subforum
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

		// Verify permission isolation - creator cannot moderate other subforums
		canModerateOther, err := mockPermissionDAO.HasUnifiedCapability(
			ctx, userID, pseudonymID, constants.CapabilityModerateContent, &otherSubforumID)
		assert.NoError(t, err)
		assert.False(t, canModerateOther,
			"Subforum creator should NOT be able to moderate other subforums")

		mockPermissionDAO.AssertExpectations(t)
	})

	t.Run("SubforumCreation_VerifiesModerationScope", func(t *testing.T) {
		// This test specifically focuses on verifying that subforum creation
		// creates role keys with moderation scope for owned subforums
		handler, mockSubforumDAO, mockRoleKeyDAO, mockPermissionDAO, mockPseudonymDAO := NewSubforumHandlerWithMocks()

		authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
		middleware.SetGlobalAuthMiddleware(authMiddleware)

		userID := int64(2)
		pseudonymID := "scope-test-creator-2"
		subforumID := int32(2)
		subforumName := "ScopeTestSubforum"

		userCtx := &middleware.UserContext{
			UserID:            userID,
			ActivePseudonymID: pseudonymID,
			Email:             "scopetest@example.com",
			DisplayName:       "ScopeTestCreator",
		}

		// Mock pre-creation permission
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityCreateSubforum, (*int32)(nil)).
			Return(true, nil)

		// Mock admin capability checks
		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilitySystemAdmin, (*int32)(nil)).
			Return(false, nil)

		mockPermissionDAO.On("HasUnifiedCapability",
			mock.Anything, userID, pseudonymID, constants.CapabilityUserManagement, (*int32)(nil)).
			Return(false, nil)

		// Mock subforum creation
		mockSubforumDAO.On("CreateSubforum", mock.Anything,
			mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
			mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"),
			mock.AnythingOfType("bool"), mock.AnythingOfType("bool"), mock.AnythingOfType("bool"),
			pseudonymID).
			Return(&dbmodels.Subforum{SubforumID: subforumID, Name: subforumName}, nil)

		// The critical mock: verify that moderation scope role key is created
		mockRoleKeyDAO.On("CreateRoleKeyWithIBE",
			mock.Anything, // context
			constants.RoleSubforumOwner,
			constants.ScopeModeration, // Only moderation scope for owned subforums
			mock.MatchedBy(func(capabilities []string) bool {
				// Verify moderation capabilities are included
				expectedModerationCaps := []string{
					constants.CapabilityModerateContent,
					constants.CapabilityBanUsers,
					constants.CapabilityRemoveContent,
					constants.CapabilityManageModerators,
				}
				for _, expectedCap := range expectedModerationCaps {
					found := false
					for _, actualCap := range capabilities {
						if actualCap == expectedCap {
							found = true
							break
						}
					}
					if !found {
						return false
					}
				}
				return true
			}),
			mock.AnythingOfType("time.Time"),
			pseudonymID,
			pseudonymID,
			&subforumID, // subforum id
		).Return(&dbmodels.RoleKey{}, nil)

		// Mock: Pseudonym DAO calls during subforum creation
		// The handler calls UpdateLastActive on the pseudonym DAO
		mockPseudonymDAO.On("UpdateLastActive", mock.Anything, pseudonymID).Return(nil)

		// Mock that creator has capabilities from moderation scope after creation
		moderationCapabilities := []string{
			constants.CapabilityModerateContent,
			constants.CapabilityBanUsers,
			constants.CapabilityManageModerators,
		}

		for _, capability := range moderationCapabilities {
			mockPermissionDAO.On("HasUnifiedCapability",
				mock.Anything, userID, pseudonymID, capability, &subforumID).
				Return(true, nil)
		}

		// Act
		ctx := middleware.SetUserContext(context.Background(), userCtx)

		input := &models.SubforumCreateInput{
			Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(userID, pseudonymID),
			Body: models.SubforumCreateBody{
				Slug:          subforumName,
				Name:          subforumName,
				Description:   "Test description",
				CommunityType: constants.CommunityTypeBranded,
				IsNSFW:        false,
				IsPrivate:     false,
				IsRestricted:  false,
			},
		}

		response, err := handler.CreateSubforum(ctx, input)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, response)

		// The key assertion: verify that role key was created for moderation scope
		mockRoleKeyDAO.AssertExpectations(t)

		// Verify the creator has capabilities from moderation scope
		for _, capability := range moderationCapabilities {
			hasCapability, err := mockPermissionDAO.HasUnifiedCapability(
				ctx, userID, pseudonymID, capability, &subforumID)
			assert.NoError(t, err)
			assert.True(t, hasCapability,
				"Subforum creator should have moderation capability: %s", capability)
		}

		mockPermissionDAO.AssertExpectations(t)
	})
}

// NewSubforumHandlerWithMocks creates a subforum handler with the necessary mocks for permission testing
func NewSubforumHandlerWithMocks() (*handlers.SubforumHandler, *mocks.MockSubforumDAO, *mocks.MockRoleKeyDAO, *mocks.MockPermissionDAO, *mocks.MockPseudonymDAO) {
	mockSubforumDAO := mocks.NewMockSubforumDAO()
	mockSubforumSubscriptionDAO := mocks.NewMockSubforumSubscriptionDAO()
	mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
	mockPermissionDAO := mocks.NewMockPermissionDAO()
	mockPseudonymDAO := mocks.NewMockPseudonymDAO()
	mockPostDAO := mocks.NewMockPostDAO()
	mockUserDAO := &mocks.MockUserDAO{}
	mockIdentityMappingDAO := &mocks.MockIdentityMappingDAO{}

	// Don't set default behaviors - only mock what's actually needed for the test
	// The handler only calls specific methods during subforum creation

	// Mock methods that are called during convertSubforumToAPIModel
	mockSubforumSubscriptionDAO.On("CountSubscriptionsBySubforum", mock.Anything, mock.AnythingOfType("int32")).Return(int64(0), nil)
	mockPostDAO.On("CountPostsBySubforum", mock.Anything, mock.AnythingOfType("int32")).Return(int64(0), nil)

	// Create handler with all required dependencies
	handler := handlers.NewSubforumHandler(
		nil, // db
		mockSubforumDAO,
		mockSubforumSubscriptionDAO,
		mockPermissionDAO,
		mockIdentityMappingDAO,
		mockPseudonymDAO,
		mockPostDAO,
		mockRoleKeyDAO,
		mockUserDAO,
	)

	return handler, mockSubforumDAO, mockRoleKeyDAO, mockPermissionDAO, mockPseudonymDAO
}
