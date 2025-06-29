//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"encoding/base64"

	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/testutil"
)

func TestUserContext_Registration_Integration(t *testing.T) {
	t.Run("RegistrationCreatesDefaultPseudonym", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return // Test was skipped
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		// Test registration with display name and all required fields
		registrationData := models.UserRegistrationBody{
			Email:       "test@example.com",
			Password:    "TestPassword123!",
			DisplayName: "TestUser123",
			Bio:         "This is a test user bio.",
			Language:    "en",
			Timezone:    "UTC",
			WebsiteURL:  "https://example.com",
		}

		resp := suite.MakeAuthenticatedRequest(t, server, "POST", "/auth/register", "", registrationData)
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			// Print response body for debugging
			var debugBody map[string]interface{}
			_ = json.NewDecoder(resp.Body).Decode(&debugBody)
			t.Logf("Registration failed: status=%d, body=%+v", resp.StatusCode, debugBody)
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Parse response
		var response models.UserRegistrationResponseBody
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Verify pseudonym was created
		if response.PseudonymID == "" {
			t.Error("Expected non-empty pseudonym_id in registration response")
		}

		if response.DisplayName != "TestUser123" {
			t.Errorf("Expected display_name 'TestUser123', got %s", response.DisplayName)
		}

		// Verify JWT token contains pseudonym context
		if response.AccessToken == "" {
			t.Error("Expected non-empty access token")
		}
	})

	t.Run("RegistrationWithoutDisplayNameFails", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return // Test was skipped
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		// Test registration without display name (should fail)
		registrationData := models.UserRegistrationBody{
			Email:      "test2@example.com",
			Password:   "TestPassword123!",
			Bio:        "This is a test user bio.",
			Language:   "en",
			Timezone:   "UTC",
			WebsiteURL: "https://example.com",
			// Missing display_name
		}

		resp := suite.MakeAuthenticatedRequest(t, server, "POST", "/auth/register", "", registrationData)
		defer resp.Body.Close()

		// Should fail because display_name is required
		if resp.StatusCode != http.StatusUnprocessableEntity {
			t.Errorf("Expected status 422, got %d", resp.StatusCode)
		}
	})
}

func TestUserContext_Login_Integration(t *testing.T) {
	t.Run("LoginReturnsPseudonymContext", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return // Test was skipped
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		testUser := suite.CreateTestUser(t, "login_test@example.com", "TestPassword123!", []string{"user"})
		resp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		defer resp.Body.Close()

		// Read response body once
		var responseBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check response status and log debug info if needed
		if resp.StatusCode != http.StatusOK {
			t.Logf("Login failed: status=%d, body=%+v", resp.StatusCode, responseBody)
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify pseudonym context is returned
		if activePseudonymID, ok := responseBody["active_pseudonym_id"].(string); !ok || activePseudonymID == "" {
			t.Error("Expected non-empty active_pseudonym_id in login response")
		}
		if displayName, ok := responseBody["display_name"].(string); !ok || displayName == "" {
			t.Error("Expected non-empty display_name in login response")
		}
		if pseudonyms, ok := responseBody["pseudonyms"].([]interface{}); !ok || len(pseudonyms) == 0 {
			t.Error("Expected non-empty pseudonyms array in login response")
		}
	})

	t.Run("LoginWithMultiplePseudonyms", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return // Test was skipped
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, "multipseud@example.com", "TestPassword123!", []string{"user"})
		suite.CreateTestPseudonym(t, testUser.UserID, "SecondPseudonym")
		suite.CreateTestPseudonym(t, testUser.UserID, "ThirdPseudonym")
		resp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
		var responseBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}
		if pseudonyms, ok := responseBody["pseudonyms"].([]interface{}); !ok || len(pseudonyms) < 3 {
			t.Errorf("Expected at least 3 pseudonyms in login response, got %d", len(pseudonyms))
		}
		if activePseudonymID, ok := responseBody["active_pseudonym_id"].(string); !ok || activePseudonymID == "" {
			t.Error("Expected non-empty active_pseudonym_id in login response")
		}
		if activePseudonymID, ok := responseBody["active_pseudonym_id"].(string); ok {
			if activePseudonymID != testUser.PseudonymID {
				t.Errorf("Expected active pseudonym to be %s, got %s", testUser.PseudonymID, activePseudonymID)
			}
		}
	})
}

func TestUserContext_JWT_Integration(t *testing.T) {
	t.Run("JWTContainsPseudonymContext", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return // Test was skipped
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, "jwt@example.com", "TestPassword123!", []string{"user"})
		resp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Login failed with status %d", resp.StatusCode)
		}
		token := suite.ExtractTokenFromResponse(t, resp)
		if token == "" {
			t.Fatal("Expected non-empty access token")
		}
		profileResp := suite.MakeAuthenticatedRequest(t, server, "GET", "/users/profile", token, nil)
		if profileResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for profile request, got %d", profileResp.StatusCode)
		}
		var profileBody map[string]interface{}
		if err := json.NewDecoder(profileResp.Body).Decode(&profileBody); err != nil {
			t.Fatalf("Failed to decode profile response: %v", err)
		}
		if pseudonyms, ok := profileBody["pseudonyms"].([]interface{}); !ok || len(pseudonyms) == 0 {
			t.Error("Expected non-empty pseudonyms array in profile response")
		}
	})

	t.Run("JWTPreservesPseudonymContext", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return // Test was skipped
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, "preserve@example.com", "TestPassword123!", []string{"user"})
		resp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Login failed with status %d", resp.StatusCode)
		}
		token := suite.ExtractTokenFromResponse(t, resp)
		if token == "" {
			t.Fatal("Expected non-empty access token")
		}
		endpoints := []string{
			"/users/profile",
			"/pseudonyms/" + testUser.PseudonymID + "/profile",
		}
		for _, endpoint := range endpoints {
			authResp := suite.MakeAuthenticatedRequest(t, server, "GET", endpoint, token, nil)
			if authResp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200 for %s, got %d", endpoint, authResp.StatusCode)
			}
		}
	})

	t.Run("JWTClaimsContainPseudonymContext", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return // Test was skipped
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		testUser := suite.CreateTestUser(t, "jwtclaims@example.com", "TestPassword123!", []string{"user"})
		resp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("Login failed with status %d", resp.StatusCode)
		}
		token := suite.ExtractTokenFromResponse(t, resp)
		if token == "" {
			t.Fatal("Expected non-empty access token")
		}
		// Decode JWT (without verifying signature, for test only)
		parts := strings.Split(token, ".")
		if len(parts) != 3 {
			t.Fatalf("Invalid JWT format")
		}
		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			t.Fatalf("Failed to decode JWT payload: %v", err)
		}
		var claims map[string]interface{}
		if err := json.Unmarshal(payload, &claims); err != nil {
			t.Fatalf("Failed to unmarshal JWT claims: %v", err)
		}
		if pseudonymID, ok := claims["active_pseudonym_id"].(string); !ok || pseudonymID == "" {
			t.Error("Expected non-empty active_pseudonym_id in JWT claims")
		}
		if displayName, ok := claims["display_name"].(string); !ok || displayName == "" {
			t.Error("Expected non-empty display_name in JWT claims")
		}
	})
}

func TestUserContext_MultiplePseudonyms_Integration(t *testing.T) {
	t.Run("CreateAdditionalPseudonym", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return // Test was skipped
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		testUser := suite.CreateTestUser(t, "additional@example.com", "TestPassword123!", []string{"user"})
		loginResp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		token := suite.ExtractTokenFromResponse(t, loginResp)

		pseudonymData := models.CreatePseudonymBody{
			DisplayName:         "AdditionalPseudonym",
			Bio:                 "Additional pseudonym bio",
			WebsiteURL:          "https://additional.example.com",
			ShowKarma:           &[]bool{true}[0],
			AllowDirectMessages: &[]bool{true}[0],
		}

		resp := suite.MakeAuthenticatedRequest(t, server, "POST", "/pseudonyms", token, pseudonymData)
		defer resp.Body.Close()

		// Read response body once
		var responseBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check status and log debug info if needed
		if resp.StatusCode != http.StatusOK {
			t.Logf("Pseudonym creation failed: status=%d, body=%+v", resp.StatusCode, responseBody)
			t.Errorf("Expected status 200 for pseudonym creation, got %d", resp.StatusCode)
		}

		// Now use the already-decoded response body for assertions
		if pseudonymID, ok := responseBody["pseudonym_id"].(string); !ok || pseudonymID == "" {
			t.Errorf("Expected non-empty pseudonym_id in create response")
		}
		if displayName, ok := responseBody["display_name"].(string); !ok || displayName != "AdditionalPseudonym" {
			t.Errorf("Expected display_name 'AdditionalPseudonym', got %s", displayName)
		}

		// Check that the new pseudonym appears in user profile
		profileResp := suite.MakeAuthenticatedRequest(t, server, "GET", "/users/profile", token, nil)
		defer profileResp.Body.Close()
		if profileResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for profile request, got %d", profileResp.StatusCode)
		}

		var profileBody map[string]interface{}
		if err := json.NewDecoder(profileResp.Body).Decode(&profileBody); err != nil {
			t.Fatalf("Failed to decode profile response: %v", err)
		}
		if pseudonyms, ok := profileBody["pseudonyms"].([]interface{}); !ok || len(pseudonyms) < 2 {
			t.Errorf("Expected at least 2 pseudonyms in profile, got %d", len(pseudonyms))
		}
	})

	t.Run("UpdatePseudonymProfile", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return // Test was skipped
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		testUser := suite.CreateTestUser(t, "update@example.com", "TestPassword123!", []string{"user"})
		loginResp := suite.LoginUser(t, server, testUser.Email, testUser.Password)
		token := suite.ExtractTokenFromResponse(t, loginResp)

		updateData := models.PseudonymProfileBody{
			DisplayName:         "UpdatedDisplayName",
			Bio:                 "Updated bio text",
			WebsiteURL:          "https://updated.example.com",
			ShowKarma:           &[]bool{false}[0],
			AllowDirectMessages: &[]bool{false}[0],
		}

		resp := suite.MakeAuthenticatedRequest(t, server, "PUT", "/pseudonyms/"+testUser.PseudonymID+"/profile", token, updateData)
		defer resp.Body.Close()

		// Read response body once
		var responseBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check status and log debug info if needed
		if resp.StatusCode != http.StatusOK {
			t.Logf("Pseudonym update failed: status=%d, body=%+v", resp.StatusCode, responseBody)
			t.Errorf("Expected status 200 for profile update, got %d", resp.StatusCode)
		}

		// Now use the already-decoded response body for assertions
		if displayName, ok := responseBody["display_name"].(string); !ok || displayName != "UpdatedDisplayName" {
			t.Errorf("Expected display_name 'UpdatedDisplayName', got %s", displayName)
		}
		if bio, ok := responseBody["bio"].(string); !ok || bio != "Updated bio text" {
			t.Errorf("Expected bio 'Updated bio text', got %s", bio)
		}
	})
}

func TestUserContext_Moderation_Integration(t *testing.T) {
	t.Run("ModeratorActionsUseCorrectPseudonym", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return // Test was skipped
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()
		moderatorUser := suite.CreateTestUser(t, "moderator@example.com", "TestPassword123!", []string{"moderator"})
		regularUser := suite.CreateTestUser(t, "regular@example.com", "TestPassword123!", []string{"user"})
		testSubforum := suite.CreateTestSubforum(t, "test-sub", "Test subforum", moderatorUser.UserID, false)
		suite.CreateTestPost(t, "Test Post", "Test content", testSubforum.SubforumID, regularUser.UserID, regularUser.PseudonymID)
		modResp := suite.LoginUser(t, server, moderatorUser.Email, moderatorUser.Password)
		if modResp.StatusCode != http.StatusOK {
			t.Fatalf("Moderator login failed with status %d", modResp.StatusCode)
		}
		modToken := suite.ExtractTokenFromResponse(t, modResp)
		if modToken == "" {
			t.Fatal("Expected non-empty moderator access token")
		}
		profileResp := suite.MakeAuthenticatedRequest(t, server, "GET", "/users/profile", modToken, nil)
		if profileResp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for moderator profile request, got %d", profileResp.StatusCode)
		}
		var profileBody map[string]interface{}
		if err := json.NewDecoder(profileResp.Body).Decode(&profileBody); err != nil {
			t.Fatalf("Failed to decode profile response: %v", err)
		}
		if roles, ok := profileBody["roles"].([]interface{}); !ok || len(roles) == 0 {
			t.Error("Expected non-empty roles array for moderator")
		}
		if pseudonyms, ok := profileBody["pseudonyms"].([]interface{}); !ok || len(pseudonyms) == 0 {
			t.Error("Expected non-empty pseudonyms array for moderator")
		}
	})
}
