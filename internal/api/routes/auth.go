package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/stephenafamo/bob"
)

// RegisterAuthRoutes registers authentication-related routes
func RegisterAuthRoutes(api huma.API, cfg *config.Config, db bob.Executor) {
	authHandler := handlers.NewAuthHandler(cfg, db)

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
}
