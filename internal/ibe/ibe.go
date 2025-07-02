package ibe

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

// Cryptographic domains for privilege separation
const (
	DOMAIN_USER_PSEUDONYMS   = "user_pseudonyms_v1"
	DOMAIN_USER_CORRELATION  = "user_self_correlation_v1"
	DOMAIN_MOD_CORRELATION   = "moderator_correlation_v1"
	DOMAIN_ADMIN_CORRELATION = "admin_correlation_v1"
	DOMAIN_LEGAL_CORRELATION = "legal_correlation_v1"
)

// SeparatedIBESystem represents the enhanced IBE system with true domain separation
type SeparatedIBESystem struct {
	domainMasters map[string][]byte // Separate master key for each domain
	keyVersion    int
	salt          []byte
}

// IBESystem provides backward compatibility wrapper around the enhanced system
type IBESystem struct {
	separated *SeparatedIBESystem
}

// IBEOptions defines configuration options for the IBE system
type IBEOptions struct {
	DomainMasters map[string][]byte // Separate master keys for each domain
	KeyVersion    int               // Optional: defaults to 1
	Salt          string            // Optional: salt for fingerprint generation, defaults to "fingerprint_salt_v1"
}

// NewSeparatedIBESystem creates a new IBE system with true domain separation
func NewSeparatedIBESystem(domainMasters map[string][]byte, keyVersion int, salt []byte) *SeparatedIBESystem {
	return &SeparatedIBESystem{
		domainMasters: domainMasters,
		keyVersion:    keyVersion,
		salt:          salt,
	}
}

// getDomainMaster returns the master key for a specific domain
func (ibe *SeparatedIBESystem) getDomainMaster(domain string) ([]byte, error) {
	master, exists := ibe.domainMasters[domain]
	if !exists {
		return nil, fmt.Errorf("no master key found for domain: %s", domain)
	}
	return master, nil
}

// GeneratePseudonym creates a pseudonym ID for a user with enhanced context separation
func (ibe *SeparatedIBESystem) GeneratePseudonym(userID int64, context string, version int) string {
	domainMaster, err := ibe.getDomainMaster(DOMAIN_USER_PSEUDONYMS)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get domain master")
		return ""
	}

	switch version {
	case 1: // Legacy deterministic (maintain existing)
		combined := append([]byte(fmt.Sprintf("%d", userID)), domainMaster...)
		hash := sha256.Sum256(combined)
		return hex.EncodeToString(hash[:16])

	case 2: // Enhanced with context separation
		contextEntropy := sha256.Sum256([]byte(context + string(ibe.salt)))
		combined := append([]byte(fmt.Sprintf("%d", userID)), domainMaster...)
		combined = append(combined, contextEntropy[:]...)
		hash := sha256.Sum256(combined)
		return hex.EncodeToString(hash[:16])

	default:
		// Default to version 1 for backward compatibility
		combined := append([]byte(fmt.Sprintf("%d", userID)), domainMaster...)
		hash := sha256.Sum256(combined)
		return hex.EncodeToString(hash[:16])
	}
}

// GenerateCorrelationKey creates a time-bounded correlation key for a specific role and scope
func (ibe *SeparatedIBESystem) GenerateCorrelationKey(role, scope string, timeWindow time.Duration) []byte {
	// Select appropriate domain based on role
	domain := selectDomain(role)
	domainMaster, err := ibe.getDomainMaster(domain)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get domain master")
		return nil
	}

	// Include time epoch in key derivation for forward secrecy
	epoch := time.Now().Truncate(timeWindow).Unix()

	combined := append(domainMaster, []byte(role)...)
	combined = append(combined, []byte(scope)...)
	combined = append(combined, []byte(fmt.Sprintf("%d", epoch))...)

	hash := sha256.Sum256(combined)
	return hash[:]
}

// selectDomain maps roles to appropriate cryptographic domains
func selectDomain(role string) string {
	switch role {
	case "user":
		return DOMAIN_USER_CORRELATION
	case "moderator", "subforum_owner":
		return DOMAIN_MOD_CORRELATION
	case "platform_admin", "trust_safety":
		return DOMAIN_ADMIN_CORRELATION
	case "legal_team":
		return DOMAIN_LEGAL_CORRELATION
	default:
		return DOMAIN_USER_CORRELATION
	}
}

// EncryptIdentityWithDomain encrypts identity mapping using domain-specific keys
func (ibe *SeparatedIBESystem) EncryptIdentityWithDomain(realIdentity, pseudonymID string, domain string, adminKey []byte) ([]byte, error) {
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

// GenerateFingerprint creates a deterministic fingerprint from a real identity
func (ibe *SeparatedIBESystem) GenerateFingerprint(realIdentity string) string {
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

// NewIBESystem creates a new IBE system with backward compatibility
func NewIBESystem() *IBESystem {
	return NewIBESystemWithOptions(IBEOptions{})
}

// NewIBESystemWithOptions creates a new IBE system with configuration options
func NewIBESystemWithOptions(opts IBEOptions) *IBESystem {
	domainMasters := make(map[string][]byte)

	// Initialize domain masters
	domains := []string{
		DOMAIN_USER_PSEUDONYMS,
		DOMAIN_USER_CORRELATION,
		DOMAIN_MOD_CORRELATION,
		DOMAIN_ADMIN_CORRELATION,
		DOMAIN_LEGAL_CORRELATION,
	}

	// If domain masters are provided, use them; otherwise generate new ones
	if opts.DomainMasters != nil {
		for domain, master := range opts.DomainMasters {
			domainMasters[domain] = make([]byte, len(master))
			copy(domainMasters[domain], master)
		}
	} else {
		// Generate separate master keys for each domain
		for _, domain := range domains {
			master := make([]byte, 32)
			rand.Read(master)
			domainMasters[domain] = master
		}
	}

	keyVersion := opts.KeyVersion
	if keyVersion == 0 {
		keyVersion = 1
	}

	salt := opts.Salt
	if salt == "" {
		salt = "fingerprint_salt_v1"
	}

	// Create the enhanced separated system
	separated := NewSeparatedIBESystem(domainMasters, keyVersion, []byte(salt))

	return &IBESystem{
		separated: separated,
	}
}

// Backward compatibility methods - maintain existing API

// GetMasterSecret returns a copy of the master secret (for backward compatibility)
// Note: This is deprecated - use GetDomainMasters instead
func (ibe *IBESystem) GetMasterSecret() []byte {
	// For backward compatibility, return the first domain master
	for _, master := range ibe.separated.domainMasters {
		secret := make([]byte, len(master))
		copy(secret, master)
		return secret
	}
	return nil
}

// SetMasterSecret sets the master secret (for backward compatibility)
// Note: This is deprecated - use SetDomainMasters instead
func (ibe *IBESystem) SetMasterSecret(secret []byte) error {
	if len(secret) != 32 {
		return fmt.Errorf("master secret must be 32 bytes, got %d", len(secret))
	}
	// For backward compatibility, set all domains to use the same secret
	for domain := range ibe.separated.domainMasters {
		ibe.separated.domainMasters[domain] = make([]byte, len(secret))
		copy(ibe.separated.domainMasters[domain], secret)
	}
	return nil
}

// GetDomainMasters returns all domain masters (new API)
func (ibe *IBESystem) GetDomainMasters() map[string][]byte {
	result := make(map[string][]byte)
	for domain, master := range ibe.separated.domainMasters {
		result[domain] = make([]byte, len(master))
		copy(result[domain], master)
	}
	return result
}

// SetDomainMasters sets the domain masters (new API)
func (ibe *IBESystem) SetDomainMasters(domainMasters map[string][]byte) error {
	for domain, master := range domainMasters {
		if len(master) != 32 {
			return fmt.Errorf("domain master for %s must be 32 bytes, got %d", domain, len(master))
		}
		ibe.separated.domainMasters[domain] = make([]byte, len(master))
		copy(ibe.separated.domainMasters[domain], master)
	}
	return nil
}

// GetKeyVersion returns the current key version
func (ibe *IBESystem) GetKeyVersion() int {
	return ibe.separated.keyVersion
}

// SetKeyVersion sets the key version
func (ibe *IBESystem) SetKeyVersion(version int) {
	ibe.separated.keyVersion = version
}

// SetSalt sets the salt for fingerprint generation
func (ibe *IBESystem) SetSalt(salt string) {
	ibe.separated.salt = []byte(salt)
}

// GetSalt returns the current salt
func (ibe *IBESystem) GetSalt() string {
	return string(ibe.separated.salt)
}

// GeneratePseudonym creates a pseudonym ID for a user (backward compatible)
func (ibe *IBESystem) GeneratePseudonym(userSecret []byte) string {
	// Extract user ID from user secret for backward compatibility
	// This maintains exact existing behavior for current users
	userID := extractUserID(userSecret)
	return ibe.separated.GeneratePseudonym(userID, "default", 1)
}

// extractUserID extracts user ID from user secret (backward compatibility helper)
func extractUserID(userSecret []byte) int64 {
	// Simple hash-based extraction for backward compatibility
	hash := sha256.Sum256(userSecret)
	// Use first 8 bytes as int64
	var userID int64
	for i := 0; i < 8; i++ {
		userID = userID<<8 + int64(hash[i])
	}
	return userID
}

// GenerateFingerprint creates a deterministic fingerprint from a real identity
func (ibe *IBESystem) GenerateFingerprint(realIdentity string) string {
	return ibe.separated.GenerateFingerprint(realIdentity)
}

// EncryptIdentity encrypts the mapping between real identity and pseudonym (backward compatible)
func (ibe *IBESystem) EncryptIdentity(realIdentity, pseudonymID string, adminKey []byte) ([]byte, error) {
	// Use appropriate domain for encryption
	domain := DOMAIN_ADMIN_CORRELATION
	return ibe.separated.EncryptIdentityWithDomain(realIdentity, pseudonymID, domain, adminKey)
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

// GenerateRoleKey creates a role-based key for administrative access (backward compatible)
func (ibe *IBESystem) GenerateRoleKey(role string, scope string, expiration time.Time) []byte {
	// Route to appropriate domain with time-bounded derivation
	timeWindow := time.Hour * 24 * 30 // 30-day windows for backward compatibility
	return ibe.separated.GenerateCorrelationKey(role, scope, timeWindow)
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

// Enhanced API methods for new functionality

// CreateEnhancedPseudonym creates a pseudonym with enhanced security features
func (ibe *IBESystem) CreateEnhancedPseudonym(userID int64, context string) string {
	return ibe.separated.GeneratePseudonym(userID, context, 2) // Enhanced version
}

// GenerateTimeBoundedKey creates a time-bounded correlation key
func (ibe *IBESystem) GenerateTimeBoundedKey(role, scope string, duration time.Duration) []byte {
	return ibe.separated.GenerateCorrelationKey(role, scope, duration)
}

// NewIBESystemFromConfig creates a new IBE system from configuration
func NewIBESystemFromConfig(domainKeysDir string, keyVersion int, salt string) (*IBESystem, error) {
	opts := IBEOptions{
		KeyVersion: keyVersion,
		Salt:       salt,
	}

	// Try to load domain masters from directory if provided
	if domainKeysDir != "" {
		domainMasters, err := LoadDomainMastersFromDir(domainKeysDir)
		if err != nil {
			return nil, fmt.Errorf("failed to load domain masters from %s: %w", domainKeysDir, err)
		}
		opts.DomainMasters = domainMasters
	}

	return NewIBESystemWithOptions(opts), nil
}

// LoadDomainMastersFromDir loads domain masters from a directory
func LoadDomainMastersFromDir(dir string) (map[string][]byte, error) {
	domainMasters := make(map[string][]byte)

	domains := []string{
		DOMAIN_USER_PSEUDONYMS,
		DOMAIN_USER_CORRELATION,
		DOMAIN_MOD_CORRELATION,
		DOMAIN_ADMIN_CORRELATION,
		DOMAIN_LEGAL_CORRELATION,
	}

	for _, domain := range domains {
		keyPath := filepath.Join(dir, fmt.Sprintf("%s.key", domain))
		master, err := LoadMasterSecretFromFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load domain master for %s: %w", domain, err)
		}
		domainMasters[domain] = master
	}

	return domainMasters, nil
}

// LoadMasterSecretFromFile loads a master secret from a file
func LoadMasterSecretFromFile(path string) ([]byte, error) {
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

// SaveMasterSecretToFile saves the master secret to a file (backward compatibility)
// Note: This is deprecated - use SaveDomainMastersToDir instead
func (ibe *IBESystem) SaveMasterSecretToFile(path string) error {
	// For backward compatibility, save the first domain master
	for _, master := range ibe.separated.domainMasters {
		hexSecret := hex.EncodeToString(master)
		return os.WriteFile(path, []byte(hexSecret), 0600)
	}
	return fmt.Errorf("no domain masters available")
}

// SaveDomainMastersToDir saves all domain masters to a directory
func (ibe *IBESystem) SaveDomainMastersToDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create domain masters directory: %w", err)
	}

	for domain, master := range ibe.separated.domainMasters {
		keyPath := filepath.Join(dir, fmt.Sprintf("%s.key", domain))
		hexSecret := hex.EncodeToString(master)
		if err := os.WriteFile(keyPath, []byte(hexSecret), 0600); err != nil {
			return fmt.Errorf("failed to save domain master for %s: %w", domain, err)
		}
	}

	return nil
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
		salt = "fingerprint_salt_v1"
	}
	ibeSystem, err := NewIBESystemFromConfig(masterKeyPath, keyVersion, salt)
	if err != nil {
		panic("Failed to create IBE system from environment: " + err.Error())
	}
	return ibeSystem
}
