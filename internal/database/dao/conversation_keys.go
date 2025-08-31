package dao

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// ConversationKeyDAOInterface defines the interface for conversation key operations
type ConversationKeyDAOInterface interface {
	CreateConversationKey(ctx context.Context, participant1UserID, participant2UserID int64, encryptedConversationKey []byte, keyFingerprint string, expiresAt time.Time) (*models.ConversationKey, error)
	GetConversationKey(ctx context.Context, conversationID uuid.UUID) (*models.ConversationKey, error)
	GetConversationKeyByParticipants(ctx context.Context, participant1UserID, participant2UserID int64) (*models.ConversationKey, error)
	UpdateConversationKey(ctx context.Context, conversationID uuid.UUID, encryptedConversationKey []byte, keyFingerprint string, expiresAt time.Time) error
	DeleteConversationKey(ctx context.Context, conversationID uuid.UUID) error
	RotateConversationKey(ctx context.Context, conversationID uuid.UUID, newEncryptedConversationKey []byte, newKeyFingerprint string, newExpiresAt time.Time) error
	GetActiveConversationKeys(ctx context.Context, userID int64) ([]*models.ConversationKey, error)
	CleanupExpiredKeys(ctx context.Context) error
}

// ConversationKeyDAO implements ConversationKeyDAOInterface
type ConversationKeyDAO struct {
	db bob.Executor
}

// NewConversationKeyDAO creates a new ConversationKeyDAO
func NewConversationKeyDAO(db bob.Executor) *ConversationKeyDAO {
	return &ConversationKeyDAO{db: db}
}

// CreateConversationKey creates a new conversation key record
func (dao *ConversationKeyDAO) CreateConversationKey(ctx context.Context, participant1UserID, participant2UserID int64, encryptedConversationKey []byte, keyFingerprint string, expiresAt time.Time) (*models.ConversationKey, error) {
	// Ensure consistent ordering of participants (smaller ID first)
	if participant1UserID > participant2UserID {
		participant1UserID, participant2UserID = participant2UserID, participant1UserID
	}

	// Create a conversation key setter
	expiresAtNull := sql.Null[time.Time]{}
	expiresAtNull.Scan(expiresAt)
	isActiveNull := sql.Null[bool]{}
	isActiveNull.Scan(true)

	keySetter := &models.ConversationKeySetter{
		Participant1UserID: &participant1UserID,
		Participant2UserID: &participant2UserID,
		EncryptedSharedKey: &encryptedConversationKey,
		KeyFingerprint:     &keyFingerprint,
		ExpiresAt:          &expiresAtNull,
		IsActive:           &isActiveNull,
	}

	// Use the generated ConversationKeys table helper
	key, err := models.ConversationKeys.Insert(keySetter).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation key: %w", err)
	}

	return key, nil
}

// GetConversationKey retrieves a conversation key by conversation ID
func (dao *ConversationKeyDAO) GetConversationKey(ctx context.Context, conversationID uuid.UUID) (*models.ConversationKey, error) {
	// Use the generated FindConversationKey function
	key, err := models.FindConversationKey(ctx, dao.db, conversationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation key: %w", err)
	}

	return key, nil
}

// GetConversationKeyByParticipants retrieves a conversation key by participant user IDs
func (dao *ConversationKeyDAO) GetConversationKeyByParticipants(ctx context.Context, participant1UserID, participant2UserID int64) (*models.ConversationKey, error) {
	// Ensure consistent ordering of participants (smaller ID first)
	if participant1UserID > participant2UserID {
		participant1UserID, participant2UserID = participant2UserID, participant1UserID
	}

	// Use the generated ConversationKeys table helper with where clause
	keys, err := models.ConversationKeys.Query(
		models.SelectWhere.ConversationKeys.Participant1UserID.EQ(participant1UserID),
		models.SelectWhere.ConversationKeys.Participant2UserID.EQ(participant2UserID),
		models.SelectWhere.ConversationKeys.IsActive.EQ(true),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation key by participants: %w", err)
	}

	if len(keys) == 0 {
		return nil, nil
	}

	return keys[0], nil
}

// UpdateConversationKey updates an existing conversation key
func (dao *ConversationKeyDAO) UpdateConversationKey(ctx context.Context, conversationID uuid.UUID, encryptedConversationKey []byte, keyFingerprint string, expiresAt time.Time) error {
	// First get the existing key
	existingKey, err := dao.GetConversationKey(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation key for update: %w", err)
	}
	if existingKey == nil {
		return fmt.Errorf("conversation key not found")
	}

	// Create an update setter
	expiresAtNull := sql.Null[time.Time]{}
	expiresAtNull.Scan(expiresAt)

	updates := &models.ConversationKeySetter{
		EncryptedSharedKey: &encryptedConversationKey,
		KeyFingerprint:     &keyFingerprint,
		ExpiresAt:          &expiresAtNull,
	}

	// Use the generated Update method
	err = existingKey.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update conversation key: %w", err)
	}

	return nil
}

// DeleteConversationKey deletes a conversation key
func (dao *ConversationKeyDAO) DeleteConversationKey(ctx context.Context, conversationID uuid.UUID) error {
	// First get the existing key
	existingKey, err := dao.GetConversationKey(ctx, conversationID)
	if err != nil {
		return fmt.Errorf("failed to get conversation key for deletion: %w", err)
	}
	if existingKey == nil {
		return nil // Already deleted
	}

	// Use the generated Delete method
	err = existingKey.Delete(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to delete conversation key: %w", err)
	}

	return nil
}

// RotateConversationKey rotates a conversation's encryption key
func (dao *ConversationKeyDAO) RotateConversationKey(ctx context.Context, conversationID uuid.UUID, newEncryptedConversationKey []byte, newKeyFingerprint string, newExpiresAt time.Time) error {
	return dao.UpdateConversationKey(ctx, conversationID, newEncryptedConversationKey, newKeyFingerprint, newExpiresAt)
}

// GetActiveConversationKeys retrieves all active conversation keys for a user
func (dao *ConversationKeyDAO) GetActiveConversationKeys(ctx context.Context, userID int64) ([]*models.ConversationKey, error) {
	// Get conversations where user is participant1 or participant2 and keys are active
	keys, err := models.ConversationKeys.Query(
		models.SelectWhere.ConversationKeys.IsActive.EQ(true),
		models.SelectWhere.ConversationKeys.ExpiresAt.GT(time.Now()),
		models.SelectWhere.ConversationKeys.Participant1UserID.EQ(userID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get active conversation keys as participant1: %w", err)
	}

	// Also get conversations where user is participant2
	keys2, err := models.ConversationKeys.Query(
		models.SelectWhere.ConversationKeys.IsActive.EQ(true),
		models.SelectWhere.ConversationKeys.ExpiresAt.GT(time.Now()),
		models.SelectWhere.ConversationKeys.Participant2UserID.EQ(userID),
	).All(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get active conversation keys as participant2: %w", err)
	}

	// Combine and return unique keys
	allKeys := append(keys, keys2...)
	return allKeys, nil
}

// CleanupExpiredKeys removes expired conversation keys
func (dao *ConversationKeyDAO) CleanupExpiredKeys(ctx context.Context) error {
	// Get all expired keys
	expiredKeys, err := models.ConversationKeys.Query(
		models.SelectWhere.ConversationKeys.ExpiresAt.LTE(time.Now()),
		models.SelectWhere.ConversationKeys.IsActive.EQ(true),
	).All(ctx, dao.db)
	if err != nil {
		return fmt.Errorf("failed to get expired conversation keys: %w", err)
	}

	// Mark expired keys as inactive
	for _, key := range expiredKeys {
		isActiveNull := sql.Null[bool]{}
		isActiveNull.Scan(false)

		updates := &models.ConversationKeySetter{
			IsActive: &isActiveNull,
		}
		err = key.Update(ctx, dao.db, updates)
		if err != nil {
			return fmt.Errorf("failed to mark conversation key as inactive: %w", err)
		}
	}

	return nil
}
