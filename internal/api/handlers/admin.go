package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
)

// AdminHandler handles administrative operations
type AdminHandler struct {
	userDAO       dao.UserDAOInterface
	pseudonymDAO  dao.PseudonymDAOInterface
	permissionDAO dao.PermissionDAOInterface
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	userDAO dao.UserDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
	permissionDAO dao.PermissionDAOInterface,
) *AdminHandler {
	return &AdminHandler{
		userDAO:       userDAO,
		pseudonymDAO:  pseudonymDAO,
		permissionDAO: permissionDAO,
	}
}

// ListUsersInput represents the input for listing users
type ListUsersInput struct {
	middleware.AuthInput
	Page  int    `query:"page" example:"1"`
	Limit int    `query:"limit" example:"25"`
	Query string `query:"q" example:"user@example.com"`
}

// ListUsersResponse represents the response for listing users
type ListUsersResponse struct {
	Status int                   `json:"-" example:"200"`
	Body   ListUsersResponseBody `json:"body"`
}

// ListUsersResponseBody represents the body of the list users response
type ListUsersResponseBody struct {
	Users      []AdminUserInfo      `json:"users"`
	Pagination apimodels.Pagination `json:"pagination"`
}

// AdminUserInfo represents user information for admin purposes
type AdminUserInfo struct {
	UserID         int64  `json:"user_id" example:"123"`
	Email          string `json:"email" example:"user@example.com"`
	CreatedAt      string `json:"created_at" example:"2024-01-01T12:00:00Z"`
	IsActive       bool   `json:"is_active" example:"true"`
	IsSuspended    bool   `json:"is_suspended" example:"false"`
	PseudonymCount int    `json:"pseudonym_count" example:"2"`
}

// NewListUsersResponse creates a new list users response
func NewListUsersResponse(users []AdminUserInfo, page, limit, total int) *ListUsersResponse {
	pages := (total + limit - 1) / limit // Ceiling division

	return &ListUsersResponse{
		Status: 200,
		Body: ListUsersResponseBody{
			Users: users,
			Pagination: apimodels.Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: pages,
			},
		},
	}
}

// ListUsers handles listing all users for admin purposes
func (h *AdminHandler) ListUsers(ctx context.Context, input *ListUsersInput) (*ListUsersResponse, error) {
	// Extract user from Huma input - authentication is required
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for admin user list")
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
			Msg("Platform admin capability required for admin user list")
		return nil, huma.Error403Forbidden("insufficient permissions: platform admin capability required")
	}

	// Build request log fields
	requestLog := log.Info().
		Str("endpoint", "admin/users").
		Str("component", "handler")

	// Add user_id to log if user context is available
	if user != nil {
		requestLog = requestLog.Int64("user_id", user.UserID)
	}

	// Add search query to log if provided
	if input.Query != "" {
		requestLog = requestLog.Str("search_query", input.Query)
	}

	requestLog.Msg("Admin user list requested")

	// Handle pagination
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	// Get total count of users for pagination
	total, err := h.userDAO.CountUsers(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count users from database")
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	// Get users with proper pagination
	users, err := h.userDAO.ListUsers(ctx, limit, (page-1)*limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get users from database")
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	// If search query is provided, filter the results in memory
	// This is a temporary solution - ideally we'd implement database-level search
	if input.Query != "" {
		var filteredUsers []*dbmodels.User
		query := input.Query
		for _, user := range users {
			// Search by user ID (convert to string for comparison)
			if fmt.Sprintf("%d", user.UserID) == query {
				filteredUsers = append(filteredUsers, user)
				continue
			}
			// Search by email (case-insensitive)
			if strings.Contains(strings.ToLower(user.Email), strings.ToLower(query)) {
				filteredUsers = append(filteredUsers, user)
			}
		}
		users = filteredUsers
		// Update total count for filtered results
		total = int64(len(users))
	}

	// Convert database users to API users
	apiUsers := make([]AdminUserInfo, 0, len(users))
	for _, user := range users {
		// Handle nullable fields
		createdAt := ""
		if user.CreatedAt.Valid {
			createdAt = user.CreatedAt.V.Format("2006-01-02T15:04:05Z")
		}

		isActive := true
		if user.IsActive.Valid {
			isActive = user.IsActive.V
		}

		isSuspended := false
		if user.IsSuspended.Valid {
			isSuspended = user.IsSuspended.V
		}

		// Count pseudonyms for this user (but don't fetch details for privacy)
		pseudonymCount := 0
		pseudonyms, err := h.pseudonymDAO.GetPseudonymsByRealIdentityDirect(ctx, user.Email)
		if err != nil {
			log.Warn().Err(err).Int64("user_id", user.UserID).Msg("Failed to get user pseudonyms")
		} else {
			pseudonymCount = len(pseudonyms)
		}

		apiUsers = append(apiUsers, AdminUserInfo{
			UserID:         user.UserID,
			Email:          user.Email,
			CreatedAt:      createdAt,
			IsActive:       isActive,
			IsSuspended:    isSuspended,
			PseudonymCount: pseudonymCount,
		})
	}

	response := NewListUsersResponse(apiUsers, page, limit, int(total))

	// Build completion log fields
	completionLog := log.Info().
		Str("endpoint", "admin/users").
		Str("component", "handler").
		Int("count", len(apiUsers)).
		Int("total", int(total))

	// Add search query to log if provided
	if input.Query != "" {
		completionLog = completionLog.Str("search_query", input.Query)
	}

	// Add user_id to log if user context is available
	if user != nil {
		completionLog = completionLog.Int64("user_id", user.UserID)
	}

	completionLog.Msg("Admin user list completed")

	return response, nil
}

// GetUserInput represents the input for getting a specific user
type GetUserInput struct {
	middleware.AuthInput
	UserID int64 `path:"user_id" example:"123"`
}

// GetUserResponse represents the response for getting a specific user
type GetUserResponse struct {
	Status int           `json:"-" example:"200"`
	Body   AdminUserInfo `json:"body"`
}

// GetUser handles getting a specific user for admin purposes
func (h *AdminHandler) GetUser(ctx context.Context, input *GetUserInput) (*GetUserResponse, error) {
	// Extract user from Huma input - authentication is required
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for admin user get")
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
			Msg("Platform admin capability required for admin user get")
		return nil, huma.Error403Forbidden("insufficient permissions: platform admin capability required")
	}

	// Build request log fields
	requestLog := log.Info().
		Str("endpoint", "admin/users/{user_id}").
		Str("component", "handler").
		Int64("target_user_id", input.UserID)

	// Add user_id to log if user context is available
	if user != nil {
		requestLog = requestLog.Int64("user_id", user.UserID)
	}

	requestLog.Msg("Admin get user requested")

	// Get user from database
	dbUser, err := h.userDAO.GetUserByID(ctx, input.UserID)
	if err != nil {
		log.Error().Err(err).Int64("target_user_id", input.UserID).Msg("Failed to get user from database")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if dbUser == nil {
		log.Warn().Int64("target_user_id", input.UserID).Msg("User not found")
		return nil, huma.Error404NotFound("user not found")
	}

	// Handle nullable fields
	createdAt := ""
	if dbUser.CreatedAt.Valid {
		createdAt = dbUser.CreatedAt.V.Format("2006-01-02T15:04:05Z")
	}

	isActive := true
	if dbUser.IsActive.Valid {
		isActive = dbUser.IsActive.V
	}

	isSuspended := false
	if dbUser.IsSuspended.Valid {
		isSuspended = dbUser.IsSuspended.V
	}

	// Count pseudonyms for this user (but don't fetch details for privacy)
	pseudonymCount := 0
	pseudonyms, err := h.pseudonymDAO.GetPseudonymsByRealIdentityDirect(ctx, dbUser.Email)
	if err != nil {
		log.Warn().Err(err).Int64("user_id", dbUser.UserID).Msg("Failed to get user pseudonyms")
	} else {
		pseudonymCount = len(pseudonyms)
	}

	apiUser := AdminUserInfo{
		UserID:         dbUser.UserID,
		Email:          dbUser.Email,
		CreatedAt:      createdAt,
		IsActive:       isActive,
		IsSuspended:    isSuspended,
		PseudonymCount: pseudonymCount,
	}

	response := &GetUserResponse{
		Status: 200,
		Body:   apiUser,
	}

	// Build completion log fields
	completionLog := log.Info().
		Str("endpoint", "admin/users/{user_id}").
		Str("component", "handler").
		Int64("target_user_id", input.UserID)

	// Add user_id to log if user context is available
	if user != nil {
		completionLog = completionLog.Int64("user_id", user.UserID)
	}

	completionLog.Msg("Admin get user completed")

	return response, nil
}

// ListPseudonymsInput represents the input for listing pseudonyms
type ListPseudonymsInput struct {
	middleware.AuthInput
	Page  int `query:"page" example:"1"`
	Limit int `query:"limit" example:"25"`
}

// ListPseudonymsResponse represents the response for listing pseudonyms
type ListPseudonymsResponse struct {
	Status int                        `json:"-" example:"200"`
	Body   ListPseudonymsResponseBody `json:"body"`
}

// ListPseudonymsResponseBody represents the body of the list pseudonyms response
type ListPseudonymsResponseBody struct {
	Pseudonyms []AdminPseudonymInfo `json:"pseudonyms"`
	Pagination apimodels.Pagination `json:"pagination"`
}

// AdminPseudonymInfo represents pseudonym information for admin purposes
type AdminPseudonymInfo struct {
	PseudonymID  string `json:"pseudonym_id" example:"354f5361a2af036b97f195e77bcaec8a"`
	DisplayName  string `json:"display_name" example:"JohnDoe"`
	Slug         string `json:"slug,omitempty" example:"john-doe"`
	KarmaScore   int    `json:"karma_score" example:"150"`
	IsActive     bool   `json:"is_active" example:"true"`
	CreatedAt    string `json:"created_at" example:"2024-01-01T12:00:00Z"`
	RealIdentity string `json:"real_identity" example:"user@example.com"`
}

// NewListPseudonymsResponse creates a new list pseudonyms response
func NewListPseudonymsResponse(pseudonyms []AdminPseudonymInfo, page, limit, total int) *ListPseudonymsResponse {
	pages := (total + limit - 1) / limit // Ceiling division

	return &ListPseudonymsResponse{
		Status: 200,
		Body: ListPseudonymsResponseBody{
			Pseudonyms: pseudonyms,
			Pagination: apimodels.Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: pages,
			},
		},
	}
}

// ListPseudonyms handles listing all pseudonyms for admin purposes
func (h *AdminHandler) ListPseudonyms(ctx context.Context, input *ListPseudonymsInput) (*ListPseudonymsResponse, error) {
	// Extract user from Huma input - authentication is required
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for admin pseudonym list")
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
			Msg("Platform admin capability required for admin pseudonym list")
		return nil, huma.Error403Forbidden("insufficient permissions: platform admin capability required")
	}

	// Build request log fields
	requestLog := log.Info().
		Str("endpoint", "admin/pseudonyms").
		Str("component", "handler")

	// Add user_id to log if user context is available
	if user != nil {
		requestLog = requestLog.Int64("user_id", user.UserID)
	}

	requestLog.Msg("Admin pseudonym list requested")

	// Handle pagination
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	// Get total count of pseudonyms for pagination
	total, err := h.pseudonymDAO.CountAllPseudonyms(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count pseudonyms from database")
		return nil, fmt.Errorf("failed to count pseudonyms: %w", err)
	}

	// Get pseudonyms with proper pagination
	// For now, we'll get all and paginate in memory, but this should be improved
	// with a dedicated paginated method in the DAO
	allPseudonyms, err := h.pseudonymDAO.GetAllPseudonyms(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get pseudonyms from database")
		return nil, fmt.Errorf("failed to get pseudonyms: %w", err)
	}

	// Handle pagination in memory
	start := (page - 1) * limit
	end := start + limit
	if end > len(allPseudonyms) {
		end = len(allPseudonyms)
	}
	if start >= len(allPseudonyms) {
		start = len(allPseudonyms)
	}

	var pseudonyms []*dbmodels.Pseudonym
	if start < len(allPseudonyms) {
		pseudonyms = allPseudonyms[start:end]
	}

	// Convert database pseudonyms to API pseudonyms
	apiPseudonyms := make([]AdminPseudonymInfo, 0, len(pseudonyms))
	for _, pseudonym := range pseudonyms {
		// Handle nullable fields
		createdAt := ""
		if pseudonym.CreatedAt.Valid {
			createdAt = pseudonym.CreatedAt.V.Format("2006-01-02T15:04:05Z")
		}

		isActive := true
		if pseudonym.IsActive.Valid {
			isActive = pseudonym.IsActive.V
		}

		slug := ""
		if pseudonym.Slug.Valid {
			slug = pseudonym.Slug.V
		}

		// Get karma score, defaulting to 0 if null
		karmaScore := 0
		if pseudonym.KarmaScore.Valid {
			karmaScore = int(pseudonym.KarmaScore.V)
		}

		apiPseudonyms = append(apiPseudonyms, AdminPseudonymInfo{
			PseudonymID:  pseudonym.PseudonymID, // Use the actual pseudonym ID
			DisplayName:  pseudonym.DisplayName,
			Slug:         slug,
			KarmaScore:   karmaScore,
			IsActive:     isActive,
			CreatedAt:    createdAt,
			RealIdentity: pseudonym.PseudonymID, // Use pseudonym ID as real identity for now
		})
	}

	response := NewListPseudonymsResponse(apiPseudonyms, page, limit, int(total))

	// Build completion log fields
	completionLog := log.Info().
		Str("endpoint", "admin/pseudonyms").
		Str("component", "handler").
		Int("count", len(apiPseudonyms)).
		Int("total", int(total))

	// Add user_id to log if user context is available
	if user != nil {
		completionLog = completionLog.Int64("user_id", user.UserID)
	}

	completionLog.Msg("Admin pseudonym list completed")

	return response, nil
}

// containsIgnoreCase performs case-insensitive string containment check
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
