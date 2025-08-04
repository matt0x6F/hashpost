package ibe

import (
	"crypto/rand"
	"testing"
	"time"
)

// createBenchmarkIBESystem creates an IBE system for benchmarking with production key sizes
func createBenchmarkIBESystem() *IBESystem {
	// Generate realistic 32-byte domain master keys (AES-256 equivalent)
	domainMasters := make(map[string][]byte)
	domains := []string{
		DOMAIN_USER_PSEUDONYMS,
		DOMAIN_USER_CORRELATION,
		DOMAIN_MOD_CORRELATION,
		DOMAIN_ADMIN_CORRELATION,
		DOMAIN_LEGAL_CORRELATION,
	}
	
	for _, domain := range domains {
		key := make([]byte, 32) // 256-bit keys for production
		rand.Read(key)
		domainMasters[domain] = key
	}
	
	return NewIBESystemWithOptions(IBEOptions{
		DomainMasters: domainMasters,
		KeyVersion:    1,
		Salt:          "production_fingerprint_salt_v1_secure_random_string",
	})
}

func BenchmarkIBESystem_GeneratePseudonym(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ibeSystem.GeneratePseudonym(int64(i), "benchmark_context", 1)
	}
}

func BenchmarkIBESystem_GeneratePseudonym_Parallel(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			ibeSystem.GeneratePseudonym(int64(i), "benchmark_context", 1)
			i++
		}
	})
}

func BenchmarkIBESystem_CreateEnhancedPseudonym(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ibeSystem.CreateEnhancedPseudonym(int64(i), "benchmark_context")
	}
}

func BenchmarkIBESystem_GenerateCorrelationKey(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ibeSystem.GenerateCorrelationKey("user", "correlation", time.Hour)
	}
}

func BenchmarkIBESystem_GenerateTimeBoundedKey(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ibeSystem.GenerateTimeBoundedKey("user", "correlation", time.Hour)
	}
}

func BenchmarkIBESystem_GenerateRoleKey(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	expiration := time.Now().Add(24 * time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ibeSystem.GenerateRoleKey("user", "scope", expiration)
	}
}

func BenchmarkIBESystem_ValidateRoleKey(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	expiration := time.Now().Add(24 * time.Hour)
	roleKey := ibeSystem.GenerateRoleKey("user", "scope", expiration)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ibeSystem.ValidateRoleKey(roleKey, "user", "scope", expiration)
	}
}

func BenchmarkIBESystem_GenerateFingerprint(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	realIdentity := "user@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ibeSystem.GenerateFingerprint(realIdentity)
	}
}

func BenchmarkIBESystem_EncryptIdentity(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	adminKey := []byte("admin_key_32_bytes_long_key_here")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ibeSystem.EncryptIdentity(realIdentity, pseudonymID, adminKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBESystem_DecryptIdentity(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	adminKey := []byte("admin_key_32_bytes_long_key_here")

	encryptedData, err := ibeSystem.EncryptIdentity(realIdentity, pseudonymID, adminKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := ibeSystem.DecryptIdentity(encryptedData, adminKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBESystem_EncryptIdentityWithDomain(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	domain := DOMAIN_USER_CORRELATION
	adminKey := []byte("admin_key_32_bytes_long_key_here")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ibeSystem.EncryptIdentityWithDomain(realIdentity, pseudonymID, domain, adminKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBESystem_DecryptIdentityWithDomain(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	domain := DOMAIN_USER_CORRELATION
	adminKey := []byte("admin_key_32_bytes_long_key_here")

	encryptedData, err := ibeSystem.EncryptIdentityWithDomain(realIdentity, pseudonymID, domain, adminKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := ibeSystem.DecryptIdentityWithDomain(encryptedData, adminKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBESystem_EncryptIdentityWithVersion(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	domain := DOMAIN_USER_CORRELATION
	version := 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ibeSystem.EncryptIdentityWithVersion(realIdentity, pseudonymID, domain, version)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBESystem_DecryptIdentityWithVersion(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	domain := DOMAIN_USER_CORRELATION
	version := 1

	encryptedData, err := ibeSystem.EncryptIdentityWithVersion(realIdentity, pseudonymID, domain, version)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := ibeSystem.DecryptIdentityWithVersion(encryptedData, domain, version)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBESystem_GenerateCorrelationKeyForVersion(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ibeSystem.GenerateCorrelationKeyForVersion("user", "correlation", time.Hour, 1)
	}
}

func BenchmarkIBESystem_EncryptFingerprintMapping(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	fingerprint := "fingerprint_123"
	pseudonymID := "pseudonym_123"
	domain := DOMAIN_USER_CORRELATION
	version := 1

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ibeSystem.EncryptFingerprintMapping(fingerprint, pseudonymID, domain, version)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark different key sizes and time windows
func BenchmarkIBESystem_GenerateTimeBoundedKey_DifferentDurations(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	durations := []time.Duration{
		time.Minute,
		time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
	}

	for _, duration := range durations {
		b.Run(duration.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ibeSystem.GenerateTimeBoundedKey("user", "correlation", duration)
			}
		})
	}
}

// Benchmark different roles
func BenchmarkIBESystem_GenerateTimeBoundedKey_DifferentRoles(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	roles := []string{"user", "moderator", "platform_admin", "legal_team"}

	for _, role := range roles {
		b.Run(role, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ibeSystem.GenerateTimeBoundedKey(role, "correlation", time.Hour)
			}
		})
	}
}

// Benchmark pseudonym generation with different contexts
func BenchmarkIBESystem_GeneratePseudonym_DifferentContexts(b *testing.B) {
	ibeSystem := createBenchmarkIBESystem()
	contexts := []string{"short", "medium_length_context", "very_long_context_name_for_benchmarking"}

	for _, context := range contexts {
		b.Run(context, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ibeSystem.GeneratePseudonym(int64(i), context, 1)
			}
		})
	}
}
