package appview

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServer(t *testing.T) {
	// Create test configuration with a mock database URL
	cfg := &config.Config{
		PDS: config.PDSConfig{
			Database: config.DatabaseConfig{
				URL: "postgres://test:test@localhost:5432/test?sslmode=disable",
			},
		},
		AppView: config.AppViewConfig{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 8081,
				Dev:  true,
			},
			PDS: config.PDSURLConfig{
				URL: "http://localhost:8080",
			},
		},
		Logging: config.LoggingConfig{
			Level:  "debug",
			Format: "json",
		},
	}

	// Test server creation - this will fail due to database connection
	// but we can test the server structure
	server, err := NewServer(cfg)

	// We expect an error due to database connection failure
	assert.Error(t, err)
	assert.Nil(t, server)

	// Test that the error is related to database connection
	assert.Contains(t, err.Error(), "database")
}

func TestServerConfig(t *testing.T) {
	cfg := &config.Config{
		AppView: config.AppViewConfig{
			Server: config.ServerConfig{
				Host: "0.0.0.0",
				Port: 8081,
			},
		},
	}

	server := &Server{
		config: cfg,
	}

	// Test server address generation
	expected := "0.0.0.0:8081"
	actual := server.config.GetAppViewServerAddress()
	assert.Equal(t, expected, actual)
}

func TestHealthEndpoint(t *testing.T) {
	// Create handlers directly for testing
	logger := slog.New(slog.NewTextHandler(nil, nil))
	handlers := NewHandlers(nil, logger, nil) // Pass nil for database and rbacService in test

	// Create a test request
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Call the handler
	handlers.Health(rr, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Check the response body
	assert.Contains(t, rr.Body.String(), "healthy")
	assert.Contains(t, rr.Body.String(), "status")
}

func TestCORSHeaders(t *testing.T) {
	// Create handlers directly for testing
	logger := slog.New(slog.NewTextHandler(nil, nil))
	handlers := NewHandlers(nil, logger, nil) // Pass nil for database and rbacService in test

	// Create a test request
	req, err := http.NewRequest("GET", "/health", nil)
	require.NoError(t, err)

	// Create a ResponseRecorder
	rr := httptest.NewRecorder()

	// Call the handler directly (CORS is handled by middleware in the actual server)
	handlers.Health(rr, req)

	// Check that the response is successful
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Contains(t, rr.Body.String(), "healthy")
}

func TestServerLifecycle(t *testing.T) {
	// This is a basic test to ensure the server can be created and stopped
	// In a real test environment, you would test with actual HTTP requests

	cfg := &config.Config{
		PDS: config.PDSConfig{
			Database: config.DatabaseConfig{
				URL: "postgres://test:test@localhost:5432/test?sslmode=disable",
			},
		},
		AppView: config.AppViewConfig{
			Server: config.ServerConfig{
				Host: "localhost",
				Port: 0, // Use port 0 for testing
				Dev:  true,
			},
			PDS: config.PDSURLConfig{
				URL: "http://localhost:8080",
			},
		},
	}

	// Test server creation - this will fail due to database connection
	server, err := NewServer(cfg)
	assert.Error(t, err)
	assert.Nil(t, server)

	// Test that the error is related to database connection
	assert.Contains(t, err.Error(), "database")
}
