package appview

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
)

// ProfileHandlers handles profile-related HTTP requests
type ProfileHandlers struct {
	queries generated.Querier
	logger  *slog.Logger
}

// NewProfileHandlers creates a new profile handlers instance
func NewProfileHandlers(queries generated.Querier, logger *slog.Logger) *ProfileHandlers {
	return &ProfileHandlers{
		queries: queries,
		logger:  logger,
	}
}

// GetUserProfile handles GET /api/v1/profiles/@{handle}
func (h *ProfileHandlers) GetUserProfile(w http.ResponseWriter, r *http.Request, handle string) {
	ctx := r.Context()

	// Remove @ prefix if present
	handle = strings.TrimPrefix(handle, "@")

	h.logger.Debug("Getting user profile", "handle", handle)

	// Get user by handle
	user, err := h.queries.GetUserByHandle(ctx, handle)
	if err != nil {
		h.logger.Error("Failed to get user by handle", "error", err, "handle", handle)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get user context for debugging
	userCtx := GetUserContext(r)
	h.logger.Debug("Profile visibility check",
		"profile_visibility", user.ProfileVisibility,
		"profile_owner_did", user.Did,
		"requesting_user_did", func() string {
			if userCtx != nil {
				return userCtx.Did
			}
			return "nil"
		}(),
		"requesting_user_handle", func() string {
			if userCtx != nil {
				return userCtx.Handle
			}
			return "nil"
		}(),
	)

	// Check profile visibility
	if !h.canViewProfile(r, string(user.ProfileVisibility), user.Did) {
		h.logger.Warn("Profile access denied",
			"profile_visibility", user.ProfileVisibility,
			"profile_owner_did", user.Did,
			"requesting_user_did", func() string {
				if userCtx != nil {
					return userCtx.Did
				}
				return "nil"
			}(),
		)
		if user.ProfileVisibility == "private" {
			http.Error(w, "Profile is private", http.StatusForbidden)
		} else if user.ProfileVisibility == "authenticated" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
		} else {
			http.Error(w, "Access denied", http.StatusForbidden)
		}
		return
	}

	// Convert to API response format
	profile := map[string]interface{}{
		"did":               user.Did,
		"handle":            user.Handle,
		"displayName":       user.DisplayName,
		"bio":               user.Bio,
		"avatarUrl":         user.AvatarUrl,
		"postCount":         user.PostCount,
		"commentCount":      user.CommentCount,
		"reputation":        user.Reputation,
		"profileVisibility": user.ProfileVisibility,
		"pdsSource":         user.PdsSource,
		"createdAt":         user.CreatedAt.Time.Format(time.RFC3339),
		"updatedAt":         user.UpdatedAt.Time.Format(time.RFC3339),
		"lastSeenAt":        nil,
	}

	if user.LastSeenAt.Valid {
		profile["lastSeenAt"] = user.LastSeenAt.Time.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

// GetUserPosts handles GET /api/v1/profiles/@{handle}/posts
func (h *ProfileHandlers) GetUserPosts(w http.ResponseWriter, r *http.Request, handle string) {
	ctx := r.Context()

	// Remove @ prefix if present
	handle = strings.TrimPrefix(handle, "@")

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

	h.logger.Debug("Getting user posts", "handle", handle, "limit", limit, "offset", offset)

	// First get the user to check visibility
	user, err := h.queries.GetUserByHandle(ctx, handle)
	if err != nil {
		h.logger.Error("Failed to get user by handle", "error", err, "handle", handle)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check profile visibility
	if !h.canViewProfile(r, string(user.ProfileVisibility), user.Did) {
		if user.ProfileVisibility == "private" {
			http.Error(w, "Profile is private", http.StatusForbidden)
		} else if user.ProfileVisibility == "authenticated" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
		} else {
			http.Error(w, "Access denied", http.StatusForbidden)
		}
		return
	}

	// Get user's posts
	posts, err := h.queries.GetPostsByAuthorDID(ctx, &generated.GetPostsByAuthorDIDParams{
		AuthorDid: user.Did,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		h.logger.Error("Failed to get user posts", "error", err, "handle", handle)
		http.Error(w, "Failed to get posts", http.StatusInternalServerError)
		return
	}

	// Convert to API response format
	userPosts := make([]map[string]interface{}, len(posts))
	for i, post := range posts {
		userPosts[i] = map[string]interface{}{
			"id":           post.ID,
			"atprotoUri":   post.AtprotoUri,
			"title":        post.Title,
			"content":      post.Content,
			"authorDid":    post.AuthorDid,
			"authorHandle": post.AuthorHandle,
			"subforumSlug": post.SubforumSlug,
			"subforumName": post.SubforumName,
			"createdAt":    post.CreatedAt.Time.Format(time.RFC3339),
			"updatedAt":    post.UpdatedAt.Time.Format(time.RFC3339),
			"upvotes":      post.Upvotes,
			"downvotes":    post.Downvotes,
			"commentCount": post.CommentCount,
			"score":        post.Score,
		}
	}

	response := map[string]interface{}{
		"posts":  userPosts,
		"total":  len(posts), // TODO: Get actual total count
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetUserComments handles GET /api/v1/profiles/@{handle}/comments
func (h *ProfileHandlers) GetUserComments(w http.ResponseWriter, r *http.Request, handle string) {
	ctx := r.Context()

	// Remove @ prefix if present
	handle = strings.TrimPrefix(handle, "@")

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

	h.logger.Debug("Getting user comments", "handle", handle, "limit", limit, "offset", offset)

	// First get the user to check visibility
	user, err := h.queries.GetUserByHandle(ctx, handle)
	if err != nil {
		h.logger.Error("Failed to get user by handle", "error", err, "handle", handle)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check profile visibility
	if !h.canViewProfile(r, string(user.ProfileVisibility), user.Did) {
		if user.ProfileVisibility == "private" {
			http.Error(w, "Profile is private", http.StatusForbidden)
		} else if user.ProfileVisibility == "authenticated" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
		} else {
			http.Error(w, "Access denied", http.StatusForbidden)
		}
		return
	}

	// Get user's comments
	comments, err := h.queries.ListCommentsByAuthor(ctx, &generated.ListCommentsByAuthorParams{
		AuthorDid: user.Did,
		Limit:     int32(limit),
		Offset:    int32(offset),
	})
	if err != nil {
		h.logger.Error("Failed to get user comments", "error", err, "handle", handle)
		http.Error(w, "Failed to get comments", http.StatusInternalServerError)
		return
	}

	// Convert to API response format
	userComments := make([]map[string]interface{}, len(comments))
	for i, comment := range comments {
		userComments[i] = map[string]interface{}{
			"id":                comment.ID,
			"atprotoUri":        comment.AtprotoUri,
			"content":           comment.Content,
			"authorDid":         comment.AuthorDid,
			"authorHandle":      comment.AuthorHandle,
			"authorDisplayName": comment.AuthorDisplayName,
			"authorAvatarUrl":   comment.AuthorAvatarUrl,
			"postId":            comment.PostID,
			"postTitle":         comment.PostTitle,
			"subforumSlug":      comment.SubforumSlug,
			"subforumName":      nil, // Not available in this query
			"parentId":          comment.ParentID,
			"createdAt":         comment.CreatedAt.Time.Format(time.RFC3339),
			"updatedAt":         comment.UpdatedAt.Time.Format(time.RFC3339),
			"upvotes":           comment.Upvotes,
			"downvotes":         comment.Downvotes,
			"score":             comment.Score,
		}
	}

	response := map[string]interface{}{
		"comments": userComments,
		"total":    len(comments), // TODO: Get actual total count
		"limit":    limit,
		"offset":   offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateProfileVisibility handles PATCH /api/v1/profiles/me/visibility
func (h *ProfileHandlers) UpdateProfileVisibility(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user context from middleware (authentication already handled by middleware)
	userCtx := GetUserContext(r)
	if userCtx == nil {
		http.Error(w, "Authentication required", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var request struct {
		Visibility string `json:"visibility"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate visibility setting
	if request.Visibility != "public" && request.Visibility != "authenticated" && request.Visibility != "private" {
		http.Error(w, "Invalid visibility setting", http.StatusBadRequest)
		return
	}

	h.logger.Debug("Updating profile visibility", "user_did", userCtx.Did, "visibility", request.Visibility)

	// Update profile visibility
	updatedUser, err := h.queries.UpdateUserProfileVisibility(ctx, &generated.UpdateUserProfileVisibilityParams{
		Did:               userCtx.Did,
		ProfileVisibility: generated.ProfileVisibility(request.Visibility),
	})
	if err != nil {
		h.logger.Error("Failed to update profile visibility", "error", err, "user_did", userCtx.Did)
		http.Error(w, "Failed to update profile visibility", http.StatusInternalServerError)
		return
	}

	// Convert to API response format
	profile := map[string]interface{}{
		"did":               updatedUser.Did,
		"handle":            updatedUser.Handle,
		"displayName":       updatedUser.DisplayName,
		"bio":               updatedUser.Bio,
		"avatarUrl":         updatedUser.AvatarUrl,
		"postCount":         updatedUser.PostCount,
		"commentCount":      updatedUser.CommentCount,
		"reputation":        updatedUser.Reputation,
		"profileVisibility": updatedUser.ProfileVisibility,
		"pdsSource":         updatedUser.PdsSource,
		"createdAt":         updatedUser.CreatedAt.Time.Format(time.RFC3339),
		"updatedAt":         updatedUser.UpdatedAt.Time.Format(time.RFC3339),
		"lastSeenAt":        nil,
	}

	if updatedUser.LastSeenAt.Valid {
		profile["lastSeenAt"] = updatedUser.LastSeenAt.Time.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profile)
}

// canViewProfile checks if the current request can view a profile based on visibility settings
func (h *ProfileHandlers) canViewProfile(r *http.Request, visibility string, profileOwnerDID string) bool {
	switch visibility {
	case "public":
		return true
	case "authenticated":
		// Check if user is authenticated
		userCtx := GetUserContext(r)
		return userCtx != nil
	case "private":
		// Only the profile owner can view
		userCtx := GetUserContext(r)
		return userCtx != nil && userCtx.Did == profileOwnerDID
	default:
		return false
	}
}
