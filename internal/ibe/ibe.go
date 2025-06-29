package ibe

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

// IBESystem represents the core IBE functionality
type IBESystem struct {
	masterSecret []byte
	keyVersion   int
}

// NewIBESystem creates a new IBE system with a master secret
func NewIBESystem() *IBESystem {
	masterSecret := make([]byte, 32)
	rand.Read(masterSecret)

	return &IBESystem{
		masterSecret: masterSecret,
		keyVersion:   1,
	}
}

// GeneratePseudonym creates a pseudonym ID for a user
func (ibe *IBESystem) GeneratePseudonym(userSecret []byte) string {
	// Combine user secret with system master secret
	combined := append(userSecret, ibe.masterSecret...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for pseudonym ID
}

// GenerateFingerprint creates a deterministic fingerprint from a real identity
// This allows correlation without revealing the actual identity
func (ibe *IBESystem) GenerateFingerprint(realIdentity string) string {
	// Combine real identity with a system-wide salt for fingerprint generation
	salt := []byte("fingerprint_salt_v1")
	combined := append([]byte(realIdentity), salt...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:16]) // Use first 16 bytes for fingerprint
}

// EncryptIdentity encrypts the mapping between real identity and pseudonym
func (ibe *IBESystem) EncryptIdentity(realIdentity, pseudonymID string, adminKey []byte) ([]byte, error) {
	// Create the mapping data with fingerprint instead of real identity
	fingerprint := ibe.GenerateFingerprint(realIdentity)
	mapping := fmt.Sprintf("%s:%s", fingerprint, pseudonymID)

	// Use admin key to derive encryption key
	key := sha256.Sum256(adminKey)

	// Create AES cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	rand.Read(nonce)

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, []byte(mapping), nil)
	return ciphertext, nil
}

// DecryptIdentity decrypts the mapping using admin key
func (ibe *IBESystem) DecryptIdentity(encryptedMapping []byte, adminKey []byte) (string, string, error) {
	// Use admin key to derive decryption key
	key := sha256.Sum256(adminKey)

	// Create AES cipher
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", "", err
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(encryptedMapping) < nonceSize {
		return "", "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedMapping[:nonceSize], encryptedMapping[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", "", err
	}

	// Parse mapping
	mapping := string(plaintext)
	// In a real implementation, you'd parse this more carefully
	return mapping, "", nil
}

// GenerateRoleKey creates a role-based key for administrative access
func (ibe *IBESystem) GenerateRoleKey(role string, scope string, expiration time.Time) []byte {
	// Combine role, scope, and expiration with master secret
	keyData := fmt.Sprintf("%s:%s:%d", role, scope, expiration.Unix())
	combined := append([]byte(keyData), ibe.masterSecret...)
	hash := sha256.Sum256(combined)
	return hash[:]
}

// ValidateRoleKey checks if a role key is valid and not expired
func (ibe *IBESystem) ValidateRoleKey(roleKey []byte, role string, scope string, expiration time.Time) bool {
	expectedKey := ibe.GenerateRoleKey(role, scope, expiration)
	if !time.Now().Before(expiration) {
		return false
	}
	if len(roleKey) != len(expectedKey) {
		return false
	}
	for i := range roleKey {
		if roleKey[i] != expectedKey[i] {
			return false
		}
	}
	return true
}
