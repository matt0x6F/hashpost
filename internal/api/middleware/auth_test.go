package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/config"
)

func TestGenerateJWT(t *testing.T) {
	userCtx := &UserContext{
		UserID:            123,
		Email:             "test@example.com",
		Roles:             []string{"user"},
		Capabilities:      []string{"create_content", "vote", "message", "report"},
		MFAEnabled:        false,
		ActivePseudonymID: "test_pseudonym_123",
		DisplayName:       "Test User",
	}

	jwtSecret := "test-jwt-secret"
	token, err := GenerateJWT(userCtx, jwtSecret, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	t.Logf("Generated JWT token: %s", token)
	t.Logf("Token length: %d characters", len(token))

	// Verify token can be parsed
	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: false,
	}
	authMiddleware := NewAuthMiddleware(jwtSecret, nil, jwtConfig, securityConfig)

	// Test cookie-based JWT
	req, err := http.NewRequest("GET", "/web/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add JWT as cookie
	req.AddCookie(&http.Cookie{
		Name:  AccessTokenCookie,
		Value: token,
	})

	extractedUser, err := authMiddleware.ExtractUserFromToken(req)
	if err != nil {
		t.Fatalf("JWT validation failed: %v", err)
	}

	// Verify extracted user
	if extractedUser.UserID != userCtx.UserID {
		t.Errorf("Expected UserID %d, got %d", userCtx.UserID, extractedUser.UserID)
	}

	if extractedUser.Email != userCtx.Email {
		t.Errorf("Expected Email %s, got %s", userCtx.Email, extractedUser.Email)
	}

	if extractedUser.TokenType != "jwt" {
		t.Errorf("Expected TokenType 'jwt', got %s", extractedUser.TokenType)
	}

	t.Logf("Successfully extracted user from JWT cookie: %+v", extractedUser)
}

func TestExtractUserFromToken_HeaderAPI(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: false,
	}
	authMiddleware := NewAuthMiddleware("test-jwt-secret", nil, jwtConfig, securityConfig)

	// Test API token in header
	req, err := http.NewRequest("GET", "/api/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add API token as Authorization header
	req.Header.Set("Authorization", "Bearer api-token-123")

	_, err = authMiddleware.ExtractUserFromToken(req)
	if err == nil {
		t.Fatal("Expected error for API token validation")
	}

	t.Logf("API token validation error (expected): %v", err)
}

func TestExtractUserFromToken_CookieJWT(t *testing.T) {
	userCtx := &UserContext{
		UserID:            123,
		Email:             "test@example.com",
		Roles:             []string{"user"},
		Capabilities:      []string{"create_content", "vote", "message", "report"},
		MFAEnabled:        false,
		ActivePseudonymID: "test_pseudonym_123",
		DisplayName:       "Test User",
	}

	jwtSecret := "test-jwt-secret"
	token, err := GenerateJWT(userCtx, jwtSecret, 24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate JWT: %v", err)
	}

	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: false,
	}
	authMiddleware := NewAuthMiddleware(jwtSecret, nil, jwtConfig, securityConfig)

	// Test cookie-based JWT
	req, err := http.NewRequest("GET", "/web/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add JWT as cookie
	req.AddCookie(&http.Cookie{
		Name:  AccessTokenCookie,
		Value: token,
	})

	extractedUser, err := authMiddleware.ExtractUserFromToken(req)
	if err != nil {
		t.Fatalf("JWT validation failed: %v", err)
	}

	// Verify extracted user
	if extractedUser.UserID != userCtx.UserID {
		t.Errorf("Expected UserID %d, got %d", userCtx.UserID, extractedUser.UserID)
	}

	if extractedUser.Email != userCtx.Email {
		t.Errorf("Expected Email %s, got %s", userCtx.Email, extractedUser.Email)
	}

	if extractedUser.TokenType != "jwt" {
		t.Errorf("Expected TokenType 'jwt', got %s", extractedUser.TokenType)
	}

	t.Logf("Successfully extracted user from JWT cookie: %+v", extractedUser)
}

func TestExtractUserFromToken_NoAuth(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: false,
	}
	authMiddleware := NewAuthMiddleware("test-jwt-secret", nil, jwtConfig, securityConfig)

	// Test no authentication
	req, err := http.NewRequest("GET", "/public/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	_, err = authMiddleware.ExtractUserFromToken(req)
	if err == nil {
		t.Fatal("Expected error for request without authentication")
	}

	t.Logf("No auth error (expected): %v", err)
}

func TestExtractUserFromToken_InvalidJWT(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: false,
	}
	authMiddleware := NewAuthMiddleware("test-jwt-secret", nil, jwtConfig, securityConfig)

	// Test invalid JWT token
	req, err := http.NewRequest("GET", "/web/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add invalid JWT as cookie
	req.AddCookie(&http.Cookie{
		Name:  AccessTokenCookie,
		Value: "invalid.jwt.token",
	})

	_, err = authMiddleware.ExtractUserFromToken(req)
	if err == nil {
		t.Fatal("Expected error for invalid JWT token")
	}

	t.Logf("Invalid JWT error (expected): %v", err)
}

func TestUserContext_HasCapability(t *testing.T) {
	userCtx := &UserContext{
		Capabilities: []string{"create_content", "vote", "message", "report"},
	}

	tests := []struct {
		name       string
		capability string
		expected   bool
	}{
		{"has create_content", "create_content", true},
		{"has vote", "vote", true},
		{"has message", "message", true},
		{"has report", "report", true},
		{"does not have admin", "admin", false},
		{"does not have empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userCtx.HasCapability(tt.capability)
			if result != tt.expected {
				t.Errorf("HasCapability(%s) = %v, expected %v", tt.capability, result, tt.expected)
			}
		})
	}
}

func TestUserContext_HasRole(t *testing.T) {
	userCtx := &UserContext{
		Roles: []string{"user", "moderator"},
	}

	tests := []struct {
		name     string
		role     string
		expected bool
	}{
		{"has user", "user", true},
		{"has moderator", "moderator", true},
		{"does not have admin", "admin", false},
		{"does not have empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userCtx.HasRole(tt.role)
			if result != tt.expected {
				t.Errorf("HasRole(%s) = %v, expected %v", tt.role, result, tt.expected)
			}
		})
	}
}

func TestUserContext_RequiresMFA(t *testing.T) {
	// Set up global auth middleware with MFA enabled for testing
	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: true, // Enable MFA for this test
	}
	authMiddleware := NewAuthMiddleware("test-jwt-secret", nil, jwtConfig, securityConfig)
	SetGlobalAuthMiddleware(authMiddleware)

	userCtx := &UserContext{
		Capabilities: []string{"create_content", "vote", "message", "report"},
	}

	tests := []struct {
		name     string
		action   string
		expected bool
	}{
		{"correlate_identities requires MFA", "correlate_identities", true},
		{"system_admin requires MFA", "system_admin", true},
		{"legal_compliance requires MFA", "legal_compliance", true},
		{"create_content does not require MFA", "create_content", false},
		{"vote does not require MFA", "vote", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userCtx.RequiresMFA(tt.action)
			if result != tt.expected {
				t.Errorf("RequiresMFA(%s) = %v, expected %v", tt.action, result, tt.expected)
			}
		})
	}
}

func TestUserContext_RequiresMFA_WithCorrelationCapabilities(t *testing.T) {
	// Set up global auth middleware with MFA enabled for testing
	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: true, // Enable MFA for this test
	}
	authMiddleware := NewAuthMiddleware("test-jwt-secret", nil, jwtConfig, securityConfig)
	SetGlobalAuthMiddleware(authMiddleware)

	userCtx := &UserContext{
		Capabilities: []string{"create_content", "vote", "correlate_fingerprints", "correlate_identities"},
	}

	tests := []struct {
		name     string
		action   string
		expected bool
	}{
		{"correlate_fingerprints requires MFA", "correlate_fingerprints", true},
		{"correlate_identities requires MFA", "correlate_identities", true},
		{"create_content does not require MFA", "create_content", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userCtx.RequiresMFA(tt.action)
			if result != tt.expected {
				t.Errorf("RequiresMFA(%s) = %v, expected %v", tt.action, result, tt.expected)
			}
		})
	}
}

func TestUserContext_RequiresMFA_Disabled(t *testing.T) {
	// Set up global auth middleware with MFA disabled for testing
	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: false, // Disable MFA for this test
	}
	authMiddleware := NewAuthMiddleware("test-jwt-secret", nil, jwtConfig, securityConfig)
	SetGlobalAuthMiddleware(authMiddleware)

	userCtx := &UserContext{
		Capabilities: []string{"create_content", "vote", "correlate_fingerprints", "correlate_identities"},
	}

	tests := []struct {
		name     string
		action   string
		expected bool
	}{
		{"correlate_fingerprints does not require MFA when disabled", "correlate_fingerprints", false},
		{"correlate_identities does not require MFA when disabled", "correlate_identities", false},
		{"system_admin does not require MFA when disabled", "system_admin", false},
		{"legal_compliance does not require MFA when disabled", "legal_compliance", false},
		{"create_content does not require MFA", "create_content", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userCtx.RequiresMFA(tt.action)
			if result != tt.expected {
				t.Errorf("RequiresMFA(%s) = %v, expected %v", tt.action, result, tt.expected)
			}
		})
	}
}
