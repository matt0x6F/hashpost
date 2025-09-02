package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
)

// RegisterAdminRoutes registers administrative routes
func RegisterAdminRoutes(api huma.API, userDAO *dao.UserDAO, pseudonymDAO *dao.PseudonymDAO, permissionDAO *dao.PermissionDAO) {
	adminHandler := handlers.NewAdminHandler(userDAO, pseudonymDAO, permissionDAO)

	// List all users (admin only)
	huma.Register(api, huma.Operation{
		OperationID: "admin-list-users",
		Method:      http.MethodGet,
		Path:        "/admin/users",
		Summary:     "List all users for admin purposes",
		Description: "Retrieves a paginated list of all users. Requires platform admin capability. Pseudonym details are not included for privacy.",
		Tags:        []string{"Admin"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, adminHandler.ListUsers)

	// Get specific user (admin only)
	huma.Register(api, huma.Operation{
		OperationID: "admin-get-user",
		Method:      http.MethodGet,
		Path:        "/admin/users/{user_id}",
		Summary:     "Get specific user for admin purposes",
		Description: "Retrieves detailed information about a specific user. Requires platform admin capability. Pseudonym details are not included for privacy.",
		Tags:        []string{"Admin"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, adminHandler.GetUser)

	// List all pseudonyms (admin only)
	huma.Register(api, huma.Operation{
		OperationID: "admin-list-pseudonyms",
		Method:      http.MethodGet,
		Path:        "/admin/pseudonyms",
		Summary:     "List all pseudonyms for admin purposes",
		Description: "Retrieves a paginated list of all pseudonyms. Requires platform admin capability.",
		Tags:        []string{"Admin"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, adminHandler.ListPseudonyms)
}
