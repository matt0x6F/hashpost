package handlers

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/rs/zerolog/log"
)

// MessagesHandler handles direct message requests
type MessagesHandler struct {
	directMessageDAO dao.DirectMessageDAOInterface
	userDAO          dao.UserDAOInterface
	pseudonymDAO     dao.PseudonymDAOInterface
}

// NewMessagesHandler creates a new messages handler
func NewMessagesHandler(
	directMessageDAO dao.DirectMessageDAOInterface,
	userDAO dao.UserDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
) *MessagesHandler {
	return &MessagesHandler{
		directMessageDAO: directMessageDAO,
		userDAO:          userDAO,
		pseudonymDAO:     pseudonymDAO,
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
