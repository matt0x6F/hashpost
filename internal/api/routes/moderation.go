package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
)

// RegisterModerationRoutes registers moderation-related routes
func RegisterModerationRoutes(api huma.API) {
	moderationHandler := handlers.NewModerationHandler()

	// Report content
	huma.Register(api, huma.Operation{
		OperationID: "report-content",
		Method:      http.MethodPost,
		Path:        "/reports",
		Summary:     "Report content or users",
		Description: "Report content or users for moderation review",
		Tags:        []string{"Moderation"},
	}, moderationHandler.ReportContent)

	// Get reports (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "get-reports",
		Method:      http.MethodGet,
		Path:        "/moderation/reports",
		Summary:     "Get reports for moderation review",
		Description: "Get reports for moderation review (moderators only)",
		Tags:        []string{"Moderation"},
	}, moderationHandler.GetReports)

	// Remove content (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "remove-content",
		Method:      http.MethodPost,
		Path:        "/moderation/content/{content_type}/{content_id}/remove",
		Summary:     "Remove content as a moderator",
		Description: "Remove content as a moderator (moderators only)",
		Tags:        []string{"Moderation"},
	}, moderationHandler.RemoveContent)

	// Ban user (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "ban-user",
		Method:      http.MethodPost,
		Path:        "/moderation/users/{pseudonym_id}/ban",
		Summary:     "Ban a user from a subforum",
		Description: "Ban a user from a subforum (moderators only)",
		Tags:        []string{"Moderation"},
	}, moderationHandler.BanUser)

	// Get moderation history (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "get-moderation-history",
		Method:      http.MethodGet,
		Path:        "/moderation/history",
		Summary:     "Get moderation action history",
		Description: "Get moderation action history for the authenticated moderator",
		Tags:        []string{"Moderation"},
	}, moderationHandler.GetModerationHistory)
}
