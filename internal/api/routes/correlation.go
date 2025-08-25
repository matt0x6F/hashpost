package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob"
)

// RegisterCorrelationRoutes registers correlation-related routes
func RegisterCorrelationRoutes(api huma.API, db bob.Executor, ibeSystem *ibe.IBESystem, pseudonymDAO *dao.PseudonymDAO, identityMappingDAO *dao.IdentityMappingDAO, postDAO *dao.PostDAO, commentDAO *dao.CommentDAO, subforumDAO *dao.SubforumDAO) {
	correlationAuditDAO := dao.NewCorrelationAuditDAO(db)
	permissionDAO := dao.NewPermissionDAO(db)
	correlationHandler := handlers.NewCorrelationHandler(db, ibeSystem, pseudonymDAO, identityMappingDAO, postDAO, commentDAO, subforumDAO, correlationAuditDAO, permissionDAO)

	// Request fingerprint correlation (moderators)
	huma.Register(api, huma.Operation{
		OperationID: "request-fingerprint-correlation",
		Method:      http.MethodPost,
		Path:        "/correlation/fingerprint",
		Summary:     "Request fingerprint-based correlation for moderation",
		Description: "Request fingerprint-based correlation for moderation purposes (moderators only)",
		Tags:        []string{"Correlation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, correlationHandler.RequestFingerprintCorrelation)

	// Request identity correlation (admins)
	huma.Register(api, huma.Operation{
		OperationID: "request-identity-correlation",
		Method:      http.MethodPost,
		Path:        "/correlation/identity",
		Summary:     "Request identity-based correlation for platform-wide investigations",
		Description: "Request identity-based correlation for platform-wide investigations (admins only)",
		Tags:        []string{"Correlation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, correlationHandler.RequestIdentityCorrelation)

	// Get correlation history
	huma.Register(api, huma.Operation{
		OperationID: "get-correlation-history",
		Method:      http.MethodGet,
		Path:        "/correlation/history",
		Summary:     "Get correlation request history",
		Description: "Get correlation request history for the authenticated user",
		Tags:        []string{"Correlation"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, correlationHandler.GetCorrelationHistory)
}
