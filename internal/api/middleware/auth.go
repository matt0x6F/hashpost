package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/rs/zerolog/log"
)

// UserContextKey is the context key for user information
type UserContextKey string

const (
	// UserContextKeyValue is the key used to store user context in request context
	UserContextKeyValue UserContextKey = "user_context"

	// Cookie names for JWT tokens
	AccessTokenCookie  = "access_token"
	RefreshTokenCookie = "refresh_token"
)

// Global auth middleware instance for Huma functions
var globalAuthMiddleware *AuthMiddleware

// SetGlobalAuthMiddleware sets the global auth middleware instance
// This should be called during server initialization
func SetGlobalAuthMiddleware(authMiddleware *AuthMiddleware) {
	globalAuthMiddleware = authMiddleware
}

// GetGlobalAuthMiddleware returns the global auth middleware instance
func GetGlobalAuthMiddleware() *AuthMiddleware {
	return globalAuthMiddleware
}

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID            int64    `json:"user_id"`
	Email             string   `json:"email"`
	Roles             []string `json:"roles"`
	Capabilities      []string `json:"capabilities"`
	MFAEnabled        bool     `json:"mfa_enabled"`
	ActivePseudonymID string   `json:"active_pseudonym_id"`
	DisplayName       string   `json:"display_name"`
	jwt.RegisteredClaims
}

// UserContext contains user information extracted from JWT token or API token
type UserContext struct {
	UserID       int64    `json:"user_id"`
	Email        string   `json:"email"`
	Roles        []string `json:"roles"`
	Capabilities []string `json:"capabilities"`
	MFAEnabled   bool     `json:"mfa_enabled"`
	// Pseudonym information for the current session
	ActivePseudonymID string `json:"active_pseudonym_id"`
	DisplayName       string `json:"display_name"`
	// Token type for tracking
	TokenType string `json:"token_type"` // "jwt" or "api_token"
}

// HasCapability checks if the user has a specific capability
func (uc *UserContext) HasCapability(capability string) bool {
	for _, cap := range uc.Capabilities {
		if cap == capability {
			return true
		}
	}
	return false
}

// HasRole checks if the user has a specific role
func (uc *UserContext) HasRole(role string) bool {
	for _, r := range uc.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// RequiresMFA checks if an action requires MFA based on user's roles
func (uc *UserContext) RequiresMFA(action string) bool {
	// Check if MFA is globally enabled
	authMiddleware := GetGlobalAuthMiddleware()
	if authMiddleware != nil && authMiddleware.securityConfig != nil && !authMiddleware.securityConfig.EnableMFA {
		// MFA is disabled globally, so no action requires MFA
		return false
	}

	// Actions that require MFA for any user
	mfaRequiredActions := map[string]bool{
		"correlate_identities": true,
		"system_admin":         true,
		"legal_compliance":     true,
	}

	// Actions that require MFA for users with correlation capabilities
	if uc.HasCapability("correlate_fingerprints") || uc.HasCapability("correlate_identities") {
		mfaRequiredActions["correlate_fingerprints"] = true
	}

	return mfaRequiredActions[action]
}

// ExtractUserFromContext extracts user context from the request context
func ExtractUserFromContext(ctx context.Context) (*UserContext, error) {
	userCtx, ok := ctx.Value(UserContextKeyValue).(*UserContext)
	if !ok {
		return nil, fmt.Errorf("user context not found in request context")
	}
	return userCtx, nil
}

// ExtractUserFromRequest extracts user context from the HTTP request
func ExtractUserFromRequest(r *http.Request) (*UserContext, error) {
	ctx := r.Context()
	return ExtractUserFromContext(ctx)
}

// SetUserContext sets user context in the request context
func SetUserContext(ctx context.Context, userCtx *UserContext) context.Context {
	return context.WithValue(ctx, UserContextKeyValue, userCtx)
}

// AuthMiddleware handles authentication and authorization
type AuthMiddleware struct {
	jwtSecret      []byte
	apiKeyDAO      *dao.APIKeyDAO
	jwtConfig      *config.JWTConfig
	securityConfig *config.SecurityConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtSecret string, apiKeyDAO *dao.APIKeyDAO, jwtConfig *config.JWTConfig, securityConfig *config.SecurityConfig) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret:      []byte(jwtSecret),
		apiKeyDAO:      apiKeyDAO,
		jwtConfig:      jwtConfig,
		securityConfig: securityConfig,
	}
}

// validateAndParseJWT validates and parses a JWT token
func (m *AuthMiddleware) validateAndParseJWT(tokenString string) (*JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// validateAPIToken validates a static API token
func (m *AuthMiddleware) validateAPIToken(tokenString string) (*UserContext, error) {
	if m.apiKeyDAO == nil {
		return nil, fmt.Errorf("API key DAO not initialized")
	}

	// Validate the API key using the DAO
	permissions, pseudonymID, err := m.apiKeyDAO.ValidateAPIKey(context.Background(), tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid API token: %w", err)
	}

	// Create user context from API key permissions
	userContext := &UserContext{
		UserID:            0,  // API keys don't have a specific user ID
		Email:             "", // API keys don't have an email
		Roles:             permissions.Roles,
		Capabilities:      permissions.Capabilities,
		MFAEnabled:        false,       // API keys don't use MFA
		ActivePseudonymID: pseudonymID, // Set the pseudonym ID from the API key
		DisplayName:       "",          // Will be loaded from pseudonym if needed
		TokenType:         "api_token",
	}

	return userContext, nil
}

// extractTokenFromRequest extracts token from either header or cookie
func (m *AuthMiddleware) extractTokenFromRequest(r *http.Request) (*UserContext, error) {
	// First, try to extract from Authorization header (for API tokens)
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			if token != "" {
				log.Debug().Str("token_type", "api_token").Msg("Attempting API token validation")
				return m.validateAPIToken(token)
			}
		}
	}

	// If no valid header token, try to extract from cookies (for JWTs)
	accessToken, err := r.Cookie(AccessTokenCookie)
	if err == nil && accessToken.Value != "" {
		log.Debug().Str("token_type", "jwt").Msg("Attempting JWT validation from cookie")
		claims, err := m.validateAndParseJWT(accessToken.Value)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT token: %w", err)
		}

		// Extract user context from JWT claims
		userContext := &UserContext{
			UserID:            claims.UserID,
			Email:             claims.Email,
			Roles:             claims.Roles,
			Capabilities:      claims.Capabilities,
			MFAEnabled:        claims.MFAEnabled,
			ActivePseudonymID: claims.ActivePseudonymID,
			DisplayName:       claims.DisplayName,
			TokenType:         "jwt",
		}

		return userContext, nil
	}

	// No valid token found
	return nil, ErrNoAuthHeader
}

// AuthInput represents authentication input for Huma handlers
type AuthInput struct {
	Authorization string `header:"Authorization" doc:"Bearer token for API authentication"`
	AccessToken   string `cookie:"access_token" doc:"JWT access token from cookie"`
}

// AuthenticatedInput is a helper type that can be embedded in input structs for protected endpoints
type AuthenticatedInput struct {
	AuthInput
}

// NewAuthenticatedInput creates a new authenticated input struct
// This can be embedded in handler input structs for protected endpoints
func NewAuthenticatedInput() *AuthenticatedInput {
	return &AuthenticatedInput{}
}

// ExtractUserFromAuthenticatedInput extracts user context from an authenticated input
func ExtractUserFromAuthenticatedInput(input *AuthenticatedInput) (*UserContext, error) {
	return ExtractUserFromHumaInput(&input.AuthInput)
}

// extractTokenFromHumaInput extracts token from either header or cookie using Huma input struct
func (m *AuthMiddleware) extractTokenFromHumaInput(input *AuthInput) (*UserContext, error) {
	log.Debug().
		Str("Authorization", input.Authorization).
		Str("AccessToken", input.AccessToken).
		Msg("Extracting token from Huma input")
	// First, try to extract from Authorization header (for JWT tokens)
	if input.Authorization != "" {
		parts := strings.Split(input.Authorization, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			token := parts[1]
			if token != "" {
				// Try JWT validation first (most common case)
				log.Debug().Str("token_type", "jwt").Msg("Attempting JWT validation from Authorization header")
				claims, err := m.validateAndParseJWT(token)
				if err == nil {
					// JWT validation succeeded
					userContext := &UserContext{
						UserID:            claims.UserID,
						Email:             claims.Email,
						Roles:             claims.Roles,
						Capabilities:      claims.Capabilities,
						MFAEnabled:        claims.MFAEnabled,
						ActivePseudonymID: claims.ActivePseudonymID,
						DisplayName:       claims.DisplayName,
						TokenType:         "jwt",
					}
					return userContext, nil
				}

				// JWT validation failed, try API token validation
				log.Debug().Str("token_type", "api_token").Msg("JWT validation failed, attempting API token validation")
				userContext, err := m.validateAPIToken(token)
				if err == nil {
					return userContext, nil
				}

				// Both JWT and API token validation failed
				return nil, fmt.Errorf("invalid token: %w", err)
			}
		}
	}

	// If no valid header token, try to extract from cookies (for JWTs)
	if input.AccessToken != "" {
		log.Debug().Str("token_type", "jwt").Msg("Attempting JWT validation from cookie")
		claims, err := m.validateAndParseJWT(input.AccessToken)
		if err != nil {
			return nil, fmt.Errorf("invalid JWT token: %w", err)
		}

		// Extract user context from JWT claims
		userContext := &UserContext{
			UserID:            claims.UserID,
			Email:             claims.Email,
			Roles:             claims.Roles,
			Capabilities:      claims.Capabilities,
			MFAEnabled:        claims.MFAEnabled,
			ActivePseudonymID: claims.ActivePseudonymID,
			DisplayName:       claims.DisplayName,
			TokenType:         "jwt",
		}

		return userContext, nil
	}

	// No valid token found
	return nil, ErrNoAuthHeader
}

// AuthenticateUser middleware extracts user from JWT token or API token and adds to context
func (m *AuthMiddleware) AuthenticateUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract user from token (either header or cookie)
		userCtx, err := m.extractTokenFromRequest(r)
		if err != nil {
			// No auth header or cookie - continue without user context
			log.Debug().Str("path", r.URL.Path).Err(err).Msg("No valid authentication token provided")
			next.ServeHTTP(w, r)
			return
		}

		// Add user context to request context
		ctx := SetUserContext(r.Context(), userCtx)
		r = r.WithContext(ctx)

		log.Debug().
			Int64("user_id", userCtx.UserID).
			Str("email", userCtx.Email).
			Str("token_type", userCtx.TokenType).
			Str("path", r.URL.Path).
			Msg("User authenticated")

		next.ServeHTTP(w, r)
	})
}

// AuthenticateUserHuma is a Huma-compatible authentication middleware
func AuthenticateUserHuma(ctx huma.Context, next func(huma.Context)) {
	// For Huma, we need to handle authentication differently
	// Extract token from headers or cookies using the input struct
	var input AuthInput

	// Try to extract authorization header
	if authHeader := ctx.Header("Authorization"); authHeader != "" {
		log.Debug().Str("auth_header", authHeader).Msg("Received Authorization header")
		input.Authorization = authHeader
	}

	log.Debug().Str("input.Authorization", input.Authorization).Msg("AuthInput before extraction")

	var userCtx *UserContext

	// Use the global auth middleware instance
	authMiddleware := GetGlobalAuthMiddleware()
	if authMiddleware == nil {
		log.Error().Msg("Global auth middleware not initialized")
		next(ctx)
		return
	}

	// Extract user context from input (header only for middleware)
	userCtx, _ = authMiddleware.extractTokenFromHumaInput(&input)

	if userCtx == nil {
		log.Debug().Msg("No valid authentication token provided or token parsing failed")
		next(ctx)
		return
	}

	// Add user context to request context
	SetUserContext(ctx.Context(), userCtx)

	log.Debug().
		Int64("user_id", userCtx.UserID).
		Str("email", userCtx.Email).
		Str("token_type", userCtx.TokenType).
		Str("path", ctx.URL().Path).
		Msg("User authenticated")

	next(ctx)
}

// ExtractUserFromTokenHuma extracts user information from JWT token or API token for Huma context
func ExtractUserFromTokenHuma(ctx huma.Context) (*UserContext, error) {
	// Create input struct to extract tokens
	var input AuthInput

	// Try to extract authorization header
	if authHeader := ctx.Header("Authorization"); authHeader != "" {
		input.Authorization = authHeader
	}

	// Note: For Huma, cookies should be accessed through input structs in handlers
	// This function is limited to header-based authentication
	// For cookie-based auth, use ExtractUserFromHumaInput instead

	// Use the global auth middleware instance
	authMiddleware := GetGlobalAuthMiddleware()
	if authMiddleware == nil {
		return nil, fmt.Errorf("global auth middleware not initialized")
	}

	// Extract user context from input (header only)
	return authMiddleware.extractTokenFromHumaInput(&input)
}

// ExtractUserFromHumaInput extracts user information from a Huma input struct
// This is the preferred way to handle authentication in Huma handlers
func ExtractUserFromHumaInput(input *AuthInput) (*UserContext, error) {
	// Use the global auth middleware instance
	authMiddleware := GetGlobalAuthMiddleware()
	if authMiddleware == nil {
		return nil, fmt.Errorf("global auth middleware not initialized")
	}

	// Extract user context from input
	return authMiddleware.extractTokenFromHumaInput(input)
}

// Example of how to use AuthInput in a Huma handler:
//
// ```go
// // In your handler function:
// func (h *MyHandler) ProtectedEndpoint(ctx context.Context, input *struct {
//     middleware.AuthInput
//     // Your other input fields here
//     Data string `json:"data"`
// }) (*MyResponse, error) {
//     // Extract user from the AuthInput embedded in your input struct
//     userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
//     if err != nil {
//         return nil, huma.Error401Unauthorized("Authentication required")
//     }
//
//     // Now you have access to user information
//     log.Info().Int64("user_id", userCtx.UserID).Msg("User accessed protected endpoint")
//
//     // Your handler logic here...
//     return &MyResponse{}, nil
// }
//
// // Or for endpoints that only need authentication (no body):
// func (h *MyHandler) GetProtectedData(ctx context.Context, input *middleware.AuthInput) (*MyResponse, error) {
//     userCtx, err := middleware.ExtractUserFromHumaInput(input)
//     if err != nil {
//         return nil, huma.Error401Unauthorized("Authentication required")
//     }
//
//     // Your handler logic here...
//     return &MyResponse{}, nil
// }

// validateAndParseJWT validates and parses a JWT token
func validateAndParseJWT(tokenString string, jwtSecret []byte) (*JWTClaims, error) {
	// Parse the token
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// RequireAuth middleware requires authentication for protected endpoints
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := ExtractUserFromRequest(r)
		if err != nil {
			log.Warn().Err(err).Str("path", r.URL.Path).Msg("Authentication required but not provided")
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		log.Debug().
			Int64("user_id", userCtx.UserID).
			Str("token_type", userCtx.TokenType).
			Str("path", r.URL.Path).
			Msg("User authenticated for protected endpoint")

		next.ServeHTTP(w, r)
	})
}

// RequireCapability middleware checks if the user has the required capability
func (m *AuthMiddleware) RequireCapability(capability string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, err := ExtractUserFromRequest(r)
			if err != nil {
				log.Warn().Err(err).Str("path", r.URL.Path).Msg("Authentication required for capability check")
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if !userCtx.HasCapability(capability) {
				log.Warn().
					Int64("user_id", userCtx.UserID).
					Str("capability", capability).
					Str("path", r.URL.Path).
					Msg("User lacks required capability")
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			log.Debug().
				Int64("user_id", userCtx.UserID).
				Str("capability", capability).
				Str("path", r.URL.Path).
				Msg("User has required capability")

			next.ServeHTTP(w, r)
		})
	}
}

// RequireMFA middleware checks if MFA is required and validated for sensitive operations
func (m *AuthMiddleware) RequireMFA(action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, err := ExtractUserFromRequest(r)
			if err != nil {
				log.Warn().Err(err).Str("path", r.URL.Path).Msg("Authentication required for MFA check")
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			// Check if MFA is globally enabled
			if m.securityConfig != nil && !m.securityConfig.EnableMFA {
				// MFA is disabled globally, so skip MFA checks
				log.Debug().
					Int64("user_id", userCtx.UserID).
					Str("action", action).
					Str("path", r.URL.Path).
					Msg("MFA check skipped - MFA disabled globally")
				next.ServeHTTP(w, r)
				return
			}

			if userCtx.RequiresMFA(action) {
				// TODO: Validate MFA token if required
				log.Warn().
					Int64("user_id", userCtx.UserID).
					Str("action", action).
					Str("path", r.URL.Path).
					Msg("MFA required but not validated")
				http.Error(w, "Multi-factor authentication required", http.StatusForbidden)
				return
			}

			log.Debug().
				Int64("user_id", userCtx.UserID).
				Str("action", action).
				Str("path", r.URL.Path).
				Msg("MFA check passed")

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole middleware checks if the user has the required role
func (m *AuthMiddleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userCtx, err := ExtractUserFromRequest(r)
			if err != nil {
				log.Warn().Err(err).Str("path", r.URL.Path).Msg("Authentication required for role check")
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if !userCtx.HasRole(role) {
				log.Warn().
					Int64("user_id", userCtx.UserID).
					Str("role", role).
					Str("path", r.URL.Path).
					Msg("User lacks required role")
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			log.Debug().
				Int64("user_id", userCtx.UserID).
				Str("role", role).
				Str("path", r.URL.Path).
				Msg("User has required role")

			next.ServeHTTP(w, r)
		})
	}
}

// ExtractUserFromToken extracts user information from JWT token or API token
func (m *AuthMiddleware) ExtractUserFromToken(r *http.Request) (*UserContext, error) {
	return m.extractTokenFromRequest(r)
}

// GenerateJWT generates a new JWT token for a user
func GenerateJWT(userCtx *UserContext, jwtSecret string, expiration time.Duration) (string, error) {
	claims := &JWTClaims{
		UserID:            userCtx.UserID,
		Email:             userCtx.Email,
		Roles:             userCtx.Roles,
		Capabilities:      userCtx.Capabilities,
		MFAEnabled:        userCtx.MFAEnabled,
		ActivePseudonymID: userCtx.ActivePseudonymID,
		DisplayName:       userCtx.DisplayName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// SetJWTCookies sets JWT tokens as HTTP-only cookies
func SetJWTCookies(w http.ResponseWriter, accessToken, refreshToken string, accessExpiry, refreshExpiry time.Duration) {
	// Set access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to false in development
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(accessExpiry),
	})

	// Set refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true, // Set to false in development
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(refreshExpiry),
	})
}

// SetJWTCookiesWithConfig sets JWT tokens as HTTP-only cookies with configuration
func (m *AuthMiddleware) SetJWTCookiesWithConfig(w http.ResponseWriter, accessToken, refreshToken string, accessExpiry, refreshExpiry time.Duration) {
	// Determine secure setting based on development mode
	secure := !m.jwtConfig.Development

	// Set access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(accessExpiry),
	})

	// Set refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(refreshExpiry),
	})
}

// ClearJWTCookies clears JWT cookies on logout
func ClearJWTCookies(w http.ResponseWriter) {
	// Clear access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-1 * time.Hour), // Expire immediately
		MaxAge:   -1,
	})

	// Clear refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     RefreshTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-1 * time.Hour), // Expire immediately
		MaxAge:   -1,
	})
}

// SetJWTCookiesHuma sets JWT tokens as HTTP-only cookies using Huma context
func SetJWTCookiesHuma(ctx huma.Context, accessToken, refreshToken string, accessExpiry, refreshExpiry time.Duration) {
	// Get the global auth middleware to access configuration
	authMiddleware := GetGlobalAuthMiddleware()
	secure := true // Default to secure
	if authMiddleware != nil && authMiddleware.jwtConfig != nil {
		secure = !authMiddleware.jwtConfig.Development
	}

	// Set access token cookie
	accessCookie := &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
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
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(refreshExpiry),
	}
	ctx.AppendHeader("Set-Cookie", refreshCookie.String())

	log.Info().
		Str("component", "auth_middleware").
		Msg("JWT cookies set successfully")
}

// ClearJWTCookiesHuma clears JWT cookies using Huma context
func ClearJWTCookiesHuma(ctx huma.Context) {
	// Get the global auth middleware to access configuration
	authMiddleware := GetGlobalAuthMiddleware()
	secure := true // Default to secure
	if authMiddleware != nil && authMiddleware.jwtConfig != nil {
		secure = !authMiddleware.jwtConfig.Development
	}

	// Clear access token cookie by setting it to expire in the past
	accessCookie := &http.Cookie{
		Name:     AccessTokenCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
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
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-24 * time.Hour), // Expire in the past
		MaxAge:   -1,
	}
	ctx.AppendHeader("Set-Cookie", refreshCookie.String())

	log.Info().
		Str("component", "auth_middleware").
		Msg("JWT cookies cleared successfully")
}

// Custom errors
var (
	ErrNoAuthHeader      = &AuthError{Code: "NO_AUTH_HEADER", Message: "Authorization header or cookie is required"}
	ErrInvalidAuthHeader = &AuthError{Code: "INVALID_AUTH_HEADER", Message: "Invalid authorization header format"}
	ErrInvalidToken      = &AuthError{Code: "INVALID_TOKEN", Message: "Invalid or expired token"}
	ErrInsufficientPerms = &AuthError{Code: "INSUFFICIENT_PERMISSIONS", Message: "Insufficient permissions for this operation"}
	ErrMFARequired       = &AuthError{Code: "MFA_REQUIRED", Message: "Multi-factor authentication required for this operation"}
)

// AuthError represents authentication/authorization errors
type AuthError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *AuthError) Error() string {
	return e.Message
}
