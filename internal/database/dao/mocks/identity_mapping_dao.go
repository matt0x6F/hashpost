package mocks

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockIdentityMappingDAO is a mock implementation of IdentityMappingDAOInterface
type MockIdentityMappingDAO struct {
	mock.Mock
}

// NewMockIdentityMappingDAO creates a new mock IdentityMappingDAO
func NewMockIdentityMappingDAO() *MockIdentityMappingDAO {
	return &MockIdentityMappingDAO{}
}

func (m *MockIdentityMappingDAO) CreateIdentityMapping(ctx context.Context, mapping *models.IdentityMappingSetter) (*models.IdentityMapping, error) {
	args := m.Called(ctx, mapping)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IdentityMapping), args.Error(1)
}

func (m *MockIdentityMappingDAO) GetIdentityMappingByPseudonymID(ctx context.Context, pseudonymID string) (*models.IdentityMapping, error) {
	args := m.Called(ctx, pseudonymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.IdentityMapping), args.Error(1)
}

func (m *MockIdentityMappingDAO) GetIdentityMappingsByPseudonymID(ctx context.Context, pseudonymID string) (models.IdentityMappingSlice, error) {
	args := m.Called(ctx, pseudonymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.IdentityMappingSlice), args.Error(1)
}

func (m *MockIdentityMappingDAO) GetIdentityMappingsByUserID(ctx context.Context, userID int64) (models.IdentityMappingSlice, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return models.IdentityMappingSlice(args.Get(0).([]*models.IdentityMapping)), args.Error(1)
}

func (m *MockIdentityMappingDAO) GetIdentityMappingsByFingerprint(ctx context.Context, fingerprint string) (models.IdentityMappingSlice, error) {
	args := m.Called(ctx, fingerprint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(models.IdentityMappingSlice), args.Error(1)
}

func (m *MockIdentityMappingDAO) DeactivateIdentityMapping(ctx context.Context, mappingID string) error {
	args := m.Called(ctx, mappingID)
	return args.Error(0)
}
