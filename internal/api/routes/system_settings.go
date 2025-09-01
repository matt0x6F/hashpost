package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob"
)

// RegisterSystemSettingsRoutes registers all system settings-related routes
func RegisterSystemSettingsRoutes(api huma.API, db bob.Executor, pseudonymDAO dao.PseudonymDAOInterface) {
	// Initialize DAOs for system settings handler
	systemSettingsDAO := dao.NewSystemSettingsDAO(db)
	permissionDAO := dao.NewPermissionDAO(db)
	ibeSystem := ibe.NewIBESystemWithOptions(ibe.IBEOptions{})

	systemSettingsHandler := handlers.NewSystemSettingsHandler(systemSettingsDAO, permissionDAO, pseudonymDAO, ibeSystem)

	// System settings routes
	huma.Register(api, huma.Operation{
		OperationID: "get-system-setting",
		Method:      http.MethodGet,
		Path:        "/system/settings/{setting_key}",
		Summary:     "Get system setting",
		Description: "Retrieves a specific system setting. Requires system_admin capability.",
		Tags:        []string{"System Settings"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, systemSettingsHandler.GetSystemSetting)

	huma.Register(api, huma.Operation{
		OperationID: "update-system-setting",
		Method:      http.MethodPut,
		Path:        "/system/settings",
		Summary:     "Update system setting",
		Description: "Updates a system setting. Requires system_admin capability.",
		Tags:        []string{"System Settings"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, systemSettingsHandler.UpdateSystemSetting)
}
