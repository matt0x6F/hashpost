package handlers

import (
	"context"
	"fmt"
	"strings"

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

// SearchHandler handles search requests
type SearchHandler struct {
	db            bob.Executor
	postDAO       dao.PostDAOInterface
	userDAO       dao.UserDAOInterface
	subforumDAO   dao.SubforumDAOInterface
	pseudonymDAO  dao.PseudonymDAOInterface
	permissionDAO dao.PermissionDAOInterface
	ibeSystem     *ibe.IBESystem
}

// NewSearchHandler creates a new search handler
// For production use, pass a database executor, nil for all DAO parameters, and the IBE system
// For testing, pass nil for db and provide mocked DAOs
func NewSearchHandler(
	db bob.Executor,
	postDAO dao.PostDAOInterface,
	userDAO dao.UserDAOInterface,
	subforumDAO dao.SubforumDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
	permissionDAO dao.PermissionDAOInterface,
	ibeSystem *ibe.IBESystem,
) *SearchHandler {
	if db != nil {
		// Production mode - create real DAOs with provided IBE system
		identityMappingDAO := dao.NewIdentityMappingDAO(db)
		roleKeyDAO := dao.NewRoleKeyDAO(db, ibeSystem)
		userBlocksDAO := dao.NewUserBlocksDAO(db)
		userDAO := dao.NewUserDAO(db)
		pseudonymDAO := dao.NewPseudonymDAO(db, ibeSystem, identityMappingDAO, userDAO, roleKeyDAO, userBlocksDAO)

		return &SearchHandler{
			db:            db,
			postDAO:       dao.NewPostDAO(db),
			userDAO:       userDAO,
			subforumDAO:   dao.NewSubforumDAO(db),
			pseudonymDAO:  pseudonymDAO,
			permissionDAO: dao.NewPermissionDAO(db),
			ibeSystem:     ibeSystem,
		}
	}

	// Test mode - use provided mocked DAOs
	return &SearchHandler{
		postDAO:       postDAO,
		userDAO:       userDAO,
		subforumDAO:   subforumDAO,
		pseudonymDAO:  pseudonymDAO,
		permissionDAO: permissionDAO,
	}
}

// SearchPosts handles searching for posts across all subforums
func (h *SearchHandler) SearchPosts(ctx context.Context, input *models.SearchPostsInput) (*models.SearchPostsResponse, error) {
	// Extract user from context (search can work anonymously, so we don't embed AuthInput)
	user, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		// Search can work for anonymous users, so we'll continue without user context
		log.Debug().Msg("No user context found for search, proceeding as anonymous user")
	}

	// Build log fields
	logFields := log.Info().
		Str("endpoint", "search/posts").
		Str("component", "handler").
		Str("query", input.Query).
		Str("subforum", input.Subforum).
		Str("author", input.Author).
		Str("sort", input.Sort).
		Str("time", input.Time)

	// Add user_id to log if user context is available
	if user != nil {
		logFields = logFields.Int64("user_id", user.UserID)
	}

	logFields.Msg("Search posts requested")

	// Build search query
	query := input.Query
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Get posts from database with proper search
	posts, err := h.searchPosts(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to search posts")
		return nil, fmt.Errorf("failed to search posts: %w", err)
	}

	// Convert database posts to API posts
	apiPosts := make([]models.SearchPost, 0, len(posts))
	for _, post := range posts {
		// Handle nullable fields
		createdAt := ""
		if post.CreatedAt.Valid {
			createdAt = post.CreatedAt.V.Format("2006-01-02T15:04:05Z")
		}

		// Get subforum info
		subforum, err := h.subforumDAO.GetSubforumByID(ctx, post.SubforumID)
		if err != nil {
			log.Warn().Err(err).Int32("subforum_id", post.SubforumID).Msg("Failed to get subforum info")
			continue
		}

		// Get author info - use pseudonym ID to get display name
		pseudonym, err := h.pseudonymDAO.GetPseudonymByID(ctx, post.PseudonymID)
		if err != nil {
			log.Warn().Err(err).Str("pseudonym_id", post.PseudonymID).Msg("Failed to get author info")
			continue
		}

		// Handle nullable fields for post
		content := ""
		if post.Content.Valid {
			content = post.Content.V
		}

		score := 0
		if post.Score.Valid {
			score = int(post.Score.V)
		}

		commentCount := 0
		if post.CommentCount.Valid {
			commentCount = int(post.CommentCount.V)
		}

		apiPosts = append(apiPosts, models.SearchPost{
			PostID:       int(post.PostID),
			Title:        post.Title,
			Content:      content,
			Score:        score,
			CommentCount: commentCount,
			CreatedAt:    createdAt,
			Author: models.Author{
				PseudonymID: post.PseudonymID,
				DisplayName: pseudonym.DisplayName,
			},
			Subforum: models.SubforumInfo{
				Name:        subforum.Name,
				DisplayName: subforum.DisplayName,
			},
		})
	}

	// Get actual total from database
	total, err := h.countSearchPosts(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count search posts")
		return nil, fmt.Errorf("failed to count search posts: %w", err)
	}

	// Handle pagination
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	response := models.NewSearchPostsResponse(input.Query, apiPosts, page, limit, int(total))

	// Build completion log fields
	completionLog := log.Info().
		Str("endpoint", "search/posts").
		Str("component", "handler").
		Int("count", len(apiPosts)).
		Int("total", int(total))

	// Add user_id to log if user context is available
	if user != nil {
		completionLog = completionLog.Int64("user_id", user.UserID)
	}

	completionLog.Msg("Search posts completed")

	return response, nil
}

// searchPosts implements the actual search logic for posts
func (h *SearchHandler) searchPosts(ctx context.Context, input *models.SearchPostsInput) ([]*dbmodels.Post, error) {
	// For now, implement a simple search by getting all posts and filtering
	// In a production system, this would use full-text search (e.g., PostgreSQL FTS)

	// Get all subforums to search across
	subforums, err := h.subforumDAO.ListSubforums(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get subforums: %w", err)
	}

	var allPosts []*dbmodels.Post
	query := strings.ToLower(input.Query)

	// Search across all subforums
	for _, subforum := range subforums {
		// Skip if specific subforum filter is applied
		if input.Subforum != "" && subforum.Name != input.Subforum {
			continue
		}

		// Get posts from this subforum
		posts, err := h.postDAO.GetPostsBySubforum(ctx, subforum.SubforumID, 1, 100, "created_at", true)
		if err != nil {
			log.Warn().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to get posts from subforum")
			continue
		}

		// Filter posts by search query
		for _, post := range posts {
			// Check if post matches search criteria
			if h.postMatchesSearch(post, query, input.Author) {
				allPosts = append(allPosts, post)
			}
		}
	}

	// Apply sorting
	allPosts = h.sortPosts(allPosts, input.Sort)

	// Apply pagination
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	start := (page - 1) * limit
	end := start + limit

	if start >= len(allPosts) {
		return []*dbmodels.Post{}, nil
	}

	if end > len(allPosts) {
		end = len(allPosts)
	}

	return allPosts[start:end], nil
}

// postMatchesSearch checks if a post matches the search criteria
func (h *SearchHandler) postMatchesSearch(post *dbmodels.Post, query, author string) bool {
	// Check title
	if strings.Contains(strings.ToLower(post.Title), query) {
		return true
	}

	// Check content
	if post.Content.Valid && strings.Contains(strings.ToLower(post.Content.V), query) {
		return true
	}

	// Check author if specified
	if author != "" {
		// This would need to be enhanced to check pseudonym display name
		// For now, we'll skip author filtering
	}

	return false
}

// sortPosts sorts posts according to the specified sort order
func (h *SearchHandler) sortPosts(posts []*dbmodels.Post, sort string) []*dbmodels.Post {
	// For now, return posts as-is
	// In a production system, this would implement proper sorting
	// based on score, date, relevance, etc.
	return posts
}

// countSearchPosts counts the total number of posts matching the search criteria
func (h *SearchHandler) countSearchPosts(ctx context.Context, input *models.SearchPostsInput) (int64, error) {
	// For now, return a simple count
	// In a production system, this would use the same search logic as searchPosts
	// but only count the results

	// Get all subforums
	subforums, err := h.subforumDAO.ListSubforums(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to get subforums: %w", err)
	}

	var total int64
	query := strings.ToLower(input.Query)

	// Count posts across all subforums
	for _, subforum := range subforums {
		// Skip if specific subforum filter is applied
		if input.Subforum != "" && subforum.Name != input.Subforum {
			continue
		}

		// Get posts from this subforum
		posts, err := h.postDAO.GetPostsBySubforum(ctx, subforum.SubforumID, 1, 1000, "created_at", true)
		if err != nil {
			log.Warn().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to get posts from subforum for counting")
			continue
		}

		// Count matching posts
		for _, post := range posts {
			if h.postMatchesSearch(post, query, input.Author) {
				total++
			}
		}
	}

	return total, nil
}

// SearchUsers handles searching for users by display name
// This endpoint requires authentication and platform admin privileges
func (h *SearchHandler) SearchUsers(ctx context.Context, input *models.SearchUsersInput) (*models.SearchUsersResponse, error) {
	// Extract user from Huma input - authentication is required
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for user search")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Check if user has platform admin capability via database
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, user.UserID, user.ActivePseudonymID, constants.CapabilitySystemAdmin, nil)
	if err != nil {
		log.Error().Err(err).Int64("user_id", user.UserID).Msg("Failed to check platform admin capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().
			Int64("user_id", user.UserID).
			Msg("Platform admin capability required for user search")
		return nil, fmt.Errorf("insufficient permissions: platform admin capability required")
	}

	// Build request log fields
	requestLog := log.Info().
		Str("endpoint", "search/users").
		Str("component", "handler").
		Str("query", input.Query)

	// Add user_id to log if user context is available
	if user != nil {
		requestLog = requestLog.Int64("user_id", user.UserID)
	}

	requestLog.Msg("Search users requested")

	// Build search query
	query := input.Query
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Get users and their pseudonyms from database with proper search
	users, pseudonyms, err := h.searchUsers(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to search users")
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert database users and pseudonyms to API users
	apiUsers := make([]models.SearchUser, 0, len(users))
	for _, user := range users {
		// Handle nullable fields
		createdAt := ""
		if user.CreatedAt.Valid {
			createdAt = user.CreatedAt.V.Format("2006-01-02T15:04:05Z")
		}

		// Get the pseudonym that was already found during search
		// Since users and pseudonyms are paired in the same order, we can use the same index
		pseudonym := pseudonyms[len(apiUsers)] // Use current length as index

		// Get all pseudonyms for this user to populate the pseudonyms array
		userPseudonyms, err := h.pseudonymDAO.GetPseudonymsByRealIdentityDirect(ctx, user.Email)
		if err != nil {
			log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to get user pseudonyms for API response")
			userPseudonyms = []*dbmodels.Pseudonym{}
		}

		// Convert database pseudonyms to API pseudonyms
		apiPseudonyms := make([]struct {
			ID          string `json:"id" example:"pseudo-123"`
			DisplayName string `json:"display_name" example:"john_doe"`
			IsDefault   bool   `json:"is_default" example:"true"`
			CreatedAt   string `json:"created_at" example:"2024-01-01T12:00:00Z"`
		}, 0, len(userPseudonyms))

		for _, p := range userPseudonyms {
			pseudoCreatedAt := ""
			if p.CreatedAt.Valid {
				pseudoCreatedAt = p.CreatedAt.V.Format("2006-01-02T15:04:05Z")
			}

			apiPseudonyms = append(apiPseudonyms, struct {
				ID          string `json:"id" example:"pseudo-123"`
				DisplayName string `json:"display_name" example:"john_doe"`
				IsDefault   bool   `json:"is_default" example:"true"`
				CreatedAt   string `json:"created_at" example:"2024-01-01T12:00:00Z"`
			}{
				ID:          p.PseudonymID,
				DisplayName: p.DisplayName,
				IsDefault:   p.IsDefault,
				CreatedAt:   pseudoCreatedAt,
			})
		}

		// Calculate karma score
		karmaScore, err := h.calculateKarmaScore(ctx)
		if err != nil {
			log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to calculate karma score")
			karmaScore = 0
		}

		apiUsers = append(apiUsers, models.SearchUser{
			PseudonymID: pseudonym.PseudonymID,
			DisplayName: pseudonym.DisplayName,
			KarmaScore:  karmaScore,
			CreatedAt:   createdAt,
			Email:       user.Email,
			UserID:      user.UserID,
			Pseudonyms:  apiPseudonyms,
		})
	}

	// Get actual total from database
	total, err := h.countSearchUsers(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count search users")
		return nil, fmt.Errorf("failed to count search users: %w", err)
	}

	// Handle pagination
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	response := models.NewSearchUsersResponse(input.Query, apiUsers, page, limit, int(total))

	// Build completion log fields
	completionLog := log.Info().
		Str("endpoint", "search/users").
		Str("component", "handler").
		Int("count", len(apiUsers)).
		Int("total", int(total))

	// Add user_id to log if user context is available
	if user != nil {
		completionLog = completionLog.Int64("user_id", user.UserID)
	}

	completionLog.Msg("Search users completed")

	return response, nil
}

// searchUsers implements the actual search logic for users
func (h *SearchHandler) searchUsers(ctx context.Context, input *models.SearchUsersInput) ([]*dbmodels.User, []*dbmodels.Pseudonym, error) {
	// Get all users
	users, err := h.userDAO.ListUsers(ctx, 1000, 0)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get users: %w", err)
	}

	log.Info().Int("total_users_found", len(users)).Msg("Search: Retrieved users from database")

	query := strings.ToLower(input.Query)
	var matchingUsers []*dbmodels.User
	var matchingPseudonyms []*dbmodels.Pseudonym

	// Filter users by search criteria
	for _, user := range users {
		// Check if user email matches query first (exact match for emails)
		if strings.Contains(strings.ToLower(user.Email), query) {
			log.Debug().Int64("user_id", user.UserID).Str("email", user.Email).Msg("Search: User matched by email")

			// Debug: Show IBE system salt and fingerprint generation
			log.Debug().
				Str("ibe_salt", h.pseudonymDAO.GetIBESystemSalt()).
				Str("email", user.Email).
				Str("generated_fingerprint", h.pseudonymDAO.GenerateFingerprintForEmail(user.Email)).
				Msg("Search: IBE system debug info")

			// Get pseudonyms for the user using the direct method (bypasses role-based access control)
			pseudonyms, err := h.pseudonymDAO.GetPseudonymsByRealIdentityDirect(ctx, user.Email)
			if err != nil {
				log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to get user pseudonyms for search")
				continue
			}

			if len(pseudonyms) == 0 {
				log.Warn().Int64("user_id", user.UserID).Msg("Search: No pseudonyms found for user")
				continue
			}

			// Use the first pseudonym (or find the default one)
			var pseudonym *dbmodels.Pseudonym
			for _, p := range pseudonyms {
				if p.IsDefault {
					pseudonym = p
					break
				}
			}
			if pseudonym == nil {
				pseudonym = pseudonyms[0] // Use first if no default found
			}

			matchingUsers = append(matchingUsers, user)
			matchingPseudonyms = append(matchingPseudonyms, pseudonym)
			continue
		}

		// Get user's primary role for pseudonym access
		// For search results, we'll use the default "user" role since we want consistent results
		pseudonyms, err := h.pseudonymDAO.GetPseudonymsByRealIdentityDirect(ctx, user.Email)
		if err != nil {
			log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to get user pseudonyms for search")
			continue
		}

		if len(pseudonyms) == 0 {
			log.Warn().Int64("user_id", user.UserID).Msg("Search: No pseudonyms found for user")
			continue
		}

		// Find a pseudonym that matches the display name query
		var matchingPseudonym *dbmodels.Pseudonym
		for _, p := range pseudonyms {
			if strings.Contains(strings.ToLower(p.DisplayName), query) {
				matchingPseudonym = p
				break
			}
		}

		if matchingPseudonym != nil {
			log.Debug().Int64("user_id", user.UserID).Str("display_name", matchingPseudonym.DisplayName).Msg("Search: User matched by display name")
			matchingUsers = append(matchingUsers, user)
			matchingPseudonyms = append(matchingPseudonyms, matchingPseudonym)
		}
	}

	log.Info().Int("matching_users", len(matchingUsers)).Str("query", query).Msg("Search: Found matching users")

	// Apply pagination
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	start := (page - 1) * limit
	end := start + limit

	if start >= len(matchingUsers) {
		return []*dbmodels.User{}, []*dbmodels.Pseudonym{}, nil
	}

	if end > len(matchingUsers) {
		end = len(matchingUsers)
	}

	usersResult := matchingUsers[start:end]
	pseudonymsResult := matchingPseudonyms[start:end]

	log.Info().Int("result_count", len(usersResult)).Int("page", page).Int("limit", limit).Msg("Search: Returning paginated results")

	return usersResult, pseudonymsResult, nil
}

// countSearchUsers counts the total number of users matching the search criteria
func (h *SearchHandler) countSearchUsers(ctx context.Context, input *models.SearchUsersInput) (int64, error) {
	// Get all users
	users, err := h.userDAO.ListUsers(ctx, 1000, 0)
	if err != nil {
		return 0, fmt.Errorf("failed to get users: %w", err)
	}

	query := strings.ToLower(input.Query)
	var total int64

	log.Info().Int("total_users_for_counting", len(users)).Str("query", query).Msg("Search: Counting matching users")

	// Count matching users
	for _, user := range users {
		// Check if user email matches query first (exact match for emails)
		if strings.Contains(strings.ToLower(user.Email), query) {
			log.Debug().Int64("user_id", user.UserID).Str("email", user.Email).Msg("Search: Count: User matched by email")

			// Debug: Show IBE system salt and fingerprint generation
			log.Debug().
				Str("ibe_salt", h.pseudonymDAO.GetIBESystemSalt()).
				Str("email", user.Email).
				Str("generated_fingerprint", h.pseudonymDAO.GenerateFingerprintForEmail(user.Email)).
				Msg("Search: Count: IBE system debug info")

			// Get pseudonyms for the user using the direct method
			pseudonyms, err := h.pseudonymDAO.GetPseudonymsByRealIdentityDirect(ctx, user.Email)
			if err != nil {
				log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to get user pseudonyms for counting")
				continue
			}

			if len(pseudonyms) > 0 {
				total++
			}
			continue
		}

		// Get pseudonyms for the user using the direct method
		pseudonyms, err := h.pseudonymDAO.GetPseudonymsByRealIdentityDirect(ctx, user.Email)
		if err != nil {
			log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to get user pseudonyms for counting")
			continue
		}

		if len(pseudonyms) == 0 {
			log.Warn().Int64("user_id", user.UserID).Msg("Search: Count: No pseudonyms found for user")
			continue
		}

		// Find a pseudonym that matches the display name query
		for _, p := range pseudonyms {
			if strings.Contains(strings.ToLower(p.DisplayName), query) {
				log.Debug().Int64("user_id", user.UserID).Str("display_name", p.DisplayName).Msg("Search: Count: User matched by display name")
				total++
				break
			}
		}
	}

	log.Info().Int64("total_matching_users", total).Str("query", query).Msg("Search: Count completed")

	return total, nil
}

// SearchPseudonyms handles searching for pseudonyms directly
// This endpoint requires authentication and platform admin privileges
func (h *SearchHandler) SearchPseudonyms(ctx context.Context, input *models.SearchPseudonymsInput) (*models.SearchPseudonymsResponse, error) {
	// Extract user from Huma input - authentication is required
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for pseudonym search")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Check if user has platform admin capability via database
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, user.UserID, user.ActivePseudonymID, constants.CapabilitySystemAdmin, nil)
	if err != nil {
		log.Error().Err(err).Int64("user_id", user.UserID).Msg("Failed to check platform admin capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().
			Int64("user_id", user.UserID).
			Msg("Platform admin capability required for pseudonym search")
		return nil, fmt.Errorf("insufficient permissions: platform admin capability required")
	}

	// Build request log fields
	requestLog := log.Info().
		Str("endpoint", "search/pseudonyms").
		Str("component", "handler").
		Str("query", input.Query)

	// Add user_id to log if user context is available
	if user != nil {
		requestLog = requestLog.Int64("user_id", user.UserID)
	}

	requestLog.Msg("Search pseudonyms requested")

	// Build search query
	query := input.Query
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Get pseudonyms from database with proper search
	pseudonyms, err := h.searchPseudonyms(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to search pseudonyms")
		return nil, fmt.Errorf("failed to search pseudonyms: %w", err)
	}

	// Convert database pseudonyms to API pseudonyms
	apiPseudonyms := make([]models.SearchPseudonym, 0, len(pseudonyms))
	for _, pseudonym := range pseudonyms {
		// Handle nullable fields
		createdAt := ""
		if pseudonym.CreatedAt.Valid {
			createdAt = pseudonym.CreatedAt.V.Format("2006-01-02T15:04:05Z")
		}

		lastActiveAt := ""
		if pseudonym.LastActiveAt.Valid {
			lastActiveAt = pseudonym.LastActiveAt.V.Format("2006-01-02T15:04:05Z")
		}

		bio := ""
		if pseudonym.Bio.Valid {
			bio = pseudonym.Bio.V
		}

		avatarURL := ""
		if pseudonym.AvatarURL.Valid {
			avatarURL = pseudonym.AvatarURL.V
		}

		websiteURL := ""
		if pseudonym.WebsiteURL.Valid {
			websiteURL = pseudonym.WebsiteURL.V
		}

		slug := ""
		if pseudonym.Slug.Valid {
			slug = pseudonym.Slug.V
		}

		karmaScore := 0
		if pseudonym.KarmaScore.Valid {
			karmaScore = int(pseudonym.KarmaScore.V)
		}

		isActive := true
		if pseudonym.IsActive.Valid {
			isActive = pseudonym.IsActive.V
		}

		showKarma := true
		if pseudonym.ShowKarma.Valid {
			showKarma = pseudonym.ShowKarma.V
		}

		allowDirectMessages := true
		if pseudonym.AllowDirectMessages.Valid {
			allowDirectMessages = pseudonym.AllowDirectMessages.V
		}

		apiPseudonyms = append(apiPseudonyms, models.SearchPseudonym{
			PseudonymID:         pseudonym.PseudonymID,
			DisplayName:         pseudonym.DisplayName,
			KarmaScore:          karmaScore,
			CreatedAt:           createdAt,
			LastActiveAt:        lastActiveAt,
			IsActive:            isActive,
			Bio:                 bio,
			AvatarURL:           avatarURL,
			WebsiteURL:          websiteURL,
			ShowKarma:           showKarma,
			AllowDirectMessages: allowDirectMessages,
			IsDefault:           pseudonym.IsDefault,
			Slug:                slug,
		})
	}

	// Get actual total from database
	total, err := h.countSearchPseudonyms(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count search pseudonyms")
		return nil, fmt.Errorf("failed to count search pseudonyms: %w", err)
	}

	// Handle pagination
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	response := models.NewSearchPseudonymsResponse(input.Query, apiPseudonyms, page, limit, int(total))

	// Build completion log fields
	completionLog := log.Info().
		Str("endpoint", "search/pseudonyms").
		Str("component", "handler").
		Int("count", len(apiPseudonyms)).
		Int("total", int(total))

	// Add user_id to log if user context is available
	if user != nil {
		completionLog = completionLog.Int64("user_id", user.UserID)
	}

	completionLog.Msg("Search pseudonyms completed")

	return response, nil
}

// PublicSearchPseudonyms handles searching for pseudonyms for co-moderator selection
// This endpoint is public and doesn't require authentication
func (h *SearchHandler) PublicSearchPseudonyms(ctx context.Context, input *models.PublicSearchPseudonymsInput) (*models.PublicSearchPseudonymsResponse, error) {
	// Build request log fields
	requestLog := log.Info().
		Str("endpoint", "search/pseudonyms/public").
		Str("component", "handler").
		Str("query", input.Query)

	requestLog.Msg("Public search pseudonyms requested")

	// Build search query
	query := input.Query
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	// Get pseudonyms from database with proper search
	pseudonyms, err := h.searchPseudonymsPublic(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to search pseudonyms publicly")
		return nil, fmt.Errorf("failed to search pseudonyms: %w", err)
	}

	// Convert database pseudonyms to public API pseudonyms (limited info)
	publicPseudonyms := make([]models.PublicSearchPseudonym, 0, len(pseudonyms))
	for _, pseudonym := range pseudonyms {
		// Only include active pseudonyms for co-moderator selection
		if !pseudonym.IsActive.Valid || !pseudonym.IsActive.V {
			continue
		}

		bio := ""
		if pseudonym.Bio.Valid {
			bio = pseudonym.Bio.V
		}

		slug := ""
		if pseudonym.Slug.Valid {
			slug = pseudonym.Slug.V
		}

		publicPseudonyms = append(publicPseudonyms, models.PublicSearchPseudonym{
			PseudonymID: pseudonym.PseudonymID,
			DisplayName: pseudonym.DisplayName,
			Slug:        slug,
			Bio:         bio,
			IsActive:    true,
		})
	}

	// Get actual total from database
	total, err := h.countSearchPseudonymsPublic(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count public search pseudonyms")
		return nil, fmt.Errorf("failed to count search pseudonyms: %w", err)
	}

	// Handle pagination
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	response := models.NewPublicSearchPseudonymsResponse(input.Query, publicPseudonyms, page, limit, int(total))

	// Build completion log fields
	completionLog := log.Info().
		Str("endpoint", "search/pseudonyms/public").
		Str("component", "handler").
		Int("count", len(publicPseudonyms)).
		Int("total", int(total))

	completionLog.Msg("Public search pseudonyms completed")

	return response, nil
}

// searchPseudonyms implements the actual search logic for pseudonyms
func (h *SearchHandler) searchPseudonyms(ctx context.Context, input *models.SearchPseudonymsInput) ([]*dbmodels.Pseudonym, error) {
	// Use DAO method instead of direct database access
	return h.pseudonymDAO.SearchPseudonyms(ctx, input.Query, input.Page, input.Limit)
}

// searchPseudonymsPublic implements the actual search logic for public pseudonym search
func (h *SearchHandler) searchPseudonymsPublic(ctx context.Context, input *models.PublicSearchPseudonymsInput) ([]*dbmodels.Pseudonym, error) {
	// Use DAO method instead of direct database access
	return h.pseudonymDAO.SearchPseudonymsPublic(ctx, input.Query, input.Page, input.Limit)
}

// countSearchPseudonyms counts the total number of pseudonyms matching the search criteria
func (h *SearchHandler) countSearchPseudonyms(ctx context.Context, input *models.SearchPseudonymsInput) (int64, error) {
	// Use DAO method instead of direct database access
	return h.pseudonymDAO.CountSearchPseudonyms(ctx, input.Query)
}

// countSearchPseudonymsPublic counts the total number of pseudonyms matching the search criteria for public search
func (h *SearchHandler) countSearchPseudonymsPublic(ctx context.Context, input *models.PublicSearchPseudonymsInput) (int64, error) {
	// Use DAO method instead of direct database access
	return h.pseudonymDAO.CountSearchPseudonymsPublic(ctx, input.Query)
}

// calculateKarmaScore calculates the karma score for a user
func (h *SearchHandler) calculateKarmaScore(ctx context.Context) (int, error) {
	// Calculate karma based on post scores
	// This is a simplified implementation - in production, this would be more sophisticated

	// Get user's posts
	posts, err := h.postDAO.GetPostsBySubforum(ctx, 0, 1, 1000, "created_at", true) // 0 means all subforums
	if err != nil {
		return 0, fmt.Errorf("failed to get posts: %w", err)
	}

	// Calculate total score
	totalScore := 0
	for _, post := range posts {
		if post.Score.Valid {
			totalScore += int(post.Score.V)
		}
	}

	return totalScore, nil
}
