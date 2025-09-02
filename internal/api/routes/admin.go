package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/services"
)

// RegisterAdminRoutes registers administrative routes
func RegisterAdminRoutes(api huma.API, userDAO *dao.UserDAO, pseudonymDAO *dao.PseudonymDAO, permissionDAO *dao.PermissionDAO, passwordResetTokenDAO *dao.PasswordResetTokenDAO, emailService *services.EmailService, config *config.Config, postDAO *dao.PostDAO, commentDAO *dao.CommentDAO, roleKeyDAO *dao.RoleKeyDAO, subforumDAO *dao.SubforumDAO) {
	adminHandler := handlers.NewAdminHandler(userDAO, pseudonymDAO, permissionDAO, passwordResetTokenDAO, emailService, config, postDAO, commentDAO, roleKeyDAO, subforumDAO)

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

	// Update specific user (admin only)
	huma.Register(api, huma.Operation{
		OperationID: "admin-update-user",
		Method:      http.MethodPut,
		Path:        "/admin/users/{user_id}",
		Summary:     "Update specific user for admin purposes",
		Description: "Updates user details. Requires platform admin capability. Password changes are not allowed through this endpoint.",
		Tags:        []string{"Admin"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, adminHandler.UpdateUser)

	// Trigger password reset for specific user (admin only)
	huma.Register(api, huma.Operation{
		OperationID: "admin-trigger-password-reset",
		Method:      http.MethodPost,
		Path:        "/admin/users/{user_id}/trigger-password-reset",
		Summary:     "Trigger password reset for specific user",
		Description: "Triggers a password reset workflow for a user. Requires platform admin capability.",
		Tags:        []string{"Admin"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, adminHandler.TriggerPasswordReset)

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

	// Get specific pseudonym (admin only)
	huma.Register(api, huma.Operation{
		OperationID: "admin-get-pseudonym",
		Method:      http.MethodGet,
		Path:        "/admin/pseudonyms/{pseudonym_id}",
		Summary:     "Get specific pseudonym for admin purposes",
		Description: "Retrieves detailed information about a specific pseudonym. Requires platform admin capability.",
		Tags:        []string{"Admin"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, adminHandler.GetPseudonym)

	// Update specific pseudonym (admin only)
	huma.Register(api, huma.Operation{
		OperationID: "admin-update-pseudonym",
		Method:      http.MethodPatch,
		Path:        "/admin/pseudonyms/{pseudonym_id}",
		Summary:     "Update specific pseudonym for admin purposes",
		Description: "Updates a specific pseudonym including roles and capabilities. Requires platform admin capability.",
		Tags:        []string{"Admin"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, adminHandler.UpdatePseudonym)
}
