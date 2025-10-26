package pds

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/matt0x6f/hashpost/internal/config"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/pds"
	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAtprotoRepoCreateRecord tests the createRecord endpoint
func TestAtprotoRepoCreateRecord_Integration(t *testing.T) {
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

	// Create test user
	// Sanitize test name by replacing slashes and hyphens with underscores
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, "-", "_")
	handle := fmt.Sprintf("testuser_%s_%d.hashpost.local", testName, time.Now().UnixNano())
	did := fmt.Sprintf("did:plc:test_user_%s_%d", testName, time.Now().UnixNano())

	user, err := db.CreateUser(context.Background(), &generated.CreateUserParams{
		Did:    did,
		Handle: handle,
	})
	require.NoError(t, err)

	// Create a default subforum for testing
	description := "General discussion"
	_, err = db.CreateSubforum(context.Background(), &generated.CreateSubforumParams{
		Name:        "General",
		Slug:        "general",
		Description: &description,
		CreatedBy:   pgtype.UUID{Bytes: user.ID, Valid: true},
		PrefixType:  "t", // topical subforum
	})
	require.NoError(t, err)

	// Create test session
	sessionID := fmt.Sprintf("session-%s-%d", t.Name(), time.Now().UnixNano())
	_, err = db.CreateUserSession(context.Background(), &generated.CreateUserSessionParams{
		SessionID: sessionID,
		UserDid:   did,
		Handle:    handle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Generate real JWT token
	session := &Session{
		ID:        sessionID,
		DID:       did,
		Handle:    handle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	authToken, _, err := server.authService.GenerateTokens(session)
	require.NoError(t, err)

	// Subforum will be provided by fixtures

	t.Run("successful_record_creation", func(t *testing.T) {
		// Create request
		requestBody := map[string]interface{}{
			"repo":       did,
			"collection": "com.hashpost.feed.post",
			"record": map[string]interface{}{
				"text":      "Hello, world!",
				"createdAt": time.Now().Format(time.RFC3339),
			},
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.repo.createRecord", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "uri")
		assert.Contains(t, response, "cid")
		assert.Contains(t, response, "record")
	})

	t.Run("unauthorized_record_creation", func(t *testing.T) {
		// Create request without auth token
		requestBody := map[string]interface{}{
			"repo":       did,
			"collection": "com.hashpost.feed.post",
			"record": map[string]interface{}{
				"text": "Hello, world!",
			},
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.repo.createRecord", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid_collection", func(t *testing.T) {
		// Create request with invalid collection
		requestBody := map[string]interface{}{
			"repo":       did,
			"collection": "invalid.collection",
			"record": map[string]interface{}{
				"text": "Hello, world!",
			},
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.repo.createRecord", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestAtprotoRepoGetRecord tests the getRecord endpoint
func TestAtprotoRepoGetRecord_Integration(t *testing.T) {
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

	// Create test user
	// Sanitize test name by replacing slashes and hyphens with underscores
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, "-", "_")
	handle := fmt.Sprintf("testuser_%s_%d.hashpost.local", testName, time.Now().UnixNano())
	did := fmt.Sprintf("did:plc:test_user_%s_%d", testName, time.Now().UnixNano())

	_, err = db.CreateUser(context.Background(), &generated.CreateUserParams{
		Did:    did,
		Handle: handle,
	})
	require.NoError(t, err)

	// Create test session
	sessionID := fmt.Sprintf("session-%s-%d", t.Name(), time.Now().UnixNano())
	_, err = db.CreateUserSession(context.Background(), &generated.CreateUserSessionParams{
		SessionID: sessionID,
		UserDid:   did,
		Handle:    handle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Generate real JWT token
	session := &Session{
		ID:        sessionID,
		DID:       did,
		Handle:    handle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	authToken, _, err := server.authService.GenerateTokens(session)
	require.NoError(t, err)

	t.Run("successful_record_retrieval", func(t *testing.T) {
		// Get user from database
		user, err := db.GetUserByDID(context.Background(), did)
		require.NoError(t, err)

		// Create a test post first
		atprotoURI := fmt.Sprintf("at://%s/com.hashpost.feed.post/test-record", did)
		userID := pgtype.UUID{Bytes: user.ID, Valid: true}
		_, err = db.CreatePost(context.Background(), &generated.CreatePostParams{
			UserID:     userID,
			Title:      "Test Post",
			Content:    "Hello, world!",
			AtprotoUri: &atprotoURI,
		})
		require.NoError(t, err)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.repo.getRecord?repo="+did+"&collection=com.hashpost.feed.post&rkey=test-record", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "uri")
		assert.Contains(t, response, "cid")
		assert.Contains(t, response, "record")
	})

	t.Run("nonexistent_record", func(t *testing.T) {
		// Create request for non-existent record
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.repo.getRecord?repo="+did+"&collection=com.hashpost.feed.post&rkey=nonexistent", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("missing_parameters", func(t *testing.T) {
		// Create request without required parameters
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.repo.getRecord", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestAtprotoRepoListRecords tests the listRecords endpoint
func TestAtprotoRepoListRecords_Integration(t *testing.T) {
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

	// Create test user
	// Sanitize test name by replacing slashes and hyphens with underscores
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, "-", "_")
	handle := fmt.Sprintf("testuser_%s_%d.hashpost.local", testName, time.Now().UnixNano())
	did := fmt.Sprintf("did:plc:test_user_%s_%d", testName, time.Now().UnixNano())

	user, err := db.CreateUser(context.Background(), &generated.CreateUserParams{
		Did:    did,
		Handle: handle,
	})
	require.NoError(t, err)

	// Create a default subforum for testing
	description := "General discussion"
	_, err = db.CreateSubforum(context.Background(), &generated.CreateSubforumParams{
		Name:        "General",
		Slug:        "general",
		Description: &description,
		CreatedBy:   pgtype.UUID{Bytes: user.ID, Valid: true},
		PrefixType:  "t", // topical subforum
	})
	require.NoError(t, err)

	// Create test session
	sessionID := fmt.Sprintf("session-%s-%d", t.Name(), time.Now().UnixNano())
	_, err = db.CreateUserSession(context.Background(), &generated.CreateUserSessionParams{
		SessionID: sessionID,
		UserDid:   did,
		Handle:    handle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Generate real JWT token
	session := &Session{
		ID:        sessionID,
		DID:       did,
		Handle:    handle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	authToken, _, err := server.authService.GenerateTokens(session)
	require.NoError(t, err)

	// Get the subforum ID for the posts
	subforums, err := db.ListSubforums(context.Background(), &generated.ListSubforumsParams{
		Limit:  1,
		Offset: 0,
	})
	require.NoError(t, err)
	require.Len(t, subforums, 1)
	subforumID := pgtype.UUID{Bytes: subforums[0].ID, Valid: true}

	// Create some test posts
	userID := pgtype.UUID{Bytes: user.ID, Valid: true}
	for i := 0; i < 3; i++ {
		atprotoURI := fmt.Sprintf("at://%s/com.hashpost.feed.post/test-record-%d", did, i)
		_, err = db.CreatePost(context.Background(), &generated.CreatePostParams{
			UserID:     userID,
			SubforumID: subforumID,
			Title:      fmt.Sprintf("Test Post %d", i),
			Content:    fmt.Sprintf("Test content %d", i),
			AtprotoUri: &atprotoURI,
		})
		require.NoError(t, err)
	}

	t.Run("successful_record_listing", func(t *testing.T) {
		// Create request
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.repo.listRecords?repo="+did+"&collection=com.hashpost.feed.post", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "records")
		records := response["records"].([]interface{})
		assert.Len(t, records, 3)
	})

	t.Run("unauthorized_record_listing", func(t *testing.T) {
		// Create request without auth token
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.repo.listRecords?repo="+did+"&collection=com.hashpost.feed.post", nil)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing_parameters", func(t *testing.T) {
		// Create request without required parameters
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.repo.listRecords", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestAtprotoRepoDeleteRecord tests the deleteRecord endpoint
func TestAtprotoRepoDeleteRecord_Integration(t *testing.T) {
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

	// Create test user
	// Sanitize test name by replacing slashes and hyphens with underscores
	testName := strings.ReplaceAll(t.Name(), "/", "_")
	testName = strings.ReplaceAll(testName, "-", "_")
	handle := fmt.Sprintf("testuser_%s_%d.hashpost.local", testName, time.Now().UnixNano())
	did := fmt.Sprintf("did:plc:test_user_%s_%d", testName, time.Now().UnixNano())

	user, err := db.CreateUser(context.Background(), &generated.CreateUserParams{
		Did:    did,
		Handle: handle,
	})
	require.NoError(t, err)

	// Create test session
	sessionID := fmt.Sprintf("session-%s-%d", t.Name(), time.Now().UnixNano())
	_, err = db.CreateUserSession(context.Background(), &generated.CreateUserSessionParams{
		SessionID: sessionID,
		UserDid:   did,
		Handle:    handle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// Generate real JWT token
	session := &Session{
		ID:        sessionID,
		DID:       did,
		Handle:    handle,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	authToken, _, err := server.authService.GenerateTokens(session)
	require.NoError(t, err)

	// Create a test post
	atprotoURI := fmt.Sprintf("at://%s/com.hashpost.feed.post/test-record", did)
	userID := pgtype.UUID{Bytes: user.ID, Valid: true}
	_, err = db.CreatePost(context.Background(), &generated.CreatePostParams{
		UserID:     userID,
		Title:      "Test Post",
		Content:    "Hello, world!",
		AtprotoUri: &atprotoURI,
	})
	require.NoError(t, err)

	t.Run("successful_record_deletion", func(t *testing.T) {
		// Create request
		requestBody := map[string]interface{}{
			"repo":       did,
			"collection": "com.hashpost.feed.post",
			"rkey":       "test-record",
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.repo.deleteRecord", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify response
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthorized_record_deletion", func(t *testing.T) {
		// Create request without auth token
		requestBody := map[string]interface{}{
			"repo":       did,
			"collection": "com.hashpost.feed.post",
			"rkey":       "test-record",
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.repo.deleteRecord", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing_parameters", func(t *testing.T) {
		// Create request without required parameters
		requestBody := map[string]interface{}{
			"repo": did,
		}
		jsonBody, err := json.Marshal(requestBody)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/xrpc/com.atproto.repo.deleteRecord", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authToken)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
