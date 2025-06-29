package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/config"
)

func TestUserContext_PseudonymContext(t *testing.T) {
	t.Run("UserContextWithPseudonym", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:            123,
			Email:             "test@example.com",
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote", "message", "report"},
			MFAEnabled:        false,
			ActivePseudonymID: "test_pseudonym_123",
			DisplayName:       "Test User",
			TokenType:         "jwt",
		}

		// Test that pseudonym context is properly set
		if userCtx.ActivePseudonymID != "test_pseudonym_123" {
			t.Errorf("Expected ActivePseudonymID 'test_pseudonym_123', got %s", userCtx.ActivePseudonymID)
		}

		if userCtx.DisplayName != "Test User" {
			t.Errorf("Expected DisplayName 'Test User', got %s", userCtx.DisplayName)
		}

		if userCtx.TokenType != "jwt" {
			t.Errorf("Expected TokenType 'jwt', got %s", userCtx.TokenType)
		}
	})

	t.Run("UserContextWithoutPseudonym", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:            123,
			Email:             "test@example.com",
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote", "message", "report"},
			MFAEnabled:        false,
			ActivePseudonymID: "", // No active pseudonym
			DisplayName:       "",
			TokenType:         "jwt",
		}

		// Test that empty pseudonym context is handled
		if userCtx.ActivePseudonymID != "" {
			t.Errorf("Expected empty ActivePseudonymID, got %s", userCtx.ActivePseudonymID)
		}

		if userCtx.DisplayName != "" {
			t.Errorf("Expected empty DisplayName, got %s", userCtx.DisplayName)
		}
	})
}

func TestJWTClaims_PseudonymContext(t *testing.T) {
	t.Run("JWTClaimsWithPseudonym", func(t *testing.T) {
		claims := &JWTClaims{
			UserID:            123,
			Email:             "test@example.com",
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote", "message", "report"},
			MFAEnabled:        false,
			ActivePseudonymID: "test_pseudonym_123",
			DisplayName:       "Test User",
		}

		// Test that pseudonym context is properly set in JWT claims
		if claims.ActivePseudonymID != "test_pseudonym_123" {
			t.Errorf("Expected ActivePseudonymID 'test_pseudonym_123', got %s", claims.ActivePseudonymID)
		}

		if claims.DisplayName != "Test User" {
			t.Errorf("Expected DisplayName 'Test User', got %s", claims.DisplayName)
		}
	})

	t.Run("JWTClaimsWithoutPseudonym", func(t *testing.T) {
		claims := &JWTClaims{
			UserID:            123,
			Email:             "test@example.com",
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote", "message", "report"},
			MFAEnabled:        false,
			ActivePseudonymID: "", // No active pseudonym
			DisplayName:       "",
		}

		// Test that empty pseudonym context is handled in JWT claims
		if claims.ActivePseudonymID != "" {
			t.Errorf("Expected empty ActivePseudonymID, got %s", claims.ActivePseudonymID)
		}

		if claims.DisplayName != "" {
			t.Errorf("Expected empty DisplayName, got %s", claims.DisplayName)
		}
	})
}

func TestGenerateJWT_PseudonymContext(t *testing.T) {
	t.Run("GenerateJWTWithPseudonym", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:            123,
			Email:             "test@example.com",
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote", "message", "report"},
			MFAEnabled:        false,
			ActivePseudonymID: "test_pseudonym_123",
			DisplayName:       "Test User",
			TokenType:         "jwt",
		}

		jwtSecret := "test-jwt-secret"
		token, err := GenerateJWT(userCtx, jwtSecret, 24*time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate JWT: %v", err)
		}

		// Verify token was generated
		if token == "" {
			t.Error("Expected non-empty JWT token")
		}

		// Parse and verify the token contains pseudonym context
		jwtConfig := &config.JWTConfig{
			Secret:      "test-jwt-secret",
			Expiration:  24 * time.Hour,
			Development: true,
		}
		securityConfig := &config.SecurityConfig{
			EnableMFA: false,
		}
		authMiddleware := NewAuthMiddleware(jwtSecret, nil, jwtConfig, securityConfig)

		// Create a mock request with the JWT token
		req, err := http.NewRequest("GET", "/test", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add JWT as cookie
		req.AddCookie(&http.Cookie{
			Name:  AccessTokenCookie,
			Value: token,
		})

		// Extract user context from token
		extractedUser, err := authMiddleware.ExtractUserFromToken(req)
		if err != nil {
			t.Fatalf("Failed to extract user from token: %v", err)
		}

		// Verify pseudonym context is preserved
		if extractedUser.ActivePseudonymID != userCtx.ActivePseudonymID {
			t.Errorf("Expected ActivePseudonymID %s, got %s", userCtx.ActivePseudonymID, extractedUser.ActivePseudonymID)
		}

		if extractedUser.DisplayName != userCtx.DisplayName {
			t.Errorf("Expected DisplayName %s, got %s", userCtx.DisplayName, extractedUser.DisplayName)
		}

		if extractedUser.TokenType != "jwt" {
			t.Errorf("Expected TokenType 'jwt', got %s", extractedUser.TokenType)
		}
	})

	t.Run("GenerateJWTWithoutPseudonym", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:            123,
			Email:             "test@example.com",
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote", "message", "report"},
			MFAEnabled:        false,
			ActivePseudonymID: "", // No active pseudonym
			DisplayName:       "",
			TokenType:         "jwt",
		}

		jwtSecret := "test-jwt-secret"
		token, err := GenerateJWT(userCtx, jwtSecret, 24*time.Hour)
		if err != nil {
			t.Fatalf("Failed to generate JWT: %v", err)
		}

		// Verify token was generated
		if token == "" {
			t.Error("Expected non-empty JWT token")
		}

		// Parse and verify the token handles empty pseudonym context
		jwtConfig := &config.JWTConfig{
			Secret:      "test-jwt-secret",
			Expiration:  24 * time.Hour,
			Development: true,
		}
		securityConfig := &config.SecurityConfig{
			EnableMFA: false,
		}
		authMiddleware := NewAuthMiddleware(jwtSecret, nil, jwtConfig, securityConfig)

		// Create a mock request with the JWT token
		req, err := http.NewRequest("GET", "/test", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		// Add JWT as cookie
		req.AddCookie(&http.Cookie{
			Name:  AccessTokenCookie,
			Value: token,
		})

		// Extract user context from token
		extractedUser, err := authMiddleware.ExtractUserFromToken(req)
		if err != nil {
			t.Fatalf("Failed to extract user from token: %v", err)
		}

		// Verify empty pseudonym context is preserved
		if extractedUser.ActivePseudonymID != "" {
			t.Errorf("Expected empty ActivePseudonymID, got %s", extractedUser.ActivePseudonymID)
		}

		if extractedUser.DisplayName != "" {
			t.Errorf("Expected empty DisplayName, got %s", extractedUser.DisplayName)
		}
	})
}

func TestUserContext_ModerationContext(t *testing.T) {
	t.Run("ModeratorWithPseudonym", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:            123,
			Email:             "moderator@example.com",
			Roles:             []string{"moderator"},
			Capabilities:      []string{"moderate_content", "ban_users", "delete_posts"},
			MFAEnabled:        false,
			ActivePseudonymID: "moderator_pseudonym_123",
			DisplayName:       "Moderator User",
			TokenType:         "jwt",
		}

		// Test that moderator has correct roles
		if !userCtx.HasRole("moderator") {
			t.Error("Expected user to have moderator role")
		}

		// Test that moderator has moderation capabilities
		if !userCtx.HasCapability("moderate_content") {
			t.Error("Expected user to have moderate_content capability")
		}

		// Test that pseudonym context is available for moderation actions
		if userCtx.ActivePseudonymID == "" {
			t.Error("Expected moderator to have active pseudonym for moderation actions")
		}
	})

	t.Run("RegularUserWithoutModeration", func(t *testing.T) {
		userCtx := &UserContext{
			UserID:            456,
			Email:             "user@example.com",
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote", "message", "report"},
			MFAEnabled:        false,
			ActivePseudonymID: "user_pseudonym_456",
			DisplayName:       "Regular User",
			TokenType:         "jwt",
		}

		// Test that regular user doesn't have moderator role
		if userCtx.HasRole("moderator") {
			t.Error("Expected user to not have moderator role")
		}

		// Test that regular user doesn't have moderation capabilities
		if userCtx.HasCapability("moderate_content") {
			t.Error("Expected user to not have moderate_content capability")
		}

		// Test that user still has pseudonym context for regular actions
		if userCtx.ActivePseudonymID == "" {
			t.Error("Expected user to have active pseudonym for regular actions")
		}
	})
}

func TestUserContext_MultiplePseudonyms(t *testing.T) {
	t.Run("UserWithMultiplePseudonyms", func(t *testing.T) {
		// Test scenario where a user has multiple pseudonyms
		// This would typically be tested in integration tests, but we can test the context here

		userCtx := &UserContext{
			UserID:            789,
			Email:             "multipseud@example.com",
			Roles:             []string{"user"},
			Capabilities:      []string{"create_content", "vote", "message", "report"},
			MFAEnabled:        false,
			ActivePseudonymID: "primary_pseudonym_789", // Currently active pseudonym
			DisplayName:       "Primary User",
			TokenType:         "jwt",
		}

		// Test that the active pseudonym is set
		if userCtx.ActivePseudonymID != "primary_pseudonym_789" {
			t.Errorf("Expected active pseudonym 'primary_pseudonym_789', got %s", userCtx.ActivePseudonymID)
		}

		// Test that the display name matches the active pseudonym
		if userCtx.DisplayName != "Primary User" {
			t.Errorf("Expected display name 'Primary User', got %s", userCtx.DisplayName)
		}

		// In a real scenario, the user could switch to a different pseudonym
		// This would require a separate endpoint to update the JWT with a new active pseudonym
		userCtx.ActivePseudonymID = "secondary_pseudonym_789"
		userCtx.DisplayName = "Secondary User"

		// Test that the pseudonym context can be updated
		if userCtx.ActivePseudonymID != "secondary_pseudonym_789" {
			t.Errorf("Expected updated active pseudonym 'secondary_pseudonym_789', got %s", userCtx.ActivePseudonymID)
		}

		if userCtx.DisplayName != "Secondary User" {
			t.Errorf("Expected updated display name 'Secondary User', got %s", userCtx.DisplayName)
		}
	})
}
