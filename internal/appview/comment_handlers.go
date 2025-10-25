package appview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
	openapi_types "github.com/oapi-codegen/runtime/types"
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

	// Get comments from database with user votes if authenticated
	var comments []*generated.ListCommentsByPostWithUserVotesRow
	userCtx := GetUserContext(r)

	if userCtx != nil {
		// User is authenticated, get comments with vote information
		comments, err = h.queries.ListCommentsByPostWithUserVotes(ctx, &generated.ListCommentsByPostWithUserVotesParams{
			PostID:  pgtype.UUID{Bytes: postID, Valid: true},
			Limit:   int32(limit),
			Offset:  int32(offset),
			UserDid: userCtx.Did,
		})
	} else {
		// User not authenticated, get comments with aggregate vote counts but no user vote
		oldComments, err := h.queries.ListCommentsByPost(ctx, &generated.ListCommentsByPostParams{
			PostID: pgtype.UUID{Bytes: postID, Valid: true},
			Limit:  int32(limit),
			Offset: int32(offset),
		})
		if err != nil {
			h.logger.Error("Failed to list comments", "error", err, "post_id", postID)
			h.writeError(w, http.StatusInternalServerError, "COMMENT_LIST_FAILED", "Failed to list comments")
			return
		}

		// Convert old format to new format and fetch vote counts for each comment
		comments = make([]*generated.ListCommentsByPostWithUserVotesRow, len(oldComments))
		for i, comment := range oldComments {
			// Fetch vote counts for this comment
			voteCounts, err := h.queries.GetCommentVoteCounts(ctx, comment.ID)
			if err != nil {
				h.logger.Error("Failed to get comment vote counts", "error", err, "comment_id", comment.ID)
				// Continue with zero counts if vote count fetch fails
				voteCounts = &generated.GetCommentVoteCountsRow{
					Upvotes:   &[]int32{0}[0],
					Downvotes: &[]int32{0}[0],
					Score:     &[]int32{0}[0],
				}
			}

			comments[i] = &generated.ListCommentsByPostWithUserVotesRow{
				ID:                comment.ID,
				AtprotoUri:        comment.AtprotoUri,
				AuthorDid:         comment.AuthorDid,
				AuthorHandle:      comment.AuthorHandle,
				PostID:            comment.PostID,
				ParentID:          comment.ParentID,
				Content:           comment.Content,
				CreatedAt:         comment.CreatedAt,
				UpdatedAt:         comment.UpdatedAt,
				PostTitle:         comment.PostTitle,
				SubforumSlug:      comment.SubforumSlug,
				AuthorDisplayName: comment.AuthorDisplayName,
				AuthorAvatarUrl:   comment.AuthorAvatarUrl,
				UserVoteType:      nil, // No vote for unauthenticated users
				Upvotes:           voteCounts.Upvotes,
				Downvotes:         voteCounts.Downvotes,
				Score:             voteCounts.Score,
			}
		}
	}
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
	commentResponses := make([]Comment, len(comments))
	for i, comment := range comments {
		commentResponses[i] = Comment{
			Id: openapi_types.UUID(comment.ID),
			Author: Author{
				Did:         comment.AuthorDid,
				Handle:      comment.AuthorHandle,
				DisplayName: comment.AuthorDisplayName,
				AvatarUrl:   comment.AuthorAvatarUrl,
			},
			Post: openapi_types.UUID(comment.PostID.Bytes),
			Parent: func() *openapi_types.UUID {
				if comment.ParentID.Valid {
					u := openapi_types.UUID(comment.ParentID.Bytes)
					return &u
				} else {
					return nil
				}
			}(),
			Content:   comment.Content,
			CreatedAt: comment.CreatedAt.Time,
			UpdatedAt: &comment.UpdatedAt.Time,
			Upvotes: func() *int {
				if comment.Upvotes != nil {
					v := int(*comment.Upvotes)
					return &v
				} else {
					return nil
				}
			}(),
			Downvotes: func() *int {
				if comment.Downvotes != nil {
					v := int(*comment.Downvotes)
					return &v
				} else {
					return nil
				}
			}(),
			UserVote: func() *CommentUserVote {
				if comment.UserVoteType != nil {
					v := CommentUserVote(*comment.UserVoteType)
					return &v
				} else {
					return nil
				}
			}(),
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

	// Get user profile from AppView database (same as posts)
	userProfile, err := h.queries.GetUserByDID(r.Context(), userCtx.Did)
	if err != nil {
		h.logger.Error("Failed to get user profile", "error", err, "did", userCtx.Did)
		h.writeError(w, http.StatusInternalServerError, "USER_NOT_FOUND", "User profile not found")
		return
	}

	// Get the post to find its atproto URI
	post, err := h.queries.GetAppViewPostByID(r.Context(), postID)
	if err != nil {
		h.logger.Error("Failed to get post", "error", err, "post_id", postID)
		h.writeError(w, http.StatusNotFound, "POST_NOT_FOUND", "Post not found")
		return
	}

	// Parse parent ID if specified and get parent comment URI
	var parentCommentURI *string
	if req.ParentID != "" {
		parentID, err := uuid.Parse(req.ParentID)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "INVALID_PARENT_ID", "Invalid parent_id format")
			return
		}
		parentComment, err := h.queries.GetCommentWithPost(r.Context(), parentID)
		if err != nil {
			h.writeError(w, http.StatusNotFound, "PARENT_NOT_FOUND", "Parent comment not found")
			return
		}
		parentCommentURI = &parentComment.AtprotoUri
	}

	// Create comment record in PDS
	commentRecord := map[string]interface{}{
		"$type":     "com.hashpost.feed.comment",
		"text":      req.Content,
		"post":      post.AtprotoUri,
		"createdAt": time.Now().Format(time.RFC3339),
	}

	if parentCommentURI != nil {
		commentRecord["parent"] = *parentCommentURI
	}

	// Proxy to PDS to create the comment record
	pdsResponse, err := h.proxyToPDS(r, "POST", "/xrpc/com.atproto.repo.createRecord", map[string]interface{}{
		"repo":       userCtx.Did,
		"collection": "com.hashpost.feed.comment",
		"record":     commentRecord,
	})

	if err != nil {
		h.logger.Error("Failed to create comment in PDS", "error", err)
		h.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to create comment")
		return
	}

	// Parse PDS response to get the created comment URI
	var pdsResult struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}
	if err := json.Unmarshal(pdsResponse, &pdsResult); err != nil {
		h.logger.Error("Failed to parse PDS response", "error", err)
		h.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to parse PDS response")
		return
	}

	// Wait a moment for the event to be processed by AppView
	// In a production system, you might want to poll or use a more sophisticated approach
	time.Sleep(100 * time.Millisecond)

	// Get the created comment from AppView database
	createdComment, err := h.queries.GetCommentByURI(r.Context(), pdsResult.URI)
	if err != nil {
		h.logger.Error("Failed to get created comment from AppView", "error", err, "uri", pdsResult.URI)
		h.writeError(w, http.StatusInternalServerError, "SYNC_ERROR", "Comment created but not yet available")
		return
	}

	// Return created comment with user profile data (same as posts)
	response := Comment{
		Id: openapi_types.UUID(createdComment.ID),
		Author: Author{
			Did:         userProfile.Did,
			Handle:      userProfile.Handle,
			DisplayName: userProfile.DisplayName, // Use display name from user profile
			AvatarUrl:   userProfile.AvatarUrl,
		},
		Post: openapi_types.UUID(postID),
		Parent: func() *openapi_types.UUID {
			if createdComment.ParentID.Valid {
				u := openapi_types.UUID(createdComment.ParentID.Bytes)
				return &u
			} else {
				return nil
			}
		}(),
		Content:   req.Content,
		CreatedAt: createdComment.CreatedAt.Time,
		UpdatedAt: &createdComment.UpdatedAt.Time,
		Upvotes:   func() *int { v := 0; return &v }(),
		Downvotes: func() *int { v := 0; return &v }(),
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

	response := Comment{
		Id: openapi_types.UUID(updatedComment.ID),
		Author: Author{
			Did:         updatedComment.AuthorDid,
			Handle:      updatedComment.AuthorHandle,
			DisplayName: updatedComment.AuthorDisplayName,
			AvatarUrl:   updatedComment.AuthorAvatarUrl,
		},
		Post: openapi_types.UUID(updatedComment.PostID.Bytes),
		Parent: func() *openapi_types.UUID {
			if updatedComment.ParentID.Valid {
				u := openapi_types.UUID(updatedComment.ParentID.Bytes)
				return &u
			} else {
				return nil
			}
		}(),
		Content:   updatedComment.Content,
		CreatedAt: updatedComment.CreatedAt.Time,
		UpdatedAt: &updatedComment.UpdatedAt.Time,
		Upvotes: func() *int {
			if updatedComment.Upvotes != nil {
				v := int(*updatedComment.Upvotes)
				return &v
			} else {
				return nil
			}
		}(),
		Downvotes: func() *int {
			if updatedComment.Downvotes != nil {
				v := int(*updatedComment.Downvotes)
				return &v
			} else {
				return nil
			}
		}(),
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

// Helper methods for PDS integration

func (h *Handlers) getPDSURL(did string) string {
	// TODO: Implement proper PDS discovery
	// For now, return the HashPost PDS URL
	// In Docker environment, use service name; in local dev, use localhost
	return "http://hashpost-pds:8080" // HashPost PDS URL
}

func (h *Handlers) createRecordInPDS(ctx context.Context, pdsURL, accessToken, repo, collection string, record map[string]interface{}) (uri, cid string, err error) {
	// Make HTTP request to PDS com.atproto.repo.createRecord endpoint
	url := fmt.Sprintf("%s/xrpc/com.atproto.repo.createRecord", pdsURL)

	// Prepare request body
	requestBody := map[string]interface{}{
		"repo":       repo,
		"collection": collection,
		"record":     record,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Name", "appview") // Mark this as a service-to-service call
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to make request to PDS: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("PDS returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response struct {
		URI string `json:"uri"`
		CID string `json:"cid"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", "", fmt.Errorf("failed to decode PDS response: %w", err)
	}

	return response.URI, response.CID, nil
}

func (h *Handlers) waitForCommentSync(ctx context.Context, uri string) (*generated.AppviewComment, error) {
	// Poll for comment to appear in AppView database
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			comment, err := h.queries.GetCommentByURI(ctx, uri)
			if err == nil {
				return comment, nil
			}
		}
	}
}
