package handlers

import (
	"context"
	"fmt"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// ContentHandler handles content-related requests
type ContentHandler struct {
	db                bob.Executor
	postDAO           *dao.PostDAO
	commentDAO        *dao.CommentDAO
	subforumDAO       *dao.SubforumDAO
	pseudonymDAO      *dao.PseudonymDAO
	voteDAO           *dao.VoteDAO
	permissionChecker *middleware.PermissionChecker
}

// NewContentHandler creates a new content handler
func NewContentHandler(db bob.Executor) *ContentHandler {
	return &ContentHandler{
		db:                db,
		postDAO:           dao.NewPostDAO(db),
		commentDAO:        dao.NewCommentDAO(db),
		subforumDAO:       dao.NewSubforumDAO(db),
		pseudonymDAO:      dao.NewPseudonymDAO(db),
		voteDAO:           dao.NewVoteDAO(db),
		permissionChecker: middleware.NewPermissionChecker(db),
	}
}

// GetPosts handles getting posts from a subforum
func (h *ContentHandler) GetPosts(ctx context.Context, input *models.PostListInput) (*models.PostListResponse, error) {
	subforumName := input.SubforumName

	log.Info().
		Str("endpoint", "subforums/posts").
		Str("component", "handler").
		Str("subforum_name", subforumName).
		Int("page", input.Page).
		Int("limit", input.Limit).
		Str("sort", input.Sort).
		Str("time", input.Time).
		Msg("Get posts requested")

	// Check if subforum exists
	subforum, err := h.subforumDAO.GetSubforumByName(ctx, subforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to get subforum")
		return nil, err
	}
	if subforum == nil {
		log.Warn().Str("subforum_name", subforumName).Msg("Subforum not found")
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check user permissions for private subforums
	// Allow access if IsPrivate is null or false, deny only if explicitly true
	if subforum.IsPrivate.Valid && subforum.IsPrivate.V {
		// Extract user context from request context
		userCtx, err := middleware.ExtractUserFromContext(ctx)
		if err != nil {
			log.Warn().Err(err).Str("subforum_name", subforumName).Msg("User context not available for private subforum access")
			return nil, huma.Error401Unauthorized("authentication required for private subforum")
		}

		// Check if user has access to this private subforum using RBAC
		canAccess, err := h.permissionChecker.CheckPrivateSubforumAccess(ctx, userCtx.UserID, subforum.SubforumID)
		if err != nil {
			log.Error().Err(err).
				Int64("user_id", userCtx.UserID).
				Int32("subforum_id", subforum.SubforumID).
				Str("subforum_name", subforumName).
				Msg("Failed to check private subforum access")
			return nil, fmt.Errorf("failed to verify subforum access")
		}

		if !canAccess {
			log.Warn().
				Int64("user_id", userCtx.UserID).
				Int32("subforum_id", subforum.SubforumID).
				Str("subforum_name", subforumName).
				Msg("User denied access to private subforum")
			return nil, huma.Error403Forbidden("access denied to private subforum")
		}

		log.Info().
			Int64("user_id", userCtx.UserID).
			Str("subforum_name", subforumName).
			Msg("User granted access to private subforum")
	}

	// Determine sort field and direction from input.Sort
	sortField := "created_at"
	sortDesc := true
	switch input.Sort {
	case models.PostSortNew:
		sortField = "created_at"
		sortDesc = true
	case models.PostSortTop:
		sortField = "score"
		sortDesc = true
	case models.PostSortOld:
		sortField = "created_at"
		sortDesc = false
	case models.PostSortComments:
		sortField = "comment_count"
		sortDesc = true
	case models.PostSortViews:
		sortField = "view_count"
		sortDesc = true
		// Add more mappings as needed
	}

	// Get posts from database
	posts, err := h.postDAO.GetPostsBySubforum(ctx, subforum.SubforumID, input.Page, input.Limit, sortField, sortDesc)
	if err != nil {
		log.Error().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to get posts")
		return nil, err
	}

	// Count total posts for pagination
	total, err := h.postDAO.CountPostsBySubforum(ctx, subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to count posts")
		return nil, err
	}

	// Convert database posts to API models
	apiPosts := make([]models.Post, len(posts))
	for i, post := range posts {
		apiPosts[i] = h.convertDBPostToAPIPost(post)
	}

	response := models.NewPostListResponse(apiPosts, input.Page, input.Limit, int(total))

	log.Info().
		Str("endpoint", "subforums/posts").
		Str("component", "handler").
		Str("subforum_name", subforumName).
		Int("count", len(apiPosts)).
		Int("total", int(total)).
		Msg("Get posts completed")

	return response, nil
}

// CreatePost handles creating a new post
func (h *ContentHandler) CreatePost(ctx context.Context, input *models.PostCreateInput) (*models.PostResponse, error) {
	subforumName := input.SubforumName
	title := input.Body.Title
	content := input.Body.Content
	postType := input.Body.PostType
	url := input.Body.URL
	isNSFW := input.Body.IsNSFW
	isSpoiler := input.Body.IsSpoiler

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for post creation")
		return nil, fmt.Errorf("authentication required")
	}

	pseudonymID := userCtx.ActivePseudonymID
	displayName := userCtx.DisplayName

	log.Info().
		Str("endpoint", "subforums/create-post").
		Str("component", "handler").
		Int64("user_id", userCtx.UserID).
		Str("subforum_name", subforumName).
		Str("title", title).
		Str("post_type", postType).
		Msg("Create post requested")

	// Validate input
	if title == "" {
		return nil, huma.Error400BadRequest("title is required")
	}
	if content == "" && postType == "text" {
		return nil, huma.Error400BadRequest("content is required for text posts")
	}
	if url == "" && postType == "link" {
		return nil, huma.Error400BadRequest("URL is required for link posts")
	}

	// Check if subforum exists
	subforum, err := h.subforumDAO.GetSubforumByName(ctx, subforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to get subforum")
		return nil, err
	}
	if subforum == nil {
		log.Warn().Str("subforum_name", subforumName).Msg("Subforum not found")
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check user permissions for private/restricted subforums
	// Allow access if IsPrivate is null or false, deny only if explicitly true
	if subforum.IsPrivate.Valid && subforum.IsPrivate.V {
		// Check if user has access to this private subforum using RBAC
		canAccess, err := h.permissionChecker.CheckPrivateSubforumAccess(ctx, userCtx.UserID, subforum.SubforumID)
		if err != nil {
			log.Error().Err(err).
				Int64("user_id", userCtx.UserID).
				Int32("subforum_id", subforum.SubforumID).
				Str("subforum_name", subforumName).
				Msg("Failed to check private subforum access")
			return nil, fmt.Errorf("failed to verify subforum access")
		}

		if !canAccess {
			log.Warn().
				Int64("user_id", userCtx.UserID).
				Int32("subforum_id", subforum.SubforumID).
				Str("subforum_name", subforumName).
				Msg("User denied access to private subforum for post creation")
			return nil, huma.Error403Forbidden("access denied to private subforum")
		}

		log.Info().
			Int64("user_id", userCtx.UserID).
			Str("subforum_name", subforumName).
			Msg("User granted access to private subforum for post creation")
	}

	// Create post in database
	var urlPtr *string
	if url != "" {
		urlPtr = &url
	}

	post, err := h.postDAO.CreatePost(ctx, subforum.SubforumID, pseudonymID, title, content, postType, urlPtr, isNSFW, isSpoiler)
	if err != nil {
		log.Error().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to create post")
		return nil, err
	}

	response := models.NewPostResponse(int(post.PostID), title, content, postType, pseudonymID, displayName)

	log.Info().
		Str("endpoint", "subforums/create-post").
		Str("component", "handler").
		Int64("user_id", userCtx.UserID).
		Int64("post_id", post.PostID).
		Msg("Create post completed")

	return response, nil
}

// GetPostDetails handles getting detailed information about a specific post
func (h *ContentHandler) GetPostDetails(ctx context.Context, input *models.PostDetailsInput) (*models.PostDetailsResponse, error) {
	postID := input.PostID
	sort := input.Sort

	log.Info().
		Str("endpoint", "posts/details").
		Str("component", "handler").
		Int64("post_id", postID).
		Str("sort", sort).
		Msg("Get post details requested")

	// Get post by ID
	post, err := h.postDAO.GetPostByID(ctx, postID)
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to get post")
		return nil, err
	}
	if post == nil {
		log.Warn().Int64("post_id", postID).Msg("Post not found")
		return nil, fmt.Errorf("post not found: %d", postID)
	}

	// Check if post is removed
	if post.IsRemoved.Valid && post.IsRemoved.V {
		log.Warn().Int64("post_id", postID).Msg("Post is removed")
		return nil, fmt.Errorf("post is removed")
	}

	// Get comments for the post
	comments, err := h.commentDAO.GetCommentsByPostWithNestedReplies(ctx, postID)
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to get comments")
		return nil, err
	}

	// Convert database post and comments to API models
	apiPost := h.convertDBPostToAPIPost(post)
	apiComments := make([]models.Comment, len(comments))
	for i, comment := range comments {
		apiComments[i] = h.convertDBCommentToAPICommentWithReplies(comment)
	}

	response := models.NewPostDetailsResponse(apiPost, apiComments)

	log.Info().
		Str("endpoint", "posts/details").
		Str("component", "handler").
		Int64("post_id", postID).
		Int("comment_count", len(apiComments)).
		Msg("Get post details completed")

	return response, nil
}

// VoteOnPost handles voting on a post
func (h *ContentHandler) VoteOnPost(ctx context.Context, input *models.PostVoteInput) (*models.VoteResponse, error) {
	postID := input.PostID
	voteValue := input.Body.VoteValue

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for voting")
		return nil, fmt.Errorf("authentication required")
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
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to get post")
		return nil, err
	}
	if post == nil {
		log.Warn().Int64("post_id", postID).Msg("Post not found")
		return nil, fmt.Errorf("post not found: %d", postID)
	}

	// Check if post is removed
	if post.IsRemoved.Valid && post.IsRemoved.V {
		log.Warn().Int64("post_id", postID).Msg("Cannot vote on removed post")
		return nil, fmt.Errorf("cannot vote on removed post")
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

	// Update post score in database
	err = h.postDAO.UpdatePostScore(ctx, postID, int32(score), int32(upvotes), int32(downvotes))
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to update post score")
		return nil, err
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

// CreateComment handles creating a new comment
func (h *ContentHandler) CreateComment(ctx context.Context, input *models.CommentInput) (*models.CommentResponse, error) {
	postID := input.PostID
	content := input.Body.Content
	parentCommentID := input.Body.ParentCommentID

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for comment creation")
		return nil, fmt.Errorf("authentication required")
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
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to get post")
		return nil, err
	}
	if post == nil {
		log.Warn().Int64("post_id", postID).Msg("Post not found")
		return nil, fmt.Errorf("post not found: %d", postID)
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

// VoteOnComment handles voting on a comment
func (h *ContentHandler) VoteOnComment(ctx context.Context, input *models.CommentVoteInput) (*models.CommentVoteResponse, error) {
	commentID := input.CommentID
	voteValue := input.Body.VoteValue

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for voting")
		return nil, fmt.Errorf("authentication required")
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

	// Update comment score in database
	err = h.commentDAO.UpdateCommentScore(ctx, commentID, int32(score), int32(upvotes), int32(downvotes))
	if err != nil {
		log.Error().Err(err).Int64("comment_id", commentID).Msg("Failed to update comment score")
		return nil, err
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

// convertDBPostToAPIPost converts a database post to an API post model
func (h *ContentHandler) convertDBPostToAPIPost(dbPost *dbmodels.Post) models.Post {
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
	// Try to extract user context from the current context
	if userCtx, err := middleware.ExtractUserFromContext(context.Background()); err == nil && userCtx != nil {
		vote, err := h.voteDAO.GetVoteByPseudonymAndContent(context.Background(), userCtx.ActivePseudonymID, "post", dbPost.PostID)
		if err == nil && vote != nil {
			userVote = int(vote.VoteValue)
		}
	}

	apiPost := models.Post{
		PostID:       int(dbPost.PostID),
		Title:        dbPost.Title,
		Content:      dbPost.Content.V,
		PostType:     dbPost.PostType,
		URL:          dbPost.URL.V,
		IsSelfPost:   dbPost.IsSelfPost.V,
		IsNSFW:       dbPost.IsNSFW.V,
		IsSpoiler:    dbPost.IsSpoiler.V,
		Score:        int(dbPost.Score.V),
		Upvotes:      int(dbPost.Upvotes.V),
		Downvotes:    int(dbPost.Downvotes.V),
		CommentCount: int(dbPost.CommentCount.V),
		ViewCount:    int(dbPost.ViewCount.V),
		CreatedAt:    dbPost.CreatedAt.V.Format("2006-01-02T15:04:05Z"),
		UserVote:     userVote,
		IsSaved:      false, // TODO: Implement saved posts functionality
	}

	// Set author info
	apiPost.Author.PseudonymID = dbPost.PseudonymID
	apiPost.Author.DisplayName = displayName

	// Set subforum info
	apiPost.Subforum.SubforumID = int(dbPost.SubforumID)
	apiPost.Subforum.Name = subforumName
	apiPost.Subforum.DisplayName = subforumDisplayName

	return apiPost
}

// convertDBCommentToAPIComment converts a database comment to an API comment model
func (h *ContentHandler) convertDBCommentToAPIComment(dbComment *dbmodels.Comment) models.Comment {
	// Get pseudonym display name
	displayName := "Unknown"
	if dbComment.R.Pseudonym != nil {
		displayName = dbComment.R.Pseudonym.DisplayName
	}

	// Get user vote if authenticated
	userVote := 0
	// Try to extract user context from the current context
	if userCtx, err := middleware.ExtractUserFromContext(context.Background()); err == nil && userCtx != nil {
		vote, err := h.voteDAO.GetVoteByPseudonymAndContent(context.Background(), userCtx.ActivePseudonymID, "comment", dbComment.CommentID)
		if err == nil && vote != nil {
			userVote = int(vote.VoteValue)
		}
	}

	// Convert parent comment ID
	var parentCommentID *int
	if dbComment.ParentCommentID.Valid {
		parentID := int(dbComment.ParentCommentID.V)
		parentCommentID = &parentID
	}

	apiComment := models.Comment{
		CommentID:       int(dbComment.CommentID),
		Content:         dbComment.Content,
		ParentCommentID: parentCommentID,
		Score:           int(dbComment.Score.V),
		CreatedAt:       dbComment.CreatedAt.V.Format("2006-01-02T15:04:05Z"),
		UserVote:        userVote,
		Replies:         []models.Comment{}, // Empty for non-nested conversion
	}

	// Set author info
	apiComment.Author.PseudonymID = dbComment.PseudonymID
	apiComment.Author.DisplayName = displayName

	return apiComment
}

// convertDBCommentToAPICommentWithReplies converts a database comment to an API comment model with nested replies
func (h *ContentHandler) convertDBCommentToAPICommentWithReplies(dbComment *dbmodels.Comment) models.Comment {
	// Get pseudonym display name
	displayName := "Unknown"
	if dbComment.R.Pseudonym != nil {
		displayName = dbComment.R.Pseudonym.DisplayName
	}

	// Get user vote if authenticated
	userVote := 0
	// Try to extract user context from the current context
	if userCtx, err := middleware.ExtractUserFromContext(context.Background()); err == nil && userCtx != nil {
		vote, err := h.voteDAO.GetVoteByPseudonymAndContent(context.Background(), userCtx.ActivePseudonymID, "comment", dbComment.CommentID)
		if err == nil && vote != nil {
			userVote = int(vote.VoteValue)
		}
	}

	// Convert parent comment ID
	var parentCommentID *int
	if dbComment.ParentCommentID.Valid {
		parentID := int(dbComment.ParentCommentID.V)
		parentCommentID = &parentID
	}

	// Convert nested replies recursively
	replies := make([]models.Comment, len(dbComment.R.ReverseComments))
	for i, reply := range dbComment.R.ReverseComments {
		replies[i] = h.convertDBCommentToAPICommentWithReplies(reply)
	}

	apiComment := models.Comment{
		CommentID:       int(dbComment.CommentID),
		Content:         dbComment.Content,
		ParentCommentID: parentCommentID,
		Score:           int(dbComment.Score.V),
		CreatedAt:       dbComment.CreatedAt.V.Format("2006-01-02T15:04:05Z"),
		UserVote:        userVote,
		Replies:         replies,
	}

	// Set author info
	apiComment.Author.PseudonymID = dbComment.PseudonymID
	apiComment.Author.DisplayName = displayName

	return apiComment
}
