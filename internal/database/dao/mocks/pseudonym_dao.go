package mocks

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockPseudonymDAO is a mock implementation of PseudonymDAOInterface with data injection support
type MockPseudonymDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	pseudonyms map[string]*models.Pseudonym
}

// NewMockPseudonymDAO creates a new mock PseudonymDAO with optional initial data
func NewMockPseudonymDAO() *MockPseudonymDAO {
	return &MockPseudonymDAO{
		pseudonyms: make(map[string]*models.Pseudonym),
	}
}

// InjectPseudonym injects a pseudonym into the mock for testing
func (m *MockPseudonymDAO) InjectPseudonym(pseudonym *models.Pseudonym) {
	m.pseudonyms[pseudonym.PseudonymID] = pseudonym
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockPseudonymDAO) SetDefaultBehavior() {
	// Default behavior for GetPseudonymByID
	m.On("GetPseudonymByID", mock.Anything, mock.AnythingOfType("string")).Return(
		func(ctx context.Context, pseudonymID string) (*models.Pseudonym, error) {
			if pseudonym, exists := m.pseudonyms[pseudonymID]; exists {
				return pseudonym, nil
			}
			return nil, nil
		},
	).Maybe()

	// Default behavior for GetUserIDByPseudonym
	m.On("GetUserIDByPseudonym", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, pseudonymID, roleName, scope string) (int64, error) {
			// Return a mock user ID for testing
			return 123, nil
		},
	)
}

func (m *MockPseudonymDAO) CreatePseudonymWithIdentityMapping(ctx context.Context, userID int64, displayName string) (*models.Pseudonym, error) {
	args := m.Called(ctx, userID, displayName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Pseudonym), args.Error(1)
}

func (m *MockPseudonymDAO) GetPseudonymByID(ctx context.Context, pseudonymID string) (*models.Pseudonym, error) {
	args := m.Called(ctx, pseudonymID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string) (*models.Pseudonym, error)); ok {
		return fn(ctx, pseudonymID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Pseudonym), args.Error(1)
}

func (m *MockPseudonymDAO) GetPseudonymByDisplayName(ctx context.Context, displayName string) (*models.Pseudonym, error) {
	args := m.Called(ctx, displayName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Pseudonym), args.Error(1)
}

func (m *MockPseudonymDAO) GetPseudonymsByUserID(ctx context.Context, userID int64, roleName, scope string) ([]*models.Pseudonym, error) {
	args := m.Called(ctx, userID, roleName, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Pseudonym), args.Error(1)
}

func (m *MockPseudonymDAO) GetDefaultPseudonymByUserID(ctx context.Context, userID int64, roleName, scope string) (*models.Pseudonym, error) {
	args := m.Called(ctx, userID, roleName, scope)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Pseudonym), args.Error(1)
}

func (m *MockPseudonymDAO) UpdatePseudonym(ctx context.Context, pseudonymID string, updates *models.PseudonymSetter) error {
	args := m.Called(ctx, pseudonymID, updates)
	return args.Error(0)
}

func (m *MockPseudonymDAO) DeletePseudonym(ctx context.Context, pseudonymID string) error {
	args := m.Called(ctx, pseudonymID)
	return args.Error(0)
}

func (m *MockPseudonymDAO) VerifyPseudonymOwnership(ctx context.Context, pseudonymID string, userID int64, roleName, scope string) (bool, error) {
	args := m.Called(ctx, pseudonymID, userID, roleName, scope)
	return args.Bool(0), args.Error(1)
}

func (m *MockPseudonymDAO) GetUserIDByPseudonym(ctx context.Context, pseudonymID, roleName, scope string) (int64, error) {
	args := m.Called(ctx, pseudonymID, roleName, scope)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, string, string) (int64, error)); ok {
		return fn(ctx, pseudonymID, roleName, scope)
	}

	// Fallback to direct return values
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPseudonymDAO) UpdateLastActive(ctx context.Context, pseudonymID string) error {
	args := m.Called(ctx, pseudonymID)
	return args.Error(0)
}

func (m *MockPseudonymDAO) DeactivatePseudonym(ctx context.Context, pseudonymID string, userID int64, roleName, scope string) error {
	args := m.Called(ctx, pseudonymID, userID, roleName, scope)
	return args.Error(0)
}
