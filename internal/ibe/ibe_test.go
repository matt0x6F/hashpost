package ibe

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestIBESystem_GeneratePseudonym(t *testing.T) {
	ibe := NewIBESystem()

	// Test that same user secret generates same pseudonym
	userSecret1 := []byte("test_user_secret_1")
	userSecret2 := []byte("test_user_secret_2")

	pseudonym1 := ibe.GeneratePseudonym(userSecret1)
	pseudonym2 := ibe.GeneratePseudonym(userSecret1)
	pseudonym3 := ibe.GeneratePseudonym(userSecret2)

	if pseudonym1 != pseudonym2 {
		t.Errorf("Same user secret should generate same pseudonym: %s != %s", pseudonym1, pseudonym2)
	}

	if pseudonym1 == pseudonym3 {
		t.Errorf("Different user secrets should generate different pseudonyms")
	}

	if len(pseudonym1) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("Pseudonym should be 32 characters long, got %d", len(pseudonym1))
	}
}

func TestIBESystem_EncryptDecryptIdentity(t *testing.T) {
	ibe := NewIBESystem()

	realIdentity := "alice@example.com"
	pseudonymID := "a1b2c3d4e5f6g7h8"
	adminKey := []byte("test_admin_key")

	// Encrypt the mapping
	encrypted, err := ibe.EncryptIdentity(realIdentity, pseudonymID, adminKey)
	if err != nil {
		t.Fatalf("Failed to encrypt identity: %v", err)
	}

	// Decrypt the mapping
	decrypted, _, err := ibe.DecryptIdentity(encrypted, adminKey)
	if err != nil {
		t.Fatalf("Failed to decrypt identity: %v", err)
	}

	expectedFingerprint := ibe.GenerateFingerprint(realIdentity)
	expected := expectedFingerprint + ":" + pseudonymID
	if decrypted != expected {
		t.Errorf("Decrypted result doesn't match expected: got %s, want %s", decrypted, expected)
	}
}

func TestIBESystem_GenerateRoleKey(t *testing.T) {
	ibe := NewIBESystem()

	role := "site_admin"
	scope := "full_correlation"
	expiration := time.Now().AddDate(0, 1, 0)

	// Generate role key
	key1 := ibe.GenerateRoleKey(role, scope, expiration)
	key2 := ibe.GenerateRoleKey(role, scope, expiration)

	// Same parameters should generate same key
	if string(key1) != string(key2) {
		t.Errorf("Same parameters should generate same role key")
	}

	// Different parameters should generate different keys
	key3 := ibe.GenerateRoleKey("different_role", scope, expiration)
	if string(key1) == string(key3) {
		t.Errorf("Different roles should generate different keys")
	}

	if len(key1) != 32 {
		t.Errorf("Role key should be 32 bytes long, got %d", len(key1))
	}
}

func TestIBESystem_ValidateRoleKey(t *testing.T) {
	ibe := NewIBESystem()

	role := "trust_safety"
	scope := "harassment_investigation"
	expiration := time.Now().AddDate(0, 1, 0)

	// Generate valid key
	validKey := ibe.GenerateRoleKey(role, scope, expiration)

	// Test valid key
	if !ibe.ValidateRoleKey(validKey, role, scope, expiration) {
		t.Errorf("Valid role key should pass validation")
	}

	// Test expired key
	expiredExpiration := time.Now().AddDate(0, -1, 0) // Past date
	expiredKey := ibe.GenerateRoleKey(role, scope, expiredExpiration)
	if ibe.ValidateRoleKey(expiredKey, role, scope, expiredExpiration) {
		t.Errorf("Expired role key should fail validation")
	}

	// Test wrong role - should fail because the key was generated with different role
	wrongRoleKey := ibe.GenerateRoleKey("wrong_role", scope, expiration)
	if ibe.ValidateRoleKey(wrongRoleKey, role, scope, expiration) {
		t.Errorf("Wrong role should fail validation")
	}

	// Test wrong scope - should fail because the key was generated with different scope
	wrongScopeKey := ibe.GenerateRoleKey(role, "wrong_scope", expiration)
	if ibe.ValidateRoleKey(wrongScopeKey, role, scope, expiration) {
		t.Errorf("Wrong scope should fail validation")
	}
}

func TestIBESystem_MultiplePseudonymsPerUser(t *testing.T) {
	ibe := NewIBESystem()
	realIdentity := "alice@example.com"
	adminKey := ibe.GenerateRoleKey("site_admin", "full_correlation", time.Now().AddDate(1, 0, 0))

	// Generate two different pseudonym secrets for the same user
	userSecret1 := []byte("alice_secret_1")
	userSecret2 := []byte("alice_secret_2")

	pseudonym1 := ibe.GeneratePseudonym(userSecret1)
	pseudonym2 := ibe.GeneratePseudonym(userSecret2)

	if pseudonym1 == pseudonym2 {
		t.Errorf("Different pseudonym secrets should yield different pseudonyms")
	}

	// Encrypt both mappings
	enc1, err1 := ibe.EncryptIdentity(realIdentity, pseudonym1, adminKey)
	enc2, err2 := ibe.EncryptIdentity(realIdentity, pseudonym2, adminKey)
	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to encrypt identity mappings: %v, %v", err1, err2)
	}

	// Decrypt and check that both map to the same fingerprint
	mapping1, _, err1 := ibe.DecryptIdentity(enc1, adminKey)
	mapping2, _, err2 := ibe.DecryptIdentity(enc2, adminKey)
	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to decrypt identity mappings: %v, %v", err1, err2)
	}

	expectedFingerprint := ibe.GenerateFingerprint(realIdentity)
	expected1 := expectedFingerprint + ":" + pseudonym1
	expected2 := expectedFingerprint + ":" + pseudonym2
	if mapping1 != expected1 {
		t.Errorf("Decrypted mapping1 incorrect: got %s, want %s", mapping1, expected1)
	}
	if mapping2 != expected2 {
		t.Errorf("Decrypted mapping2 incorrect: got %s, want %s", mapping2, expected2)
	}
}

func TestIBESystem_FingerprintGeneration(t *testing.T) {
	ibe := NewIBESystem()
	realIdentity := "alice@example.com"

	// Generate fingerprint
	fingerprint1 := ibe.GenerateFingerprint(realIdentity)
	fingerprint2 := ibe.GenerateFingerprint(realIdentity)

	// Fingerprint should be deterministic
	if fingerprint1 != fingerprint2 {
		t.Errorf("Fingerprint should be deterministic: %s != %s", fingerprint1, fingerprint2)
	}

	// Different identities should have different fingerprints
	otherIdentity := "bob@example.com"
	otherFingerprint := ibe.GenerateFingerprint(otherIdentity)
	if fingerprint1 == otherFingerprint {
		t.Errorf("Different identities should have different fingerprints")
	}

	// Fingerprint should be 32 characters (16 bytes in hex)
	if len(fingerprint1) != 32 {
		t.Errorf("Fingerprint should be 32 characters long, got %d", len(fingerprint1))
	}
}

func TestIBESystem_FingerprintBasedCorrelation(t *testing.T) {
	ibe := NewIBESystem()
	realIdentity := "alice@example.com"
	adminKey := ibe.GenerateRoleKey("site_admin", "full_correlation", time.Now().AddDate(1, 0, 0))

	// Generate two pseudonyms for the same user
	userSecret1 := []byte("alice_secret_1")
	userSecret2 := []byte("alice_secret_2")
	pseudonym1 := ibe.GeneratePseudonym(userSecret1)
	pseudonym2 := ibe.GeneratePseudonym(userSecret2)

	// Encrypt mappings (now using fingerprints)
	enc1, err1 := ibe.EncryptIdentity(realIdentity, pseudonym1, adminKey)
	enc2, err2 := ibe.EncryptIdentity(realIdentity, pseudonym2, adminKey)
	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to encrypt identity mappings: %v, %v", err1, err2)
	}

	// Decrypt mappings
	mapping1, _, err1 := ibe.DecryptIdentity(enc1, adminKey)
	mapping2, _, err2 := ibe.DecryptIdentity(enc2, adminKey)
	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to decrypt identity mappings: %v, %v", err1, err2)
	}

	// Extract fingerprints from mappings
	expectedFingerprint := ibe.GenerateFingerprint(realIdentity)
	expected1 := expectedFingerprint + ":" + pseudonym1
	expected2 := expectedFingerprint + ":" + pseudonym2

	if mapping1 != expected1 {
		t.Errorf("Decrypted mapping1 incorrect: got %s, want %s", mapping1, expected1)
	}
	if mapping2 != expected2 {
		t.Errorf("Decrypted mapping2 incorrect: got %s, want %s", mapping2, expected2)
	}

	// Both mappings should contain the same fingerprint
	if !strings.Contains(mapping1, expectedFingerprint) || !strings.Contains(mapping2, expectedFingerprint) {
		t.Errorf("Both mappings should contain the same fingerprint: %s", expectedFingerprint)
	}
}

func TestIBESystem_TestConfiguration(t *testing.T) {
	// Create IBE system with test configuration (same as in integration tests)
	testMasterSecret := []byte("test_master_secret_32_bytes_long_key")
	ibeSystem := NewIBESystemWithOptions(IBEOptions{
		DomainMasters: map[string][]byte{
			DOMAIN_USER_PSEUDONYMS:   testMasterSecret,
			DOMAIN_USER_CORRELATION:  testMasterSecret,
			DOMAIN_MOD_CORRELATION:   testMasterSecret,
			DOMAIN_ADMIN_CORRELATION: testMasterSecret,
			DOMAIN_LEGAL_CORRELATION: testMasterSecret,
		},
		KeyVersion: 1,
		Salt:       "test_fingerprint_salt_v1",
	})

	// Test that fingerprint generation is consistent
	email := "test@example.com"
	fingerprint1 := ibeSystem.GenerateFingerprint(email)
	fingerprint2 := ibeSystem.GenerateFingerprint(email)

	if fingerprint1 != fingerprint2 {
		t.Errorf("Fingerprint generation is not consistent: %s != %s", fingerprint1, fingerprint2)
	}

	// Test that role key generation is consistent
	roleKey1 := ibeSystem.GenerateTestRoleKey("user", "self_correlation")
	roleKey2 := ibeSystem.GenerateTestRoleKey("user", "self_correlation")

	if !bytes.Equal(roleKey1, roleKey2) {
		t.Errorf("Role key generation is not consistent")
	}

	// Test that encryption/decryption works with test keys
	pseudonymID := "test_pseudonym_123"
	encrypted, err := ibeSystem.EncryptIdentity(email, pseudonymID, roleKey1)
	if err != nil {
		t.Fatalf("Failed to encrypt identity: %v", err)
	}

	decrypted, _, err := ibeSystem.DecryptIdentity(encrypted, roleKey1)
	if err != nil {
		t.Fatalf("Failed to decrypt identity: %v", err)
	}

	expectedMapping := fmt.Sprintf("%s:%s", fingerprint1, pseudonymID)
	if decrypted != expectedMapping {
		t.Errorf("Decrypted mapping mismatch: expected %s, got %s", expectedMapping, decrypted)
	}

	t.Logf("Test IBE system configuration verified:")
	t.Logf("  Email: %s", email)
	t.Logf("  Fingerprint: %s", fingerprint1)
	t.Logf("  Role key length: %d", len(roleKey1))
	t.Logf("  Decrypted mapping: %s", decrypted)
}

func TestIBESystem_DomainSeparation(t *testing.T) {
	// Create IBE system with test configuration
	ibeSystem := NewIBESystemWithOptions(IBEOptions{
		DomainMasters: map[string][]byte{
			DOMAIN_USER_PSEUDONYMS:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_USER_CORRELATION:  []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_MOD_CORRELATION:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_ADMIN_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_LEGAL_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
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
	if bytes.Equal(userKey, modKey) {
		t.Error("User and moderator keys should be different (different domains)")
	}
	if bytes.Equal(userKey, adminKey) {
		t.Error("User and admin keys should be different (different domains)")
	}
	if bytes.Equal(userKey, legalKey) {
		t.Error("User and legal keys should be different (different domains)")
	}
	if bytes.Equal(modKey, adminKey) {
		t.Error("Moderator and admin keys should be different (different domains)")
	}
	if bytes.Equal(modKey, legalKey) {
		t.Error("Moderator and legal keys should be different (different domains)")
	}
	if bytes.Equal(adminKey, legalKey) {
		t.Error("Admin and legal keys should be different (different domains)")
	}

	// Test that same role gets same key within time window
	userKey2 := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
	if !bytes.Equal(userKey, userKey2) {
		t.Error("Same role and scope should get same key within time window")
	}
}

func TestIBESystem_TimeBoundedKeys(t *testing.T) {
	// Create IBE system with test configuration
	ibeSystem := NewIBESystemWithOptions(IBEOptions{
		DomainMasters: map[string][]byte{
			DOMAIN_USER_PSEUDONYMS:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_USER_CORRELATION:  []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_MOD_CORRELATION:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_ADMIN_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_LEGAL_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
		},
		KeyVersion: 1,
		Salt:       "test_fingerprint_salt_v1",
	})

	// Test that different time windows produce different keys
	key1 := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
	key2 := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Minute)

	if bytes.Equal(key1, key2) {
		t.Error("Different time windows should produce different keys")
	}

	// Test that same time window produces same key
	key3 := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
	if !bytes.Equal(key1, key3) {
		t.Error("Same time window should produce same key")
	}
}

func TestIBESystem_EnhancedPseudonyms(t *testing.T) {
	// Create IBE system with test configuration
	ibeSystem := NewIBESystemWithOptions(IBEOptions{
		DomainMasters: map[string][]byte{
			DOMAIN_USER_PSEUDONYMS:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_USER_CORRELATION:  []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_MOD_CORRELATION:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_ADMIN_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_LEGAL_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
		},
		KeyVersion: 1,
		Salt:       "test_fingerprint_salt_v1",
	})

	userID := int64(12345)

	// Test legacy pseudonym generation (version 1)
	legacyPseudonym := ibeSystem.separated.GeneratePseudonym(userID, "default", 1)

	// Test enhanced pseudonym generation (version 2)
	enhancedPseudonym := ibeSystem.CreateEnhancedPseudonym(userID, "tech_context")

	// Test that different contexts produce different pseudonyms
	enhancedPseudonym2 := ibeSystem.CreateEnhancedPseudonym(userID, "crypto_context")

	if enhancedPseudonym == enhancedPseudonym2 {
		t.Error("Different contexts should produce different pseudonyms")
	}

	// Test that same context produces same pseudonym
	enhancedPseudonym3 := ibeSystem.CreateEnhancedPseudonym(userID, "tech_context")
	if enhancedPseudonym != enhancedPseudonym3 {
		t.Error("Same context should produce same pseudonym")
	}

	// Test that legacy and enhanced pseudonyms are different
	if legacyPseudonym == enhancedPseudonym {
		t.Error("Legacy and enhanced pseudonyms should be different")
	}

	t.Logf("Legacy pseudonym: %s", legacyPseudonym)
	t.Logf("Enhanced pseudonym (tech): %s", enhancedPseudonym)
	t.Logf("Enhanced pseudonym (crypto): %s", enhancedPseudonym2)
}

func TestIBESystem_DomainIsolation(t *testing.T) {
	// Create IBE system with test configuration
	ibeSystem := NewIBESystemWithOptions(IBEOptions{
		DomainMasters: map[string][]byte{
			DOMAIN_USER_PSEUDONYMS:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_USER_CORRELATION:  []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_MOD_CORRELATION:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_ADMIN_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_LEGAL_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
		},
		KeyVersion: 1,
		Salt:       "test_fingerprint_salt_v1",
	})

	// Test that user pseudonym domain is isolated from correlation domains
	userID := int64(12345)
	pseudonym := ibeSystem.separated.GeneratePseudonym(userID, "default", 1)

	// Generate correlation keys for different roles
	userCorrKey := ibeSystem.separated.GenerateCorrelationKey("user", "correlation", time.Hour)
	modCorrKey := ibeSystem.separated.GenerateCorrelationKey("moderator", "correlation", time.Hour)
	adminCorrKey := ibeSystem.separated.GenerateCorrelationKey("platform_admin", "correlation", time.Hour)

	// Test that pseudonym generation doesn't interfere with correlation keys
	pseudonym2 := ibeSystem.separated.GeneratePseudonym(userID, "default", 1)
	userCorrKey2 := ibeSystem.separated.GenerateCorrelationKey("user", "correlation", time.Hour)
	modCorrKey2 := ibeSystem.separated.GenerateCorrelationKey("moderator", "correlation", time.Hour)
	adminCorrKey2 := ibeSystem.separated.GenerateCorrelationKey("platform_admin", "correlation", time.Hour)

	// Pseudonyms should be consistent
	if pseudonym != pseudonym2 {
		t.Error("Pseudonym generation should be deterministic")
	}

	// Correlation keys should be consistent within time window
	if !bytes.Equal(userCorrKey, userCorrKey2) {
		t.Error("User correlation keys should be consistent within time window")
	}
	if !bytes.Equal(modCorrKey, modCorrKey2) {
		t.Error("Moderator correlation keys should be consistent within time window")
	}
	if !bytes.Equal(adminCorrKey, adminCorrKey2) {
		t.Error("Admin correlation keys should be consistent within time window")
	}
}

func TestIBESystem_BackwardCompatibility(t *testing.T) {
	// Create IBE system with test configuration
	ibeSystem := NewIBESystemWithOptions(IBEOptions{
		DomainMasters: map[string][]byte{
			DOMAIN_USER_PSEUDONYMS:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_USER_CORRELATION:  []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_MOD_CORRELATION:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_ADMIN_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_LEGAL_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
		},
		KeyVersion: 1,
		Salt:       "test_fingerprint_salt_v1",
	})

	// Test that existing API methods work exactly as before
	userSecret := []byte("test_user_secret")
	pseudonym := ibeSystem.GeneratePseudonym(userSecret)

	// Test that pseudonym is deterministic
	pseudonym2 := ibeSystem.GeneratePseudonym(userSecret)
	if pseudonym != pseudonym2 {
		t.Error("Backward compatible pseudonym generation should be deterministic")
	}

	// Test that role key generation works
	roleKey := ibeSystem.GenerateRoleKey("user", "correlation", time.Now().Add(time.Hour))
	if len(roleKey) != 32 {
		t.Error("Role key should be 32 bytes")
	}

	// Test that fingerprint generation works
	fingerprint := ibeSystem.GenerateFingerprint("test@example.com")
	fingerprint2 := ibeSystem.GenerateFingerprint("test@example.com")
	if fingerprint != fingerprint2 {
		t.Error("Fingerprint generation should be deterministic")
	}

	// Test that encryption/decryption works
	encrypted, err := ibeSystem.EncryptIdentity("test@example.com", pseudonym, roleKey)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	decrypted, _, err := ibeSystem.DecryptIdentity(encrypted, roleKey)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if !strings.Contains(decrypted, fingerprint) {
		t.Error("Decrypted data should contain the fingerprint")
	}
}

func TestIBESystem_ForwardSecrecy(t *testing.T) {
	// Create IBE system with test configuration
	ibeSystem := NewIBESystemWithOptions(IBEOptions{
		DomainMasters: map[string][]byte{
			DOMAIN_USER_PSEUDONYMS:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_USER_CORRELATION:  []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_MOD_CORRELATION:   []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_ADMIN_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
			DOMAIN_LEGAL_CORRELATION: []byte("test_master_secret_32_bytes_long_key"),
		},
		KeyVersion: 1,
		Salt:       "test_fingerprint_salt_v1",
	})

	// Test that keys change over time (forward secrecy)
	key1 := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)

	// Simulate time passing by using a different time window
	key2 := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour*2)

	if bytes.Equal(key1, key2) {
		t.Error("Keys should change over time for forward secrecy")
	}

	// Test that same time window produces same key
	key3 := ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
	if !bytes.Equal(key1, key3) {
		t.Error("Same time window should produce same key")
	}
}
