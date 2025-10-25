package appview

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
)

// SubscriptionHandlers handles subscription operations for subforums
type SubscriptionHandlers struct {
	queries *generated.Queries
	logger  *slog.Logger
}

// NewSubscriptionHandlers creates a new subscription handlers instance
func NewSubscriptionHandlers(queries *generated.Queries, logger *slog.Logger) *SubscriptionHandlers {
	return &SubscriptionHandlers{
		queries: queries,
		logger:  logger,
	}
}

// SubscribeToSubforum handles POST /api/v1/subforums/{slug}/subscribe
func (sh *SubscriptionHandlers) SubscribeToSubforum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract subforum slug from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		sh.writeError(w, http.StatusBadRequest, "INVALID_SUBFORUM_SLUG", "Invalid subforum slug")
		return
	}
	slug := pathParts[3]

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		sh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := r.Context()

	// Check if subforum exists and get its atproto URI
	subforum, err := sh.queries.GetSubforumBySlug(ctx, slug)
	if err != nil {
		sh.logger.Error("Subforum not found", "error", err, "slug", slug)
		sh.writeError(w, http.StatusNotFound, "SUBFORUM_NOT_FOUND", "Subforum not found")
		return
	}

	// Create subscription record in PDS
	subscriptionRecord := map[string]interface{}{
		"$type":     "com.hashpost.graph.subscription",
		"subject":   subforum.AtprotoUri,
		"createdAt": time.Now().Format(time.RFC3339),
	}

	// Proxy to PDS to create the subscription record
	pdsResponse, err := sh.proxyToPDS(r, "POST", "/xrpc/com.atproto.repo.createRecord", map[string]interface{}{
		"repo":       userCtx.Did,
		"collection": "com.hashpost.graph.subscription",
		"record":     subscriptionRecord,
	})

	if err != nil {
		sh.logger.Error("Failed to create subscription in PDS", "error", err)
		sh.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to create subscription")
		return
	}

	// Parse PDS response to get the created subscription URI
	var pdsResult struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(pdsResponse, &pdsResult); err != nil {
		sh.logger.Error("Failed to parse PDS response", "error", err)
		sh.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to parse PDS response")
		return
	}

	// Wait a moment for the event to be processed by AppView
	time.Sleep(100 * time.Millisecond)

	response := SubscriptionResponse{
		SubforumSlug: slug,
		SubscribedAt: time.Now(),
	}

	sh.writeJSON(w, http.StatusCreated, response)
}

// UnsubscribeFromSubforum handles DELETE /api/v1/subforums/{slug}/subscribe
func (sh *SubscriptionHandlers) UnsubscribeFromSubforum(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract subforum slug from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		sh.writeError(w, http.StatusBadRequest, "INVALID_SUBFORUM_SLUG", "Invalid subforum slug")
		return
	}
	slug := pathParts[3]

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		sh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := r.Context()

	// Delete subscription
	err := sh.queries.DeleteSubscription(ctx, &generated.DeleteSubscriptionParams{
		UserDid:      userCtx.Did,
		SubforumSlug: slug,
	})
	if err != nil {
		sh.logger.Error("Failed to delete subscription", "error", err, "user_did", userCtx.Did, "slug", slug)
		sh.writeError(w, http.StatusInternalServerError, "UNSUBSCRIPTION_FAILED", "Failed to delete subscription")
		return
	}

	// Update subscriber count
	err = sh.queries.UpdateSubforumSubscriberCount(ctx, slug)
	if err != nil {
		sh.logger.Error("Failed to update subscriber count", "error", err, "slug", slug)
		// Don't fail the request, just log the error
	}

	sh.writeJSON(w, http.StatusOK, map[string]string{"message": "Unsubscribed successfully"})
}

// ListMySubscriptionsWithParams handles GET /api/v1/me/subscriptions with parsed parameters
func (sh *SubscriptionHandlers) ListMySubscriptionsWithParams(w http.ResponseWriter, r *http.Request, limit int, offset int) {
	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		sh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := r.Context()

	// Get user subscriptions
	subscriptions, err := sh.queries.ListUserSubscriptions(ctx, &generated.ListUserSubscriptionsParams{
		UserDid: userCtx.Did,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		sh.logger.Error("Failed to list user subscriptions", "error", err, "user_did", userCtx.Did)
		sh.writeError(w, http.StatusInternalServerError, "SUBSCRIPTION_LIST_FAILED", "Failed to list subscriptions")
		return
	}

	// Get total count
	total, err := sh.queries.CountUserSubscriptions(ctx, userCtx.Did)
	if err != nil {
		sh.logger.Error("Failed to count user subscriptions", "error", err, "user_did", userCtx.Did)
		sh.writeError(w, http.StatusInternalServerError, "SUBSCRIPTION_COUNT_FAILED", "Failed to count subscriptions")
		return
	}

	// Convert to response format
	subscriptionResponses := make([]SubscriptionResponse, len(subscriptions))
	for i, sub := range subscriptions {
		subscriptionResponses[i] = SubscriptionResponse{
			SubforumSlug: sub.SubforumSlug,
			SubforumName: &sub.SubforumName,
			SubscribedAt: sub.CreatedAt.Time,
		}
	}

	response := SubscriptionListResponse{
		Subscriptions: subscriptionResponses,
		Total:         int(total),
		Limit:         limit,
		Offset:        offset,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListMySubscriptions handles GET /api/v1/me/subscriptions
func (sh *SubscriptionHandlers) ListMySubscriptions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		sh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Parse pagination parameters
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

	ctx := r.Context()

	// Get user subscriptions
	subscriptions, err := sh.queries.ListUserSubscriptions(ctx, &generated.ListUserSubscriptionsParams{
		UserDid: userCtx.Did,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		sh.logger.Error("Failed to list user subscriptions", "error", err, "user_did", userCtx.Did)
		sh.writeError(w, http.StatusInternalServerError, "SUBSCRIPTION_LIST_FAILED", "Failed to list subscriptions")
		return
	}

	// Get total count
	total, err := sh.queries.CountUserSubscriptions(ctx, userCtx.Did)
	if err != nil {
		sh.logger.Error("Failed to count user subscriptions", "error", err, "user_did", userCtx.Did)
		sh.writeError(w, http.StatusInternalServerError, "SUBSCRIPTION_COUNT_FAILED", "Failed to count subscriptions")
		return
	}

	// Convert to response format
	subscriptionResponses := make([]SubscriptionResponse, len(subscriptions))
	for i, sub := range subscriptions {
		subscriptionResponses[i] = SubscriptionResponse{
			SubforumSlug: sub.SubforumSlug,
			SubforumName: &sub.SubforumName,
			SubscribedAt: sub.CreatedAt.Time,
		}
	}

	response := SubscriptionListResponse{
		Subscriptions: subscriptionResponses,
		Total:         int(total),
		Limit:         limit,
		Offset:        offset,
	}

	sh.writeJSON(w, http.StatusOK, response)
}

// GetUserSubscription handles GET /api/v1/subforums/{slug}/subscribe
func (sh *SubscriptionHandlers) GetUserSubscription(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract subforum slug from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		sh.writeError(w, http.StatusBadRequest, "INVALID_SUBFORUM_SLUG", "Invalid subforum slug")
		return
	}
	slug := pathParts[3]

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		sh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := r.Context()

	// Check if user is subscribed
	subscription, err := sh.queries.GetUserSubscription(ctx, &generated.GetUserSubscriptionParams{
		UserDid:      userCtx.Did,
		SubforumSlug: slug,
	})
	if err != nil {
		// User is not subscribed
		sh.writeJSON(w, http.StatusOK, map[string]interface{}{
			"subscribed": false,
		})
		return
	}

	response := map[string]interface{}{
		"subscribed":    true,
		"subforum_slug": subscription.SubforumSlug,
		"subscribed_at": subscription.CreatedAt.Time.Format("2006-01-02T15:04:05.000Z"),
	}

	sh.writeJSON(w, http.StatusOK, response)
}

// ListSubforumSubscribers handles GET /api/v1/subforums/{slug}/subscribers
func (sh *SubscriptionHandlers) ListSubforumSubscribers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract subforum slug from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		sh.writeError(w, http.StatusBadRequest, "INVALID_SUBFORUM_SLUG", "Invalid subforum slug")
		return
	}
	slug := pathParts[3]

	// Parse pagination parameters
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

	ctx := r.Context()

	// Get subforum subscribers
	subscribers, err := sh.queries.ListSubforumSubscribers(ctx, &generated.ListSubforumSubscribersParams{
		SubforumSlug: slug,
		Limit:        int32(limit),
		Offset:       int32(offset),
	})
	if err != nil {
		sh.logger.Error("Failed to list subforum subscribers", "error", err, "slug", slug)
		sh.writeError(w, http.StatusInternalServerError, "SUBSCRIBER_LIST_FAILED", "Failed to list subscribers")
		return
	}

	// Get total count
	total, err := sh.queries.CountSubforumSubscribers(ctx, slug)
	if err != nil {
		sh.logger.Error("Failed to count subforum subscribers", "error", err, "slug", slug)
		sh.writeError(w, http.StatusInternalServerError, "SUBSCRIBER_COUNT_FAILED", "Failed to count subscribers")
		return
	}

	// Convert to response format
	subscriberResponses := make([]map[string]interface{}, len(subscribers))
	for i, sub := range subscribers {
		subscriberResponses[i] = map[string]interface{}{
			"user_did":      sub.UserDid,
			"user_handle":   sub.UserHandle,
			"subscribed_at": sub.SubscribedAt.Time.Format("2006-01-02T15:04:05.000Z"),
		}
	}

	response := map[string]interface{}{
		"subscribers": subscriberResponses,
		"total":       int(total),
		"limit":       limit,
		"offset":      offset,
	}

	sh.writeJSON(w, http.StatusOK, response)
}

// ListMyModeratedSubforums handles GET /api/v1/me/moderated-subforums
func (sh *SubscriptionHandlers) ListMyModeratedSubforums(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		sh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := r.Context()

	// Get moderated subforums
	subforums, err := sh.queries.GetModeratedSubforums(ctx, userCtx.Did)
	if err != nil {
		sh.logger.Error("Failed to list moderated subforums", "error", err, "user_did", userCtx.Did)
		sh.writeError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list moderated subforums")
		return
	}

	// Convert to response format
	response := make([]map[string]interface{}, len(subforums))
	for i, sf := range subforums {
		response[i] = map[string]interface{}{
			"id":                sf.ID,
			"name":              sf.Name,
			"slug":              sf.Slug,
			"description":       sf.Description,
			"created_by":        sf.CreatedByDid,
			"created_by_handle": sf.CreatedByHandle,
			"created_at":        sf.CreatedAt,
			"subscriber_count":  sf.SubscriberCount,
			"post_count":        sf.PostCount,
			"comment_count":     sf.CommentCount,
			"prefix_type":       sf.PrefixType,
		}
	}

	sh.writeJSON(w, http.StatusOK, response)
}

// Helper methods

func (sh *SubscriptionHandlers) writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}

func (sh *SubscriptionHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// proxyToPDS proxies a request to the PDS server
func (sh *SubscriptionHandlers) proxyToPDS(r *http.Request, method, path string, body interface{}) ([]byte, error) {
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
	pdsURL := "http://hashpost-pds:8080" + path
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
