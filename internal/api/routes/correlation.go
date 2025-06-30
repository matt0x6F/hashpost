package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob"
)

// RegisterCorrelationRoutes registers administrative correlation routes
func RegisterCorrelationRoutes(api huma.API, db bob.Executor, ibeSystem *ibe.IBESystem, securePseudonymDAO *dao.SecurePseudonymDAO, identityMappingDAO *dao.IdentityMappingDAO, postDAO *dao.PostDAO, commentDAO *dao.CommentDAO) {
	correlationHandler := handlers.NewCorrelationHandler(db, ibeSystem, securePseudonymDAO, identityMappingDAO, postDAO, commentDAO)

	// Request fingerprint correlation (moderators)
	huma.Register(api, huma.Operation{
		OperationID: "request-fingerprint-correlation",
		Method:      http.MethodPost,
		Path:        "/admin/correlation/fingerprint",
		Summary:     "Request fingerprint-based correlation for moderation",
		Description: "Request fingerprint-based correlation for moderation purposes (moderators only)",
		Tags:        []string{"Administration", "Correlation"},
	}, correlationHandler.RequestFingerprintCorrelation)

	// Request identity correlation (admins)
	huma.Register(api, huma.Operation{
		OperationID: "request-identity-correlation",
		Method:      http.MethodPost,
		Path:        "/admin/correlation/identity",
		Summary:     "Request identity-based correlation for platform-wide investigations",
		Description: "Request identity-based correlation for platform-wide investigations (admins only)",
		Tags:        []string{"Administration", "Correlation"},
	}, correlationHandler.RequestIdentityCorrelation)

	// Get correlation history
	huma.Register(api, huma.Operation{
		OperationID: "get-correlation-history",
		Method:      http.MethodGet,
		Path:        "/admin/correlation/history",
		Summary:     "Get correlation request history",
		Description: "Get correlation request history for the authenticated user",
		Tags:        []string{"Administration", "Correlation"},
	}, correlationHandler.GetCorrelationHistory)
}
