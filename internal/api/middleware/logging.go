package middleware

import (
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/logger"
)

// LoggingMiddleware provides structured request logging for all operations
func LoggingMiddleware(ctx huma.Context, next func(huma.Context)) {
	start := time.Now()
	log := logger.GetRequestLogger()

	// Log the incoming request
	log.Info().
		Str("method", ctx.Method()).
		Str("path", ctx.URL().Path).
		Str("remote_addr", ctx.RemoteAddr()).
		Str("user_agent", ctx.Header("User-Agent")).
		Msg("Request started")

	// Call the next middleware/handler
	next(ctx)

	// Log the response
	duration := time.Since(start)
	log.Info().
		Str("method", ctx.Method()).
		Str("path", ctx.URL().Path).
		Int("status", ctx.Status()).
		Dur("duration", duration).
		Str("remote_addr", ctx.RemoteAddr()).
		Msg("Request completed")
}
