package mocks

import (
	"context"
	"database/sql"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockSubforumDAO is a mock implementation of SubforumDAOInterface with data injection support
type MockSubforumDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	subforums       map[int32]*models.Subforum
	subforumsByName map[string]*models.Subforum
	allSubforums    []*models.Subforum
}

// NewMockSubforumDAO creates a new mock SubforumDAO with optional initial data
func NewMockSubforumDAO() *MockSubforumDAO {
	return &MockSubforumDAO{
		subforums:       make(map[int32]*models.Subforum),
		subforumsByName: make(map[string]*models.Subforum),
		allSubforums:    make([]*models.Subforum, 0),
	}
}

// InjectSubforum injects a subforum into the mock for testing
func (m *MockSubforumDAO) InjectSubforum(subforum *models.Subforum) {
	m.subforums[subforum.SubforumID] = subforum
	m.subforumsByName[subforum.Name] = subforum
}

// InjectSubforumByName injects a subforum that should be returned when querying by name
func (m *MockSubforumDAO) InjectSubforumByName(name string, subforum *models.Subforum) {
	m.subforumsByName[name] = subforum
}

// InjectSubforumByCommunityTypeAndName injects a subforum that should be returned when querying by community type and name
func (m *MockSubforumDAO) InjectSubforumByCommunityTypeAndName(communityType, name string, subforum *models.Subforum) {
	key := communityType + "/" + name
	m.subforumsByName[key] = subforum
}

// InjectAllSubforums injects all subforums that should be returned when listing
func (m *MockSubforumDAO) InjectAllSubforums(subforums []*models.Subforum) {
	m.allSubforums = subforums
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockSubforumDAO) SetDefaultBehavior() {
	// Default behavior for GetSubforumByID
	m.On("GetSubforumByID", mock.Anything, mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, subforumID int32) (*models.Subforum, error) {
			if subforum, exists := m.subforums[subforumID]; exists {
				return subforum, nil
			}
			return nil, sql.ErrNoRows
		},
	)

	// Default behavior for GetSubforumByName
	m.On("GetSubforumByName", mock.Anything, mock.AnythingOfType("string")).Return(
		func(ctx context.Context, name string) (*models.Subforum, error) {
			if subforum, exists := m.subforumsByName[name]; exists {
				return subforum, nil
			}
			return nil, sql.ErrNoRows
		},
	)

	// Default behavior for GetSubforumByCommunityTypeAndName
	m.On("GetSubforumByCommunityTypeAndName", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, communityType, name string) (*models.Subforum, error) {
			key := communityType + "/" + name
			if subforum, exists := m.subforumsByName[key]; exists {
				return subforum, nil
			}
			return nil, sql.ErrNoRows
		},
	)

	// Default behavior for ListSubforums
	m.On("ListSubforums", mock.Anything).Return(
		func(ctx context.Context) ([]*models.Subforum, error) {
			return m.allSubforums, nil
		},
	)

	// Default behavior for CreateSubforum
	m.On("CreateSubforum", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("bool"), mock.AnythingOfType("bool"), mock.AnythingOfType("bool"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, name, displayName, description, sidebarText, rulesText, communityType, governanceStyle string, isNSFW, isPrivate, isRestricted bool, ownerPseudonymID string) (*models.Subforum, error) {
			// Create a mock subforum with the provided data
			subforum := &models.Subforum{
				SubforumID:       123, // Mock ID
				Name:             name,
				DisplayName:      displayName,
				Description:      sql.Null[string]{V: description, Valid: true},
				SidebarText:      sql.Null[string]{V: sidebarText, Valid: true},
				RulesText:        sql.Null[string]{V: rulesText, Valid: true},
				CommunityType:    communityType,
				GovernanceStyle:  governanceStyle,
				IsNSFW:           sql.Null[bool]{V: isNSFW, Valid: true},
				IsPrivate:        sql.Null[bool]{V: isPrivate, Valid: true},
				IsRestricted:     sql.Null[bool]{V: isRestricted, Valid: true},
				OwnerPseudonymID: sql.Null[string]{V: ownerPseudonymID, Valid: ownerPseudonymID != ""},
			}

			// Store the subforum
			m.subforums[subforum.SubforumID] = subforum
			m.subforumsByName[subforum.Name] = subforum

			return subforum, nil
		},
	)

	// Default behavior for UpdatePostCount
	m.On("UpdatePostCount", mock.Anything, mock.AnythingOfType("int32"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, subforumID int32, postCount int32) error {
			return nil
		},
	)
}

// CreateSubforum creates a new subforum
func (m *MockSubforumDAO) CreateSubforum(ctx context.Context, name, displayName, description, sidebarText, rulesText, communityType, governanceStyle string, isNSFW, isPrivate, isRestricted bool, ownerPseudonymID string) (*models.Subforum, error) {
	args := m.Called(ctx, name, displayName, description, sidebarText, rulesText, communityType, governanceStyle, isNSFW, isPrivate, isRestricted, ownerPseudonymID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, string, string, string, string, string, string, bool, bool, bool, string) (*models.Subforum, error)); ok {
		return fn(ctx, name, displayName, description, sidebarText, rulesText, communityType, governanceStyle, isNSFW, isPrivate, isRestricted, ownerPseudonymID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subforum), args.Error(1)
}

// GetSubforumByID retrieves a subforum by ID
func (m *MockSubforumDAO) GetSubforumByID(ctx context.Context, subforumID int32) (*models.Subforum, error) {
	args := m.Called(ctx, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int32) (*models.Subforum, error)); ok {
		return fn(ctx, subforumID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subforum), args.Error(1)
}

// GetSubforumByName retrieves a subforum by name
func (m *MockSubforumDAO) GetSubforumByName(ctx context.Context, name string) (*models.Subforum, error) {
	args := m.Called(ctx, name)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string) (*models.Subforum, error)); ok {
		return fn(ctx, name)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subforum), args.Error(1)
}

// ListSubforums retrieves all subforums
func (m *MockSubforumDAO) ListSubforums(ctx context.Context) ([]*models.Subforum, error) {
	args := m.Called(ctx)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context) ([]*models.Subforum, error)); ok {
		return fn(ctx)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Subforum), args.Error(1)
}

// UpdatePostCount updates the post count for a subforum
func (m *MockSubforumDAO) UpdatePostCount(ctx context.Context, subforumID int32, postCount int32) error {
	args := m.Called(ctx, subforumID, postCount)
	return args.Error(0)
}

// GetSubforumByCommunityTypeAndName retrieves a subforum by community type and name
func (m *MockSubforumDAO) GetSubforumByCommunityTypeAndName(ctx context.Context, communityType, name string) (*models.Subforum, error) {
	args := m.Called(ctx, communityType, name)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, string) (*models.Subforum, error)); ok {
		return fn(ctx, communityType, name)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subforum), args.Error(1)
}

// ListSubforumsByCommunityType retrieves subforums by community type
func (m *MockSubforumDAO) ListSubforumsByCommunityType(ctx context.Context, communityType string) ([]*models.Subforum, error) {
	args := m.Called(ctx, communityType)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string) ([]*models.Subforum, error)); ok {
		return fn(ctx, communityType)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Subforum), args.Error(1)
}

// UpdateSubscriberCount updates the subscriber count for a subforum
func (m *MockSubforumDAO) UpdateSubscriberCount(ctx context.Context, subforumID int32, subscriberCount int32) error {
	args := m.Called(ctx, subforumID, subscriberCount)
	return args.Error(0)
}

// UpdateSettings mocks updating subforum settings
func (m *MockSubforumDAO) UpdateSettings(ctx context.Context, subforumID int32, allowImages, allowVideos, allowPolls, requireFlair, isPrivate, isRestricted, isNSFW bool, minimumAccountAgeHours, minimumKarmaRequired int, description, sidebarText string) error {
	args := m.Called(ctx, subforumID, allowImages, allowVideos, allowPolls, requireFlair, isPrivate, isRestricted, isNSFW, minimumAccountAgeHours, minimumKarmaRequired, description, sidebarText)
	return args.Error(0)
}

// UpdateRules mocks updating subforum rules
func (m *MockSubforumDAO) UpdateRules(ctx context.Context, subforumID int32, rules []byte) error {
	args := m.Called(ctx, subforumID, rules)
	return args.Error(0)
}
