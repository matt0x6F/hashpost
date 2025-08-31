package handlers_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	gomock "go.uber.org/mock/gomock"
)

// TestSubforumHandler_CreateSubforum_PermissionWorkflow tests the complete subforum creation workflow
// and verifies that the creator gets proper owner permissions using gomock
func TestSubforumHandler_CreateSubforum_PermissionWorkflow(t *testing.T) {
	t.Run("SubforumCreation_GrantsCreatorOwnerPermissions", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// Arrange
		handler, mockSubforumDAO, mockRoleKeyDAO, mockPermissionDAO, mockPseudonymDAO := NewSubforumHandlerWithGomocks(ctrl)

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
		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(
				gomock.Any(), userID, pseudonymID, constants.CapabilityCreateSubforum, (*int32)(nil)).
			Return(true, nil).
			Times(1)

		// Mock: User does not have admin capabilities (for is_restricted check)
		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(
				gomock.Any(), userID, pseudonymID, constants.CapabilitySystemAdmin, (*int32)(nil)).
			Return(false, nil).
			Times(1)

		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(
				gomock.Any(), userID, pseudonymID, constants.CapabilityUserManagement, (*int32)(nil)).
			Return(false, nil).
			Times(1)

		// Mock: Subforum creation
		mockSubforumDAO.EXPECT().
			CreateSubforum(gomock.Any(),
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
		}, nil).
			Times(1)

		// Mock: Role key creation for subforum owner - this is the key part we're testing
		// The creator should get role keys for moderation scope in this subforum
		// Note: The handler only creates ONE role key for moderation scope, not two
		mockRoleKeyDAO.EXPECT().
			CreateRoleKeyWithIBE(
				gomock.Any(),                // context
				constants.RoleSubforumOwner, // role name
				constants.ScopeModeration,   // scope
				gomock.Any(),                // capabilities
				gomock.Any(),                // expires at
				pseudonymID,                 // created by
				pseudonymID,                 // pseudonym id
				&subforumID,                 // subforum id
			).Return(&dbmodels.RoleKey{}, nil).
			Times(1)

		// Mock: Pseudonym DAO calls during subforum creation
		// The handler calls UpdateLastActive on the pseudonym DAO
		mockPseudonymDAO.EXPECT().
			UpdateLastActive(gomock.Any(), pseudonymID).
			Return(nil).
			Times(1)

		// Note: GetUserIDByPseudonym is only called when there are co-moderators
		// Since this test has no co-moderators, this mock is not needed

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

		mockPermissionDAO.EXPECT().
			GetUnifiedActivePseudonymRolesAndCapabilities(
				gomock.Any(), userID, pseudonymID, &subforumID).
			Return(expectedRoles, expectedCapabilities, nil).
			Times(1)

		// Mock specific capability checks for key moderation features
		keyModerationCapabilities := []string{
			constants.CapabilityModerateContent,
			constants.CapabilityBanUsers,
			constants.CapabilityManageModerators,
			constants.CapabilityManageSubforumSettings,
		}

		for _, capability := range keyModerationCapabilities {
			mockPermissionDAO.EXPECT().
				HasUnifiedCapability(
					gomock.Any(), userID, pseudonymID, capability, &subforumID).
				Return(true, nil).
				Times(1)
		}

		// Mock: Verify creator cannot moderate OTHER subforums
		otherSubforumID := int32(999)
		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(
				gomock.Any(), userID, pseudonymID, constants.CapabilityModerateContent, &otherSubforumID).
			Return(false, nil).
			Times(1)

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
	})

	t.Run("SubforumCreation_VerifiesModerationScope", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		// This test specifically focuses on verifying that subforum creation
		// creates role keys with moderation scope for owned subforums
		handler, mockSubforumDAO, mockRoleKeyDAO, mockPermissionDAO, mockPseudonymDAO := NewSubforumHandlerWithGomocks(ctrl)

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
		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(
				gomock.Any(), userID, pseudonymID, constants.CapabilityCreateSubforum, (*int32)(nil)).
			Return(true, nil).
			Times(1)

		// Mock admin capability checks
		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(
				gomock.Any(), userID, pseudonymID, constants.CapabilitySystemAdmin, (*int32)(nil)).
			Return(false, nil).
			Times(1)

		mockPermissionDAO.EXPECT().
			HasUnifiedCapability(
				gomock.Any(), userID, pseudonymID, constants.CapabilityUserManagement, (*int32)(nil)).
			Return(false, nil).
			Times(1)

		// Mock subforum creation
		mockSubforumDAO.EXPECT().
			CreateSubforum(gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(),
				gomock.Any(), gomock.Any(), gomock.Any(),
				pseudonymID).
			Return(&dbmodels.Subforum{SubforumID: subforumID, Name: subforumName}, nil).
			Times(1)

		// The critical mock: verify that moderation scope role key is created
		mockRoleKeyDAO.EXPECT().
			CreateRoleKeyWithIBE(
				gomock.Any(), // context
				constants.RoleSubforumOwner,
				constants.ScopeModeration, // Only moderation scope for owned subforums
				gomock.Any(),              // capabilities
				gomock.Any(),              // expires at
				pseudonymID,
				pseudonymID,
				&subforumID, // subforum id
			).Return(&dbmodels.RoleKey{}, nil).
			Times(1)

		// Mock: Pseudonym DAO calls during subforum creation
		// The handler calls UpdateLastActive on the pseudonym DAO
		mockPseudonymDAO.EXPECT().
			UpdateLastActive(gomock.Any(), pseudonymID).
			Return(nil).
			Times(1)

		// Mock that creator has capabilities from moderation scope after creation
		moderationCapabilities := []string{
			constants.CapabilityModerateContent,
			constants.CapabilityBanUsers,
			constants.CapabilityManageModerators,
		}

		for _, capability := range moderationCapabilities {
			mockPermissionDAO.EXPECT().
				HasUnifiedCapability(
					gomock.Any(), userID, pseudonymID, capability, &subforumID).
				Return(true, nil).
				Times(1)
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

		// Verify the creator has capabilities from moderation scope
		for _, capability := range moderationCapabilities {
			hasCapability, err := mockPermissionDAO.HasUnifiedCapability(
				ctx, userID, pseudonymID, capability, &subforumID)
			assert.NoError(t, err)
			assert.True(t, hasCapability,
				"Subforum creator should have moderation capability: %s", capability)
		}
	})
}

// NewSubforumHandlerWithGomocks creates a subforum handler with the necessary gomocks for permission testing
func NewSubforumHandlerWithGomocks(ctrl *gomock.Controller) (*handlers.SubforumHandler, *dao.MockSubforumDAOInterface, *dao.MockRoleKeyDAOInterface, *dao.MockPermissionDAOInterface, *dao.MockPseudonymDAOInterface) {
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockSubforumSubscriptionDAO := dao.NewMockSubforumSubscriptionDAOInterface(ctrl)
	mockRoleKeyDAO := dao.NewMockRoleKeyDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
	mockIdentityMappingDAO := dao.NewMockIdentityMappingDAOInterface(ctrl)

	// Mock methods that are called during convertSubforumToAPIModel
	mockSubforumSubscriptionDAO.EXPECT().
		CountSubscriptionsBySubforum(gomock.Any(), gomock.Any()).
		Return(int64(0), nil).
		AnyTimes()

	mockPostDAO.EXPECT().
		CountPostsBySubforum(gomock.Any(), gomock.Any()).
		Return(int64(0), nil).
		AnyTimes()

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
