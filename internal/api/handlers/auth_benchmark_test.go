package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"golang.org/x/crypto/bcrypt"
)

// Benchmark password hashing operations
func BenchmarkPasswordHashing(b *testing.B) {
	password := "test_password_123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordVerification(b *testing.B) {
	password := "test_password_123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := bcrypt.CompareHashAndPassword(hash, []byte(password))
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark different password lengths
func BenchmarkPasswordHashing_DifferentLengths(b *testing.B) {
	passwords := []string{
		"short",
		"medium_length_password",
		"very_long_password_with_many_characters_for_benchmarking",
	}

	for _, password := range passwords {
		b.Run(fmt.Sprintf("%d_chars", len(password)), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark JWT operations with production key sizes
func BenchmarkJWT_Creation(b *testing.B) {
	// Use realistic 256-bit JWT secret (32 bytes)
	secret := "production_jwt_secret_256_bit_key_for_benchmarking_secure_random_string_32_bytes_long"
	claims := &middleware.JWTClaims{
		UserID:            1,
		Email:             "test@example.com",
		ActivePseudonymID: "pseudonym_123",
		Capabilities:      []string{"create_content", "vote"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		_, err := token.SignedString([]byte(secret))
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWT_Validation(b *testing.B) {
	// Use realistic 256-bit JWT secret (32 bytes)
	secret := "production_jwt_secret_256_bit_key_for_benchmarking_secure_random_string_32_bytes_long"
	claims := &middleware.JWTClaims{
		UserID:            1,
		Email:             "test@example.com",
		ActivePseudonymID: "pseudonym_123",
		Capabilities:      []string{"create_content", "vote"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.ParseWithClaims(tokenString, &middleware.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJWT_Validation_InvalidToken(b *testing.B) {
	invalidToken := "invalid.jwt.token"
	// Use realistic 256-bit JWT secret (32 bytes)
	secret := "production_jwt_secret_256_bit_key_for_benchmarking_secure_random_string_32_bytes_long"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.ParseWithClaims(invalidToken, &middleware.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		// We expect this to fail, but we're benchmarking the validation attempt
		_ = err
	}
}

func BenchmarkJWT_Validation_ExpiredToken(b *testing.B) {
	// Use realistic 256-bit JWT secret (32 bytes)
	secret := "production_jwt_secret_256_bit_key_for_benchmarking_secure_random_string_32_bytes_long"
	claims := &middleware.JWTClaims{
		UserID:            1,
		Email:             "test@example.com",
		ActivePseudonymID: "pseudonym_123",
		Capabilities:      []string{"create_content", "vote"},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // Expired
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jwt.ParseWithClaims(tokenString, &middleware.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		// We expect this to fail, but we're benchmarking the validation attempt
		_ = err
	}
}

// Benchmark JWT operations with different payload sizes
func BenchmarkJWT_Validation_DifferentPayloadSizes(b *testing.B) {
	// Use realistic 256-bit JWT secret (32 bytes)
	secret := "production_jwt_secret_256_bit_key_for_benchmarking_secure_random_string_32_bytes_long"

	testCases := []struct {
		name         string
		capabilities []string
	}{
		{
			name:         "minimal",
			capabilities: []string{"create_content"},
		},
		{
			name:         "medium",
			capabilities: []string{"create_content", "vote", "message", "report"},
		},
		{
			name:         "large",
			capabilities: []string{"create_content", "vote", "message", "report", "moderate_content", "ban_users", "manage_subforum_settings", "manage_moderators", "correlate_fingerprints", "correlate_identities"},
		},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			claims := &middleware.JWTClaims{
				UserID:            1,
				Email:             "test@example.com",
				ActivePseudonymID: "pseudonym_123",
				Capabilities:      tc.capabilities,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				},
			}

			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			tokenString, err := token.SignedString([]byte(secret))
			if err != nil {
				b.Fatal(err)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := jwt.ParseWithClaims(tokenString, &middleware.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(secret), nil
				})
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// Benchmark SHA256 operations (used in various crypto operations)
func BenchmarkSHA256_Hashing(b *testing.B) {
	data := []byte("test_data_for_sha256_hashing_benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hash := sha256.Sum256(data)
		_ = hex.EncodeToString(hash[:])
	}
}

// Benchmark random number generation (used in token generation)
func BenchmarkRandomBytes(b *testing.B) {
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

// Benchmark hex encoding/decoding operations
func BenchmarkHexEncoding(b *testing.B) {
	data := make([]byte, 32)
	rand.Read(data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = hex.EncodeToString(data)
	}
}

func BenchmarkHexDecoding(b *testing.B) {
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

// Benchmark bcrypt with different costs
func BenchmarkBcrypt_DifferentCosts(b *testing.B) {
	password := "test_password_123"
	costs := []int{10, 12, 14}

	for _, cost := range costs {
		b.Run(fmt.Sprintf("cost_%d", cost), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_, err := bcrypt.GenerateFromPassword([]byte(password), cost)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
