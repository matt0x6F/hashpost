package handlers

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/database/dao/mocks"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/matt0x6f/hashpost/internal/fixtures"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob/types"
)

// ModerationHandler handles moderation-related requests
type ModerationHandler struct {
	reportDAO           dao.ReportDAOInterface
	moderationActionDAO dao.ModerationActionDAOInterface
	userBanDAO          dao.UserBanDAOInterface
	securePseudonymDAO  dao.SecurePseudonymDAOInterface
	subforumDAO         dao.SubforumDAOInterface
	postDAO             dao.PostDAOInterface
	commentDAO          dao.CommentDAOInterface
}

// NewModerationHandler creates a new moderation handler with interface dependencies
func NewModerationHandler(
	reportDAO dao.ReportDAOInterface,
	moderationActionDAO dao.ModerationActionDAOInterface,
	userBanDAO dao.UserBanDAOInterface,
	securePseudonymDAO dao.SecurePseudonymDAOInterface,
	subforumDAO dao.SubforumDAOInterface,
	postDAO dao.PostDAOInterface,
	commentDAO dao.CommentDAOInterface,
) *ModerationHandler {
	return &ModerationHandler{
		reportDAO:           reportDAO,
		moderationActionDAO: moderationActionDAO,
		userBanDAO:          userBanDAO,
		securePseudonymDAO:  securePseudonymDAO,
		subforumDAO:         subforumDAO,
		postDAO:             postDAO,
		commentDAO:          commentDAO,
	}
}

// NewModerationHandlerWithMocks creates a new moderation handler with mock DAOs and fixture data
func NewModerationHandlerWithMocks() *ModerationHandler {
	// Create mock DAOs
	mockReportDAO := mocks.NewMockReportDAO()
	mockModerationActionDAO := mocks.NewMockModerationActionDAO()
	mockUserBanDAO := mocks.NewMockUserBanDAO()
	mockSecurePseudonymDAO := mocks.NewMockSecurePseudonymDAO()
	mockSubforumDAO := mocks.NewMockSubforumDAO()
	mockPostDAO := mocks.NewMockPostDAO()
	mockCommentDAO := mocks.NewMockCommentDAO()

	// Inject fixture data into mocks
	mockReportDAO.InjectReport(fixtures.CreateTestReport())
	mockReportDAO.InjectReport(fixtures.CreateTestReportWithResolution())
	mockReportDAO.InjectReportsByStatus("pending", []*dbmodels.Report{fixtures.CreateTestReport()})
	mockReportDAO.InjectReportsByStatus("resolved", []*dbmodels.Report{fixtures.CreateTestReportWithResolution()})
	mockReportDAO.InjectCount("pending", 1)
	mockReportDAO.InjectCount("resolved", 1)
	mockReportDAO.SetDefaultBehavior()

	mockModerationActionDAO.InjectAction(fixtures.CreateTestModerationAction())
	mockModerationActionDAO.InjectAction(fixtures.CreateTestModerationActionWithDetails())
	mockModerationActionDAO.InjectActionsByType("remove_post", []*dbmodels.ModerationAction{fixtures.CreateTestModerationAction()})
	mockModerationActionDAO.InjectCount("remove_post", 1)
	mockModerationActionDAO.SetDefaultBehavior()

	mockUserBanDAO.InjectBan(fixtures.CreateTestUserBan())
	mockUserBanDAO.InjectBan(fixtures.CreateTestPermanentUserBan())
	mockUserBanDAO.InjectBan(fixtures.CreateTestInactiveUserBan())
	mockUserBanDAO.InjectBansBySubforum(1, []*dbmodels.UserBan{fixtures.CreateTestUserBan()})
	mockUserBanDAO.InjectCount("subforum_1", 1)
	mockUserBanDAO.SetDefaultBehavior()

	// Set up mock secure pseudonym DAO with fixture data
	mockSecurePseudonymDAO.InjectPseudonym(fixtures.CreateTestPseudonym())
	mockSecurePseudonymDAO.SetDefaultBehavior()

	// Set up mock post DAO with fixture data
	mockPostDAO.InjectPost(fixtures.CreateTestPost())
	mockPostDAO.SetDefaultBehavior()

	// Set up mock comment DAO with fixture data
	mockCommentDAO.InjectComment(fixtures.CreateTestComment())
	mockCommentDAO.SetDefaultBehavior()

	return &ModerationHandler{
		reportDAO:           mockReportDAO,
		moderationActionDAO: mockModerationActionDAO,
		userBanDAO:          mockUserBanDAO,
		securePseudonymDAO:  mockSecurePseudonymDAO,
		subforumDAO:         mockSubforumDAO,
		postDAO:             mockPostDAO,
		commentDAO:          mockCommentDAO,
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

// validateModeratorPermissions validates that the user has moderator permissions
func (h *ModerationHandler) validateModeratorPermissions(userCtx *middleware.UserContext) error {
	if !userCtx.HasCapability("moderate_content") {
		return fmt.Errorf("user does not have moderation permissions")
	}
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
	// Load reporter pseudonym
	reporterPseudonym, err := h.securePseudonymDAO.GetPseudonymByID(ctx, report.ReporterPseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", report.ReporterPseudonymID).Msg("Failed to load reporter pseudonym")
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

	return &apiReport, nil
}

// ReportContent handles reporting content or users
func (h *ModerationHandler) ReportContent(ctx context.Context, input *apimodels.ReportInput) (*apimodels.ReportResponse, error) {
	// Extract user from context
	userCtx, err := h.extractUserFromContext(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
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

// GetReports handles getting reports for moderation review
func (h *ModerationHandler) GetReports(ctx context.Context, input *apimodels.ReportsListInput) (*apimodels.ReportsListResponse, error) {
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
		Str("endpoint", "moderation/reports").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Str("status", input.Status).
		Msg("Get reports requested")

	// Get reports from database
	page := input.Page
	if page <= 0 {
		page = 1
	}
	limit := input.Limit
	if limit <= 0 {
		limit = 25
	}

	reports, err := h.reportDAO.GetReports(ctx, input.Status, page, limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get reports")
		return nil, err
	}

	// Get total count
	total, err := h.reportDAO.CountReports(ctx, input.Status)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count reports")
		return nil, err
	}

	// Convert database models to API models with related data
	apiReports := make([]apimodels.Report, len(reports))
	for i, report := range reports {
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
		Str("endpoint", "moderation/reports").
		Str("component", "handler").
		Int("moderator_id", int(userCtx.UserID)).
		Int("count", len(reports)).
		Int("total", int(total)).
		Msg("Get reports completed")

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
