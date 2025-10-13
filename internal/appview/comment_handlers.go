package appview

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
)

// CommentsHandler handles GET and POST /api/v1/comments
func (h *Handlers) CommentsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.ListComments(w, r)
	case http.MethodPost:
		h.CreateComment(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// CommentByIDHandler handles GET, PUT, DELETE /api/v1/comments/{id}
func (h *Handlers) CommentByIDHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// For now, just return method not allowed for GET comments by ID
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	case http.MethodPut:
		h.UpdateComment(w, r)
	case http.MethodDelete:
		h.DeleteComment(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// ListComments handles GET /api/v1/comments
func (h *Handlers) ListComments(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	postIDStr := r.URL.Query().Get("post_id")
	if postIDStr == "" {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Missing post_id parameter")
		return
	}

	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_post_id", "Invalid post_id format")
		return
	}

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

	h.logger.Debug("Listing comments", "post_id", postID, "limit", limit, "offset", offset)

	ctx := r.Context()

	// Get comments from database
	comments, err := h.queries.ListCommentsByPost(ctx, &generated.ListCommentsByPostParams{
		PostID: pgtype.UUID{Bytes: postID, Valid: true},
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		h.logger.Error("Failed to list comments", "error", err, "post_id", postID)
		h.writeError(w, http.StatusInternalServerError, "COMMENT_LIST_FAILED", "Failed to list comments")
		return
	}

	// Get total count
	total, err := h.queries.CountCommentsByPost(ctx, pgtype.UUID{Bytes: postID, Valid: true})
	if err != nil {
		h.logger.Error("Failed to count comments", "error", err, "post_id", postID)
		h.writeError(w, http.StatusInternalServerError, "COMMENT_COUNT_FAILED", "Failed to count comments")
		return
	}

	// Convert to response format
	commentResponses := make([]map[string]interface{}, len(comments))
	for i, comment := range comments {
		commentResponses[i] = map[string]interface{}{
			"id":         comment.ID.String(),
			"author":     comment.AuthorDid,
			"post":       comment.PostID.String(),
			"parent":     comment.ParentID.String(),
			"content":    comment.Content,
			"created_at": comment.CreatedAt.Time.Format(time.RFC3339),
			"updated_at": comment.UpdatedAt.Time.Format(time.RFC3339),
			"upvotes":    int(*comment.Upvotes),
			"downvotes":  int(*comment.Downvotes),
			"score":      int(*comment.Score),
		}
	}

	response := map[string]interface{}{
		"comments": commentResponses,
		"total":    int(total),
		"limit":    limit,
		"offset":   offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateComment handles POST /api/v1/comments
func (h *Handlers) CreateComment(w http.ResponseWriter, r *http.Request) {
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
	var req struct {
		Content  string `json:"content"`
		PostID   string `json:"post_id"`
		ParentID string `json:"parent_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Content == "" || req.PostID == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "content and post_id are required")
		return
	}

	postID, err := uuid.Parse(req.PostID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post_id format")
		return
	}

	var parentID pgtype.UUID
	if req.ParentID != "" {
		parsedParentID, err := uuid.Parse(req.ParentID)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "INVALID_PARENT_ID", "Invalid parent_id format")
			return
		}
		parentID = pgtype.UUID{Bytes: parsedParentID, Valid: true}
	}

	ctx := r.Context()

	// Create comment in database
	comment, err := h.queries.CreateComment(ctx, &generated.CreateCommentParams{
		AtprotoUri:   "at://placeholder.uri", // This would come from PDS in real implementation
		AuthorDid:    userCtx.Did,
		AuthorHandle: userCtx.Handle,
		PostID:       pgtype.UUID{Bytes: postID, Valid: true},
		ParentID:     parentID,
		Content:      req.Content,
	})
	if err != nil {
		h.logger.Error("Failed to create comment", "error", err, "user_did", userCtx.Did, "post_id", postID)
		h.writeError(w, http.StatusInternalServerError, "COMMENT_CREATE_FAILED", "Failed to create comment")
		return
	}

	// Update post comment count
	err = h.queries.UpdatePostCommentCount(ctx, pgtype.UUID{Bytes: postID, Valid: true})
	if err != nil {
		h.logger.Error("Failed to update comment count", "error", err, "post_id", postID)
		// Don't fail the request, just log the error
	}

	response := map[string]interface{}{
		"id":         comment.ID.String(),
		"content":    comment.Content,
		"post":       comment.PostID.String(),
		"parent":     comment.ParentID.String(),
		"author":     comment.AuthorDid,
		"created_at": comment.CreatedAt.Time.Format(time.RFC3339),
		"upvotes":    int(*comment.Upvotes),
		"downvotes":  int(*comment.Downvotes),
		"score":      int(*comment.Score),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// UpdateComment handles PUT /api/v1/comments/{id}
func (h *Handlers) UpdateComment(w http.ResponseWriter, r *http.Request) {
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

	// Extract comment ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.writeError(w, http.StatusBadRequest, "INVALID_COMMENT_ID", "Invalid comment ID")
		return
	}
	commentIDStr := pathParts[3]
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_COMMENT_ID", "Invalid comment ID format")
		return
	}

	// Parse request body
	var req struct {
		Content string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Content == "" {
		h.writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "content is required")
		return
	}

	ctx := r.Context()

	// Check if comment exists and user is the author
	comment, err := h.queries.GetCommentByID(ctx, commentID)
	if err != nil {
		h.logger.Error("Comment not found", "error", err, "comment_id", commentID)
		h.writeError(w, http.StatusNotFound, "COMMENT_NOT_FOUND", "Comment not found")
		return
	}

	if comment.AuthorDid != userCtx.Did {
		h.writeError(w, http.StatusForbidden, "FORBIDDEN", "You can only update your own comments")
		return
	}

	// Update comment in database
	updatedComment, err := h.queries.UpdateComment(ctx, &generated.UpdateCommentParams{
		ID:      commentID,
		Content: req.Content,
	})
	if err != nil {
		h.logger.Error("Failed to update comment", "error", err, "comment_id", commentID)
		h.writeError(w, http.StatusInternalServerError, "COMMENT_UPDATE_FAILED", "Failed to update comment")
		return
	}

	response := map[string]interface{}{
		"id":         updatedComment.ID.String(),
		"content":    updatedComment.Content,
		"post":       updatedComment.PostID.String(),
		"parent":     updatedComment.ParentID.String(),
		"author":     updatedComment.AuthorDid,
		"created_at": updatedComment.CreatedAt.Time.Format(time.RFC3339),
		"updated_at": updatedComment.UpdatedAt.Time.Format(time.RFC3339),
		"upvotes":    int(*updatedComment.Upvotes),
		"downvotes":  int(*updatedComment.Downvotes),
		"score":      int(*updatedComment.Score),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteComment handles DELETE /api/v1/comments/{id}
func (h *Handlers) DeleteComment(w http.ResponseWriter, r *http.Request) {
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

	// Extract comment ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		h.writeError(w, http.StatusBadRequest, "INVALID_COMMENT_ID", "Invalid comment ID")
		return
	}
	commentIDStr := pathParts[3]
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "INVALID_COMMENT_ID", "Invalid comment ID format")
		return
	}

	ctx := r.Context()

	// Check if comment exists and user is the author
	comment, err := h.queries.GetCommentByID(ctx, commentID)
	if err != nil {
		h.logger.Error("Comment not found", "error", err, "comment_id", commentID)
		h.writeError(w, http.StatusNotFound, "COMMENT_NOT_FOUND", "Comment not found")
		return
	}

	if comment.AuthorDid != userCtx.Did {
		h.writeError(w, http.StatusForbidden, "FORBIDDEN", "You can only delete your own comments")
		return
	}

	// Delete comment from database
	err = h.queries.DeleteComment(ctx, commentID)
	if err != nil {
		h.logger.Error("Failed to delete comment", "error", err, "comment_id", commentID)
		h.writeError(w, http.StatusInternalServerError, "COMMENT_DELETE_FAILED", "Failed to delete comment")
		return
	}

	// Update post comment count
	err = h.queries.UpdatePostCommentCount(ctx, comment.PostID)
	if err != nil {
		h.logger.Error("Failed to update comment count", "error", err, "post_id", comment.PostID)
		// Don't fail the request, just log the error
	}

	h.logger.Info("Comment deleted", "comment_id", commentID, "user_did", userCtx.Did)
	w.WriteHeader(http.StatusNoContent)
}
