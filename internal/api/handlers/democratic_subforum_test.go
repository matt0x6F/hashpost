package handlers

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
)

func TestSubforumHandler_CreateSubforum_DemocraticValidation(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("missing co-moderators validation", func(t *testing.T) {
		// Setup mocks
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockPermissionDAO := &mocks.MockPermissionDAO{}
		mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
		mockPseudonymDAO := &mocks.MockPseudonymDAO{}
		mockUserDAO := &mocks.MockUserDAO{}

		// Create handler
		handler := &SubforumHandler{
			subforumDAO:   mockSubforumDAO,
			permissionDAO: mockPermissionDAO,
			roleKeyDAO:    mockRoleKeyDAO,
			pseudonymDAO:  mockPseudonymDAO,
			userDAO:       mockUserDAO,
		}

		// Create proper JWT token for authentication
		userCtx := &middleware.UserContext{
			UserID:            1,
			ActivePseudonymID: "test-pseudonym-id",
		}
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, tokenErr)

		// Setup input with missing co-moderators
		input := &models.SubforumCreateInput{
			Authorization: "Bearer " + token,
			Body: models.SubforumCreateBody{
				Slug:          "test-subforum",
				Name:          "Test Subforum",
				Description:   "A test subforum",
				CommunityType: constants.CommunityTypeTopical, // Democratic
				CoModerators:  []string{},                     // Missing!
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
		assert.Contains(t, err.Error(), "democratic subforums require exactly 2 co-moderators")

		// Verify mocks
		mockPermissionDAO.AssertExpectations(t)
	})

	t.Run("creator selects own pseudonym validation", func(t *testing.T) {
		// Setup mocks
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockPermissionDAO := &mocks.MockPermissionDAO{}
		mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
		mockPseudonymDAO := &mocks.MockPseudonymDAO{}
		mockUserDAO := &mocks.MockUserDAO{}

		// Create handler
		handler := &SubforumHandler{
			subforumDAO:   mockSubforumDAO,
			permissionDAO: mockPermissionDAO,
			roleKeyDAO:    mockRoleKeyDAO,
			pseudonymDAO:  mockPseudonymDAO,
			userDAO:       mockUserDAO,
		}

		// Create proper JWT token for authentication
		userCtx := &middleware.UserContext{
			UserID:            1,
			ActivePseudonymID: "test-pseudonym-id",
		}
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, tokenErr)

		// Setup input with creator's own pseudonym as co-moderator
		input := &models.SubforumCreateInput{
			Authorization: "Bearer " + token,
			Body: models.SubforumCreateBody{
				Slug:          "test-subforum",
				Name:          "Test Subforum",
				Description:   "A test subforum",
				CommunityType: constants.CommunityTypeTopical,              // Democratic
				CoModerators:  []string{"test-pseudonym-id", "pseudonym3"}, // Creator's own ID!
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
		assert.Contains(t, err.Error(), "cannot select your own pseudonym as co-moderator")

		// Verify mocks
		mockPermissionDAO.AssertExpectations(t)
	})
}

func TestSubforumHandler_CreateSubforum_OwnedValidation(t *testing.T) {
	// Initialize global auth middleware for testing
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("owned subforum with co-moderators should fail", func(t *testing.T) {
		// Setup mocks
		mockSubforumDAO := &mocks.MockSubforumDAO{}
		mockPermissionDAO := &mocks.MockPermissionDAO{}
		mockRoleKeyDAO := &mocks.MockRoleKeyDAO{}
		mockPseudonymDAO := &mocks.MockPseudonymDAO{}
		mockUserDAO := &mocks.MockUserDAO{}

		// Create handler
		handler := &SubforumHandler{
			subforumDAO:   mockSubforumDAO,
			permissionDAO: mockPermissionDAO,
			roleKeyDAO:    mockRoleKeyDAO,
			pseudonymDAO:  mockPseudonymDAO,
			userDAO:       mockUserDAO,
		}

		// Create proper JWT token for authentication
		userCtx := &middleware.UserContext{
			UserID:            1,
			ActivePseudonymID: "test-pseudonym-id",
		}
		token, tokenErr := middleware.GenerateJWT(userCtx, "test-secret", time.Hour)
		require.NoError(t, tokenErr)

		// Setup input with co-moderators for owned subforum (should fail)
		input := &models.SubforumCreateInput{
			Authorization: "Bearer " + token,
			Body: models.SubforumCreateBody{
				Slug:          "test-subforum",
				Name:          "Test Subforum",
				Description:   "A test subforum",
				CommunityType: constants.CommunityTypeBranded,       // Owned (branded)
				CoModerators:  []string{"pseudonym2", "pseudonym3"}, // Should not be allowed!
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
		assert.Contains(t, err.Error(), "owned subforums do not use co-moderators")

		// Verify mocks
		mockPermissionDAO.AssertExpectations(t)
	})
}
