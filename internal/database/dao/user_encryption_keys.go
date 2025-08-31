package dao

import (
	"context"
	"fmt"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// UserEncryptionKeyDAOInterface defines the interface for user encryption key operations
type UserEncryptionKeyDAOInterface interface {
	CreateUserEncryptionKey(ctx context.Context, userID int64, encryptedMasterKey, encryptedMessageKeys, encryptedSignatureKey, publicSignatureKey []byte, keyFingerprint string) (*models.UserEncryptionKey, error)
	GetUserEncryptionKey(ctx context.Context, userID int64) (*models.UserEncryptionKey, error)
	GetUserPublicKey(ctx context.Context, userID int64) ([]byte, error)
	UpdateUserEncryptionKey(ctx context.Context, userID int64, encryptedMasterKey, encryptedMessageKeys, encryptedSignatureKey, publicSignatureKey []byte, keyFingerprint string) error
	DeleteUserEncryptionKey(ctx context.Context, userID int64) error
	RotateUserEncryptionKey(ctx context.Context, userID int64, newEncryptedMasterKey, newEncryptedMessageKeys, newEncryptedSignatureKey, newPublicSignatureKey []byte, newKeyFingerprint string) error
}

// UserEncryptionKeyDAO implements UserEncryptionKeyDAOInterface
type UserEncryptionKeyDAO struct {
	db bob.Executor
}

// NewUserEncryptionKeyDAO creates a new UserEncryptionKeyDAO
func NewUserEncryptionKeyDAO(db bob.Executor) *UserEncryptionKeyDAO {
	return &UserEncryptionKeyDAO{db: db}
}

// CreateUserEncryptionKey creates a new user encryption key record
func (dao *UserEncryptionKeyDAO) CreateUserEncryptionKey(ctx context.Context, userID int64, encryptedMasterKey, encryptedMessageKeys, encryptedSignatureKey, publicSignatureKey []byte, keyFingerprint string) (*models.UserEncryptionKey, error) {
	// Create a user encryption key setter
	keySetter := &models.UserEncryptionKeySetter{
		UserID:                &userID,
		EncryptedMasterKey:    &encryptedMasterKey,
		EncryptedMessageKeys:  &encryptedMessageKeys,
		EncryptedSignatureKey: &encryptedSignatureKey,
		PublicSignatureKey:    &publicSignatureKey,
		KeyFingerprint:        &keyFingerprint,
	}

	// Use the generated UserEncryptionKeys table helper
	key, err := models.UserEncryptionKeys.Insert(keySetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create user encryption key: %w", err)
	}

	return key, nil
}

// GetUserEncryptionKey retrieves a user encryption key by user ID
func (dao *UserEncryptionKeyDAO) GetUserEncryptionKey(ctx context.Context, userID int64) (*models.UserEncryptionKey, error) {
	// Use the generated UserEncryptionKeys table helper with where clause
	keys, err := models.UserEncryptionKeys.Query(
		models.SelectWhere.UserEncryptionKeys.UserID.EQ(userID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get user encryption key: %w", err)
	}

	if len(keys) == 0 {
		return nil, nil
	}

	return keys[0], nil
}

// GetUserPublicKey retrieves a user's public key for encryption
func (dao *UserEncryptionKeyDAO) GetUserPublicKey(ctx context.Context, userID int64) ([]byte, error) {
	// Get the user's encryption key record
	userKey, err := dao.GetUserEncryptionKey(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user encryption key: %w", err)
	}
	if userKey == nil {
		return nil, fmt.Errorf("user encryption key not found")
	}

	// Return the public signature key (this should contain the RSA public key)
	// In a real implementation, this might be a separate field for RSA public keys
	return userKey.PublicSignatureKey, nil
}

// UpdateUserEncryptionKey updates an existing user encryption key
func (dao *UserEncryptionKeyDAO) UpdateUserEncryptionKey(ctx context.Context, userID int64, encryptedMasterKey, encryptedMessageKeys, encryptedSignatureKey, publicSignatureKey []byte, keyFingerprint string) error {
	// First get the existing key
	existingKey, err := dao.GetUserEncryptionKey(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user encryption key for update: %w", err)
	}
	if existingKey == nil {
		return fmt.Errorf("user encryption key not found")
	}

	// Create an update setter
	updates := &models.UserEncryptionKeySetter{
		EncryptedMasterKey:    &encryptedMasterKey,
		EncryptedMessageKeys:  &encryptedMessageKeys,
		EncryptedSignatureKey: &encryptedSignatureKey,
		PublicSignatureKey:    &publicSignatureKey,
		KeyFingerprint:        &keyFingerprint,
	}

	// Use the generated Update method
	err = existingKey.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update user encryption key: %w", err)
	}

	return nil
}

// DeleteUserEncryptionKey deletes a user encryption key
func (dao *UserEncryptionKeyDAO) DeleteUserEncryptionKey(ctx context.Context, userID int64) error {
	// First get the existing key
	existingKey, err := dao.GetUserEncryptionKey(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user encryption key for deletion: %w", err)
	}
	if existingKey == nil {
		return nil // Already deleted
	}

	// Use the generated Delete method
	err = existingKey.Delete(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete user encryption key: %w", err)
	}

	return nil
}

// RotateUserEncryptionKey rotates a user's encryption keys
func (dao *UserEncryptionKeyDAO) RotateUserEncryptionKey(ctx context.Context, userID int64, newEncryptedMasterKey, newEncryptedMessageKeys, newEncryptedSignatureKey, newPublicSignatureKey []byte, newKeyFingerprint string) error {
	return dao.UpdateUserEncryptionKey(ctx, userID, newEncryptedMasterKey, newEncryptedMessageKeys, newEncryptedSignatureKey, newPublicSignatureKey, newKeyFingerprint)
}
