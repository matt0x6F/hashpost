package mocks

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/mock"
)

// MockPermissionDAO is a mock implementation of PermissionDAOInterface with data injection support
type MockPermissionDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	accessPermissions map[string]bool // key: "userID-subforumID"
	capabilities      map[string]bool // key: "userID-capability"
	roles             map[string]bool // key: "userID-role"
}

// NewMockPermissionDAO creates a new mock PermissionDAO with optional initial data
func NewMockPermissionDAO() *MockPermissionDAO {
	return &MockPermissionDAO{
		accessPermissions: make(map[string]bool),
		capabilities:      make(map[string]bool),
		roles:             make(map[string]bool),
	}
}

// InjectAccessPermission injects an access permission for testing
func (m *MockPermissionDAO) InjectAccessPermission(userID int64, subforumID int32, canAccess bool) {
	key := fmt.Sprintf("%d-%d", userID, subforumID)
	m.accessPermissions[key] = canAccess
}

// InjectSubforumCapability injects a subforum capability for testing
func (m *MockPermissionDAO) InjectSubforumCapability(userID int64, subforumID int32, capability string, hasCapability bool) {
	key := fmt.Sprintf("%d-%d-%s", userID, subforumID, capability)
	m.capabilities[key] = hasCapability
}

// InjectModerationAbility injects moderation ability for testing
func (m *MockPermissionDAO) InjectModerationAbility(userID int64, subforumID int32, canModerate bool) {
	key := fmt.Sprintf("%d-%d", userID, subforumID)
	m.roles[key] = canModerate
}

// InjectSubforumCapabilityWithActivePseudonym injects a capability for a user, subforum, and active pseudonym
func (m *MockPermissionDAO) InjectSubforumCapabilityWithActivePseudonym(userID int64, subforumID int32, capability string, activePseudonymID string, hasCapability bool) {
	key := fmt.Sprintf("%d-%d-%s-%s", userID, subforumID, capability, activePseudonymID)
	m.capabilities[key] = hasCapability
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockPermissionDAO) SetDefaultBehavior() {
	// Default behavior for CanAccessPrivateSubforum
	m.On("CanAccessPrivateSubforum", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, userID int64, subforumID int32) (bool, error) {
			key := fmt.Sprintf("%d-%d", userID, subforumID)
			if canAccess, exists := m.accessPermissions[key]; exists {
				return canAccess, nil
			}
			return false, nil
		},
	)

	// Default behavior for HasSubforumCapability
	m.On("HasSubforumCapability", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int32"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
			key := fmt.Sprintf("%d-%d-%s", userID, subforumID, capability)
			if hasCapability, exists := m.capabilities[key]; exists {
				return hasCapability, nil
			}
			return false, nil
		},
	)

	// Default behavior for CanModerateSubforum
	m.On("CanModerateSubforum", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, userID int64, subforumID int32) (bool, error) {
			key := fmt.Sprintf("%d-%d", userID, subforumID)
			if canModerate, exists := m.roles[key]; exists {
				return canModerate, nil
			}
			return false, nil
		},
	)

	// Default behavior for GetUserSubforumCapabilities
	m.On("GetUserSubforumCapabilities", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
			return []string{}, nil
		},
	)

	// Default behavior for HasSubforumCapabilityWithActivePseudonym
	m.On("HasSubforumCapabilityWithActivePseudonym", mock.Anything, mock.AnythingOfType("int64"), mock.AnythingOfType("int32"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(
		func(ctx context.Context, userID int64, subforumID int32, capability string, activePseudonymID string) (bool, error) {
			key := fmt.Sprintf("%d-%d-%s-%s", userID, subforumID, capability, activePseudonymID)
			if hasCapability, exists := m.capabilities[key]; exists {
				return hasCapability, nil
			}
			return false, nil
		},
	)
}

// CanAccessPrivateSubforum checks if a user can access a private subforum
func (m *MockPermissionDAO) CanAccessPrivateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	args := m.Called(ctx, userID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, int32) (bool, error)); ok {
		return fn(ctx, userID, subforumID)
	}

	// Fallback to direct return values
	return args.Get(0).(bool), args.Error(1)
}

// HasSubforumCapability checks if a user has a specific capability for a subforum
func (m *MockPermissionDAO) HasSubforumCapability(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
	args := m.Called(ctx, userID, subforumID, capability)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, int32, string) (bool, error)); ok {
		return fn(ctx, userID, subforumID, capability)
	}

	// Fallback to direct return values
	return args.Get(0).(bool), args.Error(1)
}

// HasSubforumCapabilityWithActivePseudonym checks if a user has a specific capability for a subforum using only the active pseudonym
func (m *MockPermissionDAO) HasSubforumCapabilityWithActivePseudonym(ctx context.Context, userID int64, subforumID int32, capability string, activePseudonymID string) (bool, error) {
	args := m.Called(ctx, userID, subforumID, capability, activePseudonymID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, int32, string, string) (bool, error)); ok {
		return fn(ctx, userID, subforumID, capability, activePseudonymID)
	}

	// Fallback to direct return values
	return args.Get(0).(bool), args.Error(1)
}

// CanAccessPrivateSubforumWithActivePseudonym checks if a user can access a private subforum using only the active pseudonym
func (m *MockPermissionDAO) CanAccessPrivateSubforumWithActivePseudonym(ctx context.Context, userID int64, subforumID int32, activePseudonymID string) (bool, error) {
	args := m.Called(ctx, userID, subforumID, activePseudonymID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, int32, string) (bool, error)); ok {
		return fn(ctx, userID, subforumID, activePseudonymID)
	}

	// Fallback to direct return values
	return args.Get(0).(bool), args.Error(1)
}

// CanModerateSubforum checks if a user can moderate a subforum
func (m *MockPermissionDAO) CanModerateSubforum(ctx context.Context, userID int64, subforumID int32) (bool, error) {
	args := m.Called(ctx, userID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, int32) (bool, error)); ok {
		return fn(ctx, userID, subforumID)
	}

	// Fallback to direct return values
	return args.Get(0).(bool), args.Error(1)
}

// GetUserSubforumRoles gets the roles a user has for a specific subforum
func (m *MockPermissionDAO) GetUserSubforumRoles(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
	args := m.Called(ctx, userID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, int32) ([]string, error)); ok {
		return fn(ctx, userID, subforumID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return []string{}, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// GetUserSubforumCapabilities gets the capabilities a user has for a specific subforum
func (m *MockPermissionDAO) GetUserSubforumCapabilities(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
	args := m.Called(ctx, userID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, int32) ([]string, error)); ok {
		return fn(ctx, userID, subforumID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return []string{}, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// GetActivePseudonymRolesAndCapabilities gets the roles and capabilities for a specific active pseudonym
func (m *MockPermissionDAO) GetActivePseudonymRolesAndCapabilities(ctx context.Context, userID int64, activePseudonymID string) ([]string, []string, error) {
	args := m.Called(ctx, userID, activePseudonymID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, string) ([]string, []string, error)); ok {
		return fn(ctx, userID, activePseudonymID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return []string{}, []string{}, args.Error(2)
	}
	return args.Get(0).([]string), args.Get(1).([]string), args.Error(2)
}

// GetUnifiedActivePseudonymRolesAndCapabilities gets the unified roles and capabilities for a specific active pseudonym
func (m *MockPermissionDAO) GetUnifiedActivePseudonymRolesAndCapabilities(ctx context.Context, userID int64, activePseudonymID string, subforumID *int32) ([]string, []string, error) {
	args := m.Called(ctx, userID, activePseudonymID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, string, *int32) ([]string, []string, error)); ok {
		return fn(ctx, userID, activePseudonymID, subforumID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return []string{}, []string{}, args.Error(2)
	}
	return args.Get(0).([]string), args.Get(1).([]string), args.Error(2)
}

// HasUnifiedCapability checks if a user has a unified capability
func (m *MockPermissionDAO) HasUnifiedCapability(ctx context.Context, userID int64, activePseudonymID string, capability string, subforumID *int32) (bool, error) {
	args := m.Called(ctx, userID, activePseudonymID, capability, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64, string, string, *int32) (bool, error)); ok {
		return fn(ctx, userID, activePseudonymID, capability, subforumID)
	}

	// Fallback to direct return values
	return args.Get(0).(bool), args.Error(1)
}
