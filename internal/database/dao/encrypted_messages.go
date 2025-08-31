package dao

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// EncryptedMessageDAOInterface defines the interface for encrypted message operations
type EncryptedMessageDAOInterface interface {
	CreateEncryptedMessage(ctx context.Context, conversationID uuid.UUID, encryptedContent []byte, iv []byte, contentHash string, keyVersion int32, signature []byte) (*models.EncryptedMessage, error)
	GetEncryptedMessage(ctx context.Context, messageID int64) (*models.EncryptedMessage, error)
	GetMessagesByConversation(ctx context.Context, conversationID uuid.UUID) ([]*models.EncryptedMessage, error)
	UpdateEncryptedMessage(ctx context.Context, messageID int64, encryptedContent []byte, iv []byte, contentHash string, keyVersion int32, signature []byte) error
	DeleteEncryptedMessage(ctx context.Context, messageID int64) error
	GetMessageCountByConversation(ctx context.Context, conversationID uuid.UUID) (int64, error)
	SearchMessagesByContentHash(ctx context.Context, contentHash string) ([]*models.EncryptedMessage, error)
}

// EncryptedMessageDAO implements EncryptedMessageDAOInterface
type EncryptedMessageDAO struct {
	db bob.Executor
}

// NewEncryptedMessageDAO creates a new EncryptedMessageDAO
func NewEncryptedMessageDAO(db bob.Executor) *EncryptedMessageDAO {
	return &EncryptedMessageDAO{db: db}
}

// CreateEncryptedMessage creates a new encrypted message record
func (dao *EncryptedMessageDAO) CreateEncryptedMessage(ctx context.Context, conversationID uuid.UUID, encryptedContent []byte, iv []byte, contentHash string, keyVersion int32, signature []byte) (*models.EncryptedMessage, error) {
	// Create an encrypted message setter
	messageSetter := &models.EncryptedMessageSetter{
		ConversationID:   &conversationID,
		EncryptedContent: &encryptedContent,
		Iv:               &iv,
		ContentHash:      &contentHash,
		KeyVersion:       &keyVersion,
		Signature:        &signature,
	}

	// Use the generated EncryptedMessages table helper
	message, err := models.EncryptedMessages.Insert(messageSetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create encrypted message: %w", err)
	}

	return message, nil
}

// GetEncryptedMessage retrieves an encrypted message by message ID
func (dao *EncryptedMessageDAO) GetEncryptedMessage(ctx context.Context, messageID int64) (*models.EncryptedMessage, error) {
	// Use the generated FindEncryptedMessage function
	message, err := models.FindEncryptedMessage(ctx, dao.db, messageID)
	if err != nil {
		return nil, fmt.Errorf("failed to get encrypted message: %w", err)
	}

	return message, nil
}

// GetMessagesByConversation retrieves messages for a specific conversation
func (dao *EncryptedMessageDAO) GetMessagesByConversation(ctx context.Context, conversationID uuid.UUID) ([]*models.EncryptedMessage, error) {
	// Use the generated EncryptedMessages table helper with where clause
	messages, err := models.EncryptedMessages.Query(
		models.SelectWhere.EncryptedMessages.ConversationID.EQ(conversationID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by conversation: %w", err)
	}

	return messages, nil
}

// UpdateEncryptedMessage updates an existing encrypted message
func (dao *EncryptedMessageDAO) UpdateEncryptedMessage(ctx context.Context, messageID int64, encryptedContent []byte, iv []byte, contentHash string, keyVersion int32, signature []byte) error {
	// First get the existing message
	existingMessage, err := dao.GetEncryptedMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get encrypted message for update: %w", err)
	}
	if existingMessage == nil {
		return fmt.Errorf("encrypted message not found")
	}

	// Create an update setter
	updates := &models.EncryptedMessageSetter{
		EncryptedContent: &encryptedContent,
		Iv:               &iv,
		ContentHash:      &contentHash,
		KeyVersion:       &keyVersion,
		Signature:        &signature,
	}

	// Use the generated Update method
	err = existingMessage.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update encrypted message: %w", err)
	}

	return nil
}

// DeleteEncryptedMessage deletes an encrypted message
func (dao *EncryptedMessageDAO) DeleteEncryptedMessage(ctx context.Context, messageID int64) error {
	// First get the existing message
	existingMessage, err := dao.GetEncryptedMessage(ctx, messageID)
	if err != nil {
		return fmt.Errorf("failed to get encrypted message for deletion: %w", err)
	}
	if existingMessage == nil {
		return nil // Already deleted
	}

	// Use the generated Delete method
	err = existingMessage.Delete(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete encrypted message: %w", err)
	}

	return nil
}

// GetMessageCountByConversation gets the total count of messages in a conversation
func (dao *EncryptedMessageDAO) GetMessageCountByConversation(ctx context.Context, conversationID uuid.UUID) (int64, error) {
	// Use the generated EncryptedMessages table helper with count
	count, err := models.EncryptedMessages.Query(
		models.SelectWhere.EncryptedMessages.ConversationID.EQ(conversationID),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get message count by conversation: %w", err)
	}

	return count, nil
}

// SearchMessagesByContentHash searches for messages by their content hash
func (dao *EncryptedMessageDAO) SearchMessagesByContentHash(ctx context.Context, contentHash string) ([]*models.EncryptedMessage, error) {
	// Use the generated EncryptedMessages table helper with where clause
	messages, err := models.EncryptedMessages.Query(
		models.SelectWhere.EncryptedMessages.ContentHash.EQ(contentHash),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to search messages by content hash: %w", err)
	}

	return messages, nil
}
