package pds

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOAuthService_NewOAuthService(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}

	oauthService := NewOAuthService(authService, nil, logger)

	assert.NotNil(t, oauthService)
	assert.Equal(t, authService, oauthService.authService)
	assert.Equal(t, logger, oauthService.logger)
}

func TestOAuthService_GetClientMetadata_MissingClientID(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	oauthService := NewOAuthService(authService, nil, logger)

	// Create request without client_id
	req := httptest.NewRequest(http.MethodGet, "/oauth/client_metadata", nil)
	w := httptest.NewRecorder()

	// Call handler
	oauthService.GetClientMetadata(w, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "client_id parameter required")
}

func TestOAuthService_GetClientMetadata_WrongMethod(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	oauthService := NewOAuthService(authService, nil, logger)

	// Create POST request (should be GET)
	req := httptest.NewRequest(http.MethodPost, "/oauth/client_metadata?client_id=test", nil)
	w := httptest.NewRecorder()

	// Call handler
	oauthService.GetClientMetadata(w, req)

	// Should return 405 Method Not Allowed
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestOAuthService_HandleAuthorization_WrongMethod(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	oauthService := NewOAuthService(authService, nil, logger)

	// Create POST request (should be GET)
	req := httptest.NewRequest(http.MethodPost, "/oauth/authorize", nil)
	w := httptest.NewRecorder()

	// Call handler
	oauthService.HandleAuthorization(w, req)

	// Should return 405 Method Not Allowed
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestOAuthService_HandleToken_WrongMethod(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	oauthService := NewOAuthService(authService, nil, logger)

	// Create GET request (should be POST)
	req := httptest.NewRequest(http.MethodGet, "/oauth/token", nil)
	w := httptest.NewRecorder()

	// Call handler
	oauthService.HandleToken(w, req)

	// Should return 405 Method Not Allowed
	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	assert.Contains(t, w.Body.String(), "Method not allowed")
}

func TestOAuthService_HandleToken_InvalidJSON(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	oauthService := NewOAuthService(authService, nil, logger)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	oauthService.HandleToken(w, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid JSON")
}

func TestOAuthService_HandleToken_UnsupportedGrantType(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	oauthService := NewOAuthService(authService, nil, logger)

	// Create request with unsupported grant type
	tokenRequest := TokenRequest{
		GrantType: "client_credentials", // Unsupported
		Code:      "valid-code-123",
	}
	reqBody, _ := json.Marshal(tokenRequest)
	req := httptest.NewRequest(http.MethodPost, "/oauth/token", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call handler
	oauthService.HandleToken(w, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Unsupported grant_type")
}

func TestOAuthService_HelperMethods(t *testing.T) {
	logger := testutil.CreateMockLogger()
	authService := &AuthService{logger: logger}
	oauthService := NewOAuthService(authService, nil, logger)

	t.Run("isValidRedirectURI", func(t *testing.T) {
		validURIs := []string{"https://example.com/callback", "https://app.example.com/oauth"}

		// Valid redirect URI
		assert.True(t, oauthService.isValidRedirectURI(validURIs, "https://example.com/callback"))

		// Invalid redirect URI
		assert.False(t, oauthService.isValidRedirectURI(validURIs, "https://evil.com/callback"))

		// Empty redirect URI
		assert.False(t, oauthService.isValidRedirectURI(validURIs, ""))

		// Partial match (should be exact)
		assert.False(t, oauthService.isValidRedirectURI(validURIs, "https://example.com/callback/extra"))
	})

	// Note: generateAuthorizationCode requires a database connection
	// Testing it would require proper mocking of the database layer

	t.Run("validateAuthorizationCode", func(t *testing.T) {
		// Test with empty code
		authCode, err := oauthService.validateAuthorizationCode("", "test-client-123", "https://example.com/callback")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "empty authorization code")
		assert.Nil(t, authCode)
	})
}

func TestOAuthService_AuthorizationRequest(t *testing.T) {
	// Test AuthorizationRequest struct
	req := &AuthorizationRequest{
		ClientID:            "test-client-123",
		RedirectURI:         "https://example.com/callback",
		ResponseType:        "code",
		Scope:               "read write",
		State:               "random-state",
		Nonce:               "random-nonce",
		CodeChallenge:       "challenge",
		CodeChallengeMethod: "S256",
	}

	assert.Equal(t, "test-client-123", req.ClientID)
	assert.Equal(t, "https://example.com/callback", req.RedirectURI)
	assert.Equal(t, "code", req.ResponseType)
	assert.Equal(t, "read write", req.Scope)
	assert.Equal(t, "random-state", req.State)
	assert.Equal(t, "random-nonce", req.Nonce)
	assert.Equal(t, "challenge", req.CodeChallenge)
	assert.Equal(t, "S256", req.CodeChallengeMethod)
}

func TestOAuthService_AuthorizationCode(t *testing.T) {
	now := time.Now()
	code := &AuthorizationCode{
		Code:        "auth-code-123",
		ClientID:    "test-client-123",
		UserDID:     "did:plc:test-user",
		RedirectURI: "https://example.com/callback",
		Scope:       "read write",
		ExpiresAt:   now.Add(10 * time.Minute),
		Nonce:       "test-nonce",
	}

	assert.Equal(t, "auth-code-123", code.Code)
	assert.Equal(t, "test-client-123", code.ClientID)
	assert.Equal(t, "did:plc:test-user", code.UserDID)
	assert.Equal(t, "https://example.com/callback", code.RedirectURI)
	assert.Equal(t, "read write", code.Scope)
	assert.Equal(t, "test-nonce", code.Nonce)
	assert.True(t, code.ExpiresAt.After(now))
}

func TestOAuthService_TokenRequest(t *testing.T) {
	req := &TokenRequest{
		GrantType:    "authorization_code",
		Code:         "auth-code-123",
		RedirectURI:  "https://example.com/callback",
		ClientID:     "test-client-123",
		CodeVerifier: "code-verifier",
	}

	assert.Equal(t, "authorization_code", req.GrantType)
	assert.Equal(t, "auth-code-123", req.Code)
	assert.Equal(t, "https://example.com/callback", req.RedirectURI)
	assert.Equal(t, "test-client-123", req.ClientID)
	assert.Equal(t, "code-verifier", req.CodeVerifier)
}

func TestOAuthService_TokenResponse(t *testing.T) {
	response := &TokenResponse{
		AccessToken:  "access-token-123",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
		RefreshToken: "refresh-token-123",
		Scope:        "read write",
	}

	assert.Equal(t, "access-token-123", response.AccessToken)
	assert.Equal(t, "Bearer", response.TokenType)
	assert.Equal(t, 3600, response.ExpiresIn)
	assert.Equal(t, "refresh-token-123", response.RefreshToken)
	assert.Equal(t, "read write", response.Scope)
}

func TestOAuthService_OAuthClient(t *testing.T) {
	now := time.Now()
	client := &OAuthClient{
		ClientID:      "test-client-123",
		ClientName:    "Test Client",
		RedirectURIs:  []string{"https://example.com/callback"},
		Scopes:        []string{"read", "write"},
		GrantTypes:    []string{"authorization_code"},
		ResponseTypes: []string{"code"},
		CreatedAt:     now,
	}

	assert.Equal(t, "test-client-123", client.ClientID)
	assert.Equal(t, "Test Client", client.ClientName)
	assert.Equal(t, []string{"https://example.com/callback"}, client.RedirectURIs)
	assert.Equal(t, []string{"read", "write"}, client.Scopes)
	assert.Equal(t, []string{"authorization_code"}, client.GrantTypes)
	assert.Equal(t, []string{"code"}, client.ResponseTypes)
	assert.Equal(t, now, client.CreatedAt)
}

func TestGenerateRandomString(t *testing.T) {
	// Test the generateRandomString function
	code1 := generateRandomString(32)
	code2 := generateRandomString(32)

	// Should generate different codes
	assert.NotEqual(t, code1, code2)

	// Should be base64 encoded (length should be consistent)
	assert.Len(t, code1, 44) // 32 bytes base64 encoded = 44 characters
	assert.Len(t, code2, 44)

	// Should be valid base64 (with padding)
	assert.Regexp(t, `^[A-Za-z0-9_-]+=*$`, code1)
	assert.Regexp(t, `^[A-Za-z0-9_-]+=*$`, code2)
}
