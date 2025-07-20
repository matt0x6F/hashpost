package mocks

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockUserBlocksDAO is a mock implementation of UserBlocksDAOInterface
type MockUserBlocksDAO struct {
	mock.Mock
}

func (m *MockUserBlocksDAO) CreateUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string, blockedUserID int64) (*models.UserBlock, error) {
	args := m.Called(ctx, blockerPseudonymID, blockedPseudonymID, blockedUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserBlock), args.Error(1)
}

func (m *MockUserBlocksDAO) GetUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (*models.UserBlock, error) {
	args := m.Called(ctx, blockerPseudonymID, blockedPseudonymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserBlock), args.Error(1)
}

func (m *MockUserBlocksDAO) GetUserBlocksByBlocker(ctx context.Context, blockerPseudonymID string) ([]*models.UserBlock, error) {
	args := m.Called(ctx, blockerPseudonymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserBlock), args.Error(1)
}

func (m *MockUserBlocksDAO) GetUserBlocksByBlockedUser(ctx context.Context, blockedUserID int64) ([]*models.UserBlock, error) {
	args := m.Called(ctx, blockedUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserBlock), args.Error(1)
}

func (m *MockUserBlocksDAO) DeleteUserBlock(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) error {
	args := m.Called(ctx, blockerPseudonymID, blockedPseudonymID)
	return args.Error(0)
}

func (m *MockUserBlocksDAO) DeleteUserBlockByID(ctx context.Context, blockID int64) error {
	args := m.Called(ctx, blockID)
	return args.Error(0)
}

func (m *MockUserBlocksDAO) IsUserBlocked(ctx context.Context, blockerPseudonymID, blockedPseudonymID string) (bool, error) {
	args := m.Called(ctx, blockerPseudonymID, blockedPseudonymID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserBlocksDAO) IsPseudonymBlockedByUser(ctx context.Context, blockerPseudonymID, blockedPseudonymID string, blockedUserID int64) (bool, error) {
	args := m.Called(ctx, blockerPseudonymID, blockedPseudonymID, blockedUserID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserBlocksDAO) IsUserBlockedAtFingerprintLevel(ctx context.Context, blockerPseudonymID string, blockedUserID int64) (bool, error) {
	args := m.Called(ctx, blockerPseudonymID, blockedUserID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserBlocksDAO) IsUserBlockedByAnyPseudonym(ctx context.Context, blockerUserID int64, blockedPseudonymID string) (bool, error) {
	args := m.Called(ctx, blockerUserID, blockedPseudonymID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserBlocksDAO) GetFingerprintLevelBlocks(ctx context.Context, blockedUserID int64) ([]*models.UserBlock, error) {
	args := m.Called(ctx, blockedUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserBlock), args.Error(1)
}
