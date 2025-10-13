package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ProductionJWTService implements JWTService using real crypto operations
type ProductionJWTService struct {
	config JWTServiceConfig
}

// NewProductionJWTService creates a new production JWT service
func NewProductionJWTService(config JWTServiceConfig) *ProductionJWTService {
	if config.Algorithm == "" {
		config.Algorithm = "ES256"
	}
	if config.Expiration == 0 {
		config.Expiration = time.Hour
	}
	if !config.ValidateSignatures {
		config.ValidateSignatures = true // Always validate in production
	}

	return &ProductionJWTService{
		config: config,
	}
}

// GenerateSignedToken creates and signs a JWT token using real crypto
func (p *ProductionJWTService) GenerateSignedToken(claims map[string]interface{}, header map[string]interface{}, signingKey interface{}) (string, error) {
	// Create JWT token with claims
	token := jwt.NewWithClaims(jwt.GetSigningMethod(p.config.Algorithm), jwt.MapClaims(claims))

	// Add custom header fields
	for key, value := range header {
		token.Header[key] = value
	}

	// Sign the token
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

// ValidateAndParse validates and parses a JWT token using real crypto validation
func (p *ProductionJWTService) ValidateAndParse(token string, keyFunc func(token interface{}) (interface{}, error)) (map[string]interface{}, error) {
	// Parse and validate JWT token
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method matches our algorithm
		if token.Method.Alg() != p.config.Algorithm {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get the key using the provided key function
		return keyFunc(token)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !parsedToken.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	return map[string]interface{}(claims), nil
}

// ExtractPublicKey converts a JWK to an ECDSA public key using real crypto
func (p *ProductionJWTService) ExtractPublicKey(jwk map[string]interface{}) (*ecdsa.PublicKey, error) {
	// Extract JWK fields
	kty, ok := jwk["kty"].(string)
	if !ok || kty != "EC" {
		return nil, fmt.Errorf("unsupported key type: %v", kty)
	}

	crv, ok := jwk["crv"].(string)
	if !ok || crv != "P-256" {
		return nil, fmt.Errorf("unsupported curve: %v", crv)
	}

	xStr, ok := jwk["x"].(string)
	if !ok {
		return nil, fmt.Errorf("missing x coordinate")
	}

	yStr, ok := jwk["y"].(string)
	if !ok {
		return nil, fmt.Errorf("missing y coordinate")
	}

	// Decode coordinates
	xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return nil, fmt.Errorf("invalid x coordinate: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return nil, fmt.Errorf("invalid y coordinate: %w", err)
	}

	// Create public key
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}

	return publicKey, nil
}

// GenerateKeyPair generates a new ECDSA key pair for signing
func (p *ProductionJWTService) GenerateKeyPair() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

// JWKToPublicKey converts a JWK to an ECDSA public key
func (p *ProductionJWTService) JWKToPublicKey(jwk interface{}) (*ecdsa.PublicKey, error) {
	// Convert interface{} to map[string]interface{}
	jwkMap, ok := jwk.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JWK type")
	}

	return p.ExtractPublicKey(jwkMap)
}

// CreateJWKFromPublicKey creates a JWK from an ECDSA public key
func (p *ProductionJWTService) CreateJWKFromPublicKey(publicKey *ecdsa.PublicKey) (interface{}, error) {
	// Get coordinates and pad to 32 bytes if needed
	xBytes := publicKey.X.Bytes()
	yBytes := publicKey.Y.Bytes()

	// Pad to 32 bytes if needed
	if len(xBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(xBytes):], xBytes)
		xBytes = padded
	}
	if len(yBytes) < 32 {
		padded := make([]byte, 32)
		copy(padded[32-len(yBytes):], yBytes)
		yBytes = padded
	}

	jwk := map[string]interface{}{
		"kty": "EC",
		"crv": "P-256",
		"x":   base64.RawURLEncoding.EncodeToString(xBytes),
		"y":   base64.RawURLEncoding.EncodeToString(yBytes),
	}

	return jwk, nil
}

// CreateJWKFromPrivateKey creates a JWK from an ECDSA private key
func (p *ProductionJWTService) CreateJWKFromPrivateKey(privateKey *ecdsa.PrivateKey) (interface{}, error) {
	return p.CreateJWKFromPublicKey(&privateKey.PublicKey)
}

// MarshalJWK marshals a JWK to JSON
func (p *ProductionJWTService) MarshalJWK(jwk interface{}) (string, error) {
	jwkBytes, err := json.Marshal(jwk)
	if err != nil {
		return "", fmt.Errorf("failed to marshal jwk: %w", err)
	}
	return string(jwkBytes), nil
}
