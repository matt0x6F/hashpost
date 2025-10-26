package appview

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	generated "github.com/matt0x6f/hashpost/internal/database/generated/appview"
)

// VoteHandlers handles voting operations for posts and comments
type VoteHandlers struct {
	queries *generated.Queries
	logger  *slog.Logger
}

// NewVoteHandlers creates a new vote handlers instance
func NewVoteHandlers(queries *generated.Queries, logger *slog.Logger) *VoteHandlers {
	return &VoteHandlers{
		queries: queries,
		logger:  logger,
	}
}

// VoteRequest represents a vote request
type VoteRequest struct {
	VoteType string `json:"vote_type"` // "up" or "down"
}

// VoteOnPost handles POST/DELETE /api/v1/posts/{id}/vote
func (vh *VoteHandlers) VoteOnPost(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		vh.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID")
		return
	}
	postIDStr := pathParts[4]
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		vh.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID format")
		return
	}

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		vh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := r.Context()

	switch r.Method {
	case http.MethodPost:
		// Parse vote request
		var req VoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			vh.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
			return
		}

		// Validate vote type
		if req.VoteType != "up" && req.VoteType != "down" {
			vh.writeError(w, http.StatusBadRequest, "INVALID_VOTE_TYPE", "vote_type must be 'up' or 'down'")
			return
		}

		// Get the post to find its atproto URI
		post, err := vh.queries.GetAppViewPostByID(ctx, postID)
		if err != nil {
			vh.logger.Error("Failed to get post", "error", err, "post_id", postID)
			vh.writeError(w, http.StatusNotFound, "POST_NOT_FOUND", "Post not found")
			return
		}

		// Use the post's atproto URI for the vote subject
		atprotoURI := post.AtprotoUri

		// Create vote record in PDS
		voteRecord := map[string]interface{}{
			"$type":     "com.hashpost.feed.vote",
			"subject":   atprotoURI,
			"direction": req.VoteType,
			"createdAt": time.Now().Format(time.RFC3339),
		}

		// Proxy to PDS to create the vote record
		pdsResponse, err := vh.proxyToPDS(r, "POST", "/xrpc/com.atproto.repo.createRecord", map[string]interface{}{
			"repo":       userCtx.Did,
			"collection": "com.hashpost.feed.vote",
			"record":     voteRecord,
		})

		if err != nil {
			vh.logger.Error("Failed to create vote in PDS", "error", err)
			vh.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to create vote")
			return
		}

		// Parse PDS response to get the created vote URI
		var pdsResult struct {
			URI string `json:"uri"`
			CID string `json:"cid"`
		}
		if err := json.Unmarshal(pdsResponse, &pdsResult); err != nil {
			vh.logger.Error("Failed to parse PDS response", "error", err)
			vh.writeError(w, http.StatusInternalServerError, "PDS_ERROR", "Failed to parse PDS response")
			return
		}

		// Wait for event processing to complete and get updated vote counts
		var counts *generated.GetPostVoteCountsRow
		err = vh.waitForEventProcessing(ctx, func() error {
			var err error
			counts, err = vh.queries.GetPostVoteCounts(ctx, postID)
			return err
		}, "vote creation", postID.String())

		if err != nil {
			vh.logger.Error("Failed to get vote counts", "error", err, "post_id", postID)
			vh.writeError(w, http.StatusInternalServerError, "VOTE_COUNT_FAILED", "Failed to get vote counts")
			return
		}

		response := VoteResponse{
			VoteType:  VoteResponseVoteType(req.VoteType),
			Upvotes:   int(*counts.Upvotes),
			Downvotes: int(*counts.Downvotes),
			Score:     int(*counts.Score),
		}

		vh.writeJSON(w, http.StatusOK, response)

	case http.MethodDelete:
		// Remove vote
		err = vh.queries.DeleteVote(ctx, &generated.DeleteVoteParams{
			UserDid: userCtx.Did,
			PostID:  pgtype.UUID{Bytes: postID, Valid: true},
		})
		if err != nil {
			vh.logger.Error("Failed to delete vote", "error", err, "user_did", userCtx.Did, "post_id", postID)
			vh.writeError(w, http.StatusInternalServerError, "VOTE_DELETE_FAILED", "Failed to delete vote")
			return
		}

		// Update vote counts
		err = vh.queries.UpdatePostVoteCounts(ctx, postID)
		if err != nil {
			vh.logger.Error("Failed to update vote counts", "error", err, "post_id", postID)
			// Don't fail the request, just log the error
		}

		// Get updated vote counts
		counts, err := vh.queries.GetPostVoteCounts(ctx, postID)
		if err != nil {
			vh.logger.Error("Failed to get vote counts", "error", err, "post_id", postID)
			vh.writeError(w, http.StatusInternalServerError, "VOTE_COUNT_FAILED", "Failed to get vote counts")
			return
		}

		response := VoteResponse{
			VoteType:  "", // No vote
			Upvotes:   int(*counts.Upvotes),
			Downvotes: int(*counts.Downvotes),
			Score:     int(*counts.Score),
		}

		vh.writeJSON(w, http.StatusOK, response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// VoteOnComment handles POST/DELETE /api/v1/comments/{id}/vote
func (vh *VoteHandlers) VoteOnComment(w http.ResponseWriter, r *http.Request) {
	// Extract comment ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		vh.writeError(w, http.StatusBadRequest, "INVALID_COMMENT_ID", "Invalid comment ID")
		return
	}
	commentIDStr := pathParts[4]
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		vh.writeError(w, http.StatusBadRequest, "INVALID_COMMENT_ID", "Invalid comment ID format")
		return
	}

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		vh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := r.Context()

	switch r.Method {
	case http.MethodPost:
		// Parse vote request
		var req VoteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			vh.writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid JSON")
			return
		}

		// Validate vote type
		if req.VoteType != "up" && req.VoteType != "down" {
			vh.writeError(w, http.StatusBadRequest, "INVALID_VOTE_TYPE", "vote_type must be 'up' or 'down'")
			return
		}

		// Create or update vote
		_, err = vh.queries.CreateVoteOnComment(ctx, &generated.CreateVoteOnCommentParams{
			UserDid:   userCtx.Did,
			CommentID: pgtype.UUID{Bytes: commentID, Valid: true},
			VoteType:  req.VoteType,
		})
		if err != nil {
			vh.logger.Error("Failed to create vote", "error", err, "user_did", userCtx.Did, "comment_id", commentID)
			vh.writeError(w, http.StatusInternalServerError, "VOTE_FAILED", "Failed to create vote")
			return
		}

		// Update vote counts
		err = vh.queries.UpdateCommentVoteCounts(ctx, commentID)
		if err != nil {
			vh.logger.Error("Failed to update vote counts", "error", err, "comment_id", commentID)
			// Don't fail the request, just log the error
		}

		// Get updated vote counts
		counts, err := vh.queries.GetCommentVoteCounts(ctx, commentID)
		if err != nil {
			vh.logger.Error("Failed to get vote counts", "error", err, "comment_id", commentID)
			vh.writeError(w, http.StatusInternalServerError, "VOTE_COUNT_FAILED", "Failed to get vote counts")
			return
		}

		response := VoteResponse{
			VoteType:  VoteResponseVoteType(req.VoteType),
			Upvotes:   int(*counts.Upvotes),
			Downvotes: int(*counts.Downvotes),
			Score:     int(*counts.Score),
		}

		vh.writeJSON(w, http.StatusOK, response)

	case http.MethodDelete:
		// Remove vote
		err = vh.queries.DeleteVoteOnComment(ctx, &generated.DeleteVoteOnCommentParams{
			UserDid:   userCtx.Did,
			CommentID: pgtype.UUID{Bytes: commentID, Valid: true},
		})
		if err != nil {
			vh.logger.Error("Failed to delete vote", "error", err, "user_did", userCtx.Did, "comment_id", commentID)
			vh.writeError(w, http.StatusInternalServerError, "VOTE_DELETE_FAILED", "Failed to delete vote")
			return
		}

		// Update vote counts
		err = vh.queries.UpdateCommentVoteCounts(ctx, commentID)
		if err != nil {
			vh.logger.Error("Failed to update vote counts", "error", err, "comment_id", commentID)
			// Don't fail the request, just log the error
		}

		// Get updated vote counts
		counts, err := vh.queries.GetCommentVoteCounts(ctx, commentID)
		if err != nil {
			vh.logger.Error("Failed to get vote counts", "error", err, "comment_id", commentID)
			vh.writeError(w, http.StatusInternalServerError, "VOTE_COUNT_FAILED", "Failed to get vote counts")
			return
		}

		response := VoteResponse{
			VoteType:  "", // No vote
			Upvotes:   int(*counts.Upvotes),
			Downvotes: int(*counts.Downvotes),
			Score:     int(*counts.Score),
		}

		vh.writeJSON(w, http.StatusOK, response)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// GetUserVoteOnPost handles GET /api/v1/posts/{id}/vote
func (vh *VoteHandlers) GetUserVoteOnPost(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		vh.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID")
		return
	}
	postIDStr := pathParts[4]
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		vh.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID format")
		return
	}

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		vh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := r.Context()

	// Get user's vote on this post
	vote, err := vh.queries.GetUserVoteOnPost(ctx, &generated.GetUserVoteOnPostParams{
		UserDid: userCtx.Did,
		PostID:  pgtype.UUID{Bytes: postID, Valid: true},
	})
	if err != nil {
		// User hasn't voted on this post
		response := VoteResponse{
			VoteType:  "",
			Upvotes:   0,
			Downvotes: 0,
			Score:     0,
		}
		vh.writeJSON(w, http.StatusOK, response)
		return
	}

	// Get current vote counts
	counts, err := vh.queries.GetPostVoteCounts(ctx, postID)
	if err != nil {
		vh.logger.Error("Failed to get vote counts", "error", err, "post_id", postID)
		vh.writeError(w, http.StatusInternalServerError, "VOTE_COUNT_FAILED", "Failed to get vote counts")
		return
	}

	response := VoteResponse{
		VoteType:  VoteResponseVoteType(vote.VoteType),
		Upvotes:   int(*counts.Upvotes),
		Downvotes: int(*counts.Downvotes),
		Score:     int(*counts.Score),
	}

	vh.writeJSON(w, http.StatusOK, response)
}

// GetUserVoteOnly handles GET /api/v1/posts/{id}/user-vote (returns UserVote format)
func (vh *VoteHandlers) GetUserVoteOnly(w http.ResponseWriter, r *http.Request) {
	// Extract post ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		vh.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID")
		return
	}
	postIDStr := pathParts[4]
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		vh.writeError(w, http.StatusBadRequest, "INVALID_POST_ID", "Invalid post ID format")
		return
	}

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		vh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := r.Context()

	// Get user's vote on this post
	vote, err := vh.queries.GetUserVoteOnPost(ctx, &generated.GetUserVoteOnPostParams{
		UserDid: userCtx.Did,
		PostID:  pgtype.UUID{Bytes: postID, Valid: true},
	})
	if err != nil {
		// User hasn't voted on this post
		response := map[string]interface{}{
			"vote_type": nil,
		}
		vh.writeJSON(w, http.StatusOK, response)
		return
	}

	// Return just the vote type
	response := map[string]interface{}{
		"vote_type": vote.VoteType,
	}

	vh.writeJSON(w, http.StatusOK, response)
}

// GetUserVoteOnComment handles GET /api/v1/comments/{id}/vote
func (vh *VoteHandlers) GetUserVoteOnComment(w http.ResponseWriter, r *http.Request) {
	// Extract comment ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		vh.writeError(w, http.StatusBadRequest, "INVALID_COMMENT_ID", "Invalid comment ID")
		return
	}
	commentIDStr := pathParts[4]
	commentID, err := uuid.Parse(commentIDStr)
	if err != nil {
		vh.writeError(w, http.StatusBadRequest, "INVALID_COMMENT_ID", "Invalid comment ID format")
		return
	}

	// Get authenticated user from context
	userCtx := GetUserContext(r)
	if userCtx == nil {
		vh.writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
		return
	}

	ctx := r.Context()

	// Get user's vote on this comment
	vote, err := vh.queries.GetUserVoteOnComment(ctx, &generated.GetUserVoteOnCommentParams{
		UserDid:   userCtx.Did,
		CommentID: pgtype.UUID{Bytes: commentID, Valid: true},
	})
	if err != nil {
		// User hasn't voted on this comment
		response := VoteResponse{
			VoteType:  "",
			Upvotes:   0,
			Downvotes: 0,
			Score:     0,
		}
		vh.writeJSON(w, http.StatusOK, response)
		return
	}

	// Get current vote counts
	counts, err := vh.queries.GetCommentVoteCounts(ctx, commentID)
	if err != nil {
		vh.logger.Error("Failed to get vote counts", "error", err, "comment_id", commentID)
		vh.writeError(w, http.StatusInternalServerError, "VOTE_COUNT_FAILED", "Failed to get vote counts")
		return
	}

	response := VoteResponse{
		VoteType:  VoteResponseVoteType(vote.VoteType),
		Upvotes:   int(*counts.Upvotes),
		Downvotes: int(*counts.Downvotes),
		Score:     int(*counts.Score),
	}

	vh.writeJSON(w, http.StatusOK, response)
}

// Helper methods

func (vh *VoteHandlers) writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   code,
		"message": message,
	})
}

func (vh *VoteHandlers) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// proxyToPDS proxies a request to the PDS server
func (vh *VoteHandlers) proxyToPDS(r *http.Request, method, path string, body interface{}) ([]byte, error) {
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

// waitForEventProcessing polls the AppView database with exponential backoff
// to wait for event processing to complete. This handles the asynchronous
// nature of PDS → Event → AppView data flow.
func (vh *VoteHandlers) waitForEventProcessing(ctx context.Context, checkFunc func() error, operation string, identifier string) error {
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
				vh.logger.Debug("Event processing completed after retry",
					"operation", operation,
					"identifier", identifier,
					"attempt", attempt+1)
			}
			return nil
		}

		// If this was the last attempt, return the error
		if attempt == maxRetries-1 {
			vh.logger.Error("Event processing timeout",
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

		vh.logger.Debug("Event processing not ready, retrying",
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
