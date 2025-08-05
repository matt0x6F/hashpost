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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper function to create a test UserHandler with mocks
func NewUserHandlerWithMocks() (*UserHandler, *mocks.MockUserDAO, *mocks.MockPseudonymDAO, *mocks.MockUserPreferencesDAO, *mocks.MockUserBlocksDAO, *mocks.MockPostDAO, *mocks.MockCommentDAO) {
	mockUserDAO := &mocks.MockUserDAO{}
	mockPseudonymDAO := mocks.NewMockPseudonymDAO()
	mockUserPreferencesDAO := &mocks.MockUserPreferencesDAO{}
	mockUserBlocksDAO := &mocks.MockUserBlocksDAO{}
	mockPostDAO := mocks.NewMockPostDAO()
	mockCommentDAO := mocks.NewMockCommentDAO()

	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

	handler := NewUserHandler(
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

func TestUserHandler_GetPseudonymProfile(t *testing.T) {
	handler, _, mockPseudonymDAO, _, _, mockPostDAO, mockCommentDAO := NewUserHandlerWithMocks()

	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		pseudonymID := "test-pseudonym-id"

		// Mock pseudonym data
		testPseudonym := fixtures.CreateTestPseudonym()

		// Set up mock expectations
		mockPseudonymDAO.On("GetPseudonymByID", ctx, pseudonymID).Return(testPseudonym, nil)
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
		mockPseudonymDAO.AssertExpectations(t)
		mockPostDAO.AssertExpectations(t)
		mockCommentDAO.AssertExpectations(t)
	})

	t.Run("PseudonymNotFound", func(t *testing.T) {
		ctx := context.Background()
		pseudonymID := "non-existent-pseudonym"

		// Set up mock expectations
		mockPseudonymDAO.On("GetPseudonymByID", ctx, pseudonymID).Return(nil, nil)

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
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("PseudonymInactive", func(t *testing.T) {
		ctx := context.Background()
		pseudonymID := "inactive-pseudonym"

		// Mock inactive pseudonym
		inactivePseudonym := fixtures.CreateTestPseudonym()
		inactivePseudonym.IsActive = sql.Null[bool]{V: false, Valid: true}

		// Set up mock expectations
		mockPseudonymDAO.On("GetPseudonymByID", ctx, pseudonymID).Return(inactivePseudonym, nil)

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
		mockPseudonymDAO.AssertExpectations(t)
	})
}

func TestUserHandler_CreatePseudonym(t *testing.T) {
	handler, mockUserDAO, mockPseudonymDAO, _, _, _, _ := NewUserHandlerWithMocks()

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
		mockPseudonymDAO.On("GetPseudonymByDisplayName", ctx, "NewPseudonym").Return(nil, nil)
		mockPseudonymDAO.On("GenerateSlugFromDisplayName", ctx, "NewPseudonym").Return("new-pseudonym", nil)
		mockPseudonymDAO.On("CreatePseudonymWithIdentityMapping", ctx, int64(1), "NewPseudonym").Return(testPseudonym, nil)
		mockPseudonymDAO.On("UpdatePseudonym", ctx, "test-pseudonym-id", mock.AnythingOfType("*models.PseudonymSetter")).Return(nil)
		mockPseudonymDAO.On("GetPseudonymByID", ctx, "test-pseudonym-id").Return(testPseudonym, nil)

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
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("DisplayNameAlreadyTaken", func(t *testing.T) {
		ctx := context.Background()
		userCtx := fixtures.CreateTestUserContext()

		// Mock existing pseudonym
		existingPseudonym := fixtures.CreateTestPseudonym()

		// Set up mock expectations
		mockPseudonymDAO.On("GetPseudonymByDisplayName", ctx, "ExistingPseudonym").Return(existingPseudonym, nil)

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
		mockPseudonymDAO.AssertExpectations(t)
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
		mockPseudonymDAO := mocks.NewMockPseudonymDAO()

		mockUserDAO.On("GetUserByID", mock.Anything, int64(1)).Return(testUser, nil)
		mockPseudonymDAO.On("GetPseudonymsByUserID", mock.Anything, int64(1), "user", "authentication").Return([]*models.Pseudonym{testPseudonym}, nil)

		// Create mock PostDAO and CommentDAO
		mockPostDAO := mocks.NewMockPostDAO()
		mockCommentDAO := mocks.NewMockCommentDAO()

		// Set up mock expectations for post and comment counts
		mockPostDAO.On("CountPostsByPseudonym", mock.Anything, "test-pseudonym-id").Return(int64(5), nil)
		mockCommentDAO.On("CountCommentsByPseudonym", mock.Anything, "test-pseudonym-id").Return(int64(10), nil)

		// Create handler with mocked dependencies
		handler := NewUserHandler(
			nil, // nil db for testing
			mockUserDAO,
			mockPseudonymDAO,
			nil, // Mock UserPreferencesDAO
			nil, // Mock UserBlocksDAO
			mockPostDAO,
			mockCommentDAO,
			ibe.NewIBESystemWithOptions(ibe.IBEOptions{}),
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
		mockPseudonymDAO.AssertExpectations(t)
	})
}

// TestNewUserHandler tests the main constructor function
func TestNewUserHandler(t *testing.T) {
	t.Run("NewUserHandlerSuccess", func(t *testing.T) {
		// Create mock dependencies
		mockUserDAO := &mocks.MockUserDAO{}
		mockPseudonymDAO := &mocks.MockPseudonymDAO{}
		mockUserPreferencesDAO := &mocks.MockUserPreferencesDAO{}
		mockUserBlocksDAO := &mocks.MockUserBlocksDAO{}
		mockPostDAO := &mocks.MockPostDAO{}
		mockCommentDAO := &mocks.MockCommentDAO{}
		mockIBESystem := &ibe.IBESystem{}

		// Create handler with dependencies
		handler := NewUserHandler(
			nil, // nil db for testing
			mockUserDAO,
			mockPseudonymDAO,
			mockUserPreferencesDAO,
			mockUserBlocksDAO,
			mockPostDAO,
			mockCommentDAO,
			mockIBESystem,
		)

		// Verify handler is created
		assert.NotNil(t, handler)
	})
}

// TestUserHandler_GetUserPreferences tests the get user preferences functionality
func TestUserHandler_GetUserPreferences(t *testing.T) {
	t.Run("GetUserPreferencesSuccess", func(t *testing.T) {
		handler, _, _, mockUserPreferencesDAO, _, _, _ := NewUserHandlerWithMocks()

		// Test data
		userID := int64(1)
		testPreferences := &models.UserPreference{
			UserID:             userID,
			Timezone:           sql.Null[string]{V: "America/New_York", Valid: true},
			Language:           sql.Null[string]{V: "es", Valid: true},
			Theme:              sql.Null[string]{V: "dark", Valid: true},
			EmailNotifications: sql.Null[bool]{V: false, Valid: true},
			PushNotifications:  sql.Null[bool]{V: true, Valid: true},
			AutoHideNSFW:       sql.Null[bool]{V: false, Valid: true},
			AutoHideSpoilers:   sql.Null[bool]{V: true, Valid: true},
		}

		// Mock preferences retrieval
		mockUserPreferencesDAO.On("GetUserPreferences", mock.Anything, userID).Return(testPreferences, nil)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.UserPreferencesInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(userID, "test-pseudonym-id"),
			},
		}

		// Set up user context
		userCtx := fixtures.CreateTestUserContext()
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Call handler
		response, err := handler.GetUserPreferences(ctx, input)

		// Verify response
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "America/New_York", response.Body.Timezone)
		assert.Equal(t, "es", response.Body.Language)
		assert.Equal(t, "dark", response.Body.Theme)
		assert.False(t, response.Body.EmailNotifications)
		assert.True(t, response.Body.PushNotifications)
		assert.False(t, response.Body.AutoHideNSFW)
		assert.True(t, response.Body.AutoHideSpoilers)

		// Verify DAO calls
		mockUserPreferencesDAO.AssertExpectations(t)
	})

	t.Run("GetUserPreferencesNoPreferences", func(t *testing.T) {
		handler, _, _, mockUserPreferencesDAO, _, _, _ := NewUserHandlerWithMocks()

		// Test data
		userID := int64(1)

		// Mock preferences retrieval - no preferences found
		mockUserPreferencesDAO.On("GetUserPreferences", mock.Anything, userID).Return(nil, nil)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.UserPreferencesInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(userID, "test-pseudonym-id"),
			},
		}

		// Set up user context
		userCtx := fixtures.CreateTestUserContext()
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Call handler
		response, err := handler.GetUserPreferences(ctx, input)

		// Verify response - should return defaults
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "UTC", response.Body.Timezone)
		assert.Equal(t, "en", response.Body.Language)
		assert.Equal(t, "light", response.Body.Theme)
		assert.True(t, response.Body.EmailNotifications)
		assert.True(t, response.Body.PushNotifications)
		assert.True(t, response.Body.AutoHideNSFW)
		assert.True(t, response.Body.AutoHideSpoilers)

		// Verify DAO calls
		mockUserPreferencesDAO.AssertExpectations(t)
	})

	t.Run("GetUserPreferencesDatabaseError", func(t *testing.T) {
		handler, _, _, mockUserPreferencesDAO, _, _, _ := NewUserHandlerWithMocks()

		// Test data
		userID := int64(1)

		// Mock database error
		mockUserPreferencesDAO.On("GetUserPreferences", mock.Anything, userID).Return(nil, assert.AnError)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.UserPreferencesInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(userID, "test-pseudonym-id"),
			},
		}

		// Set up user context
		userCtx := fixtures.CreateTestUserContext()
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Call handler
		response, err := handler.GetUserPreferences(ctx, input)

		// Verify error
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get user preferences")

		// Verify DAO calls
		mockUserPreferencesDAO.AssertExpectations(t)
	})

	t.Run("GetUserPreferencesNoAuthentication", func(t *testing.T) {
		handler, _, _, _, _, _, _ := NewUserHandlerWithMocks()

		// Create input without authentication
		input := &struct {
			middleware.AuthInput
			apimodels.UserPreferencesInput
		}{}

		// Call handler
		response, err := handler.GetUserPreferences(context.Background(), input)

		// Verify error
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "Authentication required")
	})
}

// TestUserHandler_UpdateUserPreferences tests the update user preferences functionality
func TestUserHandler_UpdateUserPreferences(t *testing.T) {
	t.Run("UpdateUserPreferencesSuccess", func(t *testing.T) {
		handler, _, _, mockUserPreferencesDAO, _, _, _ := NewUserHandlerWithMocks()

		// Test data
		userID := int64(1)
		timezone := "Europe/London"
		language := "fr"
		theme := "dark"
		emailNotifications := false
		pushNotifications := true
		autoHideNSFW := false
		autoHideSpoilers := true

		updatedPreferences := &models.UserPreference{
			UserID:             userID,
			Timezone:           sql.Null[string]{V: timezone, Valid: true},
			Language:           sql.Null[string]{V: language, Valid: true},
			Theme:              sql.Null[string]{V: theme, Valid: true},
			EmailNotifications: sql.Null[bool]{V: emailNotifications, Valid: true},
			PushNotifications:  sql.Null[bool]{V: pushNotifications, Valid: true},
			AutoHideNSFW:       sql.Null[bool]{V: autoHideNSFW, Valid: true},
			AutoHideSpoilers:   sql.Null[bool]{V: autoHideSpoilers, Valid: true},
		}

		// Mock preferences update
		mockUserPreferencesDAO.On("UpsertUserPreferences", mock.Anything, userID, mock.AnythingOfType("*models.UserPreferenceSetter")).Return(updatedPreferences, nil)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.UserPreferencesInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(userID, "test-pseudonym-id"),
			},
			UserPreferencesInput: apimodels.UserPreferencesInput{
				Body: apimodels.UserPreferencesBody{
					Timezone:           timezone,
					Language:           language,
					Theme:              theme,
					EmailNotifications: &emailNotifications,
					PushNotifications:  &pushNotifications,
					AutoHideNSFW:       &autoHideNSFW,
					AutoHideSpoilers:   &autoHideSpoilers,
				},
			},
		}

		// Set up user context
		userCtx := fixtures.CreateTestUserContext()
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Call handler
		response, err := handler.UpdateUserPreferences(ctx, input)

		// Verify response
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, timezone, response.Body.Timezone)
		assert.Equal(t, language, response.Body.Language)
		assert.Equal(t, theme, response.Body.Theme)
		assert.Equal(t, emailNotifications, response.Body.EmailNotifications)
		assert.Equal(t, pushNotifications, response.Body.PushNotifications)
		assert.Equal(t, autoHideNSFW, response.Body.AutoHideNSFW)
		assert.Equal(t, autoHideSpoilers, response.Body.AutoHideSpoilers)

		// Verify DAO calls
		mockUserPreferencesDAO.AssertExpectations(t)
	})

	t.Run("UpdateUserPreferencesPartialUpdate", func(t *testing.T) {
		handler, _, _, mockUserPreferencesDAO, _, _, _ := NewUserHandlerWithMocks()

		// Test data
		userID := int64(1)
		timezone := "Asia/Tokyo"
		theme := "light"

		updatedPreferences := &models.UserPreference{
			UserID:   userID,
			Timezone: sql.Null[string]{V: timezone, Valid: true},
			Theme:    sql.Null[string]{V: theme, Valid: true},
		}

		// Mock preferences update
		mockUserPreferencesDAO.On("UpsertUserPreferences", mock.Anything, userID, mock.AnythingOfType("*models.UserPreferenceSetter")).Return(updatedPreferences, nil)

		// Create input with only some fields
		input := &struct {
			middleware.AuthInput
			apimodels.UserPreferencesInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(userID, "test-pseudonym-id"),
			},
			UserPreferencesInput: apimodels.UserPreferencesInput{
				Body: apimodels.UserPreferencesBody{
					Timezone: timezone,
					Theme:    theme,
				},
			},
		}

		// Set up user context
		userCtx := fixtures.CreateTestUserContext()
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Call handler
		response, err := handler.UpdateUserPreferences(ctx, input)

		// Verify response
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, timezone, response.Body.Timezone)
		assert.Equal(t, theme, response.Body.Theme)

		// Verify DAO calls
		mockUserPreferencesDAO.AssertExpectations(t)
	})

	t.Run("UpdateUserPreferencesDatabaseError", func(t *testing.T) {
		handler, _, _, mockUserPreferencesDAO, _, _, _ := NewUserHandlerWithMocks()

		// Test data
		userID := int64(1)
		timezone := "UTC"

		// Mock database error
		mockUserPreferencesDAO.On("UpsertUserPreferences", mock.Anything, userID, mock.AnythingOfType("*models.UserPreferenceSetter")).Return(nil, assert.AnError)

		// Create input
		input := &struct {
			middleware.AuthInput
			apimodels.UserPreferencesInput
		}{
			AuthInput: middleware.AuthInput{
				Authorization: "Bearer " + fixtures.MustGenerateTestJWTToken(userID, "test-pseudonym-id"),
			},
			UserPreferencesInput: apimodels.UserPreferencesInput{
				Body: apimodels.UserPreferencesBody{
					Timezone: timezone,
				},
			},
		}

		// Set up user context
		userCtx := fixtures.CreateTestUserContext()
		ctx := context.WithValue(context.Background(), middleware.UserContextKeyValue, userCtx)

		// Call handler
		response, err := handler.UpdateUserPreferences(ctx, input)

		// Verify error
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to update user preferences")

		// Verify DAO calls
		mockUserPreferencesDAO.AssertExpectations(t)
	})

	t.Run("UpdateUserPreferencesNoAuthentication", func(t *testing.T) {
		handler, _, _, _, _, _, _ := NewUserHandlerWithMocks()

		// Create input without authentication
		input := &struct {
			middleware.AuthInput
			apimodels.UserPreferencesInput
		}{}

		// Call handler
		response, err := handler.UpdateUserPreferences(context.Background(), input)

		// Verify error
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "Authentication required")
	})
}
