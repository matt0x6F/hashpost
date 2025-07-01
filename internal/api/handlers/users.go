package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
)

// UserHandler handles user management requests
type UserHandler struct {
	userDAO            *dao.UserDAO
	securePseudonymDAO *dao.SecurePseudonymDAO
	userPreferencesDAO *dao.UserPreferencesDAO
	userBlocksDAO      *dao.UserBlocksDAO
	postDAO            *dao.PostDAO
	commentDAO         *dao.CommentDAO
	ibeSystem          *ibe.IBESystem
}

// NewUserHandler creates a new user handler
func NewUserHandler(userDAO *dao.UserDAO, securePseudonymDAO *dao.SecurePseudonymDAO, userPreferencesDAO *dao.UserPreferencesDAO, userBlocksDAO *dao.UserBlocksDAO, postDAO *dao.PostDAO, commentDAO *dao.CommentDAO, ibeSystem *ibe.IBESystem) *UserHandler {
	return &UserHandler{
		userDAO:            userDAO,
		securePseudonymDAO: securePseudonymDAO,
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

	pseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, pseudonymID)
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
	response := apimodels.NewPseudonymProfileResponse(pseudonymID, displayName, bio, websiteURL, karmaScore, int(postCount), int(commentCount), showKarma, allowDirectMessages, createdAt, lastActiveAt)
	log.Info().Str("endpoint", "pseudonyms/profile").Str("component", "handler").Str("pseudonym_id", pseudonymID).Msg("Get pseudonym profile completed")
	return response, nil
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
	pseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, pseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to get pseudonym from database")
		return nil, fmt.Errorf("failed to get pseudonym: %w", err)
	}
	if pseudonym == nil {
		log.Warn().Str("pseudonym_id", pseudonymID).Msg("Pseudonym not found")
		return nil, fmt.Errorf("pseudonym not found")
	}

	// Use role-based access control for ownership verification
	ownsPseudonym, err := h.securePseudonymDAO.VerifyPseudonymOwnership(ctx, pseudonymID, int64(userID), "user", "self_correlation")
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Int("user_id", userID).Msg("Failed to verify pseudonym ownership")
		return nil, fmt.Errorf("failed to verify ownership: %w", err)
	}
	if !ownsPseudonym {
		log.Warn().Int("user_id", userID).Str("pseudonym_id", pseudonymID).Msg("User does not own this pseudonym")
		return nil, fmt.Errorf("unauthorized")
	}

	if input.Body.DisplayName != pseudonym.DisplayName {
		existing, _ := h.securePseudonymDAO.GetPseudonymByDisplayName(ctx, input.Body.DisplayName)
		if existing != nil {
			return nil, fmt.Errorf("display name is already taken")
		}
	}
	updates := &models.PseudonymSetter{
		DisplayName: &input.Body.DisplayName,
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
	err = h.securePseudonymDAO.UpdatePseudonym(ctx, pseudonymID, updates)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", pseudonymID).Msg("Failed to update pseudonym in database")
		return nil, fmt.Errorf("failed to update pseudonym: %w", err)
	}
	finalPseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, pseudonymID)
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
	existing, _ := h.securePseudonymDAO.GetPseudonymByDisplayName(ctx, displayName)
	if existing != nil {
		return nil, fmt.Errorf("display name is already taken")
	}

	// ✅ Use new method that creates pseudonym and identity mapping together
	pseudonym, err := h.securePseudonymDAO.CreatePseudonymWithIdentityMapping(ctx, int64(userID), displayName)
	if err != nil {
		log.Error().Err(err).Int("user_id", userID).Str("display_name", displayName).Msg("Failed to create pseudonym in database")
		return nil, fmt.Errorf("failed to create pseudonym: %w", err)
	}

	updates := &models.PseudonymSetter{}
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
		err = h.securePseudonymDAO.UpdatePseudonym(ctx, pseudonym.PseudonymID, updates)
		if err != nil {
			log.Error().Err(err).Str("pseudonym_id", pseudonym.PseudonymID).Msg("Failed to update pseudonym with additional fields")
		}
	}
	finalPseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, pseudonym.PseudonymID)
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

	// Get user roles from database to determine which role to use for authentication
	roles := []string{"user"} // Default role
	if user.Roles.Valid {
		rawValue, err := user.Roles.V.Value()
		if err == nil {
			var userRoles []string
			if err := json.Unmarshal(rawValue.([]byte), &userRoles); err == nil && len(userRoles) > 0 {
				roles = userRoles
			}
		}
	}

	// Use role-based access control for getting pseudonyms
	// Use the user's actual roles, not hardcoded "user"
	primaryRole := roles[0] // Use the first role for authentication
	pseudonyms, err := h.securePseudonymDAO.GetPseudonymsByUserID(ctx, int64(userID), primaryRole, "authentication")
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
		}
	}
	email := user.Email
	capabilities := userCtx.Capabilities
	response := apimodels.NewUserProfileResponse(userID, email, roles, capabilities, pseudonymProfiles)
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
	updates := &models.UserPreferenceSetter{}
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
	blockedPseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, blockedPseudonymID)
	if err != nil {
		log.Error().Err(err).Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Failed to get blocked pseudonym from database")
		return nil, fmt.Errorf("failed to get blocked pseudonym: %w", err)
	}
	if blockedPseudonym == nil {
		log.Warn().Str("blocked_pseudonym_id", blockedPseudonymID).Msg("Blocked pseudonym not found")
		return nil, huma.Error404NotFound("Blocked pseudonym not found")
	}

	// Use role-based access control for ownership verification
	ownsPseudonym, err := h.securePseudonymDAO.VerifyPseudonymOwnership(ctx, blockedPseudonymID, userID, "user", "self_correlation")
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
		blockedUserID, err := h.securePseudonymDAO.GetUserIDByPseudonym(ctx, blockedPseudonymID, "user", "self_correlation")
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
		blockedUserID, err := h.securePseudonymDAO.GetUserIDByPseudonym(ctx, blockedPseudonymID, "user", "self_correlation")
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
