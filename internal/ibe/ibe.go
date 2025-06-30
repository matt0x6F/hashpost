package ibe

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

// IBESystem represents the core IBE functionality
type IBESystem struct {
	masterSecret []byte
	keyVersion   int
	salt         []byte
}

// IBEOptions provides configuration options for the IBE system
type IBEOptions struct {
	MasterSecret []byte // Optional: if provided, use this instead of generating random
	KeyVersion   int    // Optional: defaults to 1
	Salt         string // Optional: salt for fingerprint generation, defaults to "fingerprint_salt_v1"
}

// NewIBESystem creates a new IBE system with a master secret
func NewIBESystem() *IBESystem {
	return NewIBESystemWithOptions(IBEOptions{})
}

// NewIBESystemWithOptions creates a new IBE system with configuration options
func NewIBESystemWithOptions(opts IBEOptions) *IBESystem {
	var masterSecret []byte
	if opts.MasterSecret != nil {
		masterSecret = make([]byte, len(opts.MasterSecret))
		copy(masterSecret, opts.MasterSecret)
	} else {
		masterSecret = make([]byte, 32)
		rand.Read(masterSecret)
	}

	keyVersion := opts.KeyVersion
	if keyVersion == 0 {
		keyVersion = 1
	}

	salt := opts.Salt
	if salt == "" {
		salt = "fingerprint_salt_v1"
	}

	return &IBESystem{
		masterSecret: masterSecret,
		keyVersion:   keyVersion,
		salt:         []byte(salt),
	}
}

// GetMasterSecret returns a copy of the master secret (for persistence)
func (ibe *IBESystem) GetMasterSecret() []byte {
	secret := make([]byte, len(ibe.masterSecret))
	copy(secret, ibe.masterSecret)
	return secret
}

// SetMasterSecret sets the master secret (for loading from persistence)
func (ibe *IBESystem) SetMasterSecret(secret []byte) error {
	if len(secret) != 32 {
		return fmt.Errorf("master secret must be 32 bytes, got %d", len(secret))
	}
	ibe.masterSecret = make([]byte, len(secret))
	copy(ibe.masterSecret, secret)
	return nil
}

// GetKeyVersion returns the current key version
func (ibe *IBESystem) GetKeyVersion() int {
	return ibe.keyVersion
}

// SetKeyVersion sets the key version
func (ibe *IBESystem) SetKeyVersion(version int) {
	ibe.keyVersion = version
}

// SetSalt sets the salt for fingerprint generation
func (ibe *IBESystem) SetSalt(salt string) {
	ibe.salt = []byte(salt)
}

// GetSalt returns the current salt
func (ibe *IBESystem) GetSalt() string {
	return string(ibe.salt)
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
	// Combine real identity with the configurable salt for fingerprint generation
	combined := append([]byte(realIdentity), ibe.salt...)
	hash := sha256.Sum256(combined)
	fingerprint := hex.EncodeToString(hash[:16]) // Use first 16 bytes for fingerprint

	// Debug logging
	log.Info().
		Str("real_identity", realIdentity).
		Str("salt", string(ibe.salt)).
		Str("salt_hex", hex.EncodeToString(ibe.salt)).
		Str("combined_hex", hex.EncodeToString(combined)).
		Str("fingerprint", fingerprint).
		Msg("Generated fingerprint")

	return fingerprint
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

// GenerateTestRoleKey creates a role-based key with a fixed expiration time for testing
func (ibe *IBESystem) GenerateTestRoleKey(role string, scope string) []byte {
	// Use a fixed expiration time for consistent testing
	fixedExpiration := time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC)
	return ibe.GenerateRoleKey(role, scope, fixedExpiration)
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

// NewIBESystemFromConfig creates a new IBE system from configuration
func NewIBESystemFromConfig(masterKeyPath string, keyVersion int, salt string) (*IBESystem, error) {
	opts := IBEOptions{
		KeyVersion: keyVersion,
		Salt:       salt,
	}

	// Try to load master secret from file if path is provided
	if masterKeyPath != "" {
		masterSecret, err := loadMasterSecretFromFile(masterKeyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load master secret from %s: %w", masterKeyPath, err)
		}
		opts.MasterSecret = masterSecret
	}

	return NewIBESystemWithOptions(opts), nil
}

// loadMasterSecretFromFile loads a master secret from a file
func loadMasterSecretFromFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expect hex-encoded 32-byte secret
	if len(data) != 64 { // 32 bytes = 64 hex chars
		return nil, fmt.Errorf("master secret file must contain exactly 64 hex characters")
	}

	secret, err := hex.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("invalid hex encoding in master secret file: %w", err)
	}

	return secret, nil
}

// SaveMasterSecretToFile saves the master secret to a file
func (ibe *IBESystem) SaveMasterSecretToFile(path string) error {
	secret := ibe.GetMasterSecret()
	hexSecret := hex.EncodeToString(secret)
	return os.WriteFile(path, []byte(hexSecret), 0600) // Read/write for owner only
}

// NewIBESystemFromEnv creates a new IBE system from environment variables
func NewIBESystemFromEnv() *IBESystem {
	masterKeyPath := os.Getenv("IBE_MASTER_KEY_PATH")
	if masterKeyPath == "" {
		masterKeyPath = "./keys/master.key"
	}
	keyVersion := 1
	if v := os.Getenv("IBE_KEY_VERSION"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil {
			keyVersion = parsed
		}
	}
	salt := os.Getenv("IBE_SALT")
	if salt == "" {
		salt = "hashpost_fingerprint_salt_v1"
	}
	ibeSystem, err := NewIBESystemFromConfig(masterKeyPath, keyVersion, salt)
	if err != nil {
		panic("Failed to create IBE system from environment: " + err.Error())
	}
	return ibeSystem
}
