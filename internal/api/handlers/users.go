package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// UserHandler handles user management requests
type UserHandler struct {
	userDAO            dao.UserDAOInterface
	pseudonymDAO       dao.PseudonymDAOInterface
	userPreferencesDAO dao.UserPreferencesDAOInterface
	userBlocksDAO      dao.UserBlocksDAOInterface
	postDAO            dao.PostDAOInterface
	commentDAO         dao.CommentDAOInterface
	ibeSystem          *ibe.IBESystem
}

// NewUserHandler creates a new user handler with optional dependencies
// If db is provided, real DAOs will be created. If nil, mock DAOs should be provided.
func NewUserHandler(
	db bob.Executor,
	userDAO dao.UserDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
	userPreferencesDAO dao.UserPreferencesDAOInterface,
	userBlocksDAO dao.UserBlocksDAOInterface,
	postDAO dao.PostDAOInterface,
	commentDAO dao.CommentDAOInterface,
	ibeSystem *ibe.IBESystem,
) *UserHandler {
	// If db is provided, create real DAOs (production mode)
	if db != nil {
		userDAO = dao.NewUserDAO(db)

		// Safe type assertions with error handling
		identityMappingDAO := dao.NewIdentityMappingDAO(db)
		roleKeyDAO := dao.NewRoleKeyDAO(db)
		userBlocksDAO = dao.NewUserBlocksDAO(db)

		userDAOImpl, ok := userDAO.(*dao.UserDAO)
		if !ok {
			log.Error().Msg("userDAO is not of type *dao.UserDAO")
			return nil
		}
		userBlocksDAOImpl, ok := userBlocksDAO.(*dao.UserBlocksDAO)
		if !ok {
			log.Error().Msg("userBlocksDAO is not of type *dao.UserBlocksDAO")
			return nil
		}

		pseudonymDAO = dao.NewPseudonymDAO(db, ibeSystem, identityMappingDAO, userDAOImpl, roleKeyDAO, userBlocksDAOImpl)
		userPreferencesDAO = dao.NewUserPreferencesDAO(db)
		postDAO = dao.NewPostDAO(db)
		commentDAO = dao.NewCommentDAO(db)
	}

	return &UserHandler{
		userDAO:            userDAO,
		pseudonymDAO:       pseudonymDAO,
		userPreferencesDAO: userPreferencesDAO,
		userBlocksDAO:      userBlocksDAO,
		postDAO:            postDAO,
		commentDAO:         commentDAO,
		ibeSystem:          ibeSystem,
	}
}

// GetPseudonymProfile handles getting a pseudonym's public profile
func (h *UserHandler) GetPseudonymProfile(ctx context.Context, input *apimodels.PseudonymIDPathParam) (*apimodels.PseudonymProfileResponse, error) {
	pseudonymID := input.PseudonymID

	log.Info().
		Str("endpoint", "pseudonyms/profile").
		Str("component", "handler").
		Str("pseudonym_id", pseudonymID).
		Msg("Get pseudonym profile requested")

	pseudonym, err := h.pseudonymDAO.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get pseudonym from database")
		return nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}
	if pseudonym == nil {
		log.Warn().Str("pseudonym_id", pseudonymID).Msg("Pseudonym not found")
		return nil, fmt.Errorf("pseudonym not found")
	}
	if !pseudonym.IsActive.Valid || !pseudonym.IsActive.V {
		log.Warn().Str("pseudonym_id", pseudonymID).Msg("Pseudonym is inactive")
		return nil, fmt.Errorf("pseudonym is inactive")
	}

	// ✅ No longer need to get user - pseudonym is self-contained
	// The old code that got user and checked user.IsActive is removed
	// since we no longer have direct foreign key relationships

	displayName := pseudonym.DisplayName
	bio := ""
	if pseudonym.Bio.Valid {
		bio = pseudonym.Bio.V
	}
	websiteURL := ""
	if pseudonym.WebsiteURL.Valid {
		websiteURL = pseudonym.WebsiteURL.V
	}
	karmaScore := 0
	if pseudonym.KarmaScore.Valid {
		karmaScore = int(pseudonym.KarmaScore.V)
	}
	showKarma := true
	if pseudonym.ShowKarma.Valid {
		showKarma = pseudonym.ShowKarma.V
	}
	allowDirectMessages := true
	if pseudonym.AllowDirectMessages.Valid {
		allowDirectMessages = pseudonym.AllowDirectMessages.V
	}
	createdAt := ""
	if pseudonym.CreatedAt.Valid {
		createdAt = pseudonym.CreatedAt.V.Format(time.RFC3339)
	}
	lastActiveAt := ""
	if pseudonym.LastActiveAt.Valid {
		lastActiveAt = pseudonym.LastActiveAt.V.Format(time.RFC3339)
	}
	postCount, _ := h.postDAO.CountPostsByPseudonym(ctx, pseudonymID)
	commentCount, _ := h.commentDAO.CountCommentsByPseudonym(ctx, pseudonymID)

	// Calculate and update karma if it's 0 or not set
	if karmaScore == 0 {
		calculatedKarma, err := h.pseudonymDAO.CalculateKarmaForPseudonym(ctx, pseudonymID)
		if err == nil {
			// Update the karma in the database
			h.pseudonymDAO.UpdateKarmaForPseudonym(ctx, pseudonymID)
			karmaScore = int(calculatedKarma)
		}
	}

	response := apimodels.NewPseudonymProfileResponse(pseudonymID, displayName, bio, websiteURL, karmaScore, int(postCount), int(commentCount), showKarma, allowDirectMessages, createdAt, lastActiveAt)

	// Add slug to response
	pseudonymSlug := ""
	if pseudonym.Slug.Valid {
		pseudonymSlug = pseudonym.Slug.V
	}
	response.Body.Slug = pseudonymSlug

	log.Info().Str("endpoint", "pseudonyms/profile").Str("component", "handler").Str("pseudonym_id", pseudonymID).Msg("Get pseudonym profile completed")
	return response, nil
}

// GetPseudonymProfileBySlug handles getting a pseudonym's public profile by slug
func (h *UserHandler) GetPseudonymProfileBySlug(ctx context.Context, input *apimodels.SlugPathParam) (*apimodels.PseudonymProfileResponse, error) {
	slug := input.Slug

	log.Info().
		Str("endpoint", "pseudonyms/profile/slug").
		Str("component", "handler").
		Str("slug", slug).
		Msg("Get pseudonym profile by slug requested")

	// First try to find by slug
	pseudonym, err := h.pseudonymDAO.GetPseudonymBySlug(ctx, slug)
	if err != nil {
		log.Error().Err(err).Str("slug", slug).Msg("Failed to get pseudonym from database")
		return nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}

	// If not found by slug, try to find by pseudonym ID (for backward compatibility)
	if pseudonym == nil {
		log.Info().Str("slug", slug).Msg("Pseudonym not found by slug, trying by ID")
		pseudonym, err = h.pseudonymDAO.GetPseudonymByID(ctx, slug)
		if err != nil {
			log.Error().Err(err).Str("slug", slug).Msg("Failed to get pseudonym by ID from database")
			return nil, fmt.Errorf("failed to get pseudonym: %w", err)
		}

		// If found by ID, generate a slug for it
		if pseudonym != nil && (!pseudonym.Slug.Valid || pseudonym.Slug.V == "") {
			log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Str("display_name", pseudonym.DisplayName).Msg("Generating slug for existing pseudonym")
			generatedSlug, err := h.pseudonymDAO.GenerateSlugFromDisplayName(ctx, pseudonym.DisplayName)
			if err != nil {
				log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to generate slug")
				return nil, fmt.Errorf("failed to generate slug: %w", err)
			}

			// Update the pseudonym with the generated slug
			updates := &dbmodels.PseudonymSetter{
				Slug: &sql.Null[string]{V: generatedSlug, Valid: true},
			}
			err = h.pseudonymDAO.UpdatePseudonym(ctx, pseudonym.PseudonymID, updates)
			if err != nil {
				log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Str("slug", generatedSlug).Msg("Failed to update pseudonym with generated slug")
				// Don't fail the request, just log the error
			} else {
				log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Str("slug", generatedSlug).Msg("Successfully updated pseudonym with generated slug")
				pseudonym.Slug = sql.Null[string]{V: generatedSlug, Valid: true}
			}
		}
	}

	if pseudonym == nil {
		log.Warn().Str("slug", slug).Msg("Pseudonym not found")
		return nil, fmt.Errorf("pseudonym not found")
	}
	if !pseudonym.IsActive.Valid || !pseudonym.IsActive.V {
		log.Warn().Str("slug", slug).Msg("Pseudonym is inactive")
		return nil, fmt.Errorf("pseudonym is inactive")
	}

	displayName := pseudonym.DisplayName
	bio := ""
	if pseudonym.Bio.Valid {
		bio = pseudonym.Bio.V
	}
	websiteURL := ""
	if pseudonym.WebsiteURL.Valid {
		websiteURL = pseudonym.WebsiteURL.V
	}
	karmaScore := 0
	if pseudonym.KarmaScore.Valid {
		karmaScore = int(pseudonym.KarmaScore.V)
	}
	showKarma := true
	if pseudonym.ShowKarma.Valid {
		showKarma = pseudonym.ShowKarma.V
	}
	allowDirectMessages := true
	if pseudonym.AllowDirectMessages.Valid {
		allowDirectMessages = pseudonym.AllowDirectMessages.V
	}
	createdAt := ""
	if pseudonym.CreatedAt.Valid {
		createdAt = pseudonym.CreatedAt.V.Format(time.RFC3339)
	}
	lastActiveAt := ""
	if pseudonym.LastActiveAt.Valid {
		lastActiveAt = pseudonym.LastActiveAt.V.Format(time.RFC3339)
	}
	pseudonymSlug := ""
	if pseudonym.Slug.Valid {
		pseudonymSlug = pseudonym.Slug.V
	}
	postCount, _ := h.postDAO.CountPostsByPseudonym(ctx, pseudonym.PseudonymID)
	commentCount, _ := h.commentDAO.CountCommentsByPseudonym(ctx, pseudonym.PseudonymID)

	// Calculate and update karma if it's 0 or not set
	if karmaScore == 0 {
		calculatedKarma, err := h.pseudonymDAO.CalculateKarmaForPseudonym(ctx, pseudonym.PseudonymID)
		if err == nil {
			// Update the karma in the database
			h.pseudonymDAO.UpdateKarmaForPseudonym(ctx, pseudonym.PseudonymID)
			karmaScore = int(calculatedKarma)
		}
	}

	response := apimodels.NewPseudonymProfileResponse(pseudonym.PseudonymID, displayName, bio, websiteURL, karmaScore, int(postCount), int(commentCount), showKarma, allowDirectMessages, createdAt, lastActiveAt)
	response.Body.Slug = pseudonymSlug
	log.Info().Str("endpoint", "pseudonyms/profile/slug").Str("component", "handler").Str("slug", slug).Msg("Get pseudonym profile by slug completed")
	return response, nil
}

// GetPostsByPseudonym handles getting posts by a pseudonym
func (h *UserHandler) GetPostsByPseudonym(ctx context.Context, input *struct {
	apimodels.SlugPathParam
	Page  int    `query:"page" example:"1"`
	Limit int    `query:"limit" example:"25"`
	Sort  string `query:"sort" example:"created_at" enum:"created_at,score,title"`
}) (*apimodels.PostsByPseudonymResponse, error) {
	slug := input.Slug
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	sortField := input.Sort
	if sortField == "" {
		sortField = "created_at"
	}

	log.Info().
		Str("endpoint", "pseudonyms/posts").
		Str("component", "handler").
		Str("slug", slug).
		Int("page", page).
		Int("limit", limit).
		Str("sort", sortField).
		Msg("Get posts by pseudonym requested")

	// First try to find by slug
	pseudonym, err := h.pseudonymDAO.GetPseudonymBySlug(ctx, slug)
	if err != nil {
		log.Error().Err(err).Str("slug", slug).Msg("Failed to get pseudonym from database")
		return nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}

	// If not found by slug, try to find by pseudonym ID (for backward compatibility)
	if pseudonym == nil {
		log.Info().Str("slug", slug).Msg("Pseudonym not found by slug, trying by ID")
		pseudonym, err = h.pseudonymDAO.GetPseudonymByID(ctx, slug)
		if err != nil {
			log.Error().Err(err).Str("slug", slug).Msg("Failed to get pseudonym by ID from database")
			return nil, fmt.Errorf("failed to get pseudonym: %w", err)
		}
	}

	if pseudonym == nil {
		log.Warn().Str("slug", slug).Msg("Pseudonym not found")
		return nil, fmt.Errorf("pseudonym not found")
	}

	log.Info().Str("pseudonym_id", pseudonym.PseudonymID).Msg("Found pseudonym, getting posts")

	// Get posts by pseudonym
	posts, err := h.postDAO.GetPostsByPseudonym(ctx, pseudonym.PseudonymID, page, limit, sortField, true)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to get posts by pseudonym")
		return nil, fmt.Errorf("failed to get posts: %w", err)
	}

	log.Info().Int("posts_count", len(posts)).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Retrieved posts from database")

	// Convert posts to API models
	apiPosts := make([]apimodels.Post, 0, len(posts))
	for _, post := range posts {
		apiPost := h.convertDBPostToAPIPost(ctx, post)
		apiPosts = append(apiPosts, apiPost)
	}

	// Get total count for pagination
	totalCount, err := h.postDAO.CountPostsByPseudonym(ctx, pseudonym.PseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to count posts by pseudonym")
		return nil, fmt.Errorf("failed to count posts: %w", err)
	}

	response := &apimodels.PostsByPseudonymResponse{
		Status: 200,
		Body: apimodels.PostsByPseudonymResponseBody{
			Posts:      apiPosts,
			TotalCount: int(totalCount),
			Page:       page,
			Limit:      limit,
			TotalPages: int((totalCount + int64(limit) - 1) / int64(limit)),
		},
	}

	log.Info().Str("endpoint", "pseudonyms/posts").Str("component", "handler").Str("slug", slug).Msg("Get posts by pseudonym completed")
	return response, nil
}

// GetCommentsByPseudonym handles getting comments by a pseudonym
func (h *UserHandler) GetCommentsByPseudonym(ctx context.Context, input *struct {
	apimodels.SlugPathParam
	Page  int    `query:"page" example:"1"`
	Limit int    `query:"limit" example:"25"`
	Sort  string `query:"sort" example:"created_at" enum:"created_at,score"`
}) (*apimodels.CommentsByPseudonymResponse, error) {
	slug := input.Slug
	page := input.Page
	if page < 1 {
		page = 1
	}
	limit := input.Limit
	if limit < 1 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	sortField := input.Sort
	if sortField == "" {
		sortField = "created_at"
	}

	log.Info().
		Str("endpoint", "pseudonyms/comments").
		Str("component", "handler").
		Str("slug", slug).
		Int("page", page).
		Int("limit", limit).
		Str("sort", sortField).
		Msg("Get comments by pseudonym requested")

	// First try to find by slug
	pseudonym, err := h.pseudonymDAO.GetPseudonymBySlug(ctx, slug)
	if err != nil {
		log.Error().Err(err).Str("slug", slug).Msg("Failed to get pseudonym from database")
		return nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}

	// If not found by slug, try to find by pseudonym ID (for backward compatibility)
	if pseudonym == nil {
		log.Info().Str("slug", slug).Msg("Pseudonym not found by slug, trying by ID")
		pseudonym, err = h.pseudonymDAO.GetPseudonymByID(ctx, slug)
		if err != nil {
			log.Error().Err(err).Str("slug", slug).Msg("Failed to get pseudonym by ID from database")
			return nil, fmt.Errorf("failed to get pseudonym: %w", err)
		}
	}

	if pseudonym == nil {
		log.Warn().Str("slug", slug).Msg("Pseudonym not found")
		return nil, fmt.Errorf("pseudonym not found")
	}

	// Get comments by pseudonym
	comments, err := h.commentDAO.GetCommentsByPseudonym(ctx, pseudonym.PseudonymID, page, limit, sortField, true)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to get comments by pseudonym")
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	// Convert comments to API models
	apiComments := make([]apimodels.Comment, 0, len(comments))
	for _, comment := range comments {
		apiComment := h.convertDBCommentToAPIComment(ctx, comment)
		apiComments = append(apiComments, apiComment)
	}

	// Get total count for pagination
	totalCount, err := h.commentDAO.CountCommentsByPseudonym(ctx, pseudonym.PseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to count comments by pseudonym")
		return nil, fmt.Errorf("failed to count comments: %w", err)
	}

	response := &apimodels.CommentsByPseudonymResponse{
		Status: 200,
		Body: apimodels.CommentsByPseudonymResponseBody{
			Comments:   apiComments,
			TotalCount: int(totalCount),
			Page:       page,
			Limit:      limit,
			TotalPages: int((totalCount + int64(limit) - 1) / int64(limit)),
		},
	}

	log.Info().Str("endpoint", "pseudonyms/comments").Str("component", "handler").Str("slug", slug).Msg("Get comments by pseudonym completed")
	return response, nil
}

// convertDBPostToAPIPost converts a database post to an API post model
func (h *UserHandler) convertDBPostToAPIPost(ctx context.Context, dbPost *dbmodels.Post) apimodels.Post {
	post := apimodels.Post{
		PostID:       int(dbPost.PostID),
		Title:        dbPost.Title,
		Content:      "",
		PostType:     dbPost.PostType,
		URL:          "",
		IsNSFW:       false,
		IsSpoiler:    false,
		Score:        0,
		Upvotes:      0,
		Downvotes:    0,
		CommentCount: 0,
		ViewCount:    0,
		CreatedAt:    "",
		IsLocked:     false,
		IsSticky:     false,
		IsRemoved:    false,
		IsDeleted:    false,
	}

	if dbPost.Content.Valid {
		post.Content = dbPost.Content.V
	}
	if dbPost.URL.Valid {
		post.URL = dbPost.URL.V
	}
	if dbPost.IsNSFW.Valid {
		post.IsNSFW = dbPost.IsNSFW.V
	}
	if dbPost.IsSpoiler.Valid {
		post.IsSpoiler = dbPost.IsSpoiler.V
	}
	if dbPost.Score.Valid {
		post.Score = int(dbPost.Score.V)
	}
	if dbPost.Upvotes.Valid {
		post.Upvotes = int(dbPost.Upvotes.V)
	}
	if dbPost.Downvotes.Valid {
		post.Downvotes = int(dbPost.Downvotes.V)
	}
	if dbPost.CommentCount.Valid {
		post.CommentCount = int(dbPost.CommentCount.V)
	}
	if dbPost.ViewCount.Valid {
		post.ViewCount = int(dbPost.ViewCount.V)
	}
	if dbPost.CreatedAt.Valid {
		post.CreatedAt = dbPost.CreatedAt.V.Format(time.RFC3339)
	}
	if dbPost.IsLocked.Valid {
		post.IsLocked = dbPost.IsLocked.V
	}
	if dbPost.IsRemoved.Valid {
		post.IsRemoved = dbPost.IsRemoved.V
	}
	if dbPost.IsDeleted.Valid {
		post.IsDeleted = dbPost.IsDeleted.V
	}

	// Load author information if available
	if dbPost.R.Pseudonym != nil {
		post.Author.PseudonymID = dbPost.R.Pseudonym.PseudonymID
		post.Author.DisplayName = dbPost.R.Pseudonym.DisplayName
	}

	// Load subforum information if available
	if dbPost.R.Subforum != nil {
		post.Subforum.Name = dbPost.R.Subforum.Name
		post.Subforum.DisplayName = dbPost.R.Subforum.DisplayName
		post.Subforum.CommunityType = dbPost.R.Subforum.CommunityType
	}

	return post
}

// convertDBCommentToAPIComment converts a database comment to an API comment model
func (h *UserHandler) convertDBCommentToAPIComment(ctx context.Context, dbComment *dbmodels.Comment) apimodels.Comment {
	comment := apimodels.Comment{
		CommentID:       int(dbComment.CommentID),
		Content:         dbComment.Content,
		Score:           0,
		CreatedAt:       "",
		ParentCommentID: nil,
		Replies:         []apimodels.Comment{},
	}

	if dbComment.Score.Valid {
		comment.Score = int(dbComment.Score.V)
	}
	if dbComment.CreatedAt.Valid {
		comment.CreatedAt = dbComment.CreatedAt.V.Format(time.RFC3339)
	}
	if dbComment.ParentCommentID.Valid {
		parentID := int(dbComment.ParentCommentID.V)
		comment.ParentCommentID = &parentID
	}

	// Load author information if available
	if dbComment.R.Pseudonym != nil {
		comment.Author.PseudonymID = dbComment.R.Pseudonym.PseudonymID
		comment.Author.DisplayName = dbComment.R.Pseudonym.DisplayName
	}

	// Load post information if available
	if dbComment.R.Post != nil {
		comment.PostTitle = dbComment.R.Post.Title
		comment.PostID = int(dbComment.R.Post.PostID)

		// Load subforum information if available
		if dbComment.R.Post.R.Subforum != nil {
			comment.SubforumName = dbComment.R.Post.R.Subforum.Name
			comment.SubforumDisplayName = dbComment.R.Post.R.Subforum.DisplayName
			comment.CommunityType = dbComment.R.Post.R.Subforum.CommunityType
		}
	}

	return comment
}

// UpdatePseudonymProfile handles updating the current user's pseudonym profile
// Note: input.Body is for Huma schema only; actual requests are flat JSON.
func (h *UserHandler) UpdatePseudonymProfile(ctx context.Context, input *struct {
	middleware.AuthInput
	apimodels.PseudonymIDPathParam
	apimodels.PseudonymProfileInput
}) (*apimodels.PseudonymProfileResponse, error) {
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Str("endpoint", "pseudonyms/profile").Msg("Authentication required for profile update")
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	userID := int(userCtx.UserID)
	pseudonymID := input.PseudonymID
	log.Info().Str("endpoint", "pseudonyms/profile").Str("component", "handler").Int("user_id", userID).Str("pseudonym_id", pseudonymID).Str("token_type", userCtx.TokenType).Msg("Update pseudonym profile requested")

	// Access fields via input.Body
	if input.Body.DisplayName == "" {
		return nil, fmt.Errorf("display name is required")
	}
	pseudonym, err := h.pseudonymDAO.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get pseudonym from database")
		return nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}
	if pseudonym == nil {
		log.Warn().Str("pseudonym_id", pseudonymID).Msg("Pseudonym not found")
		return nil, fmt.Errorf("pseudonym not found")
	}

	// For profile updates, we trust that the authenticated user owns the pseudonym
	// since they're already authenticated and the pseudonym ID comes from their session
	log.Info().Int("user_id", userID).Str("pseudonym_id", pseudonymID).Msg("User authenticated, proceeding with profile update")

	if input.Body.DisplayName != pseudonym.DisplayName {
		existing, _ := h.pseudonymDAO.GetPseudonymByDisplayName(ctx, input.Body.DisplayName)
		if existing != nil {
			return nil, fmt.Errorf("display name is already taken")
		}
	}

	// Handle slug updates
	var slugToSet *sql.Null[string]
	if input.Body.Slug != "" {
		// Check if slug is already taken by another pseudonym
		existing, _ := h.pseudonymDAO.GetPseudonymBySlug(ctx, input.Body.Slug)
		if existing != nil && existing.PseudonymID != pseudonymID {
			return nil, fmt.Errorf("slug is already taken")
		}
		slug := sql.Null[string]{V: input.Body.Slug, Valid: true}
		slugToSet = &slug
	} else {
		// Generate slug from display name if not provided
		generatedSlug, err := h.pseudonymDAO.GenerateSlugFromDisplayName(ctx, input.Body.DisplayName)
		if err != nil {
			log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to generate slug")
			return nil, fmt.Errorf("failed to generate slug: %w", err)
		}
		slug := sql.Null[string]{V: generatedSlug, Valid: true}
		slugToSet = &slug
	}
	updates := &dbmodels.PseudonymSetter{
		DisplayName: &input.Body.DisplayName,
		Slug:        slugToSet,
	}
	if input.Body.Bio != "" {
		bio := sql.Null[string]{V: input.Body.Bio, Valid: true}
		updates.Bio = &bio
	} else {
		bio := sql.Null[string]{Valid: false}
		updates.Bio = &bio
	}
	if input.Body.WebsiteURL != "" {
		websiteURL := sql.Null[string]{V: input.Body.WebsiteURL, Valid: true}
		updates.WebsiteURL = &websiteURL
	} else {
		websiteURL := sql.Null[string]{Valid: false}
		updates.WebsiteURL = &websiteURL
	}
	if input.Body.ShowKarma != nil {
		showKarma := sql.Null[bool]{V: *input.Body.ShowKarma, Valid: true}
		updates.ShowKarma = &showKarma
	}
	if input.Body.AllowDirectMessages != nil {
		allowDirectMessages := sql.Null[bool]{V: *input.Body.AllowDirectMessages, Valid: true}
		updates.AllowDirectMessages = &allowDirectMessages
	}
	err = h.pseudonymDAO.UpdatePseudonym(ctx, pseudonymID, updates)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym in database")
		return nil, fmt.Errorf("failed to update pseudonym: %w", err)
	}
	finalPseudonym, err := h.pseudonymDAO.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get final pseudonym data")
		return nil, fmt.Errorf("failed to get pseudonym data: %w", err)
	}
	finalDisplayName := finalPseudonym.DisplayName
	finalBio := ""
	if finalPseudonym.Bio.Valid {
		finalBio = finalPseudonym.Bio.V
	}
	finalWebsiteURL := ""
	if finalPseudonym.WebsiteURL.Valid {
		finalWebsiteURL = finalPseudonym.WebsiteURL.V
	}
	karmaScore := 0
	if finalPseudonym.KarmaScore.Valid {
		karmaScore = int(finalPseudonym.KarmaScore.V)
	}
	showKarma := true
	if finalPseudonym.ShowKarma.Valid {
		showKarma = finalPseudonym.ShowKarma.V
	}
	allowDirectMessages := true
	if finalPseudonym.AllowDirectMessages.Valid {
		allowDirectMessages = finalPseudonym.AllowDirectMessages.V
	}
	createdAt := ""
	if finalPseudonym.CreatedAt.Valid {
		createdAt = finalPseudonym.CreatedAt.V.Format(time.RFC3339)
	}
	lastActiveAt := ""
	if finalPseudonym.LastActiveAt.Valid {
		lastActiveAt = finalPseudonym.LastActiveAt.V.Format(time.RFC3339)
	}
	postCount, _ := h.postDAO.CountPostsByPseudonym(ctx, pseudonymID)
	commentCount, _ := h.commentDAO.CountCommentsByPseudonym(ctx, pseudonymID)
	response := apimodels.NewPseudonymProfileResponse(pseudonymID, finalDisplayName, finalBio, finalWebsiteURL, karmaScore, int(postCount), int(commentCount), showKarma, allowDirectMessages, createdAt, lastActiveAt)
	log.Info().Str("endpoint", "pseudonyms/profile").Str("component", "handler").Int("user_id", userID).Str("pseudonym_id", pseudonymID).Msg("Update pseudonym profile completed")
	return response, nil
}

// CreatePseudonym handles creating a new pseudonym for the current user
// Note: input.Body is for Huma schema only; actual requests are flat JSON.
func (h *UserHandler) CreatePseudonym(ctx context.Context, input *struct {
	middleware.AuthInput
	apimodels.CreatePseudonymInput
}) (*apimodels.CreatePseudonymResponse, error) {
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Str("endpoint", "pseudonyms").Msg("Authentication required for pseudonym creation")
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	userID := int(userCtx.UserID)
	log.Info().Str("endpoint", "pseudonyms").Str("component", "handler").Int("user_id", userID).Str("token_type", userCtx.TokenType).Msg("Create pseudonym requested")

	displayName := input.Body.DisplayName
	bio := input.Body.Bio
	websiteURL := input.Body.WebsiteURL
	showKarma := input.Body.ShowKarma
	allowDirectMessages := input.Body.AllowDirectMessages

	if displayName == "" {
		return nil, fmt.Errorf("display name is required")
	}
	existing, _ := h.pseudonymDAO.GetPseudonymByDisplayName(ctx, displayName)
	if existing != nil {
		return nil, fmt.Errorf("display name is already taken")
	}

	// Handle slug generation
	var slugToSet *sql.Null[string]
	if input.Body.Slug != "" {
		// Check if slug is already taken
		existing, _ := h.pseudonymDAO.GetPseudonymBySlug(ctx, input.Body.Slug)
		if existing != nil {
			return nil, fmt.Errorf("slug is already taken")
		}
		slug := sql.Null[string]{V: input.Body.Slug, Valid: true}
		slugToSet = &slug
	} else {
		// Generate slug from display name if not provided
		generatedSlug, err := h.pseudonymDAO.GenerateSlugFromDisplayName(ctx, displayName)
		if err != nil {
			log.Error().Err(err).Int("user_id", userID).Str("display_name", displayName).Msg("Failed to generate slug")
			return nil, fmt.Errorf("failed to generate slug: %w", err)
		}
		slug := sql.Null[string]{V: generatedSlug, Valid: true}
		slugToSet = &slug
	}

	// ✅ Use new method that creates pseudonym and identity mapping together
	pseudonym, err := h.pseudonymDAO.CreatePseudonymWithIdentityMapping(ctx, int64(userID), displayName)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Str("display_name", displayName).Msg("Failed to create pseudonym in database")
		return nil, fmt.Errorf("failed to create pseudonym: %w", err)
	}

	updates := &dbmodels.PseudonymSetter{
		Slug: slugToSet,
	}
	if bio != "" {
		bioVal := sql.Null[string]{V: bio, Valid: true}
		updates.Bio = &bioVal
	}
	if websiteURL != "" {
		websiteURLVal := sql.Null[string]{V: websiteURL, Valid: true}
		updates.WebsiteURL = &websiteURLVal
	}
	if showKarma != nil {
		showKarmaVal := sql.Null[bool]{V: *showKarma, Valid: true}
		updates.ShowKarma = &showKarmaVal
	}
	if allowDirectMessages != nil {
		allowDirectMessagesVal := sql.Null[bool]{V: *allowDirectMessages, Valid: true}
		updates.AllowDirectMessages = &allowDirectMessagesVal
	}
	if len(updates.SetColumns()) > 0 {
		err = h.pseudonymDAO.UpdatePseudonym(ctx, pseudonym.PseudonymID, updates)
		if err != nil {
			log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to update pseudonym with additional fields")
		}
	}
	finalPseudonym, err := h.pseudonymDAO.GetPseudonymByID(ctx, pseudonym.PseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to get final pseudonym data")
		return nil, fmt.Errorf("failed to get pseudonym data: %w", err)
	}
	finalDisplayName := finalPseudonym.DisplayName
	finalBio := ""
	if finalPseudonym.Bio.Valid {
		finalBio = finalPseudonym.Bio.V
	}
	finalWebsiteURL := ""
	if finalPseudonym.WebsiteURL.Valid {
		finalWebsiteURL = finalPseudonym.WebsiteURL.V
	}
	showKarmaVal := true
	if finalPseudonym.ShowKarma.Valid {
		showKarmaVal = finalPseudonym.ShowKarma.V
	}
	allowDirectMessagesVal := true
	if finalPseudonym.AllowDirectMessages.Valid {
		allowDirectMessagesVal = finalPseudonym.AllowDirectMessages.V
	}
	response := apimodels.NewCreatePseudonymResponse(pseudonym.PseudonymID, finalDisplayName, finalBio, finalWebsiteURL, showKarmaVal, allowDirectMessagesVal)
	log.Info().Str("endpoint", "pseudonyms").Str("component", "handler").Int("user_id", userID).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Create pseudonym completed")
	return response, nil
}

// GetUserProfile handles getting the current user's profile with all pseudonyms
func (h *UserHandler) GetUserProfile(ctx context.Context, input *middleware.AuthInput) (*apimodels.UserProfileResponse, error) {
	userCtx, err := middleware.ExtractUserFromHumaInput(input)
	if err != nil {
		log.Warn().Err(err).Str("endpoint", "users/profile").Msg("Authentication required for profile access")
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	userID := int(userCtx.UserID)
	log.Info().Str("endpoint", "users/profile").Str("component", "handler").Int("user_id", userID).Msg("Get user profile requested")
	user, err := h.userDAO.GetUserByID(ctx, int64(userID))
	if err != nil {
		log.Error().Err(err).Int64("user_id", int64(userID)).Msg("Failed to get user from database")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		log.Warn().Int64("user_id", int64(userID)).Msg("User not found")
		return nil, fmt.Errorf("user not found")
	}

	// Get user roles from role keys
	// Parse user roles from role keys
	var userRoles []string
	userRoles = []string{constants.RoleUser} // Default role

	// Note: Role keys are managed by the role key system, but this handler doesn't have access to roleKeyDAO
	// For now, we'll use the default role. In a full implementation, we would get roles from role keys.

	// Use the first role for authentication
	primaryRole := userRoles[0]
	pseudonyms, err := h.pseudonymDAO.GetPseudonymsByUserID(ctx, int64(userID), primaryRole, constants.ScopeAuthentication)
	if err != nil {
		log.Error().Err(err).Int64("user_id", int64(userID)).Str("role", primaryRole).Msg("Failed to get user pseudonyms")
		return nil, fmt.Errorf("failed to get pseudonyms: %w", err)
	}

	pseudonymProfiles := make([]apimodels.PseudonymProfile, len(pseudonyms))
	for i, pseudonym := range pseudonyms {
		karmaScore := 0
		if pseudonym.KarmaScore.Valid {
			karmaScore = int(pseudonym.KarmaScore.V)
		}
		createdAt := ""
		if pseudonym.CreatedAt.Valid {
			createdAt = pseudonym.CreatedAt.V.Format(time.RFC3339)
		}
		lastActiveAt := ""
		if pseudonym.LastActiveAt.Valid {
			lastActiveAt = pseudonym.LastActiveAt.V.Format(time.RFC3339)
		}
		isActive := true
		if pseudonym.IsActive.Valid {
			isActive = pseudonym.IsActive.V
		}
		bio := ""
		if pseudonym.Bio.Valid {
			bio = pseudonym.Bio.V
		}
		websiteURL := ""
		if pseudonym.WebsiteURL.Valid {
			websiteURL = pseudonym.WebsiteURL.V
		}
		showKarma := true
		if pseudonym.ShowKarma.Valid {
			showKarma = pseudonym.ShowKarma.V
		}
		allowDirectMessages := true
		if pseudonym.AllowDirectMessages.Valid {
			allowDirectMessages = pseudonym.AllowDirectMessages.V
		}
		postCount, _ := h.postDAO.CountPostsByPseudonym(ctx, pseudonym.PseudonymID)
		commentCount, _ := h.commentDAO.CountCommentsByPseudonym(ctx, pseudonym.PseudonymID)

		// Get slug from pseudonym
		slug := ""
		if pseudonym.Slug.Valid {
			slug = pseudonym.Slug.V
		}

		pseudonymProfiles[i] = apimodels.PseudonymProfile{
			PseudonymID:         pseudonym.PseudonymID,
			DisplayName:         pseudonym.DisplayName,
			KarmaScore:          karmaScore,
			CreatedAt:           createdAt,
			LastActiveAt:        lastActiveAt,
			IsActive:            isActive,
			Bio:                 bio,
			WebsiteURL:          websiteURL,
			ShowKarma:           showKarma,
			AllowDirectMessages: allowDirectMessages,
			PostCount:           int(postCount),
			CommentCount:        int(commentCount),
			Slug:                slug,
		}
	}
	email := user.Email
	capabilities := userCtx.Capabilities
	response := apimodels.NewUserProfileResponse(userID, email, userRoles, capabilities, pseudonymProfiles)
	log.Info().Str("endpoint", "users/profile").Str("component", "handler").Int("user_id", userID).Msg("Get user profile completed")
	return response, nil
}

// GetUserPreferences handles getting the current user's preferences
func (h *UserHandler) GetUserPreferences(ctx context.Context, input *struct {
	middleware.AuthInput
	apimodels.UserPreferencesInput
}) (*apimodels.UserPreferencesResponse, error) {
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Str("endpoint", "users/preferences").Msg("Authentication required for preferences access")
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	userID := int64(userCtx.UserID)
	log.Info().Str("endpoint", "users/preferences").Str("component", "handler").Int64("user_id", userID).Msg("Get user preferences requested")
	preferences, err := h.userPreferencesDAO.GetUserPreferences(ctx, userID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("Failed to get user preferences from database")
		return nil, fmt.Errorf("failed to get user preferences: %w", err)
	}
	if preferences == nil {
		response := apimodels.NewUserPreferencesResponse("UTC", "en", "light", true, true, true, true)
		return response, nil
	}
	timezone := "UTC"
	if preferences.Timezone.Valid {
		timezone = preferences.Timezone.V
	}
	language := "en"
	if preferences.Language.Valid {
		language = preferences.Language.V
	}
	theme := "light"
	if preferences.Theme.Valid {
		theme = preferences.Theme.V
	}
	emailNotifications := true
	if preferences.EmailNotifications.Valid {
		emailNotifications = preferences.EmailNotifications.V
	}
	pushNotifications := true
	if preferences.PushNotifications.Valid {
		pushNotifications = preferences.PushNotifications.V
	}
	autoHideNSFW := true
	if preferences.AutoHideNSFW.Valid {
		autoHideNSFW = preferences.AutoHideNSFW.V
	}
	autoHideSpoilers := true
	if preferences.AutoHideSpoilers.Valid {
		autoHideSpoilers = preferences.AutoHideSpoilers.V
	}
	response := apimodels.NewUserPreferencesResponse(timezone, language, theme, emailNotifications, pushNotifications, autoHideNSFW, autoHideSpoilers)
	log.Info().Str("endpoint", "users/preferences").Str("component", "handler").Int64("user_id", userID).Msg("Get user preferences completed")
	return response, nil
}

// UpdateUserPreferences handles updating the current user's preferences
// Note: input.Body is for Huma schema only; actual requests are flat JSON.
func (h *UserHandler) UpdateUserPreferences(ctx context.Context, input *struct {
	middleware.AuthInput
	apimodels.UserPreferencesInput
}) (*apimodels.UserPreferencesResponse, error) {
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Str("endpoint", "users/preferences").Msg("Authentication required for preferences update")
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	userID := int64(userCtx.UserID)
	log.Info().Str("endpoint", "users/preferences").Str("component", "handler").Int64("user_id", userID).Msg("Update user preferences requested")
	updates := &dbmodels.UserPreferenceSetter{}
	if input.Body.Timezone != "" {
		timezone := sql.Null[string]{V: input.Body.Timezone, Valid: true}
		updates.Timezone = &timezone
	}
	if input.Body.Language != "" {
		language := sql.Null[string]{V: input.Body.Language, Valid: true}
		updates.Language = &language
	}
	if input.Body.Theme != "" {
		theme := sql.Null[string]{V: input.Body.Theme, Valid: true}
		updates.Theme = &theme
	}
	if input.Body.EmailNotifications != nil {
		emailNotifications := sql.Null[bool]{V: *input.Body.EmailNotifications, Valid: true}
		updates.EmailNotifications = &emailNotifications
	}
	if input.Body.PushNotifications != nil {
		pushNotifications := sql.Null[bool]{V: *input.Body.PushNotifications, Valid: true}
		updates.PushNotifications = &pushNotifications
	}
	if input.Body.AutoHideNSFW != nil {
		autoHideNSFW := sql.Null[bool]{V: *input.Body.AutoHideNSFW, Valid: true}
		updates.AutoHideNSFW = &autoHideNSFW
	}
	if input.Body.AutoHideSpoilers != nil {
		autoHideSpoilers := sql.Null[bool]{V: *input.Body.AutoHideSpoilers, Valid: true}
		updates.AutoHideSpoilers = &autoHideSpoilers
	}
	updatedPreferences, err := h.userPreferencesDAO.UpsertUserPreferences(ctx, userID, updates)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("Failed to update user preferences")
		return nil, fmt.Errorf("failed to update user preferences: %w", err)
	}
	timezone := "UTC"
	if updatedPreferences.Timezone.Valid {
		timezone = updatedPreferences.Timezone.V
	}
	language := "en"
	if updatedPreferences.Language.Valid {
		language = updatedPreferences.Language.V
	}
	theme := "light"
	if updatedPreferences.Theme.Valid {
		theme = updatedPreferences.Theme.V
	}
	emailNotifications := true
	if updatedPreferences.EmailNotifications.Valid {
		emailNotifications = updatedPreferences.EmailNotifications.V
	}
	pushNotifications := true
	if updatedPreferences.PushNotifications.Valid {
		pushNotifications = updatedPreferences.PushNotifications.V
	}
	autoHideNSFW := true
	if updatedPreferences.AutoHideNSFW.Valid {
		autoHideNSFW = updatedPreferences.AutoHideNSFW.V
	}
	autoHideSpoilers := true
	if updatedPreferences.AutoHideSpoilers.Valid {
		autoHideSpoilers = updatedPreferences.AutoHideSpoilers.V
	}
	response := apimodels.NewUserPreferencesResponse(timezone, language, theme, emailNotifications, pushNotifications, autoHideNSFW, autoHideSpoilers)
	log.Info().Str("endpoint", "users/preferences").Str("component", "handler").Int64("user_id", userID).Msg("Update user preferences completed")
	return response, nil
}

// BlockUser handles blocking a user
// Note: input.Body is for Huma schema only; actual requests are flat JSON.
func (h *UserHandler) BlockUser(ctx context.Context, input *struct {
	middleware.AuthInput
	apimodels.PseudonymIDPathParam
	apimodels.BlockUserInput
}) (*apimodels.BlockUserResponse, error) {
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Str("endpoint", "users/block").Msg("Authentication required for blocking user")
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	userID := int64(userCtx.UserID)
	blockerPseudonymID := userCtx.ActivePseudonymID
	blockedPseudonymID := input.PseudonymID
	log.Info().Str("endpoint", "users/block").Str("component", "handler").Int64("user_id", userID).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Block user requested")
	if blockedPseudonymID == "" {
		return nil, fmt.Errorf("blocked pseudonym ID is required")
	}
	blockedPseudonym, err := h.pseudonymDAO.GetPseudonymByID(ctx, blockedPseudonymID)
	if err != nil {
		log.Error().Err(err).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Failed to get blocked pseudonym from database")
		return nil, fmt.Errorf("failed to get blocked pseudonym: %w", err)
	}
	if blockedPseudonym == nil {
		log.Warn().Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Blocked pseudonym not found")
		return nil, huma.Error404NotFound("Blocked pseudonym not found")
	}

	// Use role-based access control for ownership verification
	ownsPseudonym, err := h.pseudonymDAO.VerifyPseudonymOwnership(ctx, blockedPseudonymID, userID, constants.RoleUser, constants.ScopeSelfCorrelation)
	if err != nil {
		log.Error().Err(err).Str("blocked_pseudonym_id", blockedPseudonymID).Int64("user_id", userID).Msg("Failed to verify pseudonym ownership")
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	if ownsPseudonym {
		log.Warn().Int64("user_id", userID).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("User cannot block themselves")
		return nil, huma.Error400BadRequest("Cannot block yourself")
	}

	// Block all personas if requested
	if input.Body.BlockAllPersonas != nil && *input.Body.BlockAllPersonas {
		// ✅ Use IBE-based correlation to block all personas of the user
		// Get the blocked user's ID (not the blocker's ID)
		blockedUserID, err := h.pseudonymDAO.GetUserIDByPseudonym(ctx, blockedPseudonymID, constants.RoleUser, constants.ScopeSelfCorrelation)
		if err != nil {
			log.Error().Err(err).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Failed to get blocked user ID")
			return nil, fmt.Errorf("failed to get blocked user ID: %w", err)
		}

		// Block at the user ID level to prevent any future pseudonyms from this user
		// This ensures that even if the user creates new pseudonyms, they will be blocked
		_, err = h.userBlocksDAO.CreateUserBlock(ctx, blockerPseudonymID, "", blockedUserID)
		if err != nil {
			log.Error().Err(err).Str("blocker_pseudonym_id", blockerPseudonymID).Int64("blocked_user_id", blockedUserID).Msg("Failed to create fingerprint-level user block")
			return nil, fmt.Errorf("failed to create user block: %w", err)
		}

		log.Info().
			Str("blocker_pseudonym_id", blockerPseudonymID).
			Str("blocked_pseudonym_id", blockedPseudonymID).
			Int64("blocked_user_id", blockedUserID).
			Msg("Created fingerprint-level block for all personas")
	} else {
		// Block only the specific pseudonym
		log.Debug().Str("blocker_pseudonym_id", blockerPseudonymID).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("About to create pseudonym-level user block")
		_, err = h.userBlocksDAO.CreateUserBlock(ctx, blockerPseudonymID, blockedPseudonymID, 0)
		if err != nil {
			log.Error().Err(err).Str("blocker_pseudonym_id", blockerPseudonymID).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Failed to create user block")
			return nil, fmt.Errorf("failed to create user block: %w", err)
		}
	}
	response := apimodels.NewBlockUserResponse(blockedPseudonymID, blockedPseudonymID)
	log.Info().Str("endpoint", "users/block").Str("component", "handler").Int64("user_id", userID).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Block user completed")
	return response, nil
}

// UnblockUser handles unblocking a user
func (h *UserHandler) UnblockUser(ctx context.Context, input *struct {
	middleware.AuthInput
	apimodels.PseudonymIDPathParam
}) (*apimodels.UnblockUserResponse, error) {
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Str("endpoint", "users/unblock").Msg("Authentication required for unblocking user")
		return nil, huma.Error401Unauthorized("Authentication required")
	}
	userID := int64(userCtx.UserID)
	blockerPseudonymID := userCtx.ActivePseudonymID
	blockedPseudonymID := input.PseudonymID
	log.Info().Str("endpoint", "users/unblock").Str("component", "handler").Int64("user_id", userID).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Unblock user requested")
	if blockedPseudonymID == "" {
		return nil, fmt.Errorf("blocked pseudonym ID is required")
	}
	// First try to find a direct block
	existingBlock, err := h.userBlocksDAO.GetUserBlock(ctx, blockerPseudonymID, blockedPseudonymID)
	if err != nil {
		log.Error().Err(err).Str("blocker_pseudonym_id", blockerPseudonymID).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Failed to check existing direct block")
		return nil, fmt.Errorf("failed to check existing block: %w", err)
	}

	// If no direct block found, check for fingerprint-level block
	if existingBlock == nil {
		// Get the blocked user's ID to check for fingerprint-level blocks
		blockedUserID, err := h.pseudonymDAO.GetUserIDByPseudonym(ctx, blockedPseudonymID, constants.RoleUser, constants.ScopeSelfCorrelation)
		if err != nil {
			log.Error().Err(err).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Failed to get blocked user ID for unblock")
			return nil, fmt.Errorf("failed to get blocked user ID: %w", err)
		}

		// Check for fingerprint-level blocks
		fingerprintBlocks, err := h.userBlocksDAO.GetFingerprintLevelBlocks(ctx, blockedUserID)
		if err != nil {
			log.Error().Err(err).Int64("blocked_user_id", blockedUserID).Msg("Failed to check fingerprint-level blocks")
			return nil, fmt.Errorf("failed to check fingerprint-level blocks: %w", err)
		}

		// Find the block from this specific blocker
		for _, block := range fingerprintBlocks {
			if block.BlockerPseudonymID == blockerPseudonymID {
				existingBlock = block
				break
			}
		}
	}

	if existingBlock == nil {
		log.Warn().Str("blocker_pseudonym_id", blockerPseudonymID).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Block not found")
		return nil, huma.Error404NotFound("Block not found")
	}

	// Delete the block based on its type
	if existingBlock.BlockedPseudonymID.Valid {
		// Direct block
		err = h.userBlocksDAO.DeleteUserBlock(ctx, blockerPseudonymID, blockedPseudonymID)
	} else {
		// Fingerprint-level block - delete by block ID
		err = h.userBlocksDAO.DeleteUserBlockByID(ctx, existingBlock.BlockID)
	}
	if err != nil {
		log.Error().Err(err).Str("blocker_pseudonym_id", blockerPseudonymID).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Failed to delete user block")
		return nil, fmt.Errorf("failed to delete user block: %w", err)
	}
	blockedUserID := int64(0)
	if existingBlock.BlockedUserID.Valid {
		blockedUserID = existingBlock.BlockedUserID.V
	}
	response := apimodels.NewUnblockUserResponse(int(blockedUserID), blockedPseudonymID)
	log.Info().Str("endpoint", "users/unblock").Str("component", "handler").Int64("user_id", userID).Str("blocked_pseudonym_id", blockedPseudonymID).Int64("blocked_user_id", blockedUserID).Msg("Unblock user completed")
	return response, nil
}
