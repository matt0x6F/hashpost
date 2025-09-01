package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
)

// convertDBPostToAPIPost converts a database post to an API post model
func (h *ContentHandler) convertDBPostToAPIPost(ctx context.Context, dbPost *dbmodels.Post) models.Post {
	// Get pseudonym display name
	displayName := "Unknown"
	if dbPost.R.Pseudonym != nil {
		displayName = dbPost.R.Pseudonym.DisplayName
	}

	// Get subforum info
	subforumName := "Unknown"
	subforumDisplayName := "Unknown"
	if dbPost.R.Subforum != nil {
		subforumName = dbPost.R.Subforum.Name
		subforumDisplayName = dbPost.R.Subforum.DisplayName
	}

	// Get user vote if authenticated
	userVote := 0
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil || userCtx == nil {
		log.Warn().Msg("User context missing in convertDBPostToAPIPost")
	} else {
		vote, err := h.voteDAO.GetVoteByPseudonymAndContent(ctx, userCtx.ActivePseudonymID, "post", dbPost.PostID)
		if err == nil && vote != nil {
			userVote = int(vote.VoteValue)
		}
	}

	apiPost := models.Post{
		PostID:       int(dbPost.PostID),
		Slug:         dbPost.Slug.V,
		Title:        dbPost.Title,
		Content:      dbPost.Content.V,
		PostType:     dbPost.PostType,
		URL:          dbPost.URL.V,
		IsSelfPost:   dbPost.IsSelfPost.V,
		IsNSFW:       dbPost.IsNSFW.V,
		IsSpoiler:    dbPost.IsSpoiler.V,
		IsLocked:     dbPost.IsLocked.V,
		IsSticky:     dbPost.IsStickied.V,
		IsRemoved:    dbPost.IsRemoved.V,
		Score:        int(dbPost.Score.V),
		Upvotes:      int(dbPost.Upvotes.V),
		Downvotes:    int(dbPost.Downvotes.V),
		CommentCount: int(dbPost.CommentCount.V),
		CreatedAt:    dbPost.CreatedAt.V.Format("2006-01-02T15:04:05Z"),
		UserVote:     userVote,
	}

	// Set author info
	apiPost.Author.PseudonymID = dbPost.PseudonymID
	apiPost.Author.DisplayName = displayName

	// Set subforum info
	apiPost.Subforum.Name = subforumName
	apiPost.Subforum.DisplayName = subforumDisplayName

	return apiPost
}

// convertDBCommentToAPICommentWithReplies converts a database comment to an API comment model with nested replies
func (h *ContentHandler) convertDBCommentToAPICommentWithReplies(ctx context.Context, dbComment *dbmodels.Comment) models.Comment {
	displayName := "Unknown"
	if dbComment.R.Pseudonym != nil {
		displayName = dbComment.R.Pseudonym.DisplayName
	}

	userVote := 0
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil || userCtx == nil {
		log.Warn().Msg("User context missing in convertDBCommentToAPICommentWithReplies")
	} else {
		log.Info().Str("pseudonym_id", userCtx.ActivePseudonymID).Msg("User context found in convertDBCommentToAPICommentWithReplies")
		// Only get user vote if comment is not deleted (freeze voting for deleted comments)
		if !dbComment.IsDeleted.Valid || !dbComment.IsDeleted.V {
			vote, err := h.voteDAO.GetVoteByPseudonymAndContent(ctx, userCtx.ActivePseudonymID, "comment", dbComment.CommentID)
			if err == nil && vote != nil {
				userVote = int(vote.VoteValue)
			}
		}
	}

	var parentCommentID *int
	if dbComment.ParentCommentID.Valid {
		parentID := int(dbComment.ParentCommentID.V)
		parentCommentID = &parentID
	}

	replies := make([]models.Comment, len(dbComment.R.ReverseComments))
	for i, reply := range dbComment.R.ReverseComments {
		replies[i] = h.convertDBCommentToAPICommentWithReplies(ctx, reply)
	}

	// Handle deleted comments
	content := dbComment.Content
	authorDisplayName := displayName
	isDeleted := false
	deletedAt := ""
	deleteReason := ""
	var deletedBy struct {
		PseudonymID string `json:"pseudonym_id" example:"user_pseudonym_id"`
		DisplayName string `json:"display_name" example:"user_name"`
	}

	if dbComment.IsDeleted.Valid && dbComment.IsDeleted.V {
		content = "[deleted]"
		authorDisplayName = "[deleted]"
		isDeleted = true
		deletedAt = dbComment.DeletedByPseudonymAt.V.Format("2006-01-02T15:04:05Z")
		deleteReason = dbComment.DeletedByPseudonymReason.V

		// Get deleted by info
		if dbComment.R.DeletedByPseudonymPseudonym != nil {
			deletedBy.PseudonymID = dbComment.DeletedByPseudonymID.V
			deletedBy.DisplayName = dbComment.R.DeletedByPseudonymPseudonym.DisplayName
		}
	}

	apiComment := models.Comment{
		CommentID:       int(dbComment.CommentID),
		Content:         content,
		ParentCommentID: parentCommentID,
		Score:           int(dbComment.Score.V),
		CreatedAt:       dbComment.CreatedAt.V.Format("2006-01-02T15:04:05Z"),
		UserVote:        userVote,
		Replies:         replies,
		IsDeleted:       isDeleted,
		DeletedAt:       deletedAt,
		DeleteReason:    deleteReason,
		DeletedBy:       deletedBy,
	}

	apiComment.Author.PseudonymID = dbComment.PseudonymID
	apiComment.Author.DisplayName = authorDisplayName

	return apiComment
}

// parseSubforumName parses a full subforum name (e.g., "t/subforum-name") into community type and name
func (h *ContentHandler) parseSubforumName(fullName string) (communityType, subforumName string, err error) {
	// Handle different formats:
	// 1. "t/subforum-name" -> communityType: "t", subforumName: "subforum-name"
	// 2. "subforum-name" -> communityType: "h", subforumName: "subforum-name" (default for h/ subforums)

	if fullName == "" {
		return "", "", fmt.Errorf("subforum name cannot be empty")
	}

	// Check if it contains a slash (community type prefix)
	if strings.Contains(fullName, "/") {
		parts := strings.SplitN(fullName, "/", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid subforum name format: expected 'community-type/name'")
		}

		communityType = parts[0]
		subforumName = parts[1]

		// Validate community type
		validTypes := []string{constants.CommunityTypeTopical, constants.CommunityTypeGeographic, constants.CommunityTypeBranded, constants.CommunityTypeCreator, "h"}
		isValid := false
		for _, validType := range validTypes {
			if communityType == validType {
				isValid = true
				break
			}
		}

		if !isValid {
			return "", "", fmt.Errorf("invalid community type: %s", communityType)
		}

		return communityType, subforumName, nil
	}

	// No slash found, treat as h/ subforum (default)
	return "h", fullName, nil
}
