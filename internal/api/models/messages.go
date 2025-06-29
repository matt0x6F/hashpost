package models

import "time"

// DirectMessageInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type DirectMessageInputBody struct {
	RecipientPseudonymID string `json:"recipient_pseudonym_id" example:"def789ghi012..." required:"true"`
	Content              string `json:"content" example:"Hello! I wanted to discuss..." required:"true"`
}

// DirectMessageInput represents direct message request (for OpenAPI schema only)
type DirectMessageInput struct {
	Body DirectMessageInputBody `json:"body"`
}

// DirectMessageListInput represents direct message list request parameters
type DirectMessageListInput struct {
	Page  int `query:"page" example:"1"`
	Limit int `query:"limit" example:"25"`
}

// DirectMessage represents a direct message
type DirectMessage struct {
	MessageID            int    `json:"message_id" example:"123"`
	SenderPseudonymID    string `json:"sender_pseudonym_id" example:"def789ghi012..."`
	SenderDisplayName    string `json:"sender_display_name" example:"sender_name"`
	RecipientPseudonymID string `json:"recipient_pseudonym_id" example:"abc123def456..."`
	Content              string `json:"content" example:"Hello! I wanted to discuss..."`
	IsRead               bool   `json:"is_read" example:"false"`
	CreatedAt            string `json:"created_at" example:"2024-01-01T20:00:00Z"`
}

// DirectMessageResponseBody represents the body of direct message creation response
type DirectMessageResponseBody struct {
	MessageID            int    `json:"message_id" example:"123"`
	RecipientPseudonymID string `json:"recipient_pseudonym_id" example:"def789ghi012..."`
	Content              string `json:"content" example:"Hello! I wanted to discuss..."`
	CreatedAt            string `json:"created_at" example:"2024-01-01T20:00:00Z"`
}

// DirectMessageListResponseBody represents the body of direct message list response
type DirectMessageListResponseBody struct {
	Messages   []DirectMessage `json:"messages"`
	Pagination Pagination      `json:"pagination"`
}

// DirectMessageResponse represents direct message creation response
type DirectMessageResponse struct {
	Status int                       `json:"-" example:"200"`
	Body   DirectMessageResponseBody `json:"body"`
}

// DirectMessageListResponse represents direct message list response
type DirectMessageListResponse struct {
	Status int                           `json:"-" example:"200"`
	Body   DirectMessageListResponseBody `json:"body"`
}

// NewDirectMessageResponse creates a new direct message response
func NewDirectMessageResponse(messageID int, recipientPseudonymID, content string) *DirectMessageResponse {
	return &DirectMessageResponse{
		Status: 200,
		Body: DirectMessageResponseBody{
			MessageID:            messageID,
			RecipientPseudonymID: recipientPseudonymID,
			Content:              content,
			CreatedAt:            time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// NewDirectMessageListResponse creates a new direct message list response
func NewDirectMessageListResponse(messages []DirectMessage, page, limit, total int) *DirectMessageListResponse {
	pages := (total + limit - 1) / limit // Ceiling division

	return &DirectMessageListResponse{
		Status: 200,
		Body: DirectMessageListResponseBody{
			Messages: messages,
			Pagination: Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: pages,
			},
		},
	}
}
