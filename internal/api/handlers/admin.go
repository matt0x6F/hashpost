package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/config"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/services"
	"github.com/rs/zerolog/log"
)

// AdminHandler handles administrative operations
type AdminHandler struct {
	userDAO               dao.UserDAOInterface
	pseudonymDAO          dao.PseudonymDAOInterface
	permissionDAO         dao.PermissionDAOInterface
	passwordResetTokenDAO dao.PasswordResetTokenDAOInterface
	emailService          *services.EmailService
	config                *config.Config
	postDAO               dao.PostDAOInterface
	commentDAO            dao.CommentDAOInterface
	roleKeyDAO            dao.RoleKeyDAOInterface
	subforumDAO           dao.SubforumDAOInterface
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	userDAO dao.UserDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
	permissionDAO dao.PermissionDAOInterface,
	passwordResetTokenDAO dao.PasswordResetTokenDAOInterface,
	emailService *services.EmailService,
	config *config.Config,
	postDAO dao.PostDAOInterface,
	commentDAO dao.CommentDAOInterface,
	roleKeyDAO dao.RoleKeyDAOInterface,
	subforumDAO dao.SubforumDAOInterface,
) *AdminHandler {
	return &AdminHandler{
		userDAO:               userDAO,
		pseudonymDAO:          pseudonymDAO,
		permissionDAO:         permissionDAO,
		passwordResetTokenDAO: passwordResetTokenDAO,
		emailService:          emailService,
		config:                config,
		postDAO:               postDAO,
		commentDAO:            commentDAO,
		roleKeyDAO:            roleKeyDAO,
		subforumDAO:           subforumDAO,
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
	Users      []AdminUserInfo   `json:"users"`
	Pagination models.Pagination `json:"pagination"`
}

// AdminUserInfo represents user information for admin purposes
type AdminUserInfo struct {
	UserID         int64  `json:"user_id" example:"123"`
	Email          string `json:"email" example:"user@example.com"`
	CreatedAt      string `json:"created_at" example:"2024-01-01T12:00:00Z"`
	IsActive       bool   `json:"is_active" example:"true"`
	IsSuspended    bool   `json:"is_suspended" example:"false"`
	PseudonymCount int    `json:"pseudonym_count" example:"2"`
	EmailVerified  bool   `json:"email_verified" example:"true"`
}

// NewListUsersResponse creates a new list users response
func NewListUsersResponse(users []AdminUserInfo, page, limit, total int) *ListUsersResponse {
	pages := (total + limit - 1) / limit // Ceiling division

	return &ListUsersResponse{
		Status: 200,
		Body: ListUsersResponseBody{
			Users: users,
			Pagination: models.Pagination{
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

		// Handle email verification status
		emailVerified := false
		if user.EmailVerified.Valid {
			emailVerified = user.EmailVerified.V
		}

		apiUsers = append(apiUsers, AdminUserInfo{
			UserID:         user.UserID,
			Email:          user.Email,
			CreatedAt:      createdAt,
			IsActive:       isActive,
			IsSuspended:    isSuspended,
			PseudonymCount: pseudonymCount,
			EmailVerified:  emailVerified,
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

	// Handle email verification status
	emailVerified := false
	if dbUser.EmailVerified.Valid {
		emailVerified = dbUser.EmailVerified.V
	}

	apiUser := AdminUserInfo{
		UserID:         dbUser.UserID,
		Email:          dbUser.Email,
		CreatedAt:      createdAt,
		IsActive:       isActive,
		IsSuspended:    isSuspended,
		PseudonymCount: pseudonymCount,
		EmailVerified:  emailVerified,
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
	Pagination models.Pagination    `json:"pagination"`
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
			Pagination: models.Pagination{
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

// generateToken generates a random token for password reset
func (h *AdminHandler) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// UpdateUserInput represents the input for updating a user
type UpdateUserInput struct {
	middleware.AuthInput
	UserID int64 `path:"user_id" example:"123"`
	Body   struct {
		Email         *string `json:"email,omitempty" example:"newemail@example.com"`
		IsActive      *bool   `json:"is_active,omitempty" example:"true"`
		IsSuspended   *bool   `json:"is_suspended,omitempty" example:"false"`
		EmailVerified *bool   `json:"email_verified,omitempty" example:"true"`
	} `json:"body"`
}

// UpdateUserResponse represents the response for updating a user
type UpdateUserResponse struct {
	Status int           `json:"-" example:"200"`
	Body   AdminUserInfo `json:"body"`
}

// UpdateUser handles updating a user for admin purposes
func (h *AdminHandler) UpdateUser(ctx context.Context, input *UpdateUserInput) (*UpdateUserResponse, error) {
	// Extract user from Huma input - authentication is required
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for admin user update")
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
			Msg("Platform admin capability required for admin user update")
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

	requestLog.Msg("Admin update user requested")

	// Get existing user from database
	dbUser, err := h.userDAO.GetUserByID(ctx, input.UserID)
	if err != nil {
		log.Error().Err(err).Int64("target_user_id", input.UserID).Msg("Failed to get user from database")
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if dbUser == nil {
		log.Warn().Int64("target_user_id", input.UserID).Msg("User not found")
		return nil, huma.Error404NotFound("user not found")
	}

	// Build update setter
	updates := &dbmodels.UserSetter{}

	// Only update fields that are provided
	if input.Body.Email != nil {
		updates.Email = input.Body.Email
	}
	if input.Body.IsActive != nil {
		updates.IsActive = &sql.Null[bool]{V: *input.Body.IsActive, Valid: true}
	}
	if input.Body.IsSuspended != nil {
		updates.IsSuspended = &sql.Null[bool]{V: *input.Body.IsSuspended, Valid: true}
	}
	if input.Body.EmailVerified != nil {
		updates.EmailVerified = &sql.Null[bool]{V: *input.Body.EmailVerified, Valid: true}
	}

	// Set updated timestamp
	now := time.Now()
	updates.UpdatedAt = &sql.Null[time.Time]{V: now, Valid: true}

	// Apply updates
	if err := h.userDAO.UpdateUser(ctx, input.UserID, updates); err != nil {
		log.Error().Err(err).Int64("target_user_id", input.UserID).Msg("Failed to update user in database")
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Get updated user to return
	updatedUser, err := h.userDAO.GetUserByID(ctx, input.UserID)
	if err != nil {
		log.Error().Err(err).Int64("target_user_id", input.UserID).Msg("Failed to get updated user from database")
		return nil, fmt.Errorf("failed to get updated user: %w", err)
	}

	// Convert to API response format
	createdAt := ""
	if updatedUser.CreatedAt.Valid {
		createdAt = updatedUser.CreatedAt.V.Format("2006-01-02T15:04:05Z")
	}

	isActive := true
	if updatedUser.IsActive.Valid {
		isActive = updatedUser.IsActive.V
	}

	isSuspended := false
	if updatedUser.IsSuspended.Valid {
		isSuspended = updatedUser.IsSuspended.V
	}

	// Count pseudonyms for this user
	pseudonymCount := 0
	pseudonyms, err := h.pseudonymDAO.GetPseudonymsByRealIdentityDirect(ctx, updatedUser.Email)
	if err != nil {
		log.Warn().Err(err).Int64("user_id", updatedUser.UserID).Msg("Failed to get user pseudonyms")
	} else {
		pseudonymCount = len(pseudonyms)
	}

	// Handle email verification status
	emailVerified := false
	if updatedUser.EmailVerified.Valid {
		emailVerified = updatedUser.EmailVerified.V
	}

	apiUser := AdminUserInfo{
		UserID:         updatedUser.UserID,
		Email:          updatedUser.Email,
		CreatedAt:      createdAt,
		IsActive:       isActive,
		IsSuspended:    isSuspended,
		PseudonymCount: pseudonymCount,
		EmailVerified:  emailVerified,
	}

	response := &UpdateUserResponse{
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

	completionLog.Msg("Admin update user completed")

	return response, nil
}

// TriggerPasswordResetInput represents the input for triggering a password reset
type TriggerPasswordResetInput struct {
	middleware.AuthInput
	UserID int64 `path:"user_id" example:"123"`
}

// TriggerPasswordResetResponse represents the response for triggering a password reset
type TriggerPasswordResetResponse struct {
	Status int `json:"-" example:"200"`
	Body   struct {
		Message string `json:"message" example:"Password reset email sent"`
	} `json:"body"`
}

// TriggerPasswordReset handles triggering a password reset for a user
func (h *AdminHandler) TriggerPasswordReset(ctx context.Context, input *TriggerPasswordResetInput) (*TriggerPasswordResetResponse, error) {
	// Extract user from Huma input - authentication is required
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for admin password reset trigger")
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
			Msg("Platform admin capability required for admin password reset trigger")
		return nil, huma.Error403Forbidden("insufficient permissions: platform admin capability required")
	}

	// Build request log fields
	requestLog := log.Info().
		Str("endpoint", "admin/users/{user_id}/trigger-password-reset").
		Str("component", "handler").
		Int64("target_user_id", input.UserID)

	// Add user_id to log if user context is available
	if user != nil {
		requestLog = requestLog.Int64("user_id", user.UserID)
	}

	requestLog.Msg("Admin trigger password reset requested")

	// Get target user from database
	targetUser, err := h.userDAO.GetUserByID(ctx, input.UserID)
	if err != nil {
		log.Error().Err(err).Int64("target_user_id", input.UserID).Msg("Failed to get target user from database")
		return nil, fmt.Errorf("failed to get target user: %w", err)
	}

	if targetUser == nil {
		log.Warn().Int64("target_user_id", input.UserID).Msg("Target user not found")
		return nil, huma.Error404NotFound("user not found")
	}

	// Generate reset token
	resetToken, err := h.generateToken()
	if err != nil {
		log.Error().
			Err(err).
			Int64("target_user_id", input.UserID).
			Msg("Failed to generate reset token")
		return nil, fmt.Errorf("failed to generate reset token: %w", err)
	}

	// Store reset token in database with expiration
	expiresAt := time.Now().Add(1 * time.Hour) // Token expires in 1 hour
	err = h.passwordResetTokenDAO.CreateToken(ctx, targetUser.UserID, resetToken, expiresAt)
	if err != nil {
		log.Error().
			Err(err).
			Int64("target_user_id", input.UserID).
			Msg("Failed to store reset token")
		return nil, fmt.Errorf("failed to store reset token: %w", err)
	}

	// Send password reset email
	if h.emailService != nil {
		resetURL := fmt.Sprintf("%s/reset-password/confirm?token=%s", h.config.Server.SiteURL, resetToken)
		err = h.emailService.SendEmail(ctx, "password_reset", targetUser.Email, targetUser.Email, map[string]interface{}{
			"reset_url": resetURL,
		})
		if err != nil {
			log.Error().
				Err(err).
				Int64("target_user_id", input.UserID).
				Msg("Failed to send password reset email")
			return nil, fmt.Errorf("failed to send password reset email: %w", err)
		}
	}

	log.Info().
		Int64("target_user_id", input.UserID).
		Msg("Admin password reset email sent")

	response := &TriggerPasswordResetResponse{
		Status: 200,
		Body: struct {
			Message string `json:"message" example:"Password reset email sent"`
		}{
			Message: "Password reset email sent",
		},
	}

	// Build completion log fields
	completionLog := log.Info().
		Str("endpoint", "admin/users/{user_id}/trigger-password-reset").
		Str("component", "handler").
		Int64("target_user_id", input.UserID)

	// Add user_id to log if user context is available
	if user != nil {
		completionLog = completionLog.Int64("user_id", user.UserID)
	}

	completionLog.Msg("Admin trigger password reset completed")

	return response, nil
}

// ListUserPseudonymsInput represents the input for listing pseudonyms for a specific user
type ListUserPseudonymsInput struct {
	middleware.AuthInput
	UserID int64 `path:"user_id" example:"123"`
}

// ListUserPseudonymsResponse represents the response for listing user pseudonyms
type ListUserPseudonymsResponse struct {
	Status int                            `json:"-" example:"200"`
	Body   ListUserPseudonymsResponseBody `json:"body"`
}

// ListUserPseudonymsResponseBody represents the body of the list user pseudonyms response
type ListUserPseudonymsResponseBody struct {
	Pseudonyms []AdminPseudonymInfo `json:"pseudonyms"`
}

// ListUserPseudonyms handles listing pseudonyms for a specific user
func (h *AdminHandler) ListUserPseudonyms(ctx context.Context, input *ListUserPseudonymsInput) (*ListUserPseudonymsResponse, error) {
	// Extract user from Huma input - authentication is required
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for admin user pseudonym list")
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
			Msg("Platform admin capability required for admin user pseudonym list")
		return nil, huma.Error403Forbidden("insufficient permissions: platform admin capability required")
	}

	userID := input.UserID
	log.Info().
		Str("endpoint", "admin/users/{user_id}/pseudonyms").
		Str("component", "handler").
		Int64("target_user_id", userID).
		Int64("user_id", user.UserID).
		Msg("Admin user pseudonym list requested")

	// Get pseudonyms for the specific user
	pseudonyms, err := h.pseudonymDAO.GetPseudonymsByUserIDDirect(ctx, userID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("Failed to get user pseudonyms from database")
		return nil, fmt.Errorf("failed to get user pseudonyms: %w", err)
	}

	// Convert to API format
	apiPseudonyms := make([]AdminPseudonymInfo, 0, len(pseudonyms))
	for _, pseudonym := range pseudonyms {
		// Get karma score
		karmaScore, _ := h.pseudonymDAO.CalculateKarmaForPseudonym(ctx, pseudonym.PseudonymID)

		// Handle nullable fields
		createdAt := ""
		if pseudonym.CreatedAt.Valid {
			createdAt = pseudonym.CreatedAt.V.Format("2006-01-02T15:04:05Z")
		}

		slug := ""
		if pseudonym.Slug.Valid {
			slug = pseudonym.Slug.V
		}

		isActive := true
		if pseudonym.IsActive.Valid {
			isActive = pseudonym.IsActive.V
		}

		// Get user info for real identity
		userInfo, err := h.userDAO.GetUserByID(ctx, userID)
		realIdentity := ""
		if err == nil && userInfo != nil {
			realIdentity = userInfo.Email
		}

		apiPseudonym := AdminPseudonymInfo{
			PseudonymID:  pseudonym.PseudonymID,
			DisplayName:  pseudonym.DisplayName,
			Slug:         slug,
			KarmaScore:   int(karmaScore),
			IsActive:     isActive,
			CreatedAt:    createdAt,
			RealIdentity: realIdentity,
		}

		apiPseudonyms = append(apiPseudonyms, apiPseudonym)
	}

	response := &ListUserPseudonymsResponse{
		Status: 200,
		Body: ListUserPseudonymsResponseBody{
			Pseudonyms: apiPseudonyms,
		},
	}

	log.Info().
		Str("endpoint", "admin/users/{user_id}/pseudonyms").
		Str("component", "handler").
		Int64("target_user_id", userID).
		Int64("user_id", user.UserID).
		Int("pseudonym_count", len(apiPseudonyms)).
		Msg("Admin user pseudonym list completed")

	return response, nil
}

// GetPseudonymInput represents the input for getting a specific pseudonym
type GetPseudonymInput struct {
	middleware.AuthInput
	PseudonymID string `path:"pseudonym_id" example:"354f5361a2af036b97f195e77bcaec8a"`
}

// GetPseudonymResponse represents the response for getting a specific pseudonym
type GetPseudonymResponse struct {
	Status int                  `json:"-" example:"200"`
	Body   AdminPseudonymDetail `json:"body"`
}

// AdminPseudonymDetail represents detailed pseudonym information for admin purposes
type AdminPseudonymDetail struct {
	PseudonymID         string               `json:"pseudonym_id" example:"354f5361a2af036b97f195e77bcaec8a"`
	DisplayName         string               `json:"display_name" example:"JohnDoe"`
	Slug                string               `json:"slug,omitempty" example:"john-doe"`
	KarmaScore          int                  `json:"karma_score" example:"150"`
	IsActive            bool                 `json:"is_active" example:"true"`
	CreatedAt           string               `json:"created_at" example:"2024-01-01T12:00:00Z"`
	RealIdentity        string               `json:"real_identity" example:"user@example.com"`
	Bio                 string               `json:"bio,omitempty" example:"A brief description of the user."`
	WebsiteURL          string               `json:"website_url,omitempty" example:"https://example.com"`
	ShowKarma           bool                 `json:"show_karma" example:"true"`
	AllowDirectMessages bool                 `json:"allow_direct_messages" example:"true"`
	PostCount           int                  `json:"post_count" example:"10"`
	CommentCount        int                  `json:"comment_count" example:"5"`
	Roles               []string             `json:"roles,omitempty"`
	Capabilities        []string             `json:"capabilities,omitempty"`
	RoleKeys            []AdminRoleKeyDetail `json:"role_keys,omitempty"`
}

// AdminRoleKeyDetail represents detailed role key information for admin purposes
type AdminRoleKeyDetail struct {
	KeyID        string   `json:"key_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	RoleName     string   `json:"role_name" example:"moderator"`
	Scope        string   `json:"scope" example:"moderation"`
	Capabilities []string `json:"capabilities" example:"moderate_content,ban_users"`
	SubforumID   *int32   `json:"subforum_id,omitempty" example:"123"`
	SubforumName string   `json:"subforum_name,omitempty" example:"General Discussion"`
	ExpiresAt    string   `json:"expires_at" example:"2025-12-31T23:59:59Z"`
	IsActive     bool     `json:"is_active" example:"true"`
}

// RoleKeyUpdate represents a role key update in the admin pseudonym update
type RoleKeyUpdate struct {
	RoleName     string   `json:"role_name" example:"moderator"`
	Scope        string   `json:"scope" example:"moderation"`
	Capabilities []string `json:"capabilities" example:"moderate_content,ban_users"`
	SubforumID   *int32   `json:"subforum_id,omitempty" example:"123"`
	ExpiresAt    string   `json:"expires_at" example:"2025-12-31T23:59:59Z"`
	IsActive     bool     `json:"is_active" example:"true"`
}

// UpdatePseudonymInput represents the input for updating a pseudonym
type UpdatePseudonymInput struct {
	middleware.AuthInput
	PseudonymID string `path:"pseudonym_id" example:"354f5361a2af036b97f195e77bcaec8a"`
	Body        struct {
		DisplayName         *string          `json:"display_name,omitempty" example:"JohnDoe"`
		Slug                *string          `json:"slug,omitempty" example:"john-doe"`
		IsActive            *bool            `json:"is_active,omitempty" example:"true"`
		Bio                 *string          `json:"bio,omitempty" example:"A brief description of the user."`
		WebsiteURL          *string          `json:"website_url,omitempty" example:"https://example.com"`
		ShowKarma           *bool            `json:"show_karma,omitempty" example:"true"`
		AllowDirectMessages *bool            `json:"allow_direct_messages,omitempty" example:"true"`
		RoleKeys            *[]RoleKeyUpdate `json:"role_keys,omitempty"`
	} `json:"body"`
}

// UpdatePseudonymResponse represents the response for updating a pseudonym
type UpdatePseudonymResponse struct {
	Status int                  `json:"-" example:"200"`
	Body   AdminPseudonymDetail `json:"body"`
}

// GetPseudonym handles getting a specific pseudonym for admin purposes
func (h *AdminHandler) GetPseudonym(ctx context.Context, input *GetPseudonymInput) (*GetPseudonymResponse, error) {
	// Extract user from Huma input - authentication is required
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for admin pseudonym get")
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
			Msg("Platform admin capability required for admin pseudonym get")
		return nil, huma.Error403Forbidden("insufficient permissions: platform admin capability required")
	}

	pseudonymID := input.PseudonymID
	log.Info().
		Str("endpoint", "admin/pseudonyms/{pseudonym_id}").
		Str("component", "handler").
		Str("pseudonym_id", pseudonymID).
		Int64("user_id", user.UserID).
		Msg("Admin get pseudonym requested")

	// Get pseudonym from database
	pseudonym, err := h.pseudonymDAO.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get pseudonym from database")
		return nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}
	if pseudonym == nil {
		log.Warn().Str("pseudonym_id", pseudonymID).Msg("Pseudonym not found")
		return nil, huma.Error404NotFound("pseudonym not found")
	}

	// Get user ID from pseudonym using the DAO method
	userID, err := h.pseudonymDAO.GetUserIDByPseudonym(ctx, pseudonymID, constants.RolePlatformAdmin, constants.ScopeCorrelation)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get user ID from pseudonym")
		return nil, fmt.Errorf("failed to get user ID from pseudonym: %w", err)
	}

	// Get user info for real identity
	userInfo, err := h.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("Failed to get user info")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Get post and comment counts
	postCount, _ := h.postDAO.CountPostsByPseudonym(ctx, pseudonymID)
	commentCount, _ := h.commentDAO.CountCommentsByPseudonym(ctx, pseudonymID)

	// Get roles and capabilities from role keys
	roles := []string{}
	capabilities := []string{}

	roleKeys, err := h.roleKeyDAO.ListRoleKeysByPseudonym(ctx, pseudonymID)
	if err != nil {
		log.Warn().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get role keys")
	} else {
		for _, roleKey := range roleKeys {
			// RoleName is a string, not nullable
			roles = append(roles, roleKey.RoleName)

			// Capabilities is types.JSON[json.RawMessage], need to get the value
			capabilitiesBytes, err := roleKey.Capabilities.Value()
			if err == nil {
				var caps []string
				if err := json.Unmarshal(capabilitiesBytes.([]byte), &caps); err == nil {
					capabilities = append(capabilities, caps...)
				}
			}
		}
	}

	// Remove duplicates
	roles = removeDuplicates(roles)
	capabilities = removeDuplicates(capabilities)

	// Convert role keys to API format
	apiRoleKeys := []AdminRoleKeyDetail{}
	if roleKeys != nil {
		for _, roleKey := range roleKeys {
			// Get capabilities from JSON
			var roleKeyCapabilities []string
			capabilitiesBytes, err := roleKey.Capabilities.Value()
			if err == nil {
				if err := json.Unmarshal(capabilitiesBytes.([]byte), &roleKeyCapabilities); err != nil {
					log.Warn().Err(err).Str("key_id", roleKey.KeyID.String()).Msg("Failed to unmarshal capabilities")
					roleKeyCapabilities = []string{}
				}
			}

			// Get subforum name if applicable
			subforumName := ""
			if roleKey.SubforumID.Valid {
				subforum, err := h.subforumDAO.GetSubforumByID(ctx, roleKey.SubforumID.V)
				if err == nil && subforum != nil {
					subforumName = subforum.Name
				}
			}

			// Format expiration date
			expiresAt := ""
			if !roleKey.ExpiresAt.IsZero() {
				expiresAt = roleKey.ExpiresAt.Format("2006-01-02T15:04:05Z")
			}

			apiRoleKey := AdminRoleKeyDetail{
				KeyID:        roleKey.KeyID.String(),
				RoleName:     roleKey.RoleName,
				Scope:        roleKey.Scope,
				Capabilities: roleKeyCapabilities,
				SubforumID: func() *int32 {
					if roleKey.SubforumID.Valid {
						return &roleKey.SubforumID.V
					}
					return nil
				}(),
				SubforumName: subforumName,
				ExpiresAt:    expiresAt,
				IsActive:     roleKey.IsActive.Valid && roleKey.IsActive.V,
			}
			apiRoleKeys = append(apiRoleKeys, apiRoleKey)
		}
	}

	// Convert to API response format
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

	karmaScore := 0
	if pseudonym.KarmaScore.Valid {
		karmaScore = int(pseudonym.KarmaScore.V)
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

	apiPseudonym := AdminPseudonymDetail{
		PseudonymID:         pseudonym.PseudonymID,
		DisplayName:         pseudonym.DisplayName,
		Slug:                slug,
		KarmaScore:          karmaScore,
		IsActive:            isActive,
		CreatedAt:           createdAt,
		RealIdentity:        userInfo.Email,
		Bio:                 bio,
		WebsiteURL:          websiteURL,
		ShowKarma:           showKarma,
		AllowDirectMessages: allowDirectMessages,
		PostCount:           int(postCount),
		CommentCount:        int(commentCount),
		Roles:               roles,
		Capabilities:        capabilities,
		RoleKeys:            apiRoleKeys,
	}

	response := &GetPseudonymResponse{
		Status: 200,
		Body:   apiPseudonym,
	}

	log.Info().
		Str("endpoint", "admin/pseudonyms/{pseudonym_id}").
		Str("component", "handler").
		Str("pseudonym_id", pseudonymID).
		Msg("Admin get pseudonym completed")

	return response, nil
}

// UpdatePseudonym handles updating a pseudonym for admin purposes
func (h *AdminHandler) UpdatePseudonym(ctx context.Context, input *UpdatePseudonymInput) (*UpdatePseudonymResponse, error) {
	// Extract user from Huma input - authentication is required
	user, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Warn().Err(err).Msg("Authentication required for admin pseudonym update")
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
			Msg("Platform admin capability required for admin pseudonym update")
		return nil, huma.Error403Forbidden("insufficient permissions: platform admin capability required")
	}

	pseudonymID := input.PseudonymID
	log.Info().
		Str("endpoint", "admin/pseudonyms/{pseudonym_id}").
		Str("component", "handler").
		Str("pseudonym_id", pseudonymID).
		Int64("user_id", user.UserID).
		Msg("Admin update pseudonym requested")

	// Get existing pseudonym from database
	existingPseudonym, err := h.pseudonymDAO.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get pseudonym from database")
		return nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}
	if existingPseudonym == nil {
		log.Warn().Str("pseudonym_id", pseudonymID).Msg("Pseudonym not found")
		return nil, huma.Error404NotFound("pseudonym not found")
	}

	// Build update setter
	updates := &dbmodels.PseudonymSetter{}

	// Only update fields that are provided
	if input.Body.DisplayName != nil {
		// Check if display name is already taken by another pseudonym
		if *input.Body.DisplayName != existingPseudonym.DisplayName {
			existing, _ := h.pseudonymDAO.GetPseudonymByDisplayName(ctx, *input.Body.DisplayName)
			if existing != nil && existing.PseudonymID != pseudonymID {
				return nil, huma.Error400BadRequest("display name is already taken")
			}
		}
		updates.DisplayName = input.Body.DisplayName
	}

	if input.Body.Slug != nil {
		// Check if slug is already taken by another pseudonym
		if *input.Body.Slug != "" {
			existing, _ := h.pseudonymDAO.GetPseudonymBySlug(ctx, *input.Body.Slug)
			if existing != nil && existing.PseudonymID != pseudonymID {
				return nil, huma.Error400BadRequest("slug is already taken")
			}
			slug := sql.Null[string]{V: *input.Body.Slug, Valid: true}
			updates.Slug = &slug
		} else {
			slug := sql.Null[string]{Valid: false}
			updates.Slug = &slug
		}
	}

	if input.Body.IsActive != nil {
		isActive := sql.Null[bool]{V: *input.Body.IsActive, Valid: true}
		updates.IsActive = &isActive
	}

	if input.Body.Bio != nil {
		if *input.Body.Bio != "" {
			bio := sql.Null[string]{V: *input.Body.Bio, Valid: true}
			updates.Bio = &bio
		} else {
			bio := sql.Null[string]{Valid: false}
			updates.Bio = &bio
		}
	}

	if input.Body.WebsiteURL != nil {
		if *input.Body.WebsiteURL != "" {
			websiteURL := sql.Null[string]{V: *input.Body.WebsiteURL, Valid: true}
			updates.WebsiteURL = &websiteURL
		} else {
			websiteURL := sql.Null[string]{Valid: false}
			updates.WebsiteURL = &websiteURL
		}
	}

	if input.Body.ShowKarma != nil {
		showKarma := sql.Null[bool]{V: *input.Body.ShowKarma, Valid: true}
		updates.ShowKarma = &showKarma
	}

	if input.Body.AllowDirectMessages != nil {
		allowDirectMessages := sql.Null[bool]{V: *input.Body.AllowDirectMessages, Valid: true}
		updates.AllowDirectMessages = &allowDirectMessages
	}

	// Apply pseudonym updates if any
	if len(updates.SetColumns()) > 0 {
		err = h.pseudonymDAO.UpdatePseudonym(ctx, pseudonymID, updates)
		if err != nil {
			log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym in database")
			return nil, fmt.Errorf("failed to update pseudonym: %w", err)
		}
	}

	// Handle role and capability updates if provided
	if input.Body.RoleKeys != nil {
		// Get current role keys for this pseudonym
		currentRoleKeys, err := h.roleKeyDAO.ListRoleKeysByPseudonym(ctx, pseudonymID)
		if err != nil {
			log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get current role keys")
			return nil, fmt.Errorf("failed to get current role keys: %w", err)
		}

		// Create a map of current roles for easy lookup
		currentRoles := make(map[string]bool)
		for _, roleKey := range currentRoleKeys {
			currentRoles[roleKey.RoleName] = true
		}

		// Handle role key updates
		for _, newRoleKey := range *input.Body.RoleKeys {
			// Determine if this is an update or a new role key
			var existingRoleKey *dbmodels.RoleKey
			for _, currentRoleKey := range currentRoleKeys {
				// Compare subforum IDs properly handling sql.Null
				subforumMatch := false
				if newRoleKey.SubforumID == nil {
					subforumMatch = !currentRoleKey.SubforumID.Valid
				} else if currentRoleKey.SubforumID.Valid {
					subforumMatch = currentRoleKey.SubforumID.V == *newRoleKey.SubforumID
				}

				if currentRoleKey.RoleName == newRoleKey.RoleName &&
					currentRoleKey.Scope == newRoleKey.Scope &&
					subforumMatch {
					existingRoleKey = currentRoleKey
					break
				}
			}

			if existingRoleKey == nil {
				// New role key
				expiresAt, err := time.Parse(time.RFC3339, newRoleKey.ExpiresAt)
				if err != nil {
					log.Error().Err(err).Str("pseudonym_id", pseudonymID).Str("role_name", newRoleKey.RoleName).Msg("Failed to parse expires_at")
					return nil, fmt.Errorf("failed to parse expires_at for new role key: %w", err)
				}

				_, err = h.roleKeyDAO.CreateRoleKey(
					ctx,
					newRoleKey.RoleName,
					newRoleKey.Scope,
					[]byte{}, // Empty key data for now
					newRoleKey.Capabilities,
					expiresAt,
					pseudonymID,
					pseudonymID,
					newRoleKey.SubforumID,
				)
				if err != nil {
					log.Error().Err(err).Str("pseudonym_id", pseudonymID).Str("role_name", newRoleKey.RoleName).Msg("Failed to create new role key")
					return nil, fmt.Errorf("failed to create new role key for role %s: %w", newRoleKey.RoleName, err)
				}
				log.Info().Str("pseudonym_id", pseudonymID).Str("role_name", newRoleKey.RoleName).Msg("Added new role key")
			} else {
				// Update existing role key
				expiresAt, err := time.Parse(time.RFC3339, newRoleKey.ExpiresAt)
				if err != nil {
					log.Error().Err(err).Str("pseudonym_id", pseudonymID).Str("role_name", newRoleKey.RoleName).Msg("Failed to parse expires_at")
					return nil, fmt.Errorf("failed to parse expires_at for updated role key: %w", err)
				}

				// Handle active state changes
				currentIsActive := existingRoleKey.IsActive.Valid && existingRoleKey.IsActive.V
				if newRoleKey.IsActive != currentIsActive {
					if newRoleKey.IsActive {
						// For now, we'll create a new role key to reactivate since we don't have direct DB access
						// This is a limitation of the current DAO design
						log.Info().Str("pseudonym_id", pseudonymID).Str("role_name", newRoleKey.RoleName).Msg("Role key reactivation not supported in current implementation")
					} else {
						if err := h.roleKeyDAO.DeactivateRoleKey(ctx, existingRoleKey.KeyID.String()); err != nil {
							log.Error().Err(err).Str("pseudonym_id", pseudonymID).Str("role_name", newRoleKey.RoleName).Msg("Failed to deactivate role key")
							return nil, fmt.Errorf("failed to deactivate role key for role %s: %w", newRoleKey.RoleName, err)
						}
						log.Info().Str("pseudonym_id", pseudonymID).Str("role_name", newRoleKey.RoleName).Msg("Deactivated role key")
					}
				}

				// Update capabilities if provided
				if len(newRoleKey.Capabilities) > 0 {
					// Deactivate the old role key
					if err := h.roleKeyDAO.DeactivateRoleKey(ctx, existingRoleKey.KeyID.String()); err != nil {
						log.Error().Err(err).Str("pseudonym_id", pseudonymID).Str("role_name", newRoleKey.RoleName).Msg("Failed to deactivate old role key")
						return nil, fmt.Errorf("failed to deactivate old role key for role %s: %w", newRoleKey.RoleName, err)
					}

					// Create new role key with updated capabilities
					_, err := h.roleKeyDAO.CreateRoleKey(
						ctx,
						existingRoleKey.RoleName,
						existingRoleKey.Scope,
						[]byte{}, // Empty key data for now
						newRoleKey.Capabilities,
						expiresAt,
						pseudonymID,
						pseudonymID,
						func() *int32 {
							if existingRoleKey.SubforumID.Valid {
								return &existingRoleKey.SubforumID.V
							}
							return nil
						}(),
					)
					if err != nil {
						log.Error().Err(err).Str("pseudonym_id", pseudonymID).Str("role_name", newRoleKey.RoleName).Msg("Failed to create new role key with updated capabilities")
						return nil, fmt.Errorf("failed to create new role key with updated capabilities for role %s: %w", newRoleKey.RoleName, err)
					}
					log.Info().Str("pseudonym_id", pseudonymID).Str("role_name", newRoleKey.RoleName).Interface("capabilities", newRoleKey.Capabilities).Msg("Updated capabilities for role")
				}
			}
		}

		log.Info().
			Str("pseudonym_id", pseudonymID).
			Interface("role_keys", input.Body.RoleKeys).
			Msg("Role and capability updates completed successfully")
	}

	// Get updated pseudonym to return
	updatedPseudonym, err := h.pseudonymDAO.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get updated pseudonym from database")
		return nil, fmt.Errorf("failed to get updated pseudonym: %w", err)
	}

	// Get user ID from pseudonym using the DAO method
	userID, err := h.pseudonymDAO.GetUserIDByPseudonym(ctx, pseudonymID, constants.RolePlatformAdmin, constants.ScopeCorrelation)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get user ID from pseudonym")
		return nil, fmt.Errorf("failed to get user ID from pseudonym: %w", err)
	}

	// Get user info for real identity
	userInfo, err := h.userDAO.GetUserByID(ctx, userID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userID).Msg("Failed to get user info")
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	// Get post and comment counts
	postCount, _ := h.postDAO.CountPostsByPseudonym(ctx, pseudonymID)
	commentCount, _ := h.commentDAO.CountCommentsByPseudonym(ctx, pseudonymID)

	// Get roles and capabilities from role keys
	roles := []string{}
	capabilities := []string{}

	roleKeys, err := h.roleKeyDAO.ListRoleKeysByPseudonym(ctx, pseudonymID)
	if err != nil {
		log.Warn().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get role keys")
	} else {
		for _, roleKey := range roleKeys {
			// RoleName is a string, not nullable
			roles = append(roles, roleKey.RoleName)

			// Capabilities is types.JSON[json.RawMessage], need to get the value
			capabilitiesBytes, err := roleKey.Capabilities.Value()
			if err == nil {
				var caps []string
				if err := json.Unmarshal(capabilitiesBytes.([]byte), &caps); err == nil {
					capabilities = append(capabilities, caps...)
				}
			}
		}
	}

	// Remove duplicates
	roles = removeDuplicates(roles)
	capabilities = removeDuplicates(capabilities)

	// Convert role keys to API format
	apiRoleKeys := []AdminRoleKeyDetail{}
	if roleKeys != nil {
		for _, roleKey := range roleKeys {
			// Get capabilities from JSON
			var roleKeyCapabilities []string
			capabilitiesBytes, err := roleKey.Capabilities.Value()
			if err == nil {
				if err := json.Unmarshal(capabilitiesBytes.([]byte), &roleKeyCapabilities); err != nil {
					log.Warn().Err(err).Str("key_id", roleKey.KeyID.String()).Msg("Failed to unmarshal capabilities")
					roleKeyCapabilities = []string{}
				}
			}

			// Get subforum name if applicable
			subforumName := ""
			if roleKey.SubforumID.Valid {
				subforum, err := h.subforumDAO.GetSubforumByID(ctx, roleKey.SubforumID.V)
				if err == nil && subforum != nil {
					subforumName = subforum.Name
				}
			}

			// Format expiration date
			expiresAt := ""
			if !roleKey.ExpiresAt.IsZero() {
				expiresAt = roleKey.ExpiresAt.Format("2006-01-02T15:04:05Z")
			}

			apiRoleKey := AdminRoleKeyDetail{
				KeyID:        roleKey.KeyID.String(),
				RoleName:     roleKey.RoleName,
				Scope:        roleKey.Scope,
				Capabilities: roleKeyCapabilities,
				SubforumID: func() *int32 {
					if roleKey.SubforumID.Valid {
						return &roleKey.SubforumID.V
					}
					return nil
				}(),
				SubforumName: subforumName,
				ExpiresAt:    expiresAt,
				IsActive:     roleKey.IsActive.Valid && roleKey.IsActive.V,
			}
			apiRoleKeys = append(apiRoleKeys, apiRoleKey)
		}
	}

	// Convert to API response format
	createdAt := ""
	if updatedPseudonym.CreatedAt.Valid {
		createdAt = updatedPseudonym.CreatedAt.V.Format("2006-01-02T15:04:05Z")
	}

	isActive := true
	if updatedPseudonym.IsActive.Valid {
		isActive = updatedPseudonym.IsActive.V
	}

	slug := ""
	if updatedPseudonym.Slug.Valid {
		slug = updatedPseudonym.Slug.V
	}

	karmaScore := 0
	if updatedPseudonym.KarmaScore.Valid {
		karmaScore = int(updatedPseudonym.KarmaScore.V)
	}

	bio := ""
	if updatedPseudonym.Bio.Valid {
		bio = updatedPseudonym.Bio.V
	}

	websiteURL := ""
	if updatedPseudonym.WebsiteURL.Valid {
		websiteURL = updatedPseudonym.WebsiteURL.V
	}

	showKarma := true
	if updatedPseudonym.ShowKarma.Valid {
		showKarma = updatedPseudonym.ShowKarma.V
	}

	allowDirectMessages := true
	if updatedPseudonym.AllowDirectMessages.Valid {
		allowDirectMessages = updatedPseudonym.AllowDirectMessages.V
	}

	apiPseudonym := AdminPseudonymDetail{
		PseudonymID:         updatedPseudonym.PseudonymID,
		DisplayName:         updatedPseudonym.DisplayName,
		Slug:                slug,
		KarmaScore:          karmaScore,
		IsActive:            isActive,
		CreatedAt:           createdAt,
		RealIdentity:        userInfo.Email,
		Bio:                 bio,
		WebsiteURL:          websiteURL,
		ShowKarma:           showKarma,
		AllowDirectMessages: allowDirectMessages,
		PostCount:           int(postCount),
		CommentCount:        int(commentCount),
		Roles:               roles,
		Capabilities:        capabilities,
		RoleKeys:            apiRoleKeys,
	}

	response := &UpdatePseudonymResponse{
		Status: 200,
		Body:   apiPseudonym,
	}

	log.Info().
		Str("endpoint", "admin/pseudonyms/{pseudonym_id}").
		Str("component", "handler").
		Str("pseudonym_id", pseudonymID).
		Msg("Admin update pseudonym completed")

	return response, nil
}

// Helper function to remove duplicates from string slices
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
