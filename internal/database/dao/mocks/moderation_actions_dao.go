package mocks

import (
	"context"
	"database/sql"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockModerationActionDAO is a mock implementation of ModerationActionDAOInterface with data injection support
type MockModerationActionDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	actions            map[int64]*models.ModerationAction
	actionsByType      map[string][]*models.ModerationAction
	actionsByModerator map[int64][]*models.ModerationAction
	counts             map[string]int64 // key: "actionType"
}

// NewMockModerationActionDAO creates a new mock ModerationActionDAO with optional initial data
func NewMockModerationActionDAO() *MockModerationActionDAO {
	return &MockModerationActionDAO{
		actions:            make(map[int64]*models.ModerationAction),
		actionsByType:      make(map[string][]*models.ModerationAction),
		actionsByModerator: make(map[int64][]*models.ModerationAction),
		counts:             make(map[string]int64),
	}
}

// InjectAction injects a moderation action into the mock for testing
func (m *MockModerationActionDAO) InjectAction(action *models.ModerationAction) {
	m.actions[action.ActionID] = action
}

// InjectActionsByType injects actions that should be returned when querying by action type
func (m *MockModerationActionDAO) InjectActionsByType(actionType string, actions []*models.ModerationAction) {
	m.actionsByType[actionType] = actions
}

// InjectActionsByModerator injects actions that should be returned when querying by moderator
func (m *MockModerationActionDAO) InjectActionsByModerator(moderatorUserID int64, actions []*models.ModerationAction) {
	m.actionsByModerator[moderatorUserID] = actions
}

// InjectCount injects a count that should be returned for count operations
func (m *MockModerationActionDAO) InjectCount(actionType string, count int64) {
	m.counts[actionType] = count
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockModerationActionDAO) SetDefaultBehavior() {
	// Default behavior for CreateModerationAction
	m.On("CreateModerationAction", mock.Anything, mock.AnythingOfType("*models.ModerationActionSetter")).Return(
		func(ctx context.Context, actionSetter *models.ModerationActionSetter) (*models.ModerationAction, error) {
			// Create a mock action with the provided data
			action := &models.ModerationAction{
				ActionID:             456, // Mock ID
				ModeratorUserID:      *actionSetter.ModeratorUserID,
				ModeratorPseudonymID: *actionSetter.ModeratorPseudonymID,
				ActionType:           *actionSetter.ActionType,
				CreatedAt:            *actionSetter.CreatedAt,
			}

			// Set optional fields if provided
			if actionSetter.TargetContentType != nil {
				action.TargetContentType = *actionSetter.TargetContentType
			}
			if actionSetter.TargetContentID != nil {
				action.TargetContentID = *actionSetter.TargetContentID
			}
			if actionSetter.ActionDetails != nil {
				action.ActionDetails = *actionSetter.ActionDetails
			}

			return action, nil
		},
	)

	// Default behavior for GetModerationActionByID
	m.On("GetModerationActionByID", mock.Anything, mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, actionID int64) (*models.ModerationAction, error) {
			if action, exists := m.actions[actionID]; exists {
				return action, nil
			}
			return nil, sql.ErrNoRows
		},
	)

	// Default behavior for GetModerationActions
	m.On("GetModerationActions", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(
		func(ctx context.Context, actionType string, page, limit int) ([]*models.ModerationAction, error) {
			if actions, exists := m.actionsByType[actionType]; exists {
				return actions, nil
			}
			return []*models.ModerationAction{}, nil
		},
	)

	// Default behavior for GetModerationActionsByModerator
	m.On("GetModerationActionsByModerator", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(
		func(ctx context.Context, moderatorUserID int64, page, limit int) ([]*models.ModerationAction, error) {
			if actions, exists := m.actionsByModerator[moderatorUserID]; exists {
				return actions, nil
			}
			return []*models.ModerationAction{}, nil
		},
	)

	// Default behavior for CountModerationActions
	m.On("CountModerationActions", mock.Anything, mock.AnythingOfType("string")).Return(
		func(ctx context.Context, actionType string) (int64, error) {
			count := m.counts[actionType]
			return count, nil
		},
	)
}

// CreateModerationAction creates a new moderation action
func (m *MockModerationActionDAO) CreateModerationAction(ctx context.Context, action *models.ModerationActionSetter) (*models.ModerationAction, error) {
	args := m.Called(ctx, action)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, *models.ModerationActionSetter) (*models.ModerationAction, error)); ok {
		return fn(ctx, action)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ModerationAction), args.Error(1)
}

// GetModerationActionByID retrieves a moderation action by ID
func (m *MockModerationActionDAO) GetModerationActionByID(ctx context.Context, actionID int64) (*models.ModerationAction, error) {
	args := m.Called(ctx, actionID)
	if args.Get(0) == nil {
		return nil, sql.ErrNoRows
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64) (*models.ModerationAction, error)); ok {
		return fn(ctx, actionID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.ModerationAction), args.Error(1)
}

// GetModerationActions retrieves moderation actions with filtering and pagination
func (m *MockModerationActionDAO) GetModerationActions(ctx context.Context, actionType string, page, limit int) ([]*models.ModerationAction, error) {
	args := m.Called(ctx, actionType, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, int, int) ([]*models.ModerationAction, error)); ok {
		return fn(ctx, actionType, page, limit)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ModerationAction), args.Error(1)
}

// CountModerationActions counts moderation actions with optional action type filter
func (m *MockModerationActionDAO) CountModerationActions(ctx context.Context, actionType string) (int64, error) {
	args := m.Called(ctx, actionType)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string) (int64, error)); ok {
		return fn(ctx, actionType)
	}

	// Fallback to direct return values
	return args.Get(0).(int64), args.Error(1)
}

// GetModerationActionsByModerator retrieves moderation actions by a specific moderator
func (m *MockModerationActionDAO) GetModerationActionsByModerator(ctx context.Context, moderatorUserID int64, page, limit int) ([]*models.ModerationAction, error) {
	args := m.Called(ctx, moderatorUserID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, int, int) ([]*models.ModerationAction, error)); ok {
		return fn(ctx, moderatorUserID, page, limit)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ModerationAction), args.Error(1)
}
