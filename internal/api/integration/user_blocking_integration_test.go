//go:build integration

package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserBlockingIntegration(t *testing.T) {
	t.Run("PseudonymLevelBlocking", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test users
		user1 := suite.CreateTestUser(t, "user1@example.com", "password123", []string{"user"})
		user2 := suite.CreateTestUser(t, "user2@example.com", "password123", []string{"user"})

		// Create pseudonyms for the users
		pseudonym2 := suite.CreateTestPseudonym(t, user2.UserID, "User2Pseudonym")

		// Login users and get tokens
		loginResp1 := suite.LoginUser(t, server, user1.Email, user1.Password)
		token1 := suite.ExtractTokenFromResponse(t, loginResp1)
		loginResp2 := suite.LoginUser(t, server, user2.Email, user2.Password)
		token2 := suite.ExtractTokenFromResponse(t, loginResp2)

		// Get the active pseudonym from the login response (this is the one that will be used for blocking)
		var loginResponse map[string]interface{}
		err := json.NewDecoder(loginResp1.Body).Decode(&loginResponse)
		require.NoError(t, err, "Failed to decode login response")
		activePseudonymID, ok := loginResponse["active_pseudonym_id"].(string)
		require.True(t, ok, "Expected active_pseudonym_id in login response")
		t.Logf("Active pseudonym ID from login: %s", activePseudonymID)

		// Block pseudonym2 from active pseudonym (pseudonym-level block)
		blockInput := map[string]interface{}{
			"block_all_personas": false,
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", pseudonym2.PseudonymID), token1, blockInput)
		t.Logf("Block response status: %d", resp.StatusCode)
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Block response body: %s", string(body))
		}
		assert.Equal(t, 200, resp.StatusCode, "Expected block request to succeed")

		// Verify the specific pseudonym block was created
		allBlocks, err := suite.UserBlockDAO.GetUserBlocksByBlocker(context.Background(), activePseudonymID)
		if err == nil {
			t.Logf("All blocks for blocker pseudonym %s: %v", activePseudonymID, allBlocks)
		}
		// Get user ID for pseudonym2 to check fingerprint-level blocks
		user2ID, err := suite.SecurePseudonymDAO.GetUserIDByPseudonym(context.Background(), pseudonym2.PseudonymID, "user", "self_correlation")
		require.NoError(t, err, "Failed to get user ID for pseudonym2")

		blocked, err := suite.UserBlockDAO.IsPseudonymBlockedByUser(context.Background(), activePseudonymID, pseudonym2.PseudonymID, user2ID)
		if err != nil {
			t.Logf("Error checking if user is blocked: %v", err)
		}
		require.NoError(t, err, "Failed to check if user is blocked")
		assert.True(t, blocked, "Expected pseudonym2 to be blocked by pseudonym1")

		// Try to create a new pseudonym for user2 (should succeed since it's not fingerprint-level blocking)
		uniqueDisplayName := fmt.Sprintf("User2NewPseudonym_%d", time.Now().UnixNano())
		newPseudonymInput := map[string]interface{}{
			"display_name":          uniqueDisplayName,
			"bio":                   "New bio",
			"website_url":           "",
			"show_karma":            true,
			"allow_direct_messages": true,
		}
		resp = suite.MakeAuthenticatedRequest(t, server, "POST", "/pseudonyms", token2, newPseudonymInput)

		// Log the response body if it's not 200
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Pseudonym creation failed with status %d: %s", resp.StatusCode, string(body))
		}

		assert.Equal(t, 200, resp.StatusCode, "Expected new pseudonym creation to succeed")
	})

	t.Run("FingerprintLevelBlocking", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test users
		user1 := suite.CreateTestUser(t, "user1@example.com", "password123", []string{"user"})
		user2 := suite.CreateTestUser(t, "user2@example.com", "password123", []string{"user"})

		// Create pseudonyms for the users
		pseudonym2 := suite.CreateTestPseudonym(t, user2.UserID, "User2Pseudonym")

		// Login user1
		loginResp1 := suite.LoginUser(t, server, user1.Email, user1.Password)
		token1 := suite.ExtractTokenFromResponse(t, loginResp1)

		// Get the active pseudonym from the login response (this is the one that will be used for blocking)
		var loginResponse map[string]interface{}
		err := json.NewDecoder(loginResp1.Body).Decode(&loginResponse)
		require.NoError(t, err, "Failed to decode login response")
		activePseudonymID, ok := loginResponse["active_pseudonym_id"].(string)
		require.True(t, ok, "Expected active_pseudonym_id in login response")
		t.Logf("Active pseudonym ID from login: %s", activePseudonymID)

		// Block all personas of user2
		blockInput := map[string]interface{}{
			"block_all_personas": true,
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", pseudonym2.PseudonymID), token1, blockInput)
		t.Logf("Fingerprint-level block response status: %d", resp.StatusCode)
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Fingerprint-level block response body: %s", string(body))
		}
		assert.Equal(t, 200, resp.StatusCode, "Expected fingerprint-level block request to succeed")

		// Verify fingerprint-level block was created
		fingerprintBlocks, err := suite.UserBlockDAO.GetFingerprintLevelBlocks(context.Background(), user2.UserID)
		if err != nil {
			t.Logf("Error getting fingerprint-level blocks: %v", err)
		}
		require.NoError(t, err, "Failed to get fingerprint-level blocks")
		t.Logf("Fingerprint-level blocks for user %d: %v", user2.UserID, fingerprintBlocks)
		assert.Greater(t, len(fingerprintBlocks), 0, "Expected fingerprint-level block to exist")

		// DEBUG: Print all blocks for user2 to see what was created
		allBlocks, err := suite.UserBlockDAO.GetUserBlocksByBlockedUser(context.Background(), user2.UserID)
		if err != nil {
			t.Logf("Error getting all blocks for user2: %v", err)
		} else {
			t.Logf("All blocks for user2: %v", allBlocks)
		}

		// Verify that any new pseudonym from user2 would be blocked
		// Use the active pseudonym ID from login for checking the block
		blocked, err := suite.UserBlockDAO.IsUserBlockedAtFingerprintLevel(context.Background(), activePseudonymID, user2.UserID)
		if err != nil {
			t.Logf("Error checking fingerprint-level block: %v", err)
		}
		require.NoError(t, err, "Failed to check fingerprint-level block")
		t.Logf("Is user2 blocked at fingerprint level by %s: %t", activePseudonymID, blocked)
		assert.True(t, blocked, "Expected user2 to be blocked at fingerprint level")
	})

	t.Run("BlockingSelfPrevention", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test user
		user1 := suite.CreateTestUser(t, "user1@example.com", "password123", []string{"user"})
		pseudonym2 := suite.CreateTestPseudonym(t, user1.UserID, "User1Pseudonym")

		// Login user
		loginResp1 := suite.LoginUser(t, server, user1.Email, user1.Password)
		token1 := suite.ExtractTokenFromResponse(t, loginResp1)

		// Try to block self
		blockInput := map[string]interface{}{
			"block_all_personas": true,
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", pseudonym2.PseudonymID), token1, blockInput)
		t.Logf("Self-block response status: %d", resp.StatusCode)
		if resp.StatusCode != 400 {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Self-block response body: %s", string(body))
		}
		assert.Equal(t, 400, resp.StatusCode, "Expected self-blocking to fail with 400")
	})

	t.Run("UnblockingPseudonym", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test users
		user1 := suite.CreateTestUser(t, "user1@example.com", "password123", []string{"user"})
		user2 := suite.CreateTestUser(t, "user2@example.com", "password123", []string{"user"})

		// Create pseudonyms
		pseudonym2 := suite.CreateTestPseudonym(t, user2.UserID, "User2Pseudonym")

		// Login user1
		loginResp1 := suite.LoginUser(t, server, user1.Email, user1.Password)
		token1 := suite.ExtractTokenFromResponse(t, loginResp1)

		// Get the active pseudonym from the login response (this is the one that will be used for blocking/unblocking)
		var loginResponse map[string]interface{}
		err := json.NewDecoder(loginResp1.Body).Decode(&loginResponse)
		require.NoError(t, err, "Failed to decode login response")
		activePseudonymID, ok := loginResponse["active_pseudonym_id"].(string)
		require.True(t, ok, "Expected active_pseudonym_id in login response")
		t.Logf("Active pseudonym ID from login: %s", activePseudonymID)

		// Block pseudonym2 (pseudonym-level block)
		blockInput := map[string]interface{}{
			"block_all_personas": false,
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", pseudonym2.PseudonymID), token1, blockInput)
		t.Logf("Block response status: %d", resp.StatusCode)
		assert.Equal(t, 200, resp.StatusCode, "Expected block request to succeed")

		// Verify block exists
		// Get user ID for pseudonym2 to check fingerprint-level blocks
		user2ID, err := suite.SecurePseudonymDAO.GetUserIDByPseudonym(context.Background(), pseudonym2.PseudonymID, "user", "self_correlation")
		require.NoError(t, err, "Failed to get user ID for pseudonym2")

		blocked, err := suite.UserBlockDAO.IsPseudonymBlockedByUser(context.Background(), activePseudonymID, pseudonym2.PseudonymID, user2ID)
		if err != nil {
			t.Logf("Error checking if user is blocked: %v", err)
		}
		require.NoError(t, err, "Failed to check if user is blocked")
		assert.True(t, blocked, "Expected pseudonym2 to be blocked")

		// Unblock pseudonym2
		resp = suite.MakeAuthenticatedRequest(t, server, "DELETE", fmt.Sprintf("/users/%s/block", pseudonym2.PseudonymID), token1, nil)
		t.Logf("Unblock response status: %d", resp.StatusCode)
		if resp.StatusCode != 200 {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Unblock response body: %s", string(body))
		}
		assert.Equal(t, 200, resp.StatusCode, "Expected unblock request to succeed")

		// Verify block is removed
		blocked, err = suite.UserBlockDAO.IsPseudonymBlockedByUser(context.Background(), activePseudonymID, pseudonym2.PseudonymID, user2ID)
		if err != nil {
			t.Logf("Error checking if user is still blocked: %v", err)
		}
		require.NoError(t, err, "Failed to check if user is still blocked")
		assert.False(t, blocked, "Expected pseudonym2 to no longer be blocked")
	})

	t.Run("MultiplePseudonymsBlocking", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test users
		user1 := suite.CreateTestUser(t, "user1@example.com", "password123", []string{"user"})
		user2 := suite.CreateTestUser(t, "user2@example.com", "password123", []string{"user"})

		// Create multiple pseudonyms for user2
		pseudonym2a := suite.CreateTestPseudonym(t, user2.UserID, "User2PseudonymA")
		pseudonym2b := suite.CreateTestPseudonym(t, user2.UserID, "User2PseudonymB")

		// Login user1
		loginResp1 := suite.LoginUser(t, server, user1.Email, user1.Password)
		token1 := suite.ExtractTokenFromResponse(t, loginResp1)

		// Get the active pseudonym from the login response (this is the one that will be used for blocking)
		var loginResponse map[string]interface{}
		err := json.NewDecoder(loginResp1.Body).Decode(&loginResponse)
		require.NoError(t, err, "Failed to decode login response")

		activePseudonymID, ok := loginResponse["active_pseudonym_id"].(string)
		require.True(t, ok, "Expected active_pseudonym_id in login response")
		t.Logf("Active pseudonym ID from login: %s", activePseudonymID)

		// Block all personas of user2
		blockInput := map[string]interface{}{
			"block_all_personas": true,
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", pseudonym2a.PseudonymID), token1, blockInput)
		t.Logf("Fingerprint-level block response status: %d", resp.StatusCode)
		assert.Equal(t, 200, resp.StatusCode, "Expected fingerprint-level block request to succeed")

		// Verify both pseudonyms are blocked
		// Get user ID for pseudonym2a to check fingerprint-level blocks
		user2aID, err := suite.SecurePseudonymDAO.GetUserIDByPseudonym(context.Background(), pseudonym2a.PseudonymID, "user", "self_correlation")
		require.NoError(t, err, "Failed to get user ID for pseudonym2a")

		// Get user ID for pseudonym2b to check fingerprint-level blocks
		user2bID, err := suite.SecurePseudonymDAO.GetUserIDByPseudonym(context.Background(), pseudonym2b.PseudonymID, "user", "self_correlation")
		require.NoError(t, err, "Failed to get user ID for pseudonym2b")

		// DEBUG: Print user IDs for both pseudonyms
		t.Logf("DEBUG: pseudonym2a.PseudonymID=%s, user2aID=%d", pseudonym2a.PseudonymID, user2aID)
		t.Logf("DEBUG: pseudonym2b.PseudonymID=%s, user2bID=%d", pseudonym2b.PseudonymID, user2bID)
		t.Logf("DEBUG: user2aID == user2bID: %t", user2aID == user2bID)

		// Use the active pseudonym ID from login for checking blocks
		blockedA, err := suite.UserBlockDAO.IsPseudonymBlockedByUser(context.Background(), activePseudonymID, pseudonym2a.PseudonymID, user2aID)
		if err != nil {
			t.Logf("Error checking if pseudonym2a is blocked: %v", err)
		}
		require.NoError(t, err, "Failed to check if pseudonym2a is blocked")
		assert.True(t, blockedA, "Expected pseudonym2a to be blocked")

		blockedB, err := suite.UserBlockDAO.IsPseudonymBlockedByUser(context.Background(), activePseudonymID, pseudonym2b.PseudonymID, user2bID)
		if err != nil {
			t.Logf("Error checking if pseudonym2b is blocked: %v", err)
		}
		require.NoError(t, err, "Failed to check if pseudonym2b is blocked")
		assert.True(t, blockedB, "Expected pseudonym2b to be blocked")

		// DEBUG: Print all fingerprint-level blocks for user2
		blocks, err := suite.UserBlockDAO.GetFingerprintLevelBlocks(context.Background(), user2aID)
		require.NoError(t, err, "Failed to get fingerprint-level blocks for user2")
		t.Logf("Fingerprint-level blocks for user2 (user_id=%d): count=%d", user2aID, len(blocks))
		for i, block := range blocks {
			t.Logf("  Block %d: block_id=%d blocker_pseudonym_id=%s blocked_user_id=%v blocked_pseudonym_id=%v", i, block.BlockID, block.BlockerPseudonymID, block.BlockedUserID, block.BlockedPseudonymID)
		}
	})

	t.Run("BlockingNonExistentPseudonym", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test user
		user1 := suite.CreateTestUser(t, "user1@example.com", "password123", []string{"user"})

		// Login user1
		loginResp1 := suite.LoginUser(t, server, user1.Email, user1.Password)
		token1 := suite.ExtractTokenFromResponse(t, loginResp1)

		// Try to block nonexistent pseudonym
		blockInput := map[string]interface{}{
			"block_all_personas": true,
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", "nonexistent-pseudonym-id"), token1, blockInput)
		t.Logf("Block nonexistent pseudonym response status: %d", resp.StatusCode)
		if resp.StatusCode != 404 {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Block nonexistent pseudonym response body: %s", string(body))
		}
		assert.Equal(t, 404, resp.StatusCode, "Expected blocking nonexistent pseudonym to fail with 404")
	})

	t.Run("BlockingWithoutAuthentication", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test user
		user2 := suite.CreateTestUser(t, "user2@example.com", "password123", []string{"user"})
		suite.CreateTestPseudonym(t, user2.UserID, "User2Pseudonym")

		// Try to block without authentication
		blockInput := map[string]interface{}{
			"block_all_personas": true,
		}
		resp := suite.MakeRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", "some-pseudonym-id"), blockInput)
		t.Logf("Unauthenticated block response status: %d", resp.StatusCode)
		if resp.StatusCode != 401 {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Unauthenticated block response body: %s", string(body))
		}
		assert.Equal(t, 401, resp.StatusCode, "Expected unauthenticated block request to fail with 401")
	})

	t.Run("BlockSelfError", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test user
		user1 := suite.CreateTestUser(t, "user1@example.com", "password123", []string{"user"})
		pseudonym2 := suite.CreateTestPseudonym(t, user1.UserID, "User1Pseudonym")

		// Login user
		loginResp1 := suite.LoginUser(t, server, user1.Email, user1.Password)
		token1 := suite.ExtractTokenFromResponse(t, loginResp1)

		// Try to block self with fingerprint-level blocking
		blockInput := map[string]interface{}{
			"block_all_personas": true,
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", pseudonym2.PseudonymID), token1, blockInput)
		t.Logf("Self fingerprint-level block response status: %d", resp.StatusCode)
		if resp.StatusCode != 400 {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Self fingerprint-level block response body: %s", string(body))
		}
		assert.Equal(t, 400, resp.StatusCode, "Expected self fingerprint-level blocking to fail with 400")
	})

	t.Run("BlockNonexistentPseudonym", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test user
		user1 := suite.CreateTestUser(t, "user1@example.com", "password123", []string{"user"})

		// Login user1
		loginResp1 := suite.LoginUser(t, server, user1.Email, user1.Password)
		token1 := suite.ExtractTokenFromResponse(t, loginResp1)

		// Try to block nonexistent pseudonym
		blockInput := map[string]interface{}{
			"block_all_personas": true,
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", "nonexistent-pseudonym-id"), token1, blockInput)
		t.Logf("Block nonexistent pseudonym response status: %d", resp.StatusCode)
		if resp.StatusCode != 404 {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Block nonexistent pseudonym response body: %s", string(body))
		}
		assert.Equal(t, 404, resp.StatusCode, "Expected blocking nonexistent pseudonym to fail with 404")
	})

	t.Run("BlockMultiplePseudonyms", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test users
		user1 := suite.CreateTestUser(t, "user1@example.com", "password123", []string{"user"})
		user2 := suite.CreateTestUser(t, "user2@example.com", "password123", []string{"user"})

		// Create multiple pseudonyms for user2
		pseudonym2a := suite.CreateTestPseudonym(t, user2.UserID, "User2PseudonymA")
		pseudonym2b := suite.CreateTestPseudonym(t, user2.UserID, "User2PseudonymB")

		// Login user1
		loginResp1 := suite.LoginUser(t, server, user1.Email, user1.Password)
		token1 := suite.ExtractTokenFromResponse(t, loginResp1)

		// Get the active pseudonym from the login response (this is the one that will be used for blocking)
		var loginResponse map[string]interface{}
		err := json.NewDecoder(loginResp1.Body).Decode(&loginResponse)
		require.NoError(t, err, "Failed to decode login response")

		activePseudonymID, ok := loginResponse["active_pseudonym_id"].(string)
		require.True(t, ok, "Expected active_pseudonym_id in login response")
		t.Logf("Active pseudonym ID from login: %s", activePseudonymID)

		// Block pseudonym2a
		blockInput := map[string]interface{}{
			"block_all_personas": true,
		}
		resp := suite.MakeAuthenticatedRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", pseudonym2a.PseudonymID), token1, blockInput)
		t.Logf("Block pseudonym2a response status: %d", resp.StatusCode)
		assert.Equal(t, 200, resp.StatusCode, "Expected block request to succeed")

		// Block pseudonym2b
		blockInput = map[string]interface{}{
			"block_all_personas": true,
		}
		resp = suite.MakeAuthenticatedRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", pseudonym2b.PseudonymID), token1, blockInput)
		t.Logf("Block pseudonym2b response status: %d", resp.StatusCode)
		assert.Equal(t, 200, resp.StatusCode, "Expected second block request to succeed")

		// Verify both pseudonyms are blocked
		// Get user ID for pseudonym2a to check fingerprint-level blocks
		user2aID, err := suite.SecurePseudonymDAO.GetUserIDByPseudonym(context.Background(), pseudonym2a.PseudonymID, "user", "self_correlation")
		require.NoError(t, err, "Failed to get user ID for pseudonym2a")

		// Get user ID for pseudonym2b to check fingerprint-level blocks
		user2bID, err := suite.SecurePseudonymDAO.GetUserIDByPseudonym(context.Background(), pseudonym2b.PseudonymID, "user", "self_correlation")
		require.NoError(t, err, "Failed to get user ID for pseudonym2b")

		// DEBUG: Print user IDs for both pseudonyms
		t.Logf("DEBUG: pseudonym2a.PseudonymID=%s, user2aID=%d", pseudonym2a.PseudonymID, user2aID)
		t.Logf("DEBUG: pseudonym2b.PseudonymID=%s, user2bID=%d", pseudonym2b.PseudonymID, user2bID)
		t.Logf("DEBUG: user2aID == user2bID: %t", user2aID == user2bID)

		// Use the active pseudonym ID from login for checking blocks
		blockedA, err := suite.UserBlockDAO.IsPseudonymBlockedByUser(context.Background(), activePseudonymID, pseudonym2a.PseudonymID, user2aID)
		if err != nil {
			t.Logf("Error checking if pseudonym2a is blocked: %v", err)
		}
		require.NoError(t, err, "Failed to check if pseudonym2a is blocked")
		assert.True(t, blockedA, "Expected pseudonym2a to be blocked")

		blockedB, err := suite.UserBlockDAO.IsPseudonymBlockedByUser(context.Background(), activePseudonymID, pseudonym2b.PseudonymID, user2bID)
		if err != nil {
			t.Logf("Error checking if pseudonym2b is blocked: %v", err)
		}
		require.NoError(t, err, "Failed to check if pseudonym2b is blocked")
		assert.True(t, blockedB, "Expected pseudonym2b to be blocked")
	})

	t.Run("UnauthenticatedBlock", func(t *testing.T) {
		suite := testutil.NewIntegrationTestSuite(t)
		if suite == nil {
			return
		}
		defer suite.Cleanup()
		server := suite.CreateTestServer()
		defer server.Close()

		// Create test user
		user1 := suite.CreateTestUser(t, "user1@example.com", "password123", []string{"user"})
		suite.CreateTestPseudonym(t, user1.UserID, "User1Pseudonym")

		// Try to block without authentication
		blockInput := map[string]interface{}{
			"block_all_personas": true,
		}
		resp := suite.MakeRequest(t, server, "POST", fmt.Sprintf("/users/%s/block", "some-pseudonym-id"), blockInput)
		t.Logf("Unauthenticated block response status: %d", resp.StatusCode)
		if resp.StatusCode != 401 {
			body, _ := io.ReadAll(resp.Body)
			t.Logf("Unauthenticated block response body: %s", string(body))
		}
		assert.Equal(t, 401, resp.StatusCode, "Expected unauthenticated block request to fail with 401")
	})
}
