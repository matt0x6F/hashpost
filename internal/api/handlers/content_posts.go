package handlers

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/rs/zerolog/log"
)

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

	// Handle authentication - try to get user context from middleware
	var userCtx *middleware.UserContext
	var err error

	// First, try to get user context from middleware (header-based auth)
	userCtx, err = middleware.ExtractUserFromContext(ctx)
	if err != nil {
		// If no user context from middleware, try cookie-based auth from input
		userCtx, err = middleware.ExtractUserFromHumaInput(&input.AuthInput)
		if err != nil {
			// No authentication - proceed as anonymous user
			log.Debug().Msg("No user context found, proceeding as anonymous user")
		}
	}

	// Parse subforum name to extract community type and actual name
	communityType, actualSubforumName, err := h.parseSubforumName(subforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to parse subforum name")
		return nil, fmt.Errorf("invalid subforum name format: %w", err)
	}

	// Get subforum by community type and name
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, communityType, actualSubforumName)
	if subforum == nil {
		log.Warn().Str("subforum_name", subforumName).Msg("Subforum not found")
		return nil, fmt.Errorf("subforum not found: %s", subforumName)
	}
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to get subforum")
		return nil, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Check user permissions for private subforums
	// Allow access if IsPrivate is null or false, deny only if explicitly true
	if subforum.IsPrivate.Valid && subforum.IsPrivate.V {
		if userCtx == nil {
			log.Warn().Str("subforum_name", subforumName).Msg("User context not available for private subforum access")
			return nil, huma.Error401Unauthorized("authentication required for private subforum")
		}

		// Check if user has access to this private subforum using RBAC
		// Use the secure method that checks only the active pseudonym
		canAccess, err := h.permissionChecker.CheckPrivateSubforumAccessWithActivePseudonym(ctx, userCtx.UserID, subforum.SubforumID, userCtx.ActivePseudonymID)
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

	// If we have user context, set it in the context for conversion functions
	if userCtx != nil {
		ctx = middleware.SetUserContext(ctx, userCtx)
	}

	// Convert database posts to API models
	apiPosts := make([]models.Post, len(posts))
	for i, post := range posts {
		apiPosts[i] = h.convertDBPostToAPIPost(ctx, post)
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
	isSticky := input.Body.IsSticky
	isLocked := input.Body.IsLocked

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for post creation")
		return nil, huma.Error401Unauthorized("Authentication required")
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
		Bool("is_sticky", isSticky).
		Bool("is_locked", isLocked).
		Msg("Create post requested")

	// Validate input
	if title == "" {
		return nil, huma.Error400BadRequest("title is required")
	}
	if content == "" && postType == "text" {
		return nil, huma.Error400BadRequest("content is required for text posts")
	}
	if url == nil && postType == "link" {
		return nil, huma.Error400BadRequest("URL is required for link posts")
	}

	// Parse subforum name to extract community type and actual name
	communityType, actualSubforumName, err := h.parseSubforumName(subforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to parse subforum name")
		return nil, fmt.Errorf("invalid subforum name format: %w", err)
	}

	// Get subforum by community type and name
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, communityType, actualSubforumName)
	if subforum == nil {
		log.Warn().Str("subforum_name", subforumName).Msg("Subforum not found")
		return nil, fmt.Errorf("subforum not found: %s", subforumName)
	}
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to get subforum")
		return nil, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Check moderator permissions for sticky/locked options
	if isSticky || isLocked {
		canModerate := false
		var err error

		// Only check permissions if permission checker is available
		if h.permissionChecker != nil {
			canModerate, err = h.permissionChecker.CheckSubforumCapabilityWithActivePseudonym(ctx, userCtx.UserID, subforum.SubforumID, "moderate_content", userCtx.ActivePseudonymID)
			if err != nil {
				log.Error().Err(err).
					Int64("user_id", userCtx.UserID).
					Int32("subforum_id", subforum.SubforumID).
					Msg("Failed to check moderator permissions")
				return nil, fmt.Errorf("failed to verify moderator permissions")
			}
		}

		if !canModerate {
			log.Info().
				Int64("user_id", userCtx.UserID).
				Int32("subforum_id", subforum.SubforumID).
				Bool("is_sticky", isSticky).
				Bool("is_locked", isLocked).
				Msg("User attempted to create sticky/locked post without moderator permissions - dropping moderator options")
			// Silently drop moderator-only options
			isSticky = false
			isLocked = false
		} else {
			log.Info().
				Int64("user_id", userCtx.UserID).
				Int32("subforum_id", subforum.SubforumID).
				Bool("is_sticky", isSticky).
				Bool("is_locked", isLocked).
				Msg("Moderator creating post with special properties")
		}
	}

	// Check user permissions for private/restricted subforums
	// Allow access if IsPrivate is null or false, deny only if explicitly true
	if subforum.IsPrivate.Valid && subforum.IsPrivate.V {
		// Check if user has access to this private subforum using RBAC
		canAccess := false
		var err error

		// Only check permissions if permission checker is available
		if h.permissionChecker != nil {
			canAccess, err = h.permissionChecker.CheckPrivateSubforumAccessWithActivePseudonym(ctx, userCtx.UserID, subforum.SubforumID, userCtx.ActivePseudonymID)
			if err != nil {
				log.Error().Err(err).
					Int64("user_id", userCtx.UserID).
					Int32("subforum_id", subforum.SubforumID).
					Str("subforum_name", subforumName).
					Msg("Failed to check private subforum access")
				return nil, fmt.Errorf("failed to verify subforum access")
			}
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
	if url != nil {
		urlPtr = url
	}

	post, err := h.postDAO.CreatePost(ctx, subforum.SubforumID, pseudonymID, title, content, postType, urlPtr, isNSFW, isSpoiler)
	if err != nil {
		log.Error().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to create post")
		return nil, err
	}

	// Set sticky/locked status if requested
	if isSticky {
		err = h.postDAO.SetSticky(ctx, post.PostID, true)
		if err != nil {
			log.Error().Err(err).Int64("post_id", post.PostID).Msg("Failed to set post as sticky")
			// Don't fail the request, just log the error
		}
	}

	if isLocked {
		err = h.postDAO.SetLocked(ctx, post.PostID, true)
		if err != nil {
			log.Error().Err(err).Int64("post_id", post.PostID).Msg("Failed to set post as locked")
			// Don't fail the request, just log the error
		}
	}

	// Update last active timestamp for the pseudonym
	err = h.pseudonymDAO.UpdateLastActive(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	// Get the slug from the created post
	var slug string
	if post.Slug.Valid {
		slug = post.Slug.V
	} else {
		// Fallback: generate slug if not set
		slug = generateSlug(title, post.PostID)
	}

	response := models.NewPostResponse(int(post.PostID), title, content, postType, pseudonymID, displayName, slug)

	log.Info().
		Str("endpoint", "subforums/create-post").
		Str("component", "handler").
		Int64("user_id", userCtx.UserID).
		Int64("post_id", post.PostID).
		Bool("is_sticky", isSticky).
		Bool("is_locked", isLocked).
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
		log.Warn().Int64("post_id", postID).Msg("Post is removed")
		return nil, fmt.Errorf("post is removed")
	}

	// Check if post is deleted
	if post.IsDeleted.Valid && post.IsDeleted.V {
		log.Warn().Int64("post_id", postID).Msg("Post is deleted")
		return nil, fmt.Errorf("post is deleted")
	}

	// Get comments for the post
	comments, err := h.commentDAO.GetCommentsByPostWithNestedReplies(ctx, postID)
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to get comments")
		return nil, err
	}

	// Convert database post and comments to API models
	apiPost := h.convertDBPostToAPIPost(ctx, post)
	apiComments := make([]models.Comment, len(comments))
	for i, comment := range comments {
		apiComments[i] = h.convertDBCommentToAPICommentWithReplies(ctx, comment)
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

// GetPostBySlug handles getting detailed information about a specific post by slug
func (h *ContentHandler) GetPostBySlug(ctx context.Context, input *models.PostBySlugInput) (*models.PostDetailsResponse, error) {
	subforumName := input.SubforumName
	slug := input.Slug
	sort := input.Sort

	log.Info().
		Str("endpoint", "subforums/posts/slug").
		Str("component", "handler").
		Str("subforum_name", subforumName).
		Str("slug", slug).
		Str("sort", sort).
		Msg("Get post by slug requested")

	// Handle authentication - try middleware context first, then input struct
	var userCtx *middleware.UserContext
	var err error

	// First, try to get user context from middleware (header-based auth)
	userCtx, err = middleware.ExtractUserFromContext(ctx)
	if err != nil {
		// If no user context from middleware, try cookie-based auth from input
		userCtx, err = middleware.ExtractUserFromHumaInput(&input.AuthInput)
		if err != nil {
			// No authentication - proceed as anonymous user
			log.Debug().Msg("No user context found, proceeding as anonymous user")
		}
	}

	// Parse subforum name to extract community type and actual name
	communityType, actualSubforumName, err := h.parseSubforumName(subforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to parse subforum name")
		return nil, fmt.Errorf("invalid subforum name format: %w", err)
	}

	// Get subforum by community type and name
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, communityType, actualSubforumName)
	if subforum == nil {
		log.Warn().Str("subforum_name", subforumName).Msg("Subforum not found")
		return nil, fmt.Errorf("subforum not found: %s", subforumName)
	}
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to get subforum")
		return nil, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Get post by subforum and slug
	post, err := h.postDAO.GetPostBySubforumAndSlug(ctx, subforum.SubforumID, slug)
	if post == nil {
		log.Warn().Str("slug", slug).Str("subforum_name", subforumName).Msg("Post not found")
		return nil, fmt.Errorf("post not found: %s", slug)
	}
	if err != nil {
		log.Error().Err(err).Str("slug", slug).Int32("subforum_id", subforum.SubforumID).Msg("Failed to get post")
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Check if post is removed
	if post.IsRemoved.Valid && post.IsRemoved.V {
		log.Warn().Int64("post_id", post.PostID).Msg("Post is removed")
		return nil, huma.Error404NotFound("post is removed")
	}

	// Check if post is deleted
	if post.IsDeleted.Valid && post.IsDeleted.V {
		log.Warn().Int64("post_id", post.PostID).Msg("Post is deleted")
		return nil, huma.Error404NotFound("post is deleted")
	}

	// Get comments for the post
	comments, err := h.commentDAO.GetCommentsByPostWithNestedReplies(ctx, post.PostID)
	if err != nil {
		log.Error().Err(err).Int64("post_id", post.PostID).Msg("Failed to get comments")
		return nil, err
	}

	// If we have user context, set it in the context for conversion functions
	if userCtx != nil {
		ctx = middleware.SetUserContext(ctx, userCtx)
	}

	// Convert database post and comments to API models
	apiPost := h.convertDBPostToAPIPost(ctx, post)
	apiComments := make([]models.Comment, len(comments))
	for i, comment := range comments {
		apiComments[i] = h.convertDBCommentToAPICommentWithReplies(ctx, comment)
	}

	response := models.NewPostDetailsResponse(apiPost, apiComments)

	log.Info().
		Str("endpoint", "subforums/posts/slug").
		Str("component", "handler").
		Str("subforum_name", subforumName).
		Str("slug", slug).
		Int("comment_count", len(apiComments)).
		Msg("Get post by slug completed")

	return response, nil
}

// LockPost handles locking/unlocking a post
func (h *ContentHandler) LockPost(ctx context.Context, input *models.PostLockInput) (*models.PostResponse, error) {
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	post, err := h.postDAO.GetPostByID(ctx, input.PostID)
	if err != nil || post == nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	canModerate, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, "moderate_content", &post.SubforumID)
	if err != nil || !canModerate {
		return nil, huma.Error403Forbidden("Moderator permission required")
	}
	if err := h.postDAO.SetLocked(ctx, input.PostID, input.Body.Locked); err != nil {
		return nil, fmt.Errorf("failed to update lock state: %w", err)
	}

	// Update last active timestamp for the pseudonym since moderation represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, userCtx.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", userCtx.ActivePseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	post, err = h.postDAO.GetPostByID(ctx, input.PostID)
	if err != nil || post == nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	apiPost := h.convertDBPostToAPIPost(ctx, post)
	return models.NewPostResponse(apiPost.PostID, apiPost.Title, apiPost.Content, apiPost.PostType, apiPost.Author.PseudonymID, apiPost.Author.DisplayName, apiPost.Slug), nil
}

// Sticky/Unsticky Post
func (h *ContentHandler) StickyPost(ctx context.Context, input *models.PostStickyInput) (*models.PostResponse, error) {
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	post, err := h.postDAO.GetPostByID(ctx, input.PostID)
	if err != nil || post == nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	canModerate, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, "moderate_content", &post.SubforumID)
	if err != nil || !canModerate {
		return nil, huma.Error403Forbidden("Moderator permission required")
	}
	if err := h.postDAO.SetSticky(ctx, input.PostID, input.Body.Sticky); err != nil {
		return nil, fmt.Errorf("failed to update sticky state: %w", err)
	}

	// Update last active timestamp for the pseudonym since moderation represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, userCtx.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", userCtx.ActivePseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	post, err = h.postDAO.GetPostByID(ctx, input.PostID)
	if err != nil || post == nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	apiPost := h.convertDBPostToAPIPost(ctx, post)
	return models.NewPostResponse(apiPost.PostID, apiPost.Title, apiPost.Content, apiPost.PostType, apiPost.Author.PseudonymID, apiPost.Author.DisplayName, apiPost.Slug), nil
}

// Remove/Restore Post
func (h *ContentHandler) RemovePost(ctx context.Context, input *models.PostRemoveInput) (*models.PostResponse, error) {
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	post, err := h.postDAO.GetPostByID(ctx, input.PostID)
	if err != nil || post == nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	canModerate, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, "moderate_content", &post.SubforumID)
	if err != nil || !canModerate {
		return nil, huma.Error403Forbidden("Moderator permission required")
	}
	if err := h.postDAO.SetRemoved(ctx, input.PostID, input.Body.Removed); err != nil {
		return nil, fmt.Errorf("failed to update removed state: %w", err)
	}

	// Update last active timestamp for the pseudonym since moderation represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, userCtx.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", userCtx.ActivePseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	post, err = h.postDAO.GetPostByID(ctx, input.PostID)
	if err != nil || post == nil {
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}
	apiPost := h.convertDBPostToAPIPost(ctx, post)
	return models.NewPostResponse(apiPost.PostID, apiPost.Title, apiPost.Content, apiPost.PostType, apiPost.Author.PseudonymID, apiPost.Author.DisplayName, apiPost.Slug), nil
}

// EditPost handles editing a post
func (h *ContentHandler) EditPost(ctx context.Context, input *models.PostEditInput) (*models.PostEditResponse, error) {
	postID := input.PostID
	title := input.Body.Title
	content := input.Body.Content

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("User context not available for post editing")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	pseudonymID := userCtx.ActivePseudonymID
	displayName := userCtx.DisplayName

	log.Info().
		Str("endpoint", "posts/edit").
		Str("component", "handler").
		Int64("post_id", postID).
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", pseudonymID).
		Msg("Edit post requested")

	// Validate input
	if title == "" {
		return nil, huma.Error400BadRequest("title is required")
	}
	if content == "" {
		return nil, huma.Error400BadRequest("content is required")
	}

	// Check if post exists
	post, err := h.postDAO.GetPostByID(ctx, postID)
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to get post")
		return nil, err
	}
	if post == nil {
		log.Warn().Int64("post_id", postID).Msg("Post not found")
		return nil, huma.Error404NotFound("post not found")
	}

	// Check if post is removed
	if post.IsRemoved.Valid && post.IsRemoved.V {
		log.Warn().Int64("post_id", postID).Msg("Cannot edit removed post")
		return nil, huma.Error400BadRequest("cannot edit removed post")
	}

	// Check if user owns the post
	if post.PseudonymID != pseudonymID {
		log.Warn().Int64("post_id", postID).Str("pseudonym_id", pseudonymID).Msg("User does not own post")
		return nil, huma.Error403Forbidden("you can only edit your own posts")
	}

	err = h.postDAO.UpdatePost(ctx, postID, title, content)
	if err != nil {
		log.Error().Err(err).Int64("post_id", postID).Msg("Failed to update post")
		return nil, err
	}

	// Update last active timestamp for the pseudonym since editing represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	response := models.NewPostEditResponse(int(postID), title, content, pseudonymID, displayName, "", true)

	log.Info().
		Str("endpoint", "posts/edit").
		Str("component", "handler").
		Int64("post_id", postID).
		Msg("Edit post completed")

	return response, nil
}

// DeletePost allows the post author to delete their own post (soft delete)
func (h *ContentHandler) DeletePost(ctx context.Context, input *models.PostDeleteInput) (*models.PostDeleteResponse, error) {
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		return nil, huma.Error401Unauthorized("User not authenticated")
	}

	pseudonymID := user.ActivePseudonymID
	if pseudonymID == "" {
		return nil, huma.Error422UnprocessableEntity("No active pseudonym")
	}

	err = h.postDAO.MarkPostAsDeletedByPseudonym(ctx, input.PostID, pseudonymID, input.Body.Reason)
	if err != nil {
		log.Error().Err(err).Int64("post_id", input.PostID).Str("pseudonym_id", pseudonymID).Msg("Failed to delete post by user")
		return nil, huma.Error500InternalServerError("Failed to delete post")
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

	// Since we filter out deleted posts, we can't get the post after deletion
	// We'll construct the response with the information we have
	now := time.Now()
	response := &models.PostDeleteResponse{
		Status: 200,
		Body: models.PostDeleteResponseBody{
			PostID:       int(input.PostID),
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

// generateSlug creates a URL-friendly slug from a title and post ID
func generateSlug(title string, postID int64) string {
	// Convert to lowercase and remove special characters
	re := regexp.MustCompile(`[^a-zA-Z0-9\s]`)
	slug := re.ReplaceAllString(strings.ToLower(title), "")

	// Replace spaces with hyphens
	slug = regexp.MustCompile(`\s+`).ReplaceAllString(slug, "-")

	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")

	// Limit to 50 characters
	if len(slug) > 50 {
		slug = slug[:50]
	}

	// If slug is empty, use a default
	if slug == "" {
		slug = "post"
	}

	// Add deterministic suffix based on post ID
	return fmt.Sprintf("%s-%d", slug, postID)
}
