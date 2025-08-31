package mocks

import (
	"context"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockEncryptedMessageDAO is a mock of EncryptedMessageDAOInterface interface.
type MockEncryptedMessageDAO struct {
	mock.Mock
}

// CreateEncryptedMessage mocks base method.
func (m *MockEncryptedMessageDAO) CreateEncryptedMessage(ctx context.Context, conversationID uuid.UUID, encryptedContent, iv []byte, contentHash string, keyVersion int32, signature []byte) (*models.EncryptedMessage, error) {
	args := m.Called(ctx, conversationID, encryptedContent, iv, contentHash, keyVersion, signature)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EncryptedMessage), args.Error(1)
}

// GetEncryptedMessage mocks base method.
func (m *MockEncryptedMessageDAO) GetEncryptedMessage(ctx context.Context, messageID int64) (*models.EncryptedMessage, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EncryptedMessage), args.Error(1)
}

// GetMessagesByConversation mocks base method.
func (m *MockEncryptedMessageDAO) GetMessagesByConversation(ctx context.Context, conversationID uuid.UUID) ([]*models.EncryptedMessage, error) {
	args := m.Called(ctx, conversationID)
	return args.Get(0).([]*models.EncryptedMessage), args.Error(1)
}

// UpdateEncryptedMessage mocks base method.
func (m *MockEncryptedMessageDAO) UpdateEncryptedMessage(ctx context.Context, messageID int64, encryptedContent, iv []byte, contentHash string, keyVersion int32, signature []byte) error {
	args := m.Called(ctx, messageID, encryptedContent, iv, contentHash, keyVersion, signature)
	return args.Error(0)
}

// DeleteEncryptedMessage mocks base method.
func (m *MockEncryptedMessageDAO) DeleteEncryptedMessage(ctx context.Context, messageID int64) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

// GetMessageCountByConversation mocks base method.
func (m *MockEncryptedMessageDAO) GetMessageCountByConversation(ctx context.Context, conversationID uuid.UUID) (int64, error) {
	args := m.Called(ctx, conversationID)
	return args.Get(0).(int64), args.Error(1)
}

// SearchMessagesByContentHash mocks base method.
func (m *MockEncryptedMessageDAO) SearchMessagesByContentHash(ctx context.Context, contentHash string) ([]*models.EncryptedMessage, error) {
	args := m.Called(ctx, contentHash)
	return args.Get(0).([]*models.EncryptedMessage), args.Error(1)
}
