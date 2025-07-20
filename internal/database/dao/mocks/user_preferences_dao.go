package mocks

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockUserPreferencesDAO is a mock implementation of UserPreferencesDAOInterface
type MockUserPreferencesDAO struct {
	mock.Mock
}

func (m *MockUserPreferencesDAO) CreateUserPreferences(ctx context.Context, userID int64, preferences *models.UserPreferenceSetter) (*models.UserPreference, error) {
	args := m.Called(ctx, userID, preferences)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserPreference), args.Error(1)
}

func (m *MockUserPreferencesDAO) GetUserPreferences(ctx context.Context, userID int64) (*models.UserPreference, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserPreference), args.Error(1)
}

func (m *MockUserPreferencesDAO) UpdateUserPreferences(ctx context.Context, userID int64, updates *models.UserPreferenceSetter) error {
	args := m.Called(ctx, userID, updates)
	return args.Error(0)
}

func (m *MockUserPreferencesDAO) UpsertUserPreferences(ctx context.Context, userID int64, preferences *models.UserPreferenceSetter) (*models.UserPreference, error) {
	args := m.Called(ctx, userID, preferences)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserPreference), args.Error(1)
}
