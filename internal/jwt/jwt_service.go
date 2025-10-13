package jwt

import (
	"crypto/ecdsa"
	"time"
)

// JWTService defines the interface for JWT operations
// This abstraction allows different implementations for production vs testing
type JWTService interface {
	// GenerateSignedToken creates and signs a JWT token
	GenerateSignedToken(claims map[string]interface{}, header map[string]interface{}, signingKey interface{}) (string, error)

	// ValidateAndParse validates and parses a JWT token
	ValidateAndParse(token string, keyFunc func(token interface{}) (interface{}, error)) (map[string]interface{}, error)

	// ExtractPublicKey converts a JWK to an ECDSA public key
	ExtractPublicKey(jwk map[string]interface{}) (*ecdsa.PublicKey, error)

	// GenerateKeyPair generates a new ECDSA key pair for signing
	GenerateKeyPair() (*ecdsa.PrivateKey, error)

	// JWKToPublicKey converts a JWK to an ECDSA public key
	JWKToPublicKey(jwk interface{}) (*ecdsa.PublicKey, error)

	// CreateJWKFromPublicKey creates a JWK from an ECDSA public key
	CreateJWKFromPublicKey(publicKey *ecdsa.PublicKey) (interface{}, error)

	// CreateJWKFromPrivateKey creates a JWK from an ECDSA private key
	CreateJWKFromPrivateKey(privateKey *ecdsa.PrivateKey) (interface{}, error)

	// MarshalJWK marshals a JWK to JSON
	MarshalJWK(jwk interface{}) (string, error)
}

// JWTServiceConfig holds configuration for JWT service implementations
type JWTServiceConfig struct {
	// Algorithm to use for signing (e.g., "ES256", "ES256K")
	Algorithm string

	// Token expiration time
	Expiration time.Duration

	// Whether to validate signatures in test mode
	ValidateSignatures bool
}
