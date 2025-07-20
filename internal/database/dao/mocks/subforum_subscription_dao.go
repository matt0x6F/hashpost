package mocks

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockSubforumSubscriptionDAO is a mock implementation of SubforumSubscriptionDAOInterface with data injection support
type MockSubforumSubscriptionDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	subscriptions            map[string]*models.SubforumSubscription // key: "pseudonymID-subforumID"
	subscriptionsByPseudonym map[string][]*models.SubforumSubscription
	subscriptionsBySubforum  map[int32][]*models.SubforumSubscription
	counts                   map[string]int64 // key: "subforumID" or "pseudonymID"
	subscribedStatus         map[string]bool  // key: "pseudonymID-subforumID"
	favoriteStatus           map[string]bool  // key: "pseudonymID-subforumID"
}

// NewMockSubforumSubscriptionDAO creates a new mock SubforumSubscriptionDAO with optional initial data
func NewMockSubforumSubscriptionDAO() *MockSubforumSubscriptionDAO {
	return &MockSubforumSubscriptionDAO{
		subscriptions:            make(map[string]*models.SubforumSubscription),
		subscriptionsByPseudonym: make(map[string][]*models.SubforumSubscription),
		subscriptionsBySubforum:  make(map[int32][]*models.SubforumSubscription),
		counts:                   make(map[string]int64),
		subscribedStatus:         make(map[string]bool),
		favoriteStatus:           make(map[string]bool),
	}
}

// InjectSubscription injects a subscription into the mock for testing
func (m *MockSubforumSubscriptionDAO) InjectSubscription(subscription *models.SubforumSubscription) {
	key := fmt.Sprintf("%s-%d", subscription.PseudonymID, subscription.SubforumID)
	m.subscriptions[key] = subscription
}

// InjectSubscriptionsByPseudonym injects subscriptions that should be returned when querying by pseudonym
func (m *MockSubforumSubscriptionDAO) InjectSubscriptionsByPseudonym(pseudonymID string, subscriptions []*models.SubforumSubscription) {
	m.subscriptionsByPseudonym[pseudonymID] = subscriptions
}

// InjectSubscriptionsBySubforum injects subscriptions that should be returned when querying by subforum
func (m *MockSubforumSubscriptionDAO) InjectSubscriptionsBySubforum(subforumID int32, subscriptions []*models.SubforumSubscription) {
	m.subscriptionsBySubforum[subforumID] = subscriptions
}

// InjectCount injects a count that should be returned for count operations
func (m *MockSubforumSubscriptionDAO) InjectCount(key string, count int64) {
	m.counts[key] = count
}

// InjectSubscribedStatus injects a subscribed status for testing
func (m *MockSubforumSubscriptionDAO) InjectSubscribedStatus(pseudonymID string, subforumID int32, isSubscribed bool) {
	key := fmt.Sprintf("%s-%d", pseudonymID, subforumID)
	m.subscribedStatus[key] = isSubscribed
}

// InjectFavoriteStatus injects a favorite status for testing
func (m *MockSubforumSubscriptionDAO) InjectFavoriteStatus(pseudonymID string, subforumID int32, isFavorite bool) {
	key := fmt.Sprintf("%s-%d", pseudonymID, subforumID)
	m.favoriteStatus[key] = isFavorite
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockSubforumSubscriptionDAO) SetDefaultBehavior() {
	// Default behavior for CreateSubscription
	m.On("CreateSubscription", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int32"), mock.AnythingOfType("bool")).Return(
		func(ctx context.Context, pseudonymID string, subforumID int32, isFavorite bool) (*models.SubforumSubscription, error) {
			// Create a mock subscription with the provided data
			subscription := &models.SubforumSubscription{
				SubscriptionID: 456, // Mock ID
				PseudonymID:    pseudonymID,
				SubforumID:     subforumID,
				IsFavorite:     sql.Null[bool]{V: isFavorite, Valid: true},
				SubscribedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
			}

			// Store the subscription
			key := fmt.Sprintf("%s-%d", pseudonymID, subforumID)
			m.subscriptions[key] = subscription

			return subscription, nil
		},
	)

	// Default behavior for GetSubscription
	m.On("GetSubscription", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, pseudonymID string, subforumID int32) (*models.SubforumSubscription, error) {
			key := fmt.Sprintf("%s-%d", pseudonymID, subforumID)
			if subscription, exists := m.subscriptions[key]; exists {
				return subscription, nil
			}
			return nil, sql.ErrNoRows
		},
	)

	// Default behavior for GetSubscriptionsByPseudonym
	m.On("GetSubscriptionsByPseudonym", mock.Anything, mock.AnythingOfType("string")).Return(
		func(ctx context.Context, pseudonymID string) ([]*models.SubforumSubscription, error) {
			if subscriptions, exists := m.subscriptionsByPseudonym[pseudonymID]; exists {
				return subscriptions, nil
			}
			return []*models.SubforumSubscription{}, nil
		},
	)

	// Default behavior for IsSubscribed
	m.On("IsSubscribed", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, pseudonymID string, subforumID int32) (bool, error) {
			key := fmt.Sprintf("%s-%d", pseudonymID, subforumID)
			if isSubscribed, exists := m.subscribedStatus[key]; exists {
				return isSubscribed, nil
			}
			return false, nil
		},
	)

	// Default behavior for IsFavorite
	m.On("IsFavorite", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, pseudonymID string, subforumID int32) (bool, error) {
			key := fmt.Sprintf("%s-%d", pseudonymID, subforumID)
			if isFavorite, exists := m.favoriteStatus[key]; exists {
				return isFavorite, nil
			}
			return false, nil
		},
	)

	// Default behavior for DeleteSubscription
	m.On("DeleteSubscription", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, pseudonymID string, subforumID int32) error {
			key := fmt.Sprintf("%s-%d", pseudonymID, subforumID)
			if _, exists := m.subscriptions[key]; exists {
				delete(m.subscriptions, key)
				return nil
			}
			return sql.ErrNoRows
		},
	)

	// Default behavior for CountSubscriptionsBySubforum
	m.On("CountSubscriptionsBySubforum", mock.Anything, mock.AnythingOfType("int32")).Return(
		func(ctx context.Context, subforumID int32) (int64, error) {
			count := m.counts[fmt.Sprintf("subforum_%d", subforumID)]
			return count, nil
		},
	)

}

// CreateSubscription creates a new subscription
func (m *MockSubforumSubscriptionDAO) CreateSubscription(ctx context.Context, pseudonymID string, subforumID int32, isFavorite bool) (*models.SubforumSubscription, error) {
	args := m.Called(ctx, pseudonymID, subforumID, isFavorite)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, int32, bool) (*models.SubforumSubscription, error)); ok {
		return fn(ctx, pseudonymID, subforumID, isFavorite)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SubforumSubscription), args.Error(1)
}

// GetSubscription retrieves a subscription
func (m *MockSubforumSubscriptionDAO) GetSubscription(ctx context.Context, pseudonymID string, subforumID int32) (*models.SubforumSubscription, error) {
	args := m.Called(ctx, pseudonymID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, int32) (*models.SubforumSubscription, error)); ok {
		return fn(ctx, pseudonymID, subforumID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.SubforumSubscription), args.Error(1)
}

// GetSubscriptionsByPseudonym retrieves subscriptions by pseudonym
func (m *MockSubforumSubscriptionDAO) GetSubscriptionsByPseudonym(ctx context.Context, pseudonymID string) ([]*models.SubforumSubscription, error) {
	args := m.Called(ctx, pseudonymID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string) ([]*models.SubforumSubscription, error)); ok {
		return fn(ctx, pseudonymID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.SubforumSubscription), args.Error(1)
}

// IsSubscribed checks if a pseudonym is subscribed to a subforum
func (m *MockSubforumSubscriptionDAO) IsSubscribed(ctx context.Context, pseudonymID string, subforumID int32) (bool, error) {
	args := m.Called(ctx, pseudonymID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, int32) (bool, error)); ok {
		return fn(ctx, pseudonymID, subforumID)
	}

	// Fallback to direct return values
	return args.Get(0).(bool), args.Error(1)
}

// IsFavorite checks if a pseudonym has favorited a subforum
func (m *MockSubforumSubscriptionDAO) IsFavorite(ctx context.Context, pseudonymID string, subforumID int32) (bool, error) {
	args := m.Called(ctx, pseudonymID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, int32) (bool, error)); ok {
		return fn(ctx, pseudonymID, subforumID)
	}

	// Fallback to direct return values
	return args.Get(0).(bool), args.Error(1)
}

// DeleteSubscription deletes a subscription
func (m *MockSubforumSubscriptionDAO) DeleteSubscription(ctx context.Context, pseudonymID string, subforumID int32) error {
	args := m.Called(ctx, pseudonymID, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, int32) error); ok {
		return fn(ctx, pseudonymID, subforumID)
	}

	// Fallback to direct return values
	return args.Error(0)
}

// CountSubscriptionsBySubforum counts subscriptions by subforum
func (m *MockSubforumSubscriptionDAO) CountSubscriptionsBySubforum(ctx context.Context, subforumID int32) (int64, error) {
	args := m.Called(ctx, subforumID)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int32) (int64, error)); ok {
		return fn(ctx, subforumID)
	}

	// Fallback to direct return values
	return args.Get(0).(int64), args.Error(1)
}
