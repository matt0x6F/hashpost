package handlers

import (
	"context"
	"fmt"
	"strings"

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
	db           bob.Executor
	postDAO      dao.PostDAOInterface
	userDAO      dao.UserDAOInterface
	subforumDAO  dao.SubforumDAOInterface
	pseudonymDAO dao.PseudonymDAOInterface
	ibeSystem    *ibe.IBESystem
}

// NewSearchHandler creates a new search handler
// For production use, pass a database executor and nil for all DAO parameters
// For testing, pass nil for db and provide mocked DAOs
func NewSearchHandler(
	db bob.Executor,
	postDAO dao.PostDAOInterface,
	userDAO dao.UserDAOInterface,
	subforumDAO dao.SubforumDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
) *SearchHandler {
	if db != nil {
		// Production mode - create real DAOs
		ibeSystem := ibe.NewTestIBESystem()
		identityMappingDAO := dao.NewIdentityMappingDAO(db)
		roleKeyDAO := dao.NewRoleKeyDAO(db)
		userBlocksDAO := dao.NewUserBlocksDAO(db)
		userDAO := dao.NewUserDAO(db)
		securePseudonymDAO := dao.NewPseudonymDAO(db, ibeSystem, identityMappingDAO, userDAO, roleKeyDAO, userBlocksDAO)

		return &SearchHandler{
			db:           db,
			postDAO:      dao.NewPostDAO(db),
			userDAO:      userDAO,
			subforumDAO:  dao.NewSubforumDAO(db),
			pseudonymDAO: securePseudonymDAO,
			ibeSystem:    ibeSystem,
		}
	}

	// Test mode - use provided mocked DAOs
	return &SearchHandler{
		postDAO:      postDAO,
		userDAO:      userDAO,
		subforumDAO:  subforumDAO,
		pseudonymDAO: pseudonymDAO,
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
	// Extract user from context - authentication is required
	user, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for user search")
		return nil, fmt.Errorf("authentication required: %w", err)
	}

	// Check if user has platform admin role
	if !user.HasRole("platform_admin") {
		log.Warn().
			Int64("user_id", user.UserID).
			Str("roles", strings.Join(user.Roles, ",")).
			Msg("Platform admin role required for user search")
		return nil, fmt.Errorf("insufficient permissions: platform admin role required")
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

	// Get users from database with proper search
	users, err := h.searchUsers(ctx, input)
	if err != nil {
		log.Error().Err(err).Msg("Failed to search users")
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	// Convert database users to API users
	apiUsers := make([]models.SearchUser, 0, len(users))
	for _, user := range users {
		// Handle nullable fields
		createdAt := ""
		if user.CreatedAt.Valid {
			createdAt = user.CreatedAt.V.Format("2006-01-02T15:04:05Z")
		}

		// Get default pseudonym for the user
		pseudonym, err := h.pseudonymDAO.GetDefaultPseudonymByUserID(ctx, user.UserID, "user", "global")
		if err != nil {
			log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to get user pseudonym")
			continue
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
func (h *SearchHandler) searchUsers(ctx context.Context, input *models.SearchUsersInput) ([]*dbmodels.User, error) {
	// Get all users
	users, err := h.userDAO.ListUsers(ctx, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	query := strings.ToLower(input.Query)
	var matchingUsers []*dbmodels.User

	// Filter users by search criteria
	for _, user := range users {
		// Get default pseudonym for the user
		pseudonym, err := h.pseudonymDAO.GetDefaultPseudonymByUserID(ctx, user.UserID, "user", "global")
		if err != nil {
			log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to get user pseudonym for search")
			continue
		}

		// Check if pseudonym display name matches query
		if strings.Contains(strings.ToLower(pseudonym.DisplayName), query) {
			matchingUsers = append(matchingUsers, user)
		}
	}

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
		return []*dbmodels.User{}, nil
	}

	if end > len(matchingUsers) {
		end = len(matchingUsers)
	}

	return matchingUsers[start:end], nil
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

	// Count matching users
	for _, user := range users {
		// Get default pseudonym for the user
		pseudonym, err := h.pseudonymDAO.GetDefaultPseudonymByUserID(ctx, user.UserID, "user", "global")
		if err != nil {
			log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to get user pseudonym for counting")
			continue
		}

		// Check if pseudonym display name matches query
		if strings.Contains(strings.ToLower(pseudonym.DisplayName), query) {
			total++
		}
	}

	return total, nil
}

// calculateKarmaScore calculates the karma score for a user
func (h *SearchHandler) calculateKarmaScore(ctx context.Context) (int, error) {
	// Calculate karma based on post and comment scores
	// This is a simplified implementation - in production, this would be more sophisticated

	// Get user's posts
	posts, err := h.postDAO.GetPostsBySubforum(ctx, 0, 1, 1000, "created_at", true) // 0 means all subforums
	if err != nil {
		return 0, fmt.Errorf("failed to get posts for karma calculation: %w", err)
	}

	var totalScore int
	for _, post := range posts {
		// Check if this post belongs to the user
		// This would need to be enhanced to check pseudonym ownership
		if post.Score.Valid {
			totalScore += int(post.Score.V)
		}
	}

	// For now, return a simple score
	// In production, this would consider:
	// - Post scores
	// - Comment scores
	// - Time decay
	// - Community standing
	return totalScore, nil
}
