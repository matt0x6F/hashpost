package models

import "time"

// HealthInput represents the health check input (empty for GET)
type HealthInput struct{}

// HealthResponseBody represents the body of health check response
type HealthResponseBody struct {
	Status    string `json:"status" example:"healthy"`
	Timestamp string `json:"timestamp" example:"2024-01-01T00:00:00Z"`
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status int                `json:"-" example:"200"`
	Body   HealthResponseBody `json:"body"`
}

// NewHealthResponse creates a new health response with current timestamp
func NewHealthResponse() *HealthResponse {
	return &HealthResponse{
		Status: 200,
		Body: HealthResponseBody{
			Status:    "healthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}
}
