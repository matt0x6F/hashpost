package mocks

import (
	"context"
	"fmt"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockSubforumModeratorDAO is a mock implementation of SubforumModeratorDAOInterface with data injection support
type MockSubforumModeratorDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	moderatorsBySubforum  map[int32][]*models.SubforumModerator
	moderatorsByPseudonym map[string]*models.SubforumModerator // key: "pseudonymID-subforumID"
}

// NewMockSubforumModeratorDAO creates a new mock SubforumModeratorDAO with optional initial data
func NewMockSubforumModeratorDAO() *MockSubforumModeratorDAO {
	return &MockSubforumModeratorDAO{
		moderatorsBySubforum:  make(map[int32][]*models.SubforumModerator),
		moderatorsByPseudonym: make(map[string]*models.SubforumModerator),
	}
}

// InjectModeratorsBySubforum injects moderators for a subforum
func (m *MockSubforumModeratorDAO) InjectModeratorsBySubforum(subforumID int32, moderators []*models.SubforumModerator) {
	m.moderatorsBySubforum[subforumID] = moderators
}

// InjectModeratorByPseudonym injects a moderator for a specific pseudonym and subforum
func (m *MockSubforumModeratorDAO) InjectModeratorByPseudonym(pseudonymID string, subforumID int32, moderator *models.SubforumModerator) {
	key := fmt.Sprintf("%s-%d", pseudonymID, subforumID)
	m.moderatorsByPseudonym[key] = moderator
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockSubforumModeratorDAO) SetDefaultBehavior() {
	// Default behavior for GetModeratorsBySubforum
	m.On("GetModeratorsBySubforum", mock.Anything, mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, subforumID int32) ([]*models.SubforumModerator, error) {
			if moderators, exists := m.moderatorsBySubforum[subforumID]; exists {
				return moderators, nil
			}
			return []*models.SubforumModerator{}, nil
		},
	)

	// Default behavior for GetModeratorByPseudonym
	m.On("GetModeratorByPseudonym", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, pseudonymID string, subforumID int32) (*models.SubforumModerator, error) {
			key := fmt.Sprintf("%s-%d", pseudonymID, subforumID)
			if moderator, exists := m.moderatorsByPseudonym[key]; exists {
				return moderator, nil
			}
			return nil, nil
		},
	)

	// Default behavior for CreateModerator
	m.On("CreateModerator", mock.Anything, mock.AnythingOfType("int32"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, subforumID int32, pseudonymID, role string, addedByPseudonymID string) (*models.SubforumModerator, error) {
			moderator := &models.SubforumModerator{
				ModeratorID: 1,
				SubforumID:  subforumID,
				PseudonymID: pseudonymID,
				Role:        role,
			}
			return moderator, nil
		},
	)

	// Default behavior for DeleteModerator
	m.On("DeleteModerator", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, pseudonymID string, subforumID int32) error {
			return nil
		},
	)

	// Default behavior for UpdateModeratorRole
	m.On("UpdateModeratorRole", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int32"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, pseudonymID string, subforumID int32, newRole string) error {
			return nil
		},
	)
}

// GetModeratorsBySubforum retrieves all moderators for a subforum
func (m *MockSubforumModeratorDAO) GetModeratorsBySubforum(ctx context.Context, subforumID int32) ([]*models.SubforumModerator, error) {
	args := m.Called(ctx, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int32) ([]*models.SubforumModerator, error)); ok {
		return fn(ctx, subforumID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SubforumModerator), args.Error(1)
}

// GetModeratorByPseudonym retrieves a moderator by pseudonym ID and subforum ID
func (m *MockSubforumModeratorDAO) GetModeratorByPseudonym(ctx context.Context, pseudonymID string, subforumID int32) (*models.SubforumModerator, error) {
	args := m.Called(ctx, pseudonymID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, int32) (*models.SubforumModerator, error)); ok {
		return fn(ctx, pseudonymID, subforumID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SubforumModerator), args.Error(1)
}

// CreateModerator creates a new moderator
func (m *MockSubforumModeratorDAO) CreateModerator(ctx context.Context, subforumID int32, pseudonymID, role string, addedByPseudonymID string) (*models.SubforumModerator, error) {
	args := m.Called(ctx, subforumID, pseudonymID, role, addedByPseudonymID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int32, string, string, string) (*models.SubforumModerator, error)); ok {
		return fn(ctx, subforumID, pseudonymID, role, addedByPseudonymID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SubforumModerator), args.Error(1)
}

// DeleteModerator removes a moderator
func (m *MockSubforumModeratorDAO) DeleteModerator(ctx context.Context, pseudonymID string, subforumID int32) error {
	args := m.Called(ctx, pseudonymID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, int32) error); ok {
		return fn(ctx, pseudonymID, subforumID)
	}

	// Fallback to direct return values
	return args.Error(0)
}

// UpdateModeratorRole updates the role of a moderator
func (m *MockSubforumModeratorDAO) UpdateModeratorRole(ctx context.Context, pseudonymID string, subforumID int32, newRole string) error {
	args := m.Called(ctx, pseudonymID, subforumID, newRole)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, int32, string) error); ok {
		return fn(ctx, pseudonymID, subforumID, newRole)
	}

	// Fallback to direct return values
	return args.Error(0)
}
