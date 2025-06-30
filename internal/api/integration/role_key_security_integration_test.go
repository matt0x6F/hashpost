package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/testutil"
)

// TestRoleKeySecurity tests the new role-based security model
func TestRoleKeySecurity(t *testing.T) {
	t.Run("AuthenticationKeyScope", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create a test user and ensure default keys
		testUser := suite.CreateTestUser(t, "auth-test@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, testUser.UserID)

		// Create test user with pseudonyms
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

		// Verify pseudonym ownership using self-correlation key (not authentication key)
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

		// Create a test user and ensure default keys
		testUser := suite.CreateTestUser(t, "self-test@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, testUser.UserID)

		// Create test user with pseudonym
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
		otherPseudonym := suite.CreateTestPseudonym(t, otherUser.UserID, "OtherUser")

		ownsOtherPseudonym, err := suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), otherPseudonym.PseudonymID, testUser.UserID, "user", "self_correlation")
		if err != nil {
			t.Fatalf("Self-correlation key should handle cross-user verification gracefully: %v", err)
		}
		if ownsOtherPseudonym {
			t.Error("Self-correlation key should not verify ownership of other users' pseudonyms")
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

		// Test that different key scopes have different capabilities
		// Authentication key should work for basic operations
		pseudonyms, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUser.UserID, "platform_admin", "authentication")
		if err != nil {
			t.Errorf("Authentication key should work for basic operations: %v", err)
		}

		// Self-correlation key should work for ownership verification
		if len(pseudonyms) > 0 {
			_, err = suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), pseudonyms[0].PseudonymID, testUser.UserID, "platform_admin", "self_correlation")
			if err != nil {
				t.Errorf("Self-correlation key should work for ownership verification: %v", err)
			}
		}

		// Admin correlation key should work for all operations
		_, err = suite.SecurePseudonymDAO.GetPseudonymsByRealIdentity(context.Background(), testUser.Email, "platform_admin", "correlation")
		if err != nil {
			t.Errorf("Admin correlation key should work for all operations: %v", err)
		}
	})

	t.Run("CrossUserAccessPrevention", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create test users
		user1 := suite.CreateTestUser(t, "cross-test1@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, user1.UserID)
		user2 := suite.CreateTestUser(t, "cross-test2@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, user2.UserID)
		suite.CreateTestPseudonym(t, user1.UserID, "CrossUser1")
		suite.CreateTestPseudonym(t, user2.UserID, "CrossUser2")

		// Test that user1 cannot access user2's pseudonyms with authentication key
		// Note: Authentication keys can access any user's pseudonyms, but cannot verify ownership
		user2Pseudonyms, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), user2.UserID, "user", "authentication")
		if err != nil {
			t.Error("Authentication key should be able to access any user's pseudonyms")
		}

		// Test that user1 cannot verify ownership of user2's pseudonyms
		if len(user2Pseudonyms) > 0 {
			ownsOtherPseudonym, err := suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), user2Pseudonyms[0].PseudonymID, user1.UserID, "user", "self_correlation")
			if err == nil && ownsOtherPseudonym {
				t.Error("User should not be able to verify ownership of other users' pseudonyms")
			}
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

		// Test with a real pseudonym ID instead of a fake one
		pseudonyms, err := suite.SecurePseudonymDAO.GetPseudonymsByUserID(context.Background(), testUser.UserID, "user", "authentication")
		if err != nil {
			t.Fatalf("Valid key should work for getting pseudonyms: %v", err)
		}
		if len(pseudonyms) > 0 {
			_, err = suite.SecurePseudonymDAO.VerifyPseudonymOwnership(context.Background(), pseudonyms[0].PseudonymID, testUser.UserID, "invalid_role", "invalid_scope")
			if err == nil {
				t.Error("Invalid role/scope should not work for ownership verification")
			}
		}
	})

	t.Run("RoleKeyDatabaseOperations", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Test role key creation and retrieval
		ctx := context.Background()
		roleKeyDAO := suite.RoleKeyDAO

		// Create a test role key
		capabilities := []string{"test_capability", "another_capability"}
		expiresAt := time.Now().AddDate(0, 1, 0) // Expire in 1 month
		createdBy := suite.CreateTestUser(t, "rolekey-admin@example.com", "password123", []string{"user"}).UserID

		roleKey, err := roleKeyDAO.CreateRoleKey(ctx, "test_role", "test_scope", []byte("test_key_data"), capabilities, expiresAt, createdBy)
		if err != nil {
			t.Fatalf("Failed to create role key: %v", err)
		}

		if roleKey == nil {
			t.Fatal("Created role key should not be nil")
		}

		// Retrieve the role key
		retrievedKey, err := roleKeyDAO.GetRoleKey(ctx, "test_role", "test_scope")
		if err != nil {
			t.Fatalf("Failed to retrieve role key: %v", err)
		}

		if retrievedKey.RoleName != "test_role" {
			t.Errorf("Expected role name 'test_role', got '%s'", retrievedKey.RoleName)
		}

		if retrievedKey.Scope != "test_scope" {
			t.Errorf("Expected scope 'test_scope', got '%s'", retrievedKey.Scope)
		}

		// Test capability validation
		hasCapability, err := roleKeyDAO.ValidateKeyCapability(ctx, "test_role", "test_scope", "test_capability")
		if err != nil {
			t.Fatalf("Failed to validate capability: %v", err)
		}
		if !hasCapability {
			t.Error("Role key should have the test_capability")
		}

		// Test invalid capability
		hasInvalidCapability, err := roleKeyDAO.ValidateKeyCapability(ctx, "test_role", "test_scope", "invalid_capability")
		if err != nil {
			t.Fatalf("Failed to validate invalid capability: %v", err)
		}
		if hasInvalidCapability {
			t.Error("Role key should not have the invalid_capability")
		}

		// Test listing role keys
		roleKeys, err := roleKeyDAO.ListRoleKeys(ctx)
		if err != nil {
			t.Fatalf("Failed to list role keys: %v", err)
		}
		if len(roleKeys) == 0 {
			t.Error("Should have at least one role key")
		}

		// Test deactivating role key
		err = roleKeyDAO.DeactivateRoleKey(ctx, roleKey.KeyID.String())
		if err != nil {
			t.Fatalf("Failed to deactivate role key: %v", err)
		}

		// Verify the key is deactivated
		_, err = roleKeyDAO.GetRoleKey(ctx, "test_role", "test_scope")
		if err == nil {
			t.Error("Deactivated role key should not be retrievable")
		}
	})

	t.Run("DefaultKeysSetup", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create a test user with admin role
		testUser := suite.CreateTestUser(t, "default-keys-test@example.com", "password123", []string{"platform_admin"})

		// Test that default keys are created
		suite.EnsureDefaultKeys(t, testUser.UserID)

		// Verify default keys exist
		defaultKeys := []struct {
			roleName string
			scope    string
		}{
			{"platform_admin", "authentication"},
			{"platform_admin", "self_correlation"},
			{"platform_admin", "correlation"},
		}

		for _, keyDef := range defaultKeys {
			roleKey, err := suite.RoleKeyDAO.GetRoleKey(context.Background(), keyDef.roleName, keyDef.scope)
			if err != nil {
				t.Errorf("Default key for role=%s scope=%s should exist: %v", keyDef.roleName, keyDef.scope, err)
			}
			if roleKey == nil {
				t.Errorf("Default key for role=%s scope=%s should not be nil", keyDef.roleName, keyDef.scope)
			}
		}
	})

	t.Run("LoginWithRoleKeys", func(t *testing.T) {
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

		// Test login with authentication key
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
			t.Errorf("Expected status OK, got %d", resp.StatusCode)
		}

		// Verify that the login response contains the expected data
		var loginResponse map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
			t.Fatalf("Failed to decode login response: %v", err)
		}

		// Check that the response contains expected fields
		if _, ok := loginResponse["user_id"]; !ok {
			t.Error("Login response should contain user_id")
		}
		if _, ok := loginResponse["access_token"]; !ok {
			t.Error("Login response should contain access_token")
		}
		if _, ok := loginResponse["active_pseudonym_id"]; !ok {
			t.Error("Login response should contain active_pseudonym_id")
		}
	})

	t.Run("DebugRoleKeyCapabilities", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()

		// Create a test user and ensure default keys
		testUser := suite.CreateTestUser(t, "debug-capabilities@example.com", "password123", []string{"user"})
		suite.EnsureDefaultKeys(t, testUser.UserID)

		// Debug: Check what capabilities are actually set on the role keys
		ctx := context.Background()

		// Check user authentication key
		authKey, err := suite.RoleKeyDAO.GetRoleKey(ctx, "user", "authentication")
		if err != nil {
			t.Fatalf("Failed to get auth key: %v", err)
		}

		capabilitiesBytes, err := authKey.Capabilities.Value()
		if err != nil {
			t.Fatalf("Failed to get capabilities value: %v", err)
		}

		var capabilities []string
		if err := json.Unmarshal(capabilitiesBytes.([]byte), &capabilities); err != nil {
			t.Fatalf("Failed to unmarshal capabilities: %v", err)
		}

		t.Logf("User authentication key capabilities: %v", capabilities)

		// Check if it has the expected capability
		hasCapability, err := suite.RoleKeyDAO.ValidateKeyCapability(ctx, "user", "authentication", "access_own_pseudonyms")
		if err != nil {
			t.Fatalf("Failed to validate capability: %v", err)
		}

		t.Logf("Has access_own_pseudonyms capability: %v", hasCapability)

		// Check user self-correlation key
		selfKey, err := suite.RoleKeyDAO.GetRoleKey(ctx, "user", "self_correlation")
		if err != nil {
			t.Fatalf("Failed to get self-correlation key: %v", err)
		}

		capabilitiesBytes, err = selfKey.Capabilities.Value()
		if err != nil {
			t.Fatalf("Failed to get capabilities value: %v", err)
		}

		if err := json.Unmarshal(capabilitiesBytes.([]byte), &capabilities); err != nil {
			t.Fatalf("Failed to unmarshal capabilities: %v", err)
		}

		t.Logf("User self-correlation key capabilities: %v", capabilities)

		// Check if it has the expected capability
		hasCapability, err = suite.RoleKeyDAO.ValidateKeyCapability(ctx, "user", "self_correlation", "verify_own_pseudonym_ownership")
		if err != nil {
			t.Fatalf("Failed to validate capability: %v", err)
		}

		t.Logf("Has verify_own_pseudonym_ownership capability: %v", hasCapability)
	})
}
