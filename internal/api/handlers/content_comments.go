package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
)

// CreateComment handles creating a new comment
func (h *ContentHandler) CreateComment(ctx context.Context, input *models.CommentInput) (*models.CommentResponse, error) {
	postID := input.PostID
	content := input.Body.Content
	parentCommentID := input.Body.ParentCommentID

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for comment creation")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	pseudonymID := userCtx.ActivePseudonymID
	displayName := userCtx.DisplayName

	log.Info().
		Str("endpoint", "posts/comments").
		Str("component", "handler").
		Int64("post_id", postID).
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", pseudonymID).
		Msg("Create comment requested")

	// Validate input
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Check if post exists
	post, err := h.postDAO.GetPostByID(ctx, postID)
	if post == nil {
		log.Warn().Int64("post_id", postID).Msg("Post not found")
		return nil, fmt.Errorf("post not found: %d", postID)
	}
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to get post")
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Check if post is removed
	if post.IsRemoved.Valid && post.IsRemoved.V {
		log.Warn().Int64("post_id", postID).Msg("Cannot comment on removed post")
		return nil, fmt.Errorf("cannot comment on removed post")
	}

	// Check if post is locked
	if post.IsLocked.Valid && post.IsLocked.V {
		log.Warn().Int64("post_id", postID).Msg("Cannot comment on locked post")
		return nil, fmt.Errorf("cannot comment on locked post")
	}

	// Validate parent comment if provided
	if parentCommentID != nil {
		parentComment, err := h.commentDAO.GetCommentByID(ctx, int64(*parentCommentID))
		if err != nil {
			log.Error().Err(err).Int("parent_comment_id", *parentCommentID).Msg("Failed to get parent comment")
			return nil, err
		}
		if parentComment == nil {
			log.Warn().Int("parent_comment_id", *parentCommentID).Msg("Parent comment not found")
			return nil, fmt.Errorf("parent comment not found: %d", *parentCommentID)
		}
		if parentComment.PostID != postID {
			log.Warn().Int("parent_comment_id", *parentCommentID).Int64("post_id", postID).Msg("Parent comment does not belong to post")
			return nil, fmt.Errorf("parent comment does not belong to post")
		}
	}

	// Convert parent comment ID to int64 pointer for DAO
	var parentCommentID64 *int64
	if parentCommentID != nil {
		parentID := int64(*parentCommentID)
		parentCommentID64 = &parentID
	}

	// Create comment in database
	comment, err := h.commentDAO.CreateComment(ctx, postID, pseudonymID, content, parentCommentID64)
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to create comment")
		return nil, err
	}

	// Update last active timestamp for the pseudonym
	err = h.pseudonymDAO.UpdateLastActive(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	// Update post comment count
	err = h.postDAO.UpdateCommentCount(ctx, postID, post.CommentCount.V+1)
	if err != nil {
		log.Warn().Err(err).Int64("post_id", postID).Msg("Failed to update post comment count")
		// Don't fail the request for this
	}

	response := models.NewCommentResponse(int(comment.CommentID), content, parentCommentID, pseudonymID, displayName)

	log.Info().
		Str("endpoint", "posts/comments").
		Str("component", "handler").
		Int64("post_id", postID).
		Int64("comment_id", comment.CommentID).
		Msg("Create comment completed")

	return response, nil
}

// EditComment handles editing a comment
func (h *ContentHandler) EditComment(ctx context.Context, input *models.CommentEditInput) (*models.CommentEditResponse, error) {
	commentID := input.CommentID
	content := input.Body.Content
	editReason := input.Body.EditReason

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for comment editing")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	pseudonymID := userCtx.ActivePseudonymID
	displayName := userCtx.DisplayName

	log.Info().
		Str("endpoint", "comments/edit").
		Str("component", "handler").
		Int64("comment_id", commentID).
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", pseudonymID).
		Msg("Edit comment requested")

	// Validate input
	if content == "" {
		return nil, huma.Error400BadRequest("content is required")
	}

	// Check if comment exists
	comment, err := h.commentDAO.GetCommentByID(ctx, commentID)
	if err != nil {
		log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to get comment")
		return nil, huma.Error404NotFound("comment not found")
	}
	if comment == nil {
		log.Warn().Int64("comment_id", commentID).Msg("Comment not found")
		return nil, huma.Error404NotFound("comment not found")
	}

	// Check if comment is removed
	if comment.IsRemoved.Valid && comment.IsRemoved.V {
		log.Warn().Int64("comment_id", commentID).Msg("Cannot edit removed comment")
		return nil, huma.Error400BadRequest("cannot edit removed comment")
	}

	// Check if user owns the comment
	if comment.PseudonymID != pseudonymID {
		log.Warn().Int64("comment_id", commentID).Str("pseudonym_id", pseudonymID).Msg("User does not own comment")
		return nil, huma.Error403Forbidden("you can only edit your own comments")
	}

	// Update comment in database
	err = h.commentDAO.UpdateComment(ctx, commentID, content, editReason)
	if err != nil {
		log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to update comment")
		return nil, err
	}

	// Update last active timestamp for the pseudonym since editing represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	// Convert parent comment ID for response
	var parentCommentID *int
	if comment.ParentCommentID.Valid {
		parentID := int(comment.ParentCommentID.V)
		parentCommentID = &parentID
	}

	response := models.NewCommentEditResponse(int(commentID), content, parentCommentID, pseudonymID, displayName, editReason, true)

	log.Info().
		Str("endpoint", "comments/edit").
		Str("component", "handler").
		Int64("comment_id", commentID).
		Msg("Edit comment completed")

	return response, nil
}

// RemoveComment handles removing/restoring a comment (moderators only)
func (h *ContentHandler) RemoveComment(ctx context.Context, input *models.CommentRemoveInput) (*models.CommentRemoveResponse, error) {
	commentID := input.CommentID
	removed := input.Body.Removed
	reason := input.Body.Reason

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for comment removal")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	pseudonymID := userCtx.ActivePseudonymID
	displayName := userCtx.DisplayName

	log.Info().
		Str("endpoint", "comments/remove").
		Str("component", "handler").
		Int64("comment_id", commentID).
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", pseudonymID).
		Bool("removed", removed).
		Msg("Remove comment requested")

	// Check if comment exists
	comment, err := h.commentDAO.GetCommentByID(ctx, commentID)
	if err != nil {
		log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to get comment")
		return nil, huma.Error404NotFound("comment not found")
	}
	if comment == nil {
		log.Warn().Int64("comment_id", commentID).Msg("Comment not found")
		return nil, huma.Error404NotFound("comment not found")
	}

	// Check if user owns the comment or has moderation permissions
	canModerate := false
	if comment.PseudonymID == pseudonymID {
		// User owns the comment
		canModerate = true
	} else {
		// Check if user has moderation permissions for the post's subforum
		post, err := h.postDAO.GetPostByID(ctx, comment.PostID)
		if err != nil {
			log.Error().Err(err).Int64("post_id", comment.PostID).Msg("Failed to get post for permission check")
			return nil, err
		}
		if post != nil {
			canModerate, err = h.permissionChecker.CheckSubforumCapability(ctx, userCtx.UserID, post.SubforumID, "moderate_content")
			if err != nil {
				log.Error().Err(err).Int64("user_id", userCtx.UserID).Int32("subforum_id", post.SubforumID).Msg("Failed to check moderation permission")
				return nil, err
			}
		}
	}

	if !canModerate {
		log.Warn().Int64("comment_id", commentID).Str("pseudonym_id", pseudonymID).Msg("User lacks permission to remove comment")
		return nil, huma.Error403Forbidden("insufficient permissions to remove comment")
	}

	// Update comment removal status
	err = h.commentDAO.SetCommentRemoved(ctx, commentID, removed, reason, pseudonymID)
	if err != nil {
		log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to update comment removal status")
		return nil, err
	}

	// Update last active timestamp for the pseudonym since moderation represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	response := models.NewCommentRemoveResponse(int(commentID), removed, reason, pseudonymID, displayName)

	log.Info().
		Str("endpoint", "comments/remove").
		Str("component", "handler").
		Int64("comment_id", commentID).
		Bool("removed", removed).
		Msg("Remove comment completed")

	return response, nil
}

// DeleteComment allows the comment author to delete their own comment (soft delete)
func (h *ContentHandler) DeleteComment(ctx context.Context, input *models.CommentDeleteInput) (*models.CommentDeleteResponse, error) {
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	pseudonymID := user.ActivePseudonymID
	if pseudonymID == "" {
		return nil, huma.Error422UnprocessableEntity("No active pseudonym")
	}

	err = h.commentDAO.MarkCommentAsDeletedByPseudonym(ctx, input.CommentID, pseudonymID, input.Body.Reason)
	if err != nil {
		log.Error().Err(err).Int64("comment_id", input.CommentID).Str("pseudonym_id", pseudonymID).Msg("Failed to delete comment by user")
		return nil, huma.Error500InternalServerError("Failed to delete comment")
	}

	// Update last active timestamp for the pseudonym since deleting represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	// Get user info for response
	pseudonymInfo, err := h.pseudonymDAO.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get pseudonym info for deletion response")
		// Continue without user info
	}

	// Since we filter out deleted comments, we can't get the comment after deletion
	// We'll construct the response with the information we have
	now := time.Now()
	response := &models.CommentDeleteResponse{
		Status: 200,
		Body: models.CommentDeleteResponseBody{
			CommentID:    int(input.CommentID),
			DeletedAt:    now.Format(time.RFC3339),
			DeleteReason: input.Body.Reason,
			DeletedBy: struct {
				PseudonymID string `json:"pseudonym_id" example:"user_pseudonym_id"`
				DisplayName string `json:"display_name" example:"user_name"`
			}{
				PseudonymID: pseudonymID,
				DisplayName: pseudonymInfo.DisplayName,
			},
		},
	}

	return response, nil
}

// ReportComment handles reporting a comment
func (h *ContentHandler) ReportComment(ctx context.Context, input *models.CommentReportInput) (*models.CommentReportResponse, error) {
	commentID := input.CommentID
	reportReason := input.Body.ReportReason
	reportDetails := input.Body.ReportDetails

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for comment reporting")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	pseudonymID := userCtx.ActivePseudonymID
	displayName := userCtx.DisplayName

	log.Info().
		Str("endpoint", "comments/report").
		Str("component", "handler").
		Int64("comment_id", commentID).
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", pseudonymID).
		Str("report_reason", reportReason).
		Msg("Report comment requested")

	// Validate input
	if reportReason == "" {
		return nil, huma.Error400BadRequest("report_reason is required")
	}

	// Check if comment exists
	comment, err := h.commentDAO.GetCommentByID(ctx, commentID)
	if err != nil {
		log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to get comment")
		return nil, err
	}
	if comment == nil {
		log.Warn().Int64("comment_id", commentID).Msg("Comment not found")
		return nil, huma.Error404NotFound("comment not found")
	}

	// Check if comment is already removed
	if comment.IsRemoved.Valid && comment.IsRemoved.V {
		log.Warn().Int64("comment_id", commentID).Msg("Cannot report removed comment")
		return nil, huma.Error400BadRequest("cannot report removed comment")
	}

	// Check if user is reporting their own comment
	if comment.PseudonymID == pseudonymID {
		log.Warn().Int64("comment_id", commentID).Str("pseudonym_id", pseudonymID).Msg("User cannot report their own comment")
		return nil, huma.Error400BadRequest("you cannot report your own comment")
	}

	// Create report in database
	contentIDNull := sql.Null[int64]{V: commentID, Valid: true}
	reportDetailsNull := sql.Null[string]{V: reportDetails, Valid: true}
	statusNull := sql.Null[string]{V: "pending", Valid: true}

	reportSetter := &dbmodels.ReportSetter{
		ReporterPseudonymID: &pseudonymID,
		ContentType:         &[]string{"comment"}[0],
		ContentID:           &contentIDNull,
		ReportReason:        &reportReason,
		ReportDetails:       &reportDetailsNull,
		Status:              &statusNull,
	}

	report, err := h.reportDAO.CreateReport(ctx, reportSetter)
	if err != nil {
		log.Error().Err(err).Int64("comment_id", commentID).Str("pseudonym_id", pseudonymID).Msg("Failed to create report")
		return nil, huma.Error500InternalServerError("Failed to create report")
	}

	// Update last active timestamp for the pseudonym since reporting represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	response := models.NewCommentReportResponse(int(report.ReportID), int(commentID), reportReason, reportDetails, pseudonymID, displayName)

	log.Info().
		Str("endpoint", "comments/report").
		Str("component", "handler").
		Int64("comment_id", commentID).
		Int64("report_id", report.ReportID).
		Msg("Report comment completed")

	return response, nil
}
