package pds

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasswordUpdateIntegration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.Cleanup()

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
			"currentPassword": "testpassword123",
			"newPassword":     "123", // Too short
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

