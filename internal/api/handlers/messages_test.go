package handlers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// NewMessagesHandlerWithMocks creates a MessagesHandler with mocked dependencies
func NewMessagesHandlerWithMocks() (*handlers.MessagesHandler, *mocks.MockDirectMessageDAO, *mocks.MockUserDAO) {
	mockDirectMessageDAO := &mocks.MockDirectMessageDAO{}
	mockUserDAO := &mocks.MockUserDAO{}

	// Create handler using the test constructor
	handler := handlers.NewMessagesHandler(nil, mockDirectMessageDAO, mockUserDAO)

	return handler, mockDirectMessageDAO, mockUserDAO
}

// setupTestAuthMiddleware sets up the global auth middleware for testing
func setupTestAuthMiddleware() {
	authMiddleware := middleware.NewAuthMiddleware("test-secret", nil, &config.JWTConfig{
		Secret:      "test-secret",
		Expiration:  time.Hour,
		Development: true,
	}, &config.SecurityConfig{
		EnableMFA: false,
	})
	middleware.SetGlobalAuthMiddleware(authMiddleware)
}

// createAuthenticatedInput creates an input with a valid JWT token for testing
func createAuthenticatedInput(userID int64, activePseudonymID string, displayName string) *models.DirectMessageInput {
	// Create a user context
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false, // roles and capabilities deprecated
	}

	// Generate a JWT token
	token, _ := middleware.GenerateJWT(user, "test-secret", time.Hour)

	return &models.DirectMessageInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
	}
}

// createAuthenticatedListInput creates an input with a valid JWT token for testing
func createAuthenticatedListInput(userID int64, activePseudonymID string, displayName string) *models.DirectMessageListInput {
	// Create a user context
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false, // roles and capabilities deprecated
	}

	// Generate a JWT token
	token, _ := middleware.GenerateJWT(user, "test-secret", time.Hour)

	return &models.DirectMessageListInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
	}
}

// TestMessagesHandler_SendDirectMessage tests the send direct message functionality
func TestMessagesHandler_SendDirectMessage(t *testing.T) {
	// Set up global auth middleware for testing
	setupTestAuthMiddleware()

	t.Run("SendDirectMessageSuccess", func(t *testing.T) {
		handler, mockDirectMessageDAO, _ := NewMessagesHandlerWithMocks()

		// Test data
		userID := int64(1)
		activePseudonymID := "sender-pseudonym-123"
		recipientPseudonymID := "recipient-pseudonym-456"
		content := "Hello! This is a test message."
		displayName := "TestSender"

		// Create context
		ctx := context.Background()

		// Mock direct message creation
		expectedMessage := &dbmodels.DirectMessage{
			MessageID:            123,
			SenderPseudonymID:    activePseudonymID,
			RecipientPseudonymID: recipientPseudonymID,
			Content:              content,
			CreatedAt:            sql.Null[time.Time]{V: time.Now(), Valid: true},
		}
		mockDirectMessageDAO.On("CreateDirectMessage", mock.Anything, activePseudonymID, recipientPseudonymID, content).Return(expectedMessage, nil)

		// Create authenticated input
		input := createAuthenticatedInput(userID, activePseudonymID, displayName)
		input.Body = models.DirectMessageInputBody{
			RecipientPseudonymID: recipientPseudonymID,
			Content:              content,
		}

		// Call handler
		response, err := handler.SendDirectMessage(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, int(expectedMessage.MessageID), response.Body.MessageID)
		assert.Equal(t, recipientPseudonymID, response.Body.RecipientPseudonymID)
		assert.Equal(t, content, response.Body.Content)
		assert.NotEmpty(t, response.Body.CreatedAt)

		mockDirectMessageDAO.AssertExpectations(t)
	})

	t.Run("SendDirectMessageNoAuthentication", func(t *testing.T) {
		handler, _, _ := NewMessagesHandlerWithMocks()

		// Create context
		ctx := context.Background()

		// Create input without authentication
		input := &models.DirectMessageInput{
			Body: models.DirectMessageInputBody{
				RecipientPseudonymID: "recipient-pseudonym-456",
				Content:              "Test message",
			},
		}

		// Call handler
		response, err := handler.SendDirectMessage(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "authentication required")
	})

	t.Run("SendDirectMessageDatabaseError", func(t *testing.T) {
		handler, mockDirectMessageDAO, _ := NewMessagesHandlerWithMocks()

		// Test data
		userID := int64(1)
		activePseudonymID := "sender-pseudonym-123"
		recipientPseudonymID := "recipient-pseudonym-456"
		content := "Hello! This is a test message."
		displayName := "TestSender"

		// Create context
		ctx := context.Background()

		// Mock database error
		mockDirectMessageDAO.On("CreateDirectMessage", mock.Anything, activePseudonymID, recipientPseudonymID, content).Return(nil, assert.AnError)

		// Create authenticated input
		input := createAuthenticatedInput(userID, activePseudonymID, displayName)
		input.Body = models.DirectMessageInputBody{
			RecipientPseudonymID: recipientPseudonymID,
			Content:              content,
		}

		// Call handler
		response, err := handler.SendDirectMessage(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to send message")

		mockDirectMessageDAO.AssertExpectations(t)
	})
}

// TestMessagesHandler_GetDirectMessages tests the get direct messages functionality
func TestMessagesHandler_GetDirectMessages(t *testing.T) {
	// Set up global auth middleware for testing
	setupTestAuthMiddleware()

	t.Run("GetDirectMessagesSuccess", func(t *testing.T) {
		handler, mockDirectMessageDAO, _ := NewMessagesHandlerWithMocks()

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		page := 1
		limit := 25

		// Create context
		ctx := context.Background()

		// Mock messages
		now := time.Now()
		mockMessages := []*dbmodels.DirectMessage{
			{
				MessageID:            1,
				SenderPseudonymID:    "sender-1",
				RecipientPseudonymID: activePseudonymID,
				Content:              "Message 1",
				IsRead:               sql.Null[bool]{V: false, Valid: true},
				CreatedAt:            sql.Null[time.Time]{V: now, Valid: true},
			},
			{
				MessageID:            2,
				SenderPseudonymID:    "sender-2",
				RecipientPseudonymID: activePseudonymID,
				Content:              "Message 2",
				IsRead:               sql.Null[bool]{V: true, Valid: true},
				CreatedAt:            sql.Null[time.Time]{V: now.Add(-time.Hour), Valid: true},
			},
		}

		// Mock database calls
		mockDirectMessageDAO.On("GetDirectMessagesByPseudonym", mock.Anything, activePseudonymID, page, limit).Return(mockMessages, nil)
		mockDirectMessageDAO.On("CountDirectMessagesByPseudonym", mock.Anything, activePseudonymID).Return(int64(2), nil)

		// Create authenticated input
		input := createAuthenticatedListInput(userID, activePseudonymID, displayName)
		input.Page = page
		input.Limit = limit

		// Call handler
		response, err := handler.GetDirectMessages(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Len(t, response.Body.Messages, 2)
		assert.Equal(t, page, response.Body.Pagination.Page)
		assert.Equal(t, limit, response.Body.Pagination.Limit)
		assert.Equal(t, 2, response.Body.Pagination.Total)
		assert.Equal(t, 1, response.Body.Pagination.Pages)

		// Verify first message
		assert.Equal(t, 1, response.Body.Messages[0].MessageID)
		assert.Equal(t, "sender-1", response.Body.Messages[0].SenderPseudonymID)
		assert.Equal(t, activePseudonymID, response.Body.Messages[0].RecipientPseudonymID)
		assert.Equal(t, "Message 1", response.Body.Messages[0].Content)
		assert.False(t, response.Body.Messages[0].IsRead)
		assert.NotEmpty(t, response.Body.Messages[0].CreatedAt)

		// Verify second message
		assert.Equal(t, 2, response.Body.Messages[1].MessageID)
		assert.Equal(t, "sender-2", response.Body.Messages[1].SenderPseudonymID)
		assert.Equal(t, activePseudonymID, response.Body.Messages[1].RecipientPseudonymID)
		assert.Equal(t, "Message 2", response.Body.Messages[1].Content)
		assert.True(t, response.Body.Messages[1].IsRead)
		assert.NotEmpty(t, response.Body.Messages[1].CreatedAt)

		mockDirectMessageDAO.AssertExpectations(t)
	})

	t.Run("GetDirectMessagesNoAuthentication", func(t *testing.T) {
		handler, _, _ := NewMessagesHandlerWithMocks()

		// Create context
		ctx := context.Background()

		// Create input without authentication
		input := &models.DirectMessageListInput{
			Page:  1,
			Limit: 25,
		}

		// Call handler
		response, err := handler.GetDirectMessages(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "authentication required")
	})

	t.Run("GetDirectMessagesDatabaseError", func(t *testing.T) {
		handler, mockDirectMessageDAO, _ := NewMessagesHandlerWithMocks()

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		page := 1
		limit := 25

		// Create context
		ctx := context.Background()

		// Mock database error
		mockDirectMessageDAO.On("GetDirectMessagesByPseudonym", mock.Anything, activePseudonymID, page, limit).Return(nil, assert.AnError)

		// Create authenticated input
		input := createAuthenticatedListInput(userID, activePseudonymID, displayName)
		input.Page = page
		input.Limit = limit

		// Call handler
		response, err := handler.GetDirectMessages(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to get messages")

		mockDirectMessageDAO.AssertExpectations(t)
	})

	t.Run("GetDirectMessagesCountError", func(t *testing.T) {
		handler, mockDirectMessageDAO, _ := NewMessagesHandlerWithMocks()

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		page := 1
		limit := 25

		// Create context
		ctx := context.Background()

		// Mock successful messages retrieval but failed count
		mockMessages := []*dbmodels.DirectMessage{}
		mockDirectMessageDAO.On("GetDirectMessagesByPseudonym", mock.Anything, activePseudonymID, page, limit).Return(mockMessages, nil)
		mockDirectMessageDAO.On("CountDirectMessagesByPseudonym", mock.Anything, activePseudonymID).Return(int64(0), assert.AnError)

		// Create authenticated input
		input := createAuthenticatedListInput(userID, activePseudonymID, displayName)
		input.Page = page
		input.Limit = limit

		// Call handler
		response, err := handler.GetDirectMessages(ctx, input)

		// Assertions
		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to count messages")

		mockDirectMessageDAO.AssertExpectations(t)
	})

	t.Run("GetDirectMessagesWithNullFields", func(t *testing.T) {
		handler, mockDirectMessageDAO, _ := NewMessagesHandlerWithMocks()

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		page := 1
		limit := 25

		// Create context
		ctx := context.Background()

		// Mock messages with null fields
		mockMessages := []*dbmodels.DirectMessage{
			{
				MessageID:            1,
				SenderPseudonymID:    "sender-1",
				RecipientPseudonymID: activePseudonymID,
				Content:              "Message with null fields",
				IsRead:               sql.Null[bool]{},      // Null
				CreatedAt:            sql.Null[time.Time]{}, // Null
			},
		}

		// Mock database calls
		mockDirectMessageDAO.On("GetDirectMessagesByPseudonym", mock.Anything, activePseudonymID, page, limit).Return(mockMessages, nil)
		mockDirectMessageDAO.On("CountDirectMessagesByPseudonym", mock.Anything, activePseudonymID).Return(int64(1), nil)

		// Create authenticated input
		input := createAuthenticatedListInput(userID, activePseudonymID, displayName)
		input.Page = page
		input.Limit = limit

		// Call handler
		response, err := handler.GetDirectMessages(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Len(t, response.Body.Messages, 1)
		assert.False(t, response.Body.Messages[0].IsRead)    // Should default to false for null
		assert.Empty(t, response.Body.Messages[0].CreatedAt) // Should be empty for null

		mockDirectMessageDAO.AssertExpectations(t)
	})
}

// TestMessagesHandler_NewMessagesHandler tests the main constructor function
func TestMessagesHandler_NewMessagesHandler(t *testing.T) {
	t.Run("NewMessagesHandlerSuccess", func(t *testing.T) {
		// Create mock dependencies
		mockDirectMessageDAO := &mocks.MockDirectMessageDAO{}
		mockUserDAO := &mocks.MockUserDAO{}

		// Create handler with dependencies
		handler := handlers.NewMessagesHandler(
			nil, // nil db for testing
			mockDirectMessageDAO,
			mockUserDAO,
		)

		// Verify handler is created
		assert.NotNil(t, handler)
	})
}
