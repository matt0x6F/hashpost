package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

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

// ContentHandlerConfig holds configuration for creating a ContentHandler
type ContentHandlerConfig struct {
	DB                 bob.Executor
	RawDB              *sql.DB
	IBESystem          *ibe.IBESystem
	IdentityMappingDAO dao.IdentityMappingDAOInterface
	UserDAO            dao.UserDAOInterface
	PostDAO            dao.PostDAOInterface
	CommentDAO         dao.CommentDAOInterface
	SubforumDAO        dao.SubforumDAOInterface
	PseudonymDAO       dao.PseudonymDAOInterface
	VoteDAO            dao.VoteDAOInterface
	UserBlocksDAO      dao.UserBlocksDAOInterface
	RoleKeyDAO         dao.RoleKeyDAOInterface
	PermissionChecker  middleware.PermissionCheckerInterface
	PermissionDAO      dao.PermissionDAOInterface
	ReportDAO          dao.ReportDAOInterface
}

// NewContentHandlerConfig creates a new configuration for ContentHandler
func NewContentHandlerConfig(db bob.Executor, rawDB *sql.DB, ibeSystem *ibe.IBESystem) *ContentHandlerConfig {
	return &ContentHandlerConfig{
		DB:        db,
		RawDB:     rawDB,
		IBESystem: ibeSystem,
	}
}

// ContentHandler handles content-related requests
type ContentHandler struct {
	db                 bob.Executor
	rawDB              *sql.DB
	ibeSystem          *ibe.IBESystem
	identityMappingDAO dao.IdentityMappingDAOInterface
	userDAO            dao.UserDAOInterface
	postDAO            dao.PostDAOInterface
	commentDAO         dao.CommentDAOInterface
	subforumDAO        dao.SubforumDAOInterface
	pseudonymDAO       dao.PseudonymDAOInterface
	voteDAO            dao.VoteDAOInterface
	permissionChecker  middleware.PermissionCheckerInterface
	permissionDAO      dao.PermissionDAOInterface
	reportDAO          dao.ReportDAOInterface
}

// NewContentHandler creates a new content handler with optional dependencies
// If db is provided, real DAOs will be created. If nil, mock DAOs should be provided.
func NewContentHandler(
	db bob.Executor,
	rawDB *sql.DB,
	ibeSystem *ibe.IBESystem,
	identityMappingDAO dao.IdentityMappingDAOInterface,
	userDAO dao.UserDAOInterface,
	postDAO dao.PostDAOInterface,
	commentDAO dao.CommentDAOInterface,
	subforumDAO dao.SubforumDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
	voteDAO dao.VoteDAOInterface,
	userBlocksDAO dao.UserBlocksDAOInterface,
	roleKeyDAO dao.RoleKeyDAOInterface,
	permissionChecker middleware.PermissionCheckerInterface,
	permissionDAO dao.PermissionDAOInterface,
	reportDAO dao.ReportDAOInterface,
) *ContentHandler {
	// If db is provided, create real DAOs (production mode)
	if db != nil {
		roleKeyDAO = dao.NewRoleKeyDAO(db, nil)
		userBlocksDAO = dao.NewUserBlocksDAO(db)

		// Safe type assertions with error handling
		identityMappingDAOImpl, ok := identityMappingDAO.(*dao.IdentityMappingDAO)
		if !ok {
			log.Error().Msg("identityMappingDAO is not of type *dao.IdentityMappingDAO")
			return nil
		}
		userDAOImpl, ok := userDAO.(*dao.UserDAO)
		if !ok {
			log.Error().Msg("userDAO is not of type *dao.UserDAO")
			return nil
		}
		roleKeyDAOImpl, ok := roleKeyDAO.(*dao.RoleKeyDAO)
		if !ok {
			log.Error().Msg("roleKeyDAO is not of type *dao.RoleKeyDAO")
			return nil
		}
		userBlocksDAOImpl, ok := userBlocksDAO.(*dao.UserBlocksDAO)
		if !ok {
			log.Error().Msg("userBlocksDAO is not of type *dao.UserBlocksDAO")
			return nil
		}

		pseudonymDAO = dao.NewPseudonymDAO(db, ibeSystem, identityMappingDAOImpl, userDAOImpl, roleKeyDAOImpl, userBlocksDAOImpl)
		permissionDAO = dao.NewPermissionDAO(db)
		postDAO = dao.NewPostDAO(db)
		commentDAO = dao.NewCommentDAO(db)
		subforumDAO = dao.NewSubforumDAO(db)
		voteDAO = dao.NewVoteDAO(db)
		permissionChecker = middleware.NewPermissionChecker(db)
		reportDAO = dao.NewReportDAO(db)
	}

	return &ContentHandler{
		db:                 db,
		rawDB:              rawDB,
		ibeSystem:          ibeSystem,
		identityMappingDAO: identityMappingDAO,
		userDAO:            userDAO,
		postDAO:            postDAO,
		commentDAO:         commentDAO,
		subforumDAO:        subforumDAO,
		pseudonymDAO:       pseudonymDAO,
		voteDAO:            voteDAO,
		permissionChecker:  permissionChecker,
		permissionDAO:      permissionDAO,
		reportDAO:          reportDAO,
	}
}

// NewContentHandlerFromConfig creates a new content handler from a configuration struct
func NewContentHandlerFromConfig(cfg *ContentHandlerConfig) *ContentHandler {
	return NewContentHandler(
		cfg.DB,
		cfg.RawDB,
		cfg.IBESystem,
		cfg.IdentityMappingDAO,
		cfg.UserDAO,
		cfg.PostDAO,
		cfg.CommentDAO,
		cfg.SubforumDAO,
		cfg.PseudonymDAO,
		cfg.VoteDAO,
		cfg.UserBlocksDAO,
		cfg.RoleKeyDAO,
		cfg.PermissionChecker,
		cfg.PermissionDAO,
		cfg.ReportDAO,
	)
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
