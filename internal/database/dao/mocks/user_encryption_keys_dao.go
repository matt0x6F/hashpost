package mocks

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockUserEncryptionKeyDAO is a mock of UserEncryptionKeyDAOInterface interface.
type MockUserEncryptionKeyDAO struct {
	mock.Mock
}

// CreateUserEncryptionKey mocks base method.
func (m *MockUserEncryptionKeyDAO) CreateUserEncryptionKey(ctx context.Context, userID int64, encryptedMasterKey, encryptedMessageKeys, encryptedSignatureKey, publicSignatureKey []byte, keyFingerprint string) (*models.UserEncryptionKey, error) {
	args := m.Called(ctx, userID, encryptedMasterKey, encryptedMessageKeys, encryptedSignatureKey, publicSignatureKey, keyFingerprint)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserEncryptionKey), args.Error(1)
}

// GetUserEncryptionKey mocks base method.
func (m *MockUserEncryptionKeyDAO) GetUserEncryptionKey(ctx context.Context, userID int64) (*models.UserEncryptionKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.UserEncryptionKey), args.Error(1)
}

// UpdateUserEncryptionKey mocks base method.
func (m *MockUserEncryptionKeyDAO) UpdateUserEncryptionKey(ctx context.Context, userID int64, encryptedMasterKey, encryptedMessageKeys, encryptedSignatureKey, publicSignatureKey []byte, keyFingerprint string) error {
	args := m.Called(ctx, userID, encryptedMasterKey, encryptedMessageKeys, encryptedSignatureKey, publicSignatureKey, keyFingerprint)
	return args.Error(0)
}

// DeleteUserEncryptionKey mocks base method.
func (m *MockUserEncryptionKeyDAO) DeleteUserEncryptionKey(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

// RotateUserEncryptionKey mocks base method.
func (m *MockUserEncryptionKeyDAO) RotateUserEncryptionKey(ctx context.Context, userID int64, encryptedMasterKey, encryptedMessageKeys, encryptedSignatureKey, publicSignatureKey []byte, keyFingerprint string) error {
	args := m.Called(ctx, userID, encryptedMasterKey, encryptedMessageKeys, encryptedSignatureKey, publicSignatureKey, keyFingerprint)
	return args.Error(0)
}

// GetUserPublicKey mocks base method.
func (m *MockUserEncryptionKeyDAO) GetUserPublicKey(ctx context.Context, userID int64) ([]byte, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]byte), args.Error(1)
}
