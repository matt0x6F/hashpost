package handlers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// SubforumHandler handles subforum-related requests
type SubforumHandler struct {
	subforumDAO             *dao.SubforumDAO
	subforumSubscriptionDAO *dao.SubforumSubscriptionDAO
	permissionDAO           *dao.PermissionDAO
	db                      bob.Executor
}

// NewSubforumHandler creates a new subforum handler
func NewSubforumHandler(db bob.Executor) *SubforumHandler {
	return &SubforumHandler{
		subforumDAO:             dao.NewSubforumDAO(db),
		subforumSubscriptionDAO: dao.NewSubforumSubscriptionDAO(db),
		permissionDAO:           dao.NewPermissionDAO(db),
		db:                      db,
	}
}

// GetSubforums handles getting a list of subforums
func (h *SubforumHandler) GetSubforums(ctx context.Context, input *models.SubforumListInput) (*models.SubforumsListResponse, error) {
	log.Info().
		Str("endpoint", "subforums").
		Str("component", "handler").
		Int("page", input.Page).
		Int("limit", input.Limit).
		Str("sort", input.Sort).
		Msg("Get subforums requested")

	// Extract user context for permission checks
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		log.Debug().Msg("No user context found, proceeding as anonymous user")
	}

	// Get subforums from database
	subforums, err := h.subforumDAO.ListSubforums(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subforums from database")
		return nil, fmt.Errorf("failed to get subforums: %w", err)
	}

	// Filter subforums based on user permissions
	filteredSubforums := make([]*dbmodels.Subforum, 0)
	for _, subforum := range subforums {
		// Check if subforum is private and user has access
		if subforum.IsPrivate.Valid && subforum.IsPrivate.V {
			if userCtx == nil {
				// Anonymous users cannot access private subforums
				continue
			}

			canAccess, err := h.permissionDAO.CanAccessPrivateSubforum(ctx, userCtx.UserID, subforum.SubforumID)
			if err != nil {
				log.Warn().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to check private subforum access")
				continue
			}

			if !canAccess {
				continue
			}
		}

		filteredSubforums = append(filteredSubforums, subforum)
	}

	// Apply sorting
	h.sortSubforums(filteredSubforums, input.Sort)

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
	if start >= len(filteredSubforums) {
		start = len(filteredSubforums)
	}
	if end > len(filteredSubforums) {
		end = len(filteredSubforums)
	}

	paginatedSubforums := filteredSubforums[start:end]

	// Convert to API models
	apiSubforums := make([]models.Subforum, len(paginatedSubforums))
	for i, subforum := range paginatedSubforums {
		apiSubforums[i] = h.convertSubforumToAPIModel(subforum)
	}

	response := models.NewSubforumListResponse(apiSubforums, page, limit, len(filteredSubforums))

	log.Info().
		Str("endpoint", "subforums").
		Str("component", "handler").
		Int("count", len(apiSubforums)).
		Int("total", len(filteredSubforums)).
		Msg("Get subforums completed")

	return response, nil
}

// sortSubforums sorts subforums based on the specified sort field
func (h *SubforumHandler) sortSubforums(subforums []*dbmodels.Subforum, sortField string) {
	sort.Slice(subforums, func(i, j int) bool {
		switch sortField {
		case "name":
			return subforums[i].Name < subforums[j].Name
		case "subscribers":
			subI := int32(0)
			subJ := int32(0)
			if subforums[i].SubscriberCount.Valid {
				subI = subforums[i].SubscriberCount.V
			}
			if subforums[j].SubscriberCount.Valid {
				subJ = subforums[j].SubscriberCount.V
			}
			return subI > subJ // Descending order
		case "posts":
			postI := int32(0)
			postJ := int32(0)
			if subforums[i].PostCount.Valid {
				postI = subforums[i].PostCount.V
			}
			if subforums[j].PostCount.Valid {
				postJ = subforums[j].PostCount.V
			}
			return postI > postJ // Descending order
		case "created_at":
			timeI := time.Now()
			timeJ := time.Now()
			if subforums[i].CreatedAt.Valid {
				timeI = subforums[i].CreatedAt.V
			}
			if subforums[j].CreatedAt.Valid {
				timeJ = subforums[j].CreatedAt.V
			}
			return timeI.After(timeJ) // Descending order (newest first)
		default:
			// Default to name sorting
			return subforums[i].Name < subforums[j].Name
		}
	})
}

// convertSubforumToAPIModel converts a database subforum model to an API model
func (h *SubforumHandler) convertSubforumToAPIModel(subforum *dbmodels.Subforum) models.Subforum {
	// Extract description
	description := ""
	if subforum.Description.Valid {
		description = subforum.Description.V
	}

	// Extract sidebar text
	sidebarText := ""
	if subforum.SidebarText.Valid {
		sidebarText = subforum.SidebarText.V
	}

	// Extract rules text
	rulesText := ""
	if subforum.RulesText.Valid {
		rulesText = subforum.RulesText.V
	}

	// Extract boolean flags
	isNSFW := false
	if subforum.IsNSFW.Valid {
		isNSFW = subforum.IsNSFW.V
	}

	isPrivate := false
	if subforum.IsPrivate.Valid {
		isPrivate = subforum.IsPrivate.V
	}

	isRestricted := false
	if subforum.IsRestricted.Valid {
		isRestricted = subforum.IsRestricted.V
	}

	// Get subscriber count
	subscriberCount, err := h.subforumSubscriptionDAO.CountSubscriptionsBySubforum(context.Background(), subforum.SubforumID)
	if err != nil {
		log.Warn().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to get subscriber count")
		subscriberCount = 0
	}

	// Get post count - use the stored value for now since we don't have PostDAO access
	postCount := 0
	if subforum.PostCount.Valid {
		postCount = int(subforum.PostCount.V)
	}

	// Convert timestamps
	createdAt := time.Now()
	if subforum.CreatedAt.Valid {
		createdAt = subforum.CreatedAt.V
	}

	updatedAt := time.Now()
	if subforum.UpdatedAt.Valid {
		updatedAt = subforum.UpdatedAt.V
	}

	return models.Subforum{
		Name:            subforum.Name,
		DisplayName:     subforum.DisplayName,
		Description:     description,
		SidebarText:     sidebarText,
		RulesText:       rulesText,
		IsNSFW:          isNSFW,
		IsPrivate:       isPrivate,
		IsRestricted:    isRestricted,
		SubscriberCount: int(subscriberCount),
		PostCount:       postCount,
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}
}

// GetSubforumDetails handles getting detailed information about a specific subforum
func (h *SubforumHandler) GetSubforumDetails(ctx context.Context, input *models.SubforumSubscriptionInput) (*models.SubforumDetailsResponse, error) {
	subforumName := input.SubforumName

	log.Info().
		Str("endpoint", "subforums/details").
		Str("component", "handler").
		Str("subforum_name", subforumName).
		Msg("Get subforum details requested")

	// Extract user context for permission checks
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		log.Debug().Msg("No user context found, proceeding as anonymous user")
	}

	// Get subforum details from database
	subforum, err := h.subforumDAO.GetSubforumByName(ctx, subforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to get subforum from database")
		return nil, fmt.Errorf("failed to get subforum: %w", err)
	}
	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check if user has access to private subforums
	if subforum.IsPrivate.Valid && subforum.IsPrivate.V {
		if userCtx == nil {
			return nil, huma.Error403Forbidden("access denied: private subforum requires authentication")
		}

		canAccess, err := h.permissionDAO.CanAccessPrivateSubforum(ctx, userCtx.UserID, subforum.SubforumID)
		if err != nil {
			log.Error().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to check private subforum access")
			return nil, fmt.Errorf("failed to check access permissions: %w", err)
		}

		if !canAccess {
			return nil, huma.Error403Forbidden("access denied: insufficient permissions for private subforum")
		}
	}

	// Get moderator information
	moderators, err := h.getSubforumModerators(ctx, subforum.SubforumID)
	if err != nil {
		log.Warn().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to get subforum moderators")
		// Continue without moderator information
	}

	// Check subscription status if user is authenticated
	var isSubscribed, isFavorite bool
	if userCtx != nil {
		isSubscribed, err = h.subforumSubscriptionDAO.IsSubscribed(ctx, userCtx.ActivePseudonymID, subforum.SubforumID)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to check subscription status")
		}

		isFavorite, err = h.subforumSubscriptionDAO.IsFavorite(ctx, userCtx.ActivePseudonymID, subforum.SubforumID)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to check favorite status")
		}
	}

	// Convert to API models
	apiSubforum := h.convertSubforumToAPIModel(subforum)
	apiModerators := h.convertModeratorsToAPIModels(moderators)

	response := models.NewSubforumDetailsResponse(apiSubforum, apiModerators, isSubscribed, isFavorite)

	log.Info().
		Str("endpoint", "subforums/details").
		Str("component", "handler").
		Str("subforum_name", subforumName).
		Int("subforum_id", int(subforum.SubforumID)).
		Msg("Get subforum details completed")

	return response, nil
}

// getSubforumModerators retrieves moderators for a subforum
func (h *SubforumHandler) getSubforumModerators(ctx context.Context, subforumID int32) ([]*dbmodels.SubforumModerator, error) {
	moderators, err := dbmodels.SubforumModerators.Query(
		dbmodels.SelectWhere.SubforumModerators.SubforumID.EQ(subforumID),
	).All(ctx, h.db)
	if err != nil {
		return nil, fmt.Errorf("failed to get subforum moderators: %w", err)
	}

	// Load pseudonym relationships for all moderators
	if len(moderators) > 0 {
		err = dbmodels.SubforumModeratorSlice(moderators).LoadPseudonym(ctx, h.db)
		if err != nil {
			log.Warn().Err(err).Int32("subforum_id", subforumID).Msg("Failed to load moderator pseudonyms")
			// Continue without pseudonym data
		}
	}

	return moderators, nil
}

// convertModeratorsToAPIModels converts database moderator models to API models
func (h *SubforumHandler) convertModeratorsToAPIModels(moderators []*dbmodels.SubforumModerator) []models.SubforumModerator {
	apiModerators := make([]models.SubforumModerator, len(moderators))
	for i, moderator := range moderators {
		displayName := moderator.PseudonymID // Fallback to pseudonym ID
		if moderator.R.Pseudonym != nil {
			displayName = moderator.R.Pseudonym.DisplayName
		}

		apiModerators[i] = models.SubforumModerator{
			PseudonymID:   moderator.PseudonymID,
			DisplayName:   displayName,
			ModeratorType: moderator.Role,                  // Use Role field from DB as ModeratorType
			AddedAt:       time.Now().Format(time.RFC3339), // For now, use current time
		}
	}
	return apiModerators
}

// SubscribeToSubforum handles subscribing to a subforum
func (h *SubforumHandler) SubscribeToSubforum(ctx context.Context, input *models.SubforumSubscriptionInput) (*models.SubforumSubscriptionResponse, error) {
	// Extract user from context
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for subscription")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	subforumName := input.SubforumName

	log.Info().
		Str("endpoint", "subforums/subscribe").
		Str("component", "handler").
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", userCtx.ActivePseudonymID).
		Str("subforum_name", subforumName).
		Msg("Subscribe to subforum requested")

	// Get subforum by name
	subforum, err := h.subforumDAO.GetSubforumByName(ctx, subforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to get subforum")
		return nil, fmt.Errorf("failed to get subforum: %w", err)
	}
	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check if user is already subscribed
	isSubscribed, err := h.subforumSubscriptionDAO.IsSubscribed(ctx, userCtx.ActivePseudonymID, subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check subscription status")
		return nil, fmt.Errorf("failed to check subscription status: %w", err)
	}

	if isSubscribed {
		log.Info().
			Str("subforum_name", subforumName).
			Str("pseudonym_id", userCtx.ActivePseudonymID).
			Msg("User already subscribed to subforum")
		return nil, huma.Error409Conflict("already subscribed to subforum")
	}

	// Create subscription
	_, err = h.subforumSubscriptionDAO.CreateSubscription(ctx, userCtx.ActivePseudonymID, subforum.SubforumID, false)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create subscription")
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Get updated subscriber count
	subscriberCount, err := h.subforumSubscriptionDAO.CountSubscriptionsBySubforum(ctx, subforum.SubforumID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get updated subscriber count")
		subscriberCount = 0 // Use 0 if we can't get the count
	}

	response := models.NewSubforumSubscriptionResponse(int(subforum.SubforumID), subforumName, true, int(subscriberCount))

	log.Info().
		Str("endpoint", "subforums/subscribe").
		Str("component", "handler").
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", userCtx.ActivePseudonymID).
		Str("subforum_name", subforumName).
		Int32("subforum_id", subforum.SubforumID).
		Int64("subscriber_count", subscriberCount).
		Msg("Subscribe to subforum completed")

	return response, nil
}

// UnsubscribeFromSubforum handles unsubscribing from a subforum
func (h *SubforumHandler) UnsubscribeFromSubforum(ctx context.Context, input *models.SubforumSubscriptionInput) (*models.SubforumSubscriptionResponse, error) {
	// Extract user from context
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for unsubscription")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	subforumName := input.SubforumName

	log.Info().
		Str("endpoint", "subforums/unsubscribe").
		Str("component", "handler").
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", userCtx.ActivePseudonymID).
		Str("subforum_name", subforumName).
		Msg("Unsubscribe from subforum requested")

	// Get subforum by name
	subforum, err := h.subforumDAO.GetSubforumByName(ctx, subforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to get subforum")
		return nil, fmt.Errorf("failed to get subforum: %w", err)
	}
	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check if user is subscribed
	isSubscribed, err := h.subforumSubscriptionDAO.IsSubscribed(ctx, userCtx.ActivePseudonymID, subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check subscription status")
		return nil, fmt.Errorf("failed to check subscription status: %w", err)
	}

	if !isSubscribed {
		log.Info().
			Str("subforum_name", subforumName).
			Str("pseudonym_id", userCtx.ActivePseudonymID).
			Msg("User not subscribed to subforum")
		return nil, huma.Error409Conflict("not subscribed to subforum")
	}

	// Delete subscription
	err = h.subforumSubscriptionDAO.DeleteSubscription(ctx, userCtx.ActivePseudonymID, subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to delete subscription")
		return nil, fmt.Errorf("failed to delete subscription: %w", err)
	}

	// Get updated subscriber count
	subscriberCount, err := h.subforumSubscriptionDAO.CountSubscriptionsBySubforum(ctx, subforum.SubforumID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get updated subscriber count")
		subscriberCount = 0 // Use 0 if we can't get the count
	}

	response := models.NewSubforumSubscriptionResponse(int(subforum.SubforumID), subforumName, false, int(subscriberCount))

	log.Info().
		Str("endpoint", "subforums/unsubscribe").
		Str("component", "handler").
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", userCtx.ActivePseudonymID).
		Str("subforum_name", subforumName).
		Int32("subforum_id", subforum.SubforumID).
		Int64("subscriber_count", subscriberCount).
		Msg("Unsubscribe from subforum completed")

	return response, nil
}

// CreateSubforum handles creating a new subforum
func (h *SubforumHandler) CreateSubforum(ctx context.Context, input *models.SubforumCreateInput) (*models.SubforumDetailsResponse, error) {
	log.Info().Str("endpoint", "subforums/create").Str("component", "handler").Msg("Create subforum requested")

	// Debug: Log the received input
	log.Debug().
		Str("slug", input.Body.Slug).
		Str("name", input.Body.Name).
		Str("description", input.Body.Description).
		Str("sidebar_text", input.Body.SidebarText).
		Str("rules_text", input.Body.RulesText).
		Bool("is_nsfw", input.Body.IsNSFW).
		Bool("is_private", input.Body.IsPrivate).
		Bool("is_restricted", input.Body.IsRestricted).
		Msg("Received subforum creation input")

	// Extract user context from the authentication fields
	authInput := &middleware.AuthInput{
		Authorization: input.Authorization,
		AccessToken:   input.AccessToken,
	}
	userCtx, err := middleware.ExtractUserFromHumaInput(authInput)
	if err != nil || userCtx == nil {
		log.Warn().Msg("Authentication required for subforum creation")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Check capability
	if !userCtx.HasCapability("create_subforum") {
		log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks create_subforum capability")
		return nil, huma.Error403Forbidden("insufficient permissions to create subforum")
	}

	// Validate required fields
	if input.Body.Slug == "" {
		return nil, huma.Error400BadRequest("slug is required")
	}
	if input.Body.Name == "" {
		return nil, huma.Error400BadRequest("name is required")
	}
	if input.Body.Description == "" {
		return nil, huma.Error400BadRequest("description is required")
	}

	// Only admins can set is_restricted; otherwise, force to false
	isRestricted := false
	if userCtx.HasCapability("system_admin") || userCtx.HasCapability("user_management") {
		isRestricted = input.Body.IsRestricted
	}

	// Use defaults for optional fields if not provided
	sidebarText := input.Body.SidebarText
	rulesText := input.Body.RulesText
	isNSFW := input.Body.IsNSFW
	isPrivate := input.Body.IsPrivate

	// Create the subforum in the database
	subforum, err := h.subforumDAO.CreateSubforum(
		ctx,
		input.Body.Slug, // Slug is used as the unique identifier (maps to db 'name')
		input.Body.Name,
		input.Body.Description,
		sidebarText,
		rulesText,
		isNSFW,
		isPrivate,
		isRestricted,
	)
	if err != nil {
		log.Error().Err(err).Str("slug", input.Body.Slug).Msg("Failed to create subforum")
		return nil, huma.Error400BadRequest(err.Error())
	}

	// Convert to API model
	apiSubforum := h.convertSubforumToAPIModel(subforum)

	// For now, moderators, isSubscribed, isFavorite are empty/default
	return models.NewSubforumDetailsResponse(apiSubforum, nil, false, false), nil
}
