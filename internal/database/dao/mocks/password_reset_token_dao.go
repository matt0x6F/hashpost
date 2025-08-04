package mocks

import (
	"context"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockPasswordResetTokenDAO is a mock implementation of PasswordResetTokenDAOInterface
type MockPasswordResetTokenDAO struct {
	mock.Mock
}

func (m *MockPasswordResetTokenDAO) CreateToken(ctx context.Context, userID int64, token string, expiresAt time.Time) error {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Error(0)
}

func (m *MockPasswordResetTokenDAO) GetToken(ctx context.Context, token string) (*models.PasswordResetToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.PasswordResetToken), args.Error(1)
}

func (m *MockPasswordResetTokenDAO) MarkTokenAsUsed(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockPasswordResetTokenDAO) DeleteExpiredTokens(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockPasswordResetTokenDAO) DeleteTokensByUserID(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}