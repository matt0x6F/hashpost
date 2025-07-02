//go:build integration

package integration

import (
	"bytes"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIBEKeyGeneration_EnhancedArchitecture(t *testing.T) {
	// Create a temporary directory for test keys
	tempDir := t.TempDir()

	// Test the enhanced IBE system with domain separation
	t.Run("Domain Separation", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create IBE system with test configuration
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
			DomainMasters: map[string][]byte{
				ibe.DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			},
			KeyVersion: 1,
			Salt:       "test_fingerprint_salt_v1",
		})

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
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create IBE system
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
			DomainMasters: map[string][]byte{
				ibe.DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			},
			KeyVersion: 1,
			Salt:       "test_fingerprint_salt_v1",
		})

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
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create IBE system
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
			DomainMasters: map[string][]byte{
				ibe.DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			},
			KeyVersion: 1,
			Salt:       "test_fingerprint_salt_v1",
		})

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

	t.Run("Key File Generation", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create IBE system
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
			DomainMasters: map[string][]byte{
				ibe.DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			},
			KeyVersion: 1,
			Salt:       "test_fingerprint_salt_v1",
		})

		// Test master key file generation
		masterKeyPath := filepath.Join(tempDir, "master.key")
		err := ibeSystem.SaveMasterSecretToFile(masterKeyPath)
		require.NoError(t, err, "Should save master key to file")

		// Verify file exists and has correct permissions
		info, err := os.Stat(masterKeyPath)
		require.NoError(t, err, "Should be able to stat master key file")
		assert.Equal(t, os.FileMode(0600), info.Mode()&0777, "Master key file should have 600 permissions")

		// Debug: print master key file contents and length
		masterKeyFileContents, err := os.ReadFile(masterKeyPath)
		require.NoError(t, err, "Should read master key file for debug")
		t.Logf("Master key file contents: %q", masterKeyFileContents)
		t.Logf("Master key file length: %d", len(masterKeyFileContents))

		// Load master key from file using the production loader
		masterSecretBytes, err := ibe.LoadMasterSecretFromFile(masterKeyPath)
		require.NoError(t, err, "Should load master secret from file")

		// Verify both systems generate same keys
		loadedIBE := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
			DomainMasters: nil, // Will be loaded from file
			KeyVersion:    1,
			Salt:          "test_fingerprint_salt_v1",
		})

		err = loadedIBE.SetMasterSecret(masterSecretBytes)
		require.NoError(t, err, "Should set master secret from file")

		// Verify both systems generate same keys
		originalKey := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
		loadedKey := loadedIBE.GenerateTimeBoundedKey("user", "correlation", time.Hour)

		assert.True(t, bytes.Equal(originalKey, loadedKey), "Keys should be identical after loading from file")
	})

	t.Run("Domain Key Generation", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create IBE system
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
			DomainMasters: map[string][]byte{
				ibe.DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			},
			KeyVersion: 1,
			Salt:       "test_fingerprint_salt_v1",
		})

		// Test domain key generation
		domains := []string{
			ibe.DOMAIN_USER_PSEUDONYMS,
			ibe.DOMAIN_USER_CORRELATION,
			ibe.DOMAIN_MOD_CORRELATION,
			ibe.DOMAIN_ADMIN_CORRELATION,
			ibe.DOMAIN_LEGAL_CORRELATION,
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
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create IBE system
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
			DomainMasters: map[string][]byte{
				ibe.DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			},
			KeyVersion: 1,
			Salt:       "test_fingerprint_salt_v1",
		})

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
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create IBE system
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
			DomainMasters: map[string][]byte{
				ibe.DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			},
			KeyVersion: 1,
			Salt:       "test_fingerprint_salt_v1",
		})

		// Test backward compatibility with legacy methods
		userSecret := []byte("test_user_secret")
		legacyPseudonym := ibeSystem.GeneratePseudonym(userSecret)

		// Legacy pseudonym should still work
		assert.NotEmpty(t, legacyPseudonym, "Legacy pseudonym generation should work")
		assert.Len(t, legacyPseudonym, 32, "Legacy pseudonym should be 32 hex characters")

		// Test legacy role key generation
		legacyRoleKey := ibeSystem.GenerateRoleKey("user", "authentication", time.Now().Add(time.Hour))
		assert.NotEmpty(t, legacyRoleKey, "Legacy role key generation should work")
		assert.Len(t, legacyRoleKey, 32, "Legacy role key should be 32 bytes")
	})
}

func TestIBEKeyGeneration_IntegrationWithDatabase(t *testing.T) {
	t.Run("Database Integration", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test user
		testUser := suite.CreateTestUser(t, "test@example.com", "password123", []string{"user"})

		// Create IBE system
		ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
			DomainMasters: map[string][]byte{
				ibe.DOMAIN_USER_PSEUDONYMS:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_USER_CORRELATION:  []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_MOD_CORRELATION:   []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_ADMIN_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
				ibe.DOMAIN_LEGAL_CORRELATION: []byte("0123456789abcdef0123456789abcdef"),
			},
			KeyVersion: 1,
			Salt:       "test_fingerprint_salt_v1",
		})

		// Test pseudonym generation for database user
		pseudonym := ibeSystem.CreateEnhancedPseudonym(testUser.UserID, "database_test")
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

		// Expect the fingerprint, not the email
		expectedFingerprint := ibeSystem.GenerateFingerprint(realIdentity)
		expectedMapping := expectedFingerprint + ":" + pseudonymID
		assert.Equal(t, expectedMapping, decryptedRealIdentity, "Decrypted real identity should be the fingerprint mapping")
		// The decrypted pseudonym is not used in this implementation
		// Optionally, assert that it's empty
		assert.Empty(t, decryptedPseudonym, "Decrypted pseudonym should be empty (not used)")
	})
}
