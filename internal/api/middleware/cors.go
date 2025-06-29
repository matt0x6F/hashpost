package middleware

import (
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/config"
)

// CORSMiddleware adds CORS headers to responses
func CORSMiddleware(corsConfig *config.CORSConfig) func(huma.Context, func(huma.Context)) {
	return func(ctx huma.Context, next func(huma.Context)) {
		// Get the origin from the request
		origin := ctx.Header("Origin")

		// Determine the allowed origin
		allowedOrigin := determineAllowedOrigin(origin, corsConfig.AllowedOrigins)

		// Add CORS headers
		ctx.SetHeader("Access-Control-Allow-Origin", allowedOrigin)
		ctx.SetHeader("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
		ctx.SetHeader("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))

		if corsConfig.AllowCredentials {
			ctx.SetHeader("Access-Control-Allow-Credentials", "true")
		}

		ctx.SetHeader("Access-Control-Max-Age", string(rune(corsConfig.MaxAge)))

		// Handle preflight OPTIONS requests
		if ctx.Method() == "OPTIONS" {
			// Set status to 200 for preflight requests
			ctx.SetStatus(http.StatusOK)
			// Don't call next() for OPTIONS requests
			return
		}

		// Call the next middleware/handler for non-OPTIONS requests
		next(ctx)
	}
}

// CORSMiddlewareHTTP is a standard HTTP middleware for CORS
func CORSMiddlewareHTTP(corsConfig *config.CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the origin from the request
			origin := r.Header.Get("Origin")

			// Determine the allowed origin
			allowedOrigin := determineAllowedOrigin(origin, corsConfig.AllowedOrigins)

			// Add CORS headers
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))

			if corsConfig.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			w.Header().Set("Access-Control-Max-Age", string(rune(corsConfig.MaxAge)))

			// Handle preflight OPTIONS requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// determineAllowedOrigin determines the appropriate origin to allow based on the request origin and allowed origins
func determineAllowedOrigin(requestOrigin string, allowedOrigins []string) string {
	// If no request origin, return the first allowed origin
	if requestOrigin == "" {
		if len(allowedOrigins) > 0 {
			return allowedOrigins[0]
		}
		return "*"
	}

	// If wildcard is allowed, return the request origin
	if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
		return requestOrigin
	}

	// Check if the request origin is in the allowed origins
	for _, allowed := range allowedOrigins {
		if allowed == requestOrigin {
			return requestOrigin
		}
	}

	// If not found, return the first allowed origin (or empty string)
	if len(allowedOrigins) > 0 {
		return allowedOrigins[0]
	}

	return ""
}
