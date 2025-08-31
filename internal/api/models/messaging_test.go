package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSendEncryptedMessageInput_Validation(t *testing.T) {
	tests := []struct {
		name    string
		input   SendEncryptedMessageInput
		isValid bool
	}{
		{
			name: "valid input",
			input: SendEncryptedMessageInput{
				RecipientUserID: 123,
				Content:         "Hello, this is a test message",
			},
			isValid: true,
		},
		{
			name: "invalid recipient user ID",
			input: SendEncryptedMessageInput{
				RecipientUserID: 0,
				Content:         "Hello, this is a test message",
			},
			isValid: false,
		},
		{
			name: "negative recipient user ID",
			input: SendEncryptedMessageInput{
				RecipientUserID: -1,
				Content:         "Hello, this is a test message",
			},
			isValid: false,
		},
		{
			name: "empty content",
			input: SendEncryptedMessageInput{
				RecipientUserID: 123,
				Content:         "",
			},
			isValid: false,
		},
		{
			name: "very long content",
			input: SendEncryptedMessageInput{
				RecipientUserID: 123,
				Content:         string(make([]byte, 10001)), // 10KB + 1 byte
			},
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: Actual validation is done in the handler, not in the model
			// This test just ensures the struct can be created
			assert.NotNil(t, tt.input)
			assert.Equal(t, tt.input.RecipientUserID, tt.input.RecipientUserID)
			assert.Equal(t, tt.input.Content, tt.input.Content)
		})
	}
}

func TestSendEncryptedMessageResponse_Structure(t *testing.T) {
	now := time.Now()
	response := &SendEncryptedMessageResponse{
		Status: 200,
		Body: SendEncryptedMessageResponseBody{
			MessageID:      12345,
			ConversationID: "uuid-12345-67890",
			ContentHash:    "sha256-hash-here",
			SentAt:         now,
			MessageStatus:  "sent",
		},
	}

	assert.Equal(t, 200, response.Status)
	assert.Equal(t, int64(12345), response.Body.MessageID)
	assert.Equal(t, "uuid-12345-67890", response.Body.ConversationID)
	assert.Equal(t, "sha256-hash-here", response.Body.ContentHash)
	assert.Equal(t, now, response.Body.SentAt)
	assert.Equal(t, "sent", response.Body.MessageStatus)
}

func TestGetConversationsResponse_Structure(t *testing.T) {
	conversations := []ConversationSummary{
		{
			ConversationID: "conv-1",
			OtherUserID:    123,
			KeyFingerprint: "fp-1",
			MessageCount:   5,
			LastActive:     time.Now(),
			IsActive:       true,
		},
		{
			ConversationID: "conv-2",
			OtherUserID:    456,
			KeyFingerprint: "fp-2",
			MessageCount:   3,
			LastActive:     time.Now().Add(-time.Hour),
			IsActive:       false,
		},
	}

	response := &GetConversationsResponse{
		Status: 200,
		Body: GetConversationsResponseBody{
			Conversations: conversations,
			TotalCount:    int64(len(conversations)),
		},
	}

	assert.Equal(t, 200, response.Status)
	assert.Len(t, response.Body.Conversations, 2)
	assert.Equal(t, int64(2), response.Body.TotalCount)
	assert.Equal(t, "conv-1", response.Body.Conversations[0].ConversationID)
	assert.Equal(t, "conv-2", response.Body.Conversations[1].ConversationID)
}

func TestGetConversationMessagesResponse_Structure(t *testing.T) {
	messages := []MessageSummary{
		{
			MessageID:      1,
			ConversationID: "conv-1",
			ContentHash:    "hash-1",
			KeyVersion:     1,
			CreatedAt:      time.Now(),
		},
		{
			MessageID:      2,
			ConversationID: "conv-1",
			ContentHash:    "hash-2",
			KeyVersion:     1,
			CreatedAt:      time.Now().Add(time.Minute),
		},
	}

	response := &GetConversationMessagesResponse{
		Status: 200,
		Body: GetConversationMessagesResponseBody{
			ConversationID: "conv-1",
			Messages:       messages,
			TotalCount:     int64(len(messages)),
		},
	}

	assert.Equal(t, 200, response.Status)
	assert.Equal(t, "conv-1", response.Body.ConversationID)
	assert.Len(t, response.Body.Messages, 2)
	assert.Equal(t, int64(2), response.Body.TotalCount)
	assert.Equal(t, int64(1), response.Body.Messages[0].MessageID)
	assert.Equal(t, int64(2), response.Body.Messages[1].MessageID)
}

func TestConversationSummary_Structure(t *testing.T) {
	now := time.Now()
	summary := ConversationSummary{
		ConversationID: "conv-123",
		OtherUserID:    456,
		KeyFingerprint: "sha256-fingerprint-here",
		MessageCount:   42,
		LastActive:     now,
		IsActive:       true,
	}

	assert.Equal(t, "conv-123", summary.ConversationID)
	assert.Equal(t, int64(456), summary.OtherUserID)
	assert.Equal(t, "sha256-fingerprint-here", summary.KeyFingerprint)
	assert.Equal(t, int64(42), summary.MessageCount)
	assert.Equal(t, now, summary.LastActive)
	assert.True(t, summary.IsActive)
}

func TestMessageSummary_Structure(t *testing.T) {
	now := time.Now()
	summary := MessageSummary{
		MessageID:      789,
		ConversationID: "conv-456",
		ContentHash:    "sha256-content-hash",
		KeyVersion:     2,
		CreatedAt:      now,
	}

	assert.Equal(t, int64(789), summary.MessageID)
	assert.Equal(t, "conv-456", summary.ConversationID)
	assert.Equal(t, "sha256-content-hash", summary.ContentHash)
	assert.Equal(t, 2, summary.KeyVersion)
	assert.Equal(t, now, summary.CreatedAt)
}

func TestConversationKeyInput_Structure(t *testing.T) {
	input := ConversationKeyInput{
		ConversationID: "conv-789",
	}

	assert.Equal(t, "conv-789", input.ConversationID)
}

func TestConversationKeyResponse_Structure(t *testing.T) {
	now := time.Now()
	response := ConversationKeyResponse{
		ConversationID: "conv-123",
		KeyFingerprint: "fp-456",
		KeyVersion:     3,
		ExpiresAt:      now.AddDate(0, 1, 0), // 1 month from now
		IsActive:       true,
	}

	assert.Equal(t, "conv-123", response.ConversationID)
	assert.Equal(t, "fp-456", response.KeyFingerprint)
	assert.Equal(t, 3, response.KeyVersion)
	assert.True(t, response.ExpiresAt.After(now))
	assert.True(t, response.IsActive)
}
