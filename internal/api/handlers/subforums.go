package handlers

import (
	"context"
	"fmt"
	"sort"
	"time"

	"database/sql"
	"encoding/json"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
)

// SubforumHandler handles subforum-related requests
type SubforumHandler struct {
	subforumDAO             dao.SubforumDAOInterface
	subforumSubscriptionDAO dao.SubforumSubscriptionDAOInterface
	permissionDAO           dao.PermissionDAOInterface
	subforumModeratorDAO    dao.SubforumModeratorDAOInterface
	identityMappingDAO      dao.IdentityMappingDAOInterface
	pseudonymDAO            dao.PseudonymDAOInterface
	postDAO                 dao.PostDAOInterface
	db                      bob.Executor
}

// NewSubforumHandler creates a new subforum handler with optional dependencies
// If db is provided, real DAOs will be created. If nil, mock DAOs should be provided.
func NewSubforumHandler(
	db bob.Executor,
	subforumDAO dao.SubforumDAOInterface,
	subforumSubscriptionDAO dao.SubforumSubscriptionDAOInterface,
	permissionDAO dao.PermissionDAOInterface,
	subforumModeratorDAO dao.SubforumModeratorDAOInterface,
	identityMappingDAO dao.IdentityMappingDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
	postDAO dao.PostDAOInterface,
) *SubforumHandler {
	// If db is provided, create real DAOs (production mode)
	if db != nil {
		subforumDAO = dao.NewSubforumDAO(db)
		subforumSubscriptionDAO = dao.NewSubforumSubscriptionDAO(db)
		permissionDAO = dao.NewPermissionDAO(db)
		subforumModeratorDAO = dao.NewSubforumModeratorDAO(db)
		identityMappingDAO = dao.NewIdentityMappingDAO(db)
		// Note: pseudonymDAO requires additional dependencies, so it should be passed in
		postDAO = dao.NewPostDAO(db)
	}

	return &SubforumHandler{
		subforumDAO:             subforumDAO,
		subforumSubscriptionDAO: subforumSubscriptionDAO,
		permissionDAO:           permissionDAO,
		subforumModeratorDAO:    subforumModeratorDAO,
		identityMappingDAO:      identityMappingDAO,
		pseudonymDAO:            pseudonymDAO,
		postDAO:                 postDAO,
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

			canAccess, err := h.permissionDAO.CanAccessPrivateSubforumWithActivePseudonym(ctx, userCtx.UserID, subforum.SubforumID, userCtx.ActivePseudonymID)
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

	// Get actual post count from database
	postCount, err := h.postDAO.CountPostsBySubforum(context.Background(), subforum.SubforumID)
	if err != nil {
		log.Warn().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to get post count")
		postCount = 0
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
		PostCount:       int(postCount),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
		CommunityType:   subforum.CommunityType,
		GovernanceStyle: subforum.GovernanceStyle,
		OwnerPseudonymID: func() string {
			if subforum.OwnerPseudonymID.Valid {
				return subforum.OwnerPseudonymID.V
			}
			return ""
		}(),
	}
}

// GetSubforumDetails handles getting detailed information about a specific subforum
func (h *SubforumHandler) GetSubforumDetails(ctx context.Context, input *struct {
	middleware.AuthInput
	models.SubforumSubscriptionInput
}) (*models.SubforumDetailsResponse, error) {
	communityType := input.SubforumSubscriptionInput.CommunityType
	subforumName := input.SubforumSubscriptionInput.SubforumName

	log.Info().
		Str("endpoint", "subforums/details").
		Str("component", "handler").
		Str("community_type", communityType).
		Str("subforum_name", subforumName).
		Msg("Get subforum details requested")

	// Try to extract user context for subscription checks
	var userCtx *middleware.UserContext
	var err error
	if input.AuthInput.Authorization != "" || input.AuthInput.AccessToken != "" {
		// If Authorization header or AccessToken cookie is present, try to extract user context
		userCtx, err = middleware.ExtractUserFromHumaInput(&input.AuthInput)
		if err != nil {
			log.Debug().Msg("Failed to extract user context from auth input, proceeding as anonymous user")
		}
	} else {
		log.Debug().Msg("No auth token found, proceeding as anonymous user")
	}

	// Get subforum details from database using community type and name
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, communityType, subforumName)
	if err != nil {
		log.Error().Err(err).Str("community_type", communityType).Str("subforum_name", subforumName).Msg("Failed to get subforum from database")
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

		canAccess, err := h.permissionDAO.CanAccessPrivateSubforumWithActivePseudonym(ctx, userCtx.UserID, subforum.SubforumID, userCtx.ActivePseudonymID)
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
		Str("community_type", communityType).
		Str("subforum_name", subforumName).
		Int("subforum_id", int(subforum.SubforumID)).
		Msg("Get subforum details completed")

	return response, nil
}

// getSubforumModerators retrieves moderators for a subforum
func (h *SubforumHandler) getSubforumModerators(ctx context.Context, subforumID int32) ([]*dbmodels.SubforumModerator, error) {
	moderators, err := h.subforumModeratorDAO.GetModeratorsBySubforum(ctx, subforumID)
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
func (h *SubforumHandler) SubscribeToSubforum(ctx context.Context, input *struct {
	middleware.AuthInput
	models.SubforumSubscriptionInput
}) (*models.SubforumSubscriptionResponse, error) {
	// Extract user from input
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for subscription")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	subforumName := input.SubforumSubscriptionInput.SubforumName

	log.Info().
		Str("endpoint", "subforums/subscribe").
		Str("component", "handler").
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", userCtx.ActivePseudonymID).
		Str("subforum_name", subforumName).
		Msg("Subscribe to subforum requested")

	// Get subforum by community type and name
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.SubforumSubscriptionInput.CommunityType, subforumName)
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
func (h *SubforumHandler) UnsubscribeFromSubforum(ctx context.Context, input *struct {
	middleware.AuthInput
	models.SubforumSubscriptionInput
}) (*models.SubforumSubscriptionResponse, error) {
	// Extract user from input
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for unsubscription")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	subforumName := input.SubforumSubscriptionInput.SubforumName

	log.Info().
		Str("endpoint", "subforums/unsubscribe").
		Str("component", "handler").
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", userCtx.ActivePseudonymID).
		Str("subforum_name", subforumName).
		Msg("Unsubscribe from subforum requested")

	// Get subforum by community type and name
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.SubforumSubscriptionInput.CommunityType, subforumName)
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

	// Extract user from context
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for subforum creation")
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

	// Enforce governance style based on community type
	var governanceStyle string
	switch input.Body.CommunityType {
	case "t", "g":
		governanceStyle = "democratic"
	case "b", "c":
		governanceStyle = "owned"
	default:
		return nil, huma.Error400BadRequest("invalid community type")
	}

	// Create the subforum in the database
	subforum, err := h.subforumDAO.CreateSubforum(
		ctx,
		input.Body.Slug, // Slug is used as the unique identifier (maps to db 'name')
		input.Body.Name,
		input.Body.Description,
		sidebarText,
		rulesText,
		input.Body.CommunityType,
		governanceStyle, // Use enforced governance style, not from request
		isNSFW,
		isPrivate,
		isRestricted,
		userCtx.ActivePseudonymID, // Owner is the creating user's active pseudonym
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

// GetSubforumSettings returns the settings for a specific subforum
func (h *SubforumHandler) GetSubforumSettings(ctx context.Context, input *struct {
	middleware.AuthInput
	models.SubforumSettingsGetInput
}) (*models.SubforumSettingsResponse, error) {
	log.Info().Str("endpoint", "subforums/settings").Str("component", "handler").Msg("Get subforum settings requested")

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for subforum settings")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Get subforum first
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.Type, input.Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subforum")
		return nil, fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check capability for managing subforum settings using unified permission system
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityManageSubforumSettings, &subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check unified capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks manage_subforum_settings capability")
		return nil, huma.Error403Forbidden("insufficient permissions to view subforum settings")
	}

	// Convert database model to settings
	settings := models.SubforumSettings{
		AllowImages:  subforum.AllowImages.Valid && subforum.AllowImages.V,
		AllowVideos:  subforum.AllowVideos.Valid && subforum.AllowVideos.V,
		AllowPolls:   subforum.AllowPolls.Valid && subforum.AllowPolls.V,
		RequireFlair: subforum.RequireFlair.Valid && subforum.RequireFlair.V,
		MinimumAccountAgeHours: func() int {
			if subforum.MinimumAccountAgeHours.Valid {
				return int(subforum.MinimumAccountAgeHours.V)
			}
			return 0
		}(),
		MinimumKarmaRequired: func() int {
			if subforum.MinimumKarmaRequired.Valid {
				return int(subforum.MinimumKarmaRequired.V)
			}
			return 0
		}(),
		IsPrivate:    subforum.IsPrivate.Valid && subforum.IsPrivate.V,
		IsRestricted: subforum.IsRestricted.Valid && subforum.IsRestricted.V,
		IsNSFW:       subforum.IsNSFW.Valid && subforum.IsNSFW.V,
		// Note: These fields need to be added to the database schema in a future migration
		// For now, using sensible defaults until the schema is updated
		AutoModerationEnabled: false,
		RequireApproval:       false,
		AllowCrossposts:       true,
		Description: func() string {
			if subforum.Description.Valid {
				return subforum.Description.V
			}
			return ""
		}(),
		SidebarText: func() string {
			if subforum.SidebarText.Valid {
				return subforum.SidebarText.V
			}
			return ""
		}(),
	}

	updatedAt := time.Now()
	if subforum.UpdatedAt.Valid {
		updatedAt = subforum.UpdatedAt.V
	}

	return &models.SubforumSettingsResponse{
		Status: 200,
		Body: struct {
			SubforumID int32                   `json:"subforum_id" example:"123"`
			Name       string                  `json:"name" example:"golang"`
			Settings   models.SubforumSettings `json:"settings"`
			UpdatedAt  string                  `json:"updated_at" example:"2023-01-01T00:00:00Z"`
		}{
			SubforumID: subforum.SubforumID,
			Name:       subforum.Name,
			Settings:   settings,
			UpdatedAt:  updatedAt.Format(time.RFC3339),
		},
	}, nil
}

// UpdateSubforumSettings updates the settings for a specific subforum
func (h *SubforumHandler) UpdateSubforumSettings(ctx context.Context, input *struct {
	middleware.AuthInput
	models.SubforumSettingsInput
}) (*models.SubforumSettingsResponse, error) {
	log.Info().Str("endpoint", "subforums/settings").Str("component", "handler").Msg("Update subforum settings requested")

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for subforum settings")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Get subforum first
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.Type, input.Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subforum")
		return nil, fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check capability for managing subforum settings using unified permission system
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityManageSubforumSettings, &subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check unified capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks manage_subforum_settings capability")
		return nil, huma.Error403Forbidden("insufficient permissions to update subforum settings")
	}

	// Update subforum settings
	updateSetter := &dbmodels.SubforumSetter{
		AllowImages:            &sql.Null[bool]{V: input.Body.AllowImages, Valid: true},
		AllowVideos:            &sql.Null[bool]{V: input.Body.AllowVideos, Valid: true},
		AllowPolls:             &sql.Null[bool]{V: input.Body.AllowPolls, Valid: true},
		RequireFlair:           &sql.Null[bool]{V: input.Body.RequireFlair, Valid: true},
		MinimumAccountAgeHours: &sql.Null[int32]{V: int32(input.Body.MinimumAccountAgeHours), Valid: true},
		MinimumKarmaRequired:   &sql.Null[int32]{V: int32(input.Body.MinimumKarmaRequired), Valid: true},
		IsPrivate:              &sql.Null[bool]{V: input.Body.IsPrivate, Valid: true},
		IsRestricted:           &sql.Null[bool]{V: input.Body.IsRestricted, Valid: true},
		IsNSFW:                 &sql.Null[bool]{V: input.Body.IsNSFW, Valid: true},
		Description:            &sql.Null[string]{V: input.Body.Description, Valid: true},
		SidebarText:            &sql.Null[string]{V: input.Body.SidebarText, Valid: true},
		UpdatedAt:              &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	if err := subforum.Update(ctx, h.db, updateSetter); err != nil {
		log.Error().Err(err).Msg("Failed to update subforum settings")
		return nil, fmt.Errorf("failed to update subforum settings: %w", err)
	}

	log.Info().
		Str("endpoint", "subforums/settings").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int32("subforum_id", subforum.SubforumID).
		Msg("Subforum settings updated")

	return &models.SubforumSettingsResponse{
		Status: 200,
		Body: struct {
			SubforumID int32                   `json:"subforum_id" example:"123"`
			Name       string                  `json:"name" example:"golang"`
			Settings   models.SubforumSettings `json:"settings"`
			UpdatedAt  string                  `json:"updated_at" example:"2023-01-01T00:00:00Z"`
		}{
			SubforumID: subforum.SubforumID,
			Name:       subforum.Name,
			Settings:   input.Body,
			UpdatedAt:  time.Now().Format(time.RFC3339),
		},
	}, nil
}

// GetModeratorTeam returns the moderator team for a specific subforum
func (h *SubforumHandler) GetModeratorTeam(ctx context.Context, input *struct {
	middleware.AuthInput
	models.ModeratorTeamInput
}) (*models.ModeratorTeamResponse, error) {
	log.Info().Str("endpoint", "subforums/moderator-team").Str("component", "handler").Msg("Get moderator team requested")

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for moderator team")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Get subforum first
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.Type, input.Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subforum")
		return nil, fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check capability for managing moderators using unified permission system
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityManageModerators, &subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check unified capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks manage_moderators capability")
		return nil, huma.Error403Forbidden("insufficient permissions to view moderator team")
	}

	// Get moderator team
	moderators, err := h.subforumModeratorDAO.GetModeratorsBySubforum(ctx, subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get moderator team")
		return nil, fmt.Errorf("failed to get moderator team: %w", err)
	}

	// Convert to API models
	var members []models.ModeratorTeamMember
	var owner models.ModeratorTeamMember

	for _, mod := range moderators {
		// Get pseudonym details to get display name
		pseudonym, err := h.pseudonymDAO.GetPseudonymByID(ctx, mod.PseudonymID)
		displayName := ""
		if err == nil && pseudonym != nil {
			displayName = pseudonym.DisplayName
		}

		member := models.ModeratorTeamMember{
			PseudonymID:  mod.PseudonymID,
			DisplayName:  displayName,
			Role:         mod.Role,
			Capabilities: []string{}, // Will be populated from permissions JSON below
			AddedAt: func() string {
				if mod.AddedAt.Valid {
					return mod.AddedAt.V.Format(time.RFC3339)
				}
				return ""
			}(),
			AddedBy: func() string {
				if mod.AddedByPseudonymID.Valid {
					return mod.AddedByPseudonymID.V
				}
				return ""
			}(),
			// Note: IsActive field needs to be added to the database schema in a future migration
			// For now, assuming all moderators are active
			IsActive: true,
		}

		// Parse capabilities from permissions JSON
		if mod.Permissions.Valid {
			rawValue, err := mod.Permissions.V.Value()
			if err == nil {
				if bytes, ok := rawValue.([]byte); ok {
					var permissions []string
					if err := json.Unmarshal(bytes, &permissions); err == nil {
						member.Capabilities = permissions
					}
				}
			}
		}

		// Determine if this is the owner
		if subforum.OwnerPseudonymID.Valid && mod.PseudonymID == subforum.OwnerPseudonymID.V {
			owner = member
		} else {
			members = append(members, member)
		}
	}

	return &models.ModeratorTeamResponse{
		Status: 200,
		Body: struct {
			SubforumID int32                        `json:"subforum_id" example:"123"`
			Name       string                       `json:"name" example:"golang"`
			Members    []models.ModeratorTeamMember `json:"members"`
			Owner      models.ModeratorTeamMember   `json:"owner"`
		}{
			SubforumID: subforum.SubforumID,
			Name:       subforum.Name,
			Members:    members,
			Owner:      owner,
		},
	}, nil
}

// AddModerator adds a new moderator to the subforum
func (h *SubforumHandler) AddModerator(ctx context.Context, input *struct {
	middleware.AuthInput
	models.AddModeratorInput
}) (*models.ModeratorTeamMember, error) {
	log.Info().Str("endpoint", "subforums/add-moderator").Str("component", "handler").Msg("Add moderator requested")

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for adding moderator")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Get subforum first
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.Type, input.Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subforum")
		return nil, fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check capability for managing moderators using unified permission system
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityManageModerators, &subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check unified capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks manage_moderators capability")
		return nil, huma.Error403Forbidden("insufficient permissions to add moderator")
	}

	// Add moderator
	if _, err := h.subforumModeratorDAO.CreateModerator(ctx, subforum.SubforumID, input.Body.PseudonymID, input.Body.Role, userCtx.ActivePseudonymID); err != nil {
		log.Error().Err(err).Msg("Failed to add moderator")
		return nil, fmt.Errorf("failed to add moderator: %w", err)
	}

	log.Info().
		Str("endpoint", "subforums/add-moderator").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int32("subforum_id", subforum.SubforumID).
		Str("moderator_id", input.Body.PseudonymID).
		Msg("Moderator added")

	return &models.ModeratorTeamMember{
		PseudonymID:  input.Body.PseudonymID,
		DisplayName:  "", // TODO: Get from pseudonym table
		Role:         input.Body.Role,
		Capabilities: input.Body.Capabilities,
		AddedAt:      time.Now().Format(time.RFC3339),
		AddedBy:      userCtx.ActivePseudonymID,
		IsActive:     true,
	}, nil
}

// UpdateModerator updates an existing moderator's permissions
func (h *SubforumHandler) UpdateModerator(ctx context.Context, input *struct {
	middleware.AuthInput
	models.UpdateModeratorInput
}) (*models.ModeratorTeamMember, error) {
	log.Info().Str("endpoint", "subforums/update-moderator").Str("component", "handler").Msg("Update moderator requested")

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for updating moderator")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Get subforum first
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.Type, input.Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subforum")
		return nil, fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check capability for managing moderators using unified permission system
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityManageModerators, &subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check unified capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks manage_moderators capability")
		return nil, huma.Error403Forbidden("insufficient permissions to update moderator")
	}

	// Update moderator role
	if err := h.subforumModeratorDAO.UpdateModeratorRole(ctx, input.PseudonymID, subforum.SubforumID, input.Body.Role); err != nil {
		log.Error().Err(err).Msg("Failed to update moderator")
		return nil, fmt.Errorf("failed to update moderator: %w", err)
	}

	log.Info().
		Str("endpoint", "subforums/update-moderator").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int32("subforum_id", subforum.SubforumID).
		Str("moderator_id", input.PseudonymID).
		Msg("Moderator updated")

	return &models.ModeratorTeamMember{
		PseudonymID:  input.PseudonymID,
		DisplayName:  "", // TODO: Get from pseudonym table
		Role:         input.Body.Role,
		Capabilities: input.Body.Capabilities,
		AddedAt:      "", // TODO: Get from existing record
		AddedBy:      "", // TODO: Get from existing record
		IsActive:     input.Body.IsActive,
	}, nil
}

// RemoveModerator removes a moderator from the subforum
func (h *SubforumHandler) RemoveModerator(ctx context.Context, input *struct {
	middleware.AuthInput
	models.RemoveModeratorInput
}) (*models.RemoveModeratorResponse, error) {
	log.Info().Str("endpoint", "subforums/remove-moderator").Str("component", "handler").Msg("Remove moderator requested")

	// Extract user from AuthInput
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Authentication required for removing moderator")
		return nil, huma.Error401Unauthorized("authentication required")
	}

	// Get subforum first
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.Type, input.Name)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subforum")
		return nil, fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Check capability for managing moderators using unified permission system
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityManageModerators, &subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check unified capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks manage_moderators capability")
		return nil, huma.Error403Forbidden("insufficient permissions to remove moderator")
	}

	// Remove moderator
	if err := h.subforumModeratorDAO.DeleteModerator(ctx, input.PseudonymID, subforum.SubforumID); err != nil {
		log.Error().Err(err).Msg("Failed to remove moderator")
		return nil, fmt.Errorf("failed to remove moderator: %w", err)
	}

	log.Info().
		Str("endpoint", "subforums/remove-moderator").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int32("subforum_id", subforum.SubforumID).
		Str("moderator_id", input.PseudonymID).
		Msg("Moderator removed")

	return &models.RemoveModeratorResponse{
		Success: true,
		Message: "Moderator removed successfully",
	}, nil
}

// GetPseudonymSubscriptions handles GET /pseudonyms/{pseudonym_id}/subscriptions
func (h *SubforumHandler) GetPseudonymSubscriptions(ctx context.Context, input *struct {
	middleware.AuthInput
	models.PseudonymSubscriptionsInput
}) (*models.SubforumSubscriptionsResponse, error) {
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		log := zerolog.Ctx(ctx)
		log.Error().Err(err).Msg("Failed to extract user from context in GetPseudonymSubscriptions")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	log := zerolog.Ctx(ctx)
	log.Info().Int64("user_id", userCtx.UserID).Str("pseudonym_id", input.PseudonymSubscriptionsInput.PseudonymID).Msg("Checking pseudonym ownership for subscriptions")

	// Only allow if the pseudonym belongs to the user
	identityMappings, err := h.identityMappingDAO.GetIdentityMappingsByUserID(ctx, userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to fetch user identity mappings")
		return nil, huma.Error500InternalServerError("Failed to fetch user pseudonyms")
	}

	// Check if the requested pseudonym_id is owned by this user
	userOwnsPseudonym := false
	for _, mapping := range identityMappings {
		if mapping.PseudonymID == input.PseudonymSubscriptionsInput.PseudonymID {
			userOwnsPseudonym = true
			break
		}
	}

	if !userOwnsPseudonym {
		log.Warn().Int64("user_id", userCtx.UserID).Str("pseudonym_id", input.PseudonymSubscriptionsInput.PseudonymID).Msg("User does not own the requested pseudonym")
		return nil, huma.Error403Forbidden("Access denied: pseudonym not owned by user")
	}

	// Get subscriptions for the pseudonym
	subscriptions, err := h.subforumSubscriptionDAO.GetSubscriptionsByPseudonym(ctx, input.PseudonymSubscriptionsInput.PseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", input.PseudonymSubscriptionsInput.PseudonymID).Msg("Failed to fetch pseudonym subscriptions")
		return nil, huma.Error500InternalServerError("Failed to fetch subscriptions")
	}

	// Convert to API models
	apiSubforums := make([]models.Subforum, 0, len(subscriptions))
	for _, subscription := range subscriptions {
		// Get the full subforum details
		subforum, err := h.subforumDAO.GetSubforumByID(ctx, subscription.SubforumID)
		if err != nil {
			log.Error().Err(err).Int32("subforum_id", subscription.SubforumID).Msg("Failed to fetch subforum details")
			continue
		}

		// Handle nullable fields
		description := ""
		if subforum.Description.Valid {
			description = subforum.Description.V
		}

		sidebarText := ""
		if subforum.SidebarText.Valid {
			sidebarText = subforum.SidebarText.V
		}

		rulesText := ""
		if subforum.RulesText.Valid {
			rulesText = subforum.RulesText.V
		}

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

		subscriberCount := 0
		if subforum.SubscriberCount.Valid {
			subscriberCount = int(subforum.SubscriberCount.V)
		}

		// Get actual post count from database
		postCount, err := h.postDAO.CountPostsBySubforum(ctx, subforum.SubforumID)
		if err != nil {
			log.Warn().Err(err).Int32("subforum_id", subforum.SubforumID).Msg("Failed to get post count")
			postCount = 0
		}

		createdAt := time.Now()
		if subforum.CreatedAt.Valid {
			createdAt = subforum.CreatedAt.V
		}

		updatedAt := time.Now()
		if subforum.UpdatedAt.Valid {
			updatedAt = subforum.UpdatedAt.V
		}

		apiSubforums = append(apiSubforums, models.Subforum{
			Name:            subforum.Name,
			DisplayName:     subforum.DisplayName,
			Description:     description,
			SidebarText:     sidebarText,
			RulesText:       rulesText,
			IsNSFW:          isNSFW,
			IsPrivate:       isPrivate,
			IsRestricted:    isRestricted,
			SubscriberCount: subscriberCount,
			PostCount:       int(postCount),
			CreatedAt:       createdAt,
			UpdatedAt:       updatedAt,
		})
	}

	log.Info().
		Str("endpoint", "pseudonyms/subscriptions").
		Str("component", "handler").
		Int64("user_id", userCtx.UserID).
		Str("pseudonym_id", input.PseudonymSubscriptionsInput.PseudonymID).
		Int("subscription_count", len(apiSubforums)).
		Msg("Successfully retrieved pseudonym subscriptions")

	return &models.SubforumSubscriptionsResponse{
		Status: 200,
		Body: models.SubforumSubscriptionsResponseBody{
			Subforums: apiSubforums,
		},
	}, nil
}
