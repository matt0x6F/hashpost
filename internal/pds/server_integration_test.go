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

	"github.com/matt0x6f/hashpost/internal/config"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IntegrationTestSuite provides a complete test environment for PDS integration tests
type IntegrationTestSuite struct {
	server     *Server
	db         *generated.Queries
	cleanup    func()
	baseURL    string
	client     *http.Client
	userDID    string
	userHandle string
	authToken  string
}

// SetupIntegrationTestSuite creates a complete test environment
func SetupIntegrationTestSuite(t *testing.T) *IntegrationTestSuite {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Setup test database using pgtestdb
	pool := testutil.SetupPDSTestDB(t)

	// Create test config
	cfg := &config.Config{
		PDS: config.PDSConfig{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8080,
				Dev:  true,
			},
			Database: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "hashpost_test",
				Username: "hashpost",
				Password: "password",
				SSLMode:  "disable",
			},
			Redis: config.RedisConfig{
				URL: "redis://localhost:6379",
			},
			Atproto: config.AtprotoConfig{
				DIDResolver: "https://plc.directory",
				HandleBase:  "hashpost.local",
			},
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "json",
		},
	}

	// Create database queries
	db := generated.New(pool)

	// Create server
	server, err := NewServer(cfg, db)
	require.NoError(t, err)

	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	suite := &IntegrationTestSuite{
		server:  server,
		db:      db,
		cleanup: nil, // pgtestdb handles cleanup automatically
		baseURL: "http://localhost:8080",
		client:  client,
	}

	return suite
}

// Cleanup cleans up the test environment
func (s *IntegrationTestSuite) Cleanup() {
	// pgtestdb handles cleanup automatically, no manual cleanup needed
}

// CreateTestUser creates a test user and returns DID and handle
func (s *IntegrationTestSuite) CreateTestUser(t *testing.T) (string, string) {
	// Use simple timestamp for uniqueness
	timestamp := time.Now().UnixNano()
	handle := fmt.Sprintf("testuser%d.hashpost.local", timestamp)
	did := fmt.Sprintf("did:plc:testuser%d", timestamp)

	// Create user in database
	_, err := s.db.CreateUser(context.Background(), &generated.CreateUserParams{
		Did:    did,
		Handle: handle,
	})
	require.NoError(t, err)

	s.userDID = did
	s.userHandle = handle

	return did, handle
}

// CreateTestSession creates a test session and returns auth token
func (s *IntegrationTestSuite) CreateTestSession(t *testing.T) string {
	// Create test user first
	s.CreateTestUser(t)

	// Create session
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())
	_, err := s.db.CreateUserSession(context.Background(), &generated.CreateUserSessionParams{
		SessionID: sessionID,
		UserDid:   s.userDID,
		Handle:    s.userHandle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Generate real JWT token using auth service
	session := &Session{
		ID:        sessionID,
		DID:       s.userDID,
		Handle:    s.userHandle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	authToken, _, err := s.server.authService.GenerateTokens(session)
	require.NoError(t, err)
	s.authToken = authToken

	return authToken
}

// TestAtprotoServerCreateSession tests the createSession endpoint
func TestAtprotoServerCreateSession_Integration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.Cleanup()

	t.Run("successful_authentication", func(t *testing.T) {
		// Create test user with password
		did, handle := suite.CreateTestUser(t)

		// Get user from database
		user, err := suite.db.GetUserByDID(context.Background(), did)
		require.NoError(t, err)

		// Hash password using auth service
		hashedPassword, err := suite.server.authService.HashPassword("testpassword123")
		require.NoError(t, err)

		// Update user with hashed password
		err = suite.db.UpdateUserPasswordHash(context.Background(), &generated.UpdateUserPasswordHashParams{
			ID:           user.ID,
			PasswordHash: &hashedPassword,
		})
		require.NoError(t, err)

		// Create request
		requestBody := map[string]interface{}{
			"identifier": handle,
			"password":   "testpassword123",
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.createSession", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "accessJwt")
		assert.Contains(t, response, "refreshJwt")
		assert.Contains(t, response, "handle")
		assert.Contains(t, response, "did")
		assert.Equal(t, handle, response["handle"])
		assert.Equal(t, did, response["did"])
	})

	t.Run("invalid_credentials", func(t *testing.T) {
		// Create test user
		did, handle := suite.CreateTestUser(t)

		// Get user from database and set password hash
		user, err := suite.db.GetUserByDID(context.Background(), did)
		require.NoError(t, err)

		// Hash password using auth service
		hashedPassword, err := suite.server.authService.HashPassword("correctpassword123")
		require.NoError(t, err)

		// Update user with hashed password
		err = suite.db.UpdateUserPasswordHash(context.Background(), &generated.UpdateUserPasswordHashParams{
			ID:           user.ID,
			PasswordHash: &hashedPassword,
		})
		require.NoError(t, err)

		// Create request with wrong password
		requestBody := map[string]interface{}{
			"identifier": handle,
			"password":   "wrongpassword",
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.createSession", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("nonexistent_user", func(t *testing.T) {
		// Create request for non-existent user
		requestBody := map[string]interface{}{
			"identifier": "nonexistent.hashpost.local",
			"password":   "testpassword123",
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.createSession", bytes.NewReader(jsonBody))
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

// TestAtprotoServerCreateAccount tests the createAccount endpoint
func TestAtprotoServerCreateAccount_Integration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.Cleanup()

	t.Run("successful_account_creation", func(t *testing.T) {
		handle := fmt.Sprintf("newuser%d.hashpost.local", time.Now().Unix())
		did := fmt.Sprintf("did:plc:new-user-%d", time.Now().Unix())

		// Create request
		requestBody := map[string]interface{}{
			"handle":   handle,
			"password": "newpassword123",
			"email":    "test@example.com",
			"did":      did,
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.createAccount", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "accessJwt")
		assert.Contains(t, response, "refreshJwt")
		assert.Contains(t, response, "handle")
		assert.Contains(t, response, "did")
		assert.Equal(t, handle, response["handle"])
		assert.Equal(t, did, response["did"])

		// Verify user was created in database
		user, err := suite.db.GetUserByHandle(context.Background(), handle)
		require.NoError(t, err)
		assert.Equal(t, handle, user.Handle)
		assert.Equal(t, did, user.Did)
	})

	t.Run("duplicate_handle", func(t *testing.T) {
		// Create existing user with specific handle
		handle := fmt.Sprintf("existinguser%d.hashpost.local", time.Now().Unix())
		did := fmt.Sprintf("did:plc:existing-user-%d", time.Now().Unix())

		// Create user in database
		_, err := suite.db.CreateUser(context.Background(), &generated.CreateUserParams{
			Did:    did,
			Handle: handle,
		})
		require.NoError(t, err)

		// Try to create account with same handle
		requestBody := map[string]interface{}{
			"handle":   handle,
			"password": "newpassword123",
			"email":    "test2@example.com", // Use different email to avoid email constraint
			"did":      fmt.Sprintf("did:plc:duplicate-user-%d", time.Now().Unix()),
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.createAccount", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusConflict, w.Code)
	})
}

// TestAtprotoServerGetSession tests the getSession endpoint
func TestAtprotoServerGetSession_Integration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.Cleanup()

	t.Run("successful_session_retrieval", func(t *testing.T) {
		// Create test session
		authToken := suite.CreateTestSession(t)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.server.getSession", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "handle")
		assert.Contains(t, response, "did")
		assert.Equal(t, suite.userHandle, response["handle"])
		assert.Equal(t, suite.userDID, response["did"])
	})

	t.Run("invalid_token", func(t *testing.T) {
		// Create request with invalid token
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.server.getSession", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestAtprotoServerDeleteSession tests the deleteSession endpoint
func TestAtprotoServerDeleteSession_Integration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.Cleanup()

	t.Run("successful_session_deletion", func(t *testing.T) {
		// Create test session
		authToken := suite.CreateTestSession(t)

		// Create request
		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.deleteSession", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid_token", func(t *testing.T) {
		// Create request with invalid token
		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.deleteSession", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestAtprotoServerRefreshSession tests the refreshSession endpoint
func TestAtprotoServerRefreshSession_Integration(t *testing.T) {
	suite := SetupIntegrationTestSuite(t)
	defer suite.Cleanup()

	t.Run("successful_token_refresh", func(t *testing.T) {
		// Create test session
		authToken := suite.CreateTestSession(t)

		// Create request
		requestBody := map[string]interface{}{
			"refreshJwt": authToken,
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.refreshSession", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		suite.server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "accessJwt")
		assert.Contains(t, response, "refreshJwt")
	})

	t.Run("invalid_refresh_token", func(t *testing.T) {
		// Create request with invalid refresh token
		requestBody := map[string]interface{}{
			"refreshJwt": "invalid-token",
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.server.refreshSession", bytes.NewReader(jsonBody))
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
