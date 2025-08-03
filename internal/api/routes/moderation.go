package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
)

// RegisterModerationRoutes registers moderation-related routes
func RegisterModerationRoutes(api huma.API, reportDAO *dao.ReportDAO, moderationActionDAO *dao.ModerationActionDAO, userBanDAO *dao.UserBanDAO, pseudonymDAO *dao.PseudonymDAO, subforumDAO *dao.SubforumDAO, postDAO *dao.PostDAO, commentDAO *dao.CommentDAO, voteDAO *dao.VoteDAO, permissionDAO *dao.PermissionDAO) {
	moderationHandler := handlers.NewModerationHandler(reportDAO, moderationActionDAO, userBanDAO, pseudonymDAO, subforumDAO, postDAO, commentDAO, voteDAO, permissionDAO)

	// Moderation dashboard statistics
	huma.Register(api, huma.Operation{
		OperationID: "get-moderation-stats",
		Method:      http.MethodGet,
		Path:        "/moderation/{subforum_path}/stats",
		Summary:     "Get moderation dashboard statistics",
		Description: "Get statistics for the moderation dashboard including pending reports, banned users, and engagement metrics",
		Tags:        []string{"Moderation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, moderationHandler.GetModerationStats)

	// Moderation engagement analytics
	huma.Register(api, huma.Operation{
		OperationID: "get-engagement-analytics",
		Method:      http.MethodGet,
		Path:        "/moderation/{subforum_path}/engagement",
		Summary:     "Get engagement analytics data",
		Description: "Get detailed engagement analytics including posts, comments, votes, and voting sentiment over time",
		Tags:        []string{"Moderation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, moderationHandler.GetEngagementAnalytics)

	// Report content
	huma.Register(api, huma.Operation{
		OperationID: "report-content",
		Method:      http.MethodPost,
		Path:        "/reports",
		Summary:     "Report content or users",
		Description: "Report content or users for moderation review",
		Tags:        []string{"Moderation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, moderationHandler.ReportContent)

	// Get reports for specific subforum (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "get-subforum-reports",
		Method:      http.MethodGet,
		Path:        "/moderation/{subforum_path}/reports",
		Summary:     "Get reports for specific subforum",
		Description: "Get reports for moderation review for a specific subforum (moderators only)",
		Tags:        []string{"Moderation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, moderationHandler.GetSubforumReports)

	// Remove content (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "remove-content",
		Method:      http.MethodPost,
		Path:        "/moderation/content/{content_type}/{content_id}/remove",
		Summary:     "Remove content as a moderator",
		Description: "Remove content as a moderator (moderators only)",
		Tags:        []string{"Moderation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, moderationHandler.RemoveContent)

	// Ban user (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "ban-user",
		Method:      http.MethodPost,
		Path:        "/moderation/users/{pseudonym_id}/ban",
		Summary:     "Ban a user from a subforum",
		Description: "Ban a user from a subforum (moderators only)",
		Tags:        []string{"Moderation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, moderationHandler.BanUser)

	// Resolve report (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "resolve-report",
		Method:      http.MethodPost,
		Path:        "/moderation/reports/{report_id}/resolve",
		Summary:     "Resolve a report",
		Description: "Resolve or dismiss a report (moderators only)",
		Tags:        []string{"Moderation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, moderationHandler.ResolveReport)

	// Get moderation history (moderators only)
	huma.Register(api, huma.Operation{
		OperationID: "get-moderation-history",
		Method:      http.MethodGet,
		Path:        "/moderation/history",
		Summary:     "Get moderation action history",
		Description: "Get moderation action history for the authenticated moderator",
		Tags:        []string{"Moderation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, moderationHandler.GetModerationHistory)
}
