package mocks

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/services"
	"github.com/stretchr/testify/mock"
)

// MockKeyManagementService is a mock of KeyManagementServiceInterface
type MockKeyManagementService struct {
	mock.Mock
}

// Ensure MockKeyManagementService implements KeyManagementServiceInterface
var _ services.KeyManagementServiceInterface = (*MockKeyManagementService)(nil)

// EnsureMessagingKeys mocks base method
func (m *MockKeyManagementService) EnsureMessagingKeys(ctx context.Context, userID int64, pseudonymID string) error {
	args := m.Called(ctx, userID, pseudonymID)
	return args.Error(0)
}

// GenerateMessageKey mocks base method
func (m *MockKeyManagementService) GenerateMessageKey() (*services.MessageKey, error) {
	args := m.Called()
	return args.Get(0).(*services.MessageKey), args.Error(1)
}

// GenerateConversationKey mocks base method
func (m *MockKeyManagementService) GenerateConversationKey() (*services.ConversationKey, error) {
	args := m.Called()
	return args.Get(0).(*services.ConversationKey), args.Error(1)
}
