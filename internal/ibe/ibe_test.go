package ibe

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestIBESystem creates an IBE system with consistent test configuration
func createTestIBESystem() *IBESystem {
	return NewIBESystemWithOptions(IBEOptions{
		DomainMasters: map[string][]byte{
			DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
			DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
			DOMAIN_USER_MESSAGING:    []byte("0123456789abcdef0123456789abcdef"),
			DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
			DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
		},
		KeyVersion: 1,
		Salt:       "test_fingerprint_salt_v1",
	})
}

func TestIBESystem_EnhancedArchitecture(t *testing.T) {
	// Create a temporary directory for test keys
	tempDir := t.TempDir()

	t.Run("Domain Separation", func(t *testing.T) {
		// Create IBE system with test configuration
		ibeSystem := createTestIBESystem()

		// Test that different roles get different domains
		userKey := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
		modKey := ibeSystem.GenerateTimeBoundedKey("moderator", "correlation", time.Hour)
		adminKey := ibeSystem.GenerateTimeBoundedKey("platform_admin", "correlation", time.Hour)
		legalKey := ibeSystem.GenerateTimeBoundedKey("legal_team", "correlation", time.Hour)

		// All keys should be different (different domains)
		assert.False(t, bytes.Equal(userKey, modKey), "User and moderator keys should be different")
		assert.False(t, bytes.Equal(userKey, adminKey), "User and admin keys should be different")
		assert.False(t, bytes.Equal(userKey, legalKey), "User and legal keys should be different")
		assert.False(t, bytes.Equal(modKey, adminKey), "Moderator and admin keys should be different")
		assert.False(t, bytes.Equal(modKey, legalKey), "Moderator and legal keys should be different")
		assert.False(t, bytes.Equal(adminKey, legalKey), "Admin and legal keys should be different")

		// Test that same role gets same key in same time window
		userKey2 := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
		assert.True(t, bytes.Equal(userKey, userKey2), "Same role should get same key in same time window")
	})

	t.Run("Time Bounded Keys", func(t *testing.T) {
		// Create IBE system
		ibeSystem := createTestIBESystem()

		// Test different time windows
		hourKey := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
		dayKey := ibeSystem.GenerateTimeBoundedKey("user", "correlation", 24*time.Hour)
		weekKey := ibeSystem.GenerateTimeBoundedKey("user", "correlation", 7*24*time.Hour)

		// Keys should be different for different time windows
		assert.False(t, bytes.Equal(hourKey, dayKey), "Hour and day keys should be different")
		assert.False(t, bytes.Equal(hourKey, weekKey), "Hour and week keys should be different")
		assert.False(t, bytes.Equal(dayKey, weekKey), "Day and week keys should be different")
	})

	t.Run("Enhanced Pseudonyms", func(t *testing.T) {
		// Create IBE system
		ibeSystem := createTestIBESystem()

		// Test enhanced pseudonym generation
		pseudonym1 := ibeSystem.CreateEnhancedPseudonym(1, "test_context_1")
		pseudonym2 := ibeSystem.CreateEnhancedPseudonym(1, "test_context_2")
		pseudonym3 := ibeSystem.CreateEnhancedPseudonym(2, "test_context_1")

		// Different contexts should generate different pseudonyms
		assert.NotEqual(t, pseudonym1, pseudonym2, "Different contexts should generate different pseudonyms")
		assert.NotEqual(t, pseudonym1, pseudonym3, "Different user IDs should generate different pseudonyms")
		assert.NotEqual(t, pseudonym2, pseudonym3, "Different user IDs and contexts should generate different pseudonyms")

		// Same user ID and context should generate same pseudonym
		pseudonym1Again := ibeSystem.CreateEnhancedPseudonym(1, "test_context_1")
		assert.Equal(t, pseudonym1, pseudonym1Again, "Same user ID and context should generate same pseudonym")
	})

	t.Run("Domain Key Generation", func(t *testing.T) {
		// Create IBE system
		ibeSystem := createTestIBESystem()

		// Test domain key generation and saving
		domainKeysDir := filepath.Join(tempDir, "domains")
		err := ibeSystem.SaveDomainMastersToDir(domainKeysDir)
		require.NoError(t, err, "Should save domain master keys to directory")

		// Verify domain key files exist and have correct permissions
		domains := []string{
			DOMAIN_USER_PSEUDONYMS,
			DOMAIN_USER_CORRELATION,
			DOMAIN_MOD_CORRELATION,
			DOMAIN_ADMIN_CORRELATION,
			DOMAIN_LEGAL_CORRELATION,
		}

		for _, domain := range domains {
			keyPath := filepath.Join(domainKeysDir, fmt.Sprintf("%s.key", domain))
			info, err := os.Stat(keyPath)
			require.NoError(t, err, "Should be able to stat domain key file for %s", domain)
			assert.Equal(t, os.FileMode(0600), info.Mode()&0777, "Domain key file should have 600 permissions")

			// Load domain key from file
			data, err := os.ReadFile(keyPath)
			require.NoError(t, err, "Should read domain key file for %s", domain)

			// Expect hex-encoded 32-byte secret
			require.Len(t, data, 64, "Domain key file should contain exactly 64 hex characters")

			domainKeyBytes, err := hex.DecodeString(string(data))
			require.NoError(t, err, "Should decode domain key for %s", domain)
			assert.Len(t, domainKeyBytes, 32, "Domain key should be 32 bytes")
		}

		// Load domain masters from directory
		loadedDomainMasters, err := LoadDomainMastersFromDir(domainKeysDir)
		require.NoError(t, err, "Should load domain masters from directory")

		// Create new IBE system with loaded domain masters
		loadedIBE := NewIBESystemWithOptions(IBEOptions{
			DomainMasters: loadedDomainMasters,
			KeyVersion:    1,
			Salt:          "test_fingerprint_salt_v1",
		})

		// Verify both systems generate same keys
		originalKey := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
		loadedKey := loadedIBE.GenerateTimeBoundedKey("user", "correlation", time.Hour)

		assert.True(t, bytes.Equal(originalKey, loadedKey), "Keys should be identical after loading from directory")
	})

	t.Run("Domain Key Generation", func(t *testing.T) {
		// Create IBE system
		ibeSystem := createTestIBESystem()

		// Test domain key generation
		domains := []string{
			DOMAIN_USER_PSEUDONYMS,
			DOMAIN_USER_CORRELATION,
			DOMAIN_MOD_CORRELATION,
			DOMAIN_ADMIN_CORRELATION,
			DOMAIN_LEGAL_CORRELATION,
		}

		domainKeys := make(map[string][]byte)
		for _, domain := range domains {
			// Generate domain key (this would be implemented in the IBE system)
			// For now, we'll test that we can generate keys for different roles
			key := ibeSystem.GenerateTimeBoundedKey("user", domain, time.Hour)
			domainKeys[domain] = key
		}

		// All domain keys should be different
		seenKeys := make(map[string]bool)
		for domain, key := range domainKeys {
			keyHex := hex.EncodeToString(key)
			assert.False(t, seenKeys[keyHex], "Domain %s should have unique key", domain)
			seenKeys[keyHex] = true
		}
	})

	t.Run("Role Key Validation", func(t *testing.T) {
		// Create IBE system
		ibeSystem := createTestIBESystem()

		// Test role key generation and validation
		roles := []string{"user", "moderator", "platform_admin", "trust_safety", "legal_team"}
		scopes := []string{"authentication", "correlation"}

		for _, role := range roles {
			for _, scope := range scopes {
				// Generate role key
				roleKey := ibeSystem.GenerateTimeBoundedKey(role, scope, time.Hour)

				// Validate role key (this would be implemented in the IBE system)
				// For now, we'll just verify the key is not empty
				assert.NotEmpty(t, roleKey, "Role key for %s:%s should not be empty", role, scope)
				assert.Len(t, roleKey, 32, "Role key should be 32 bytes")

				// Test that same role/scope/time generates same key
				roleKey2 := ibeSystem.GenerateTimeBoundedKey(role, scope, time.Hour)
				assert.True(t, bytes.Equal(roleKey, roleKey2), "Same role/scope/time should generate same key")
			}
		}
	})

	t.Run("Backward Compatibility", func(t *testing.T) {
		// Create IBE system
		ibeSystem := createTestIBESystem()

		// Test backward compatibility with legacy methods
		userSecret := []byte("test_user_secret")
		legacyPseudonym := ibeSystem.GeneratePseudonymFromUserSecret(userSecret)

		// Legacy pseudonym should still work
		assert.NotEmpty(t, legacyPseudonym, "Legacy pseudonym generation should work")
		assert.Len(t, legacyPseudonym, 32, "Legacy pseudonym should be 32 hex characters")

		// Test legacy role key generation
		legacyRoleKey := ibeSystem.GenerateRoleKey("user", "authentication", time.Now().Add(time.Hour))
		assert.NotEmpty(t, legacyRoleKey, "Legacy role key generation should work")
		assert.Len(t, legacyRoleKey, 32, "Legacy role key should be 32 bytes")
	})
}

func TestIBESystem_IntegrationWithDatabase(t *testing.T) {
	t.Run("Database Integration", func(t *testing.T) {
		// Create IBE system
		ibeSystem := createTestIBESystem()

		// Test pseudonym generation for database user
		pseudonym := ibeSystem.CreateEnhancedPseudonym(1, "database_test")
		assert.NotEmpty(t, pseudonym, "Should generate pseudonym for database user")

		// Test fingerprint generation
		realIdentity := "test@example.com"
		fingerprint := ibeSystem.GenerateFingerprint(realIdentity)
		assert.NotEmpty(t, fingerprint, "Should generate fingerprint for real identity")
		assert.Len(t, fingerprint, 32, "Fingerprint should be 32 hex characters")

		// Test identity encryption/decryption
		adminKey := ibeSystem.GenerateTimeBoundedKey("platform_admin", "correlation", time.Hour)

		// Use the pseudonym ID format that matches the system
		pseudonymID := pseudonym // The pseudonym is already in the correct format

		encryptedMapping, err := ibeSystem.EncryptIdentity(realIdentity, pseudonymID, adminKey)
		require.NoError(t, err, "Should encrypt identity mapping")

		// Debug: print encrypted mapping
		t.Logf("Encrypted mapping: %s", encryptedMapping)

		decryptedRealIdentity, decryptedPseudonym, err := ibeSystem.DecryptIdentity(encryptedMapping, adminKey)
		require.NoError(t, err, "Should decrypt identity mapping")
		// Debug: print decrypted values
		t.Logf("Decrypted real identity: %s", decryptedRealIdentity)
		t.Logf("Decrypted pseudonym: %s", decryptedPseudonym)

		// Expect the fingerprint, not the real identity
		expectedFingerprint := ibeSystem.GenerateFingerprint(realIdentity)
		expectedMapping := expectedFingerprint + ":" + pseudonymID
		assert.Equal(t, expectedMapping, decryptedRealIdentity, "Decrypted real identity should be the fingerprint mapping")
		// The decrypted pseudonym is not used in this implementation
		// Optionally, assert that it's empty
		assert.Empty(t, decryptedPseudonym, "Decrypted pseudonym should be empty (not used)")
	})
}

func TestIBESystem_Configuration(t *testing.T) {
	t.Run("Default Configuration", func(t *testing.T) {
		ibeSystem := NewIBESystemWithOptions(createDefaultIBEOptions())

		// Test default values
		assert.Equal(t, int32(1), ibeSystem.GetKeyVersion(), "Default key version should be 1")
		assert.Equal(t, "test_fingerprint_salt_v1", ibeSystem.GetSalt(), "Default salt should be test_fingerprint_salt_v1")

		// Test that we can generate keys with default config
		key := ibeSystem.GenerateTimeBoundedKey("user", "test", time.Hour)
		assert.NotEmpty(t, key, "Should generate key with default configuration")
		assert.Len(t, key, 32, "Generated key should be 32 bytes")
	})

	t.Run("Custom Configuration", func(t *testing.T) {
		ibeSystem := NewIBESystemWithOptions(IBEOptions{
			KeyVersion: int32(2),
			Salt:       "custom_salt_v2",
		})

		// Test custom values
		assert.Equal(t, int32(2), ibeSystem.GetKeyVersion(), "Custom key version should be 2")
		assert.Equal(t, "custom_salt_v2", ibeSystem.GetSalt(), "Custom salt should be custom_salt_v2")

		// Test that we can generate keys with custom config
		key := ibeSystem.GenerateTimeBoundedKey("user", "test", time.Hour)
		assert.NotEmpty(t, key, "Should generate key with custom configuration")
		assert.Len(t, key, 32, "Generated key should be 32 bytes")
	})
}

func TestIBESystem_FileOperations(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("Save and Load Domain Masters", func(t *testing.T) {
		ibeSystem := createTestIBESystem()

		// Save domain masters
		err := ibeSystem.SaveDomainMastersToDir(tempDir)
		require.NoError(t, err, "Should save domain masters to directory")

		// Load domain masters
		loadedMasters, err := LoadDomainMastersFromDir(tempDir)
		require.NoError(t, err, "Should load domain masters from directory")

		// Verify all domains are present
		expectedDomains := []string{
			DOMAIN_USER_PSEUDONYMS,
			DOMAIN_USER_CORRELATION,
			DOMAIN_MOD_CORRELATION,
			DOMAIN_ADMIN_CORRELATION,
			DOMAIN_LEGAL_CORRELATION,
		}

		for _, domain := range expectedDomains {
			master, exists := loadedMasters[domain]
			assert.True(t, exists, "Domain %s should be present", domain)
			assert.Len(t, master, 32, "Domain master for %s should be 32 bytes", domain)
		}
	})

	t.Run("Load Domain Masters from Non-existent Directory", func(t *testing.T) {
		_, err := LoadDomainMastersFromDir("/non/existent/directory")
		assert.Error(t, err, "Should return error for non-existent directory")
	})
}

func TestIBESystem_IdentityOperations(t *testing.T) {
	ibeSystem := createTestIBESystem()

	t.Run("Fingerprint Generation", func(t *testing.T) {
		// Test fingerprint generation
		identity := "test@example.com"
		fingerprint1 := ibeSystem.GenerateFingerprint(identity)
		fingerprint2 := ibeSystem.GenerateFingerprint(identity)

		// Same identity should generate same fingerprint
		assert.Equal(t, fingerprint1, fingerprint2, "Same identity should generate same fingerprint")
		assert.Len(t, fingerprint1, 32, "Fingerprint should be 32 hex characters")

		// Different identities should generate different fingerprints
		identity2 := "test2@example.com"
		fingerprint3 := ibeSystem.GenerateFingerprint(identity2)
		assert.NotEqual(t, fingerprint1, fingerprint3, "Different identities should generate different fingerprints")
	})

	t.Run("Identity Encryption and Decryption", func(t *testing.T) {
		// Test identity encryption and decryption
		realIdentity := "test@example.com"
		pseudonymID := "abc123def456"
		adminKey := ibeSystem.GenerateTimeBoundedKey("platform_admin", "correlation", time.Hour)

		// Encrypt identity mapping
		encryptedMapping, err := ibeSystem.EncryptIdentity(realIdentity, pseudonymID, adminKey)
		require.NoError(t, err, "Should encrypt identity mapping")
		assert.NotEmpty(t, encryptedMapping, "Encrypted mapping should not be empty")

		// Decrypt identity mapping
		decryptedRealIdentity, decryptedPseudonym, err := ibeSystem.DecryptIdentity(encryptedMapping, adminKey)
		require.NoError(t, err, "Should decrypt identity mapping")

		// Verify decrypted values - expect fingerprint, not real identity
		expectedFingerprint := ibeSystem.GenerateFingerprint(realIdentity)
		expectedMapping := expectedFingerprint + ":" + pseudonymID
		assert.Equal(t, expectedMapping, decryptedRealIdentity, "Decrypted real identity should be the fingerprint mapping")
		assert.Empty(t, decryptedPseudonym, "Decrypted pseudonym should be empty (not used)")
	})

	t.Run("Identity Decryption with Wrong Key", func(t *testing.T) {
		// Test decryption with wrong key
		realIdentity := "test@example.com"
		pseudonymID := "abc123def456"
		adminKey := ibeSystem.GenerateTimeBoundedKey("platform_admin", "correlation", time.Hour)
		wrongKey := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)

		// Encrypt with admin key
		encryptedMapping, err := ibeSystem.EncryptIdentity(realIdentity, pseudonymID, adminKey)
		require.NoError(t, err, "Should encrypt identity mapping")

		// Try to decrypt with wrong key
		_, _, err = ibeSystem.DecryptIdentity(encryptedMapping, wrongKey)
		assert.Error(t, err, "Should fail to decrypt with wrong key")
	})
}

// TestMultiVersionKeyMigration tests the multi-version key system during migration
func TestMultiVersionKeyMigration(t *testing.T) {
	// Create key registry with multiple versions
	registry := NewKeyVersionRegistry()

	// Add old key version (version 1) - make sure keys are different
	oldDomainKeys := map[string][]byte{
		DOMAIN_USER_CORRELATION:  []byte("old-user-domain-key-32-bytes-long"),
		DOMAIN_ADMIN_CORRELATION: []byte("old-admin-domain-key-32-bytes"),
	}
	registry.AddKeyVersion(1, oldDomainKeys, "old_salt_v1")

	// Add new key version (version 2) - make sure keys are different
	newDomainKeys := map[string][]byte{
		DOMAIN_USER_CORRELATION:  []byte("new-user-domain-key-32-bytes-long"),
		DOMAIN_ADMIN_CORRELATION: []byte("new-admin-domain-key-32-bytes"),
	}
	registry.AddKeyVersion(2, newDomainKeys, "new_salt_v2")

	// Create IBE system with multi-version support
	ibeSystem := NewIBESystemWithOptions(IBEOptions{
		DomainMasters: newDomainKeys, // Current keys are version 2
		KeyVersion:    2,
		Salt:          "new_salt_v2",
		KeyRegistry:   registry,
	})

	// Enable migration mode BEFORE any operations
	ibeSystem.EnableMigrationMode()

	// Verify migration mode is enabled
	assert.True(t, registry.MigrationMode)

	// Test data
	realIdentity := "user123@example.com"
	pseudonymID := "pseudonym-456"
	domain := DOMAIN_USER_CORRELATION

	t.Run("EncryptWithOldKey", func(t *testing.T) {
		// Encrypt with old key version
		encrypted, err := ibeSystem.EncryptIdentityWithVersion(realIdentity, pseudonymID, domain, 1)
		assert.NoError(t, err)
		assert.NotNil(t, encrypted)

		// Decrypt with old key version
		decryptedRealIdentity, decryptedPseudonymID, err := ibeSystem.DecryptIdentityWithVersion(encrypted, domain, 1)
		assert.NoError(t, err)
		// Expect fingerprint, not real identity
		expectedFingerprint := ibeSystem.GenerateFingerprint(realIdentity)
		assert.Equal(t, expectedFingerprint, decryptedRealIdentity)
		assert.Equal(t, pseudonymID, decryptedPseudonymID)
	})

	t.Run("EncryptWithNewKey", func(t *testing.T) {
		// Encrypt with new key version
		encrypted, err := ibeSystem.EncryptIdentityWithVersion(realIdentity, pseudonymID, domain, 2)
		assert.NoError(t, err)
		assert.NotNil(t, encrypted)

		// Decrypt with new key version
		decryptedRealIdentity, decryptedPseudonymID, err := ibeSystem.DecryptIdentityWithVersion(encrypted, domain, 2)
		assert.NoError(t, err)
		// Expect fingerprint, not real identity
		expectedFingerprint := ibeSystem.GenerateFingerprint(realIdentity)
		assert.Equal(t, expectedFingerprint, decryptedRealIdentity)
		assert.Equal(t, pseudonymID, decryptedPseudonymID)
	})

	t.Run("CrossVersionDecryptionFails", func(t *testing.T) {
		// Encrypt with old key version
		encrypted, err := ibeSystem.EncryptIdentityWithVersion(realIdentity, pseudonymID, domain, 1)
		assert.NoError(t, err)

		// Try to decrypt with new key version (should fail)
		_, _, err = ibeSystem.DecryptIdentityWithVersion(encrypted, domain, 2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to decrypt")
	})

	t.Run("KeyVersionManagement", func(t *testing.T) {
		// Test key version info retrieval
		keyInfo, err := registry.GetKeyVersionInfo(1)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), keyInfo.Version)
		assert.True(t, keyInfo.IsActive)
		assert.False(t, keyInfo.IsDeprecated)

		// Test deprecating old key version
		err = registry.DeprecateKeyVersion(1)
		assert.NoError(t, err)

		// Verify old key is now deprecated
		keyInfo, err = registry.GetKeyVersionInfo(1)
		assert.NoError(t, err)
		assert.True(t, keyInfo.IsDeprecated)
		assert.False(t, keyInfo.IsActive)
		assert.NotNil(t, keyInfo.DeprecatedAt)

		// Verify new key is still active
		keyInfo, err = registry.GetKeyVersionInfo(2)
		assert.NoError(t, err)
		assert.False(t, keyInfo.IsDeprecated)
		assert.True(t, keyInfo.IsActive)
	})

	t.Run("MigrationModeOperations", func(t *testing.T) {
		// Test that migration mode allows access to deprecated keys
		assert.True(t, registry.MigrationMode)

		// Should still be able to decrypt with deprecated key
		encrypted, err := ibeSystem.EncryptIdentityWithVersion(realIdentity, pseudonymID, domain, 1)
		assert.NoError(t, err)

		decryptedRealIdentity, decryptedPseudonymID, err := ibeSystem.DecryptIdentityWithVersion(encrypted, domain, 1)
		assert.NoError(t, err)
		// Expect fingerprint, not real identity
		expectedFingerprint := ibeSystem.GenerateFingerprint(realIdentity)
		assert.Equal(t, expectedFingerprint, decryptedRealIdentity)
		assert.Equal(t, pseudonymID, decryptedPseudonymID)

		// Disable migration mode
		ibeSystem.DisableMigrationMode()
		assert.False(t, registry.MigrationMode)
	})

	t.Run("CorrelationKeyVersioning", func(t *testing.T) {
		// Re-enable migration mode for this test
		ibeSystem.EnableMigrationMode()

		// Test correlation key generation with different versions
		role := "platform_admin"
		scope := "correlation"
		timeWindow := 24 * time.Hour

		// Generate correlation key with old version
		oldCorrelationKey := ibeSystem.GenerateCorrelationKeyForVersion(role, scope, timeWindow, 1)
		assert.NotNil(t, oldCorrelationKey)

		// Generate correlation key with new version
		newCorrelationKey := ibeSystem.GenerateCorrelationKeyForVersion(role, scope, timeWindow, 2)
		assert.NotNil(t, newCorrelationKey)

		// Keys should be different due to different domain keys
		assert.NotEqual(t, oldCorrelationKey, newCorrelationKey)

		// Test that same version produces same key for same time window
		oldCorrelationKey2 := ibeSystem.GenerateCorrelationKeyForVersion(role, scope, timeWindow, 1)
		assert.Equal(t, oldCorrelationKey, oldCorrelationKey2)

		// Test that different time windows produce different keys
		differentTimeWindow := 12 * time.Hour
		oldCorrelationKey3 := ibeSystem.GenerateCorrelationKeyForVersion(role, scope, differentTimeWindow, 1)
		assert.NotEqual(t, oldCorrelationKey, oldCorrelationKey3)
	})
}

// TestKeyVersionRegistry tests the key version registry functionality
func TestKeyVersionRegistry(t *testing.T) {
	registry := NewKeyVersionRegistry()

	t.Run("AddKeyVersion", func(t *testing.T) {
		domainKeys := map[string][]byte{
			DOMAIN_USER_CORRELATION: []byte("test-domain-key-32-bytes-long"),
		}

		registry.AddKeyVersion(1, domainKeys, "test_salt")

		keyInfo, err := registry.GetKeyVersionInfo(1)
		assert.NoError(t, err)
		assert.Equal(t, int32(1), keyInfo.Version)
		assert.Equal(t, "test_salt", keyInfo.Salt)
		assert.True(t, keyInfo.IsActive)
		assert.False(t, keyInfo.IsDeprecated)
	})

	t.Run("GetDomainKeyForVersion", func(t *testing.T) {
		domainKeys := map[string][]byte{
			DOMAIN_USER_CORRELATION: []byte("test-domain-key-32-bytes-long"),
		}

		registry.AddKeyVersion(1, domainKeys, "test_salt")

		key, err := registry.GetDomainKeyForVersion(1, DOMAIN_USER_CORRELATION)
		assert.NoError(t, err)
		assert.Equal(t, []byte("test-domain-key-32-bytes-long"), key)

		// Test non-existent domain
		_, err = registry.GetDomainKeyForVersion(1, "non_existent_domain")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no key found for domain")
	})

	t.Run("DeprecateKeyVersion", func(t *testing.T) {
		domainKeys := map[string][]byte{
			DOMAIN_USER_CORRELATION: []byte("test-domain-key-32-bytes-long"),
		}

		registry.AddKeyVersion(1, domainKeys, "test_salt")

		// Initially active
		keyInfo, err := registry.GetKeyVersionInfo(1)
		assert.NoError(t, err)
		assert.True(t, keyInfo.IsActive)

		// Deprecate
		err = registry.DeprecateKeyVersion(1)
		assert.NoError(t, err)

		// Now deprecated
		keyInfo, err = registry.GetKeyVersionInfo(1)
		assert.NoError(t, err)
		assert.True(t, keyInfo.IsDeprecated)
		assert.False(t, keyInfo.IsActive)
		assert.NotNil(t, keyInfo.DeprecatedAt)
	})

	t.Run("DeprecateNonExistentVersion", func(t *testing.T) {
		err := registry.DeprecateKeyVersion(999)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key version 999 not found")
	})
}
