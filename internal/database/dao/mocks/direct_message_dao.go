package mocks

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockDirectMessageDAO is a mock implementation of DirectMessageDAOInterface
type MockDirectMessageDAO struct {
	mock.Mock
}

// CreateDirectMessage mocks the CreateDirectMessage method
func (m *MockDirectMessageDAO) CreateDirectMessage(ctx context.Context, senderPseudonymID, recipientPseudonymID, content string) (*models.DirectMessage, error) {
	args := m.Called(ctx, senderPseudonymID, recipientPseudonymID, content)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DirectMessage), args.Error(1)
}

// GetDirectMessageByID mocks the GetDirectMessageByID method
func (m *MockDirectMessageDAO) GetDirectMessageByID(ctx context.Context, messageID int64) (*models.DirectMessage, error) {
	args := m.Called(ctx, messageID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.DirectMessage), args.Error(1)
}

// GetDirectMessagesByPseudonym mocks the GetDirectMessagesByPseudonym method
func (m *MockDirectMessageDAO) GetDirectMessagesByPseudonym(ctx context.Context, pseudonymID string, page, limit int) ([]*models.DirectMessage, error) {
	args := m.Called(ctx, pseudonymID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DirectMessage), args.Error(1)
}

// GetDirectMessagesBetweenPseudonyms mocks the GetDirectMessagesBetweenPseudonyms method
func (m *MockDirectMessageDAO) GetDirectMessagesBetweenPseudonyms(ctx context.Context, pseudonymID1, pseudonymID2 string, page, limit int) ([]*models.DirectMessage, error) {
	args := m.Called(ctx, pseudonymID1, pseudonymID2, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.DirectMessage), args.Error(1)
}

// CountDirectMessagesByPseudonym mocks the CountDirectMessagesByPseudonym method
func (m *MockDirectMessageDAO) CountDirectMessagesByPseudonym(ctx context.Context, pseudonymID string) (int64, error) {
	args := m.Called(ctx, pseudonymID)
	return args.Get(0).(int64), args.Error(1)
}

// MarkMessageAsRead mocks the MarkMessageAsRead method
func (m *MockDirectMessageDAO) MarkMessageAsRead(ctx context.Context, messageID int64) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

// DeleteDirectMessage mocks the DeleteDirectMessage method
func (m *MockDirectMessageDAO) DeleteDirectMessage(ctx context.Context, messageID int64) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

// IsUserBlocked mocks the IsUserBlocked method
func (m *MockDirectMessageDAO) IsUserBlocked(ctx context.Context, senderPseudonymID, recipientPseudonymID string) (bool, error) {
	args := m.Called(ctx, senderPseudonymID, recipientPseudonymID)
	return args.Bool(0), args.Error(1)
}
