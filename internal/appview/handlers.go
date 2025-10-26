package appview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
)

// Handlers contains simple HTTP handlers for the AppView
type Handlers struct {
	queries     *generated.Queries
	logger      *slog.Logger
	pdsURL      string
	rbacService *RBACService
}

// NewHandlers creates a new simple handlers instance
func NewHandlers(queries *generated.Queries, logger *slog.Logger, rbacService *RBACService) *Handlers {
	return &Handlers{
		queries:     queries,
		logger:      logger,
		pdsURL:      "http://hashpost-pds:8080", // Default PDS URL
		rbacService: rbacService,
	}
}

// ListSubforumsWithParams handles GET /api/v1/subforums with parsed parameters
func (h *Handlers) ListSubforumsWithParams(w http.ResponseWriter, r *http.Request, limit int, offset int) {
	h.logger.Debug("Listing subforums", "limit", limit, "offset", offset)

	// Query subforums from database
	subforums, err := h.queries.ListAppViewSubforums(r.Context(), &generated.ListAppViewSubforumsParams{
		Limit:  int32(limit),
		Offset: int32(offset),
	})

	h.logger.Debug("Query result", "count", len(subforums), "error", err)
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
			"prefix_type":       subforum.PrefixType,
		}
	}

	response := map[string]interface{}{
		"subforums": subforumList,
		"total":     len(subforumList),
		"limit":     limit,
		"offset":    offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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

	h.logger.Debug("Query result", "count", len(subforums), "error", err)
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
			"prefix_type":       subforum.PrefixType,
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
		"prefix_type":       subforum.PrefixType,
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

// writeJSONResponse writes a JSON response
func (h *Handlers) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
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
		PrefixType  string `json:"prefix_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Name == "" || req.Slug == "" || req.PrefixType == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "name, slug, and prefix_type are required")
		return
	}

	// Validate prefix_type
	if req.PrefixType != "h" && req.PrefixType != "r" && req.PrefixType != "t" {
		h.writeError(w, http.StatusBadRequest, "INVALID_PREFIX_TYPE", "prefix_type must be 'h', 'r', or 't'")
		return
	}

	// Get user context from authentication
	userCtx := GetUserContext(r)
	if userCtx == nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Check permissions for 'h' prefix (HashPost-centric, admin only)
	if req.PrefixType == "h" {
		hasAdminRole := false
		for _, role := range userCtx.Roles {
			if role.RoleName == "platform_admin" {
				hasAdminRole = true
				break
			}
		}
		if !hasAdminRole {
			h.writeError(w, http.StatusForbidden, "INSUFFICIENT_PERMISSIONS", "Only platform administrators can create HashPost-centric subforums")
			return
		}
	}

	// Auto-prepend prefix to slug
	prefixedSlug := req.PrefixType + "-" + req.Slug

	// Create subforum in database
	description := &req.Description
	if req.Description == "" {
		description = nil
	}
	createdSubforum, err := h.queries.CreateAppViewSubforum(r.Context(), &generated.CreateAppViewSubforumParams{
		Name:            req.Name,
		Slug:            prefixedSlug,
		Description:     description,
		CreatedByDid:    userCtx.Did,
		CreatedByHandle: userCtx.Handle,
		PrefixType:      req.PrefixType,
	})
	if err != nil {
		h.logger.Error("Failed to create subforum", "error", err)
		h.writeError(w, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create subforum")
		return
	}

	// Assign subforum_owner role to creator
	err = h.rbacService.AssignSubforumRole(r.Context(), prefixedSlug, userCtx.Did, "subforum_owner", userCtx.Did, nil)
	if err != nil {
		h.logger.Error("Failed to assign subforum_owner role", "error", err, "slug", prefixedSlug, "user", userCtx.Did)
		// Continue anyway - subforum was created successfully
	}

	// Assign subforum_moderator role to creator
	err = h.rbacService.AssignSubforumRole(r.Context(), prefixedSlug, userCtx.Did, "subforum_moderator", userCtx.Did, nil)
	if err != nil {
		h.logger.Error("Failed to assign subforum_moderator role", "error", err, "slug", prefixedSlug, "user", userCtx.Did)
		// Continue anyway - subforum was created successfully
	}

	// Auto-subscribe creator to the subforum
	_, err = h.queries.CreateSubscription(r.Context(), &generated.CreateSubscriptionParams{
		UserDid:      userCtx.Did,
		UserHandle:   userCtx.Handle,
		SubforumSlug: prefixedSlug,
	})
	if err != nil {
		h.logger.Error("Failed to auto-subscribe creator", "error", err, "slug", prefixedSlug, "user", userCtx.Did)
		// Continue anyway - subforum was created successfully
	}

	// Update subscriber count
	err = h.queries.UpdateSubforumSubscriberCount(r.Context(), prefixedSlug)
	if err != nil {
		h.logger.Error("Failed to update subscriber count", "error", err, "slug", prefixedSlug)
		// Continue anyway
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
		"comment_count":     createdSubforum.CommentCount,
		"prefix_type":       createdSubforum.PrefixType,
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

// GetPostMetrics handles GET /api/v1/posts/{id}/metrics
func (h *Handlers) GetPostMetrics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	postIDStr := pathParts[4]

	ctx := r.Context()

	// Parse post ID as UUID
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		http.Error(w, "Invalid post ID format", http.StatusBadRequest)
		return
	}

	// Get post metrics from database
	commentCount, err := h.queries.GetPostCommentCount(ctx, postID)
	if err != nil {
		h.logger.Error("Failed to get post comment count", "error", err, "post_id", postIDStr)
		http.Error(w, "Failed to get post metrics", http.StatusInternalServerError)
		return
	}

	voteCounts, err := h.queries.GetPostVoteCounts(ctx, postID)
	if err != nil {
		h.logger.Error("Failed to get post vote counts", "error", err, "post_id", postIDStr)
		http.Error(w, "Failed to get post metrics", http.StatusInternalServerError)
		return
	}

	// Calculate total votes
	totalVotes := int32(0)
	if voteCounts.Upvotes != nil {
		totalVotes += *voteCounts.Upvotes
	}
	if voteCounts.Downvotes != nil {
		totalVotes += *voteCounts.Downvotes
	}

	response := PostMetrics{
		Upvotes:      int(*voteCounts.Upvotes),
		Downvotes:    int(*voteCounts.Downvotes),
		Score:        int(*voteCounts.Score),
		CommentCount: int(*commentCount),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetPostModerationState handles GET /api/v1/posts/{id}/moderation
func (h *Handlers) GetPostModerationState(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		http.Error(w, "Invalid post ID", http.StatusBadRequest)
		return
	}
	postIDStr := pathParts[4]

	// For now, return a default moderation state since we don't have the method yet
	response := map[string]interface{}{
		"post_id":      postIDStr,
		"state":        "approved",
		"reason":       nil,
		"moderated_at": time.Now().Format(time.RFC3339),
		"moderated_by": nil,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// formatTimestamptz formats a pgtype.Timestamptz as RFC3339 string
func formatTimestamptz(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return time.Now().Format(time.RFC3339)
	}
	return ts.Time.Format(time.RFC3339)
}

// UpdatePassword handles POST /api/v1/auth/updatePassword
func (h *Handlers) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.CurrentPassword == "" || req.NewPassword == "" {
		http.Error(w, "currentPassword and newPassword required", http.StatusBadRequest)
		return
	}

	// Get the user's token from the Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	// Create request body for PDS
	reqBody, err := json.Marshal(req)
	if err != nil {
		h.logger.Error("Failed to marshal request", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Proxy the request to the PDS server
	pdsURL := h.pdsURL + "/xrpc/com.atproto.server.updatePassword"

	// Create request to PDS server
	pdsReq, err := http.NewRequest("POST", pdsURL, strings.NewReader(string(reqBody)))
	if err != nil {
		h.logger.Error("Failed to create PDS request", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Copy headers from original request
	pdsReq.Header.Set("Content-Type", "application/json")
	pdsReq.Header.Set("Authorization", authHeader)

	// Make request to PDS server
	client := &http.Client{Timeout: 30 * time.Second}
	pdsResp, err := client.Do(pdsReq)
	if err != nil {
		h.logger.Error("Failed to call PDS server", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer pdsResp.Body.Close()

	// Copy response status and headers
	w.WriteHeader(pdsResp.StatusCode)

	// Copy response body
	if pdsResp.StatusCode != http.StatusOK {
		// Read error response from PDS
		body := make([]byte, 1024)
		n, _ := pdsResp.Body.Read(body)
		http.Error(w, string(body[:n]), pdsResp.StatusCode)
		return
	}

	// Success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// proxyToPDS proxies a request to the PDS server
func (h *Handlers) proxyToPDS(r *http.Request, method, path string, body interface{}) ([]byte, error) {
	// Marshal request body if provided
	var reqBody []byte
	var err error
	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, err
		}
	}

	// Create request to PDS
	pdsURL := h.pdsURL + path
	req, err := http.NewRequest(method, pdsURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	// Copy headers from original request
	for key, values := range r.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Set content type for JSON requests
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add service-to-service header
	req.Header.Set("X-Service-Name", "appview")

	// Make request to PDS
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for error status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("PDS request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// waitForEventProcessing polls the AppView database with exponential backoff
// to wait for event processing to complete. This handles the asynchronous
// nature of PDS → Event → AppView data flow.
func (h *Handlers) waitForEventProcessing(ctx context.Context, checkFunc func() error, operation string, identifier string) error {
	const (
		maxRetries = 5
		baseDelay  = 100 * time.Millisecond
		maxDelay   = 1600 * time.Millisecond
	)

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Try the check function
		if err := checkFunc(); err == nil {
			// Success - return immediately
			if attempt > 0 {
				h.logger.Debug("Event processing completed after retry",
					"operation", operation,
					"identifier", identifier,
					"attempt", attempt+1)
			}
			return nil
		}

		// If this was the last attempt, return the error
		if attempt == maxRetries-1 {
			h.logger.Error("Event processing timeout",
				"operation", operation,
				"identifier", identifier,
				"attempts", maxRetries)
			return fmt.Errorf("timeout waiting for %s %s after %d attempts", operation, identifier, maxRetries)
		}

		// Calculate delay with exponential backoff
		delay := baseDelay * time.Duration(1<<uint(attempt)) // 100ms, 200ms, 400ms, 800ms, 1600ms
		if delay > maxDelay {
			delay = maxDelay
		}

		h.logger.Debug("Event processing not ready, retrying",
			"operation", operation,
			"identifier", identifier,
			"attempt", attempt+1,
			"delay_ms", delay.Milliseconds())

		// Wait before next attempt
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	return fmt.Errorf("unexpected end of retry loop")
}
