package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/config"
)

func TestDualAuthentication_HeaderAPIToken(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: false,
	}
	authMiddleware := NewAuthMiddleware("test-jwt-secret", nil, jwtConfig, securityConfig)
	SetGlobalAuthMiddleware(authMiddleware)

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
	t.Logf("This error is expected since API token validation is not yet implemented")
}

func TestDualAuthentication_CookieJWT(t *testing.T) {
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
	SetGlobalAuthMiddleware(authMiddleware)

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

func TestDualAuthentication_Priority(t *testing.T) {
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
	SetGlobalAuthMiddleware(authMiddleware)

	// Test that header takes priority over cookie
	req, err := http.NewRequest("GET", "/api/test", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Add both API token in header and JWT in cookie
	req.Header.Set("Authorization", "Bearer api-token-123")
	req.AddCookie(&http.Cookie{
		Name:  AccessTokenCookie,
		Value: token,
	})

	_, err = authMiddleware.ExtractUserFromToken(req)
	if err == nil {
		t.Fatal("Expected error for API token validation")
	}

	t.Logf("API token validation error (expected): %v", err)
	t.Logf("This confirms that header-based tokens take priority over cookie-based tokens")
}

func TestDualAuthentication_NoAuth(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: false,
	}
	authMiddleware := NewAuthMiddleware("test-jwt-secret", nil, jwtConfig, securityConfig)
	SetGlobalAuthMiddleware(authMiddleware)

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

func TestDualAuthentication_InvalidJWT(t *testing.T) {
	jwtConfig := &config.JWTConfig{
		Secret:      "test-jwt-secret",
		Expiration:  24 * time.Hour,
		Development: true,
	}
	securityConfig := &config.SecurityConfig{
		EnableMFA: false,
	}
	authMiddleware := NewAuthMiddleware("test-jwt-secret", nil, jwtConfig, securityConfig)
	SetGlobalAuthMiddleware(authMiddleware)

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

func TestCookieManagement(t *testing.T) {
	// Create a test user context
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
	expiration := 24 * time.Hour

	// Generate JWT tokens
	accessToken, err := GenerateJWT(userCtx, jwtSecret, expiration)
	if err != nil {
		t.Fatalf("Failed to generate access token: %v", err)
	}

	refreshToken, err := GenerateJWT(userCtx, jwtSecret, 7*24*time.Hour)
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	// Create a mock response writer for testing cookie functions
	mockWriter := &mockResponseWriter{
		headers: make(http.Header),
	}

	// Test setting cookies
	SetJWTCookies(mockWriter, accessToken, refreshToken, 1*time.Hour, 7*24*time.Hour)

	// Check that cookies were set in headers
	cookies := mockWriter.headers.Values("Set-Cookie")
	if len(cookies) != 2 {
		t.Errorf("Expected 2 Set-Cookie headers, got %d", len(cookies))
		return
	}

	// Verify access token cookie header
	accessCookieHeader := cookies[0]
	if !contains(accessCookieHeader, AccessTokenCookie) {
		t.Errorf("Expected access token cookie header to contain %s, got %s", AccessTokenCookie, accessCookieHeader)
	}
	if !contains(accessCookieHeader, accessToken) {
		t.Errorf("Expected access token cookie header to contain token value")
	}
	if !contains(accessCookieHeader, "HttpOnly") {
		t.Error("Expected HttpOnly flag in access token cookie")
	}

	// Verify refresh token cookie header
	refreshCookieHeader := cookies[1]
	if !contains(refreshCookieHeader, RefreshTokenCookie) {
		t.Errorf("Expected refresh token cookie header to contain %s, got %s", RefreshTokenCookie, refreshCookieHeader)
	}
	if !contains(refreshCookieHeader, refreshToken) {
		t.Errorf("Expected refresh token cookie header to contain token value")
	}
	if !contains(refreshCookieHeader, "HttpOnly") {
		t.Error("Expected HttpOnly flag in refresh token cookie")
	}

	t.Logf("Successfully set %d cookies", len(cookies))

	// Test clearing cookies
	ClearJWTCookies(mockWriter)

	// Check that cookies were cleared
	clearedCookies := mockWriter.headers.Values("Set-Cookie")
	if len(clearedCookies) != 4 {
		t.Errorf("Expected 4 Set-Cookie headers after clearing (2 set + 2 cleared), got %d", len(clearedCookies))
		return
	}

	// Verify cleared cookies have empty values
	for i := 2; i < 4; i++ {
		cookieHeader := clearedCookies[i]
		if !(contains(cookieHeader, "Max-Age=0") || contains(cookieHeader, "Max-Age=-1")) {
			t.Errorf("Expected cleared cookie to have Max-Age=0 or Max-Age=-1, got %s", cookieHeader)
		}
	}

	t.Logf("Successfully cleared cookies")
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// mockResponseWriter is a simple mock for testing cookie functions
type mockResponseWriter struct {
	headers http.Header
}

func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

func (m *mockResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {}
