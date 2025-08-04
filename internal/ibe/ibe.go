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
	"strings"
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

// KeyVersionInfo represents a specific key version with its domain keys
type KeyVersionInfo struct {
	Version      int
	DomainKeys   map[string][]byte
	Salt         string
	CreatedAt    time.Time
	DeprecatedAt *time.Time
	IsActive     bool
	IsDeprecated bool
}

// KeyVersionRegistry manages multiple key versions for migration scenarios
type KeyVersionRegistry struct {
	CurrentVersion     int
	ActiveVersions     map[int]*KeyVersionInfo
	DeprecatedVersions map[int]*KeyVersionInfo
	MigrationMode      bool
}

// NewKeyVersionRegistry creates a new key version registry
func NewKeyVersionRegistry() *KeyVersionRegistry {
	return &KeyVersionRegistry{
		CurrentVersion:     1,
		ActiveVersions:     make(map[int]*KeyVersionInfo),
		DeprecatedVersions: make(map[int]*KeyVersionInfo),
		MigrationMode:      false,
	}
}

// AddKeyVersion adds a new key version to the registry
func (r *KeyVersionRegistry) AddKeyVersion(version int, domainKeys map[string][]byte, salt string) {
	keyInfo := &KeyVersionInfo{
		Version:      version,
		DomainKeys:   domainKeys,
		Salt:         salt,
		CreatedAt:    time.Now(),
		IsActive:     true,
		IsDeprecated: false,
	}

	r.ActiveVersions[version] = keyInfo
}

// DeprecateKeyVersion marks a key version as deprecated (for migration)
func (r *KeyVersionRegistry) DeprecateKeyVersion(version int) error {
	keyInfo, exists := r.ActiveVersions[version]
	if !exists {
		return fmt.Errorf("key version %d not found", version)
	}

	now := time.Now()
	keyInfo.DeprecatedAt = &now
	keyInfo.IsDeprecated = true
	keyInfo.IsActive = false

	// Move to deprecated versions
	r.DeprecatedVersions[version] = keyInfo
	delete(r.ActiveVersions, version)

	return nil
}

// GetKeyVersionInfo retrieves key information for a specific version
func (r *KeyVersionRegistry) GetKeyVersionInfo(version int) (*KeyVersionInfo, error) {
	// Check active versions first
	if keyInfo, exists := r.ActiveVersions[version]; exists {
		return keyInfo, nil
	}

	// Check deprecated versions
	if keyInfo, exists := r.DeprecatedVersions[version]; exists {
		return keyInfo, nil
	}

	return nil, fmt.Errorf("key version %d not found", version)
}

// GetDomainKeyForVersion gets the domain key for a specific version and domain
func (r *KeyVersionRegistry) GetDomainKeyForVersion(version int, domain string) ([]byte, error) {
	keyInfo, err := r.GetKeyVersionInfo(version)
	if err != nil {
		return nil, err
	}

	domainKey, exists := keyInfo.DomainKeys[domain]
	if !exists {
		return nil, fmt.Errorf("no key found for domain %s in version %d", domain, version)
	}

	return domainKey, nil
}

// EnableMigrationMode enables migration mode where both old and new keys are available
func (r *KeyVersionRegistry) EnableMigrationMode() {
	r.MigrationMode = true
}

// DisableMigrationMode disables migration mode
func (r *KeyVersionRegistry) DisableMigrationMode() {
	r.MigrationMode = false
}

// IBESystem represents the enhanced Identity-Based Encryption (IBE) system with true domain separation.
//
// Purpose:
// The IBESystem struct is designed to provide cryptographic domain separation and multi-version key management
// for secure and flexible encryption. It ensures that each domain operates independently with its own master key.
//
// Key Management Approach:
//   - `domainMasters`: A map where each domain is associated with a unique master key. This enables strict separation
//     of cryptographic operations across domains.
//   - `keyVersion`: Indicates the current version of the keys being used. This allows for key rotation and versioning.
//   - `keyRegistry`: A registry that supports multiple key versions, enabling seamless migration between key versions
//     and ensuring backward compatibility.
//
// Relationship Between Domain Masters and Key Versions:
// Each domain has its own master key stored in `domainMasters`. The `keyVersion` field specifies which version of
// the keys is currently active. The `keyRegistry` provides additional support for managing multiple versions of keys,
// allowing the system to operate in migration mode where both old and new keys are available.
type IBESystem struct {
	domainMasters map[string][]byte // Separate master key for each domain
	keyVersion    int
	salt          []byte
	keyRegistry   *KeyVersionRegistry // Multi-version key support
}

// IBEOptions defines configuration options for the IBE system
type IBEOptions struct {
	DomainMasters map[string][]byte   // Separate master keys for each domain
	KeyVersion    int                 // Optional: defaults to 1
	Salt          string              // Optional: salt for fingerprint generation, defaults to "fingerprint_salt_v1"
	KeyRegistry   *KeyVersionRegistry // Optional: multi-version key registry
}

// NewIBESystem creates a new IBE system with true domain separation
func NewIBESystem(domainMasters map[string][]byte, keyVersion int, salt []byte) *IBESystem {
	return &IBESystem{
		domainMasters: domainMasters,
		keyVersion:    keyVersion,
		salt:          salt,
		keyRegistry:   NewKeyVersionRegistry(),
	}
}

// getDomainMaster returns the master key for a specific domain
func (ibe *IBESystem) getDomainMaster(domain string) ([]byte, error) {
	master, exists := ibe.domainMasters[domain]
	if !exists {
		return nil, fmt.Errorf("no master key found for domain: %s", domain)
	}
	return master, nil
}

// getDomainMasterForVersion gets the master key for a specific domain and version
func (ibe *IBESystem) getDomainMasterForVersion(domain string, version int) ([]byte, error) {
	// If we have a key registry with multiple versions, use it
	if ibe.keyRegistry != nil && ibe.keyRegistry.MigrationMode {
		return ibe.keyRegistry.GetDomainKeyForVersion(version, domain)
	}

	// Fall back to current domain masters
	return ibe.getDomainMaster(domain)
}

// GeneratePseudonym creates a pseudonym ID for a user with enhanced context separation
func (ibe *IBESystem) GeneratePseudonym(userID int64, context string, version int) string {
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
func (ibe *IBESystem) GenerateCorrelationKey(role, scope string, timeWindow time.Duration) []byte {
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
	// Include time window in key derivation to ensure different windows produce different keys
	combined = append(combined, []byte(fmt.Sprintf("%d", timeWindow.Nanoseconds()))...)

	hash := sha256.Sum256(combined)
	return hash[:]
}

// GenerateCorrelationKeyForVersion creates a time-bounded correlation key for a specific version
func (ibe *IBESystem) GenerateCorrelationKeyForVersion(role, scope string, timeWindow time.Duration, version int) []byte {
	// Select appropriate domain based on role
	domain := selectDomain(role)
	domainMaster, err := ibe.getDomainMasterForVersion(domain, version)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get domain master for version")
		return nil
	}

	// Include time epoch in key derivation for forward secrecy
	epoch := time.Now().Truncate(timeWindow).Unix()

	combined := append(domainMaster, []byte(role)...)
	combined = append(combined, []byte(scope)...)
	combined = append(combined, []byte(fmt.Sprintf("%d", epoch))...)
	// Include time window in key derivation to ensure different windows produce different keys
	combined = append(combined, []byte(fmt.Sprintf("%d", timeWindow.Nanoseconds()))...)

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

// EncryptIdentityWithDomain encrypts identity mapping with domain-specific key
func (ibe *IBESystem) EncryptIdentityWithDomain(realIdentity, pseudonymID string, domain string, adminKey []byte) ([]byte, error) {
	// Use AES-GCM for authenticated encryption
	key := sha256.Sum256(adminKey)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create fingerprint for privacy protection
	fingerprint := ibe.GenerateFingerprint(realIdentity)

	// Create mapping data with fingerprint, not real identity
	mapping := fmt.Sprintf("%s:%s", fingerprint, pseudonymID)
	plaintext := []byte(mapping)

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// EncryptIdentityWithVersion encrypts identity mapping with version-specific key
func (ibe *IBESystem) EncryptIdentityWithVersion(realIdentity, pseudonymID string, domain string, version int) ([]byte, error) {
	// Get domain key for specific version
	domainKey, err := ibe.getDomainMasterForVersion(domain, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain key for version %d: %w", version, err)
	}

	return ibe.EncryptIdentityWithDomain(realIdentity, pseudonymID, domain, domainKey)
}

// DecryptIdentityWithVersion decrypts identity mapping with version-specific key
func (ibe *IBESystem) DecryptIdentityWithVersion(encryptedData []byte, domain string, version int) (string, string, error) {
	// Get domain key for specific version
	domainKey, err := ibe.getDomainMasterForVersion(domain, version)
	if err != nil {
		return "", "", fmt.Errorf("failed to get domain key for version %d: %w", version, err)
	}

	return ibe.DecryptIdentityWithDomain(encryptedData, domainKey)
}

// DecryptIdentityWithDomain decrypts identity mapping with domain-specific key
func (ibe *IBESystem) DecryptIdentityWithDomain(encryptedData []byte, adminKey []byte) (string, string, error) {
	// Use AES-GCM for authenticated decryption
	key := sha256.Sum256(adminKey)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", "", fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extract nonce
	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return "", "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to decrypt: %w", err)
	}

	// Parse mapping (contains fingerprint:pseudonym_id)
	mapping := string(plaintext)
	parts := strings.Split(mapping, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid mapping format")
	}

	// Return fingerprint and pseudonym_id
	return parts[0], parts[1], nil
}

// GenerateFingerprint creates a fingerprint for real identity
func (ibe *IBESystem) GenerateFingerprint(realIdentity string) string {
	// Create fingerprint using salt
	combined := append([]byte(realIdentity), ibe.salt...)
	hash := sha256.Sum256(combined)
	return hex.EncodeToString(hash[:16])
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

	// Create the enhanced system
	ibe := NewIBESystem(domainMasters, keyVersion, []byte(salt))

	// If a key registry is provided, use it
	if opts.KeyRegistry != nil {
		ibe.keyRegistry = opts.KeyRegistry
	}

	return ibe
}

// Backward compatibility methods - maintain existing API

// GetMasterSecret returns a copy of the master secret (for backward compatibility)
// Note: This is deprecated - use GetDomainMasters instead
func (ibe *IBESystem) GetMasterSecret() []byte {
	// For backward compatibility, return the first domain master
	for _, master := range ibe.domainMasters {
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
	for domain := range ibe.domainMasters {
		ibe.domainMasters[domain] = make([]byte, len(secret))
		copy(ibe.domainMasters[domain], secret)
	}
	return nil
}

// GetDomainMasters returns all domain masters (new API)
func (ibe *IBESystem) GetDomainMasters() map[string][]byte {
	result := make(map[string][]byte)
	for domain, master := range ibe.domainMasters {
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
		ibe.domainMasters[domain] = make([]byte, len(master))
		copy(ibe.domainMasters[domain], master)
	}
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

// Multi-version key management methods

// GetKeyRegistry returns the key version registry
func (ibe *IBESystem) GetKeyRegistry() *KeyVersionRegistry {
	return ibe.keyRegistry
}

// EnableMigrationMode enables migration mode where both old and new keys are available
func (ibe *IBESystem) EnableMigrationMode() {
	if ibe.keyRegistry != nil {
		ibe.keyRegistry.EnableMigrationMode()
	}
}

// DisableMigrationMode disables migration mode
func (ibe *IBESystem) DisableMigrationMode() {
	if ibe.keyRegistry != nil {
		ibe.keyRegistry.DisableMigrationMode()
	}
}

// AddKeyVersion adds a new key version to the registry
func (ibe *IBESystem) AddKeyVersion(version int, domainKeys map[string][]byte, salt string) {
	if ibe.keyRegistry != nil {
		ibe.keyRegistry.AddKeyVersion(version, domainKeys, salt)
	}
}

// DeprecateKeyVersion marks a key version as deprecated (for migration)
func (ibe *IBESystem) DeprecateKeyVersion(version int) error {
	if ibe.keyRegistry != nil {
		return ibe.keyRegistry.DeprecateKeyVersion(version)
	}
	return fmt.Errorf("key registry not available")
}

// GeneratePseudonymFromUserSecret creates a pseudonym ID for a user from a user secret (backward compatible)
func (ibe *IBESystem) GeneratePseudonymFromUserSecret(userSecret []byte) string {
	// Extract user ID from user secret for backward compatibility
	// This maintains exact existing behavior for current users
	userID := extractUserID(userSecret)
	return ibe.GeneratePseudonym(userID, "default", 1)
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

// EncryptIdentity encrypts the mapping between real identity and pseudonym (backward compatible)
func (ibe *IBESystem) EncryptIdentity(realIdentity, pseudonymID string, adminKey []byte) ([]byte, error) {
	// Use appropriate domain for encryption
	domain := DOMAIN_ADMIN_CORRELATION
	return ibe.EncryptIdentityWithDomain(realIdentity, pseudonymID, domain, adminKey)
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
	return ibe.GenerateCorrelationKey(role, scope, timeWindow)
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
	return ibe.GeneratePseudonym(userID, context, 2) // Enhanced version
}

// GenerateTimeBoundedKey creates a time-bounded correlation key
func (ibe *IBESystem) GenerateTimeBoundedKey(role, scope string, duration time.Duration) []byte {
	return ibe.GenerateCorrelationKey(role, scope, duration)
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
		data, err := os.ReadFile(keyPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load domain master for %s: %w", domain, err)
		}

		// Expect hex-encoded 32-byte secret
		if len(data) != 64 { // 32 bytes = 64 hex chars
			return nil, fmt.Errorf("domain master file %s must contain exactly 64 hex characters", domain)
		}

		secret, err := hex.DecodeString(string(data))
		if err != nil {
			return nil, fmt.Errorf("invalid hex encoding in domain master file %s: %w", domain, err)
		}
		domainMasters[domain] = secret
	}

	return domainMasters, nil
}

// SaveDomainMastersToDir saves all domain masters to a directory
func (ibe *IBESystem) SaveDomainMastersToDir(dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create domain masters directory: %w", err)
	}

	for domain, master := range ibe.domainMasters {
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
	domainKeysDir := os.Getenv("IBE_DOMAIN_KEYS_DIR")
	if domainKeysDir == "" {
		domainKeysDir = "./keys/domains"
	}
	var keyVersion int32 = 1
	if v := os.Getenv("IBE_KEY_VERSION"); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 32); err == nil {
			keyVersion = int32(parsed)
		}
	}
	salt := os.Getenv("IBE_SALT")
	if salt == "" {
		salt = "fingerprint_salt_v1"
	}
	ibeSystem, err := NewIBESystemFromConfig(domainKeysDir, int(keyVersion), salt)
	if err != nil {
		panic("Failed to create IBE system from environment: " + err.Error())
	}
	return ibeSystem
}

// createDefaultIBEOptions creates default IBE options for testing
func createDefaultIBEOptions() IBEOptions {
	return IBEOptions{
		DomainMasters: map[string][]byte{
			DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
			DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
			DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
			DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
		},
		KeyVersion: 1,
		Salt:       "test_fingerprint_salt_v1",
	}
}

// NewTestIBESystem creates a new IBE system with test configuration
func NewTestIBESystem() *IBESystem {
	return NewIBESystemWithOptions(createDefaultIBEOptions())
}

// EncryptFingerprintMapping encrypts a fingerprint-to-pseudonym mapping directly
func (ibe *IBESystem) EncryptFingerprintMapping(fingerprint, pseudonymID string, domain string, version int) ([]byte, error) {
	// Get domain key for specific version
	domainKey, err := ibe.getDomainMasterForVersion(domain, version)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain key for version %d: %w", version, err)
	}

	// Use AES-GCM for authenticated encryption
	key := sha256.Sum256(domainKey)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Create mapping data with fingerprint directly (no additional hashing)
	mapping := fmt.Sprintf("%s:%s", fingerprint, pseudonymID)
	plaintext := []byte(mapping)

	// Generate nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt
	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}
