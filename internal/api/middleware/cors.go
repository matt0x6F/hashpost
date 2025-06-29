package middleware

import (
	"github.com/danielgtaylor/huma/v2"
)

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware(ctx huma.Context, next func(huma.Context)) {
	// Add CORS headers
	ctx.SetHeader("Access-Control-Allow-Origin", "*")
	ctx.SetHeader("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	ctx.SetHeader("Access-Control-Allow-Headers", "Content-Type, Authorization")

	// Handle preflight requests
	if ctx.Method() == "OPTIONS" {
		ctx.SetStatus(200)
		return
	}

	// Call the next middleware/handler
	next(ctx)
}
