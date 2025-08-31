package services

import (
	"context"
)

// EncryptionServiceInterface defines the interface for encryption operations
type EncryptionServiceInterface interface {
	GenerateAESKey() ([]byte, error)
	GenerateSignatureKeyPair() (*SignatureKeyPair, error)
	EncryptAES(plaintext, key []byte) (*EncryptedMessage, error)
	DecryptAES(key []byte, encryptedMsg *EncryptedMessage) ([]byte, error)
	EncryptMessageKey(messageKey, masterKey []byte) ([]byte, error)
	DecryptMessageKey(encryptedMessageKey, masterKey []byte) ([]byte, error)
	EncryptWithPublicKey(publicKeyPEM []byte, data []byte) ([]byte, error)
	DecryptWithPrivateKey(privateKeyPEM []byte, encryptedData []byte) ([]byte, error)
	SignMessage(message, privateKey []byte) ([]byte, error)
	VerifySignature(publicKey, message, signature []byte) bool
}

// KeyManagementServiceInterface defines the interface for key management operations
type KeyManagementServiceInterface interface {
	EnsureMessagingKeys(ctx context.Context, userID int64, pseudonymID string) error
	GenerateMessageKey() (*MessageKey, error)
	GenerateConversationKey() (*ConversationKey, error)
}
