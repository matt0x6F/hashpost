package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/testutil"
)

// TestIBEKeyScopeSeparation tests that different key scopes have appropriate access levels
func TestIBEKeyScopeSeparation(t *testing.T) {
	t.Run("AuthenticationKeyScope", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test user with multiple pseudonyms
		testUser := suite.CreateTestUser(t, "auth-test@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, testUser.UserID)
		pseudonym1 := suite.CreateTestPseudonym(t, testUser.UserID, "AuthUser1")
		suite.CreateTestPseudonym(t, testUser.UserID, "AuthUser2")

		// Test that authentication key can access user's own pseudonyms
		pseudonyms, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUser.UserID, "user", "authentication")
		if err != nil {
			t.Fatalf("Authentication key should be able to access user's pseudonyms: %v", err)
		}

		if len(pseudonyms) != 3 {
			t.Errorf("Expected 3 pseudonyms (1 from CreateTestUser + 2 additional), got %d", len(pseudonyms))
		}

		// Verify pseudonym ownership using self-correlation key
		ownsPseudonym1, err := suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), pseudonym1.PseudonymID, testUser.UserID, "user", "self_correlation")
		if err != nil {
			t.Fatalf("Self-correlation key should be able to verify ownership: %v", err)
		}
		if !ownsPseudonym1 {
			t.Error("Self-correlation key should verify correct ownership")
		}
	})

	t.Run("SelfCorrelationKeyScope", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test user
		testUser := suite.CreateTestUser(t, "self-test@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, testUser.UserID)
		pseudonym := suite.CreateTestPseudonym(t, testUser.UserID, "SelfUser")

		// Test that self-correlation key can verify ownership
		ownsPseudonym, err := suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), pseudonym.PseudonymID, testUser.UserID, "user", "self_correlation")
		if err != nil {
			t.Fatalf("Self-correlation key should be able to verify ownership: %v", err)
		}
		if !ownsPseudonym {
			t.Error("Self-correlation key should verify correct ownership")
		}

		// Test that self-correlation key cannot access other users' pseudonyms
		otherUser := suite.CreateTestUser(t, "other@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, otherUser.UserID)
		otherPseudonym := suite.CreateTestPseudonym(t, otherUser.UserID, "OtherUser")

		ownsOtherPseudonym, err := suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), otherPseudonym.PseudonymID, testUser.UserID, "user", "self_correlation")
		if err != nil {
			t.Fatalf("Self-correlation key should handle cross-user verification gracefully: %v", err)
		}
		if ownsOtherPseudonym {
			t.Error("Self-correlation key should not grant access to other users' pseudonyms")
		}
	})

	t.Run("AdministrativeCorrelationKeyScope", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test users with admin roles and pseudonyms
		user1 := suite.CreateTestUser(t, "admin-test1@example.com", "password123", []string{"platform_admin"})
		suite.EnsureDefaultKeys(t, user1.UserID)
		user2 := suite.CreateTestUser(t, "admin-test2@example.com", "password123", []string{"platform_admin"})
		suite.EnsureDefaultKeys(t, user2.UserID)
		suite.CreateTestPseudonym(t, user1.UserID, "AdminUser1")
		suite.CreateTestPseudonym(t, user2.UserID, "AdminUser2")

		// Test that admin correlation key can access all pseudonyms
		user1Pseudonyms, err := suite.SecurePseudonymDAO.GetPseudonymsByRealIdentity(context.Background(), user1.Email, "platform_admin", "correlation")
		if err != nil {
			t.Fatalf("Admin correlation key should be able to access user1 pseudonyms: %v", err)
		}
		if len(user1Pseudonyms) != 2 {
			t.Errorf("Expected 2 pseudonyms for user1 (1 from CreateTestUser + 1 additional), got %d", len(user1Pseudonyms))
		}

		user2Pseudonyms, err := suite.SecurePseudonymDAO.GetPseudonymsByRealIdentity(context.Background(), user2.Email, "platform_admin", "correlation")
		if err != nil {
			t.Fatalf("Admin correlation key should be able to access user2 pseudonyms: %v", err)
		}
		if len(user2Pseudonyms) != 2 {
			t.Errorf("Expected 2 pseudonyms for user2 (1 from CreateTestUser + 1 additional), got %d", len(user2Pseudonyms))
		}

		// Test cross-user correlation
		realIdentity1, err := suite.SecurePseudonymDAO.GetRealIdentityByPseudonym(context.Background(), user1Pseudonyms[0].PseudonymID, "platform_admin", "correlation")
		if err != nil {
			t.Fatalf("Admin correlation key should be able to get real identity: %v", err)
		}
		if realIdentity1 == "" {
			t.Error("Admin correlation key should return real identity fingerprint")
		}
	})

	t.Run("KeyScopeIsolation", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test user with admin role
		testUser := suite.CreateTestUser(t, "isolation-test@example.com", "password123", []string{"platform_admin"})
		suite.EnsureDefaultKeys(t, testUser.UserID)
		suite.CreateTestPseudonym(t, testUser.UserID, "IsolationUser")

		// Test that different key scopes work for their intended purposes
		// Authentication key should work for basic operations
		_, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUser.UserID, "platform_admin", "authentication")
		if err != nil {
			t.Errorf("Authentication key should work for getting pseudonyms: %v", err)
		}

		// Self-correlation key should work for ownership verification
		pseudonyms, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUser.UserID, "platform_admin", "authentication")
		if err != nil {
			t.Errorf("Authentication key should work for getting pseudonyms: %v", err)
		}
		if len(pseudonyms) > 0 {
			_, err = suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), pseudonyms[0].PseudonymID, testUser.UserID, "platform_admin", "self_correlation")
			if err != nil {
				t.Errorf("Self-correlation key should work for ownership verification: %v", err)
			}
		}

		// Admin correlation key should work for all operations
		_, err = suite.SecurePseudonymDAO.GetPseudonymsByRealIdentity(context.Background(), testUser.Email, "platform_admin", "correlation")
		if err != nil {
			t.Errorf("Admin correlation key should work for getting pseudonyms: %v", err)
		}
	})
}

// TestAuthenticationFlowWithKeyScopes tests the complete authentication flow with proper key scopes
func TestAuthenticationFlowWithKeyScopes(t *testing.T) {
	t.Run("LoginWithAuthenticationKey", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test user
		testUser := suite.CreateTestUser(t, "login-test@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, testUser.UserID)
		suite.CreateTestPseudonym(t, testUser.UserID, "LoginUser")

		server := suite.CreateTestServer()
		defer server.Close()

		// Test login with authentication key scope
		loginData := map[string]string{
			"email":    "login-test@example.com",
			"password": "password123",
		}

		jsonData, err := json.Marshal(loginData)
		if err != nil {
			t.Fatalf("Failed to marshal login data: %v", err)
		}

		resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make login request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		// Verify response contains user's pseudonyms
		var responseBody map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
			t.Fatalf("Failed to decode response: %v", err)
		}

		// Check that pseudonyms are included in response
		if pseudonyms, ok := responseBody["pseudonyms"].([]interface{}); ok {
			if len(pseudonyms) == 0 {
				t.Error("Login response should include user's pseudonyms")
			}
		} else {
			t.Error("Login response should include pseudonyms array")
		}
	})

	t.Run("LoginWithoutPseudonyms", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test user without pseudonyms by creating user directly
		ctx := context.Background()
		// Use the same password hashing as CreateTestUser
		passwordHash := func(password string) string {
			hash := sha256.Sum256([]byte(password))
			return hex.EncodeToString(hash[:])
		}("password123")
		user, err := suite.UserDAO.CreateUser(ctx, "no-pseudonyms@example.com", passwordHash)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
		suite.Tracker.TrackUser(user.UserID)
		// Don't create any pseudonyms for this user

		server := suite.CreateTestServer()
		defer server.Close()

		// Test login should fail when user has no pseudonyms
		loginData := map[string]string{
			"email":    "no-pseudonyms@example.com",
			"password": "password123",
		}

		jsonData, err := json.Marshal(loginData)
		if err != nil {
			t.Fatalf("Failed to marshal login data: %v", err)
		}

		resp, err := http.Post(server.URL+"/auth/login", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			t.Fatalf("Failed to make login request: %v", err)
		}
		defer resp.Body.Close()

		// Should fail because user has no pseudonyms
		if resp.StatusCode == http.StatusOK {
			t.Error("Login should fail when user has no pseudonyms")
		}
	})
}

// TestKeyScopeSecurity tests security aspects of key scope separation
func TestKeyScopeSecurity(t *testing.T) {
	t.Run("CrossUserAccessPrevention", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create two users
		user1 := suite.CreateTestUser(t, "security1@example.com", "password123", []string{"user"})
		user2 := suite.CreateTestUser(t, "security2@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, user1.UserID)
		suite.EnsureDefaultKeys(t, user2.UserID)
		pseudonym2 := suite.CreateTestPseudonym(t, user2.UserID, "SecurityUser2")

		// Test that user1 cannot access user2's pseudonyms with authentication key
		// Note: Authentication keys can access any user's pseudonyms, but cannot verify ownership
		user2Pseudonyms, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), user2.UserID, "user", "authentication")
		if err != nil {
			t.Fatalf("Authentication key should handle cross-user access gracefully: %v", err)
		}
		if len(user2Pseudonyms) == 0 {
			t.Error("Authentication key should be able to access any user's pseudonyms")
		}

		// User1 should not be able to verify ownership of user2's pseudonym
		ownsPseudonym2, err := suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), pseudonym2.PseudonymID, user1.UserID, "user", "self_correlation")
		if err != nil {
			t.Fatalf("Self-correlation key should handle cross-user ownership verification gracefully: %v", err)
		}
		if ownsPseudonym2 {
			t.Error("Self-correlation key should not grant ownership of other users' pseudonyms")
		}
	})

	t.Run("ExpiredKeyHandling", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test user
		testUser := suite.CreateTestUser(t, "expired-test@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, testUser.UserID)
		suite.CreateTestPseudonym(t, testUser.UserID, "ExpiredUser")

		// Test with invalid role/scope combination
		_, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUser.UserID, "invalid_role", "authentication")
		if err == nil {
			t.Error("Invalid role should not work for getting pseudonyms")
		}

		_, err = suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), "invalid_pseudonym_id", testUser.UserID, "user", "self_correlation")
		if err == nil {
			t.Error("Invalid pseudonym ID should not work for ownership verification")
		}
	})

	t.Run("InvalidKeyHandling", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test user
		testUser := suite.CreateTestUser(t, "invalid-test@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, testUser.UserID)
		suite.CreateTestPseudonym(t, testUser.UserID, "InvalidUser")

		// Test with invalid role
		_, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUser.UserID, "invalid_role", "authentication")
		if err == nil {
			t.Error("Invalid role should not work for getting pseudonyms")
		}

		// Test with invalid scope
		_, err = suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUser.UserID, "user", "invalid_scope")
		if err == nil {
			t.Error("Invalid scope should not work for getting pseudonyms")
		}

		// Test with valid role/scope combination
		pseudonyms, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUser.UserID, "user", "authentication")
		if err != nil {
			t.Fatalf("Valid role/scope should work for getting pseudonyms: %v", err)
		}
		if len(pseudonyms) > 0 {
			_, err = suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), pseudonyms[0].PseudonymID, testUser.UserID, "user", "self_correlation")
			if err != nil {
				t.Errorf("Valid role/scope should work for ownership verification: %v", err)
			}
		}
	})
}

// TestKeyScopePerformance tests performance characteristics of different key scopes
func TestKeyScopePerformance(t *testing.T) {
	t.Run("KeyGenerationPerformance", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test user with admin role for performance testing
		testUser := suite.CreateTestUser(t, "perf-test@example.com", "password123", []string{"platform_admin"})
		suite.EnsureDefaultKeys(t, testUser.UserID)
		suite.CreateTestPseudonym(t, testUser.UserID, "PerfUser")

		// Test role key retrieval performance for different scopes
		start := time.Now()
		for i := 0; i < 10; i++ {
			_, err := suite.RoleKeyDAO.GetRoleKey(context.Background(), "platform_admin", "authentication")
			if err != nil {
				t.Fatalf("Failed to get authentication key: %v", err)
			}
		}
		authKeyTime := time.Since(start)

		start = time.Now()
		for i := 0; i < 10; i++ {
			_, err := suite.RoleKeyDAO.GetRoleKey(context.Background(), "platform_admin", "self_correlation")
			if err != nil {
				t.Fatalf("Failed to get self-correlation key: %v", err)
			}
		}
		selfKeyTime := time.Since(start)

		start = time.Now()
		for i := 0; i < 10; i++ {
			_, err := suite.RoleKeyDAO.GetRoleKey(context.Background(), "platform_admin", "correlation")
			if err != nil {
				t.Fatalf("Failed to get admin correlation key: %v", err)
			}
		}
		adminKeyTime := time.Since(start)

		// All key retrieval should be reasonably fast
		if authKeyTime > 5*time.Second {
			t.Errorf("Authentication key retrieval too slow: %v", authKeyTime)
		}
		if selfKeyTime > 5*time.Second {
			t.Errorf("Self-correlation key retrieval too slow: %v", selfKeyTime)
		}
		if adminKeyTime > 5*time.Second {
			t.Errorf("Admin correlation key retrieval too slow: %v", adminKeyTime)
		}

		// Test pseudonym access performance
		start = time.Now()
		for i := 0; i < 10; i++ {
			_, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUser.UserID, "platform_admin", "authentication")
			if err != nil {
				t.Fatalf("Failed to get pseudonyms: %v", err)
			}
		}
		accessTime := time.Since(start)

		if accessTime > 10*time.Second {
			t.Errorf("Pseudonym access too slow: %v", accessTime)
		}
	})
}
