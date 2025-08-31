package models

import (
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
)

// SendEncryptedMessageInput represents the input for sending an encrypted message
type SendEncryptedMessageInput struct {
	middleware.AuthInput
	RecipientUserID int64  `json:"recipient_user_id" doc:"ID of the user to send the message to"`
	Content         string `json:"content" doc:"Plain text message content to be encrypted"`
}

// SendEncryptedMessageResponseBody represents the body of encrypted message response
type SendEncryptedMessageResponseBody struct {
	MessageID      int64     `json:"message_id" example:"123"`
	ConversationID string    `json:"conversation_id" example:"uuid-string"`
	ContentHash    string    `json:"content_hash" example:"sha256-hash"`
	SentAt         time.Time `json:"sent_at" example:"2024-01-01T20:00:00Z"`
	MessageStatus  string    `json:"message_status" example:"sent"`
}

// SendEncryptedMessageResponse represents the response after sending an encrypted message
type SendEncryptedMessageResponse struct {
	Status int                              `json:"-" example:"200"`
	Body   SendEncryptedMessageResponseBody `json:"body"`
}

// GetConversationsInput represents the input for retrieving conversations
type GetConversationsInput struct {
	middleware.AuthInput
	// No additional parameters needed - returns all conversations for authenticated user
}

// GetConversationsResponseBody represents the body of conversations response
type GetConversationsResponseBody struct {
	Conversations []ConversationSummary `json:"conversations"`
	TotalCount    int64                 `json:"total_count" example:"5"`
}

// GetConversationsResponse represents the response containing user conversations
type GetConversationsResponse struct {
	Status int                          `json:"-" example:"200"`
	Body   GetConversationsResponseBody `json:"body"`
}

// ConversationSummary represents a summary of a conversation
type ConversationSummary struct {
	ConversationID string    `json:"conversation_id" doc:"UUID of the conversation"`
	OtherUserID    int64     `json:"other_user_id" doc:"ID of the other participant in the conversation"`
	KeyFingerprint string    `json:"key_fingerprint" doc:"Fingerprint of the conversation encryption key"`
	MessageCount   int64     `json:"message_count" doc:"Number of messages in the conversation"`
	LastActive     time.Time `json:"last_active" doc:"Timestamp of the last activity in the conversation"`
	IsActive       bool      `json:"is_active" doc:"Whether the conversation is currently active"`
}

// GetConversationMessagesInput represents the input for retrieving messages from a conversation
type GetConversationMessagesInput struct {
	middleware.AuthInput
	ConversationID string `path:"conversation_id" doc:"UUID of the conversation to retrieve messages from"`
}

// GetConversationMessagesResponseBody represents the body of conversation messages response
type GetConversationMessagesResponseBody struct {
	ConversationID string           `json:"conversation_id" example:"uuid-string"`
	Messages       []MessageSummary `json:"messages"`
	TotalCount     int64            `json:"total_count" example:"10"`
}

// GetConversationMessagesResponse represents the response containing conversation messages
type GetConversationMessagesResponse struct {
	Status int                                 `json:"-" example:"200"`
	Body   GetConversationMessagesResponseBody `json:"body"`
}

// MessageSummary represents a summary of an encrypted message
type MessageSummary struct {
	MessageID      int64     `json:"message_id" doc:"Unique identifier for the message"`
	ConversationID string    `json:"conversation_id" doc:"UUID of the conversation this message belongs to"`
	ContentHash    string    `json:"content_hash" doc:"SHA-256 hash of the message content"`
	KeyVersion     int       `json:"key_version" doc:"Version of the encryption key used for this message"`
	CreatedAt      time.Time `json:"created_at" doc:"Timestamp when the message was created"`
}

// ConversationKeyInput represents the input for managing conversation keys
type ConversationKeyInput struct {
	middleware.AuthInput
	ConversationID string `path:"conversation_id" doc:"UUID of the conversation to manage keys for"`
}

// ConversationKeyResponse represents the response containing conversation key information
type ConversationKeyResponse struct {
	ConversationID string    `json:"conversation_id" doc:"UUID of the conversation"`
	KeyFingerprint string    `json:"key_fingerprint" doc:"Fingerprint of the conversation encryption key"`
	KeyVersion     int       `json:"key_version" doc:"Version of the encryption key"`
	ExpiresAt      time.Time `json:"expires_at" doc:"When the current key expires"`
	IsActive       bool      `json:"is_active" doc:"Whether the key is currently active"`
}
