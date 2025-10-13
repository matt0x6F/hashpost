package appview

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
)

// Handlers contains simple HTTP handlers for the AppView
type Handlers struct {
	queries *generated.Queries
	logger  *slog.Logger
	pdsURL  string
}

// NewHandlers creates a new simple handlers instance
func NewHandlers(queries *generated.Queries, logger *slog.Logger) *Handlers {
	return &Handlers{
		queries: queries,
		logger:  logger,
		pdsURL:  "http://hashpost-pds:8080", // Default PDS URL
	}
}

// ListSubforums handles GET /api/v1/subforums
func (h *Handlers) ListSubforums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Parse query parameters
	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	h.logger.Debug("Listing subforums", "limit", limit, "offset", offset)

	// Query subforums from database
	subforums, err := h.queries.ListAppViewSubforums(r.Context(), &generated.ListAppViewSubforumsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.logger.Error("Failed to list subforums", "error", err)
		http.Error(w, "Failed to list subforums", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	subforumList := make([]map[string]interface{}, len(subforums))
	for i, subforum := range subforums {
		subforumList[i] = map[string]interface{}{
			"id":                subforum.ID,
			"name":              subforum.Name,
			"slug":              subforum.Slug,
			"description":       subforum.Description,
			"created_by":        subforum.CreatedByDid,
			"created_by_handle": subforum.CreatedByHandle,
			"created_at":        formatTimestamptz(subforum.CreatedAt),
			"updated_at":        formatTimestamptz(subforum.UpdatedAt),
			"subscriber_count":  subforum.SubscriberCount,
			"post_count":        subforum.PostCount,
			"comment_count":     subforum.CommentCount,
		}
	}

	response := map[string]interface{}{
		"subforums": subforumList,
		"total":     len(subforumList),
		"limit":     limit,
		"offset":    offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SubforumsHandler handles GET and POST /api/v1/subforums
func (h *Handlers) SubforumsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListSubforums(w, r)
	case http.MethodPost:
		h.CreateSubforum(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// SubforumBySlugHandler handles GET, PUT, DELETE /api/v1/subforums/{slug}
func (h *Handlers) SubforumBySlugHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetSubforumBySlug(w, r)
	case http.MethodPut:
		h.UpdateSubforum(w, r)
	case http.MethodDelete:
		h.DeleteSubforum(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetSubforumBySlug handles GET /api/v1/subforums/{slug}
func (h *Handlers) GetSubforumBySlug(w http.ResponseWriter, r *http.Request) {
	// Extract slug from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/subforums/")
	if path == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Missing subforum slug")
		return
	}

	h.logger.Debug("Getting subforum by slug", "slug", path)

	// Query subforum from database
	subforum, err := h.queries.GetAppViewSubforumBySlug(r.Context(), path)
	if err != nil {
		h.logger.Error("Failed to get subforum", "error", err, "slug", path)
		http.Error(w, "Subforum not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"id":                subforum.ID,
		"name":              subforum.Name,
		"slug":              subforum.Slug,
		"description":       subforum.Description,
		"created_by":        subforum.CreatedByDid,
		"created_by_handle": subforum.CreatedByHandle,
		"created_at":        formatTimestamptz(subforum.CreatedAt),
		"updated_at":        formatTimestamptz(subforum.UpdatedAt),
		"subscriber_count":  subforum.SubscriberCount,
		"post_count":        subforum.PostCount,
		"comment_count":     subforum.CommentCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Health handles GET /health
func (h *Handlers) Health(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// writeError writes an error response
func (h *Handlers) writeError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	error := map[string]interface{}{
		"error":   errorCode,
		"message": message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(error)
}

// CreateSubforum handles POST /api/v1/subforums
func (h *Handlers) CreateSubforum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		Name        string `json:"name"`
		Slug        string `json:"slug"`
		Description string `json:"description,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Name == "" || req.Slug == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name and slug are required")
		return
	}

	// Get user context from authentication
	userCtx := GetUserContext(r)
	if userCtx == nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Create subforum in database
	description := &req.Description
	if req.Description == "" {
		description = nil
	}
	createdSubforum, err := h.queries.CreateAppViewSubforum(r.Context(), &generated.CreateAppViewSubforumParams{
		Name:            req.Name,
		Slug:            req.Slug,
		Description:     description,
		CreatedByDid:    userCtx.Did,
		CreatedByHandle: userCtx.Handle,
	})
	if err != nil {
		h.logger.Error("Failed to create subforum", "error", err)
		h.writeError(w, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create subforum")
		return
	}

	subforum := map[string]interface{}{
		"id":                createdSubforum.ID,
		"name":              createdSubforum.Name,
		"slug":              createdSubforum.Slug,
		"description":       createdSubforum.Description,
		"created_by":        createdSubforum.CreatedByDid,
		"created_by_handle": createdSubforum.CreatedByHandle,
		"created_at":        formatTimestamptz(createdSubforum.CreatedAt),
		"subscriber_count":  createdSubforum.SubscriberCount,
		"post_count":        createdSubforum.PostCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(subforum)
}

// UpdateSubforum handles PUT /api/v1/subforums/{slug}
func (h *Handlers) UpdateSubforum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract subforum slug from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.writeError(w, http.StatusBadRequest, "INVALID_SUBFORUM_SLUG", "Invalid subforum slug")
		return
	}
	slug := pathParts[3]

	// Parse request body
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Name == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name is required")
		return
	}

	// Get user context from authentication
	userCtx := GetUserContext(r)
	if userCtx == nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Update subforum in database
	description := &req.Description
	if req.Description == "" {
		description = nil
	}
	updatedSubforum, err := h.queries.UpdateAppViewSubforum(r.Context(), &generated.UpdateAppViewSubforumParams{
		Slug:        slug,
		Name:        req.Name,
		Description: description,
	})
	if err != nil {
		h.logger.Error("Failed to update subforum", "error", err, "slug", slug)
		h.writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update subforum")
		return
	}

	subforum := map[string]interface{}{
		"id":                updatedSubforum.ID,
		"name":              updatedSubforum.Name,
		"slug":              updatedSubforum.Slug,
		"description":       updatedSubforum.Description,
		"created_by":        updatedSubforum.CreatedByDid,
		"created_by_handle": updatedSubforum.CreatedByHandle,
		"created_at":        formatTimestamptz(updatedSubforum.CreatedAt),
		"updated_at":        formatTimestamptz(updatedSubforum.UpdatedAt),
		"subscriber_count":  updatedSubforum.SubscriberCount,
		"post_count":        updatedSubforum.PostCount,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(subforum)
}

// DeleteSubforum handles DELETE /api/v1/subforums/{slug}
func (h *Handlers) DeleteSubforum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract subforum slug from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.writeError(w, http.StatusBadRequest, "INVALID_SUBFORUM_SLUG", "Invalid subforum slug")
		return
	}
	slug := pathParts[3]

	// Get user context from authentication
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Delete subforum from database
	err := h.queries.DeleteAppViewSubforum(r.Context(), slug)
	if err != nil {
		h.logger.Error("Failed to delete subforum", "error", err, "slug", slug)
		http.Error(w, "Failed to delete subforum", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Subforum deleted", "slug", slug)
	w.WriteHeader(http.StatusNoContent)
}

// formatTimestamptz formats a pgtype.Timestamptz as RFC3339 string
func formatTimestamptz(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return time.Now().Format(time.RFC3339)
	}
	return ts.Time.Format(time.RFC3339)
}
