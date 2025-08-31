package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/services"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// MessagesHandler handles direct message requests
type MessagesHandler struct {
	directMessageDAO     dao.DirectMessageDAOInterface
	userDAO              dao.UserDAOInterface
	pseudonymDAO         dao.PseudonymDAOInterface
	userEncryptionKeyDAO dao.UserEncryptionKeyDAOInterface
	conversationKeyDAO   dao.ConversationKeyDAOInterface
	encryptedMessageDAO  dao.EncryptedMessageDAOInterface
	encryptionService    services.EncryptionServiceInterface
	keyManagementService services.KeyManagementServiceInterface
	permissionDAO        dao.PermissionDAOInterface
	db                   bob.Executor
}

// NewMessagesHandler creates a new messages handler
func NewMessagesHandler(
	directMessageDAO dao.DirectMessageDAOInterface,
	userDAO dao.UserDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
	userEncryptionKeyDAO dao.UserEncryptionKeyDAOInterface,
	conversationKeyDAO dao.ConversationKeyDAOInterface,
	encryptedMessageDAO dao.EncryptedMessageDAOInterface,
	encryptionService services.EncryptionServiceInterface,
	keyManagementService services.KeyManagementServiceInterface,
	permissionDAO dao.PermissionDAOInterface,
	db bob.Executor,
) *MessagesHandler {
	return &MessagesHandler{
		directMessageDAO:     directMessageDAO,
		userDAO:              userDAO,
		pseudonymDAO:         pseudonymDAO,
		userEncryptionKeyDAO: userEncryptionKeyDAO,
		conversationKeyDAO:   conversationKeyDAO,
		encryptedMessageDAO:  encryptedMessageDAO,
		encryptionService:    encryptionService,
		keyManagementService: keyManagementService,
		permissionDAO:        permissionDAO,
		db:                   db,
	}
}

// SendDirectMessage handles sending a direct message to another user
func (h *MessagesHandler) SendDirectMessage(ctx context.Context, input *models.DirectMessageInput) (*models.DirectMessageResponse, error) {
	// Extract user from Huma input
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	log.Info().
		Str("endpoint", "messages").
		Str("component", "handler").
		Int64("user_id", user.UserID).
		Str("recipient_pseudonym_id", input.Body.RecipientPseudonymID).
		Msg("Send direct message requested")

	// Create the direct message
	message, err := h.directMessageDAO.CreateDirectMessage(
		ctx,
		user.ActivePseudonymID,
		input.Body.RecipientPseudonymID,
		input.Body.Content,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create direct message")
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	// Update last active timestamp for the pseudonym since sending messages represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, user.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", user.ActivePseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	response := models.NewDirectMessageResponse(
		int(message.MessageID),
		message.RecipientPseudonymID,
		message.Content,
	)

	log.Info().
		Str("endpoint", "messages").
		Str("component", "handler").
		Int64("user_id", user.UserID).
		Int64("message_id", message.MessageID).
		Msg("Send direct message completed")

	return response, nil
}

// SendEncryptedMessage handles sending an encrypted message
func (h *MessagesHandler) SendEncryptedMessage(ctx context.Context, input *models.SendEncryptedMessageInput) (*models.SendEncryptedMessageResponse, error) {
	// Extract user context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Check permissions using unified capability system
	hasCapability, err := h.checkMessagingCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilitySendDirectMessages)
	if err != nil {
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		return nil, huma.Error403Forbidden("insufficient permissions for messaging")
	}

	// Ensure messaging keys exist for the user
	err = h.keyManagementService.EnsureMessagingKeys(ctx, userCtx.UserID, userCtx.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to ensure messaging keys")
		return nil, fmt.Errorf("failed to ensure messaging keys: %w", err)
	}

	// Validate input
	if err := h.validateSendMessageInput(input); err != nil {
		return nil, huma.Error400BadRequest(err.Error())
	}

	// Get or create conversation key
	conversationKey, err := h.getOrCreateConversationKey(ctx, userCtx.UserID, input.RecipientUserID)
	if err != nil {
		log.Error().Err(err).Int64("sender_id", userCtx.UserID).Int64("recipient_id", input.RecipientUserID).Msg("Failed to get/create conversation key")
		return nil, fmt.Errorf("failed to establish secure conversation: %w", err)
	}

	// Encrypt the message content
	encryptedContent, iv, err := h.encryptMessageContent(ctx, userCtx.UserID, input.Content, conversationKey)
	if err != nil {
		log.Error().Err(err).Msg("Failed to encrypt message content")
		return nil, fmt.Errorf("failed to encrypt message: %w", err)
	}

	// Generate content hash for integrity verification
	contentHash := h.generateContentHash(input.Content)

	// Create the encrypted message
	message, err := h.encryptedMessageDAO.CreateEncryptedMessage(
		ctx,
		conversationKey.ConversationID,
		encryptedContent,
		iv,
		contentHash,
		conversationKey.KeyVersion,
		[]byte{}, // Signature will be added later
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create encrypted message")
		return nil, fmt.Errorf("failed to store encrypted message: %w", err)
	}

	// Generate response
	response := &models.SendEncryptedMessageResponse{
		Status: 200,
		Body: models.SendEncryptedMessageResponseBody{
			MessageID:      message.MessageID,
			ConversationID: conversationKey.ConversationID.String(),
			ContentHash:    contentHash,
			SentAt:         time.Now(),
			MessageStatus:  "sent",
		},
	}

	log.Info().
		Int64("message_id", message.MessageID).
		Int64("sender_id", userCtx.UserID).
		Int64("recipient_id", input.RecipientUserID).
		Msg("Encrypted message sent successfully")

	return response, nil
}

// GetConversations retrieves all conversations for the authenticated user
func (h *MessagesHandler) GetConversations(ctx context.Context, input *models.GetConversationsInput) (*models.GetConversationsResponse, error) {
	// Extract user context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Check permissions using unified capability system
	hasCapability, err := h.checkMessagingCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityReceiveDirectMessages)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to check messaging capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks receive messaging capability")
		return nil, huma.Error403Forbidden("insufficient permissions to view conversations")
	}

	// Get active conversation keys for the user
	conversationKeys, err := h.conversationKeyDAO.GetActiveConversationKeys(ctx, userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to get active conversations")
		return nil, fmt.Errorf("failed to retrieve conversations: %w", err)
	}

	// Build conversation list
	conversations := make([]models.ConversationSummary, 0, len(conversationKeys))
	for _, key := range conversationKeys {
		// Determine the other participant
		otherUserID := key.Participant1UserID
		if otherUserID == userCtx.UserID {
			otherUserID = key.Participant2UserID
		}

		// Get message count for this conversation
		messageCount, err := h.encryptedMessageDAO.GetMessageCountByConversation(ctx, key.ConversationID)
		if err != nil {
			log.Warn().Err(err).Str("conversation_id", key.ConversationID.String()).Msg("Failed to get message count")
			messageCount = 0
		}

		conversation := models.ConversationSummary{
			ConversationID: key.ConversationID.String(),
			OtherUserID:    otherUserID,
			KeyFingerprint: key.KeyFingerprint,
			MessageCount:   messageCount,
			LastActive:     key.CreatedAt.V, // Use .V for sql.Null[time.Time]
			IsActive:       key.IsActive.V,  // Use .V for sql.Null[bool]
		}
		conversations = append(conversations, conversation)
	}

	response := &models.GetConversationsResponse{
		Status: 200,
		Body: models.GetConversationsResponseBody{
			Conversations: conversations,
			TotalCount:    int64(len(conversations)),
		},
	}

	return response, nil
}

// GetConversationMessages retrieves messages for a specific conversation
func (h *MessagesHandler) GetConversationMessages(ctx context.Context, input *models.GetConversationMessagesInput) (*models.GetConversationMessagesResponse, error) {
	// Extract user context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Check permissions using unified capability system
	hasCapability, err := h.checkMessagingCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityReceiveDirectMessages)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to check messaging capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks receive messaging capability")
		return nil, huma.Error403Forbidden("insufficient permissions to view messages")
	}

	// Parse conversation ID
	conversationID := uuid.FromStringOrNil(input.ConversationID)
	if conversationID == uuid.Nil {
		return nil, huma.Error400BadRequest("invalid conversation ID format")
	}

	// Verify user is a participant in this conversation
	conversationKey, err := h.conversationKeyDAO.GetConversationKey(ctx, conversationID)
	if err != nil {
		log.Error().Err(err).Str("conversation_id", input.ConversationID).Msg("Failed to get conversation key")
		return nil, fmt.Errorf("failed to retrieve conversation: %w", err)
	}
	if conversationKey == nil {
		return nil, huma.Error404NotFound("conversation not found")
	}

	// Check if user is a participant
	if conversationKey.Participant1UserID != userCtx.UserID && conversationKey.Participant2UserID != userCtx.UserID {
		log.Warn().Int64("user_id", userCtx.UserID).Str("conversation_id", input.ConversationID).Msg("User not a participant in conversation")
		return nil, huma.Error403Forbidden("access denied to conversation")
	}

	// Get messages for the conversation
	messages, err := h.encryptedMessageDAO.GetMessagesByConversation(ctx, conversationID)
	if err != nil {
		log.Error().Err(err).Str("conversation_id", input.ConversationID).Msg("Failed to get conversation messages")
		return nil, fmt.Errorf("failed to retrieve messages: %w", err)
	}

	// Build message summaries (without decrypted content)
	messageSummaries := make([]models.MessageSummary, 0, len(messages))
	for _, msg := range messages {
		// Get creation time from the message
		createdAt := time.Now()
		if msg.CreatedAt.Valid {
			createdAt = msg.CreatedAt.V
		}

		summary := models.MessageSummary{
			MessageID:      msg.MessageID,
			ConversationID: msg.ConversationID.String(),
			ContentHash:    msg.ContentHash,
			KeyVersion:     int(msg.KeyVersion),
			CreatedAt:      createdAt,
		}
		messageSummaries = append(messageSummaries, summary)
	}

	response := &models.GetConversationMessagesResponse{
		Status: 200,
		Body: models.GetConversationMessagesResponseBody{
			ConversationID: input.ConversationID,
			Messages:       messageSummaries,
			TotalCount:     int64(len(messageSummaries)),
		},
	}

	return response, nil
}

// Helper methods

// checkMessagingCapability checks if the user has the required messaging capability
func (h *MessagesHandler) checkMessagingCapability(ctx context.Context, userID int64, pseudonymID string, capability string) (bool, error) {
	// Use the unified permission system to check messaging capabilities
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(
		ctx, userID, pseudonymID, capability, nil) // nil = global scope for messaging
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Str("capability", capability).Msg("Failed to check messaging capability")
		return false, fmt.Errorf("failed to check permissions: %w", err)
	}

	if !hasCapability {
		log.Warn().Int64("user_id", userID).Str("capability", capability).Msg("User lacks required messaging capability")
		return false, nil
	}

	return true, nil
}

// validateSendMessageInput validates the send message input
func (h *MessagesHandler) validateSendMessageInput(input *models.SendEncryptedMessageInput) error {
	if input.RecipientUserID <= 0 {
		return fmt.Errorf("invalid recipient user ID")
	}
	if len(input.Content) == 0 {
		return fmt.Errorf("message content cannot be empty")
	}
	if len(input.Content) > 10000 { // 10KB limit
		return fmt.Errorf("message content too long (max 10KB)")
	}
	return nil
}

// getOrCreateConversationKey gets an existing conversation key or creates a new one
func (h *MessagesHandler) getOrCreateConversationKey(ctx context.Context, senderID, recipientID int64) (*dbmodels.ConversationKey, error) {
	// Try to get existing conversation key
	conversationKey, err := h.conversationKeyDAO.GetConversationKeyByParticipants(ctx, senderID, recipientID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing conversation: %w", err)
	}

	// If conversation key exists and is active, return it
	if conversationKey != nil && conversationKey.IsActive.V {
		return conversationKey, nil
	}

	// Create new conversation key using the encryption service
	conversationKeyData, err := h.encryptionService.GenerateAESKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate conversation key: %w", err)
	}

	// Generate fingerprint for the key
	keyFingerprint := h.generateKeyFingerprint(conversationKeyData)

	// Encrypt the conversation key with the sender's public key
	// In production, this should be encrypted with both users' public keys
	senderPublicKey, err := h.userEncryptionKeyDAO.GetUserPublicKey(ctx, senderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sender public key: %w", err)
	}

	encryptedKey, err := h.encryptionService.EncryptWithPublicKey(senderPublicKey, conversationKeyData)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt conversation key with public key: %w", err)
	}

	expiresAt := time.Now().AddDate(0, 1, 0) // Expire in 1 month

	newConversationKey, err := h.conversationKeyDAO.CreateConversationKey(
		ctx,
		senderID,
		recipientID,
		encryptedKey,
		keyFingerprint,
		expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation key: %w", err)
	}

	return newConversationKey, nil
}

// encryptMessageContent encrypts the message content using the conversation key
func (h *MessagesHandler) encryptMessageContent(ctx context.Context, userID int64, content string, conversationKey *dbmodels.ConversationKey) ([]byte, []byte, error) {
	// Decrypt the conversation key using the user's private key
	// First, get the user's private key
	userKey, err := h.userEncryptionKeyDAO.GetUserEncryptionKey(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user encryption key: %w", err)
	}
	if userKey == nil {
		return nil, nil, fmt.Errorf("user encryption key not found")
	}

	// Decrypt the conversation key using the user's private key
	keyData, err := h.encryptionService.DecryptWithPrivateKey(userKey.EncryptedSignatureKey, conversationKey.EncryptedSharedKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt conversation key with private key: %w", err)
	}

	// Use the encryption service to encrypt the message
	encryptedMessage, err := h.encryptionService.EncryptAES(keyData, []byte(content))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to encrypt message: %w", err)
	}

	return encryptedMessage.EncryptedContent, encryptedMessage.IV, nil
}

// generateContentHash generates a SHA-256 hash of the message content
func (h *MessagesHandler) generateContentHash(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// generateKeyFingerprint generates a SHA-256 hash of the key data
func (h *MessagesHandler) generateKeyFingerprint(keyData []byte) string {
	hash := sha256.Sum256(keyData)
	return hex.EncodeToString(hash[:])
}

// GetDirectMessages handles getting direct messages for the current user
func (h *MessagesHandler) GetDirectMessages(ctx context.Context, input *models.DirectMessageListInput) (*models.DirectMessageListResponse, error) {
	// Extract user from Huma input
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("authentication required")
	}

	log.Info().
		Str("endpoint", "messages").
		Str("component", "handler").
		Int64("user_id", user.UserID).
		Msg("Get direct messages requested")

	// Get messages for the user
	messages, err := h.directMessageDAO.GetDirectMessagesByPseudonym(
		ctx,
		user.ActivePseudonymID,
		input.Page,
		input.Limit,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get direct messages")
		return nil, fmt.Errorf("failed to get messages: %w", err)
	}

	// Convert database messages to API messages
	apiMessages := make([]models.DirectMessage, len(messages))
	for i, msg := range messages {
		// Handle nullable fields
		isRead := false
		if msg.IsRead.Valid {
			isRead = msg.IsRead.V
		}

		createdAt := ""
		if msg.CreatedAt.Valid {
			createdAt = msg.CreatedAt.V.Format("2006-01-02T15:04:05Z")
		}

		apiMessages[i] = models.DirectMessage{
			MessageID:            int(msg.MessageID),
			SenderPseudonymID:    msg.SenderPseudonymID,
			SenderDisplayName:    user.DisplayName, // Use current user's display name for now
			RecipientPseudonymID: msg.RecipientPseudonymID,
			Content:              msg.Content,
			IsRead:               isRead,
			CreatedAt:            createdAt,
		}
	}

	// Get total count for pagination
	total, err := h.directMessageDAO.CountDirectMessagesByPseudonym(ctx, user.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count direct messages")
		return nil, fmt.Errorf("failed to count messages: %w", err)
	}

	response := models.NewDirectMessageListResponse(apiMessages, input.Page, input.Limit, int(total))

	log.Info().
		Str("endpoint", "messages").
		Str("component", "handler").
		Int64("user_id", user.UserID).
		Int("count", len(apiMessages)).
		Int("total", int(total)).
		Msg("Get direct messages completed")

	return response, nil
}
