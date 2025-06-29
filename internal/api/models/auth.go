package models

import (
	"net/http"
	"time"
)

// UserRegistrationBody represents the body of user registration request/response
type UserRegistrationBody struct {
	Email       string `json:"email" example:"user@example.com"`
	Password    string `json:"password" example:"secure_password"`
	DisplayName string `json:"display_name" example:"user_display_name"`
	Bio         string `json:"bio,omitempty" example:"Optional user bio"`
	WebsiteURL  string `json:"website_url,omitempty" example:"https://example.com"`
	Timezone    string `json:"timezone,omitempty" example:"UTC"`
	Language    string `json:"language,omitempty" example:"en"`
}

// UserRegistrationInput represents user registration request
type UserRegistrationInput struct {
	Body UserRegistrationBody `json:"body"`
}

// UserLoginBody represents the body of user login request
type UserLoginBody struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secure_password"`
}

// UserLoginInput represents user login request
type UserLoginInput struct {
	Body UserLoginBody `json:"body"`
}

// TokenRefreshBody represents the body of token refresh request
type TokenRefreshBody struct {
	RefreshToken string `json:"refresh_token" example:"refresh_token_here"`
}

// TokenRefreshInput represents token refresh request
type TokenRefreshInput struct {
	Body TokenRefreshBody `json:"body"`
}

// RefreshTokenBody represents the body of refresh token request
type RefreshTokenBody struct {
	RefreshToken string `json:"refresh_token" example:"refresh_token_here"`
}

// RefreshTokenInput represents refresh token request
type RefreshTokenInput struct {
	Body RefreshTokenBody `json:"body"`
}

// UserLogoutBody represents the body of user logout request
type UserLogoutBody struct {
	RefreshToken string `json:"refresh_token" example:"refresh_token_here"`
}

// UserLogoutInput represents user logout request
type UserLogoutInput struct {
	Body UserLogoutBody `json:"body"`
}

// UserInfo represents user information in responses
type UserInfo struct {
	UserID       int      `json:"user_id" example:"123"`
	Email        string   `json:"email" example:"user@example.com"`
	CreatedAt    string   `json:"created_at" example:"2024-01-01T12:00:00Z"`
	LastActiveAt string   `json:"last_active_at" example:"2024-01-01T18:00:00Z"`
	IsActive     bool     `json:"is_active" example:"true"`
	IsSuspended  bool     `json:"is_suspended" example:"false"`
	Roles        []string `json:"roles" example:"user"`
	Capabilities []string `json:"capabilities" example:"create_content,vote,message,report"`
}

// PseudonymInfo represents pseudonym information in responses
type PseudonymInfo struct {
	PseudonymID  string `json:"pseudonym_id" example:"abc123def456..."`
	DisplayName  string `json:"display_name" example:"user_display_name"`
	KarmaScore   int    `json:"karma_score" example:"0"`
	CreatedAt    string `json:"created_at" example:"2024-01-01T12:00:00Z"`
	LastActiveAt string `json:"last_active_at" example:"2024-01-01T18:00:00Z"`
	IsActive     bool   `json:"is_active" example:"true"`
}

// TokenInfo represents token information
type TokenInfo struct {
	AccessToken  string `json:"access_token" example:"jwt_token_here"`
	RefreshToken string `json:"refresh_token" example:"refresh_token_here"`
	ExpiresIn    int    `json:"expires_in" example:"3600"`
}

// UserRegistrationResponseBody represents the body of user registration response
type UserRegistrationResponseBody struct {
	UserID       int      `json:"user_id" example:"123"`
	Email        string   `json:"email" example:"user@example.com"`
	CreatedAt    string   `json:"created_at" example:"2024-01-01T12:00:00Z"`
	LastActiveAt string   `json:"last_active_at" example:"2024-01-01T18:00:00Z"`
	IsActive     bool     `json:"is_active" example:"true"`
	IsSuspended  bool     `json:"is_suspended" example:"false"`
	Roles        []string `json:"roles" example:"user"`
	Capabilities []string `json:"capabilities" example:"create_content,vote,message,report"`
	PseudonymID  string   `json:"pseudonym_id" example:"abc123def456..."`
	DisplayName  string   `json:"display_name" example:"user_display_name"`
	KarmaScore   int      `json:"karma_score" example:"0"`
	AccessToken  string   `json:"access_token" example:"jwt_token_here"`
	RefreshToken string   `json:"refresh_token" example:"refresh_token_here"`
	ExpiresIn    int      `json:"expires_in" example:"3600"`
}

// UserRegistrationResponse represents user registration response
type UserRegistrationResponse struct {
	Status int                          `json:"-" example:"200"`
	Body   UserRegistrationResponseBody `json:"body"`
}

// UserLoginResponseBody represents the body of user login response
type UserLoginResponseBody struct {
	UserID       int      `json:"user_id" example:"123"`
	Email        string   `json:"email" example:"user@example.com"`
	CreatedAt    string   `json:"created_at" example:"2024-01-01T12:00:00Z"`
	LastActiveAt string   `json:"last_active_at" example:"2024-01-01T18:00:00Z"`
	IsActive     bool     `json:"is_active" example:"true"`
	IsSuspended  bool     `json:"is_suspended" example:"false"`
	Roles        []string `json:"roles" example:"[\"user\"]"`
	Capabilities []string `json:"capabilities" example:"[\"create_content\",\"vote\",\"message\",\"report\"]"`
	// JWT tokens (also available in cookies)
	AccessToken  string `json:"access_token" example:"jwt_access_token_here"`
	RefreshToken string `json:"refresh_token" example:"jwt_refresh_token_here"`
	// Pseudonym information
	ActivePseudonymID string          `json:"active_pseudonym_id" example:"pseudonym_123"`
	DisplayName       string          `json:"display_name" example:"User123"`
	Pseudonyms        []PseudonymInfo `json:"pseudonyms"`
}

// TokenRefreshResponseBody represents the body of token refresh response
type TokenRefreshResponseBody struct {
	AccessToken string `json:"access_token" example:"new_jwt_token_here"`
	ExpiresIn   int    `json:"expires_in" example:"3600"`
}

// UserLoginResponse represents a successful user login response
type UserLoginResponse struct {
	Status int                   `json:"-" example:"200"`
	Body   UserLoginResponseBody `json:"body"`
	// JWT cookies set automatically by Huma
	Cookies []http.Cookie `header:"Set-Cookie"`
}

// TokenRefreshResponse represents a token refresh response
type TokenRefreshResponse struct {
	Status int                      `json:"-" example:"200"`
	Body   TokenRefreshResponseBody `json:"body"`
	// JWT cookie set automatically by Huma
	Cookies []http.Cookie `header:"Set-Cookie"`
}

// UserLogoutResponse represents user logout response
type UserLogoutResponse struct {
	Status  int           `json:"-" example:"200"`
	Message string        `json:"message" example:"Logout successful"`
	Cookies []http.Cookie `header:"Set-Cookie"`
}

// CurrentUserSessionResponseBody represents the body of current user session response
type CurrentUserSessionResponseBody struct {
	UserID            int             `json:"user_id" example:"123"`
	Email             string          `json:"email" example:"user@example.com"`
	CreatedAt         string          `json:"created_at" example:"2024-01-01T12:00:00Z"`
	LastActiveAt      string          `json:"last_active_at" example:"2024-01-01T18:00:00Z"`
	IsActive          bool            `json:"is_active" example:"true"`
	IsSuspended       bool            `json:"is_suspended" example:"false"`
	Roles             []string        `json:"roles" example:"[\"user\"]"`
	Capabilities      []string        `json:"capabilities" example:"[\"create_content\",\"vote\",\"message\",\"report\"]"`
	ActivePseudonymID string          `json:"active_pseudonym_id" example:"pseudonym_123"`
	DisplayName       string          `json:"display_name" example:"User123"`
	Pseudonyms        []PseudonymInfo `json:"pseudonyms"`
}

// CurrentUserSessionResponse represents current user session response
type CurrentUserSessionResponse struct {
	Status int                            `json:"-" example:"200"`
	Body   CurrentUserSessionResponseBody `json:"body"`
}

// NewUserRegistrationResponse creates a new user registration response
func NewUserRegistrationResponse(userID int, email string, roles, capabilities []string, pseudonymID, displayName string, accessToken, refreshToken string) *UserRegistrationResponse {
	now := time.Now().UTC().Format(time.RFC3339)
	return &UserRegistrationResponse{
		Status: 200,
		Body: UserRegistrationResponseBody{
			UserID:       userID,
			Email:        email,
			CreatedAt:    now,
			LastActiveAt: now,
			IsActive:     true,
			IsSuspended:  false,
			Roles:        roles,
			Capabilities: capabilities,
			PseudonymID:  pseudonymID,
			DisplayName:  displayName,
			KarmaScore:   0,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			ExpiresIn:    3600,
		},
	}
}

// NewUserLoginResponse creates a new user login response
func NewUserLoginResponse(accessToken, refreshToken string, userID int, email string, roles, capabilities []string, activePseudonymID, displayName string, pseudonyms []PseudonymInfo, isDevelopment bool) *UserLoginResponse {
	// Create cookies for JWT tokens
	accessCookie := http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Domain:   "", // Empty domain means current domain
		HttpOnly: true,
		Secure:   !isDevelopment, // Secure only in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(24 * time.Hour), // 24 hours
	}

	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Path:     "/",
		Domain:   "", // Empty domain means current domain
		HttpOnly: true,
		Secure:   !isDevelopment, // Secure only in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	return &UserLoginResponse{
		Status: 200,
		Body: UserLoginResponseBody{
			UserID:            userID,
			Email:             email,
			CreatedAt:         time.Now().Format(time.RFC3339),
			LastActiveAt:      time.Now().Format(time.RFC3339),
			IsActive:          true,
			IsSuspended:       false,
			Roles:             roles,
			Capabilities:      capabilities,
			AccessToken:       accessToken,
			RefreshToken:      refreshToken,
			ActivePseudonymID: activePseudonymID,
			DisplayName:       displayName,
			Pseudonyms:        pseudonyms,
		},
		Cookies: []http.Cookie{accessCookie, refreshCookie},
	}
}

// NewTokenRefreshResponse creates a new token refresh response
func NewTokenRefreshResponse(accessToken string, expiresIn int, isDevelopment bool) *TokenRefreshResponse {
	// Create cookie for the new access token
	accessCookie := http.Cookie{
		Name:     "access_token",
		Value:    accessToken,
		Path:     "/",
		Domain:   "", // Empty domain means current domain
		HttpOnly: true,
		Secure:   !isDevelopment, // Secure only in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(time.Duration(expiresIn) * time.Second),
	}

	return &TokenRefreshResponse{
		Status: 200,
		Body: TokenRefreshResponseBody{
			AccessToken: accessToken,
			ExpiresIn:   expiresIn,
		},
		Cookies: []http.Cookie{accessCookie},
	}
}

// NewCurrentUserSessionResponse creates a new current user session response
func NewCurrentUserSessionResponse(userID int, email string, roles, capabilities []string, activePseudonymID, displayName string, pseudonyms []PseudonymInfo) *CurrentUserSessionResponse {
	return &CurrentUserSessionResponse{
		Status: 200,
		Body: CurrentUserSessionResponseBody{
			UserID:            userID,
			Email:             email,
			CreatedAt:         time.Now().Format(time.RFC3339),
			LastActiveAt:      time.Now().Format(time.RFC3339),
			IsActive:          true,
			IsSuspended:       false,
			Roles:             roles,
			Capabilities:      capabilities,
			ActivePseudonymID: activePseudonymID,
			DisplayName:       displayName,
			Pseudonyms:        pseudonyms,
		},
	}
}

// NewUserLogoutResponse creates a new user logout response with expired cookies
func NewUserLogoutResponse(isDevelopment bool) *UserLogoutResponse {
	// Create expired cookies to clear the existing JWT cookies
	accessCookie := http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		Domain:   "", // Empty domain means current domain
		HttpOnly: true,
		Secure:   !isDevelopment, // Secure only in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-1 * time.Hour), // Expired in the past
		MaxAge:   -1,                             // Immediate expiration
	}

	refreshCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Domain:   "", // Empty domain means current domain
		HttpOnly: true,
		Secure:   !isDevelopment, // Secure only in production
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(-1 * time.Hour), // Expired in the past
		MaxAge:   -1,                             // Immediate expiration
	}

	return &UserLogoutResponse{
		Status:  200,
		Message: "Logout successful. Cookies have been cleared.",
		Cookies: []http.Cookie{accessCookie, refreshCookie},
	}
}
