package handlers

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/rs/zerolog/log"
)

// VoteOnPost handles voting on a post
func (h *ContentHandler) VoteOnPost(ctx context.Context, input *models.PostVoteInput) (*models.VoteResponse, error) {
	postID := input.PostID
	voteValue := input.Body.VoteValue

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for voting")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	pseudonymID := userCtx.ActivePseudonymID

	log.Info().
		Str("endpoint", "posts/vote").
		Str("component", "handler").
		Int64("post_id", postID).
		Int("vote_value", voteValue).
		Str("pseudonym_id", pseudonymID).
		Msg("Vote on post requested")

	// Validate vote value
	if voteValue != -1 && voteValue != 0 && voteValue != 1 {
		return nil, fmt.Errorf("invalid vote value: must be -1, 0, or 1")
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
		log.Warn().Int64("post_id", postID).Msg("Cannot vote on removed post")
		return nil, fmt.Errorf("cannot vote on removed post")
	}

	// Check if post is deleted
	if post.IsDeleted.Valid && post.IsDeleted.V {
		log.Warn().Int64("post_id", postID).Msg("Cannot vote on deleted post")
		return nil, fmt.Errorf("cannot vote on deleted post")
	}

	// Handle vote
	if voteValue == 0 {
		// Remove vote
		existingVote, err := h.voteDAO.GetVoteByPseudonymAndContent(ctx, pseudonymID, "post", postID)
		if err != nil {
			log.Error().Err(err).Int64("post_id", postID).Msg("Failed to get existing vote")
			return nil, err
		}
		if existingVote != nil {
			err = h.voteDAO.DeleteVote(ctx, existingVote.VoteID)
			if err != nil {
				log.Error().Err(err).Int64("post_id", postID).Msg("Failed to delete vote")
				return nil, err
			}
		}
	} else {
		// Create or update vote
		_, err = h.voteDAO.UpsertVote(ctx, pseudonymID, "post", postID, int32(voteValue))
		if err != nil {
			log.Error().Err(err).Int64("post_id", postID).Msg("Failed to upsert vote")
			return nil, err
		}
	}

	// Get updated vote summary
	upvotes, downvotes, _, err := h.voteDAO.GetVoteSummaryByContent(ctx, "post", postID)
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to get vote summary")
		return nil, err
	}

	score := upvotes - downvotes

	// Update last active timestamp for the pseudonym
	err = h.pseudonymDAO.UpdateLastActive(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	// Update post score in database
	err = h.postDAO.UpdatePostScore(ctx, postID, int32(score), int32(upvotes), int32(downvotes))
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to update post score")
		return nil, err
	}

	// Update karma for the post author
	err = h.pseudonymDAO.UpdateKarmaForPseudonym(ctx, post.PseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", post.PseudonymID).Msg("Failed to update karma for post author")
		// Don't fail the request for this error
	}

	response := models.NewVoteResponse(int(postID), voteValue, score, upvotes, downvotes)

	log.Info().
		Str("endpoint", "posts/vote").
		Str("component", "handler").
		Int64("post_id", postID).
		Int("vote_value", voteValue).
		Int("score", score).
		Msg("Vote on post completed")

	return response, nil
}

// VoteOnComment handles voting on a comment
func (h *ContentHandler) VoteOnComment(ctx context.Context, input *models.CommentVoteInput) (*models.CommentVoteResponse, error) {
	commentID := input.CommentID
	voteValue := input.Body.VoteValue

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for voting")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	pseudonymID := userCtx.ActivePseudonymID

	log.Info().
		Str("endpoint", "comments/vote").
		Str("component", "handler").
		Int64("comment_id", commentID).
		Int("vote_value", voteValue).
		Str("pseudonym_id", pseudonymID).
		Msg("Vote on comment requested")

	// Validate vote value
	if voteValue != -1 && voteValue != 0 && voteValue != 1 {
		return nil, fmt.Errorf("invalid vote value: must be -1, 0, or 1")
	}

	// Check if comment exists
	comment, err := h.commentDAO.GetCommentByID(ctx, commentID)
	if err != nil {
		log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to get comment")
		return nil, err
	}
	if comment == nil {
		log.Warn().Int64("comment_id", commentID).Msg("Comment not found")
		return nil, fmt.Errorf("comment not found: %d", commentID)
	}

	// Check if comment is removed
	if comment.IsRemoved.Valid && comment.IsRemoved.V {
		log.Warn().Int64("comment_id", commentID).Msg("Cannot vote on removed comment")
		return nil, fmt.Errorf("cannot vote on removed comment")
	}

	// Check if comment is deleted
	if comment.IsDeleted.Valid && comment.IsDeleted.V {
		log.Warn().Int64("comment_id", commentID).Msg("Cannot vote on deleted comment")
		return nil, fmt.Errorf("cannot vote on deleted comment")
	}

	// Handle vote
	if voteValue == 0 {
		// Remove vote
		existingVote, err := h.voteDAO.GetVoteByPseudonymAndContent(ctx, pseudonymID, "comment", commentID)
		if err != nil {
			log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to get existing vote")
			return nil, err
		}
		if existingVote != nil {
			err = h.voteDAO.DeleteVote(ctx, existingVote.VoteID)
			if err != nil {
				log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to delete vote")
				return nil, err
			}
		}
	} else {
		// Create or update vote
		_, err = h.voteDAO.UpsertVote(ctx, pseudonymID, "comment", commentID, int32(voteValue))
		if err != nil {
			log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to upsert vote")
			return nil, err
		}
	}

	// Get updated vote summary
	upvotes, downvotes, _, err := h.voteDAO.GetVoteSummaryByContent(ctx, "comment", commentID)
	if err != nil {
		log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to get vote summary")
		return nil, err
	}

	score := upvotes - downvotes

	// Update last active timestamp for the pseudonym
	err = h.pseudonymDAO.UpdateLastActive(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	// Update comment score in database
	err = h.commentDAO.UpdateCommentScore(ctx, commentID, int32(score), int32(upvotes), int32(downvotes))
	if err != nil {
		log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to update comment score")
		return nil, err
	}

	// Update karma for the comment author
	err = h.pseudonymDAO.UpdateKarmaForPseudonym(ctx, comment.PseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", comment.PseudonymID).Msg("Failed to update karma for comment author")
		// Don't fail the request for this error
	}

	response := models.NewCommentVoteResponse(int(commentID), voteValue, score, upvotes, downvotes)

	log.Info().
		Str("endpoint", "comments/vote").
		Str("component", "handler").
		Int64("comment_id", commentID).
		Int("vote_value", voteValue).
		Int("score", score).
		Msg("Vote on comment completed")

	return response, nil
}
