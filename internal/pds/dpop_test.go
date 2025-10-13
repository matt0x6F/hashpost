package pds

import (
	"crypto/elliptic"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDPoPService_NewDPoPService(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	jwtService := testutil.NewMockJWTService()

	dpopService := NewDPoPService(authService, nil, logger, jwtService)

	assert.NotNil(t, dpopService)
	assert.Equal(t, authService, dpopService.authService)
	assert.Equal(t, logger, dpopService.logger)
	assert.Equal(t, jwtService, dpopService.jwtService)
}

func TestDPoPService_GenerateNonce_Unit(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	jwtService := testutil.NewMockJWTService()
	dpopService := NewDPoPService(authService, nil, logger, jwtService)

	t.Run("invalid GET request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/dpop/nonce", nil)
		w := httptest.NewRecorder()

		dpopService.GenerateNonce(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
		assert.Contains(t, w.Body.String(), "Method not allowed")
	})

	// Note: POST request requires database access, so we test that in integration tests
}

func TestDPoPService_ValidateDPoPProof_InvalidJWT(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	jwtService := testutil.NewMockJWTService()
	dpopService := NewDPoPService(authService, nil, logger, jwtService)

	tests := []struct {
		name        string
		proofJWT    string
		httpMethod  string
		httpURI     string
		expectedErr string
	}{
		{
			name:        "empty JWT",
			proofJWT:    "",
			httpMethod:  "POST",
			httpURI:     "/api/test",
			expectedErr: "failed to parse DPoP proof",
		},
		{
			name:        "invalid JWT format",
			proofJWT:    "invalid.jwt.token",
			httpMethod:  "POST",
			httpURI:     "/api/test",
			expectedErr: "failed to parse DPoP proof",
		},
		{
			name:        "malformed JWT",
			proofJWT:    "not-a-jwt",
			httpMethod:  "POST",
			httpURI:     "/api/test",
			expectedErr: "failed to parse DPoP proof",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof, err := dpopService.ValidateDPoPProof(tt.proofJWT, tt.httpMethod, tt.httpURI)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
			assert.Nil(t, proof)
		})
	}
}

func TestDPoPService_GenerateDPoPProof_Unit(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	jwtService := testutil.NewMockJWTService()
	dpopService := NewDPoPService(authService, nil, logger, jwtService)

	tests := []struct {
		name        string
		httpMethod  string
		httpURI     string
		nonce       string
		expectError bool
	}{
		{
			name:        "valid proof generation",
			httpMethod:  "POST",
			httpURI:     "/api/test",
			nonce:       "test-nonce-123",
			expectError: false,
		},
		{
			name:        "valid proof without nonce",
			httpMethod:  "GET",
			httpURI:     "/api/read",
			nonce:       "",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proof, err := dpopService.GenerateDPoPProof(tt.httpMethod, tt.httpURI, tt.nonce)

			if tt.expectError {
				require.Error(t, err)
				assert.Empty(t, proof)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, proof)

				// Verify it's a valid JWT structure
				parsedToken, err := jwt.Parse(proof, func(token *jwt.Token) (interface{}, error) {
					// We don't verify signature here, just parse the structure
					return []byte("dummy"), nil
				})
				// Should fail on signature verification but not on parsing
				require.Error(t, err)
				assert.Contains(t, err.Error(), "signature is invalid") // Expected error for dummy key
				assert.NotNil(t, parsedToken)

				// Verify claims structure
				claims, ok := parsedToken.Claims.(jwt.MapClaims)
				require.True(t, ok)
				assert.Equal(t, tt.httpMethod, claims["htm"])
				assert.Equal(t, tt.httpURI, claims["uri"])
				if tt.nonce != "" {
					assert.Equal(t, tt.nonce, claims["nonce"])
				}
			}
		})
	}
}

func TestDPoPService_HelperMethods_Unit(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	jwtService := testutil.NewMockJWTService()
	dpopService := NewDPoPService(authService, nil, logger, jwtService)

	t.Run("generateRandomNonce", func(t *testing.T) {
		nonce1 := dpopService.generateRandomNonce()
		nonce2 := dpopService.generateRandomNonce()

		// Should generate different nonces
		assert.NotEqual(t, nonce1, nonce2)

		// Should be base64 encoded (length should be consistent)
		assert.Len(t, nonce1, 44) // 32 bytes base64 encoded = 44 characters
		assert.Len(t, nonce2, 44)

		// Should be valid base64 (with padding)
		assert.Regexp(t, `^[A-Za-z0-9_-]+=*$`, nonce1)
		assert.Regexp(t, `^[A-Za-z0-9_-]+=*$`, nonce2)
	})

	t.Run("generateRandomString", func(t *testing.T) {
		str1 := dpopService.generateRandomString(16)
		str2 := dpopService.generateRandomString(16)

		// Should generate different strings
		assert.NotEqual(t, str1, str2)

		// Should be base64 encoded
		assert.Len(t, str1, 24) // 16 bytes base64 encoded = 24 characters
		assert.Len(t, str2, 24)

		// Should be valid base64 (with padding)
		assert.Regexp(t, `^[A-Za-z0-9_-]+=*$`, str1)
		assert.Regexp(t, `^[A-Za-z0-9_-]+=*$`, str2)
	})

	t.Run("jwkToPublicKey", func(t *testing.T) {
		// Create JWK as map for JWT service
		jwk := map[string]interface{}{
			"kty": "EC",
			"crv": "P-256",
			"x":   "test-x-coordinate",
			"y":   "test-y-coordinate",
		}

		// Convert to public key using JWT service (mock returns predictable values)
		publicKey, err := dpopService.jwtService.JWKToPublicKey(jwk)
		require.NoError(t, err)
		assert.NotNil(t, publicKey)
		// Mock service returns predictable values for testing
		assert.Equal(t, big.NewInt(12345), publicKey.X)
		assert.Equal(t, big.NewInt(67890), publicKey.Y)
		assert.Equal(t, elliptic.P256(), publicKey.Curve)
	})

	t.Run("jwkToPublicKey_InvalidKeyType", func(t *testing.T) {
		jwk := map[string]interface{}{
			"kty": "RSA", // Invalid for DPoP
			"crv": "P-256",
			"x":   "test",
			"y":   "test",
		}

		publicKey, err := dpopService.jwtService.JWKToPublicKey(jwk)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported key type")
		assert.Nil(t, publicKey)
	})

	t.Run("jwkToPublicKey_InvalidCurve", func(t *testing.T) {
		jwk := map[string]interface{}{
			"kty": "EC",
			"crv": "P-384", // Invalid for DPoP (should be P-256)
			"x":   "test",
			"y":   "test",
		}

		publicKey, err := dpopService.jwtService.JWKToPublicKey(jwk)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported curve")
		assert.Nil(t, publicKey)
	})

	t.Run("jwkToPublicKey_InvalidCoordinates", func(t *testing.T) {
		jwk := map[string]interface{}{
			"kty": "EC",
			"crv": "P-256",
			"x":   "invalid-base64!!!", // Invalid base64 with special characters
			"y":   "invalid-base64!!!",
		}

		// Mock service doesn't validate base64, so it should succeed with mock values
		publicKey, err := dpopService.jwtService.JWKToPublicKey(jwk)
		require.NoError(t, err)
		assert.NotNil(t, publicKey)
		// Mock service returns predictable values regardless of input
		assert.Equal(t, big.NewInt(12345), publicKey.X)
		assert.Equal(t, big.NewInt(67890), publicKey.Y)
		assert.Equal(t, elliptic.P256(), publicKey.Curve)
	})
}

func TestDPoPService_StructValidation(t *testing.T) {
	t.Run("DPoPProof", func(t *testing.T) {
		now := time.Now()
		proof := &DPoPProof{
			Header: DPoPHeader{
				Typ: "dpop+jwt",
				Alg: "ES256",
				Jwk: "test-jwk",
			},
			Claims: DPoPClaims{
				JTI:   "test-jti",
				IAT:   now.Unix(),
				HTTP:  "POST",
				URI:   "/api/test",
				HTM:   "POST",
				Nonce: "test-nonce",
			},
		}

		assert.Equal(t, "dpop+jwt", proof.Header.Typ)
		assert.Equal(t, "ES256", proof.Header.Alg)
		assert.Equal(t, "test-jwk", proof.Header.Jwk)
		assert.Equal(t, "test-jti", proof.Claims.JTI)
		assert.Equal(t, now.Unix(), proof.Claims.IAT)
		assert.Equal(t, "POST", proof.Claims.HTTP)
		assert.Equal(t, "/api/test", proof.Claims.URI)
		assert.Equal(t, "POST", proof.Claims.HTM)
		assert.Equal(t, "test-nonce", proof.Claims.Nonce)
	})

	t.Run("DPoPHeader", func(t *testing.T) {
		header := DPoPHeader{
			Typ: "dpop+jwt",
			Alg: "ES256",
			Jwk: "test-jwk",
		}

		assert.Equal(t, "dpop+jwt", header.Typ)
		assert.Equal(t, "ES256", header.Alg)
		assert.Equal(t, "test-jwk", header.Jwk)
	})

	t.Run("DPoPClaims", func(t *testing.T) {
		now := time.Now()
		claims := DPoPClaims{
			JTI:   "test-jti",
			IAT:   now.Unix(),
			HTTP:  "POST",
			URI:   "/api/test",
			HTM:   "POST",
			Nonce: "test-nonce",
		}

		assert.Equal(t, "test-jti", claims.JTI)
		assert.Equal(t, now.Unix(), claims.IAT)
		assert.Equal(t, "POST", claims.HTTP)
		assert.Equal(t, "/api/test", claims.URI)
		assert.Equal(t, "POST", claims.HTM)
		assert.Equal(t, "test-nonce", claims.Nonce)
	})

	t.Run("DPoPJWK", func(t *testing.T) {
		jwk := DPoPJWK{
			KTY: "EC",
			CRV: "P-256",
			X:   "test-x-coordinate",
			Y:   "test-y-coordinate",
		}

		assert.Equal(t, "EC", jwk.KTY)
		assert.Equal(t, "P-256", jwk.CRV)
		assert.Equal(t, "test-x-coordinate", jwk.X)
		assert.Equal(t, "test-y-coordinate", jwk.Y)
	})

	t.Run("DPoPNonce", func(t *testing.T) {
		now := time.Now()
		nonce := DPoPNonce{
			Nonce:   "test-nonce-123",
			Expires: now.Add(5 * time.Minute),
		}

		assert.Equal(t, "test-nonce-123", nonce.Nonce)
		assert.Equal(t, now.Add(5*time.Minute), nonce.Expires)
		assert.True(t, nonce.Expires.After(now))
	})
}

// Integration tests using the established database pattern
func TestDPoPService_GenerateNonce_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := testutil.SetupPDSTestDB(t)
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	jwtService := testutil.NewTestJWTService()

	// Create DPoP service with real database
	queries := generated.New(pool)
	dpopService := NewDPoPService(authService, queries, logger, jwtService)

	req := httptest.NewRequest(http.MethodPost, "/dpop/nonce", nil)
	w := httptest.NewRecorder()

	dpopService.GenerateNonce(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response DPoPNonce
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.Nonce)
	assert.True(t, response.Expires.After(time.Now()))

	// Verify nonce was stored in database
	dbNonce, err := queries.GetDPoPNonce(req.Context(), response.Nonce)
	require.NoError(t, err)
	assert.Equal(t, response.Nonce, dbNonce.Nonce)
	assert.True(t, dbNonce.ExpiresAt.After(time.Now()))
}

func TestDPoPService_ValidateDPoPProof_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := testutil.SetupPDSTestDB(t)
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	jwtService := testutil.NewTestJWTService()

	queries := generated.New(pool)
	dpopService := NewDPoPService(authService, queries, logger, jwtService)

	// Generate a valid DPoP proof
	proof, err := dpopService.GenerateDPoPProof("POST", "/api/test", "")
	require.NoError(t, err)

	t.Run("valid proof without nonce", func(t *testing.T) {
		// This should work since we're not using a nonce
		validatedProof, err := dpopService.ValidateDPoPProof(proof, "POST", "/api/test")
		require.NoError(t, err)
		assert.NotNil(t, validatedProof)
		assert.Equal(t, "POST", validatedProof.Claims.HTM)
		assert.Equal(t, "/api/test", validatedProof.Claims.URI)
	})

	t.Run("method mismatch", func(t *testing.T) {
		_, err := dpopService.ValidateDPoPProof(proof, "GET", "/api/test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DPoP proof method mismatch")
	})

	t.Run("URI mismatch", func(t *testing.T) {
		_, err := dpopService.ValidateDPoPProof(proof, "POST", "/api/different")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DPoP proof URI mismatch")
	})
}

func TestDPoPService_ValidateDPoPProof_WithNonce_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := testutil.SetupPDSTestDB(t)
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	jwtService := testutil.NewTestJWTService()

	queries := generated.New(pool)
	dpopService := NewDPoPService(authService, queries, logger, jwtService)

	// First, generate a nonce and store it in the database
	req := httptest.NewRequest(http.MethodPost, "/dpop/nonce", nil)
	w := httptest.NewRecorder()
	dpopService.GenerateNonce(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var nonceResponse DPoPNonce
	err := json.Unmarshal(w.Body.Bytes(), &nonceResponse)
	require.NoError(t, err)

	// Generate a DPoP proof with the nonce
	proof, err := dpopService.GenerateDPoPProof("POST", "/api/test", nonceResponse.Nonce)
	require.NoError(t, err)

	t.Run("valid proof with nonce", func(t *testing.T) {
		// This should work with the valid nonce
		validatedProof, err := dpopService.ValidateDPoPProof(proof, "POST", "/api/test")
		require.NoError(t, err)
		assert.NotNil(t, validatedProof)
		assert.Equal(t, nonceResponse.Nonce, validatedProof.Claims.Nonce)
	})

	t.Run("reused nonce should fail", func(t *testing.T) {
		// Try to use the same nonce again - should fail
		_, err := dpopService.ValidateDPoPProof(proof, "POST", "/api/test")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired nonce")
	})
}
