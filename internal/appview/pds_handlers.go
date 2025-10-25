package appview

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"

	appview "github.com/matt0x6f/hashpost/internal/database/generated/appview"
)

// PDSHandlers contains handlers for PDS management endpoints
type PDSHandlers struct {
	queries *appview.Queries
	logger  *slog.Logger
}

// NewPDSHandlers creates a new PDS handlers instance
func NewPDSHandlers(queries *appview.Queries, logger *slog.Logger) *PDSHandlers {
	return &PDSHandlers{
		queries: queries,
		logger:  logger,
	}
}

// ListPDSServers handles GET /api/v1/admin/pds/servers
func (h *PDSHandlers) ListPDSServers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	servers, err := h.queries.GetPDSServerStats(r.Context())
	if err != nil {
		h.logger.Error("Failed to get PDS server stats", "error", err)
		http.Error(w, "Failed to get PDS servers", http.StatusInternalServerError)
		return
	}

	// Transform snake_case to camelCase to match OpenAPI schema
	var transformedServers []map[string]interface{}
	for _, server := range servers {
		transformedServer := map[string]interface{}{
			"pdsEndpoint":    server.PdsEndpoint,
			"userCount":      server.UserCount,
			"lastActivity":   server.LastActivity,
			"activeUsers24h": server.ActiveUsers24h,
		}
		transformedServers = append(transformedServers, transformedServer)
	}

	response := map[string]interface{}{
		"servers": transformedServers,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPDSServerDetails handles GET /api/v1/admin/pds/{endpoint}
func (h *PDSHandlers) GetPDSServerDetails(w http.ResponseWriter, r *http.Request, endpoint string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// URL decode the endpoint
	decodedEndpoint, err := url.QueryUnescape(endpoint)
	if err != nil {
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		return
	}

	details, err := h.queries.GetPDSServerDetails(r.Context(), &decodedEndpoint)
	if err != nil {
		h.logger.Error("Failed to get PDS server details", "error", err, "endpoint", decodedEndpoint)
		http.Error(w, "Failed to get PDS server details", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(details)
}

// ListPDSServerUsers handles GET /api/v1/admin/pds/{endpoint}/users
func (h *PDSHandlers) ListPDSServerUsers(w http.ResponseWriter, r *http.Request, endpoint string, params ListPDSServerUsersParams) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// URL decode the endpoint
	decodedEndpoint, err := url.QueryUnescape(endpoint)
	if err != nil {
		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
		return
	}

	limit := 20
	offset := 0
	if params.Limit != nil {
		limit = *params.Limit
	}
	if params.Offset != nil {
		offset = *params.Offset
	}

	users, err := h.queries.GetUsersByPDSSource(r.Context(), &appview.GetUsersByPDSSourceParams{
		PdsSource: &decodedEndpoint,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		h.logger.Error("Failed to get users by PDS source", "error", err, "endpoint", decodedEndpoint)
		http.Error(w, "Failed to get users", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"users":  users,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
