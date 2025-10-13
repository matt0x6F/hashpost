package pds

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bluesky-social/indigo/atproto/crypto"
	"github.com/bluesky-social/indigo/atproto/identity"
	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthService_GenerateTokens(t *testing.T) {
	// Create auth service
	logger := testutil.CreateMockLogger()
	jwtService := testutil.NewMockJWTService()
	authService := &AuthService{
		logger:     logger,
		serverDID:  "did:plc:hashpost-server",
		signingKey: generateTestSigningKey(t),
		jwtService: jwtService,
	}

	// Create test session
	session := &Session{
		ID:        "test-session-id",
		DID:       "did:plc:test-user",
		Handle:    "testuser.hashpost.local",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Test token generation
	accessToken, refreshToken, err := authService.GenerateTokens(session)

	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.NotEqual(t, accessToken, refreshToken)

	// Validate token structure (basic checks)
	assert.Contains(t, accessToken, ".")
	assert.Contains(t, refreshToken, ".")
	assert.Len(t, strings.Split(accessToken, "."), 3)  // JWT has 3 parts
	assert.Len(t, strings.Split(refreshToken, "."), 3) // JWT has 3 parts
}

func TestAuthService_ValidateToken(t *testing.T) {
	// Create auth service
	logger := testutil.CreateMockLogger()
	jwtService := testutil.NewMockJWTService()
	authService := &AuthService{
		logger:     logger,
		serverDID:  "did:plc:hashpost-server",
		signingKey: generateTestSigningKey(t),
		jwtService: jwtService,
	}

	// Create test session
	session := &Session{
		ID:        "test-session-id",
		DID:       "did:plc:test-user",
		Handle:    "testuser.hashpost.local",
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Generate tokens
	accessToken, _, err := authService.GenerateTokens(session)
	require.NoError(t, err)

	// Test token validation
	validatedSession, err := authService.ValidateToken(accessToken)

	require.NoError(t, err)
	require.NotNil(t, validatedSession)
	assert.Equal(t, session.DID, validatedSession.DID)
	assert.Equal(t, session.Handle, validatedSession.Handle)
}

func TestAuthService_ValidateToken_InvalidToken(t *testing.T) {
	// Create auth service
	logger := testutil.CreateMockLogger()
	jwtService := testutil.NewMockJWTService()
	authService := &AuthService{
		logger:     logger,
		serverDID:  "did:plc:hashpost-server",
		signingKey: generateTestSigningKey(t),
		jwtService: jwtService,
	}

	// Test invalid token
	_, err := authService.ValidateToken("invalid.token.here")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "token validation failed")
}

func TestAuthService_HashPassword(t *testing.T) {
	// Create auth service
	logger := testutil.CreateMockLogger()
	authService := &AuthService{
		logger: logger,
	}

	password := "testpassword123"

	// Test password hashing
	hash, err := authService.HashPassword(password)

	require.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	assert.Contains(t, hash, "$2a$") // bcrypt hash format
}

func TestAuthService_ValidatePasswordStrength(t *testing.T) {
	// Create auth service
	logger := testutil.CreateMockLogger()
	authService := &AuthService{
		logger: logger,
	}

	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid password",
			password:    "testpassword123",
			expectError: false,
		},
		{
			name:        "password too short",
			password:    "short",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
		{
			name:        "empty password",
			password:    "",
			expectError: true,
			errorMsg:    "password must be at least 8 characters long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := authService.ValidatePasswordStrength(tt.password)

			if tt.expectError {
				require.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAuthService_ResolveHandle(t *testing.T) {
	// Create auth service with real directory
	logger := testutil.CreateMockLogger()
	authService := &AuthService{
		directory: identity.DefaultDirectory(),
		logger:    logger,
	}

	// Test with invalid handle (should return error)
	_, err := authService.ResolveHandle(context.Background(), "nonexistent.hashpost.local")
	require.Error(t, err)
}

func TestAuthService_ResolveDID(t *testing.T) {
	// Create auth service with real directory
	logger := testutil.CreateMockLogger()
	authService := &AuthService{
		directory: identity.DefaultDirectory(),
		logger:    logger,
	}

	// Test with invalid DID (should return error)
	_, err := authService.ResolveDID(context.Background(), "did:plc:nonexistent")
	require.Error(t, err)
}

func TestAuthService_AddUserToMockDirectory(t *testing.T) {
	// Create auth service with real directory (not mock)
	logger := testutil.CreateMockLogger()
	authService := &AuthService{
		directory: identity.DefaultDirectory(),
		logger:    logger,
	}

	// Test adding user to real directory (should fail)
	err := authService.AddUserToMockDirectory(context.Background(), "did:plc:new-user", "newuser.hashpost.local")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a mock directory")
}

// Helper function to generate a test signing key
func generateTestSigningKey(t *testing.T) crypto.PrivateKey {
	key, err := crypto.GeneratePrivateKeyK256()
	require.NoError(t, err)
	return key
}
