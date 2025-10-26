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

	// Generate a new UUID for the post
	postID := uuid.New()

	// Create post record for PDS
	postRecord := map[string]interface{}{
		"$type":        "com.hashpost.feed.post",
		"title":        req.Title,
		"content":      req.Content,
		"subforumSlug": req.SubforumSlug,
		"createdAt":    time.Now().Format(time.RFC3339),
	}

	// Proxy to PDS to create the post record
	pdsResponse, err := h.proxyToPDS(r, "POST", "/xrpc/com.atproto.repo.createRecord", map[string]interface{}{
		"repo":       authorDID,
		"collection": "com.hashpost.feed.post",
		"record":     postRecord,
		"rkey":       postID.String(),
	})

	if err != nil {
		h.logger.Error("Failed to create post in PDS", "error", err, "title", req.Title)
		h.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to create post")
		return
	}

	// Parse PDS response to get the created post URI
	var pdsResult struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(pdsResponse, &pdsResult); err != nil {
		h.logger.Error("Failed to parse PDS response", "error", err)
		h.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to parse PDS response")
		return
	}

	// Wait for event processing to complete and get the created post from AppView database
	var createdPost *generated.GetAppViewPostByIDRow
	err = h.waitForEventProcessing(r.Context(), func() error {
		var err error
		createdPost, err = h.queries.GetAppViewPostByID(r.Context(), postID)
		return err
	}, "post creation", postID.String())

	if err != nil {
		h.logger.Error("Failed to get created post from AppView", "error", err, "post_id", postID)
		h.writeError(w, http.StatusInternalServerError, "POST_NOT_FOUND", "Post created but not found in AppView")
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

	// Update user post count
	err = h.queries.IncrementUserPostCount(r.Context(), authorDID)
	if err != nil {
		h.logger.Error("Failed to update user post count", "error", err, "author_did", authorDID)
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

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	h.logger.Debug("UpdatePost path analysis", "path", r.URL.Path, "pathParts", pathParts, "pathPartsLength", len(pathParts))
	if len(pathParts) < 5 {
		h.logger.Error("Invalid post ID path", "path", r.URL.Path, "pathParts", pathParts)
		h.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID")
		return
	}
	postIDStr := pathParts[4]
	h.logger.Debug("Extracted post ID string", "postIDStr", postIDStr)

	// Parse post ID as UUID
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		h.logger.Error("Failed to parse post ID as UUID", "postIDStr", postIDStr, "error", err)
		h.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID format")
		return
	}
	h.logger.Debug("Successfully parsed post ID", "postID", postID)

	// Parse request body
	var req UpdatePostJSONBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("Failed to decode request body", "error", err)
		h.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}
	h.logger.Debug("Successfully decoded request body", "title", req.Title, "contentLength", len(req.Content))

	// Validate required fields
	if req.Title == "" || req.Content == "" {
		h.logger.Error("Missing required fields", "titleEmpty", req.Title == "", "contentEmpty", req.Content == "", "title", req.Title, "contentLength", len(req.Content))
		h.writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "title and content are required")
		return
	}

	ctx := r.Context()

	// Check if post exists and user is the author
	post, err := h.queries.GetAppViewPostByID(ctx, postID)
	if err != nil {
		h.logger.Error("Post not found", "error", err, "post_id", postID)
		h.writeError(w, http.StatusNotFound, "POST_NOT_FOUND", "Post not found")
		return
	}

	if post.AuthorDid != userCtx.Did {
		h.writeError(w, http.StatusForbidden, "FORBIDDEN", "You can only edit your own posts")
		return
	}

	// Extract the collection and rkey from the atproto URI
	// URI format: at://did:plc:user123/collection/rkey
	uriParts := strings.Split(post.AtprotoUri, "/")
	if len(uriParts) < 4 {
		h.logger.Error("Invalid atproto URI format", "uri", post.AtprotoUri)
		h.writeError(w, http.StatusInternalServerError, "INVALID_URI", "Invalid post URI format")
		return
	}
	// The collection is the second-to-last part, rkey is the last part
	collection := uriParts[len(uriParts)-2]
	rkey := uriParts[len(uriParts)-1]

	h.logger.Debug("Extracted collection and rkey", "collection", collection, "rkey", rkey, "uri", post.AtprotoUri)

	// Create updated post record for PDS
	postRecord := map[string]interface{}{
		"$type":        collection, // Use the actual collection from the URI
		"title":        req.Title,
		"text":         req.Content,  // PDS expects "text" field, not "content"
		"subforumSlug": post.SubforumSlug,
		"createdAt":    post.CreatedAt.Time.Format(time.RFC3339),
	}

	// Proxy to PDS to update the post record
	pdsResponse, err := h.proxyToPDS(r, "POST", "/xrpc/com.atproto.repo.putRecord", map[string]interface{}{
		"repo":       userCtx.Did,
		"collection": collection, // Use the actual collection from the URI
		"rkey":       rkey,
		"record":     postRecord,
	})

	if err != nil {
		h.logger.Error("Failed to update post in PDS", "error", err, "post_id", postID)
		h.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to update post")
		return
	}

	// Parse PDS response
	var pdsResult struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(pdsResponse, &pdsResult); err != nil {
		h.logger.Error("Failed to parse PDS response", "error", err)
		h.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to parse PDS response")
		return
	}

	// Directly update the AppView database since event processing is not working
	// TODO: Remove this direct update once NATS event processing is fixed
	h.logger.Info("Directly updating AppView database due to event processing issues", "post_id", postID)
	
	_, err = h.queries.UpdatePostByAtprotoURI(r.Context(), &generated.UpdatePostByAtprotoURIParams{
		AtprotoUri: post.AtprotoUri,
		Title:      req.Title,
		Content:    req.Content,
	})
	if err != nil {
		h.logger.Error("Failed to update post in AppView database", "error", err, "post_id", postID)
		h.writeError(w, http.StatusInternalServerError, "UPDATE_FAILED", "Failed to update post in AppView")
		return
	}

	// Get the updated post from AppView database
	updatedPost, err := h.queries.GetAppViewPostByID(r.Context(), postID)
	if err != nil {
		h.logger.Error("Failed to get updated post from AppView", "error", err, "post_id", postID)
		h.writeError(w, http.StatusInternalServerError, "POST_NOT_FOUND", "Post updated but not found in AppView")
		return
	}

	// Return the updated post
	response := Post{
		Id: openapi_types.UUID(updatedPost.ID),
		Author: Author{
			Did:         updatedPost.AuthorDid,
			Handle:      updatedPost.AuthorHandle,
			DisplayName: updatedPost.AuthorDisplayName,
			AvatarUrl:   updatedPost.AuthorAvatarUrl,
		},
		Subforum:  updatedPost.SubforumSlug,
		Title:     updatedPost.Title,
		Content:   updatedPost.Content,
		CreatedAt: updatedPost.CreatedAt.Time,
		UpdatedAt: &updatedPost.UpdatedAt.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeletePost handles DELETE /api/v1/posts/{id}
func (h *Handlers) DeletePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		h.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		h.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID")
		return
	}
	postIDStr := pathParts[4]
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID format")
		return
	}

	ctx := r.Context()

	// Check if post exists and user is the author
	post, err := h.queries.GetAppViewPostByID(ctx, postID)
	if err != nil {
		h.logger.Error("Post not found", "error", err, "post_id", postID)
		h.writeError(w, http.StatusNotFound, "POST_NOT_FOUND", "Post not found")
		return
	}

	if post.AuthorDid != userCtx.Did {
		h.writeError(w, http.StatusForbidden, "FORBIDDEN", "You can only delete your own posts")
		return
	}

	// Delete post from PDS (canonical record)
	// Extract the rkey from the atproto URI
	// URI format: at://did:plc:user123/com.hashpost.feed.post/uuid-here
	// We need to find the last part after the collection name
	uriParts := strings.Split(post.AtprotoUri, "/")
	if len(uriParts) < 4 {
		h.logger.Error("Invalid atproto URI format", "uri", post.AtprotoUri)
		h.writeError(w, http.StatusInternalServerError, "INVALID_URI", "Invalid post URI format")
		return
	}
	// The rkey is the last part after the collection
	rkey := uriParts[len(uriParts)-1]

	// Call PDS to delete the record
	pdsResponse, err := h.proxyToPDS(r, "POST", "/xrpc/com.atproto.repo.deleteRecord", map[string]interface{}{
		"repo":       userCtx.Did,
		"collection": "com.hashpost.feed.post",
		"rkey":       rkey,
	})

	if err != nil {
		h.logger.Error("Failed to delete post from PDS", "error", err, "post_id", postID)
		h.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to delete post")
		return
	}

	// Log PDS response for debugging
	h.logger.Info("PDS delete response", "response", string(pdsResponse))

	// PDS deletion succeeded, now delete from AppView database directly
	// This ensures the UI updates immediately even if NATS events fail
	err = h.queries.DeleteAppViewPost(ctx, postID)
	if err != nil {
		h.logger.Error("Failed to delete post from AppView", "error", err, "post_id", postID)
		// Don't fail the request since PDS deletion succeeded
	}

	h.logger.Info("Post deleted", "post_id", postID, "user_did", userCtx.Did)
	w.WriteHeader(http.StatusNoContent)
}
