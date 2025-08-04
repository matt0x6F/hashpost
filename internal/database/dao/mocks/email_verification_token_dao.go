package mocks

import (
	"context"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockEmailVerificationTokenDAO is a mock implementation of EmailVerificationTokenDAOInterface
type MockEmailVerificationTokenDAO struct {
	mock.Mock
}

func (m *MockEmailVerificationTokenDAO) CreateToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Error(0)
}

func (m *MockEmailVerificationTokenDAO) GetToken(ctx context.Context, token string) (*models.EmailVerificationToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.EmailVerificationToken), args.Error(1)
}

func (m *MockEmailVerificationTokenDAO) MarkTokenAsUsed(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockEmailVerificationTokenDAO) DeleteExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockEmailVerificationTokenDAO) DeleteTokensByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}