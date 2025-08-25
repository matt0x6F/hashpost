package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/stephenafamo/bob"
)

// RegisterRulesRoutes registers all rules-related routes
func RegisterRulesRoutes(api huma.API, db bob.Executor, pseudonymDAO dao.PseudonymDAOInterface) {
	// Initialize DAOs for rules handler
	reportsDAO := dao.NewReportDAO(db)
	subforumDAO := dao.NewSubforumDAO(db)
	systemSettingsDAO := dao.NewSystemSettingsDAO(db)
	permissionDAO := dao.NewPermissionDAO(db)

	rulesHandler := handlers.NewRulesHandler(reportsDAO, subforumDAO, systemSettingsDAO, permissionDAO, pseudonymDAO, db)

	// Platform rules routes
	huma.Register(api, huma.Operation{
		OperationID: "get-platform-rules",
		Method:      http.MethodGet,
		Path:        "/platform/rules",
		Summary:     "Get platform rules",
		Description: "Retrieves all platform-wide rules for content moderation.",
		Tags:        []string{"Rules"},
	}, rulesHandler.GetPlatformRules)

	huma.Register(api, huma.Operation{
		OperationID: "update-platform-rules",
		Method:      http.MethodPut,
		Path:        "/platform/rules",
		Summary:     "Update platform rules",
		Description: "Updates platform-wide rules. Requires system_admin capability.",
		Tags:        []string{"Rules"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, rulesHandler.UpdatePlatformRules)

	// Subforum rules routes
	huma.Register(api, huma.Operation{
		OperationID: "get-subforum-rules",
		Method:      http.MethodGet,
		Path:        "/rules/subforums/{community_type}/{subforum_name}",
		Summary:     "Get subforum rules",
		Description: "Retrieves all rules for a specific subforum.",
		Tags:        []string{"Rules"},
	}, rulesHandler.GetSubforumRules)

	// Subforum rule management routes (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "create-subforum-rule",
		Method:      http.MethodPost,
		Path:        "/rules/subforums/{community_type}/{subforum_name}",
		Summary:     "Create a new subforum rule",
		Description: "Creates a new rule for a subforum. Requires moderator permissions.",
		Tags:        []string{"Rules"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, rulesHandler.CreateSubforumRule)

	huma.Register(api, huma.Operation{
		OperationID: "update-subforum-rule",
		Method:      http.MethodPut,
		Path:        "/rules/subforums/{community_type}/{subforum_name}/{rule_code}",
		Summary:     "Update a subforum rule",
		Description: "Updates an existing rule for a subforum. Requires moderator permissions.",
		Tags:        []string{"Rules"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, rulesHandler.UpdateSubforumRule)

	huma.Register(api, huma.Operation{
		OperationID: "delete-subforum-rule",
		Method:      http.MethodDelete,
		Path:        "/rules/subforums/{community_type}/{subforum_name}/{rule_code}",
		Summary:     "Delete a subforum rule",
		Description: "Deletes a rule from a subforum. Requires moderator permissions.",
		Tags:        []string{"Rules"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, rulesHandler.DeleteSubforumRule)

	// Rule violation reporting routes
	huma.Register(api, huma.Operation{
		OperationID: "report-rule-violation",
		Method:      http.MethodPost,
		Path:        "/reports/rule-violation",
		Summary:     "Report a rule violation",
		Description: "Reports a violation of a specific platform or subforum rule.",
		Tags:        []string{"Reports"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, rulesHandler.ReportRuleViolation)

	// Report forwarding routes (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "forward-report-to-platform",
		Method:      http.MethodPost,
		Path:        "/reports/{report_id}/forward",
		Summary:     "Forward a report to platform moderators",
		Description: "Forwards a subforum report to platform-level moderators with notes. Requires moderator permissions.",
		Tags:        []string{"Reports"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, rulesHandler.ForwardReportToPlatform)
}
