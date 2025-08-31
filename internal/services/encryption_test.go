package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncryptionService_GenerateAESKey(t *testing.T) {
	service := NewEncryptionService()

	key, err := service.GenerateAESKey()
	require.NoError(t, err)
	assert.Len(t, key, 32) // AES-256 key should be 32 bytes
}

func TestEncryptionService_GenerateSignatureKeyPair(t *testing.T) {
	service := NewEncryptionService()

	keyPair, err := service.GenerateSignatureKeyPair()
	require.NoError(t, err)
	assert.NotNil(t, keyPair)
	assert.Len(t, keyPair.PrivateKey, 64) // Ed25519 private key is 64 bytes
	assert.Len(t, keyPair.PublicKey, 32)  // Ed25519 public key is 32 bytes
	assert.NotEmpty(t, keyPair.KeyID)
}

func TestEncryptionService_EncryptDecryptAES(t *testing.T) {
	service := NewEncryptionService()

	// Generate a key
	key, err := service.GenerateAESKey()
	require.NoError(t, err)

	// Test data
	plaintext := []byte("Hello, encrypted world!")

	// Encrypt
	encrypted, err := service.EncryptAES(key, plaintext)
	require.NoError(t, err)
	assert.NotNil(t, encrypted)
	assert.NotEqual(t, plaintext, encrypted.EncryptedContent)
	assert.NotEmpty(t, encrypted.IV)
	assert.NotEmpty(t, encrypted.ContentHash)
	assert.Equal(t, 1, encrypted.KeyVersion)

	// Decrypt
	decrypted, err := service.DecryptAES(key, encrypted)
	require.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestEncryptionService_SignVerifyMessage(t *testing.T) {
	service := NewEncryptionService()

	// Generate key pair
	keyPair, err := service.GenerateSignatureKeyPair()
	require.NoError(t, err)

	// Test message
	message := []byte("Hello, signed world!")

	// Sign
	signature, err := service.SignMessage(keyPair.PrivateKey, message)
	require.NoError(t, err)
	assert.NotEmpty(t, signature)

	// Verify
	valid := service.VerifySignature(keyPair.PublicKey, message, signature)
	assert.True(t, valid)

	// Verify with wrong message
	invalid := service.VerifySignature(keyPair.PublicKey, []byte("Wrong message"), signature)
	assert.False(t, invalid)
}

func TestEncryptionService_MessageKey(t *testing.T) {
	service := NewEncryptionService()

	// Generate a master key
	masterKey, err := service.GenerateAESKey()
	require.NoError(t, err)

	// Generate a message key
	messageKey := &MessageKey{
		KeyID:      "test-key-1",
		KeyData:    []byte("test-key-data"),
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().AddDate(0, 0, 30),
		IsActive:   true,
		KeyVersion: 1,
	}

	// Encrypt message key
	encrypted, err := service.EncryptMessageKey(masterKey, messageKey.KeyData)
	require.NoError(t, err)
	assert.NotEqual(t, messageKey.KeyData, encrypted)

	// Decrypt message key
	decrypted, err := service.DecryptMessageKey(masterKey, encrypted)
	require.NoError(t, err)
	assert.Equal(t, messageKey.KeyData, decrypted)
}

func TestEncryptionService_SerializeDeserializeMessageKeys(t *testing.T) {
	service := NewEncryptionService()

	// Create test keys
	keys := []*MessageKey{
		{
			KeyID:      "key-1",
			KeyData:    []byte("data-1"),
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().AddDate(0, 0, 30),
			IsActive:   true,
			KeyVersion: 1,
		},
		{
			KeyID:      "key-2",
			KeyData:    []byte("data-2"),
			CreatedAt:  time.Now(),
			ExpiresAt:  time.Now().AddDate(0, 0, 30),
			IsActive:   true,
			KeyVersion: 2,
		},
	}

	// Serialize
	data, err := service.SerializeMessageKeys(keys)
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Deserialize
	deserialized, err := service.DeserializeMessageKeys(data)
	require.NoError(t, err)
	assert.Len(t, deserialized, 2)
	assert.Equal(t, keys[0].KeyID, deserialized[0].KeyID)
	assert.Equal(t, keys[1].KeyID, deserialized[1].KeyID)
}

func TestEncryptionService_KeyExpiration(t *testing.T) {
	service := NewEncryptionService()

	// Test expired key
	expiredKey := &MessageKey{
		ExpiresAt: time.Now().AddDate(0, 0, -1), // Expired yesterday
	}
	assert.True(t, service.IsKeyExpired(expiredKey))

	// Test valid key
	validKey := &MessageKey{
		ExpiresAt: time.Now().AddDate(0, 0, 1), // Expires tomorrow
	}
	assert.False(t, service.IsKeyExpired(validKey))

	// Test conversation key expiration
	expiredConvKey := &ConversationKey{
		ExpiresAt: time.Now().AddDate(0, 0, -1), // Expired yesterday
	}
	assert.True(t, service.IsConversationKeyExpired(expiredConvKey))

	validConvKey := &ConversationKey{
		ExpiresAt: time.Now().AddDate(0, 0, 1), // Expires tomorrow
	}
	assert.False(t, service.IsConversationKeyExpired(validConvKey))
}
