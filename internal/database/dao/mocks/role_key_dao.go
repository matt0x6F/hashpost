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

// NewMockRoleKeyDAO creates a new mock RoleKeyDAO
func NewMockRoleKeyDAO() *MockRoleKeyDAO {
	return &MockRoleKeyDAO{}
}

func (m *MockRoleKeyDAO) CreateRoleKey(ctx context.Context, roleName, scope string, keyData []byte, capabilities []string, expiresAt time.Time, createdByPseudonymID string, pseudonymID string, subforumID *int32) (*models.RoleKey, error) {
	args := m.Called(ctx, roleName, scope, keyData, capabilities, expiresAt, createdByPseudonymID, pseudonymID, subforumID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) GetRoleKey(ctx context.Context, pseudonymID string, scope string, subforumID *int32) (*models.RoleKey, error) {
	args := m.Called(ctx, pseudonymID, scope, subforumID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) ListRoleKeysByPseudonym(ctx context.Context, pseudonymID string) ([]*models.RoleKey, error) {
	args := m.Called(ctx, pseudonymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RoleKey), args.Error(1)
}

func (m *MockRoleKeyDAO) GetModeratorsForSubforum(ctx context.Context, subforumID int32) ([]*models.RoleKey, error) {
	args := m.Called(ctx, subforumID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.RoleKey), args.Error(1)
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

func (m *MockRoleKeyDAO) DeleteByPseudonymID(ctx context.Context, pseudonymID string) error {
	args := m.Called(ctx, pseudonymID)
	return args.Error(0)
}

func (m *MockRoleKeyDAO) DeactivateRoleKey(ctx context.Context, keyID string) error {
	args := m.Called(ctx, keyID)
	return args.Error(0)
}

func (m *MockRoleKeyDAO) ValidateKeyCapability(ctx context.Context, pseudonymID string, scope, requiredCapability string, subforumID *int32) (bool, error) {
	args := m.Called(ctx, pseudonymID, scope, requiredCapability, subforumID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleKeyDAO) GetKeyData(ctx context.Context, pseudonymID string, scope string, subforumID *int32) ([]byte, error) {
	args := m.Called(ctx, pseudonymID, scope, subforumID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockRoleKeyDAO) EnsureDefaultKeys(ctx context.Context, ibeSystem interface{}, pseudonymID string, userRoles []string) error {
	args := m.Called(ctx, ibeSystem, pseudonymID, userRoles)
	return args.Error(0)
}
