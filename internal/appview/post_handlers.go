package appview

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// PostsHandler handles both GET and POST /api/v1/posts
func (h *Handlers) PostsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListPosts(w, r)
	case http.MethodPost:
		h.CreatePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ListPosts handles GET /api/v1/posts
func (h *Handlers) ListPosts(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	limit := 20
	offset := 0
	subforum := r.URL.Query().Get("subforum")

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

	h.logger.Debug("Listing posts", "limit", limit, "offset", offset, "subforum", subforum)

	// For now, return mock data
	response := map[string]interface{}{
		"posts": []map[string]interface{}{
			{
				"id":            "550e8400-e29b-41d4-a716-446655440002",
				"author":        "550e8400-e29b-41d4-a716-446655440001",
				"subforum":      "550e8400-e29b-41d4-a716-446655440000",
				"title":         "Welcome to HashPost!",
				"content":       "This is the first post in our new HashPost community. Welcome everyone!",
				"created_at":    time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				"updated_at":    time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
				"upvotes":       15,
				"downvotes":     2,
				"comment_count": 8,
				"is_pinned":     true,
				"is_locked":     false,
			},
		},
		"total":  1,
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// PostByIDHandler handles GET, PUT, DELETE /api/v1/posts/{id}
func (h *Handlers) PostByIDHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetPostByID(w, r)
	case http.MethodPut:
		h.UpdatePost(w, r)
	case http.MethodDelete:
		h.DeletePost(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetPostByID handles GET /api/v1/posts/{id}
func (h *Handlers) GetPostByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/posts/")
	if path == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Missing post ID")
		return
	}

	h.logger.Debug("Getting post by ID", "id", path)

	// For now, return mock data
	response := map[string]interface{}{
		"id":            path,
		"author":        "550e8400-e29b-41d4-a716-446655440001",
		"subforum":      "550e8400-e29b-41d4-a716-446655440000",
		"title":         "Welcome to HashPost!",
		"content":       "This is the first post in our new HashPost community. Welcome everyone!",
		"created_at":    time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		"updated_at":    time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		"upvotes":       15,
		"downvotes":     2,
		"comment_count": 8,
		"is_pinned":     true,
		"is_locked":     false,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreatePost handles POST /api/v1/posts
func (h *Handlers) CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var req CreatePostJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Title == "" || req.Content == "" || req.SubforumSlug == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title, content, and subforum_slug are required")
		return
	}

	// For now, use placeholder values for author
	// In a real implementation, these would come from authentication
	authorDID := "did:placeholder:author"

	// Create post in database
	// Note: This would need to be implemented with the actual database operations
	// For now, we'll return a mock response
	post := map[string]interface{}{
		"id":            "placeholder-id",
		"title":         req.Title,
		"content":       req.Content,
		"subforum":      req.SubforumSlug,
		"author":        authorDID,
		"created_at":    time.Now().Format(time.RFC3339),
		"upvotes":       0,
		"downvotes":     0,
		"comment_count": 0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

// UpdatePost handles PUT /api/v1/posts/{id}
func (h *Handlers) UpdatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID")
		return
	}
	postID := pathParts[3]

	// Parse request body
	var req UpdatePostJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Title == "" || req.Content == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title and content are required")
		return
	}

	// For now, return a mock response
	// In a real implementation, this would update the database
	post := map[string]interface{}{
		"id":            postID,
		"title":         req.Title,
		"content":       req.Content,
		"updated_at":    time.Now().Format(time.RFC3339),
		"upvotes":       0,
		"downvotes":     0,
		"comment_count": 0,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(post)
}

// DeletePost handles DELETE /api/v1/posts/{id}
func (h *Handlers) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID")
		return
	}
	postID := pathParts[3]

	// For now, just return success
	// In a real implementation, this would delete from the database
	h.logger.Info("Post deleted", "post_id", postID)

	w.WriteHeader(http.StatusNoContent)
}
