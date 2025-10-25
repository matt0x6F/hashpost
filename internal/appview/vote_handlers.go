package appview

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

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

		// Create or update vote
		_, err = vh.queries.CreateVote(ctx, &generated.CreateVoteParams{
			UserDid:  userCtx.Did,
			PostID:   pgtype.UUID{Bytes: postID, Valid: true},
			VoteType: req.VoteType,
		})
		if err != nil {
			vh.logger.Error("Failed to create vote", "error", err, "user_did", userCtx.Did, "post_id", postID)
			vh.writeError(w, http.StatusInternalServerError, "VOTE_FAILED", "Failed to create vote")
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
