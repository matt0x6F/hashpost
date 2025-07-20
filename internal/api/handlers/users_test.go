package handlers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test UserHandler with mocks
func createTestUserHandler() (*UserHandler, *mocks.MockUserDAO, *mocks.MockSecurePseudonymDAO, *mocks.MockUserPreferencesDAO, *mocks.MockUserBlocksDAO, *mocks.MockPostDAO, *mocks.MockCommentDAO) {
	mockUserDAO := &mocks.MockUserDAO{}
	mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()
	mockUserPreferencesDAO := &mocks.MockUserPreferencesDAO{}
	mockUserBlocksDAO := &mocks.MockUserBlocksDAO{}
	mockPostDAO := mocks.NewMockPostDAO()
	mockCommentDAO := mocks.NewMockCommentDAO()

	ibeSystem := ibe.NewIBESystem()

	handler := NewUserHandlerWithDependencies(
		mockUserDAO,
		mockSecurePseudonymDAO,
		mockUserPreferencesDAO,
		mockUserBlocksDAO,
		mockPostDAO,
		mockCommentDAO,
		ibeSystem,
	)

	return handler, mockUserDAO, mockSecurePseudonymDAO, mockUserPreferencesDAO, mockUserBlocksDAO, mockPostDAO, mockCommentDAO
}

func TestUserHandler_GetPseudonymProfile(t *testing.T) {
	handler, _, mockSecurePseudonymDAO, _, _, mockPostDAO, mockCommentDAO := createTestUserHandler()

	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		pseudonymID := "test-pseudonym-id"

		// Mock pseudonym data
		testPseudonym := fixtures.CreateTestPseudonym()

		// Set up mock expectations
		mockSecurePseudonymDAO.On("GetPseudonymByID", ctx, pseudonymID).Return(testPseudonym, nil)
		mockPostDAO.On("CountPostsByPseudonym", ctx, pseudonymID).Return(int64(5), nil)
		mockCommentDAO.On("CountCommentsByPseudonym", ctx, pseudonymID).Return(int64(10), nil)

		// Create input
		input := &apimodels.PseudonymIDPathParam{
			PseudonymID: pseudonymID,
		}

		// Call the method
		result, err := handler.GetPseudonymProfile(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, pseudonymID, result.Body.PseudonymID)
		require.Equal(t, "TestUser", result.Body.DisplayName)
		require.Equal(t, "Test bio", result.Body.Bio)
		require.Equal(t, "https://example.com", result.Body.WebsiteURL)
		require.Equal(t, 100, result.Body.KarmaScore)
		require.Equal(t, 5, result.Body.PostCount)
		require.Equal(t, 10, result.Body.CommentCount)
		require.True(t, result.Body.ShowKarma)
		require.True(t, result.Body.AllowDirectMessages)

		// Verify mock expectations
		mockSecurePseudonymDAO.AssertExpectations(t)
		mockPostDAO.AssertExpectations(t)
		mockCommentDAO.AssertExpectations(t)
	})

	t.Run("PseudonymNotFound", func(t *testing.T) {
		ctx := context.Background()
		pseudonymID := "non-existent-pseudonym"

		// Set up mock expectations
		mockSecurePseudonymDAO.On("GetPseudonymByID", ctx, pseudonymID).Return(nil, nil)

		// Create input
		input := &apimodels.PseudonymIDPathParam{
			PseudonymID: pseudonymID,
		}

		// Call the method
		result, err := handler.GetPseudonymProfile(ctx, input)

		// Assertions
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "pseudonym not found")

		// Verify mock expectations
		mockSecurePseudonymDAO.AssertExpectations(t)
	})

	t.Run("PseudonymInactive", func(t *testing.T) {
		ctx := context.Background()
		pseudonymID := "inactive-pseudonym"

		// Mock inactive pseudonym
		inactivePseudonym := fixtures.CreateTestPseudonym()
		inactivePseudonym.IsActive = sql.Null[bool]{V: false, Valid: true}

		// Set up mock expectations
		mockSecurePseudonymDAO.On("GetPseudonymByID", ctx, pseudonymID).Return(inactivePseudonym, nil)

		// Create input
		input := &apimodels.PseudonymIDPathParam{
			PseudonymID: pseudonymID,
		}

		// Call the method
		result, err := handler.GetPseudonymProfile(ctx, input)

		// Assertions
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "pseudonym is inactive")

		// Verify mock expectations
		mockSecurePseudonymDAO.AssertExpectations(t)
	})
}

func TestUserHandler_CreatePseudonym(t *testing.T) {
	handler, mockUserDAO, mockSecurePseudonymDAO, _, _, _, _ := createTestUserHandler()

	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()

		// Mock pseudonym data
		testPseudonym := fixtures.CreateTestPseudonym()
		testPseudonym.DisplayName = "NewPseudonym"

		// Set up mock expectations
		mockSecurePseudonymDAO.On("GetPseudonymByDisplayName", ctx, "NewPseudonym").Return(nil, nil)
		mockSecurePseudonymDAO.On("CreatePseudonymWithIdentityMapping", ctx, int64(1), "NewPseudonym").Return(testPseudonym, nil)
		mockSecurePseudonymDAO.On("UpdatePseudonym", ctx, "test-pseudonym-id", mock.AnythingOfType("*models.PseudonymSetter")).Return(nil)
		mockSecurePseudonymDAO.On("GetPseudonymByID", ctx, "test-pseudonym-id").Return(testPseudonym, nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.CreatePseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			CreatePseudonymInput: apimodels.CreatePseudonymInput{
				Body: apimodels.CreatePseudonymBody{
					DisplayName:         "NewPseudonym",
					Bio:                 "New pseudonym bio",
					WebsiteURL:          "https://new.example.com",
					ShowKarma:           &[]bool{true}[0],
					AllowDirectMessages: &[]bool{true}[0],
				},
			},
		}

		// Call the method
		result, err := handler.CreatePseudonym(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, "test-pseudonym-id", result.Body.PseudonymID)
		require.Equal(t, "NewPseudonym", result.Body.DisplayName)

		// Verify mock expectations
		mockUserDAO.AssertExpectations(t)
		mockSecurePseudonymDAO.AssertExpectations(t)
	})

	t.Run("DisplayNameAlreadyTaken", func(t *testing.T) {
		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()

		// Mock existing pseudonym
		existingPseudonym := fixtures.CreateTestPseudonym()

		// Set up mock expectations
		mockSecurePseudonymDAO.On("GetPseudonymByDisplayName", ctx, "ExistingPseudonym").Return(existingPseudonym, nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.CreatePseudonymInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + token,
			},
			CreatePseudonymInput: apimodels.CreatePseudonymInput{
				Body: apimodels.CreatePseudonymBody{
					DisplayName:         "ExistingPseudonym",
					Bio:                 "New pseudonym bio",
					WebsiteURL:          "https://new.example.com",
					ShowKarma:           &[]bool{true}[0],
					AllowDirectMessages: &[]bool{true}[0],
				},
			},
		}

		// Call the method
		result, err := handler.CreatePseudonym(ctx, input)

		// Assertions
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "display name is already taken")

		// Verify mock expectations
		mockUserDAO.AssertExpectations(t)
		mockSecurePseudonymDAO.AssertExpectations(t)
	})
}

func TestUserHandler_GetUserProfile(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()

		// Mock user data
		testUser := &models.User{
			UserID:       1,
			Email:        "test@example.com",
			IsActive:     sql.Null[bool]{V: true, Valid: true},
			CreatedAt:    sql.Null[time.Time]{V: time.Now(), Valid: true},
			LastActiveAt: sql.Null[time.Time]{V: time.Now(), Valid: true},
		}

		// Mock pseudonym data
		testPseudonym := fixtures.CreateTestPseudonym()

		// Set up mock expectations
		mockUserDAO := &mocks.MockUserDAO{}
		mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()

		mockUserDAO.On("GetUserByID", mock.Anything, int64(1)).Return(testUser, nil)
		mockSecurePseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, int64(1), "user", "authentication").Return([]*models.Pseudonym{testPseudonym}, nil)

		// Create mock PostDAO and CommentDAO
		mockPostDAO := mocks.NewMockPostDAO()
		mockCommentDAO := mocks.NewMockCommentDAO()

		// Set up mock expectations for post and comment counts
		mockPostDAO.On("CountPostsByPseudonym", mock.Anything, "test-pseudonym-id").Return(int64(5), nil)
		mockCommentDAO.On("CountCommentsByPseudonym", mock.Anything, "test-pseudonym-id").Return(int64(10), nil)

		// Create handler with mocked dependencies
		handler := NewUserHandlerWithDependencies(
			mockUserDAO,
			mockSecurePseudonymDAO,
			nil, // Mock UserPreferencesDAO
			nil, // Mock UserBlocksDAO
			mockPostDAO,
			mockCommentDAO,
			ibe.NewIBESystem(),
		)

		// Generate a proper JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &middleware.AuthInput{
			Authorization: "Bearer " + token,
		}

		// Set up user context in the request context
		ctx = context.WithValue(ctx, middleware.UserContextKeyValue, userCtx)

		// Call the method
		result, err := handler.GetUserProfile(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, result)
		require.Equal(t, 1, result.Body.UserID)
		require.Equal(t, "test@example.com", result.Body.Email)
		require.Len(t, result.Body.Pseudonyms, 1)
		require.Equal(t, "test-pseudonym-id", result.Body.Pseudonyms[0].PseudonymID)
		require.Equal(t, "TestUser", result.Body.Pseudonyms[0].DisplayName)

		// Verify mock expectations
		mockUserDAO.AssertExpectations(t)
		mockSecurePseudonymDAO.AssertExpectations(t)
	})
}
