package appview

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

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

	// Check if subforum exists
	_, err := sh.queries.GetSubforumBySlug(ctx, slug)
	if err != nil {
		sh.logger.Error("Subforum not found", "error", err, "slug", slug)
		sh.writeError(w, http.StatusNotFound, "SUBFORUM_NOT_FOUND", "Subforum not found")
		return
	}

	// Create subscription
	subscription, err := sh.queries.CreateSubscription(ctx, &generated.CreateSubscriptionParams{
		UserDid:      userCtx.Did,
		UserHandle:   userCtx.Handle,
		SubforumSlug: slug,
	})
	if err != nil {
		sh.logger.Error("Failed to create subscription", "error", err, "user_did", userCtx.Did, "slug", slug)
		sh.writeError(w, http.StatusInternalServerError, "SUBSCRIPTION_FAILED", "Failed to create subscription")
		return
	}

	// Update subscriber count
	err = sh.queries.UpdateSubforumSubscriberCount(ctx, slug)
	if err != nil {
		sh.logger.Error("Failed to update subscriber count", "error", err, "slug", slug)
		// Don't fail the request, just log the error
	}

	response := SubscriptionResponse{
		SubforumSlug: subscription.SubforumSlug,
		SubscribedAt: subscription.CreatedAt.Time,
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
