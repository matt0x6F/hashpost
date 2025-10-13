package pds

import (
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

// TestAtprotoIdentityResolveHandle tests the resolveHandle endpoint
func TestAtprotoIdentityResolveHandle_Integration(t *testing.T) {
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

	t.Run("successful_handle_resolution", func(t *testing.T) {
		// Create test user
		handle := fmt.Sprintf("testuser%d.hashpost.local", time.Now().Unix())
		did := fmt.Sprintf("did:plc:test-user-%d", time.Now().Unix())

		_, err := db.CreateUser(context.Background(), &generated.CreateUserParams{
			Did:    did,
			Handle: handle,
		})
		require.NoError(t, err)

		// Create request
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.identity.resolveHandle?handle="+handle, nil)
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

		assert.Contains(t, response, "did")
		assert.Equal(t, did, response["did"])
	})

	t.Run("nonexistent_handle", func(t *testing.T) {
		// Create request for non-existent handle
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.identity.resolveHandle?handle=nonexistent.hashpost.local", nil)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("invalid_handle_format", func(t *testing.T) {
		// Create request with invalid handle format
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.identity.resolveHandle?handle=invalid-handle", nil)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing_handle_parameter", func(t *testing.T) {
		// Create request without handle parameter
		req := httptest.NewRequest(http.MethodGet, "/xrpc/com.atproto.identity.resolveHandle", nil)
		w := httptest.NewRecorder()

		// Create HTTP server and call handler
		mux := http.NewServeMux()
		server.registerAtprotoEndpoints(mux)
		mux.ServeHTTP(w, req)

		// Verify error response
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
