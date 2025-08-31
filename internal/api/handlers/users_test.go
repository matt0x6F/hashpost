package handlers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// NewUserHandlerWithGomocks creates a test UserHandler with gomock DAOs
func NewUserHandlerWithGomocks(ctrl *gomock.Controller) (*handlers.UserHandler, *dao.MockUserDAOInterface, *dao.MockPseudonymDAOInterface, *dao.MockUserPreferencesDAOInterface, *dao.MockUserBlocksDAOInterface, *dao.MockPostDAOInterface, *dao.MockCommentDAOInterface) {
	mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockUserPreferencesDAO := dao.NewMockUserPreferencesDAOInterface(ctrl)
	mockUserBlocksDAO := dao.NewMockUserBlocksDAOInterface(ctrl)
	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockCommentDAO := dao.NewMockCommentDAOInterface(ctrl)

	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

	handler := handlers.NewUserHandler(
		nil, // nil db for testing
		mockUserDAO,
		mockPseudonymDAO,
		mockUserPreferencesDAO,
		mockUserBlocksDAO,
		mockPostDAO,
		mockCommentDAO,
		ibeSystem,
	)

	return handler, mockUserDAO, mockPseudonymDAO, mockUserPreferencesDAO, mockUserBlocksDAO, mockPostDAO, mockCommentDAO
}

func TestUserHandler_GetPseudonymProfile_Gomock(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler, _, mockPseudonymDAO, _, _, mockPostDAO, mockCommentDAO := NewUserHandlerWithGomocks(ctrl)

		ctx := context.Background()
		pseudonymID := "test-pseudonym-id"

		// Mock pseudonym data
		testPseudonym := fixtures.CreateTestPseudonym()

		// Set up gomock expectations
		mockPseudonymDAO.EXPECT().
			GetPseudonymByID(ctx, pseudonymID).
			Return(testPseudonym, nil).
			Times(1)

		mockPostDAO.EXPECT().
			CountPostsByPseudonym(ctx, pseudonymID).
			Return(int64(5), nil).
			Times(1)

		mockCommentDAO.EXPECT().
			CountCommentsByPseudonym(ctx, pseudonymID).
			Return(int64(10), nil).
			Times(1)

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
	})

	t.Run("PseudonymNotFound", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler, _, mockPseudonymDAO, _, _, _, _ := NewUserHandlerWithGomocks(ctrl)

		ctx := context.Background()
		pseudonymID := "non-existent-pseudonym"

		// Set up gomock expectations
		mockPseudonymDAO.EXPECT().
			GetPseudonymByID(ctx, pseudonymID).
			Return(nil, nil).
			Times(1)

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
	})
}

// TestUserHandler_CreatePseudonym_Gomock tests the CreatePseudonym functionality using gomock
func TestUserHandler_CreatePseudonym_Gomock(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler, _, mockPseudonymDAO, _, _, _, _ := NewUserHandlerWithGomocks(ctrl)

		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()

		// Mock pseudonym data
		testPseudonym := fixtures.CreateTestPseudonym()
		testPseudonym.DisplayName = "NewPseudonym"

		// Set up gomock expectations
		mockPseudonymDAO.EXPECT().
			GetPseudonymByDisplayName(ctx, "NewPseudonym").
			Return(nil, nil).
			Times(1)

		mockPseudonymDAO.EXPECT().
			GenerateSlugFromDisplayName(ctx, "NewPseudonym").
			Return("new-pseudonym", nil).
			Times(1)

		mockPseudonymDAO.EXPECT().
			CreatePseudonymWithIdentityMapping(ctx, int64(1), "NewPseudonym").
			Return(testPseudonym, nil).
			Times(1)

		mockPseudonymDAO.EXPECT().
			UpdatePseudonym(ctx, "test-pseudonym-id", gomock.Any()).
			Return(nil).
			Times(1)

		mockPseudonymDAO.EXPECT().
			GetPseudonymByID(ctx, "test-pseudonym-id").
			Return(testPseudonym, nil).
			Times(1)

		// Set up mock expectation for UpdateLastActive
		mockPseudonymDAO.EXPECT().
			UpdateLastActive(ctx, "test-pseudonym-id").
			Return(nil).
			Times(1)

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
	})

	t.Run("DisplayNameAlreadyTaken", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler, _, mockPseudonymDAO, _, _, _, _ := NewUserHandlerWithGomocks(ctrl)

		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()

		// Mock existing pseudonym
		existingPseudonym := fixtures.CreateTestPseudonym()

		// Set up gomock expectations
		mockPseudonymDAO.EXPECT().
			GetPseudonymByDisplayName(ctx, "ExistingPseudonym").
			Return(existingPseudonym, nil).
			Times(1)

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
	})
}

// TestUserHandler_GetUserProfile_Gomock tests the GetUserProfile functionality using gomock
func TestUserHandler_GetUserProfile_Gomock(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("Success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		handler, mockUserDAO, mockPseudonymDAO, _, _, mockPostDAO, mockCommentDAO := NewUserHandlerWithGomocks(ctrl)

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

		// Set up gomock expectations
		mockUserDAO.EXPECT().
			GetUserByID(gomock.Any(), int64(1)).
			Return(testUser, nil).
			Times(1)

		mockPseudonymDAO.EXPECT().
			GetPseudonymsByUserID(gomock.Any(), int64(1), "test-pseudonym-id", "user", "authentication").
			Return([]*models.Pseudonym{testPseudonym}, nil).
			Times(1)

		// Set up mock expectations for post and comment counts
		mockPostDAO.EXPECT().
			CountPostsByPseudonym(gomock.Any(), "test-pseudonym-id").
			Return(int64(5), nil).
			Times(1)

		mockCommentDAO.EXPECT().
			CountCommentsByPseudonym(gomock.Any(), "test-pseudonym-id").
			Return(int64(10), nil).
			Times(1)

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
	})
}
