package mocks

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockUserBanDAO is a mock implementation of UserBanDAOInterface with data injection support
type MockUserBanDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	bans           map[int64]*models.UserBan
	bansBySubforum map[int32][]*models.UserBan
	bansByUser     map[int64][]*models.UserBan
	counts         map[string]int64 // key: "subforumID" or "userID"
	bannedUsers    map[string]bool  // key: "userID-subforumID"
}

// NewMockUserBanDAO creates a new mock UserBanDAO with optional initial data
func NewMockUserBanDAO() *MockUserBanDAO {
	return &MockUserBanDAO{
		bans:           make(map[int64]*models.UserBan),
		bansBySubforum: make(map[int32][]*models.UserBan),
		bansByUser:     make(map[int64][]*models.UserBan),
		counts:         make(map[string]int64),
		bannedUsers:    make(map[string]bool),
	}
}

// InjectBan injects a user ban into the mock for testing
func (m *MockUserBanDAO) InjectBan(ban *models.UserBan) {
	m.bans[ban.BanID] = ban
}

// InjectBansBySubforum injects bans that should be returned when querying by subforum
func (m *MockUserBanDAO) InjectBansBySubforum(subforumID int32, bans []*models.UserBan) {
	m.bansBySubforum[subforumID] = bans
}

// InjectBansByUser injects bans that should be returned when querying by user
func (m *MockUserBanDAO) InjectBansByUser(userID int64, bans []*models.UserBan) {
	m.bansByUser[userID] = bans
}

// InjectCount injects a count that should be returned for count operations
func (m *MockUserBanDAO) InjectCount(key string, count int64) {
	m.counts[key] = count
}

// InjectBannedUser injects a banned user status for testing
func (m *MockUserBanDAO) InjectBannedUser(userID int64, subforumID int32, isBanned bool) {
	key := fmt.Sprintf("%d-%d", userID, subforumID)
	m.bannedUsers[key] = isBanned
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockUserBanDAO) SetDefaultBehavior() {
	// Default behavior for GetUserBanByID
	m.On("GetUserBanByID", mock.Anything, mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, banID int64) (*models.UserBan, error) {
			if ban, exists := m.bans[banID]; exists {
				return ban, nil
			}
			return nil, sql.ErrNoRows
		},
	)

	// Default behavior for GetUserBansBySubforum
	m.On("GetUserBansBySubforum", mock.Anything, mock.AnythingOfType("int32"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(
		func(ctx context.Context, subforumID int32, page, limit int) ([]*models.UserBan, error) {
			if bans, exists := m.bansBySubforum[subforumID]; exists {
				return bans, nil
			}
			return []*models.UserBan{}, nil
		},
	)

	// Default behavior for GetUserBansByUser
	m.On("GetUserBansByUser", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(
		func(ctx context.Context, userID int64, page, limit int) ([]*models.UserBan, error) {
			if bans, exists := m.bansByUser[userID]; exists {
				return bans, nil
			}
			return []*models.UserBan{}, nil
		},
	)

	// Default behavior for CountUserBansBySubforum
	m.On("CountUserBansBySubforum", mock.Anything, mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, subforumID int32) (int64, error) {
			count := m.counts[fmt.Sprintf("subforum_%d", subforumID)]
			return count, nil
		},
	)

	// Default behavior for CountUserBansByUser
	m.On("CountUserBansByUser", mock.Anything, mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, userID int64) (int64, error) {
			count := m.counts[fmt.Sprintf("user_%d", userID)]
			return count, nil
		},
	)

	// Default behavior for CreateUserBan
	m.On("CreateUserBan", mock.Anything, mock.AnythingOfType("*models.UserBanSetter")).Return(
		func(ctx context.Context, banSetter *models.UserBanSetter) (*models.UserBan, error) {
			// Create a mock ban with the provided data
			ban := &models.UserBan{
				BanID:               789, // Mock ID
				SubforumID:          *banSetter.SubforumID,
				BannedUserID:        *banSetter.BannedUserID,
				BannedByUserID:      *banSetter.BannedByUserID,
				BannedByPseudonymID: *banSetter.BannedByPseudonymID,
				BanReason:           *banSetter.BanReason,
				CreatedAt:           *banSetter.CreatedAt,
			}

			// Set optional fields if provided
			if banSetter.IsPermanent != nil {
				ban.IsPermanent = *banSetter.IsPermanent
			}
			if banSetter.ExpiresAt != nil {
				ban.ExpiresAt = *banSetter.ExpiresAt
			}
			if banSetter.IsActive != nil {
				ban.IsActive = *banSetter.IsActive
			}

			return ban, nil
		},
	)

	// Default behavior for IsUserBannedFromSubforum
	m.On("IsUserBannedFromSubforum", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, userID int64, subforumID int32) (bool, error) {
			key := fmt.Sprintf("%d-%d", userID, subforumID)
			if isBanned, exists := m.bannedUsers[key]; exists {
				return isBanned, nil
			}
			return false, nil
		},
	)
}

// CreateUserBan creates a new user ban
func (m *MockUserBanDAO) CreateUserBan(ctx context.Context, ban *models.UserBanSetter) (*models.UserBan, error) {
	args := m.Called(ctx, ban)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, *models.UserBanSetter) (*models.UserBan, error)); ok {
		return fn(ctx, ban)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserBan), args.Error(1)
}

// GetUserBanByID retrieves a user ban by ID
func (m *MockUserBanDAO) GetUserBanByID(ctx context.Context, banID int64) (*models.UserBan, error) {
	args := m.Called(ctx, banID)
	if args.Get(0) == nil {
		return nil, sql.ErrNoRows
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64) (*models.UserBan, error)); ok {
		return fn(ctx, banID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserBan), args.Error(1)
}

// GetUserBansBySubforum retrieves user bans for a specific subforum
func (m *MockUserBanDAO) GetUserBansBySubforum(ctx context.Context, subforumID int32, page, limit int) ([]*models.UserBan, error) {
	args := m.Called(ctx, subforumID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int32, int, int) ([]*models.UserBan, error)); ok {
		return fn(ctx, subforumID, page, limit)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserBan), args.Error(1)
}

// GetUserBansByUser retrieves user bans for a specific user
func (m *MockUserBanDAO) GetUserBansByUser(ctx context.Context, bannedUserID int64, page, limit int) ([]*models.UserBan, error) {
	args := m.Called(ctx, bannedUserID, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, int, int) ([]*models.UserBan, error)); ok {
		return fn(ctx, bannedUserID, page, limit)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.UserBan), args.Error(1)
}

// CountUserBansBySubforum counts user bans for a specific subforum
func (m *MockUserBanDAO) CountUserBansBySubforum(ctx context.Context, subforumID int32) (int64, error) {
	args := m.Called(ctx, subforumID)
	return args.Get(0).(int64), args.Error(1)
}

// CountUserBansByUser counts user bans for a specific user
func (m *MockUserBanDAO) CountUserBansByUser(ctx context.Context, bannedUserID int64) (int64, error) {
	args := m.Called(ctx, bannedUserID)
	return args.Get(0).(int64), args.Error(1)
}

// IsUserBannedFromSubforum checks if a user is banned from a specific subforum
func (m *MockUserBanDAO) IsUserBannedFromSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	args := m.Called(ctx, userID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, int32) (bool, error)); ok {
		return fn(ctx, userID, subforumID)
	}

	// Fallback to direct return values
	return args.Get(0).(bool), args.Error(1)
}

// UpdateUserBan updates a user ban
func (m *MockUserBanDAO) UpdateUserBan(ctx context.Context, banID int64, updates *models.UserBanSetter) error {
	args := m.Called(ctx, banID, updates)
	return args.Error(0)
}

// DeactivateUserBan deactivates a user ban
func (m *MockUserBanDAO) DeactivateUserBan(ctx context.Context, banID int64) error {
	args := m.Called(ctx, banID)
	return args.Error(0)
}
