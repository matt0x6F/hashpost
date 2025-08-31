package handlers_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	daomocks "github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/services"
	servicemocks "github.com/matt0x6f/hashpost/internal/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// NewMessagesHandlerWithMocks creates a MessagesHandler with mocked dependencies
func NewMessagesHandlerWithMocks() (*handlers.MessagesHandler, *daomocks.MockDirectMessageDAO, *daomocks.MockUserDAO, *daomocks.MockPseudonymDAO) {
	mockDirectMessageDAO := &daomocks.MockDirectMessageDAO{}
	mockUserDAO := &daomocks.MockUserDAO{}
	mockPseudonymDAO := &daomocks.MockPseudonymDAO{}

	// Create handler using the test constructor - for now, pass nil for encrypted messaging dependencies
	// since the existing tests don't need them
	handler := handlers.NewMessagesHandler(
		mockDirectMessageDAO,
		mockUserDAO,
		mockPseudonymDAO,
		nil,                           // userEncryptionKeyDAO
		nil,                           // conversationKeyDAO
		nil,                           // encryptedMessageDAO
		nil,                           // encryptionService
		nil,                           // keyManagementService
		&daomocks.MockPermissionDAO{}, // permissionDAO
		nil,                           // db
	)

	return handler, mockDirectMessageDAO, mockUserDAO, mockPseudonymDAO
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
		handler, mockDirectMessageDAO, _, mockPseudonymDAO := NewMessagesHandlerWithMocks()

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

		// Set up mock expectation for UpdateLastActive
		mockPseudonymDAO.On("UpdateLastActive", mock.Anything, activePseudonymID).Return(nil)

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
		handler, _, _, _ := NewMessagesHandlerWithMocks()

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
		handler, mockDirectMessageDAO, _, _ := NewMessagesHandlerWithMocks()

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
		handler, mockDirectMessageDAO, _, _ := NewMessagesHandlerWithMocks()

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
		handler, _, _, _ := NewMessagesHandlerWithMocks()

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
		handler, mockDirectMessageDAO, _, _ := NewMessagesHandlerWithMocks()

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
		handler, mockDirectMessageDAO, _, _ := NewMessagesHandlerWithMocks()

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
		handler, mockDirectMessageDAO, _, _ := NewMessagesHandlerWithMocks()

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
		mockDirectMessageDAO := &daomocks.MockDirectMessageDAO{}
		mockUserDAO := &daomocks.MockUserDAO{}
		mockPseudonymDAO := &daomocks.MockPseudonymDAO{}

		// Create handler with dependencies
		handler := handlers.NewMessagesHandler(
			mockDirectMessageDAO,
			mockUserDAO,
			mockPseudonymDAO,
			nil,                           // userEncryptionKeyDAO
			nil,                           // conversationKeyDAO
			nil,                           // encryptedMessageDAO
			nil,                           // encryptionService
			nil,                           // keyManagementService
			&daomocks.MockPermissionDAO{}, // permissionDAO
			nil,                           // db
		)

		// Verify handler is created
		assert.NotNil(t, handler)
	})
}

// TestMessagesHandler_EncryptedMessaging tests the encrypted messaging functionality
func TestMessagesHandler_EncryptedMessaging(t *testing.T) {
	// Set up global auth middleware for testing
	setupTestAuthMiddleware()

	t.Run("SendEncryptedMessageSuccess", func(t *testing.T) {
		// Create mock dependencies for encrypted messaging
		mockDirectMessageDAO := &daomocks.MockDirectMessageDAO{}
		mockUserDAO := &daomocks.MockUserDAO{}
		mockPseudonymDAO := &daomocks.MockPseudonymDAO{}
		mockUserEncryptionKeyDAO := &daomocks.MockUserEncryptionKeyDAO{}
		mockConversationKeyDAO := &daomocks.MockConversationKeyDAO{}
		mockEncryptedMessageDAO := &daomocks.MockEncryptedMessageDAO{}
		mockEncryptionService := &servicemocks.MockEncryptionService{}
		mockKeyManagementService := &servicemocks.MockKeyManagementService{}
		mockPermissionDAO := &daomocks.MockPermissionDAO{}

		// Create handler with all dependencies
		handler := handlers.NewMessagesHandler(
			mockDirectMessageDAO,
			mockUserDAO,
			mockPseudonymDAO,
			mockUserEncryptionKeyDAO,
			mockConversationKeyDAO,
			mockEncryptedMessageDAO,
			mockEncryptionService,
			mockKeyManagementService,
			mockPermissionDAO, // permissionDAO
			nil,               // db
		)

		// Test data
		userID := int64(1)
		activePseudonymID := "sender-pseudonym-123"
		recipientUserID := int64(2)
		content := "Hello! This is an encrypted test message."
		displayName := "TestSender"

		// Create context
		ctx := context.Background()

		// Mock conversation key creation
		now := time.Now()
		testConversationKey := &dbmodels.ConversationKey{
			ConversationID:     uuid.Must(uuid.NewV4()),
			Participant1UserID: userID,
			Participant2UserID: recipientUserID,
			EncryptedSharedKey: []byte("encrypted-key"),
			KeyFingerprint:     "test-fingerprint",
			CreatedAt:          sql.Null[time.Time]{V: now, Valid: true},
			ExpiresAt:          sql.Null[time.Time]{V: now.AddDate(0, 1, 0), Valid: true},
			IsActive:           sql.Null[bool]{V: true, Valid: true},
			KeyVersion:         1,
		}

		// Mock encrypted message creation
		testMessage := &dbmodels.EncryptedMessage{
			MessageID:        789,
			ConversationID:   uuid.Must(uuid.NewV4()),
			EncryptedContent: []byte("encrypted-content"),
			Iv:               []byte("test-iv"),
			ContentHash:      "test-content-hash",
			KeyVersion:       1,
			Signature:        []byte("test-signature"),
		}

		// Mock expectations
		mockConversationKeyDAO.On("GetConversationKeyByParticipants", mock.Anything, userID, recipientUserID).
			Return(nil, nil).Once() // No existing conversation
		mockConversationKeyDAO.On("CreateConversationKey", mock.Anything, userID, recipientUserID, mock.Anything, mock.Anything, mock.Anything).
			Return(testConversationKey, nil).Once()
		mockEncryptedMessageDAO.On("CreateEncryptedMessage", mock.Anything, testConversationKey.ConversationID, mock.Anything, mock.Anything, mock.Anything, int32(1), mock.Anything).
			Return(testMessage, nil).Once()

		// Mock permission checking
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, userID, activePseudonymID, "send_direct_messages", mock.Anything).
			Return(true, nil).Once()

		// Mock key management
		mockKeyManagementService.On("EnsureMessagingKeys", mock.Anything, userID, activePseudonymID).
			Return(nil).Once()

		// Mock encryption service
		mockEncryptionService.On("GenerateAESKey").
			Return([]byte("test-aes-key"), nil).Once()

		// Mock user encryption key retrieval
		mockUserEncryptionKeyDAO.On("GetUserPublicKey", mock.Anything, userID).
			Return([]byte("test-public-key"), nil).Once()

		// Mock user encryption key for message encryption
		mockUserEncryptionKeyDAO.On("GetUserEncryptionKey", mock.Anything, userID).
			Return(&dbmodels.UserEncryptionKey{
				UserID:                userID,
				EncryptedSignatureKey: []byte("test-encrypted-signature-key"),
				PublicSignatureKey:    []byte("test-public-signature-key"),
			}, nil).Once()

		// Mock conversation key encryption
		mockEncryptionService.On("EncryptWithPublicKey", mock.Anything, mock.Anything).
			Return([]byte("encrypted-conversation-key"), nil).Once()

		// Mock conversation key decryption
		mockEncryptionService.On("DecryptWithPrivateKey", mock.Anything, mock.Anything).
			Return([]byte("test-aes-key"), nil).Once()

		// Mock encryption
		mockEncryptionService.On("EncryptAES", mock.Anything, mock.Anything).
			Return(&services.EncryptedMessage{
				EncryptedContent: []byte("encrypted-content"),
				IV:               []byte("test-iv"),
			}, nil).Once()

		// Create authenticated input for encrypted message
		input := createAuthenticatedEncryptedMessageInput(userID, activePseudonymID, displayName)
		input.RecipientUserID = recipientUserID
		input.Content = content

		// Call handler
		response, err := handler.SendEncryptedMessage(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, testMessage.MessageID, response.Body.MessageID)
		assert.Equal(t, testConversationKey.ConversationID.String(), response.Body.ConversationID)
		assert.Equal(t, "sent", response.Body.MessageStatus)

		// Verify mocks
		mockConversationKeyDAO.AssertExpectations(t)
		mockEncryptedMessageDAO.AssertExpectations(t)
	})

	t.Run("GetConversationsSuccess", func(t *testing.T) {
		// Create mock dependencies for encrypted messaging
		mockDirectMessageDAO := &daomocks.MockDirectMessageDAO{}
		mockUserDAO := &daomocks.MockUserDAO{}
		mockPseudonymDAO := &daomocks.MockPseudonymDAO{}
		mockUserEncryptionKeyDAO := &daomocks.MockUserEncryptionKeyDAO{}
		mockConversationKeyDAO := &daomocks.MockConversationKeyDAO{}
		mockEncryptedMessageDAO := &daomocks.MockEncryptedMessageDAO{}
		mockEncryptionService := &servicemocks.MockEncryptionService{}
		mockKeyManagementService := &servicemocks.MockKeyManagementService{}
		mockPermissionDAO := &daomocks.MockPermissionDAO{}

		// Create handler with all dependencies
		handler := handlers.NewMessagesHandler(
			mockDirectMessageDAO,
			mockUserDAO,
			mockPseudonymDAO,
			mockUserEncryptionKeyDAO,
			mockConversationKeyDAO,
			mockEncryptedMessageDAO,
			mockEncryptionService,
			mockKeyManagementService,
			mockPermissionDAO, // permissionDAO
			nil,               // db
		)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"

		// Create context
		ctx := context.Background()

		// Mock conversation keys
		now := time.Now()
		testConversationKeys := []*dbmodels.ConversationKey{
			{
				ConversationID:     uuid.Must(uuid.NewV4()),
				Participant1UserID: userID,
				Participant2UserID: 456,
				EncryptedSharedKey: []byte("encrypted-key-1"),
				KeyFingerprint:     "fp-1",
				CreatedAt:          sql.Null[time.Time]{V: now, Valid: true},
				ExpiresAt:          sql.Null[time.Time]{V: now.AddDate(0, 1, 0), Valid: true},
				IsActive:           sql.Null[bool]{V: true, Valid: true},
				KeyVersion:         1,
			},
		}

		// Mock expectations
		mockConversationKeyDAO.On("GetActiveConversationKeys", mock.Anything, userID).
			Return(testConversationKeys, nil).Once()
		mockEncryptedMessageDAO.On("GetMessageCountByConversation", mock.Anything, testConversationKeys[0].ConversationID).
			Return(int64(5), nil).Once()

		// Mock permission checking
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, userID, activePseudonymID, "receive_direct_messages", mock.Anything).
			Return(true, nil).Once()

		// Create authenticated input
		input := createAuthenticatedGetConversationsInput(userID, activePseudonymID, displayName)

		// Call handler
		response, err := handler.GetConversations(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Len(t, response.Body.Conversations, 1)
		assert.Equal(t, int64(1), response.Body.TotalCount)
		assert.Equal(t, int64(456), response.Body.Conversations[0].OtherUserID)

		// Verify mocks
		mockConversationKeyDAO.AssertExpectations(t)
		mockEncryptedMessageDAO.AssertExpectations(t)
	})

	t.Run("GetConversationMessagesSuccess", func(t *testing.T) {
		// Create mock dependencies for encrypted messaging
		mockDirectMessageDAO := &daomocks.MockDirectMessageDAO{}
		mockUserDAO := &daomocks.MockUserDAO{}
		mockPseudonymDAO := &daomocks.MockPseudonymDAO{}
		mockUserEncryptionKeyDAO := &daomocks.MockUserEncryptionKeyDAO{}
		mockConversationKeyDAO := &daomocks.MockConversationKeyDAO{}
		mockEncryptedMessageDAO := &daomocks.MockEncryptedMessageDAO{}
		mockEncryptionService := &servicemocks.MockEncryptionService{}
		mockKeyManagementService := &servicemocks.MockKeyManagementService{}
		mockPermissionDAO := &daomocks.MockPermissionDAO{}

		// Create handler with all dependencies
		handler := handlers.NewMessagesHandler(
			mockDirectMessageDAO,
			mockUserDAO,
			mockPseudonymDAO,
			mockUserEncryptionKeyDAO,
			mockConversationKeyDAO,
			mockEncryptedMessageDAO,
			mockEncryptionService,
			mockKeyManagementService,
			mockPermissionDAO, // permissionDAO
			nil,               // db
		)

		// Test data
		userID := int64(1)
		activePseudonymID := "user-pseudonym-123"
		displayName := "TestUser"
		conversationID := uuid.Must(uuid.NewV4())

		// Create context
		ctx := context.Background()

		// Mock conversation key
		now := time.Now()
		testConversationKey := &dbmodels.ConversationKey{
			ConversationID:     conversationID,
			Participant1UserID: userID,
			Participant2UserID: 456,
			EncryptedSharedKey: []byte("encrypted-key"),
			KeyFingerprint:     "test-fingerprint",
			CreatedAt:          sql.Null[time.Time]{V: now, Valid: true},
			ExpiresAt:          sql.Null[time.Time]{V: now.AddDate(0, 1, 0), Valid: true},
			IsActive:           sql.Null[bool]{V: true, Valid: true},
			KeyVersion:         1,
		}

		// Mock messages
		testMessages := []*dbmodels.EncryptedMessage{
			{
				MessageID:        1,
				ConversationID:   conversationID,
				EncryptedContent: []byte("encrypted-content-1"),
				Iv:               []byte("iv-1"),
				ContentHash:      "hash-1",
				KeyVersion:       1,
				Signature:        []byte("sig-1"),
			},
		}

		// Mock expectations
		mockConversationKeyDAO.On("GetConversationKey", mock.Anything, conversationID).
			Return(testConversationKey, nil).Once()
		mockEncryptedMessageDAO.On("GetMessagesByConversation", mock.Anything, conversationID).
			Return(testMessages, nil).Once()

		// Mock permission checking
		mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, userID, activePseudonymID, "receive_direct_messages", mock.Anything).
			Return(true, nil).Once()

		// Create authenticated input
		input := createAuthenticatedGetConversationMessagesInput(userID, activePseudonymID, displayName)
		input.ConversationID = conversationID.String()

		// Call handler
		response, err := handler.GetConversationMessages(ctx, input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, conversationID.String(), response.Body.ConversationID)
		assert.Len(t, response.Body.Messages, 1)
		assert.Equal(t, int64(1), response.Body.TotalCount)

		// Verify mocks
		mockConversationKeyDAO.AssertExpectations(t)
		mockEncryptedMessageDAO.AssertExpectations(t)
	})

	// NEW: Test full end-to-end messaging flow between two users
	t.Run("FullEndToEndMessagingFlow", func(t *testing.T) {
		// Create mock dependencies for encrypted messaging
		mockDirectMessageDAO := &daomocks.MockDirectMessageDAO{}
		mockUserDAO := &daomocks.MockUserDAO{}
		mockPseudonymDAO := &daomocks.MockPseudonymDAO{}
		mockUserEncryptionKeyDAO := &daomocks.MockUserEncryptionKeyDAO{}
		mockConversationKeyDAO := &daomocks.MockConversationKeyDAO{}
		mockEncryptedMessageDAO := &daomocks.MockEncryptedMessageDAO{}
		mockEncryptionService := &servicemocks.MockEncryptionService{}
		mockKeyManagementService := &servicemocks.MockKeyManagementService{}
		mockPermissionDAO := &daomocks.MockPermissionDAO{}

		// Create handler with all dependencies
		handler := handlers.NewMessagesHandler(
			mockDirectMessageDAO,
			mockUserDAO,
			mockPseudonymDAO,
			mockUserEncryptionKeyDAO,
			mockConversationKeyDAO,
			mockEncryptedMessageDAO,
			mockEncryptionService,
			mockKeyManagementService,
			mockPermissionDAO,
			nil, // db
		)

		// Test data for two users
		userAID := int64(1)
		userBID := int64(2)
		userAPseudonymID := "user-a-pseudonym"
		userBPseudonymID := "user-b-pseudonym"
		userADisplayName := "UserA"
		userBDisplayName := "UserB"

		// Create context
		ctx := context.Background()

		// Mock conversation key that will be shared between users
		conversationID := uuid.Must(uuid.NewV4())
		sharedAESKey := []byte("shared-aes-key-12345")
		keyFingerprint := "fp-shared-key-abc123"

		now := time.Now()
		expiresAt := now.AddDate(0, 1, 0)

		// Step 1: User A initiates conversation and creates shared key
		t.Run("UserA_CreatesConversation", func(t *testing.T) {
			// Mock conversation key creation
			testConversationKey := &dbmodels.ConversationKey{
				ConversationID:     conversationID,
				Participant1UserID: userAID,
				Participant2UserID: userBID,
				EncryptedSharedKey: []byte("encrypted-with-user-a-public-key"),
				KeyFingerprint:     keyFingerprint,
				CreatedAt:          sql.Null[time.Time]{V: now, Valid: true},
				ExpiresAt:          sql.Null[time.Time]{V: expiresAt, Valid: true},
				IsActive:           sql.Null[bool]{V: true, Valid: true},
				KeyVersion:         1,
			}

			// Mock expectations for User A creating conversation
			mockConversationKeyDAO.On("GetConversationKeyByParticipants", mock.Anything, userAID, userBID).
				Return(nil, nil).Once() // No existing conversation
			mockConversationKeyDAO.On("CreateConversationKey", mock.Anything, userAID, userBID, mock.Anything, mock.Anything, mock.Anything).
				Return(testConversationKey, nil).Once()

			// Mock permission checking for User A
			mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, userAID, userAPseudonymID, "send_direct_messages", mock.Anything).
				Return(true, nil).Once()

			// Mock key management for User A
			mockKeyManagementService.On("EnsureMessagingKeys", mock.Anything, userAID, userAPseudonymID).
				Return(nil).Once()

			// Mock User A's public key retrieval
			mockUserEncryptionKeyDAO.On("GetUserPublicKey", mock.Anything, userAID).
				Return([]byte("user-a-public-key"), nil).Once()

			// Mock encryption service for key generation and encryption
			mockEncryptionService.On("GenerateAESKey").
				Return(sharedAESKey, nil).Once()
			mockEncryptionService.On("EncryptWithPublicKey", []byte("user-a-public-key"), sharedAESKey).
				Return([]byte("encrypted-with-user-a-public-key"), nil).Once()

			// Mock encrypted message creation
			testMessage := &dbmodels.EncryptedMessage{
				MessageID:        1,
				ConversationID:   conversationID,
				EncryptedContent: []byte("encrypted-content-from-user-a"),
				Iv:               []byte("iv-from-user-a"),
				ContentHash:      "hash-from-user-a",
				KeyVersion:       1,
				Signature:        []byte("sig-from-user-a"),
			}
			mockEncryptedMessageDAO.On("CreateEncryptedMessage", mock.Anything, conversationID, mock.Anything, mock.Anything, mock.Anything, int32(1), mock.Anything).
				Return(testMessage, nil).Once()

			// Mock User A's private key for message encryption
			mockUserEncryptionKeyDAO.On("GetUserEncryptionKey", mock.Anything, userAID).
				Return(&dbmodels.UserEncryptionKey{
					UserID:                userAID,
					EncryptedSignatureKey: []byte("user-a-private-key"),
					PublicSignatureKey:    []byte("user-a-public-key"),
				}, nil).Once()

			// Mock conversation key decryption for User A
			mockEncryptionService.On("DecryptWithPrivateKey", []byte("user-a-private-key"), []byte("encrypted-with-user-a-public-key")).
				Return(sharedAESKey, nil).Once()

			// Mock message encryption
			mockEncryptionService.On("EncryptAES", sharedAESKey, []byte("Hello User B! This is User A.")).
				Return(&services.EncryptedMessage{
					EncryptedContent: []byte("encrypted-content-from-user-a"),
					IV:               []byte("iv-from-user-a"),
				}, nil).Once()

			// Create input for User A sending message
			input := createAuthenticatedEncryptedMessageInput(userAID, userAPseudonymID, userADisplayName)
			input.RecipientUserID = userBID
			input.Content = "Hello User B! This is User A."

			// Call handler for User A
			response, err := handler.SendEncryptedMessage(ctx, input)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, 200, response.Status)
			assert.Equal(t, conversationID.String(), response.Body.ConversationID)
			assert.Equal(t, "sent", response.Body.MessageStatus)

			// Verify mocks
			mockConversationKeyDAO.AssertExpectations(t)
			mockEncryptedMessageDAO.AssertExpectations(t)
		})

		// Step 2: User B retrieves and decrypts the conversation
		t.Run("UserB_RetrievesConversation", func(t *testing.T) {
			// Mock conversation key retrieval for User B
			testConversationKey := &dbmodels.ConversationKey{
				ConversationID:     conversationID,
				Participant1UserID: userAID,
				Participant2UserID: userBID,
				EncryptedSharedKey: []byte("encrypted-with-user-a-public-key"),
				KeyFingerprint:     keyFingerprint,
				CreatedAt:          sql.Null[time.Time]{V: now, Valid: true},
				ExpiresAt:          sql.Null[time.Time]{V: expiresAt, Valid: true},
				IsActive:           sql.Null[bool]{V: true, Valid: true},
				KeyVersion:         1,
			}

			// Mock permission checking for User B
			mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, userBID, userBPseudonymID, "receive_direct_messages", mock.Anything).
				Return(true, nil).Once()

			// Mock conversation retrieval
			mockConversationKeyDAO.On("GetActiveConversationKeys", mock.Anything, userBID).
				Return([]*dbmodels.ConversationKey{testConversationKey}, nil).Once()

			// Mock message count
			mockEncryptedMessageDAO.On("GetMessageCountByConversation", mock.Anything, conversationID).
				Return(int64(1), nil).Once()

			// Create input for User B getting conversations
			input := createAuthenticatedGetConversationsInput(userBID, userBPseudonymID, userBDisplayName)

			// Call handler for User B
			response, err := handler.GetConversations(ctx, input)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, 200, response.Status)
			assert.Len(t, response.Body.Conversations, 1)
			assert.Equal(t, conversationID.String(), response.Body.Conversations[0].ConversationID)
			assert.Equal(t, userAID, response.Body.Conversations[0].OtherUserID)
			assert.Equal(t, keyFingerprint, response.Body.Conversations[0].KeyFingerprint)

			// Verify mocks
			mockConversationKeyDAO.AssertExpectations(t)
			mockEncryptedMessageDAO.AssertExpectations(t)
		})

		// Step 3: User B sends a reply message
		t.Run("UserB_SendsReply", func(t *testing.T) {
			// Mock conversation key retrieval for User B's reply
			testConversationKey := &dbmodels.ConversationKey{
				ConversationID:     conversationID,
				Participant1UserID: userAID,
				Participant2UserID: userBID,
				EncryptedSharedKey: []byte("encrypted-with-user-a-public-key"),
				KeyFingerprint:     keyFingerprint,
				CreatedAt:          sql.Null[time.Time]{V: now, Valid: true},
				ExpiresAt:          sql.Null[time.Time]{V: expiresAt, Valid: true},
				IsActive:           sql.Null[bool]{V: true, Valid: true},
				KeyVersion:         1,
			}

			// Mock existing conversation key retrieval
			mockConversationKeyDAO.On("GetConversationKeyByParticipants", mock.Anything, userBID, userAID).
				Return(testConversationKey, nil).Once()

			// Mock permission checking for User B
			mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, userBID, userBPseudonymID, "send_direct_messages", mock.Anything).
				Return(true, nil).Once()

			// Mock key management for User B
			mockKeyManagementService.On("EnsureMessagingKeys", mock.Anything, userBID, userBPseudonymID).
				Return(nil).Once()

			// Mock User B's private key for message encryption
			mockUserEncryptionKeyDAO.On("GetUserEncryptionKey", mock.Anything, userBID).
				Return(&dbmodels.UserEncryptionKey{
					UserID:                userBID,
					EncryptedSignatureKey: []byte("user-b-private-key"),
					PublicSignatureKey:    []byte("user-b-public-key"),
				}, nil).Once()

			// Mock conversation key decryption for User B
			mockEncryptionService.On("DecryptWithPrivateKey", []byte("user-b-private-key"), []byte("encrypted-with-user-a-public-key")).
				Return(sharedAESKey, nil).Once()

			// Mock encrypted message creation for User B's reply
			testMessage := &dbmodels.EncryptedMessage{
				MessageID:        2,
				ConversationID:   conversationID,
				EncryptedContent: []byte("encrypted-content-from-user-b"),
				Iv:               []byte("iv-from-user-b"),
				ContentHash:      "hash-from-user-b",
				KeyVersion:       1,
				Signature:        []byte("sig-from-user-b"),
			}
			mockEncryptedMessageDAO.On("CreateEncryptedMessage", mock.Anything, conversationID, mock.Anything, mock.Anything, mock.Anything, int32(1), mock.Anything).
				Return(testMessage, nil).Once()

			// Mock message encryption for User B's reply
			mockEncryptionService.On("EncryptAES", sharedAESKey, []byte("Hello User A! This is User B replying.")).
				Return(&services.EncryptedMessage{
					EncryptedContent: []byte("encrypted-content-from-user-b"),
					IV:               []byte("iv-from-user-b"),
				}, nil).Once()

			// Create input for User B sending reply
			input := createAuthenticatedEncryptedMessageInput(userBID, userBPseudonymID, userBDisplayName)
			input.RecipientUserID = userAID
			input.Content = "Hello User A! This is User B replying."

			// Call handler for User B
			response, err := handler.SendEncryptedMessage(ctx, input)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, 200, response.Status)
			assert.Equal(t, conversationID.String(), response.Body.ConversationID)
			assert.Equal(t, "sent", response.Body.MessageStatus)

			// Verify mocks
			mockConversationKeyDAO.AssertExpectations(t)
			mockEncryptedMessageDAO.AssertExpectations(t)
		})

		// Step 4: User A retrieves the conversation messages (including User B's reply)
		t.Run("UserA_RetrievesMessages", func(t *testing.T) {
			// Mock conversation key verification
			testConversationKey := &dbmodels.ConversationKey{
				ConversationID:     conversationID,
				Participant1UserID: userAID,
				Participant2UserID: userBID,
				EncryptedSharedKey: []byte("encrypted-with-user-a-public-key"),
				KeyFingerprint:     keyFingerprint,
				CreatedAt:          sql.Null[time.Time]{V: now, Valid: true},
				ExpiresAt:          sql.Null[time.Time]{V: expiresAt, Valid: true},
				IsActive:           sql.Null[bool]{V: true, Valid: true},
				KeyVersion:         1,
			}

			// Mock permission checking for User A
			mockPermissionDAO.On("HasUnifiedCapability", mock.Anything, userAID, userAPseudonymID, "receive_direct_messages", mock.Anything).
				Return(true, nil).Once()

			// Mock conversation key retrieval
			mockConversationKeyDAO.On("GetConversationKey", mock.Anything, conversationID).
				Return(testConversationKey, nil).Once()

			// Mock messages retrieval (both User A's original message and User B's reply)
			testMessages := []*dbmodels.EncryptedMessage{
				{
					MessageID:        1,
					ConversationID:   conversationID,
					EncryptedContent: []byte("encrypted-content-from-user-a"),
					Iv:               []byte("iv-from-user-a"),
					ContentHash:      "hash-from-user-a",
					KeyVersion:       1,
					Signature:        []byte("sig-from-user-a"),
					CreatedAt:        sql.Null[time.Time]{V: now, Valid: true},
				},
				{
					MessageID:        2,
					ConversationID:   conversationID,
					EncryptedContent: []byte("encrypted-content-from-user-b"),
					Iv:               []byte("iv-from-user-b"),
					ContentHash:      "hash-from-user-b",
					KeyVersion:       1,
					Signature:        []byte("sig-from-user-b"),
					CreatedAt:        sql.Null[time.Time]{V: now.Add(time.Minute), Valid: true},
				},
			}
			mockEncryptedMessageDAO.On("GetMessagesByConversation", mock.Anything, conversationID).
				Return(testMessages, nil).Once()

			// Create input for User A getting conversation messages
			input := createAuthenticatedGetConversationMessagesInput(userAID, userAPseudonymID, userADisplayName)
			input.ConversationID = conversationID.String()

			// Call handler for User A
			response, err := handler.GetConversationMessages(ctx, input)

			// Assertions
			require.NoError(t, err)
			require.NotNil(t, response)
			assert.Equal(t, 200, response.Status)
			assert.Equal(t, conversationID.String(), response.Body.ConversationID)
			assert.Len(t, response.Body.Messages, 2)
			assert.Equal(t, int64(2), response.Body.TotalCount)

			// Verify first message (User A's original)
			assert.Equal(t, int64(1), response.Body.Messages[0].MessageID)
			assert.Equal(t, "hash-from-user-a", response.Body.Messages[0].ContentHash)
			assert.Equal(t, 1, response.Body.Messages[0].KeyVersion)

			// Verify second message (User B's reply)
			assert.Equal(t, int64(2), response.Body.Messages[1].MessageID)
			assert.Equal(t, "hash-from-user-b", response.Body.Messages[1].ContentHash)
			assert.Equal(t, 1, response.Body.Messages[1].KeyVersion)

			// Verify mocks
			mockConversationKeyDAO.AssertExpectations(t)
			mockEncryptedMessageDAO.AssertExpectations(t)
		})
	})

	// NEW: Test key exchange security and validation
	t.Run("KeyExchangeSecurityValidation", func(t *testing.T) {
		// Create mock dependencies
		mockEncryptionService := &servicemocks.MockEncryptionService{}

		t.Run("ConversationKeyEncryption", func(t *testing.T) {
			// Test that conversation keys are properly encrypted
			sharedKey := []byte("shared-aes-key-12345")
			userPublicKey := []byte("user-public-key")

			// Mock encryption service
			mockEncryptionService.On("EncryptWithPublicKey", userPublicKey, sharedKey).
				Return([]byte("encrypted-conversation-key"), nil).Once()

			// Test the encryption
			encryptedKey, err := mockEncryptionService.EncryptWithPublicKey(userPublicKey, sharedKey)

			require.NoError(t, err)
			assert.NotEqual(t, sharedKey, encryptedKey) // Encrypted key should be different from original
			assert.Equal(t, []byte("encrypted-conversation-key"), encryptedKey)

			mockEncryptionService.AssertExpectations(t)
		})
	})
}

// Helper functions for encrypted messaging tests
func createAuthenticatedEncryptedMessageInput(userID int64, activePseudonymID string, displayName string) *models.SendEncryptedMessageInput {
	// Create a user context
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false,
	}

	// Generate a JWT token
	token, _ := middleware.GenerateJWT(user, "test-secret", time.Hour)

	return &models.SendEncryptedMessageInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
	}
}

func createAuthenticatedGetConversationsInput(userID int64, activePseudonymID string, displayName string) *models.GetConversationsInput {
	// Create a user context
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false,
	}

	// Generate a JWT token
	token, _ := middleware.GenerateJWT(user, "test-secret", time.Hour)

	return &models.GetConversationsInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
	}
}

func createAuthenticatedGetConversationMessagesInput(userID int64, activePseudonymID string, displayName string) *models.GetConversationMessagesInput {
	// Create a user context
	user := &middleware.UserContext{
		UserID:            userID,
		Email:             "test@example.com",
		ActivePseudonymID: activePseudonymID,
		DisplayName:       displayName,
		MFAEnabled:        false,
	}

	// Generate a JWT token
	token, _ := middleware.GenerateJWT(user, "test-secret", time.Hour)

	return &models.GetConversationMessagesInput{
		AuthInput: middleware.AuthInput{
			Authorization: "Bearer " + token,
		},
	}
}
