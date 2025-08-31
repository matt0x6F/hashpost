package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
)

// KeyManagementService handles the creation and management of encryption keys for messaging
type KeyManagementService struct {
	encryptionService *EncryptionService
	roleKeyDAO        dao.RoleKeyDAOInterface
	ibeSystem         *ibe.IBESystem
}

// NewKeyManagementService creates a new key management service
func NewKeyManagementService(
	encryptionService *EncryptionService,
	roleKeyDAO dao.RoleKeyDAOInterface,
	ibeSystem *ibe.IBESystem,
) *KeyManagementService {
	return &KeyManagementService{
		encryptionService: encryptionService,
		roleKeyDAO:        roleKeyDAO,
		ibeSystem:         ibeSystem,
	}
}

// EnsureMessagingKeys ensures that a user has the necessary messaging encryption keys
func (s *KeyManagementService) EnsureMessagingKeys(ctx context.Context, userID int64, pseudonymID string) error {
	// Check if messaging keys already exist
	existing, err := s.roleKeyDAO.GetRoleKey(ctx, pseudonymID, constants.ScopeMessaging, nil)
	if err == nil && existing != nil {
		log.Info().Int64("user_id", userID).Str("pseudonym_id", pseudonymID).Msg("Messaging keys already exist")
		return nil // Keys already exist
	}

	log.Info().Int64("user_id", userID).Str("pseudonym_id", pseudonymID).Msg("Creating new messaging keys")

	// Generate new messaging role key using the messaging domain
	messagingKey := s.ibeSystem.GenerateMessagingKey(
		"user",           // role
		"messaging",      // scope (triggers DOMAIN_USER_MESSAGING)
		time.Hour*24*365, // 1 year expiration
	)

	if messagingKey == nil {
		return fmt.Errorf("failed to generate messaging key")
	}

	// Store in role_keys table using the correct method signature
	_, err = s.roleKeyDAO.CreateRoleKey(
		ctx,
		"user",                   // roleName
		constants.ScopeMessaging, // scope
		messagingKey,             // keyData
		[]string{ // capabilities
			constants.CapabilitySendDirectMessages,
			constants.CapabilityReceiveDirectMessages,
			constants.CapabilityManageConversationKeys,
		},
		time.Now().AddDate(1, 0, 0), // expiresAt
		pseudonymID,                 // createdByPseudonymID
		pseudonymID,                 // pseudonymID
		nil,                         // subforumID (global scope)
	)
	if err != nil {
		return fmt.Errorf("failed to create messaging role key: %w", err)
	}

	log.Info().Int64("user_id", userID).Str("pseudonym_id", pseudonymID).Msg("Messaging keys created successfully")
	return nil
}

// GenerateMessageKey generates a new message encryption key
func (s *KeyManagementService) GenerateMessageKey() (*MessageKey, error) {
	// Generate AES-256 key
	keyData, err := s.encryptionService.GenerateAESKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Create key metadata
	keyID := s.generateKeyID(keyData)
	now := time.Now()

	return &MessageKey{
		KeyID:      keyID,
		KeyData:    keyData,
		CreatedAt:  now,
		ExpiresAt:  now.AddDate(0, 0, 30), // 30 days
		IsActive:   true,
		KeyVersion: 1,
	}, nil
}

// GenerateConversationKey generates a new conversation encryption key
func (s *KeyManagementService) GenerateConversationKey() (*ConversationKey, error) {
	// Generate AES-256 key
	keyData, err := s.encryptionService.GenerateAESKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}

	// Create key metadata
	conversationID := s.generateConversationID(keyData)
	now := time.Now()

	return &ConversationKey{
		ConversationID: conversationID,
		KeyData:        keyData,
		CreatedAt:      now,
		ExpiresAt:      now.AddDate(0, 0, 7), // 7 days for forward secrecy
		IsActive:       true,
	}, nil
}

// GenerateSignatureKeyPair generates a new Ed25519 signing key pair
func (s *KeyManagementService) GenerateSignatureKeyPair() (*SignatureKeyPair, error) {
	return s.encryptionService.GenerateSignatureKeyPair()
}

// EncryptMessageKey encrypts a message key with a master key
func (s *KeyManagementService) EncryptMessageKey(messageKey *MessageKey, masterKey []byte) ([]byte, error) {
	// Serialize the message key
	keyData, err := s.encryptionService.SerializeMessageKeys([]*MessageKey{messageKey})
	if err != nil {
		return nil, fmt.Errorf("failed to serialize message key: %w", err)
	}

	// Encrypt with master key
	return s.encryptionService.EncryptMessageKey(keyData, masterKey)
}

// DecryptMessageKey decrypts a message key with a master key
func (s *KeyManagementService) DecryptMessageKey(encryptedKey, masterKey []byte) (*MessageKey, error) {
	// Decrypt with master key
	keyData, err := s.encryptionService.DecryptMessageKey(encryptedKey, masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message key: %w", err)
	}

	// Deserialize the message key
	keys, err := s.encryptionService.DeserializeMessageKeys(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize message key: %w", err)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no message keys found")
	}

	return keys[0], nil
}

// EncryptConversationKey encrypts a conversation key for storage
func (s *KeyManagementService) EncryptConversationKey(conversationKey *ConversationKey, userKeys map[int64][]byte) (map[int64][]byte, error) {
	encryptedKeys := make(map[int64][]byte)

	for userID, userKey := range userKeys {
		// Encrypt conversation key with user's key
		encrypted, err := s.encryptionService.EncryptMessageKey(userKey, conversationKey.KeyData)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt conversation key for user %d: %w", userID, err)
		}
		encryptedKeys[userID] = encrypted
	}

	return encryptedKeys, nil
}

// DecryptConversationKey decrypts a conversation key using a user's key
func (s *KeyManagementService) DecryptConversationKey(encryptedKey, userKey []byte) ([]byte, error) {
	return s.encryptionService.DecryptMessageKey(userKey, encryptedKey)
}

// RotateMessageKeys rotates expired message keys
func (s *KeyManagementService) RotateMessageKeys(ctx context.Context, userID int64) error {
	// This would involve:
	// 1. Getting current message keys
	// 2. Checking expiration
	// 3. Generating new keys
	// 4. Updating the database
	// For now, just log that rotation is needed
	log.Info().Int64("user_id", userID).Msg("Message key rotation requested")
	return nil
}

// RotateConversationKeys rotates expired conversation keys
func (s *KeyManagementService) RotateConversationKeys(ctx context.Context, conversationID string) error {
	// This would involve:
	// 1. Getting current conversation key
	// 2. Checking expiration
	// 3. Generating new key
	// 4. Updating the database
	// For now, just log that rotation is needed
	log.Info().Str("conversation_id", conversationID).Msg("Conversation key rotation requested")
	return nil
}

// generateKeyID generates a unique identifier for a key
func (s *KeyManagementService) generateKeyID(key []byte) string {
	hash := sha256.Sum256(key)
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for shorter ID
}

// generateConversationID generates a unique identifier for a conversation
func (s *KeyManagementService) generateConversationID(key []byte) string {
	hash := sha256.Sum256(key)
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for shorter ID
}
