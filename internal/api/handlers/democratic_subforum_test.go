package handlers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
)

// TestSubforumHandler_CreateSubforum_DemocraticValidation tests the democratic subforum validation logic
func TestSubforumHandler_CreateSubforum_DemocraticValidation(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)
	t.Run("missing_co-moderators_validation", func(t *testing.T) {
		// Setup mocks
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockSubforumSubscriptionDAO := &mocks.MockSubforumSubscriptionDAO{}
		mockPermissionDAO := &mocks.MockPermissionDAO{}
		mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
		mockPseudonymDAO := &mocks.MockPseudonymDAO{}
		mockPostDAO := &mocks.MockPostDAO{}
		mockUserDAO := &mocks.MockUserDAO{}

		// Create handler using constructor
		handler := handlers.NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, nil, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO, mockUserDAO)

		// Create proper JWT token for authentication
		userCtx := &middleware.UserContext{
			UserID:            1,
			ActivePseudonymID: "test-pseudonym-id",
		}
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, tokenErr)

		// Setup input with no co-moderators (single moderator - now valid)
		input := &models.SubforumCreateInput{
			Authorization: "Bearer " + token,
			Body: models.SubforumCreateBody{
				Slug:          "test-subforum",
				Name:          "Test Subforum",
				Description:   "A test subforum",
				CommunityType: constants.CommunityTypeTopical, // Democratic
				CoModerators:  []string{},                     // No co-moderators - valid!
			},
		}

		// Mock authentication
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "test-pseudonym-id", constants.CapabilityCreateSubforum, (*int32)(nil)).Return(true, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "test-pseudonym-id", constants.CapabilitySystemAdmin, (*int32)(nil)).Return(false, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "test-pseudonym-id", constants.CapabilityUserManagement, (*int32)(nil)).Return(false, nil)

		// Mock subforum creation
		expectedSubforum := &dbmodels.Subforum{
			SubforumID:      1,
			Name:            "test-subforum",
			DisplayName:     "Test Subforum",
			Description:     sql.Null[string]{V: "A test subforum", Valid: true},
			CommunityType:   constants.CommunityTypeTopical,
			GovernanceStyle: constants.GovernanceStyleDemocratic,
			IsNSFW:          sql.Null[bool]{V: false, Valid: true},
			IsPrivate:       sql.Null[bool]{V: false, Valid: true},
			IsRestricted:    sql.Null[bool]{V: false, Valid: true},
		}
		mockSubforumDAO.On("CreateSubforum", mock.Anything, "test-subforum", "Test Subforum", "A test subforum", "", constants.CommunityTypeTopical, constants.GovernanceStyleDemocratic, false, false, false, "").Return(expectedSubforum, nil)

		// Mock subscription count for convertSubforumToAPIModel
		mockSubforumSubscriptionDAO.On("CountSubscriptionsBySubforum", mock.Anything, int32(1)).Return(int64(0), nil)

		// Mock post count for convertSubforumToAPIModel
		mockPostDAO.On("CountPostsBySubforum", mock.Anything, int32(1)).Return(int64(0), nil)

		// Mock role key creation for single moderator
		mockRoleKeyDAO.On("CreateRoleKeyWithIBE", mock.Anything, constants.RoleElectedModerator, constants.ScopeModeration, mock.Anything, mock.Anything, "test-pseudonym-id", "test-pseudonym-id", mock.Anything).Return(&dbmodels.RoleKey{}, nil)

		// Mock pseudonym last active update
		mockPseudonymDAO.On("UpdateLastActive", mock.Anything, "test-pseudonym-id").Return(nil)

		// Create context with user
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Execute
		result, err := handler.CreateSubforum(ctx, input)

		// Verify success - no co-moderators is now valid
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-subforum", result.Body.Subforum.Name)

		// Verify mocks
		mockPermissionDAO.AssertExpectations(t)
		mockSubforumDAO.AssertExpectations(t)
		mockRoleKeyDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("co-moderators_validation", func(t *testing.T) {
		// Setup mocks
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockPermissionDAO := &mocks.MockPermissionDAO{}
		mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
		mockPseudonymDAO := &mocks.MockPseudonymDAO{}
		mockUserDAO := &mocks.MockUserDAO{}

		// Create handler using constructor
		handler := handlers.NewSubforumHandler(nil, mockSubforumDAO, nil, mockPermissionDAO, nil, mockPseudonymDAO, nil, mockRoleKeyDAO, mockUserDAO)

		// Create proper JWT token for authentication
		userCtx := &middleware.UserContext{
			UserID:            1,
			ActivePseudonymID: "test-pseudonym-id",
		}
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, tokenErr)

		// Setup input with too many co-moderators
		input := &models.SubforumCreateInput{
			Authorization: "Bearer " + token,
			Body: models.SubforumCreateBody{
				Slug:          "test-subforum",
				Name:          "Test Subforum",
				Description:   "A test subforum",
				CommunityType: constants.CommunityTypeTopical,                                                 // Democratic
				CoModerators:  []string{"pseudonym1", "pseudonym2", "pseudonym3", "pseudonym4", "pseudonym5"}, // Too many!
			},
		}

		// Mock authentication
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "test-pseudonym-id", constants.CapabilityCreateSubforum, (*int32)(nil)).Return(true, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "test-pseudonym-id", constants.CapabilitySystemAdmin, (*int32)(nil)).Return(false, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "test-pseudonym-id", constants.CapabilityUserManagement, (*int32)(nil)).Return(false, nil)

		// Create context with user
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Execute
		result, err := handler.CreateSubforum(ctx, input)

		// Verify validation error
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "democratic subforums can have at most 4 co-moderators")

		// Verify mocks
		mockPermissionDAO.AssertExpectations(t)
	})

	t.Run("co-moderators_owned_by_same_user_validation", func(t *testing.T) {
		// Setup mocks
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockPermissionDAO := &mocks.MockPermissionDAO{}
		mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
		mockPseudonymDAO := &mocks.MockPseudonymDAO{}
		mockUserDAO := &mocks.MockUserDAO{}

		// Create handler using constructor
		handler := handlers.NewSubforumHandler(nil, mockSubforumDAO, nil, mockPermissionDAO, nil, mockPseudonymDAO, nil, mockRoleKeyDAO, mockUserDAO)

		// Create proper JWT token for authentication
		userCtx := &middleware.UserContext{
			UserID:            1,
			ActivePseudonymID: "creator-pseudonym",
		}
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, tokenErr)

		// Setup input with co-moderators owned by the same user
		input := &models.SubforumCreateInput{
			Authorization: "Bearer " + token,
			Body: models.SubforumCreateBody{
				Slug:          "test-subforum",
				Name:          "Test Subforum",
				Description:   "A test subforum",
				CommunityType: constants.CommunityTypeTopical,       // Democratic
				CoModerators:  []string{"pseudonym1", "pseudonym2"}, // Both owned by user 2
			},
		}

		// Mock authentication
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "creator-pseudonym", constants.CapabilityCreateSubforum, (*int32)(nil)).Return(true, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "creator-pseudonym", constants.CapabilitySystemAdmin, (*int32)(nil)).Return(false, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "creator-pseudonym", constants.CapabilityUserManagement, (*int32)(nil)).Return(false, nil)

		// Mock pseudonym correlation - creator owned by user 1, co-moderators by user 2
		mockPseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, "creator-pseudonym", constants.RolePlatformAdmin, constants.ScopeCorrelation).Return(int64(1), nil)
		mockPseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, "pseudonym1", constants.RolePlatformAdmin, constants.ScopeCorrelation).Return(int64(2), nil)
		mockPseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, "pseudonym2", constants.RolePlatformAdmin, constants.ScopeCorrelation).Return(int64(2), nil) // Same user!

		// Note: GetPseudonymByID and GetUserByID should NOT be called because validation fails early
		// when detecting same-user ownership in the first pass

		// Create context with user
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Execute
		result, err := handler.CreateSubforum(ctx, input)

		// Verify validation error - should fail because co-moderators are owned by the same user
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "are owned by the same user - all moderators must be different users")

		// Verify mocks
		mockPermissionDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("successful_democratic_subforum_creation_single_moderator", func(t *testing.T) {
		// Setup mocks
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockSubforumSubscriptionDAO := &mocks.MockSubforumSubscriptionDAO{}
		mockPermissionDAO := &mocks.MockPermissionDAO{}
		mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
		mockPseudonymDAO := &mocks.MockPseudonymDAO{}
		mockPostDAO := &mocks.MockPostDAO{}
		mockUserDAO := &mocks.MockUserDAO{}

		// Create handler using constructor
		handler := handlers.NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, nil, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO, mockUserDAO)

		// Create proper JWT token for authentication
		userCtx := &middleware.UserContext{
			UserID:            1,
			ActivePseudonymID: "creator-pseudonym",
		}
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, tokenErr)

		// Setup input with no co-moderators (single moderator)
		input := &models.SubforumCreateInput{
			Authorization: "Bearer " + token,
			Body: models.SubforumCreateBody{
				Slug:          "test-subforum",
				Name:          "Test Subforum",
				Description:   "A test subforum",
				CommunityType: constants.CommunityTypeTopical, // Democratic
				CoModerators:  []string{},                     // No co-moderators
			},
		}

		// Mock authentication
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "creator-pseudonym", constants.CapabilityCreateSubforum, (*int32)(nil)).Return(true, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "creator-pseudonym", constants.CapabilitySystemAdmin, (*int32)(nil)).Return(false, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "creator-pseudonym", constants.CapabilityUserManagement, (*int32)(nil)).Return(false, nil)

		// Mock subforum creation
		expectedSubforum := &dbmodels.Subforum{
			SubforumID:      1,
			Name:            "test-subforum",
			DisplayName:     "Test Subforum",
			Description:     sql.Null[string]{V: "A test subforum", Valid: true},
			CommunityType:   constants.CommunityTypeTopical,
			GovernanceStyle: constants.GovernanceStyleDemocratic,
			IsNSFW:          sql.Null[bool]{V: false, Valid: true},
			IsPrivate:       sql.Null[bool]{V: false, Valid: true},
			IsRestricted:    sql.Null[bool]{V: false, Valid: true},
		}
		mockSubforumDAO.On("CreateSubforum", mock.Anything, "test-subforum", "Test Subforum", "A test subforum", "", constants.CommunityTypeTopical, constants.GovernanceStyleDemocratic, false, false, false, "").Return(expectedSubforum, nil)

		// Mock subscription count for convertSubforumToAPIModel
		mockSubforumSubscriptionDAO.On("CountSubscriptionsBySubforum", mock.Anything, int32(1)).Return(int64(0), nil)

		// Mock post count for convertSubforumToAPIModel
		mockPostDAO.On("CountPostsBySubforum", mock.Anything, int32(1)).Return(int64(0), nil)

		// Mock role key creation for single moderator
		mockRoleKeyDAO.On("CreateRoleKeyWithIBE", mock.Anything, constants.RoleElectedModerator, constants.ScopeModeration, mock.Anything, mock.Anything, "creator-pseudonym", "creator-pseudonym", mock.Anything).Return(&dbmodels.RoleKey{}, nil)

		// Mock pseudonym last active update
		mockPseudonymDAO.On("UpdateLastActive", mock.Anything, "creator-pseudonym").Return(nil)

		// Create context with user
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Execute
		result, err := handler.CreateSubforum(ctx, input)

		// Verify success
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-subforum", result.Body.Subforum.Name)

		// Verify mocks
		mockPermissionDAO.AssertExpectations(t)
		mockSubforumDAO.AssertExpectations(t)
		mockRoleKeyDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("successful_democratic_subforum_creation_with_co_moderators", func(t *testing.T) {
		// Setup mocks
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockSubforumSubscriptionDAO := &mocks.MockSubforumSubscriptionDAO{}
		mockPermissionDAO := &mocks.MockPermissionDAO{}
		mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
		mockPseudonymDAO := &mocks.MockPseudonymDAO{}
		mockPostDAO := &mocks.MockPostDAO{}
		mockUserDAO := &mocks.MockUserDAO{}

		// Create handler using constructor
		handler := handlers.NewSubforumHandler(nil, mockSubforumDAO, mockSubforumSubscriptionDAO, mockPermissionDAO, nil, mockPseudonymDAO, mockPostDAO, mockRoleKeyDAO, mockUserDAO)

		// Create proper JWT token for authentication
		userCtx := &middleware.UserContext{
			UserID:            1,
			ActivePseudonymID: "creator-pseudonym",
		}
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, tokenErr)

		// Setup input with valid co-moderators owned by different users
		input := &models.SubforumCreateInput{
			Authorization: "Bearer " + token,
			Body: models.SubforumCreateBody{
				Slug:          "test-democratic-subforum",
				Name:          "Test Democratic Subforum",
				Description:   "A test democratic subforum",
				CommunityType: constants.CommunityTypeTopical,       // Democratic
				CoModerators:  []string{"moderator1", "moderator2"}, // Different users
			},
		}

		// Mock authentication
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "creator-pseudonym", constants.CapabilityCreateSubforum, (*int32)(nil)).Return(true, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "creator-pseudonym", constants.CapabilitySystemAdmin, (*int32)(nil)).Return(false, nil)
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "creator-pseudonym", constants.CapabilityUserManagement, (*int32)(nil)).Return(false, nil)

		// Mock pseudonym correlation - all owned by different users
		mockPseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, "creator-pseudonym", constants.RolePlatformAdmin, constants.ScopeCorrelation).Return(int64(1), nil)
		mockPseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, "moderator1", constants.RolePlatformAdmin, constants.ScopeCorrelation).Return(int64(2), nil)
		mockPseudonymDAO.On("GetUserIDByPseudonym", mock.Anything, "moderator2", constants.RolePlatformAdmin, constants.ScopeCorrelation).Return(int64(3), nil)

		// Mock pseudonym existence checks for co-moderators
		mockPseudonymDAO.On("GetPseudonymByID", mock.Anything, "moderator1").Return(&dbmodels.Pseudonym{PseudonymID: "moderator1", DisplayName: "Moderator 1"}, nil)
		mockPseudonymDAO.On("GetPseudonymByID", mock.Anything, "moderator2").Return(&dbmodels.Pseudonym{PseudonymID: "moderator2", DisplayName: "Moderator 2"}, nil)

		// Mock user existence and validation checks for co-moderators
		mockUserDAO.On("GetUserByID", mock.Anything, int64(2)).Return(&dbmodels.User{UserID: 2, IsActive: sql.Null[bool]{V: true, Valid: true}, EmailVerified: sql.Null[bool]{V: true, Valid: true}}, nil)
		mockUserDAO.On("GetUserByID", mock.Anything, int64(3)).Return(&dbmodels.User{UserID: 3, IsActive: sql.Null[bool]{V: true, Valid: true}, EmailVerified: sql.Null[bool]{V: true, Valid: true}}, nil)

		// Mock subforum creation
		createdSubforum := &dbmodels.Subforum{
			SubforumID:      1,
			Name:            "test-democratic-subforum",
			DisplayName:     "Test Democratic Subforum",
			Description:     sql.Null[string]{V: "A test democratic subforum", Valid: true},
			CommunityType:   constants.CommunityTypeTopical,
			GovernanceStyle: constants.GovernanceStyleDemocratic,
		}
		mockSubforumDAO.On("CreateSubforum", mock.Anything, "test-democratic-subforum", "Test Democratic Subforum", "A test democratic subforum", "", constants.CommunityTypeTopical, constants.GovernanceStyleDemocratic, false, false, false, "").Return(createdSubforum, nil)

		// Mock subscription count for convertSubforumToAPIModel
		mockSubforumSubscriptionDAO.On("CountSubscriptionsBySubforum", mock.Anything, int32(1)).Return(int64(0), nil)

		// Mock post count for convertSubforumToAPIModel
		mockPostDAO.On("CountPostsBySubforum", mock.Anything, int32(1)).Return(int64(0), nil)

		// Mock role key creation for all 3 moderators (creator + 2 co-moderators)
		mockRoleKeyDAO.On("CreateRoleKeyWithIBE", mock.Anything, constants.RoleElectedModerator, constants.ScopeModeration, mock.Anything, mock.Anything, "creator-pseudonym", "creator-pseudonym", mock.Anything).Return(&dbmodels.RoleKey{KeyID: [16]byte{1}}, nil)
		mockRoleKeyDAO.On("CreateRoleKeyWithIBE", mock.Anything, constants.RoleElectedModerator, constants.ScopeModeration, mock.Anything, mock.Anything, "creator-pseudonym", "moderator1", mock.Anything).Return(&dbmodels.RoleKey{KeyID: [16]byte{2}}, nil)
		mockRoleKeyDAO.On("CreateRoleKeyWithIBE", mock.Anything, constants.RoleElectedModerator, constants.ScopeModeration, mock.Anything, mock.Anything, "creator-pseudonym", "moderator2", mock.Anything).Return(&dbmodels.RoleKey{KeyID: [16]byte{3}}, nil)

		// Mock pseudonym last active update
		mockPseudonymDAO.On("UpdateLastActive", mock.Anything, "creator-pseudonym").Return(nil)

		// Mock methods called during convertSubforumToAPIModel
		mockSubforumSubscriptionDAO.On("CountSubscriptionsBySubforum", mock.Anything, int32(1)).Return(int64(0), nil)
		mockPostDAO.On("CountPostsBySubforum", mock.Anything, int32(1)).Return(int64(0), nil)

		// Create context with user
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Execute
		result, err := handler.CreateSubforum(ctx, input)

		// Verify successful creation
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-democratic-subforum", result.Body.Subforum.Name)
		assert.Equal(t, "Test Democratic Subforum", result.Body.Subforum.DisplayName)

		// Verify mocks
		mockPermissionDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
		mockUserDAO.AssertExpectations(t)
		mockSubforumDAO.AssertExpectations(t)
		mockRoleKeyDAO.AssertExpectations(t)
		mockSubforumSubscriptionDAO.AssertExpectations(t)
		mockPostDAO.AssertExpectations(t)
	})
}
