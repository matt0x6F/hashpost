package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
)

// RegisterUserRoutes registers user management-related routes
func RegisterUserRoutes(api huma.API, userDAO *dao.UserDAO, pseudonymDAO *dao.PseudonymDAO, userPreferencesDAO *dao.UserPreferencesDAO, userBlocksDAO *dao.UserBlocksDAO, postDAO *dao.PostDAO, commentDAO *dao.CommentDAO) {
	userHandler := handlers.NewUserHandler(userDAO, pseudonymDAO, userPreferencesDAO, userBlocksDAO, postDAO, commentDAO)

	// Get pseudonym profile (public)
	huma.Register(api, huma.Operation{
		OperationID: "get-pseudonym-profile",
		Method:      http.MethodGet,
		Path:        "/pseudonyms/{pseudonym_id}/profile",
		Summary:     "Get a pseudonym's public profile",
		Description: "Retrieves public profile information for a pseudonym by pseudonym ID",
		Tags:        []string{"Pseudonyms"},
	}, userHandler.GetPseudonymProfile)

	// Update pseudonym profile
	huma.Register(api, huma.Operation{
		OperationID: "update-pseudonym-profile",
		Method:      http.MethodPut,
		Path:        "/pseudonyms/{pseudonym_id}/profile",
		Summary:     "Update a pseudonym's profile",
		Description: "Updates the authenticated user's pseudonym profile information",
		Tags:        []string{"Pseudonyms"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, userHandler.UpdatePseudonymProfile)

	// Create new pseudonym
	huma.Register(api, huma.Operation{
		OperationID: "create-pseudonym",
		Method:      http.MethodPost,
		Path:        "/pseudonyms",
		Summary:     "Create a new pseudonym",
		Description: "Creates a new pseudonym for the authenticated user",
		Tags:        []string{"Pseudonyms"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, userHandler.CreatePseudonym)

	// Get current user profile with all pseudonyms
	huma.Register(api, huma.Operation{
		OperationID: "get-user-profile",
		Method:      http.MethodGet,
		Path:        "/users/profile",
		Summary:     "Get the current user's profile",
		Description: "Retrieves the authenticated user's profile with all associated pseudonyms",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, userHandler.GetUserProfile)

	// Get user preferences
	huma.Register(api, huma.Operation{
		OperationID: "get-user-preferences",
		Method:      http.MethodGet,
		Path:        "/users/preferences",
		Summary:     "Get the current user's preferences",
		Description: "Retrieves the authenticated user's preferences and settings",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, userHandler.GetUserPreferences)

	// Update user preferences
	huma.Register(api, huma.Operation{
		OperationID: "update-user-preferences",
		Method:      http.MethodPut,
		Path:        "/users/preferences",
		Summary:     "Update the current user's preferences",
		Description: "Updates the authenticated user's preferences and settings",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, userHandler.UpdateUserPreferences)

	// Block pseudonym/user
	huma.Register(api, huma.Operation{
		OperationID: "block-user",
		Method:      http.MethodPost,
		Path:        "/users/{pseudonym_id}/block",
		Summary:     "Block a pseudonym",
		Description: "Blocks a user by pseudonym ID. Optionally blocks all personas of the same user.",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, userHandler.BlockUser)

	// Unblock pseudonym/user
	huma.Register(api, huma.Operation{
		OperationID: "unblock-user",
		Method:      http.MethodDelete,
		Path:        "/users/{pseudonym_id}/block",
		Summary:     "Unblock a user",
		Description: "Unblocks a previously blocked user by pseudonym ID",
		Tags:        []string{"Users"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, userHandler.UnblockUser)
}
