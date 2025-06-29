package ibe

import (
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
