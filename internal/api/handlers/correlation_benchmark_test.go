package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/ibe"
)

// createBenchmarkCorrelationHandler creates a correlation handler for benchmarking with production key sizes
func createBenchmarkCorrelationHandler() *CorrelationHandler {
	// Generate realistic 32-byte domain master keys (AES-256 equivalent)
	domainMasters := make(map[string][]byte)
	domains := []string{
		ibe.DOMAIN_USER_PSEUDONYMS,
		ibe.DOMAIN_USER_CORRELATION,
		ibe.DOMAIN_MOD_CORRELATION,
		ibe.DOMAIN_ADMIN_CORRELATION,
		ibe.DOMAIN_LEGAL_CORRELATION,
	}

	for _, domain := range domains {
		key := make([]byte, 32) // 256-bit keys for production
		rand.Read(key)
		domainMasters[domain] = key
	}

	// Create IBE system
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{
		DomainMasters: domainMasters,
		KeyVersion:    1,
		Salt:          "production_fingerprint_salt_v1_secure_random_string",
	})

	return &CorrelationHandler{
		ibeSystem: ibeSystem,
	}
}

// Benchmark IBE correlation operations
func BenchmarkIBE_GenerateCorrelationKey(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ibeSystem.GenerateCorrelationKey("user", "correlation", time.Hour)
	}
}

func BenchmarkIBE_GenerateCorrelationKey_Parallel(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			handler.ibeSystem.GenerateCorrelationKey("user", "correlation", time.Hour)
			i++
		}
	})
}

func BenchmarkIBE_GenerateCorrelationKeyForVersion(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ibeSystem.GenerateCorrelationKeyForVersion("user", "correlation", time.Hour, 1)
	}
}

// Benchmark different roles for correlation
func BenchmarkIBE_GenerateCorrelationKey_DifferentRoles(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	roles := []string{constants.RoleUser, constants.RoleModerator, constants.RolePlatformAdmin, "legal_team"}

	for _, role := range roles {
		b.Run(role, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				handler.ibeSystem.GenerateCorrelationKey(role, "correlation", time.Hour)
			}
		})
	}
}

// Benchmark different time windows for correlation
func BenchmarkIBE_GenerateCorrelationKey_DifferentTimeWindows(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	durations := []time.Duration{
		time.Minute,
		time.Hour,
		24 * time.Hour,
		7 * 24 * time.Hour,
	}

	for _, duration := range durations {
		b.Run(duration.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				handler.ibeSystem.GenerateCorrelationKey("user", "correlation", duration)
			}
		})
	}
}

// Benchmark fingerprint generation
func BenchmarkIBE_GenerateFingerprint(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	realIdentity := "user@example.com"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.ibeSystem.GenerateFingerprint(realIdentity)
	}
}

func BenchmarkIBE_GenerateFingerprint_DifferentIdentities(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	identities := []string{
		"short@test.com",
		"medium_length_email@example.com",
		"very_long_email_address_with_many_characters@example.com",
	}

	for _, identity := range identities {
		b.Run(fmt.Sprintf("%d_chars", len(identity)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				handler.ibeSystem.GenerateFingerprint(identity)
			}
		})
	}
}

// Benchmark identity encryption/decryption operations
func BenchmarkIBE_EncryptIdentity(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	adminKey := []byte("admin_key_32_bytes_long_key_here")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.ibeSystem.EncryptIdentity(realIdentity, pseudonymID, adminKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBE_DecryptIdentity(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	adminKey := []byte("admin_key_32_bytes_long_key_here")

	encryptedData, err := handler.ibeSystem.EncryptIdentity(realIdentity, pseudonymID, adminKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := handler.ibeSystem.DecryptIdentity(encryptedData, adminKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBE_EncryptIdentityWithDomain(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	domain := ibe.DOMAIN_USER_CORRELATION
	adminKey := []byte("admin_key_32_bytes_long_key_here")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.ibeSystem.EncryptIdentityWithDomain(realIdentity, pseudonymID, domain, adminKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBE_DecryptIdentityWithDomain(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	domain := ibe.DOMAIN_USER_CORRELATION
	adminKey := []byte("admin_key_32_bytes_long_key_here")

	encryptedData, err := handler.ibeSystem.EncryptIdentityWithDomain(realIdentity, pseudonymID, domain, adminKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := handler.ibeSystem.DecryptIdentityWithDomain(encryptedData, adminKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBE_EncryptIdentityWithVersion(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	domain := ibe.DOMAIN_USER_CORRELATION
	version := int32(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.ibeSystem.EncryptIdentityWithVersion(realIdentity, pseudonymID, domain, version)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkIBE_DecryptIdentityWithVersion(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	domain := ibe.DOMAIN_USER_CORRELATION
	version := int32(1)

	encryptedData, err := handler.ibeSystem.EncryptIdentityWithVersion(realIdentity, pseudonymID, domain, version)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, err := handler.ibeSystem.DecryptIdentityWithVersion(encryptedData, domain, version)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark fingerprint mapping encryption
func BenchmarkIBE_EncryptFingerprintMapping(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	fingerprint := "fingerprint_123"
	pseudonymID := "pseudonym_123"
	domain := ibe.DOMAIN_USER_CORRELATION
	version := int32(1)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := handler.ibeSystem.EncryptFingerprintMapping(fingerprint, pseudonymID, domain, version)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark different domain operations
func BenchmarkIBE_EncryptIdentity_DifferentDomains(b *testing.B) {
	handler := createBenchmarkCorrelationHandler()
	realIdentity := "user@example.com"
	pseudonymID := "pseudonym_123"
	adminKey := []byte("admin_key_32_bytes_long_key_here")
	domains := []string{
		ibe.DOMAIN_USER_CORRELATION,
		ibe.DOMAIN_MOD_CORRELATION,
		ibe.DOMAIN_ADMIN_CORRELATION,
		ibe.DOMAIN_LEGAL_CORRELATION,
	}

	for _, domain := range domains {
		b.Run(domain, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := handler.ibeSystem.EncryptIdentityWithDomain(realIdentity, pseudonymID, domain, adminKey)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark SHA256 operations used in correlation
func BenchmarkSHA256_Correlation(b *testing.B) {
	data := []byte("correlation_data_for_sha256_hashing_benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := sha256.Sum256(data)
		_ = hex.EncodeToString(hash[:])
	}
}

// Benchmark hex encoding/decoding for correlation data
func BenchmarkHexEncoding_Correlation(b *testing.B) {
	data := make([]byte, 32)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hex.EncodeToString(data)
	}
}

func BenchmarkHexDecoding_Correlation(b *testing.B) {
	data := make([]byte, 32)
	rand.Read(data)
	hexString := hex.EncodeToString(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hex.DecodeString(hexString)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark random number generation for correlation keys
func BenchmarkRandomBytes_Correlation(b *testing.B) {
	sizes := []int{16, 32, 64, 128}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("%d_bytes", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := rand.Read(make([]byte, size))
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
