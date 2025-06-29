package middleware

import (
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/rs/zerolog/log"
)

// CookieContext is a context key for storing cookies to be set
type CookieContext string

const (
	// CookieContextKey is the key used to store cookies in request context
	CookieContextKey CookieContext = "cookies_to_set"
)

// CookieData represents a cookie to be set
type CookieData struct {
	Name     string
	Value    string
	Path     string
	HttpOnly bool
	Secure   bool
	SameSite http.SameSite
	Expires  time.Time
	MaxAge   int
}

// SetCookieInContext adds a cookie to be set in the response
func SetCookieInContext(ctx huma.Context, cookie *http.Cookie) {
	// Get existing cookies from context
	cookies, _ := ctx.Context().Value(CookieContextKey).([]*http.Cookie)
	if cookies == nil {
		cookies = make([]*http.Cookie, 0)
	}

	// Add the new cookie
	cookies = append(cookies, cookie)

	// Store back in context
	ctx.Context().Value(CookieContextKey)
}

// CookieMiddleware is a middleware that can set cookies in responses
func CookieMiddleware(ctx huma.Context, next func(huma.Context)) {
	// Call the next middleware/handler
	next(ctx)

	// After the handler has processed, we can set cookies if needed
	// This would require a way to communicate from the handler to the middleware
	// For now, we'll use a simpler approach
}

// SetJWTCookiesForLogin sets JWT cookies for login responses
func SetJWTCookiesForLogin(ctx huma.Context, accessToken, refreshToken string, accessExpiry, refreshExpiry time.Duration) {
	// Set access token cookie
	accessCookie := &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to false in development
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(accessExpiry),
	}
	ctx.AppendHeader("Set-Cookie", accessCookie.String())

	// Set refresh token cookie
	refreshCookie := &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to false in development
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(refreshExpiry),
	}
	ctx.AppendHeader("Set-Cookie", refreshCookie.String())

	log.Info().
		Str("component", "cookie_middleware").
		Msg("JWT cookies set successfully for login")
}

// ClearJWTCookiesForLogout clears JWT cookies for logout responses
func ClearJWTCookiesForLogout(ctx huma.Context) {
	// Clear access token cookie by setting it to expire in the past
	accessCookie := &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to false in development
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-24 * time.Hour), // Expire in the past
		MaxAge:   -1,
	}
	ctx.AppendHeader("Set-Cookie", accessCookie.String())

	// Clear refresh token cookie by setting it to expire in the past
	refreshCookie := &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to false in development
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-24 * time.Hour), // Expire in the past
		MaxAge:   -1,
	}
	ctx.AppendHeader("Set-Cookie", refreshCookie.String())

	log.Info().
		Str("component", "cookie_middleware").
		Msg("JWT cookies cleared successfully for logout")
}

// UpdateAccessTokenCookie updates the access token cookie for token refresh
func UpdateAccessTokenCookie(ctx huma.Context, accessToken string, expiry time.Duration) {
	// Set new access token cookie
	accessCookie := &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to false in development
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(expiry),
	}
	ctx.AppendHeader("Set-Cookie", accessCookie.String())

	log.Info().
		Str("component", "cookie_middleware").
		Msg("Access token cookie updated successfully")
}
