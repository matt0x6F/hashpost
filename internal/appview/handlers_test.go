package appview

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/matt0x6f/hashpost/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlers_Health(t *testing.T) {
	// Create handlers with mock dependencies
	logger := testutil.CreateMockLogger()
	handlers := &Handlers{
		logger: logger,
	}

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	// Call handler
	handlers.Health(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check response content
	assert.Equal(t, "healthy", response["status"])
	assert.NotNil(t, response["timestamp"])
}

func TestHandlers_Structure(t *testing.T) {
	// Test that Handlers can be created and has expected fields
	logger := testutil.CreateMockLogger()
	handlers := &Handlers{
		logger: logger,
	}

	// Test that the struct has the expected fields
	require.NotNil(t, handlers.logger)
	assert.Equal(t, logger, handlers.logger)
	// pdsURL is set in NewHandlers, not in the struct literal
	assert.Equal(t, "", handlers.pdsURL) // Will be empty in struct literal
}

func TestHandlers_WriteError(t *testing.T) {
	// Create handlers with mock dependencies
	logger := testutil.CreateMockLogger()
	handlers := &Handlers{
		logger: logger,
	}

	// Test writeError method
	w := httptest.NewRecorder()
	handlers.writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Test error message")

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Check error content
	assert.Equal(t, "INVALID_REQUEST", response["error"])
	assert.Equal(t, "Test error message", response["message"])
}
