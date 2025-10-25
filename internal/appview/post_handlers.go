package appview

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
	openapi_types "github.com/oapi-codegen/runtime/types"
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

// ListPostsWithParams handles GET /api/v1/posts with parsed parameters
func (h *Handlers) ListPostsWithParams(w http.ResponseWriter, r *http.Request, limit int, offset int, subforum string) {
	h.logger.Debug("Listing posts", "limit", limit, "offset", offset, "subforum", subforum)

	// Get posts from database
	var posts interface{}
	var err error

	if subforum != "" {
		// Parse subforum parameter (format: "h/hashpost")
		subforumParts := strings.Split(subforum, "/")
		if len(subforumParts) == 2 && subforumParts[0] == "h" {
			subforumSlug := subforumParts[1]
			posts, err = h.queries.ListAppViewPostsBySubforum(r.Context(), &generated.ListAppViewPostsBySubforumParams{
				SubforumSlug: subforumSlug,
				Limit:        int32(limit),
				Offset:       int32(offset),
			})
		} else {
			h.writeError(w, http.StatusBadRequest, "INVALID_SUBFORUM", "Invalid subforum format")
			return
		}
	} else {
		posts, err = h.queries.ListAppViewPosts(r.Context(), &generated.ListAppViewPostsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
	}

	if err != nil {
		h.logger.Error("Failed to list posts", "error", err)
		h.writeError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list posts")
		return
	}

	// Convert posts to response format using proper Post types
	var postList []Post

	// Handle different post types
	switch p := posts.(type) {
	case []*generated.ListAppViewPostsRow:
		postList = make([]Post, len(p))
		for i, post := range p {
			postList[i] = Post{
				Id: openapi_types.UUID(post.ID),
				Author: Author{
					Did:         post.AuthorDid,
					Handle:      post.AuthorHandle,
					DisplayName: post.AuthorDisplayName,
					AvatarUrl:   post.AuthorAvatarUrl,
				},
				Subforum:  post.SubforumSlug,
				Title:     post.Title,
				Content:   post.Content,
				CreatedAt: post.CreatedAt.Time,
				UpdatedAt: &post.UpdatedAt.Time,
			}
		}
	case []*generated.ListAppViewPostsBySubforumRow:
		postList = make([]Post, len(p))
		for i, post := range p {
			postList[i] = Post{
				Id: openapi_types.UUID(post.ID),
				Author: Author{
					Did:         post.AuthorDid,
					Handle:      post.AuthorHandle,
					DisplayName: post.AuthorDisplayName,
					AvatarUrl:   post.AuthorAvatarUrl,
				},
				Subforum:  post.SubforumSlug,
				Title:     post.Title,
				Content:   post.Content,
				CreatedAt: post.CreatedAt.Time,
				UpdatedAt: &post.UpdatedAt.Time,
			}
		}
	default:
		h.logger.Error("Unknown post type", "type", fmt.Sprintf("%T", posts))
		h.writeError(w, http.StatusInternalServerError, "UNKNOWN_TYPE", "Unknown post type")
		return
	}

	response := map[string]interface{}{
		"posts":  postList,
		"total":  len(postList),
		"limit":  limit,
		"offset": offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
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

	// Get posts from database
	var posts interface{}
	var err error

	if subforum != "" {
		// Parse subforum parameter (format: "h/hashpost")
		subforumParts := strings.Split(subforum, "/")
		if len(subforumParts) == 2 && subforumParts[0] == "h" {
			subforumSlug := subforumParts[1]
			posts, err = h.queries.ListAppViewPostsBySubforum(r.Context(), &generated.ListAppViewPostsBySubforumParams{
				SubforumSlug: subforumSlug,
				Limit:        int32(limit),
				Offset:       int32(offset),
			})
		} else {
			h.writeError(w, http.StatusBadRequest, "INVALID_SUBFORUM", "Invalid subforum format")
			return
		}
	} else {
		posts, err = h.queries.ListAppViewPosts(r.Context(), &generated.ListAppViewPostsParams{
			Limit:  int32(limit),
			Offset: int32(offset),
		})
	}

	if err != nil {
		h.logger.Error("Failed to list posts", "error", err)
		h.writeError(w, http.StatusInternalServerError, "LIST_FAILED", "Failed to list posts")
		return
	}

	// Convert posts to response format using proper Post types
	var postList []Post

	// Handle different post types
	switch p := posts.(type) {
	case []*generated.ListAppViewPostsRow:
		postList = make([]Post, len(p))
		for i, post := range p {
			postList[i] = Post{
				Id: openapi_types.UUID(post.ID),
				Author: Author{
					Did:         post.AuthorDid,
					Handle:      post.AuthorHandle,
					DisplayName: post.AuthorDisplayName,
					AvatarUrl:   post.AuthorAvatarUrl,
				},
				Subforum:  post.SubforumSlug,
				Title:     post.Title,
				Content:   post.Content,
				CreatedAt: post.CreatedAt.Time,
				UpdatedAt: &post.UpdatedAt.Time,
			}
		}
	case []*generated.ListAppViewPostsBySubforumRow:
		postList = make([]Post, len(p))
		for i, post := range p {
			postList[i] = Post{
				Id: openapi_types.UUID(post.ID),
				Author: Author{
					Did:         post.AuthorDid,
					Handle:      post.AuthorHandle,
					DisplayName: post.AuthorDisplayName,
					AvatarUrl:   post.AuthorAvatarUrl,
				},
				Subforum:  post.SubforumSlug,
				Title:     post.Title,
				Content:   post.Content,
				CreatedAt: post.CreatedAt.Time,
				UpdatedAt: &post.UpdatedAt.Time,
			}
		}
	default:
		h.logger.Error("Unknown post type", "type", fmt.Sprintf("%T", posts))
		h.writeError(w, http.StatusInternalServerError, "UNKNOWN_TYPE", "Unknown post type")
		return
	}

	response := map[string]interface{}{
		"posts":  postList,
		"total":  len(postList),
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

	// Parse post ID as UUID
	postID, err := uuid.Parse(path)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID format")
		return
	}

	// Get post from database
	post, err := h.queries.GetAppViewPostByID(r.Context(), postID)
	if err != nil {
		h.logger.Error("Failed to get post", "error", err, "post_id", path)
		h.writeError(w, http.StatusNotFound, "POST_NOT_FOUND", "Post not found")
		return
	}

	// Return the post data using proper generated types
	response := Post{
		Id: openapi_types.UUID(post.ID),
		Author: Author{
			Did:         post.AuthorDid,
			Handle:      post.AuthorHandle,
			DisplayName: post.AuthorDisplayName,
			AvatarUrl:   post.AuthorAvatarUrl,
		},
		Subforum:  post.SubforumSlug,
		Title:     post.Title,
		Content:   post.Content,
		CreatedAt: post.CreatedAt.Time,
		UpdatedAt: &post.UpdatedAt.Time,
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

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
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

	// Use authenticated user's DID as author
	authorDID := userCtx.Did
	authorHandle := userCtx.Handle

	// Generate a new UUID for the post
	postID := uuid.New()

	// Create atproto URI for the post
	atprotoURI := fmt.Sprintf("at://%s/app.bsky.feed.post/%s", authorDID, postID.String())

	// Create post in database
	createdPost, err := h.queries.CreateAppViewPost(r.Context(), &generated.CreateAppViewPostParams{
		AtprotoUri:   atprotoURI,
		AuthorDid:    authorDID,
		AuthorHandle: authorHandle,
		SubforumSlug: req.SubforumSlug,
		Title:        req.Title,
		Content:      req.Content,
	})

	if err != nil {
		h.logger.Error("Failed to create post", "error", err, "title", req.Title)
		h.writeError(w, http.StatusInternalServerError, "CREATE_FAILED", "Failed to create post")
		return
	}

	// Update subforum post count
	err = h.queries.UpdateSubforumPostCount(r.Context(), &generated.UpdateSubforumPostCountParams{
		Slug:      req.SubforumSlug,
		PostCount: &[]int32{1}[0],
	})
	if err != nil {
		h.logger.Error("Failed to update subforum post count", "error", err, "subforum", req.SubforumSlug)
		// Don't fail the request, just log the error
	}

	// Return the created post
	post := map[string]interface{}{
		"id":            createdPost.ID.String(),
		"title":         createdPost.Title,
		"content":       createdPost.Content,
		"subforum":      createdPost.SubforumSlug,
		"author":        createdPost.AuthorDid,
		"author_handle": createdPost.AuthorHandle,
		"created_at":    createdPost.CreatedAt.Time.Format(time.RFC3339),
		"upvotes":       *createdPost.Upvotes,
		"downvotes":     *createdPost.Downvotes,
		"comment_count": *createdPost.CommentCount,
		"score":         *createdPost.Score,
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
