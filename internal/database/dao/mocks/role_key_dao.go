package mocks

import (
	"context"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockRoleKeyDAO is a mock implementation of RoleKeyDAOInterface
type MockRoleKeyDAO struct {
	mock.Mock
}

func (m *MockRoleKeyDAO) CreateRoleKey(ctx context.Context, roleName, scope string, keyData []byte, capabilities []string, expiresAt time.Time, createdBy int64) (*models.RoleKey, error) {
	args := m.Called(ctx, roleName, scope, keyData, capabilities, expiresAt, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) GetRoleKey(ctx context.Context, roleName, scope string) (*models.RoleKey, error) {
	args := m.Called(ctx, roleName, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) GetPerUserRoleKey(ctx context.Context, roleName, scope string, createdBy int64) (*models.RoleKey, error) {
	args := m.Called(ctx, roleName, scope, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) GetRoleKeyByID(ctx context.Context, keyID string) (*models.RoleKey, error) {
	args := m.Called(ctx, keyID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) ListRoleKeys(ctx context.Context) ([]*models.RoleKey, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) ListRoleKeysByRole(ctx context.Context, roleName string) ([]*models.RoleKey, error) {
	args := m.Called(ctx, roleName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) DeactivateRoleKey(ctx context.Context, keyID string) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *MockRoleKeyDAO) ValidateKeyCapability(ctx context.Context, roleName, scope, requiredCapability string) (bool, error) {
	args := m.Called(ctx, roleName, scope, requiredCapability)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleKeyDAO) GetKeyData(ctx context.Context, roleName, scope string) ([]byte, error) {
	args := m.Called(ctx, roleName, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRoleKeyDAO) GetPerUserKeyData(ctx context.Context, roleName, scope string, createdBy int64) ([]byte, error) {
	args := m.Called(ctx, roleName, scope, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRoleKeyDAO) EnsureDefaultKeys(ctx context.Context, ibeSystem interface{}, userID int64) error {
	args := m.Called(ctx, ibeSystem, userID)
	return args.Error(0)
}
