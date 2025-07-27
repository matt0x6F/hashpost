package routes

import (
	"database/sql"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob"
)

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(api huma.API, cfg *config.Config, db bob.Executor, rawDB *sql.DB, ibeSystem *ibe.IBESystem) {
	authHandler := handlers.NewAuthHandler(cfg, db, nil, nil, nil, nil, ibeSystem, nil, nil)

	// User registration
	huma.Register(api, huma.Operation{
		OperationID: "register-user",
		Method:      http.MethodPost,
		Path:        "/auth/register",
		Summary:     "Register a new user account",
		Description: "Creates a new user account with pseudonymous identity using IBE",
		Tags:        []string{"Authentication"},
	}, authHandler.RegisterUser)

	// User login
	// Note: JWT cookies are set by the client based on the response tokens
	// The response includes both access_token and refresh_token for client-side cookie management
	huma.Register(api, huma.Operation{
		OperationID: "login-user",
		Method:      http.MethodPost,
		Path:        "/auth/login",
		Summary:     "Authenticate a user",
		Description: "Authenticates a user and returns access tokens with role-based capabilities. The client should set HTTP-only cookies based on the returned tokens.",
		Tags:        []string{"Authentication"},
	}, authHandler.LoginUser)

	// User logout
	// Note: The client should clear cookies based on the logout response
	huma.Register(api, huma.Operation{
		OperationID: "logout-user",
		Method:      http.MethodPost,
		Path:        "/auth/logout",
		Summary:     "Logout a user",
		Description: "Invalidates the user's refresh token. The client should clear any stored tokens and cookies.",
		Tags:        []string{"Authentication"},
	}, authHandler.LogoutUser)

	// Token refresh
	// Note: The client should update the access token cookie based on the response
	huma.Register(api, huma.Operation{
		OperationID: "refresh-token",
		Method:      http.MethodPost,
		Path:        "/auth/refresh",
		Summary:     "Refresh an expired access token",
		Description: "Refreshes an expired access token using a valid refresh token. The client should update the access token cookie with the new token.",
		Tags:        []string{"Authentication"},
	}, authHandler.RefreshToken)

	// Get current user session
	huma.Register(api, huma.Operation{
		OperationID: "get-current-user-session",
		Method:      http.MethodGet,
		Path:        "/auth/me",
		Summary:     "Get current user session data",
		Description: "Retrieves the current user's session data including pseudonyms and metadata. Returns the same structure as login response.",
		Tags:        []string{"Authentication"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, authHandler.GetCurrentUserSession)

	// Get current user session with subforum-specific capabilities
	huma.Register(api, huma.Operation{
		OperationID: "get-current-user-session-for-subforum",
		Method:      http.MethodGet,
		Path:        "/auth/me/subforum/{subforum_name}",
		Summary:     "Get current user session data with subforum-specific capabilities",
		Description: "Retrieves the current user's session data including subforum-specific moderator capabilities and permissions.",
		Tags:        []string{"Authentication"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, authHandler.GetCurrentUserSessionForSubforum)

	// Switch active pseudonym
	huma.Register(api, huma.Operation{
		OperationID: "switch-pseudonym",
		Method:      http.MethodPost,
		Path:        "/auth/switch-pseudonym",
		Summary:     "Switch active pseudonym",
		Description: "Switches the user's active pseudonym and returns a new JWT token with updated pseudonym context.",
		Tags:        []string{"Authentication"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, authHandler.SwitchPseudonym)

	// Deactivate pseudonym
	huma.Register(api, huma.Operation{
		OperationID: "deactivate-pseudonym",
		Method:      http.MethodPost,
		Path:        "/auth/deactivate-pseudonym",
		Summary:     "Deactivate pseudonym",
		Description: "Deactivates a pseudonym owned by the current user. Deactivated pseudonyms cannot be reactivated.",
		Tags:        []string{"Authentication"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, authHandler.DeactivatePseudonym)
}
