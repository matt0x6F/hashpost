package handlers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gomock "go.uber.org/mock/gomock"
)

// NewSearchHandlerWithGomocks creates a SearchHandler with gomock dependencies
func NewSearchHandlerWithGomocks(t *testing.T) (*handlers.SearchHandler, *dao.MockPostDAOInterface, *dao.MockUserDAOInterface, *dao.MockSubforumDAOInterface, *dao.MockPseudonymDAOInterface, *dao.MockPermissionDAOInterface) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)

	// Create a test IBE system for testing
	testIBESystem := ibe.NewTestIBESystem()

	// Create handler with mocked DAOs
	handler := handlers.NewSearchHandler(
		nil, // nil db for testing
		mockPostDAO,
		mockUserDAO,
		mockSubforumDAO,
		mockPseudonymDAO,
		mockPermissionDAO,
		testIBESystem,
	)

	return handler, mockPostDAO, mockUserDAO, mockSubforumDAO, mockPseudonymDAO, mockPermissionDAO
}

// createTestSearchContext creates a context with user information
func createTestSearchContext(t *testing.T, userID int64, activePseudonymID string, displayName string) context.Context {
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false, // roles and capabilities deprecated
	}

	return context.WithValue(context.Background(), middleware.UserContextKeyValue, user)
}

func createTestPlatformAdminContext(t *testing.T, userID int64, activePseudonymID string, displayName string) context.Context {
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "admin@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false, // roles and capabilities deprecated
	}

	return context.WithValue(context.Background(), middleware.UserContextKeyValue, user)
}

// createAuthenticatedSearchUsersInput creates a SearchUsersInput with a valid JWT token for testing
func createAuthenticatedSearchUsersInput(userID int64, activePseudonymID string, displayName string, query string) *models.SearchUsersInput {
	// Create a user context
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "admin@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false, // roles and capabilities deprecated
	}

	// Generate a JWT token
	token, _ := middleware.GenerateJWT(user, "test-secret", time.Hour)

	return &models.SearchUsersInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
		Query: query,
		Page:  1,
		Limit: 25,
	}
}

// TestSearchHandler_SearchPosts tests the search posts functionality using gomock
func TestSearchHandler_SearchPosts(t *testing.T) {
	t.Run("SearchPostsSuccess", func(t *testing.T) {
		handler, mockPostDAO, _, mockSubforumDAO, mockPseudonymDAO, _ := NewSearchHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		query := "test post"

		// Create context with user
		ctx := createTestSearchContext(t, userID, activePseudonymID, displayName)

		// Mock subforums
		mockSubforums := []*dbmodels.Subforum{
			{
				SubforumID: 1,
				Name:       "General",
			},
		}
		mockSubforumDAO.EXPECT().ListSubforums(gomock.Any()).Return(mockSubforums, nil).Times(2)

		// Mock posts
		now := time.Now()
		mockPosts := []*dbmodels.Post{
			{
				PostID:      1,
				Title:       "Test Post 1",
				Content:     sql.Null[string]{V: "This is a test post about testing", Valid: true},
				SubforumID:  1,
				PseudonymID: "user-pseudonym-123",
				CreatedAt:   sql.Null[time.Time]{V: now, Valid: true},
			},
			{
				PostID:      2,
				Title:       "Another Test Post",
				Content:     sql.Null[string]{V: "Another test post with test content", Valid: true},
				SubforumID:  1,
				PseudonymID: "user-pseudonym-123",
				CreatedAt:   sql.Null[time.Time]{V: now.Add(-time.Hour), Valid: true},
			},
		}

		// Mock post retrieval calls - search uses 100, counting uses 1000
		mockPostDAO.EXPECT().GetPostsBySubforum(gomock.Any(), int32(1), 1, 100, "created_at", true).Return(mockPosts, nil).Times(1)
		mockPostDAO.EXPECT().GetPostsBySubforum(gomock.Any(), int32(1), 1, 1000, "created_at", true).Return(mockPosts, nil).Times(1)

		// Mock subforum and pseudonym calls for each post
		mockSubforumDAO.EXPECT().GetSubforumByID(gomock.Any(), int32(1)).Return(mockSubforums[0], nil).Times(2)
		mockPseudonymDAO.EXPECT().GetPseudonymByID(gomock.Any(), "user-pseudonym-123").Return(&dbmodels.Pseudonym{
			PseudonymID: "user-pseudonym-123",
			DisplayName: "TestUser",
		}, nil).Times(2)

		// Create input
		input := &models.SearchPostsInput{
			Query: query,
			Page:  1,
			Limit: 25,
		}

		// Call handler
		response, err := handler.SearchPosts(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, query, response.Body.Query)
		assert.Len(t, response.Body.Posts, 2)
		assert.Equal(t, 1, response.Body.Pagination.Page)
		assert.Equal(t, 25, response.Body.Pagination.Limit)
		assert.Equal(t, 2, response.Body.Pagination.Total)
		assert.Equal(t, 1, response.Body.Pagination.Pages)
	})

	t.Run("SearchPostsEmptyQuery", func(t *testing.T) {
		handler, _, _, _, _, _ := NewSearchHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		query := ""

		// Create context with user
		ctx := createTestSearchContext(t, userID, activePseudonymID, displayName)

		// Create input
		input := &models.SearchPostsInput{
			Query: query,
			Page:  1,
			Limit: 25,
		}

		// Call handler
		response, err := handler.SearchPosts(ctx, input)

		// Assertions - empty query should return error
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "search query is required")
	})

	t.Run("SearchPostsDatabaseError", func(t *testing.T) {
		handler, _, _, mockSubforumDAO, _, _ := NewSearchHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		query := "test"

		// Create context with user
		ctx := createTestSearchContext(t, userID, activePseudonymID, displayName)

		// Mock database error
		mockSubforumDAO.EXPECT().ListSubforums(gomock.Any()).Return(nil, assert.AnError).Times(1)

		// Create input
		input := &models.SearchPostsInput{
			Query: query,
			Page:  1,
			Limit: 25,
		}

		// Call handler
		response, err := handler.SearchPosts(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get subforums")
	})
}

// TestSearchHandler_SearchUsers tests the search users functionality using gomock
func TestSearchHandler_SearchUsers(t *testing.T) {
	// Set up global auth middleware for tests
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, nil, nil)
	middleware.SetGlobalAuthMiddleware(authMiddleware)

	t.Run("SearchUsersSuccess", func(t *testing.T) {
		handler, mockPostDAO, mockUserDAO, _, mockPseudonymDAO, mockPermissionDAO := NewSearchHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "admin-pseudonym-123"
		displayName := "AdminUser"
		query := "john"

		// Create context with platform admin user
		ctx := createTestPlatformAdminContext(t, userID, activePseudonymID, displayName)

		// Mock permission check to return true for platform admin capability
		mockPermissionDAO.EXPECT().HasUnifiedCapability(gomock.Any(), userID, activePseudonymID, constants.CapabilitySystemAdmin, (*int32)(nil)).Return(true, nil).Times(1)

		// Mock users
		now := time.Now()
		mockUsers := []*dbmodels.User{
			{
				UserID:       1,
				Email:        "john.doe@example.com",
				IsActive:     sql.Null[bool]{V: true, Valid: true},
				CreatedAt:    sql.Null[time.Time]{V: now, Valid: true},
				LastActiveAt: sql.Null[time.Time]{V: now, Valid: true},
			},
			{
				UserID:       2,
				Email:        "johnny@example.com",
				IsActive:     sql.Null[bool]{V: true, Valid: true},
				CreatedAt:    sql.Null[time.Time]{V: now.Add(-time.Hour), Valid: true},
				LastActiveAt: sql.Null[time.Time]{V: now.Add(-time.Hour), Valid: true},
			},
		}

		// Mock pseudonyms
		mockPseudonym1 := &dbmodels.Pseudonym{
			PseudonymID: "pseudonym-1",
			DisplayName: "john_doe",
		}
		mockPseudonym2 := &dbmodels.Pseudonym{
			PseudonymID: "pseudonym-2",
			DisplayName: "johnny_smith",
		}

		// Mock database calls - called twice (once for search, once for counting)
		mockUserDAO.EXPECT().ListUsers(gomock.Any(), 1000, 0).Return(mockUsers, nil).Times(2)

		// Mock IBE system calls - GetIBESystemSalt is called twice in search + twice in counting (4 total)
		mockPseudonymDAO.EXPECT().GetIBESystemSalt().Return("test-salt").Times(4)
		mockPseudonymDAO.EXPECT().GenerateFingerprintForEmail("john.doe@example.com").Return("fingerprint1").Times(2)
		mockPseudonymDAO.EXPECT().GenerateFingerprintForEmail("johnny@example.com").Return("fingerprint2").Times(2)

		// Mock pseudonym retrieval calls - called twice for each user in search + twice in counting (4 total each)
		mockPseudonymDAO.EXPECT().GetPseudonymsByRealIdentityDirect(gomock.Any(), "john.doe@example.com").Return([]*dbmodels.Pseudonym{mockPseudonym1}, nil).Times(4)
		mockPseudonymDAO.EXPECT().GetPseudonymsByRealIdentityDirect(gomock.Any(), "johnny@example.com").Return([]*dbmodels.Pseudonym{mockPseudonym2}, nil).Times(4)

		// Mock karma score calculation calls - called for each user in the search results
		mockPostDAO.EXPECT().GetPostsBySubforum(gomock.Any(), int32(0), 1, 1000, "created_at", true).Return([]*dbmodels.Post{}, nil).Times(2)

		// Create authenticated input
		input := createAuthenticatedSearchUsersInput(userID, activePseudonymID, displayName, query)

		// Call handler
		response, err := handler.SearchUsers(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, query, response.Body.Query)
		assert.Len(t, response.Body.Users, 2)
		assert.Equal(t, 1, response.Body.Pagination.Page)
		assert.Equal(t, 25, response.Body.Pagination.Limit)
		assert.Equal(t, 2, response.Body.Pagination.Total)
		assert.Equal(t, 1, response.Body.Pagination.Pages)

		// Verify first user
		assert.Equal(t, "pseudonym-1", response.Body.Users[0].PseudonymID)
		assert.Equal(t, "john_doe", response.Body.Users[0].DisplayName)
		assert.NotEmpty(t, response.Body.Users[0].CreatedAt)
	})

	t.Run("SearchUsersEmptyQuery", func(t *testing.T) {
		handler, _, _, _, _, mockPermissionDAO := NewSearchHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "admin-pseudonym-123"
		displayName := "AdminUser"

		// Create context with platform admin user
		ctx := createTestPlatformAdminContext(t, userID, activePseudonymID, displayName)

		// Mock permission check to return true for platform admin capability
		mockPermissionDAO.EXPECT().HasUnifiedCapability(gomock.Any(), userID, activePseudonymID, constants.CapabilitySystemAdmin, (*int32)(nil)).Return(true, nil).Times(1)

		// Create authenticated input with empty query
		input := createAuthenticatedSearchUsersInput(userID, activePseudonymID, displayName, "")

		// Call handler
		response, err := handler.SearchUsers(ctx, input)

		// Assertions - empty query should return error
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "search query is required")
	})

	t.Run("SearchUsersInsufficientPermissions", func(t *testing.T) {
		handler, _, _, _, _, mockPermissionDAO := NewSearchHandlerWithGomocks(t)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "RegularUser"

		// Create context with regular user
		ctx := createTestSearchContext(t, userID, activePseudonymID, displayName)

		// Mock permission check to return false for platform admin capability
		mockPermissionDAO.EXPECT().HasUnifiedCapability(gomock.Any(), userID, activePseudonymID, constants.CapabilitySystemAdmin, (*int32)(nil)).Return(false, nil).Times(1)

		// Create authenticated input
		input := createAuthenticatedSearchUsersInput(userID, activePseudonymID, displayName, "test")

		// Call handler
		response, err := handler.SearchUsers(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "insufficient permissions")
	})
}

// TestSearchHandler_NewSearchHandler tests the search handler constructor using gomock
func TestSearchHandler_NewSearchHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock DAOs
	mockPostDAO := dao.NewMockPostDAOInterface(ctrl)
	mockUserDAO := dao.NewMockUserDAOInterface(ctrl)
	mockSubforumDAO := dao.NewMockSubforumDAOInterface(ctrl)
	mockPseudonymDAO := dao.NewMockPseudonymDAOInterface(ctrl)
	mockPermissionDAO := dao.NewMockPermissionDAOInterface(ctrl)

	// Create a test IBE system
	testIBESystem := ibe.NewTestIBESystem()

	// Create handler with mocked DAOs
	handler := handlers.NewSearchHandler(
		nil, // nil db for testing
		mockPostDAO,
		mockUserDAO,
		mockSubforumDAO,
		mockPseudonymDAO,
		mockPermissionDAO,
		testIBESystem,
	)

	// Assertions
	assert.NotNil(t, handler)
	// Note: Fields are unexported, so we can't access them directly in tests
	// The constructor test verifies the handler was created successfully
}
