package mocks

import (
	"context"
	"fmt"
	"reflect"

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

func (m *MockIdentityMappingDAO) UpdateIdentityMapping(ctx context.Context, mappingID string, updates *models.IdentityMappingSetter) error {
	args := m.Called(ctx, mappingID, updates)
	return args.Error(0)
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

func (m *MockIdentityMappingDAO) GetIdentityMappingsByFingerprint(ctx context.Context, fingerprint string) (models.IdentityMappingSlice, error) {
	args := m.Called(ctx, fingerprint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	// Handle both []*models.IdentityMapping and models.IdentityMappingSlice
	switch v := args.Get(0).(type) {
	case []*models.IdentityMapping:
		return models.IdentityMappingSlice(v), args.Error(1)
	case models.IdentityMappingSlice:
		return v, args.Error(1)
	default:
		// Try to convert if it's a slice
		if reflect.TypeOf(v).Kind() == reflect.Slice {
			// Convert to []*models.IdentityMapping first
			slice := reflect.ValueOf(v)
			result := make([]*models.IdentityMapping, slice.Len())
			for i := 0; i < slice.Len(); i++ {
				result[i] = slice.Index(i).Interface().(*models.IdentityMapping)
			}
			return models.IdentityMappingSlice(result), args.Error(1)
		}
		return nil, fmt.Errorf("unexpected type for GetIdentityMappingsByFingerprint: %T", v)
	}
}

func (m *MockIdentityMappingDAO) GetAllActiveIdentityMappings(ctx context.Context) (models.IdentityMappingSlice, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return models.IdentityMappingSlice(args.Get(0).([]*models.IdentityMapping)), args.Error(1)
}

func (m *MockIdentityMappingDAO) DeactivateIdentityMapping(ctx context.Context, mappingID string) error {
	args := m.Called(ctx, mappingID)
	return args.Error(0)
}

func (m *MockIdentityMappingDAO) DeleteByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockIdentityMappingDAO) GetCorrelationData(ctx context.Context, pseudonymID string) (*models.IdentityMapping, models.IdentityMappingSlice, error) {
	args := m.Called(ctx, pseudonymID)
	if args.Get(0) == nil {
		return nil, nil, args.Error(2)
	}
	if args.Get(1) == nil {
		return args.Get(0).(*models.IdentityMapping), nil, args.Error(2)
	}
	return args.Get(0).(*models.IdentityMapping), args.Get(1).(models.IdentityMappingSlice), args.Error(2)
}
