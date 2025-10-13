package testutil

import (
	"time"

	"github.com/matt0x6f/hashpost/internal/jwt"
)

// JWTServiceFactory provides factory functions for creating JWT services
// for different testing contexts
type JWTServiceFactory struct{}

// NewJWTServiceFactory creates a new JWT service factory
func NewJWTServiceFactory() *JWTServiceFactory {
	return &JWTServiceFactory{}
}

// NewProductionJWTService creates a production JWT service with real crypto
func (f *JWTServiceFactory) NewProductionJWTService() jwt.JWTService {
	config := jwt.JWTServiceConfig{
		Algorithm:          "ES256",
		Expiration:         time.Hour,
		ValidateSignatures: true,
	}
	return jwt.NewProductionJWTService(config)
}

// NewMockJWTService creates a mock JWT service for unit tests
func (f *JWTServiceFactory) NewMockJWTService() jwt.JWTService {
	return NewMockJWTService()
}

// NewTestJWTService creates a test JWT service for integration tests
func (f *JWTServiceFactory) NewTestJWTService() jwt.JWTService {
	return NewTestJWTService()
}

// NewES256KJWTService creates a JWT service for ES256K algorithm (atproto compliance)
func (f *JWTServiceFactory) NewES256KJWTService() jwt.JWTService {
	config := jwt.JWTServiceConfig{
		Algorithm:          "ES256K",
		Expiration:         time.Hour,
		ValidateSignatures: true,
	}
	return jwt.NewProductionJWTService(config)
}

// NewES256KTestJWTService creates a test JWT service for ES256K algorithm
func (f *JWTServiceFactory) NewES256KTestJWTService() jwt.JWTService {
	return NewTestJWTService()
}

// Convenience functions for direct access

// NewProductionJWTService creates a production JWT service
func NewProductionJWTService() jwt.JWTService {
	factory := NewJWTServiceFactory()
	return factory.NewProductionJWTService()
}

// NewES256KJWTService creates a JWT service for ES256K algorithm
func NewES256KJWTService() jwt.JWTService {
	factory := NewJWTServiceFactory()
	return factory.NewES256KJWTService()
}

// NewES256KTestJWTService creates a test JWT service for ES256K algorithm
func NewES256KTestJWTService() jwt.JWTService {
	factory := NewJWTServiceFactory()
	return factory.NewES256KTestJWTService()
}
