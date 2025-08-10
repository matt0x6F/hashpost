package handlers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// NewSearchHandlerWithMocks creates a SearchHandler with mocked dependencies
func NewSearchHandlerWithMocks() (*handlers.SearchHandler, *mocks.MockPostDAO, *mocks.MockUserDAO, *mocks.MockSubforumDAO, *mocks.MockPseudonymDAO, *mocks.MockPermissionDAO) {
	mockPostDAO := &mocks.MockPostDAO{}
	mockUserDAO := &mocks.MockUserDAO{}
	mockSubforumDAO := &mocks.MockSubforumDAO{}
	mockPseudonymDAO := &mocks.MockPseudonymDAO{}
	mockPermissionDAO := &mocks.MockPermissionDAO{}

	// Create handler with mocked DAOs
	handler := handlers.NewSearchHandler(
		nil, // nil db for testing
		mockPostDAO,
		mockUserDAO,
		mockSubforumDAO,
		mockPseudonymDAO,
		mockPermissionDAO,
	)

	return handler, mockPostDAO, mockUserDAO, mockSubforumDAO, mockPseudonymDAO, mockPermissionDAO
}

// createTestContext creates a context with user information
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

// TestSearchHandler_SearchPosts tests the search posts functionality
func TestSearchHandler_SearchPosts(t *testing.T) {
	t.Run("SearchPostsSuccess", func(t *testing.T) {
		handler, mockPostDAO, _, mockSubforumDAO, mockPseudonymDAO, _ := NewSearchHandlerWithMocks()

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		query := "golang concurrency"

		// Create context with user
		ctx := createTestSearchContext(t, userID, activePseudonymID, displayName)

		// Mock posts
		now := time.Now()
		mockPosts := []*dbmodels.Post{
			{
				PostID:       1,
				Title:        "Understanding Golang Concurrency",
				Content:      sql.Null[string]{V: "This post discusses golang concurrency patterns...", Valid: true},
				PseudonymID:  "author-pseudonym-1",
				SubforumID:   1,
				Score:        sql.Null[int32]{V: 1250, Valid: true},
				CommentCount: sql.Null[int32]{V: 45, Valid: true},
				CreatedAt:    sql.Null[time.Time]{V: now, Valid: true},
			},
			{
				PostID:       2,
				Title:        "Advanced Concurrency in Go",
				Content:      sql.Null[string]{V: "Advanced patterns for golang concurrency...", Valid: true},
				PseudonymID:  "author-pseudonym-2",
				SubforumID:   1,
				Score:        sql.Null[int32]{V: 800, Valid: true},
				CommentCount: sql.Null[int32]{V: 23, Valid: true},
				CreatedAt:    sql.Null[time.Time]{V: now.Add(-time.Hour), Valid: true},
			},
		}

		// Mock subforum
		mockSubforum := &dbmodels.Subforum{
			SubforumID:  1,
			Name:        "golang",
			DisplayName: "Go Programming",
			Description: sql.Null[string]{V: "Go programming language discussions", Valid: true},
		}

		// Mock pseudonyms
		mockPseudonym1 := &dbmodels.Pseudonym{
			PseudonymID: "author-pseudonym-1",
			DisplayName: "Author1",
		}
		mockPseudonym2 := &dbmodels.Pseudonym{
			PseudonymID: "author-pseudonym-2",
			DisplayName: "Author2",
		}

		// Mock database calls
		mockSubforumDAO.On("ListSubforums", mock.Anything).Return([]*dbmodels.Subforum{mockSubforum}, nil)
		mockPostDAO.On("GetPostsBySubforum", mock.Anything, int32(1), 1, 100, "created_at", true).Return(mockPosts, nil)
		mockPostDAO.On("GetPostsBySubforum", mock.Anything, int32(1), 1, 1000, "created_at", true).Return(mockPosts, nil)
		mockSubforumDAO.On("GetSubforumByID", mock.Anything, int32(1)).Return(mockSubforum, nil)
		mockPseudonymDAO.On("GetPseudonymByID", mock.Anything, "author-pseudonym-1").Return(mockPseudonym1, nil)
		mockPseudonymDAO.On("GetPseudonymByID", mock.Anything, "author-pseudonym-2").Return(mockPseudonym2, nil)

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

		// Verify first post
		assert.Equal(t, 1, response.Body.Posts[0].PostID)
		assert.Equal(t, "Understanding Golang Concurrency", response.Body.Posts[0].Title)
		assert.Equal(t, "This post discusses golang concurrency patterns...", response.Body.Posts[0].Content)
		assert.Equal(t, 1250, response.Body.Posts[0].Score)
		assert.Equal(t, 45, response.Body.Posts[0].CommentCount)
		assert.Equal(t, "Author1", response.Body.Posts[0].Author.DisplayName)
		assert.Equal(t, "Go Programming", response.Body.Posts[0].Subforum.DisplayName)

		mockPostDAO.AssertExpectations(t)
		mockSubforumDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("SearchPostsEmptyQuery", func(t *testing.T) {
		handler, _, _, _, _, _ := NewSearchHandlerWithMocks()

		// Create context with user
		ctx := createTestSearchContext(t, 1, "user-pseudonym-123", "TestUser")

		// Create input with empty query
		input := &models.SearchPostsInput{
			Query: "",
			Page:  1,
			Limit: 25,
		}

		// Call handler
		response, err := handler.SearchPosts(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "search query is required")
	})

	t.Run("SearchPostsAnonymousUser", func(t *testing.T) {
		handler, mockPostDAO, _, mockSubforumDAO, _, _ := NewSearchHandlerWithMocks()

		// Create context without user (anonymous)
		ctx := context.Background()

		// Mock subforums (required for anonymous search)
		mockSubforums := []*dbmodels.Subforum{}
		mockSubforumDAO.On("ListSubforums", mock.Anything).Return(mockSubforums, nil)

		// Create input
		input := &models.SearchPostsInput{
			Query: "test query",
			Page:  1,
			Limit: 25,
		}

		// Call handler
		response, err := handler.SearchPosts(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Len(t, response.Body.Posts, 0)

		mockPostDAO.AssertExpectations(t)
		mockSubforumDAO.AssertExpectations(t)
	})

	t.Run("SearchPostsDatabaseError", func(t *testing.T) {
		handler, _, _, mockSubforumDAO, _, _ := NewSearchHandlerWithMocks()

		// Create context with user
		ctx := createTestSearchContext(t, 1, "user-pseudonym-123", "TestUser")

		// Mock database error
		mockSubforumDAO.On("ListSubforums", mock.Anything).Return(nil, assert.AnError)

		// Create input
		input := &models.SearchPostsInput{
			Query: "test query",
			Page:  1,
			Limit: 25,
		}

		// Call handler
		response, err := handler.SearchPosts(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to search posts")

		mockSubforumDAO.AssertExpectations(t)
	})

	t.Run("SearchPostsWithFilters", func(t *testing.T) {
		handler, mockPostDAO, _, mockSubforumDAO, mockPseudonymDAO, _ := NewSearchHandlerWithMocks()

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		query := "golang"
		subforum := "golang"
		author := "test-author"
		sort := "relevance"
		timeFilter := "week"

		// Create context with user
		ctx := createTestSearchContext(t, userID, activePseudonymID, displayName)

		// Mock posts
		mockPosts := []*dbmodels.Post{
			{
				PostID:       1,
				Title:        "Golang Best Practices",
				Content:      sql.Null[string]{V: "Best practices for golang development...", Valid: true},
				PseudonymID:  "author-pseudonym-1",
				SubforumID:   1,
				Score:        sql.Null[int32]{V: 1000, Valid: true},
				CommentCount: sql.Null[int32]{V: 30, Valid: true},
				CreatedAt:    sql.Null[time.Time]{V: time.Now(), Valid: true},
			},
		}

		// Mock subforum
		mockSubforum := &dbmodels.Subforum{
			SubforumID:  1,
			Name:        "golang",
			DisplayName: "Go Programming",
			Description: sql.Null[string]{V: "Go programming language discussions", Valid: true},
		}

		// Mock pseudonym
		mockPseudonym := &dbmodels.Pseudonym{
			PseudonymID: "author-pseudonym-1",
			DisplayName: "test-author",
		}

		// Mock database calls
		mockSubforumDAO.On("ListSubforums", mock.Anything).Return([]*dbmodels.Subforum{mockSubforum}, nil)
		mockPostDAO.On("GetPostsBySubforum", mock.Anything, int32(1), 1, 100, "created_at", true).Return(mockPosts, nil)
		mockPostDAO.On("GetPostsBySubforum", mock.Anything, int32(1), 1, 1000, "created_at", true).Return(mockPosts, nil)
		mockSubforumDAO.On("GetSubforumByID", mock.Anything, int32(1)).Return(mockSubforum, nil)
		mockPseudonymDAO.On("GetPseudonymByID", mock.Anything, "author-pseudonym-1").Return(mockPseudonym, nil)

		// Create input with filters
		input := &models.SearchPostsInput{
			Query:    query,
			Subforum: subforum,
			Author:   author,
			Sort:     sort,
			Time:     timeFilter,
			Page:     1,
			Limit:    25,
		}

		// Call handler
		response, err := handler.SearchPosts(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, query, response.Body.Query)
		assert.Len(t, response.Body.Posts, 1)

		mockPostDAO.AssertExpectations(t)
		mockSubforumDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
	})
}

// TestSearchHandler_SearchUsers tests the search users functionality
func TestSearchHandler_SearchUsers(t *testing.T) {
	t.Run("SearchUsersSuccess", func(t *testing.T) {
		handler, mockPostDAO, mockUserDAO, _, mockPseudonymDAO, mockPermissionDAO := NewSearchHandlerWithMocks()

		// Test data
		userID := int64(1)
		activePseudonymID := "admin-pseudonym-123"
		displayName := "AdminUser"
		query := "john"

		// Create context with platform admin user
		ctx := createTestPlatformAdminContext(t, userID, activePseudonymID, displayName)

		// Mock permission check to return true for platform admin capability
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, userID, activePseudonymID, "system_admin", (*int32)(nil)).Return(true, nil)

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

		// Mock database calls
		mockUserDAO.On("ListUsers", mock.Anything, 1000, 0).Return(mockUsers, nil)
		mockPseudonymDAO.On("GetDefaultPseudonymByUserID", mock.Anything, int64(1), "user", "global").Return(mockPseudonym1, nil)
		mockPseudonymDAO.On("GetDefaultPseudonymByUserID", mock.Anything, int64(2), "user", "global").Return(mockPseudonym2, nil)

		// Mock karma score calculation calls
		mockPostDAO.On("GetPostsBySubforum", mock.Anything, int32(0), 1, 1000, "created_at", true).Return([]*dbmodels.Post{}, nil)

		// Create input
		input := &models.SearchUsersInput{
			Query: query,
			Page:  1,
			Limit: 25,
		}

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

		mockUserDAO.AssertExpectations(t)
		mockPseudonymDAO.AssertExpectations(t)
	})

	t.Run("SearchUsersEmptyQuery", func(t *testing.T) {
		handler, _, _, _, _, mockPermissionDAO := NewSearchHandlerWithMocks()

		// Create context with platform admin user
		ctx := createTestPlatformAdminContext(t, 1, "admin-pseudonym-123", "AdminUser")

		// Mock permission check to return true for platform admin capability
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "admin-pseudonym-123", "system_admin", (*int32)(nil)).Return(true, nil)

		// Create input with empty query
		input := &models.SearchUsersInput{
			Query: "",
			Page:  1,
			Limit: 25,
		}

		// Call handler
		response, err := handler.SearchUsers(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "search query is required")

		// Verify mocks
		mockPermissionDAO.AssertExpectations(t)
	})

	t.Run("SearchUsersAnonymousUser", func(t *testing.T) {
		handler, _, _, _, _, _ := NewSearchHandlerWithMocks()

		// Create context without user (anonymous)
		ctx := context.Background()

		// Create input
		input := &models.SearchUsersInput{
			Query: "test query",
			Page:  1,
			Limit: 25,
		}

		// Call handler
		response, err := handler.SearchUsers(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "authentication required")
	})

	t.Run("SearchUsersDatabaseError", func(t *testing.T) {
		handler, _, mockUserDAO, _, _, mockPermissionDAO := NewSearchHandlerWithMocks()

		// Create context with platform admin user
		ctx := createTestPlatformAdminContext(t, 1, "admin-pseudonym-123", "AdminUser")

		// Mock permission check to return true for platform admin capability
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, int64(1), "admin-pseudonym-123", "system_admin", (*int32)(nil)).Return(true, nil)

		// Mock database error
		mockUserDAO.On("ListUsers", mock.Anything, 1000, 0).Return(nil, assert.AnError)

		// Create input
		input := &models.SearchUsersInput{
			Query: "test query",
			Page:  1,
			Limit: 25,
		}

		// Call handler
		response, err := handler.SearchUsers(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to search users")

		mockUserDAO.AssertExpectations(t)
		mockPermissionDAO.AssertExpectations(t)
	})
}

// TestSearchHandler_NewSearchHandler tests the constructor
func TestSearchHandler_NewSearchHandler(t *testing.T) {
	t.Run("NewSearchHandlerWithMocks", func(t *testing.T) {
		// Test constructor with mocked dependencies
		mockPostDAO := mocks.NewMockPostDAO()
		mockUserDAO := &mocks.MockUserDAO{}
		mockSubforumDAO := mocks.NewMockSubforumDAO()
		mockPseudonymDAO := mocks.NewMockPseudonymDAO()

		// Create handler with mocked dependencies
		handler := handlers.NewSearchHandler(
			nil, // Mock DB
			mockPostDAO,
			mockUserDAO,
			mockSubforumDAO,
			mockPseudonymDAO,
			&mocks.MockPermissionDAO{},
		)

		// Verify handler was created successfully
		assert.NotNil(t, handler)
		// Note: We can't access private fields directly, but we can verify the handler was created
		// The actual field assignments are tested through the handler's behavior in other tests
	})
}
