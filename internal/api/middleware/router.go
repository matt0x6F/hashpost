package middleware

import (
	"net/http"
	"time"
)

// RouterMiddleware wraps an http.Handler to add router-specific functionality
type RouterMiddleware struct {
	handler http.Handler
}

// ServeHTTP implements http.Handler interface
func (m *RouterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add router-specific headers
	w.Header().Set("X-Server", "HashPost")
	w.Header().Set("X-Request-ID", generateRequestID())

	// Add timing header
	start := time.Now()

	// Call the next handler
	m.handler.ServeHTTP(w, r)

	// Add response timing header
	duration := time.Since(start)
	w.Header().Set("X-Response-Time", duration.String())
}

// NewRouterMiddleware creates a new router middleware
func NewRouterMiddleware(handler http.Handler) http.Handler {
	return &RouterMiddleware{handler: handler}
}

// generateRequestID creates a simple request ID (in production, use a proper UUID library)
func generateRequestID() string {
	return time.Now().Format("20060102150405")
}
