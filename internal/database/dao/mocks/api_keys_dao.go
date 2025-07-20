package mocks

import (
	"context"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockAPIKeyDAO is a mock implementation of APIKeyDAOInterface
type MockAPIKeyDAO struct {
	mock.Mock
}

func (m *MockAPIKeyDAO) CreateAPIKey(ctx context.Context, keyName string, rawKey string, pseudonymID string, permissions *dao.APIKeyPermissions, expiresAt *time.Time) (*models.APIKey, error) {
	args := m.Called(ctx, keyName, rawKey, pseudonymID, permissions, expiresAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}

func (m *MockAPIKeyDAO) GetAPIKeyByHash(ctx context.Context, keyHash string) (*models.APIKey, error) {
	args := m.Called(ctx, keyHash)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}

func (m *MockAPIKeyDAO) GetAPIKeyByID(ctx context.Context, keyID int64) (*models.APIKey, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}

func (m *MockAPIKeyDAO) ValidateAPIKey(ctx context.Context, rawKey string) (*dao.APIKeyPermissions, string, error) {
	args := m.Called(ctx, rawKey)
	if args.Get(0) == nil {
		return nil, "", args.Error(2)
	}
	return args.Get(0).(*dao.APIKeyPermissions), args.String(1), args.Error(2)
}

func (m *MockAPIKeyDAO) UpdateAPIKey(ctx context.Context, keyID int64, updates *models.APIKeySetter) error {
	args := m.Called(ctx, keyID, updates)
	return args.Error(0)
}

func (m *MockAPIKeyDAO) DeleteAPIKey(ctx context.Context, keyID int64) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *MockAPIKeyDAO) DeactivateAPIKey(ctx context.Context, keyID int64) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *MockAPIKeyDAO) ActivateAPIKey(ctx context.Context, keyID int64) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *MockAPIKeyDAO) UpdateLastUsed(ctx context.Context, keyID int64) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *MockAPIKeyDAO) ListAPIKeys(ctx context.Context, limit, offset int) ([]*models.APIKey, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.APIKey), args.Error(1)
}

func (m *MockAPIKeyDAO) CountAPIKeys(ctx context.Context) (int64, error) {
	args := m.Called(ctx)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockAPIKeyDAO) GetExpiredAPIKeys(ctx context.Context) ([]*models.APIKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.APIKey), args.Error(1)
}

func (m *MockAPIKeyDAO) CleanupExpiredAPIKeys(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockAPIKeyDAO) GetAPIKeysByPseudonymID(ctx context.Context, pseudonymID string) ([]*models.APIKey, error) {
	args := m.Called(ctx, pseudonymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.APIKey), args.Error(1)
}

func (m *MockAPIKeyDAO) GetAPIKeyWithPseudonym(ctx context.Context, keyID int64) (*models.APIKey, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.APIKey), args.Error(1)
}
