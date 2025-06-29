package handlers

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/api/models"
)

// HelloHandler handles hello world requests
func HelloHandler(ctx context.Context, input *models.HelloInput) (*models.HelloResponse, error) {
	return models.NewHelloResponse(), nil
}
