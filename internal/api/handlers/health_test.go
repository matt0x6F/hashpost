package handlers_test

import (
	"context"
	"testing"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthHandler tests the health check functionality
func TestHealthHandler(t *testing.T) {
	t.Run("HealthCheckSuccess", func(t *testing.T) {
		// Create input
		input := &models.HealthInput{}

		// Call handler
		response, err := handlers.HealthHandler(context.Background(), input)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, "healthy", response.Body.Status)
		assert.NotEmpty(t, response.Body.Timestamp)

		// Verify timestamp format
		_, err = time.Parse(time.RFC3339, response.Body.Timestamp)
		assert.NoError(t, err, "Timestamp should be in RFC3339 format")
	})

	t.Run("HealthCheckWithNilInput", func(t *testing.T) {
		// Call handler with nil input
		response, err := handlers.HealthHandler(context.Background(), nil)

		// Assertions
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, "healthy", response.Body.Status)
		assert.NotEmpty(t, response.Body.Timestamp)
	})

	t.Run("HealthCheckWithCancelledContext", func(t *testing.T) {
		// Create cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// Create input
		input := &models.HealthInput{}

		// Call handler
		response, err := handlers.HealthHandler(ctx, input)

		// Assertions - health check should still work even with cancelled context
		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, 200, response.Status)
		assert.Equal(t, "healthy", response.Body.Status)
	})
}
