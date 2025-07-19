package handlers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockUserDAO is a mock implementation of UserDAOInterface
type MockUserDAO struct {
	mock.Mock
}

func (m *MockUserDAO) CreateUser(ctx context.Context, email, passwordHash string) (*models.User, error) {
	args := m.Called(ctx, email, passwordHash)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserDAO) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserDAO) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserDAO) UpdateUser(ctx context.Context, userID int64, updates *models.UserSetter) error {
	args := m.Called(ctx, userID, updates)
	return args.Error(0)
}

func (m *MockUserDAO) DeleteUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserDAO) ListUsers(ctx context.Context, limit, offset int) ([]*models.User, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserDAO) CountUsers(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockUserDAO) UpdateLastActive(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserDAO) SuspendUser(ctx context.Context, userID int64, reason string, expiresAt *time.Time) error {
	args := m.Called(ctx, userID, reason, expiresAt)
	return args.Error(0)
}

func (m *MockUserDAO) UnsuspendUser(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// MockSecurePseudonymDAO is a mock implementation of SecurePseudonymDAOInterface
type MockSecurePseudonymDAO struct {
	mock.Mock
}

func (m *MockSecurePseudonymDAO) CreatePseudonymWithIdentityMapping(ctx context.Context, userID int64, displayName string) (*models.Pseudonym, error) {
	args := m.Called(ctx, userID, displayName)
	return args.Get(0).(*models.Pseudonym), args.Error(1)
}

func (m *MockSecurePseudonymDAO) GetPseudonymByID(ctx context.Context, pseudonymID string) (*models.Pseudonym, error) {
	args := m.Called(ctx, pseudonymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Pseudonym), args.Error(1)
}

func (m *MockSecurePseudonymDAO) GetPseudonymByDisplayName(ctx context.Context, displayName string) (*models.Pseudonym, error) {
	args := m.Called(ctx, displayName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Pseudonym), args.Error(1)
}

func (m *MockSecurePseudonymDAO) GetPseudonymsByUserID(ctx context.Context, userID int64, roleName, scope string) ([]*models.Pseudonym, error) {
	args := m.Called(ctx, userID, roleName, scope)
	return args.Get(0).([]*models.Pseudonym), args.Error(1)
}

func (m *MockSecurePseudonymDAO) GetDefaultPseudonymByUserID(ctx context.Context, userID int64, roleName, scope string) (*models.Pseudonym, error) {
	args := m.Called(ctx, userID, roleName, scope)
	return args.Get(0).(*models.Pseudonym), args.Error(1)
}

func (m *MockSecurePseudonymDAO) UpdatePseudonym(ctx context.Context, pseudonymID string, updates *models.PseudonymSetter) error {
	args := m.Called(ctx, pseudonymID, updates)
	return args.Error(0)
}

func (m *MockSecurePseudonymDAO) DeletePseudonym(ctx context.Context, pseudonymID string) error {
	args := m.Called(ctx, pseudonymID)
	return args.Error(0)
}

func (m *MockSecurePseudonymDAO) VerifyPseudonymOwnership(ctx context.Context, pseudonymID string, userID int64, roleName, scope string) (bool, error) {
	args := m.Called(ctx, pseudonymID, userID, roleName, scope)
	return args.Bool(0), args.Error(1)
}

func (m *MockSecurePseudonymDAO) GetUserIDByPseudonym(ctx context.Context, pseudonymID, roleName, scope string) (int64, error) {
	args := m.Called(ctx, pseudonymID, roleName, scope)
	return args.Get(0).(int64), args.Error(1)
}

// MockUserPreferencesDAO is a mock implementation of UserPreferencesDAOInterface
type MockUserPreferencesDAO struct {
	mock.Mock
}

func (m *MockUserPreferencesDAO) CreateUserPreferences(ctx context.Context, userID int64, preferences *models.UserPreferenceSetter) (*models.UserPreference, error) {
	args := m.Called(ctx, userID, preferences)
	return args.Get(0).(*models.UserPreference), args.Error(1)
}

func (m *MockUserPreferencesDAO) GetUserPreferences(ctx context.Context, userID int64) (*models.UserPreference, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*models.UserPreference), args.Error(1)
}

func (m *MockUserPreferencesDAO) UpdateUserPreferences(ctx context.Context, userID int64, updates *models.UserPreferenceSetter) error {
	args := m.Called(ctx, userID, updates)
	return args.Error(0)
}

func (m *MockUserPreferencesDAO) UpsertUserPreferences(ctx context.Context, userID int64, preferences *models.UserPreferenceSetter) (*models.UserPreference, error) {
	args := m.Called(ctx, userID, preferences)
	return args.Get(0).(*models.UserPreference), args.Error(1)
}

// MockUserBlocksDAO is a mock implementation of UserBlocksDAOInterface
type MockUserBlocksDAO struct {
	mock.Mock
}

func (m *MockUserBlocksDAO) CreateUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string, blockedUserID int64) (*models.UserBlock, error) {
	args := m.Called(ctx, blockerPseudonymID, blockedPseudonymID, blockedUserID)
	return args.Get(0).(*models.UserBlock), args.Error(1)
}

func (m *MockUserBlocksDAO) GetUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (*models.UserBlock, error) {
	args := m.Called(ctx, blockerPseudonymID, blockedPseudonymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserBlock), args.Error(1)
}

func (m *MockUserBlocksDAO) GetUserBlocksByBlocker(ctx context.Context, blockerPseudonymID string) ([]*models.UserBlock, error) {
	args := m.Called(ctx, blockerPseudonymID)
	return args.Get(0).([]*models.UserBlock), args.Error(1)
}

func (m *MockUserBlocksDAO) GetUserBlocksByBlockedUser(ctx context.Context, blockedUserID int64) ([]*models.UserBlock, error) {
	args := m.Called(ctx, blockedUserID)
	return args.Get(0).([]*models.UserBlock), args.Error(1)
}

func (m *MockUserBlocksDAO) DeleteUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) error {
	args := m.Called(ctx, blockerPseudonymID, blockedPseudonymID)
	return args.Error(0)
}

func (m *MockUserBlocksDAO) DeleteUserBlockByID(ctx context.Context, blockID int64) error {
	args := m.Called(ctx, blockID)
	return args.Error(0)
}

func (m *MockUserBlocksDAO) IsUserBlocked(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (bool, error) {
	args := m.Called(ctx, blockerPseudonymID, blockedPseudonymID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserBlocksDAO) IsPseudonymBlockedByUser(ctx context.Context, blockerPseudonymID, blockedPseudonymID string, blockedUserID int64) (bool, error) {
	args := m.Called(ctx, blockerPseudonymID, blockedPseudonymID, blockedUserID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserBlocksDAO) IsUserBlockedAtFingerprintLevel(ctx context.Context, blockerPseudonymID string, blockedUserID int64) (bool, error) {
	args := m.Called(ctx, blockerPseudonymID, blockedUserID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserBlocksDAO) IsUserBlockedByAnyPseudonym(ctx context.Context, blockerUserID int64, blockedPseudonymID string) (bool, error) {
	args := m.Called(ctx, blockerUserID, blockedPseudonymID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserBlocksDAO) GetFingerprintLevelBlocks(ctx context.Context, blockedUserID int64) ([]*models.UserBlock, error) {
	args := m.Called(ctx, blockedUserID)
	return args.Get(0).([]*models.UserBlock), args.Error(1)
}

// MockPostDAO is a mock implementation of PostDAOInterface
type MockPostDAO struct {
	mock.Mock
}

func (m *MockPostDAO) CreatePost(ctx context.Context, subforumID int32, pseudonymID, title, content, postType string, url *string, isNSFW, isSpoiler bool) (*models.Post, error) {
	args := m.Called(ctx, subforumID, pseudonymID, title, content, postType, url, isNSFW, isSpoiler)
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostDAO) GetPostByID(ctx context.Context, postID int64) (*models.Post, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostDAO) GetPostsBySubforum(ctx context.Context, subforumID int32, page, limit int, sortField string, sortDesc bool) ([]*models.Post, error) {
	args := m.Called(ctx, subforumID, page, limit, sortField, sortDesc)
	return args.Get(0).([]*models.Post), args.Error(1)
}

func (m *MockPostDAO) CountPostsBySubforum(ctx context.Context, subforumID int32) (int64, error) {
	args := m.Called(ctx, subforumID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostDAO) CountPostsByPseudonym(ctx context.Context, pseudonymID string) (int64, error) {
	args := m.Called(ctx, pseudonymID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPostDAO) GetPostBySubforumAndSlug(ctx context.Context, subforumID int32, slug string) (*models.Post, error) {
	args := m.Called(ctx, subforumID, slug)
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostDAO) UpdatePostScore(ctx context.Context, postID int64, score, upvotes, downvotes int32) error {
	args := m.Called(ctx, postID, score, upvotes, downvotes)
	return args.Error(0)
}

func (m *MockPostDAO) IncrementViewCount(ctx context.Context, postID int64) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostDAO) UpdateCommentCount(ctx context.Context, postID int64, commentCount int32) error {
	args := m.Called(ctx, postID, commentCount)
	return args.Error(0)
}

func (m *MockPostDAO) SetLocked(ctx context.Context, postID int64, locked bool) error {
	args := m.Called(ctx, postID, locked)
	return args.Error(0)
}

func (m *MockPostDAO) SetSticky(ctx context.Context, postID int64, sticky bool) error {
	args := m.Called(ctx, postID, sticky)
	return args.Error(0)
}

func (m *MockPostDAO) SetRemoved(ctx context.Context, postID int64, removed bool) error {
	args := m.Called(ctx, postID, removed)
	return args.Error(0)
}

func (m *MockPostDAO) MarkPostAsDeletedByPseudonym(ctx context.Context, postID int64, pseudonymID string, reason string) error {
	args := m.Called(ctx, postID, pseudonymID, reason)
	return args.Error(0)
}

// MockCommentDAO is a mock implementation of CommentDAOInterface
type MockCommentDAO struct {
	mock.Mock
}

func (m *MockCommentDAO) CreateComment(ctx context.Context, postID int64, pseudonymID, content string, parentCommentID *int64) (*models.Comment, error) {
	args := m.Called(ctx, postID, pseudonymID, content, parentCommentID)
	return args.Get(0).(*models.Comment), args.Error(1)
}

func (m *MockCommentDAO) GetCommentByID(ctx context.Context, commentID int64) (*models.Comment, error) {
	args := m.Called(ctx, commentID)
	return args.Get(0).(*models.Comment), args.Error(1)
}

func (m *MockCommentDAO) GetCommentsByPost(ctx context.Context, postID int64) ([]*models.Comment, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).([]*models.Comment), args.Error(1)
}

func (m *MockCommentDAO) GetCommentsByPostWithNestedReplies(ctx context.Context, postID int64) ([]*models.Comment, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).([]*models.Comment), args.Error(1)
}

func (m *MockCommentDAO) CountCommentsByPost(ctx context.Context, postID int64) (int64, error) {
	args := m.Called(ctx, postID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCommentDAO) CountCommentsByPseudonym(ctx context.Context, pseudonymID string) (int64, error) {
	args := m.Called(ctx, pseudonymID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockCommentDAO) UpdateCommentScore(ctx context.Context, commentID int64, score, upvotes, downvotes int32) error {
	args := m.Called(ctx, commentID, score, upvotes, downvotes)
	return args.Error(0)
}

func (m *MockCommentDAO) MarkCommentAsDeletedByPseudonym(ctx context.Context, commentID int64, pseudonymID string, reason string) error {
	args := m.Called(ctx, commentID, pseudonymID, reason)
	return args.Error(0)
}

func (m *MockCommentDAO) DeleteCommentByUser(ctx context.Context, commentID int64, reason string) error {
	args := m.Called(ctx, commentID, reason)
	return args.Error(0)
}

// Helper function to create a test UserHandler with mocks
func createTestUserHandler() (*UserHandler, *MockUserDAO, *MockSecurePseudonymDAO, *MockUserPreferencesDAO, *MockUserBlocksDAO, *MockPostDAO, *MockCommentDAO) {
	mockUserDAO := &MockUserDAO{}
	mockSecurePseudonymDAO := &MockSecurePseudonymDAO{}
	mockUserPreferencesDAO := &MockUserPreferencesDAO{}
	mockUserBlocksDAO := &MockUserBlocksDAO{}
	mockPostDAO := &MockPostDAO{}
	mockCommentDAO := &MockCommentDAO{}

	// Create a mock IBE system
	mockIBESystem := &ibe.IBESystem{}

	handler := NewUserHandlerWithDependencies(
		mockUserDAO,
		mockSecurePseudonymDAO,
		mockUserPreferencesDAO,
		mockUserBlocksDAO,
		mockPostDAO,
		mockCommentDAO,
		mockIBESystem,
	)

	return handler, mockUserDAO, mockSecurePseudonymDAO, mockUserPreferencesDAO, mockUserBlocksDAO, mockPostDAO, mockCommentDAO
}

// Helper function to create a test user context
func createTestUserContext() *middleware.UserContext {
	return &middleware.UserContext{
		UserID:            1,
		Email:             "test@example.com",
		ActivePseudonymID: "test-pseudonym-id",
		DisplayName:       "TestUser",
		Roles:             []string{"user"},
		Capabilities:      []string{"user"},
		TokenType:         "jwt",
	}
}

// Helper function to create a test pseudonym
func createTestPseudonym() *models.Pseudonym {
	return &models.Pseudonym{
		PseudonymID:         "test-pseudonym-id",
		DisplayName:         "TestUser",
		Bio:                 sql.Null[string]{V: "Test bio", Valid: true},
		WebsiteURL:          sql.Null[string]{V: "https://example.com", Valid: true},
		KarmaScore:          sql.Null[int32]{V: 100, Valid: true},
		ShowKarma:           sql.Null[bool]{V: true, Valid: true},
		AllowDirectMessages: sql.Null[bool]{V: true, Valid: true},
		IsActive:            sql.Null[bool]{V: true, Valid: true},
		CreatedAt:           sql.Null[time.Time]{V: time.Now(), Valid: true},
		LastActiveAt:        sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
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
		testPseudonym := createTestPseudonym()

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
		inactivePseudonym := createTestPseudonym()
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
		userCtx := createTestUserContext()

		// Mock pseudonym data
		testPseudonym := createTestPseudonym()
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
		userCtx := createTestUserContext()

		// Mock existing pseudonym
		existingPseudonym := createTestPseudonym()

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
		handler, mockUserDAO, mockSecurePseudonymDAO, _, _, mockPostDAO, mockCommentDAO := createTestUserHandler()

		ctx := context.Background()
		userCtx := createTestUserContext()

		// Mock user data
		testUser := &models.User{
			UserID:   1,
			Email:    "test@example.com",
			IsActive: sql.Null[bool]{V: true, Valid: true},
		}

		// Mock pseudonym data
		testPseudonym := createTestPseudonym()
		pseudonyms := []*models.Pseudonym{testPseudonym}

		// Set up mock expectations
		mockUserDAO.On("GetUserByID", ctx, int64(1)).Return(testUser, nil)
		mockSecurePseudonymDAO.On("GetPseudonymsByUserID", ctx, int64(1), "user", "authentication").Return(pseudonyms, nil)
		mockPostDAO.On("CountPostsByPseudonym", ctx, "test-pseudonym-id").Return(int64(5), nil)
		mockCommentDAO.On("CountCommentsByPseudonym", ctx, "test-pseudonym-id").Return(int64(10), nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &middleware.AuthInput{
			Authorization: "Bearer " + token,
		}

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
		mockPostDAO.AssertExpectations(t)
		mockCommentDAO.AssertExpectations(t)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		handler, mockUserDAO, _, _, _, _, _ := createTestUserHandler()

		ctx := context.Background()
		userCtx := createTestUserContext()

		// Set up mock expectations
		mockUserDAO.On("GetUserByID", ctx, int64(1)).Return(nil, nil)

		// Generate JWT token
		token, err := middleware.GenerateJWT(userCtx, "test-secret", 24*time.Hour)
		require.NoError(t, err)

		// Create input
		input := &middleware.AuthInput{
			Authorization: "Bearer " + token,
		}

		// Call the method
		result, err := handler.GetUserProfile(ctx, input)

		// Assertions
		require.Error(t, err)
		require.Nil(t, result)
		require.Contains(t, err.Error(), "user not found")

		// Verify mock expectations
		mockUserDAO.AssertExpectations(t)
	})
}
