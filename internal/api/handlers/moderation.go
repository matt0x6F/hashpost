package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob/types"
)

// ModerationHandler handles moderation-related requests
type ModerationHandler struct {
	reportDAO           dao.ReportDAOInterface
	moderationActionDAO dao.ModerationActionDAOInterface
	userBanDAO          dao.UserBanDAOInterface
	securePseudonymDAO  dao.PseudonymDAOInterface
	subforumDAO         dao.SubforumDAOInterface
	postDAO             dao.PostDAOInterface
	commentDAO          dao.CommentDAOInterface
	voteDAO             dao.VoteDAOInterface
	permissionDAO       dao.PermissionDAOInterface
}

// NewModerationHandler creates a new moderation handler with interface dependencies
func NewModerationHandler(
	reportDAO dao.ReportDAOInterface,
	moderationActionDAO dao.ModerationActionDAOInterface,
	userBanDAO dao.UserBanDAOInterface,
	securePseudonymDAO dao.PseudonymDAOInterface,
	subforumDAO dao.SubforumDAOInterface,
	postDAO dao.PostDAOInterface,
	commentDAO dao.CommentDAOInterface,
	voteDAO dao.VoteDAOInterface,
	permissionDAO dao.PermissionDAOInterface,
) *ModerationHandler {
	return &ModerationHandler{
		reportDAO:           reportDAO,
		moderationActionDAO: moderationActionDAO,
		userBanDAO:          userBanDAO,
		securePseudonymDAO:  securePseudonymDAO,
		subforumDAO:         subforumDAO,
		postDAO:             postDAO,
		commentDAO:          commentDAO,
		voteDAO:             voteDAO,
		permissionDAO:       permissionDAO,
	}
}

// extractUserFromContext extracts user information from the request context
func (h *ModerationHandler) extractUserFromContext(ctx context.Context) (*middleware.UserContext, error) {
	userCtx, err := middleware.ExtractUserFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user from context: %w", err)
	}
	return userCtx, nil
}

// validateModeratorPermissions validates that the user has moderator permissions for a specific subforum
func (h *ModerationHandler) validateModeratorPermissionsForSubforum(ctx context.Context, userCtx *middleware.UserContext, subforumPath string) error {
	// Parse subforum path to extract community type and name
	communityType, subforumName, err := h.parseSubforumPath(subforumPath)
	if err != nil {
		log.Error().Err(err).Str("subforum_path", subforumPath).Msg("Failed to parse subforum path")
		return fmt.Errorf("invalid subforum path format: %w", err)
	}

	// Get subforum by community type and name
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, communityType, subforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_path", subforumPath).Msg("Failed to get subforum")
		return fmt.Errorf("failed to get subforum: %w", err)
	}
	if subforum == nil {
		return fmt.Errorf("subforum not found: %s", subforumPath)
	}

	// Check if the active pseudonym has moderator capabilities for this subforum
	activePseudonymID := userCtx.ActivePseudonymID
	if activePseudonymID == "" {
		return fmt.Errorf("no active pseudonym found")
	}

	// First, check if the active pseudonym is the owner of this subforum
	isOwner := subforum.OwnerPseudonymID.Valid && subforum.OwnerPseudonymID.V == activePseudonymID
	if isOwner {
		log.Debug().
			Int64("user_id", userCtx.UserID).
			Int32("subforum_id", subforum.SubforumID).
			Str("active_pseudonym_id", activePseudonymID).
			Msg("Found owner record - granting moderator access")
		return nil
	}

	// Check if the active pseudonym is a moderator for this subforum
	hasModerateContent, err := h.permissionDAO.HasSubforumCapabilityWithActivePseudonym(ctx, userCtx.UserID, subforum.SubforumID, "moderate_content", activePseudonymID)
	if err != nil {
		log.Error().Err(err).
			Int64("user_id", userCtx.UserID).
			Int32("subforum_id", subforum.SubforumID).
			Str("active_pseudonym_id", activePseudonymID).
			Msg("Failed to check moderator capabilities")
		return fmt.Errorf("failed to check moderator capabilities: %w", err)
	}

	if !hasModerateContent {
		log.Warn().
			Int64("user_id", userCtx.UserID).
			Int32("subforum_id", subforum.SubforumID).
			Str("active_pseudonym_id", activePseudonymID).
			Msg("User does not have moderator capabilities for this subforum")
		return fmt.Errorf("user does not have moderation permissions for this subforum")
	}

	log.Debug().
		Int64("user_id", userCtx.UserID).
		Int32("subforum_id", subforum.SubforumID).
		Str("active_pseudonym_id", activePseudonymID).
		Msg("User has moderator capabilities for this subforum")
	return nil
}

// parseSubforumPath parses a full subforum path (e.g., "b/hashpost") into community type and name
func (h *ModerationHandler) parseSubforumPath(fullPath string) (communityType, subforumName string, err error) {
	// Handle different formats:
	// 1. "b/hashpost" -> communityType: "b", subforumName: "hashpost"
	// 2. "hashpost" -> communityType: "h", subforumName: "hashpost" (default for h/ subforums)

	if fullPath == "" {
		return "", "", fmt.Errorf("subforum path cannot be empty")
	}

	// Check if it contains a slash (community type prefix)
	if strings.Contains(fullPath, "/") {
		parts := strings.SplitN(fullPath, "/", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid subforum path format: expected 'community-type/name'")
		}

		communityType = parts[0]
		subforumName = parts[1]

		// Validate community type
		validTypes := []string{"t", "g", "b", "c", "h"}
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
	return "h", fullPath, nil
}

// validateModeratorPermissions validates that the user has moderator permissions using unified permissions
func (h *ModerationHandler) validateModeratorPermissions(userCtx *middleware.UserContext) error {
	log.Info().
		Int("user_id", int(userCtx.UserID)).
		Str("active_pseudonym_id", userCtx.ActivePseudonymID).
		Msg("Validating moderator permissions using unified system")

	// For global moderation endpoints, check platform-wide capabilities
	// Get the unified roles and capabilities without a specific subforum
	_, capabilities, err := h.permissionDAO.GetUnifiedActivePseudonymRolesAndCapabilities(
		context.Background(),
		userCtx.UserID,
		userCtx.ActivePseudonymID,
		nil, // No specific subforum for global moderation
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get unified capabilities")
		return huma.Error500InternalServerError("Failed to validate permissions")
	}

	// Check for platform-wide moderation capabilities
	hasModerateContent := false
	hasSystemModeration := false

	for _, cap := range capabilities {
		if cap == "moderate_content" {
			hasModerateContent = true
		}
		if cap == "system_moderation" {
			hasSystemModeration = true
		}
	}

	log.Info().
		Int("user_id", int(userCtx.UserID)).
		Str("active_pseudonym_id", userCtx.ActivePseudonymID).
		Strs("capabilities", capabilities).
		Bool("has_moderate_content", hasModerateContent).
		Bool("has_system_moderation", hasSystemModeration).
		Msg("Platform-wide capability check")

	if !hasModerateContent && !hasSystemModeration {
		log.Error().
			Int("user_id", int(userCtx.UserID)).
			Str("active_pseudonym_id", userCtx.ActivePseudonymID).
			Strs("capabilities", capabilities).
			Msg("User lacks platform-wide moderation permissions")
		return huma.Error403Forbidden("user does not have moderation permissions")
	}

	log.Info().
		Int("user_id", int(userCtx.UserID)).
		Str("active_pseudonym_id", userCtx.ActivePseudonymID).
		Msg("User has platform-wide moderation permissions")
	return nil
}

// validateContentExists validates that the content being moderated actually exists
func (h *ModerationHandler) validateContentExists(ctx context.Context, contentType string, contentID int) error {
	switch contentType {
	case "post":
		post, err := h.postDAO.GetPostByID(ctx, int64(contentID))
		if err != nil || post == nil {
			return fmt.Errorf("post not found")
		}
	case "comment":
		comment, err := h.commentDAO.GetCommentByID(ctx, int64(contentID))
		if err != nil || comment == nil {
			return fmt.Errorf("comment not found")
		}
	default:
		return fmt.Errorf("unsupported content type: %s", contentType)
	}
	return nil
}

// removeContentFromDatabase removes content from the database based on content type
func (h *ModerationHandler) removeContentFromDatabase(ctx context.Context, contentType string, contentID int, reason string) error {
	switch contentType {
	case "post":
		// Mark post as removed
		post, err := h.postDAO.GetPostByID(ctx, int64(contentID))
		if err != nil {
			return fmt.Errorf("failed to get post: %w", err)
		}
		if post == nil {
			return fmt.Errorf("post not found")
		}

		// Update post to mark as removed
		if err := h.postDAO.SetRemoved(ctx, int64(contentID), true); err != nil {
			return fmt.Errorf("failed to mark post as removed: %w", err)
		}

		log.Info().
			Str("content_type", contentType).
			Int("content_id", contentID).
			Str("reason", reason).
			Msg("Post marked as removed")

	case "comment":
		// Mark comment as removed
		comment, err := h.commentDAO.GetCommentByID(ctx, int64(contentID))
		if err != nil {
			return fmt.Errorf("failed to get comment: %w", err)
		}
		if comment == nil {
			return fmt.Errorf("comment not found")
		}

		// Update comment to mark as removed
		if err := h.commentDAO.SetRemoved(ctx, int64(contentID), true); err != nil {
			return fmt.Errorf("failed to mark comment as removed: %w", err)
		}

		log.Info().
			Str("content_type", contentType).
			Int("content_id", contentID).
			Str("reason", reason).
			Msg("Comment marked as removed")

	default:
		return fmt.Errorf("unsupported content type for removal: %s", contentType)
	}

	return nil
}

// sendModerationNotification sends a notification about the moderation action
func (h *ModerationHandler) sendModerationNotification(ctx context.Context, actionType, contentType string, contentID int, reason string) {
	// TODO: Implement actual notification sending
	// This could involve:
	// - Sending email notifications
	// - In-app notifications
	// - Webhook calls
	// - Audit logging

	log.Info().
		Str("action_type", actionType).
		Str("content_type", contentType).
		Int("content_id", contentID).
		Str("reason", reason).
		Msg("Moderation notification sent")
}

// parseActionDetails parses action details JSON and returns structured data
func (h *ModerationHandler) parseActionDetails(actionDetails sql.Null[types.JSON[json.RawMessage]]) map[string]interface{} {
	if !actionDetails.Valid {
		return nil
	}

	// Get the raw bytes from the JSON type
	value, err := actionDetails.V.Value()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get action details value")
		return nil
	}

	// Convert to bytes
	bytes, ok := value.([]byte)
	if !ok {
		log.Error().Msg("Action details value is not bytes")
		return nil
	}

	var details map[string]interface{}
	if err := json.Unmarshal(bytes, &details); err != nil {
		log.Error().Err(err).Msg("Failed to parse action details")
		return nil
	}

	return details
}

// loadRelatedData loads related data for reports and moderation actions
func (h *ModerationHandler) loadRelatedData(ctx context.Context, report *dbmodels.Report) (*apimodels.Report, error) {
	log.Info().
		Int("report_id", int(report.ReportID)).
		Str("reporter_pseudonym_id", report.ReporterPseudonymID).
		Msg("Loading related data for report")

	// Load reporter pseudonym
	reporterPseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, report.ReporterPseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", report.ReporterPseudonymID).Msg("Failed to load reporter pseudonym")
	} else {
		log.Info().
			Str("pseudonym_id", report.ReporterPseudonymID).
			Str("display_name", reporterPseudonym.DisplayName).
			Msg("Successfully loaded reporter pseudonym")
	}

	// Load reported pseudonym if available
	var reportedPseudonym *dbmodels.Pseudonym
	if report.ReportedPseudonymID.Valid {
		reportedPseudonym, err = h.securePseudonymDAO.GetPseudonymByID(ctx, report.ReportedPseudonymID.V)
		if err != nil {
			log.Error().Err(err).Str("pseudonym_id", report.ReportedPseudonymID.V).Msg("Failed to load reported pseudonym")
		}
	}

	// Convert to API model
	apiReport := apimodels.Report{
		ReportID:     int(report.ReportID),
		ContentType:  report.ContentType,
		ReportReason: report.ReportReason,
		Status:       report.Status.V,
		CreatedAt:    report.CreatedAt.V.Format(time.RFC3339),
	}

	// Set optional fields
	if report.ContentID.Valid {
		contentID := int(report.ContentID.V)
		apiReport.ContentID = &contentID
	}

	if report.ReportedPseudonymID.Valid {
		apiReport.ReportedPseudonymID = report.ReportedPseudonymID.V
	}

	if report.ReportDetails.Valid {
		apiReport.ReportDetails = report.ReportDetails.V
	}

	// Set reporter info
	if reporterPseudonym != nil {
		apiReport.Reporter = apimodels.Reporter{
			PseudonymID: reporterPseudonym.PseudonymID,
			DisplayName: reporterPseudonym.DisplayName,
		}
	}

	// Set reported user info
	if reportedPseudonym != nil {
		apiReport.ReportedUser = apimodels.ReportedUser{
			PseudonymID: reportedPseudonym.PseudonymID,
			DisplayName: reportedPseudonym.DisplayName,
		}
	}

	// Set resolution information if report is resolved or dismissed
	if report.Status.V == "resolved" || report.Status.V == "dismissed" {
		if report.ResolutionNotes.Valid {
			apiReport.ResolutionNotes = report.ResolutionNotes.V
		}

		if report.ResolvedAt.Valid {
			apiReport.ResolvedAt = report.ResolvedAt.V.Format(time.RFC3339)
		}

		if report.ResolvedByPseudonymID.Valid {
			// Load resolver pseudonym
			resolverPseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, report.ResolvedByPseudonymID.V)
			if err == nil && resolverPseudonym != nil {
				resolvedBy := apimodels.ResolvedBy{
					PseudonymID: resolverPseudonym.PseudonymID,
					DisplayName: resolverPseudonym.DisplayName,
				}
				apiReport.ResolvedBy = &resolvedBy
			}
		}
	}

	return &apiReport, nil
}

// ReportContent handles reporting content or users
func (h *ModerationHandler) ReportContent(ctx context.Context, input *struct {
	middleware.AuthInput
	apimodels.ReportInput
}) (*apimodels.ReportResponse, error) {
	// Extract user from input
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from input")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	log.Info().
		Str("endpoint", "reports").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Str("content_type", input.Body.ContentType).
		Str("report_reason", input.Body.ReportReason).
		Msg("Report content requested")

	// Validate content exists if content ID is provided
	if input.Body.ContentID != nil {
		if err := h.validateContentExists(ctx, input.Body.ContentType, *input.Body.ContentID); err != nil {
			log.Error().Err(err).Msg("Content validation failed")
			return nil, fmt.Errorf("content not found: %w", err)
		}
	}

	// Create report in database
	status := sql.Null[string]{V: "pending", Valid: true}
	reportSetter := &dbmodels.ReportSetter{
		ReporterPseudonymID: &userCtx.ActivePseudonymID,
		ContentType:         &input.Body.ContentType,
		ReportReason:        &input.Body.ReportReason,
		Status:              &status,
		CreatedAt:           &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	// Set content ID if provided
	if input.Body.ContentID != nil {
		contentID := int64(*input.Body.ContentID)
		reportSetter.ContentID = &sql.Null[int64]{V: contentID, Valid: true}
	}

	// Set reported pseudonym ID if provided
	if input.Body.ReportedPseudonymID != "" {
		reportSetter.ReportedPseudonymID = &sql.Null[string]{V: input.Body.ReportedPseudonymID, Valid: true}
	}

	// Set report details if provided
	if input.Body.ReportDetails != "" {
		reportSetter.ReportDetails = &sql.Null[string]{V: input.Body.ReportDetails, Valid: true}
	}

	report, err := h.reportDAO.CreateReport(ctx, reportSetter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create report")
		return nil, err
	}

	response := apimodels.NewReportResponse(int(report.ReportID))

	log.Info().
		Str("endpoint", "reports").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int("report_id", int(report.ReportID)).
		Msg("Report content completed")

	return response, nil
}

// GetSubforumReports handles getting reports for a specific subforum
func (h *ModerationHandler) GetSubforumReports(ctx context.Context, input *struct {
	middleware.AuthInput
	SubforumPath string `path:"subforum_path" example:"b/hashpost"`
	Status       string `query:"status" example:"pending"`
	Page         int    `query:"page" example:"1"`
	Limit        int    `query:"limit" example:"25"`
}) (*apimodels.ReportsListResponse, error) {
	// Extract moderator from input
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from input")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	// Parse subforum path
	_, subforumName, err := h.parseSubforumPath(input.SubforumPath)
	if err != nil {
		log.Error().Err(err).Str("subforum_path", input.SubforumPath).Msg("Failed to parse subforum path")
		return nil, huma.Error400BadRequest("Invalid subforum path")
	}

	// Get subforum
	log.Info().Str("subforum_name", subforumName).Msg("Calling GetSubforumByName")
	subforum, err := h.subforumDAO.GetSubforumByName(ctx, subforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", subforumName).Msg("Failed to get subforum")
		return nil, huma.Error404NotFound("Subforum not found")
	}
	if subforum == nil {
		log.Error().Str("subforum_name", subforumName).Msg("Subforum is nil")
		return nil, huma.Error404NotFound("Subforum not found")
	}
	log.Info().Str("subforum_name", subforumName).Str("found_name", subforum.Name).Msg("Successfully retrieved subforum")

	// Validate moderator permissions for this specific subforum
	if err := h.validateModeratorPermissionsForSubforum(ctx, userCtx, input.SubforumPath); err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Str("subforum_path", input.SubforumPath).Msg("User lacks moderation permissions for subforum")
		return nil, huma.Error403Forbidden("Insufficient permissions for this subforum")
	}

	log.Info().
		Str("endpoint", "moderation/subforum/reports").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Str("subforum_path", input.SubforumPath).
		Str("status", input.Status).
		Msg("Get subforum reports requested")

	// Get reports from database (same as global endpoint for now)
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	log.Info().
		Str("status", input.Status).
		Int("page", page).
		Int("limit", limit).
		Msg("Querying reports from database")

	reports, err := h.reportDAO.GetReports(ctx, input.Status, page, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get reports")
		return nil, err
	}

	log.Info().
		Int("reports_count", len(reports)).
		Msg("Retrieved reports from database")

	// Get total count
	total, err := h.reportDAO.CountReports(ctx, input.Status)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count reports")
		return nil, err
	}

	// Convert database models to API models with related data
	apiReports := make([]apimodels.Report, len(reports))
	for i, report := range reports {
		log.Info().
			Int("report_id", int(report.ReportID)).
			Str("content_type", report.ContentType).
			Str("report_reason", report.ReportReason).
			Str("status", report.Status.V).
			Str("reporter_pseudonym_id", report.ReporterPseudonymID).
			Msg("Processing report")

		apiReport, err := h.loadRelatedData(ctx, report)
		if err != nil {
			log.Error().Err(err).Int("report_id", int(report.ReportID)).Msg("Failed to load related data")
			// Continue with basic data if related data loading fails
			apiReports[i] = apimodels.Report{
				ReportID:     int(report.ReportID),
				ContentType:  report.ContentType,
				ReportReason: report.ReportReason,
				Status:       report.Status.V,
				CreatedAt:    report.CreatedAt.V.Format(time.RFC3339),
			}
		} else {
			apiReports[i] = *apiReport
		}
	}

	response := apimodels.NewReportsListResponse(apiReports, page, limit, int(total))

	log.Info().
		Str("endpoint", "moderation/subforum/reports").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Str("subforum_path", input.SubforumPath).
		Int("count", len(reports)).
		Int("total", int(total)).
		Msg("Get subforum reports completed")

	return response, nil
}

// RemoveContent handles removing content as a moderator
func (h *ModerationHandler) RemoveContent(ctx context.Context, input *apimodels.ContentRemovalInput) (*apimodels.ContentRemovalResponse, error) {
	// Extract moderator from context
	userCtx, err := h.extractUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	// Validate moderator permissions
	if err := h.validateModeratorPermissions(userCtx); err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("User lacks moderation permissions")
		return nil, fmt.Errorf("insufficient permissions: %w", err)
	}

	log.Info().
		Str("endpoint", "moderation/content/remove").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Str("content_type", input.ContentType).
		Int("content_id", input.ContentID).
		Str("removal_reason", input.Body.RemovalReason).
		Msg("Remove content requested")

	// Validate content exists
	if err := h.validateContentExists(ctx, input.ContentType, input.ContentID); err != nil {
		log.Error().Err(err).Msg("Content validation failed")
		return nil, fmt.Errorf("content not found: %w", err)
	}

	// Create moderation action record
	actionDetails := map[string]interface{}{
		"removal_reason": input.Body.RemovalReason,
		"content_type":   input.ContentType,
		"content_id":     input.ContentID,
	}

	actionDetailsJSON, err := json.Marshal(actionDetails)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal action details")
		return nil, fmt.Errorf("failed to create action details: %w", err)
	}

	actionSetter := &dbmodels.ModerationActionSetter{
		ModeratorUserID:      &userCtx.UserID,
		ModeratorPseudonymID: &userCtx.ActivePseudonymID,
		ActionType:           &[]string{"remove_content"}[0],
		TargetContentType:    &sql.Null[string]{V: input.ContentType, Valid: true},
		TargetContentID:      &sql.Null[int64]{V: int64(input.ContentID), Valid: true},
		ActionDetails:        &sql.Null[types.JSON[json.RawMessage]]{V: types.NewJSON[json.RawMessage](actionDetailsJSON), Valid: true},
		CreatedAt:            &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	_, err = h.moderationActionDAO.CreateModerationAction(ctx, actionSetter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create moderation action")
		return nil, err
	}

	// Actually remove content from database
	if err := h.removeContentFromDatabase(ctx, input.ContentType, input.ContentID, input.Body.RemovalReason); err != nil {
		log.Error().Err(err).Msg("Failed to remove content from database")
		return nil, fmt.Errorf("failed to remove content: %w", err)
	}

	// Send notification
	h.sendModerationNotification(ctx, "remove_content", input.ContentType, input.ContentID, input.Body.RemovalReason)

	response := apimodels.NewContentRemovalResponse(input.ContentID, input.ContentType, input.Body.RemovalReason, userCtx.ActivePseudonymID, userCtx.DisplayName)

	log.Info().
		Str("endpoint", "moderation/content/remove").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Int("content_id", input.ContentID).
		Msg("Remove content completed")

	return response, nil
}

// BanUser handles banning a user from a subforum
func (h *ModerationHandler) BanUser(ctx context.Context, input *apimodels.UserBanInput) (*apimodels.UserBanResponse, error) {
	// Extract moderator from context
	userCtx, err := h.extractUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	// Validate moderator permissions
	if err := h.validateModeratorPermissions(userCtx); err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("User lacks moderation permissions")
		return nil, fmt.Errorf("insufficient permissions: %w", err)
	}

	log.Info().
		Str("endpoint", "moderation/users/ban").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Str("pseudonym_id", input.PseudonymID).
		Int("subforum_id", input.Body.SubforumID).
		Str("ban_reason", input.Body.BanReason).
		Bool("is_permanent", input.Body.IsPermanent).
		Msg("Ban user requested")

	// Get user ID from pseudonym
	userID, err := h.securePseudonymDAO.GetUserIDByPseudonym(ctx, input.PseudonymID, "user", "site")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get user ID from pseudonym")
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Check if user is already banned from this subforum
	isBanned, err := h.userBanDAO.IsUserBannedFromSubforum(ctx, userID, int32(input.Body.SubforumID))
	if err != nil {
		log.Error().Err(err).Msg("Failed to check if user is already banned")
		return nil, fmt.Errorf("failed to check ban status: %w", err)
	}
	if isBanned {
		return nil, fmt.Errorf("user is already banned from this subforum")
	}

	// Generate user fingerprint using IBE (mock for now)
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Error().Err(err).Msg("Failed to generate user fingerprint")
		return nil, fmt.Errorf("failed to generate user fingerprint: %w", err)
	}
	bannedFingerprint := hex.EncodeToString(bytes)

	// Create ban record
	subforumID := int32(input.Body.SubforumID)
	banSetter := &dbmodels.UserBanSetter{
		SubforumID:          &subforumID,
		BannedUserID:        &userID,
		BannedByUserID:      &userCtx.UserID,
		BannedByPseudonymID: &userCtx.ActivePseudonymID,
		BanReason:           &input.Body.BanReason,
		IsPermanent:         &sql.Null[bool]{V: input.Body.IsPermanent, Valid: true},
		CreatedAt:           &sql.Null[time.Time]{V: time.Now(), Valid: true},
		IsActive:            &sql.Null[bool]{V: true, Valid: true},
	}

	// Set expiration if not permanent
	if !input.Body.IsPermanent && input.Body.DurationDays != nil && *input.Body.DurationDays > 0 {
		expiresAt := time.Now().AddDate(0, 0, *input.Body.DurationDays)
		banSetter.ExpiresAt = &sql.Null[time.Time]{V: expiresAt, Valid: true}
	}

	ban, err := h.userBanDAO.CreateUserBan(ctx, banSetter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create user ban")
		return nil, err
	}

	// Send notification
	h.sendModerationNotification(ctx, "ban_user", "user", int(userID), input.Body.BanReason)

	response := apimodels.NewUserBanResponse(int(ban.BanID), bannedFingerprint, input.Body.SubforumID, input.Body.BanReason, input.Body.IsPermanent, input.Body.DurationDays, userCtx.ActivePseudonymID, userCtx.DisplayName)

	log.Info().
		Str("endpoint", "moderation/users/ban").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Int("ban_id", int(ban.BanID)).
		Str("pseudonym_id", input.PseudonymID).
		Msg("Ban user completed")

	return response, nil
}

// ResolveReport handles resolving or dismissing a report
func (h *ModerationHandler) ResolveReport(ctx context.Context, input *struct {
	middleware.AuthInput
	ReportID int `path:"report_id" example:"123"`
	apimodels.ReportResolutionInput
}) (*apimodels.ReportResolutionResponse, error) {
	// Extract moderator from input
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from input")
		return nil, huma.Error401Unauthorized("Authentication required")
	}

	// Validate moderator permissions
	if err := h.validateModeratorPermissions(userCtx); err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("User lacks moderation permissions")
		return nil, huma.Error403Forbidden("Insufficient permissions")
	}

	log.Info().
		Str("endpoint", "moderation/reports/resolve").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Int("report_id", input.ReportID).
		Str("action", input.Body.Action).
		Msg("Resolve report requested")

	// Get the report to validate it exists and is pending
	report, err := h.reportDAO.GetReportByID(ctx, int64(input.ReportID))
	if err != nil {
		log.Error().Err(err).Int("report_id", input.ReportID).Msg("Failed to get report")
		return nil, huma.Error404NotFound("Report not found")
	}
	if report == nil {
		return nil, huma.Error404NotFound("Report not found")
	}

	// Check if report is still pending
	if report.Status.V != "pending" {
		return nil, huma.Error400BadRequest("Report is not pending")
	}

	// Resolve the report
	err = h.reportDAO.ResolveReport(ctx, int64(input.ReportID), userCtx.UserID, userCtx.ActivePseudonymID, input.Body.Notes, input.Body.Action)
	if err != nil {
		log.Error().Err(err).Int("report_id", input.ReportID).Msg("Failed to resolve report")
		return nil, huma.Error500InternalServerError("Failed to resolve report")
	}

	// Create moderation action record
	actionDetails := map[string]interface{}{
		"resolution_action": input.Body.Action,
		"resolution_notes":  input.Body.Notes,
		"report_id":         input.ReportID,
	}

	actionDetailsJSON, err := json.Marshal(actionDetails)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal action details")
		return nil, huma.Error500InternalServerError("Failed to create action details")
	}

	actionSetter := &dbmodels.ModerationActionSetter{
		ModeratorUserID:      &userCtx.UserID,
		ModeratorPseudonymID: &userCtx.ActivePseudonymID,
		ActionType:           &[]string{"resolve_report"}[0],
		ActionDetails:        &sql.Null[types.JSON[json.RawMessage]]{V: types.NewJSON[json.RawMessage](actionDetailsJSON), Valid: true},
		CreatedAt:            &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	_, err = h.moderationActionDAO.CreateModerationAction(ctx, actionSetter)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create moderation action")
		// Don't fail the request if action logging fails
	}

	response := apimodels.NewReportResolutionResponse(input.ReportID, input.Body.Action, input.Body.Notes, userCtx.ActivePseudonymID, userCtx.DisplayName)

	log.Info().
		Str("endpoint", "moderation/reports/resolve").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Int("report_id", input.ReportID).
		Str("action", input.Body.Action).
		Msg("Resolve report completed")

	return response, nil
}

// GetModerationHistory handles getting moderation action history
func (h *ModerationHandler) GetModerationHistory(ctx context.Context, input *apimodels.ModerationHistoryInput) (*apimodels.ModerationHistoryResponse, error) {
	// Extract moderator from context
	userCtx, err := h.extractUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	// Validate moderator permissions
	if err := h.validateModeratorPermissions(userCtx); err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("User lacks moderation permissions")
		return nil, fmt.Errorf("insufficient permissions: %w", err)
	}

	log.Info().
		Str("endpoint", "moderation/history").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Str("action_type", input.ActionType).
		Msg("Get moderation history requested")

	// Get moderation actions from database
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	actions, err := h.moderationActionDAO.GetModerationActions(ctx, input.ActionType, page, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get moderation actions")
		return nil, err
	}

	// Get total count
	total, err := h.moderationActionDAO.CountModerationActions(ctx, input.ActionType)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count moderation actions")
		return nil, err
	}

	// Convert database models to API models
	apiActions := make([]apimodels.ModerationAction, len(actions))
	for i, action := range actions {
		apiActions[i] = apimodels.ModerationAction{
			ActionID:          int(action.ActionID),
			ActionType:        action.ActionType,
			TargetContentType: action.TargetContentType.V,
			TargetContentID:   int(action.TargetContentID.V),
			CreatedAt:         action.CreatedAt.V.Format(time.RFC3339),
		}

		// Parse action details if available
		if action.ActionDetails.Valid {
			details := h.parseActionDetails(action.ActionDetails)
			if details != nil {
				// Set specific fields based on action type
				if removalReason, ok := details["removal_reason"].(string); ok {
					apiActions[i].ActionDetails.RemovalReason = removalReason
				}
			}
		}

		// Set moderator info
		apiActions[i].Moderator = apimodels.Moderator{
			PseudonymID: action.ModeratorPseudonymID,
			DisplayName: userCtx.DisplayName, // Use current user's display name as fallback
		}

		// Load moderator pseudonym details if available
		moderatorPseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, action.ModeratorPseudonymID)
		if err == nil && moderatorPseudonym != nil {
			apiActions[i].Moderator.DisplayName = moderatorPseudonym.DisplayName
		}

		// Set subforum info if available
		if action.SubforumID.Valid {
			apiActions[i].Subforum = apimodels.SubforumModerator{
				PseudonymID:   action.ModeratorPseudonymID,
				DisplayName:   apiActions[i].Moderator.DisplayName,
				ModeratorType: "moderator",
			}
		}
	}

	response := apimodels.NewModerationHistoryResponse(apiActions, page, limit, int(total))

	log.Info().
		Str("endpoint", "moderation/history").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Int("count", len(actions)).
		Int("total", int(total)).
		Msg("Get moderation history completed")

	return response, nil
}

// ModerationStats represents statistics for the moderation dashboard
type ModerationStats struct {
	PendingReports int `json:"pending_reports"`
	BannedUsers    int `json:"banned_users"`
	ModActions     int `json:"mod_actions"`
	TotalPosts     int `json:"total_posts"`
	TotalComments  int `json:"total_comments"`
	TotalVotes     int `json:"total_votes"`
	AvgEngagement  int `json:"avg_engagement"`
}

// EngagementData represents engagement analytics data
type EngagementData struct {
	TimeRange  string                `json:"time_range"`
	DataPoints []EngagementDataPoint `json:"data_points"`
}

// EngagementDataPoint represents a single data point in engagement analytics
type EngagementDataPoint struct {
	Date             string `json:"date"`
	Posts            int    `json:"posts"`
	Comments         int    `json:"comments"`
	PostVotes        int    `json:"post_votes"`
	CommentVotes     int    `json:"comment_votes"`
	TotalVotes       int    `json:"total_votes"`
	PostUpvotes      int    `json:"post_upvotes"`
	PostDownvotes    int    `json:"post_downvotes"`
	CommentUpvotes   int    `json:"comment_upvotes"`
	CommentDownvotes int    `json:"comment_downvotes"`
}

// ModerationInput represents the input for moderation endpoints
type ModerationInput struct {
	middleware.AuthInput
	SubforumPath string `path:"subforum_path" example:"b/hashpost"`
	TimeRange    string `query:"time_range" example:"14d" enum:"7d,14d,30d"`
}

// ModerationStatsOutput represents the output for moderation stats
type ModerationStatsOutput struct {
	Body ModerationStats `json:"body"`
}

// EngagementOutput represents the output for engagement analytics
type EngagementOutput struct {
	Body EngagementData `json:"body"`
}

// CacheEntry represents a cached value with expiration
type CacheEntry struct {
	Value      interface{}
	Expiration time.Time
}

// ModerationCache is a thread-safe LRU cache for moderation data
type ModerationCache struct {
	cache   map[string]CacheEntry
	mutex   sync.RWMutex
	maxSize int
}

// NewModerationCache creates a new moderation cache
func NewModerationCache(maxSize int) *ModerationCache {
	return &ModerationCache{
		cache:   make(map[string]CacheEntry),
		maxSize: maxSize,
	}
}

// Get retrieves a value from the cache
func (c *ModerationCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.Expiration) {
		// Expired, remove it
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.cache, key)
		c.mutex.Unlock()
		c.mutex.RLock()
		return nil, false
	}

	return entry.Value, true
}

// Set stores a value in the cache
func (c *ModerationCache) Set(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Remove expired entries first
	for k, entry := range c.cache {
		if time.Now().After(entry.Expiration) {
			delete(c.cache, k)
		}
	}

	// If cache is full, remove oldest entry (simple implementation)
	if len(c.cache) >= c.maxSize {
		var oldestKey string
		var oldestTime time.Time
		for k, entry := range c.cache {
			if oldestKey == "" || entry.Expiration.Before(oldestTime) {
				oldestKey = k
				oldestTime = entry.Expiration
			}
		}
		if oldestKey != "" {
			delete(c.cache, oldestKey)
		}
	}

	c.cache[key] = CacheEntry{
		Value:      value,
		Expiration: time.Now().Add(ttl),
	}
}

// Clear clears the entire cache
func (c *ModerationCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.cache = make(map[string]CacheEntry)
}

// Global cache instance
var moderationCache = NewModerationCache(100)

// GetModerationStats returns statistics for the moderation dashboard
func (h *ModerationHandler) GetModerationStats(ctx context.Context, input *ModerationInput) (*ModerationStatsOutput, error) {
	// Handle authentication - try to get user context from middleware
	var userCtx *middleware.UserContext
	var err error

	// First, try to get user context from middleware (header-based auth)
	userCtx, err = middleware.ExtractUserFromContext(ctx)
	if err != nil {
		// If no user context from middleware, try cookie-based auth from input
		userCtx, err = middleware.ExtractUserFromHumaInput(&input.AuthInput)
		if err != nil {
			return nil, huma.Error403Forbidden("Authentication required")
		}
	}

	// Check if user has moderation permissions for this specific subforum
	err = h.validateModeratorPermissionsForSubforum(ctx, userCtx, input.SubforumPath)
	if err != nil {
		return nil, huma.Error403Forbidden("Moderation permissions required")
	}

	// Try to get from cache first
	cacheKey := "mod_stats:" + input.SubforumPath
	if cached, exists := moderationCache.Get(cacheKey); exists {
		if stats, ok := cached.(ModerationStats); ok {
			return &ModerationStatsOutput{Body: stats}, nil
		}
	}

	log.Info().
		Msg("No cached stats found, calculating fresh stats")

	// Parse time range parameter and calculate since date
	var since time.Time
	var days int
	switch input.TimeRange {
	case "7d":
		since = time.Now().AddDate(0, 0, -7)
		days = 7
	case "14d":
		since = time.Now().AddDate(0, 0, -14)
		days = 14
	case "30d":
		since = time.Now().AddDate(0, 0, -30)
		days = 30
	default:
		// Default to 30 days if invalid time range
		since = time.Now().AddDate(0, 0, -30)
		days = 30
	}

	// Get moderation statistics
	pendingReports, err := h.reportDAO.GetPendingReportsCount(ctx, input.SubforumPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending reports count: %w", err)
	}

	bannedUsers, err := h.userBanDAO.GetBannedUsersCount(ctx, input.SubforumPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get banned users count: %w", err)
	}

	modActions, err := h.moderationActionDAO.GetModActionsCount(ctx, input.SubforumPath, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get moderation actions count: %w", err)
	}

	totalPosts, err := h.postDAO.GetPostsCount(ctx, input.SubforumPath, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get posts count: %w", err)
	}

	totalComments, err := h.commentDAO.GetCommentsCount(ctx, input.SubforumPath, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments count: %w", err)
	}

	totalVotes, err := h.voteDAO.GetVotesCount(ctx, input.SubforumPath, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get votes count: %w", err)
	}

	// Calculate average engagement (posts + comments + votes) over the specified time range
	avgEngagement := 0
	if totalPosts+totalComments+totalVotes > 0 {
		avgEngagement = (totalPosts + totalComments + totalVotes) / days // Average per day over the time range
	}

	stats := ModerationStats{
		PendingReports: pendingReports,
		BannedUsers:    bannedUsers,
		ModActions:     modActions,
		TotalPosts:     totalPosts,
		TotalComments:  totalComments,
		TotalVotes:     totalVotes,
		AvgEngagement:  avgEngagement,
	}

	// Cache the result for 5 minutes
	moderationCache.Set(cacheKey, stats, 5*time.Minute)

	return &ModerationStatsOutput{Body: stats}, nil
}

// GetEngagementAnalytics returns engagement analytics data
func (h *ModerationHandler) GetEngagementAnalytics(ctx context.Context, input *ModerationInput) (*EngagementOutput, error) {
	// Handle authentication - try to get user context from middleware
	var userCtx *middleware.UserContext
	var err error

	// First, try to get user context from middleware (header-based auth)
	userCtx, err = middleware.ExtractUserFromContext(ctx)
	if err != nil {
		// If no user context from middleware, try cookie-based auth from input
		userCtx, err = middleware.ExtractUserFromHumaInput(&input.AuthInput)
		if err != nil {
			return nil, huma.Error403Forbidden("Authentication required")
		}
	}

	// Check if user has moderation permissions for this specific subforum
	err = h.validateModeratorPermissionsForSubforum(ctx, userCtx, input.SubforumPath)
	if err != nil {
		return nil, huma.Error403Forbidden("Moderation permissions required")
	}

	// Try to get from cache first
	cacheKey := "engagement:" + input.SubforumPath + ":" + input.TimeRange
	if cached, exists := moderationCache.Get(cacheKey); exists {
		if data, ok := cached.(EngagementData); ok {
			return &EngagementOutput{Body: data}, nil
		}
	}

	// Calculate time range
	days := 7
	switch input.TimeRange {
	case "14d":
		days = 14
	case "30d":
		days = 30
	}

	// Generate data points for each day
	dataPoints := make([]EngagementDataPoint, days)
	today := time.Now()

	for i := 0; i < days; i++ {
		date := today.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

		// Get data for this specific day (not cumulative)
		posts, _ := h.postDAO.GetPostsCountForDateRange(ctx, input.SubforumPath, startOfDay, endOfDay)
		comments, _ := h.commentDAO.GetCommentsCountForDateRange(ctx, input.SubforumPath, startOfDay, endOfDay)
		postVotes, _ := h.voteDAO.GetPostVotesCountForDateRange(ctx, input.SubforumPath, startOfDay, endOfDay)
		commentVotes, _ := h.voteDAO.GetCommentVotesCountForDateRange(ctx, input.SubforumPath, startOfDay, endOfDay)

		// Get upvote/downvote breakdown for this specific day
		postUpvotes, _ := h.voteDAO.GetPostUpvotesCountForDateRange(ctx, input.SubforumPath, startOfDay, endOfDay)
		postDownvotes, _ := h.voteDAO.GetPostDownvotesCountForDateRange(ctx, input.SubforumPath, startOfDay, endOfDay)
		commentUpvotes, _ := h.voteDAO.GetCommentUpvotesCountForDateRange(ctx, input.SubforumPath, startOfDay, endOfDay)
		commentDownvotes, _ := h.voteDAO.GetCommentDownvotesCountForDateRange(ctx, input.SubforumPath, startOfDay, endOfDay)

		dataPoints[days-1-i] = EngagementDataPoint{
			Date:             startOfDay.Format("2006-01-02"),
			Posts:            posts,
			Comments:         comments,
			PostVotes:        postVotes,
			CommentVotes:     commentVotes,
			TotalVotes:       postVotes + commentVotes,
			PostUpvotes:      postUpvotes,
			PostDownvotes:    postDownvotes,
			CommentUpvotes:   commentUpvotes,
			CommentDownvotes: commentDownvotes,
		}
	}

	data := EngagementData{
		TimeRange:  input.TimeRange,
		DataPoints: dataPoints,
	}

	// Cache the result for 10 minutes (engagement data is more expensive)
	moderationCache.Set(cacheKey, data, 10*time.Minute)

	return &EngagementOutput{Body: data}, nil
}
