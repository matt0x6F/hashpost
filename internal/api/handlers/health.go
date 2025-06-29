package handlers

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/rs/zerolog/log"
)

// HealthHandler handles health check requests
func HealthHandler(ctx context.Context, input *models.HealthInput) (*models.HealthResponse, error) {
	log.Info().
		Str("endpoint", "health").
		Str("component", "handler").
		Msg("Health check requested")

	response := models.NewHealthResponse()

	log.Info().
		Str("endpoint", "health").
		Str("component", "handler").
		Str("status", "healthy").
		Str("timestamp", response.Body.Timestamp).
		Msg("Health check completed")

	return response, nil
}
