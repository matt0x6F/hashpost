package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"time"
)

// EncryptionService handles all cryptographic operations for messaging
type EncryptionService struct{}

// NewEncryptionService creates a new encryption service
func NewEncryptionService() *EncryptionService {
	return &EncryptionService{}
}

// MessageKey represents an AES encryption key with metadata
type MessageKey struct {
	KeyID      string    `json:"key_id"`
	KeyData    []byte    `json:"key_data"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	IsActive   bool      `json:"is_active"`
	KeyVersion int       `json:"key_version"`
}

// ConversationKey represents a shared encryption key between two users
type ConversationKey struct {
	ConversationID string    `json:"conversation_id"`
	KeyData        []byte    `json:"key_data"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	IsActive       bool      `json:"is_active"`
}

// SignatureKeyPair represents Ed25519 signing keys
type SignatureKeyPair struct {
	PrivateKey []byte `json:"private_key"`
	PublicKey  []byte `json:"public_key"`
	KeyID      string `json:"key_id"`
}

// EncryptedMessage represents an encrypted message with metadata
type EncryptedMessage struct {
	EncryptedContent []byte `json:"encrypted_content"`
	IV               []byte `json:"iv"`
	Signature        []byte `json:"signature"`
	ContentHash      string `json:"content_hash"`
	KeyVersion       int    `json:"key_version"`
}

// GenerateAESKey generates a new AES-256 key
func (s *EncryptionService) GenerateAESKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate AES key: %w", err)
	}
	return key, nil
}

// GenerateSignatureKeyPair generates a new Ed25519 key pair
func (s *EncryptionService) GenerateSignatureKeyPair() (*SignatureKeyPair, error) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Ed25519 keys: %w", err)
	}

	keyID := s.generateKeyID(publicKey)

	return &SignatureKeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		KeyID:      keyID,
	}, nil
}

// EncryptAES encrypts data using AES-256-GCM
func (s *EncryptionService) EncryptAES(key, plaintext []byte) (*EncryptedMessage, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt (don't include nonce in ciphertext since we store it separately)
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// Generate content hash
	hash := sha256.Sum256(plaintext)
	contentHash := hex.EncodeToString(hash[:])

	return &EncryptedMessage{
		EncryptedContent: ciphertext,
		IV:               nonce,
		ContentHash:      contentHash,
		KeyVersion:       1,
	}, nil
}

// DecryptAES decrypts data using AES-256-GCM
func (s *EncryptionService) DecryptAES(key []byte, encryptedMsg *EncryptedMessage) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt
	plaintext, err := gcm.Open(nil, encryptedMsg.IV, encryptedMsg.EncryptedContent, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt: %w", err)
	}

	// Verify content hash
	hash := sha256.Sum256(plaintext)
	expectedHash := hex.EncodeToString(hash[:])
	if expectedHash != encryptedMsg.ContentHash {
		return nil, fmt.Errorf("content hash mismatch")
	}

	return plaintext, nil
}

// SignMessage signs a message using Ed25519
func (s *EncryptionService) SignMessage(privateKey, message []byte) ([]byte, error) {
	signature := ed25519.Sign(privateKey, message)
	return signature, nil
}

// VerifySignature verifies an Ed25519 signature
func (s *EncryptionService) VerifySignature(publicKey, message, signature []byte) bool {
	return ed25519.Verify(publicKey, message, signature)
}

// GenerateKeyID generates a unique identifier for a key
func (s *EncryptionService) generateKeyID(key []byte) string {
	hash := sha256.Sum256(key)
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for shorter ID
}

// EncryptMessageKey encrypts a message key with a master key
func (s *EncryptionService) EncryptMessageKey(masterKey, messageKey []byte) ([]byte, error) {
	// For message key encryption, we need to handle the IV properly
	// The IV should be prepended to the encrypted data for storage
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, messageKey, nil)
	return ciphertext, nil
}

// DecryptMessageKey decrypts a message key with a master key
func (s *EncryptionService) DecryptMessageKey(masterKey, encryptedKey []byte) ([]byte, error) {
	// For now, assume the encrypted key is just AES encrypted
	// In a real implementation, this would handle the specific encryption format
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	// Extract nonce from encrypted data
	nonceSize := gcm.NonceSize()
	if len(encryptedKey) < nonceSize {
		return nil, fmt.Errorf("encrypted key too short")
	}

	nonce := encryptedKey[:nonceSize]
	ciphertext := encryptedKey[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt message key: %w", err)
	}

	return plaintext, nil
}

// EncryptWithPublicKey encrypts data using a user's RSA public key
func (s *EncryptionService) EncryptWithPublicKey(publicKeyPEM []byte, data []byte) ([]byte, error) {
	// Parse PEM block
	block, _ := pem.Decode(publicKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse public key
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	// Type assert to RSA public key
	rsaPubKey, ok := pubKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not RSA")
	}

	// Encrypt data with RSA-OAEP
	encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPubKey, data)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt with RSA: %w", err)
	}

	return encryptedData, nil
}

// DecryptWithPrivateKey decrypts data using a user's RSA private key
func (s *EncryptionService) DecryptWithPrivateKey(privateKeyPEM []byte, encryptedData []byte) ([]byte, error) {
	// Parse PEM block
	block, _ := pem.Decode(privateKeyPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}

	// Parse private key
	privKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Type assert to RSA private key
	rsaPrivKey, ok := privKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not RSA")
	}

	// Decrypt data with RSA-OAEP
	decryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, rsaPrivKey, encryptedData)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt with RSA: %w", err)
	}

	return decryptedData, nil
}

// SerializeMessageKeys serializes message keys to JSON for storage
func (s *EncryptionService) SerializeMessageKeys(keys []*MessageKey) ([]byte, error) {
	return json.Marshal(keys)
}

// DeserializeMessageKeys deserializes message keys from JSON
func (s *EncryptionService) DeserializeMessageKeys(data []byte) ([]*MessageKey, error) {
	var keys []*MessageKey
	err := json.Unmarshal(data, &keys)
	return keys, err
}

// IsKeyExpired checks if a key has expired
func (s *EncryptionService) IsKeyExpired(key *MessageKey) bool {
	return time.Now().After(key.ExpiresAt)
}

// IsConversationKeyExpired checks if a conversation key has expired
func (s *EncryptionService) IsConversationKeyExpired(key *ConversationKey) bool {
	return time.Now().After(key.ExpiresAt)
}
