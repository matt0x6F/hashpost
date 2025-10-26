package pds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordUpdateIntegration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.Cleanup()

	// Create test user with password hash
	timestamp := time.Now().UnixNano()
	handle := fmt.Sprintf("testuser%d.hashpost.local", timestamp)
	did := fmt.Sprintf("did:plc:testuser%d", timestamp)

	// Hash the test password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpassword123"), bcrypt.DefaultCost)
	require.NoError(t, err)
	passwordHash := string(hashedPassword)

	// Create user with password
	_, err = suite.db.CreateUserWithPassword(context.Background(), &generated.CreateUserWithPasswordParams{
		Handle:       handle,
		Did:          did,
		Email:        nil,
		PasswordHash: &passwordHash,
	})
	require.NoError(t, err)

	// Set user info in suite
	suite.userDID = did
	suite.userHandle = handle

	// Create session manually
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	_, err = suite.db.CreateUserSession(context.Background(), &generated.CreateUserSessionParams{
		SessionID: sessionID,
		UserDid:   did,
		Handle:    handle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Generate auth token
	session := &Session{
		ID:        sessionID,
		DID:       did,
		Handle:    handle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	authToken, _, err := suite.server.authService.GenerateTokens(session)
	require.NoError(t, err)
	suite.authToken = authToken

	t.Run("successful_password_update", func(t *testing.T) {
		// Create request to update password
		requestBody := map[string]interface{}{
			"currentPassword": "testpassword123",
			"newPassword":     "newpassword123",
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.updatePassword", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid_current_password", func(t *testing.T) {
		// Create request with wrong current password
		requestBody := map[string]interface{}{
			"currentPassword": "wrongpassword",
			"newPassword":     "newpassword123",
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.updatePassword", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("weak_new_password", func(t *testing.T) {
		// Create request with weak new password
		requestBody := map[string]interface{}{
			"currentPassword": "newpassword123", // Use the updated password from previous test
			"newPassword":     "123",            // Too short
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.updatePassword", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing_authorization", func(t *testing.T) {
		// Create request without authorization header
		requestBody := map[string]interface{}{
			"currentPassword": "testpassword123",
			"newPassword":     "newpassword123",
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.updatePassword", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
