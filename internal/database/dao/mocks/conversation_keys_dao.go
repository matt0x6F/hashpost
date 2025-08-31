package mocks

import (
	"context"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockConversationKeyDAO is a mock of ConversationKeyDAOInterface interface.
type MockConversationKeyDAO struct {
	mock.Mock
}

// CreateConversationKey mocks base method.
func (m *MockConversationKeyDAO) CreateConversationKey(ctx context.Context, participant1UserID, participant2UserID int64, encryptedConversationKey []byte, keyFingerprint string, expiresAt time.Time) (*models.ConversationKey, error) {
	args := m.Called(ctx, participant1UserID, participant2UserID, encryptedConversationKey, keyFingerprint, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConversationKey), args.Error(1)
}

// GetConversationKey mocks base method.
func (m *MockConversationKeyDAO) GetConversationKey(ctx context.Context, conversationID uuid.UUID) (*models.ConversationKey, error) {
	args := m.Called(ctx, conversationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConversationKey), args.Error(1)
}

// GetConversationKeyByParticipants mocks base method.
func (m *MockConversationKeyDAO) GetConversationKeyByParticipants(ctx context.Context, participant1UserID, participant2UserID int64) (*models.ConversationKey, error) {
	args := m.Called(ctx, participant1UserID, participant2UserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ConversationKey), args.Error(1)
}

// UpdateConversationKey mocks base method.
func (m *MockConversationKeyDAO) UpdateConversationKey(ctx context.Context, conversationID uuid.UUID, encryptedConversationKey []byte, keyFingerprint string, expiresAt time.Time) error {
	args := m.Called(ctx, conversationID, encryptedConversationKey, keyFingerprint, expiresAt)
	return args.Error(0)
}

// DeleteConversationKey mocks base method.
func (m *MockConversationKeyDAO) DeleteConversationKey(ctx context.Context, conversationID uuid.UUID) error {
	args := m.Called(ctx, conversationID)
	return args.Error(0)
}

// RotateConversationKey mocks base method.
func (m *MockConversationKeyDAO) RotateConversationKey(ctx context.Context, conversationID uuid.UUID, encryptedConversationKey []byte, keyFingerprint string, expiresAt time.Time) error {
	args := m.Called(ctx, conversationID, encryptedConversationKey, keyFingerprint, expiresAt)
	return args.Error(0)
}

// GetActiveConversationKeys mocks base method.
func (m *MockConversationKeyDAO) GetActiveConversationKeys(ctx context.Context, userID int64) ([]*models.ConversationKey, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]*models.ConversationKey), args.Error(1)
}

// CleanupExpiredKeys mocks base method.
func (m *MockConversationKeyDAO) CleanupExpiredKeys(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
