package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	dbmodels "github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/types"
)

// RulesHandler handles rule-related requests
type RulesHandler struct {
	reportDAO         dao.ReportDAOInterface
	subforumDAO       dao.SubforumDAOInterface
	systemSettingsDAO dao.SystemSettingsDAOInterface
	permissionDAO     dao.PermissionDAOInterface
	db                bob.Executor
}

// NewRulesHandler creates a new rules handler
func NewRulesHandler(
	reportDAO dao.ReportDAOInterface,
	subforumDAO dao.SubforumDAOInterface,
	systemSettingsDAO dao.SystemSettingsDAOInterface,
	permissionDAO dao.PermissionDAOInterface,
	db bob.Executor,
) *RulesHandler {
	return &RulesHandler{
		reportDAO:         reportDAO,
		subforumDAO:       subforumDAO,
		systemSettingsDAO: systemSettingsDAO,
		permissionDAO:     permissionDAO,
		db:                db,
	}
}

// GetPlatformRules returns all platform rules
func (h *RulesHandler) GetPlatformRules(ctx context.Context, input *apimodels.PlatformRulesInput) (*apimodels.PlatformRulesResponse, error) {
	// Get platform rules from system settings
	setting, err := h.systemSettingsDAO.GetSetting(ctx, "platform_rules")
	if err != nil {
		log.Error().Err(err).Msg("Failed to get platform rules setting")
		return nil, fmt.Errorf("failed to get platform rules: %w", err)
	}

	if setting == nil {
		// Return empty rules if no platform rules are configured
		return &apimodels.PlatformRulesResponse{
			Status: 200,
			Body: struct {
				Rules []apimodels.Rule `json:"rules"`
			}{
				Rules: []apimodels.Rule{},
			},
		}, nil
	}

	// Parse JSON rules
	var rules []apimodels.Rule
	if err := json.Unmarshal([]byte(setting.SettingValue), &rules); err != nil {
		log.Error().Err(err).Msg("Failed to parse platform rules JSON")
		return nil, fmt.Errorf("failed to parse platform rules: %w", err)
	}

	// Filter by active status if requested
	if input.ActiveOnly {
		var activeRules []apimodels.Rule
		for _, rule := range rules {
			if rule.Active {
				activeRules = append(activeRules, rule)
			}
		}
		rules = activeRules
	}

	return &apimodels.PlatformRulesResponse{
		Status: 200,
		Body: struct {
			Rules []apimodels.Rule `json:"rules"`
		}{
			Rules: rules,
		},
	}, nil
}

// GetSubforumRules returns rules for a specific subforum
func (h *RulesHandler) GetSubforumRules(ctx context.Context, input *apimodels.SubforumRulesInput) (*apimodels.SubforumRulesResponse, error) {
	// Get subforum by name and community type
	subforum, err := h.subforumDAO.GetSubforumByName(ctx, input.SubforumName)
	if err != nil {
		log.Error().Err(err).Str("subforum_name", input.SubforumName).Msg("Failed to get subforum")
		return nil, fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Parse subforum rules JSON
	var rules []apimodels.Rule
	if subforum.SubforumRules.Valid {
		// Get the raw bytes from the JSON type
		value, err := subforum.SubforumRules.V.Value()
		if err != nil {
			log.Error().Err(err).Str("subforum_name", input.SubforumName).Msg("Failed to get subforum rules value")
			return nil, fmt.Errorf("failed to parse subforum rules: %w", err)
		}

		// Convert to bytes
		bytes, ok := value.([]byte)
		if !ok {
			log.Error().Str("subforum_name", input.SubforumName).Msg("Subforum rules value is not bytes")
			return nil, fmt.Errorf("failed to parse subforum rules: %w", err)
		}

		if err := json.Unmarshal(bytes, &rules); err != nil {
			log.Error().Err(err).Str("subforum_name", input.SubforumName).Msg("Failed to parse subforum rules JSON")
			return nil, fmt.Errorf("failed to parse subforum rules: %w", err)
		}
	}

	// Filter by active status if requested
	if input.ActiveOnly {
		var activeRules []apimodels.Rule
		for _, rule := range rules {
			if rule.Active {
				activeRules = append(activeRules, rule)
			}
		}
		rules = activeRules
	}

	return &apimodels.SubforumRulesResponse{
		Status: 200,
		Body: struct {
			SubforumID int32            `json:"subforum_id" example:"1"`
			Name       string           `json:"name" example:"golang"`
			Rules      []apimodels.Rule `json:"rules"`
		}{
			SubforumID: subforum.SubforumID,
			Name:       subforum.Name,
			Rules:      rules,
		},
	}, nil
}

// CreateSubforumRule creates a new rule for a subforum
func (h *RulesHandler) CreateSubforumRule(ctx context.Context, input *apimodels.RuleCreateInput) (*apimodels.Rule, error) {
	// Extract user from context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	// Validate moderator permissions for the subforum
	if err := h.validateModeratorPermissionsForSubforum(ctx, userCtx, input.CommunityType, input.SubforumName); err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("User lacks moderation permissions")
		return nil, fmt.Errorf("insufficient permissions: %w", err)
	}

	// Get subforum
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.CommunityType, input.SubforumName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subforum")
		return nil, fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Get existing rules
	var rules []apimodels.Rule
	if subforum.SubforumRules.Valid {
		// Get the raw bytes from the JSON type
		value, err := subforum.SubforumRules.V.Value()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get existing subforum rules value")
			return nil, fmt.Errorf("failed to parse existing rules: %w", err)
		}

		// Convert to bytes
		bytes, ok := value.([]byte)
		if !ok {
			log.Error().Msg("Existing subforum rules value is not bytes")
			return nil, fmt.Errorf("failed to parse existing rules: %w", err)
		}

		if err := json.Unmarshal(bytes, &rules); err != nil {
			log.Error().Err(err).Msg("Failed to parse existing subforum rules")
			return nil, fmt.Errorf("failed to parse existing rules: %w", err)
		}
	}

	// Check if rule code already exists
	for _, rule := range rules {
		if rule.Code == input.Body.Code {
			return nil, huma.Error400BadRequest("rule with this code already exists")
		}
	}

	// Create new rule
	newRule := apimodels.Rule{
		Code:        input.Body.Code,
		Name:        input.Body.Name,
		Description: input.Body.Description,
		Category:    input.Body.Category,
		Severity:    input.Body.Severity,
		Active:      input.Body.Active,
	}

	// Add to rules array
	rules = append(rules, newRule)

	// Convert back to JSON
	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal rules to JSON")
		return nil, fmt.Errorf("failed to save rules: %w", err)
	}

	// Update subforum with new rules
	updateSetter := &dbmodels.SubforumSetter{
		UpdatedAt: &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	// Set the JSON field properly
	jsonValue := types.JSON[json.RawMessage]{}
	if err := jsonValue.Scan(rulesJSON); err != nil {
		log.Error().Err(err).Msg("Failed to scan rules JSON")
		return nil, fmt.Errorf("failed to save rules: %w", err)
	}
	updateSetter.SubforumRules = &sql.Null[types.JSON[json.RawMessage]]{V: jsonValue, Valid: true}

	// Use the Update method on the subforum model
	if err := subforum.Update(ctx, h.db, updateSetter); err != nil {
		log.Error().Err(err).Msg("Failed to update subforum rules")
		return nil, fmt.Errorf("failed to save rule: %w", err)
	}

	log.Info().
		Str("endpoint", "subforum/rules").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int32("subforum_id", subforum.SubforumID).
		Str("rule_code", newRule.Code).
		Msg("Subforum rule created")

	return &newRule, nil
}

// UpdateSubforumRule updates an existing rule for a subforum
func (h *RulesHandler) UpdateSubforumRule(ctx context.Context, input *apimodels.RuleUpdateInput) (*apimodels.Rule, error) {
	// Extract user from context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	// Validate moderator permissions for the subforum
	if err := h.validateModeratorPermissionsForSubforum(ctx, userCtx, input.CommunityType, input.SubforumName); err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("User lacks moderation permissions")
		return nil, fmt.Errorf("insufficient permissions: %w", err)
	}

	// Get subforum
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.CommunityType, input.SubforumName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subforum")
		return nil, fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Get existing rules
	var rules []apimodels.Rule
	if subforum.SubforumRules.Valid {
		// Get the raw bytes from the JSON type
		value, err := subforum.SubforumRules.V.Value()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get existing subforum rules value")
			return nil, fmt.Errorf("failed to parse existing rules: %w", err)
		}

		// Convert to bytes
		bytes, ok := value.([]byte)
		if !ok {
			log.Error().Msg("Existing subforum rules value is not bytes")
			return nil, fmt.Errorf("failed to parse existing rules: %w", err)
		}

		if err := json.Unmarshal(bytes, &rules); err != nil {
			log.Error().Err(err).Msg("Failed to parse existing subforum rules")
			return nil, fmt.Errorf("failed to parse existing rules: %w", err)
		}
	}

	// Find and update the rule
	var updatedRule *apimodels.Rule
	for i, rule := range rules {
		if rule.Code == input.RuleCode {
			// Update fields if provided
			if input.Body.Name != nil {
				rule.Name = *input.Body.Name
			}
			if input.Body.Description != nil {
				rule.Description = *input.Body.Description
			}
			if input.Body.Category != nil {
				rule.Category = *input.Body.Category
			}
			if input.Body.Severity != nil {
				rule.Severity = *input.Body.Severity
			}
			if input.Body.Active != nil {
				rule.Active = *input.Body.Active
			}
			rules[i] = rule
			updatedRule = &rule
			break
		}
	}

	if updatedRule == nil {
		return nil, huma.Error404NotFound("rule not found")
	}

	// Convert back to JSON
	rulesJSON, err := json.Marshal(rules)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal rules to JSON")
		return nil, fmt.Errorf("failed to save rules: %w", err)
	}

	// Update subforum with updated rules
	updateSetter := &dbmodels.SubforumSetter{
		UpdatedAt: &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	// Set the JSON field properly
	jsonValue := types.JSON[json.RawMessage]{}
	if err := jsonValue.Scan(rulesJSON); err != nil {
		log.Error().Err(err).Msg("Failed to scan rules JSON")
		return nil, fmt.Errorf("failed to save rules: %w", err)
	}
	updateSetter.SubforumRules = &sql.Null[types.JSON[json.RawMessage]]{V: jsonValue, Valid: true}

	// Use the Update method on the subforum model
	if err := subforum.Update(ctx, h.db, updateSetter); err != nil {
		log.Error().Err(err).Msg("Failed to update subforum rules")
		return nil, fmt.Errorf("failed to save rule: %w", err)
	}

	log.Info().
		Str("endpoint", "subforum/rules").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int32("subforum_id", subforum.SubforumID).
		Str("rule_code", updatedRule.Code).
		Msg("Subforum rule updated")

	return updatedRule, nil
}

// DeleteSubforumRule deletes a rule from a subforum
func (h *RulesHandler) DeleteSubforumRule(ctx context.Context, input *apimodels.RuleDeleteInput) (*apimodels.RuleDeleteResponse, error) {
	// Extract user from context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	// Validate moderator permissions for the subforum
	if err := h.validateModeratorPermissionsForSubforum(ctx, userCtx, input.CommunityType, input.SubforumName); err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("User lacks moderation permissions")
		return nil, fmt.Errorf("insufficient permissions: %w", err)
	}

	// Get subforum
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, input.CommunityType, input.SubforumName)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get subforum")
		return nil, fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return nil, huma.Error404NotFound("subforum not found")
	}

	// Get existing rules
	var rules []apimodels.Rule
	if subforum.SubforumRules.Valid {
		// Get the raw bytes from the JSON type
		value, err := subforum.SubforumRules.V.Value()
		if err != nil {
			log.Error().Err(err).Msg("Failed to get existing subforum rules value")
			return nil, fmt.Errorf("failed to parse existing rules: %w", err)
		}

		// Convert to bytes
		bytes, ok := value.([]byte)
		if !ok {
			log.Error().Msg("Existing subforum rules value is not bytes")
			return nil, fmt.Errorf("failed to parse existing rules: %w", err)
		}

		if err := json.Unmarshal(bytes, &rules); err != nil {
			log.Error().Err(err).Msg("Failed to parse existing subforum rules")
			return nil, fmt.Errorf("failed to parse existing rules: %w", err)
		}
	}

	// Find and remove the rule
	var deletedRule *apimodels.Rule
	var newRules []apimodels.Rule
	for _, rule := range rules {
		if rule.Code == input.RuleCode {
			deletedRule = &rule
		} else {
			newRules = append(newRules, rule)
		}
	}

	if deletedRule == nil {
		return nil, huma.Error404NotFound("rule not found")
	}

	// Convert back to JSON
	rulesJSON, err := json.Marshal(newRules)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal rules to JSON")
		return nil, fmt.Errorf("failed to save rules: %w", err)
	}

	// Update subforum with updated rules
	updateSetter := &dbmodels.SubforumSetter{
		UpdatedAt: &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	// Set the JSON field properly
	jsonValue := types.JSON[json.RawMessage]{}
	if err := jsonValue.Scan(rulesJSON); err != nil {
		log.Error().Err(err).Msg("Failed to scan rules JSON")
		return nil, fmt.Errorf("failed to save rules: %w", err)
	}
	updateSetter.SubforumRules = &sql.Null[types.JSON[json.RawMessage]]{V: jsonValue, Valid: true}

	// Use the Update method on the subforum model
	if err := subforum.Update(ctx, h.db, updateSetter); err != nil {
		log.Error().Err(err).Msg("Failed to update subforum rules")
		return nil, fmt.Errorf("failed to delete rule: %w", err)
	}

	log.Info().
		Str("endpoint", "subforum/rules").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int32("subforum_id", subforum.SubforumID).
		Str("rule_code", deletedRule.Code).
		Msg("Subforum rule deleted")

	return &apimodels.RuleDeleteResponse{
		Code:      deletedRule.Code,
		DeletedAt: time.Now().Format(time.RFC3339),
		DeletedBy: userCtx.DisplayName,
	}, nil
}

// ReportRuleViolation reports a violation of a specific rule
func (h *RulesHandler) ReportRuleViolation(ctx context.Context, input *apimodels.RuleViolationInput) (*apimodels.RuleViolationResponse, error) {
	// Extract user from context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	log.Info().
		Str("endpoint", "reports/rule-violation").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Str("rule_code", input.Body.RuleCode).
		Str("rule_type", input.Body.RuleType).
		Msg("Rule violation report requested")

	// Validate rule exists
	if input.Body.RuleType == "platform" {
		// Check if platform rule exists
		setting, err := h.systemSettingsDAO.GetSetting(ctx, "platform_rules")
		if err != nil {
			log.Error().Err(err).Msg("Failed to get platform rules")
			return nil, fmt.Errorf("failed to validate rule: %w", err)
		}

		if setting != nil {
			var rules []apimodels.Rule
			if err := json.Unmarshal([]byte(setting.SettingValue), &rules); err != nil {
				log.Error().Err(err).Msg("Failed to parse platform rules")
				return nil, fmt.Errorf("failed to validate rule: %w", err)
			}

			ruleExists := false
			for _, rule := range rules {
				if rule.Code == input.Body.RuleCode && rule.Active {
					ruleExists = true
					break
				}
			}

			if !ruleExists {
				return nil, huma.Error400BadRequest("invalid or inactive platform rule")
			}
		}
	} else if input.Body.RuleType == "subforum" {
		// For subforum rules, we'd need to check the specific subforum
		// This is a simplified version - in practice you'd need to determine the subforum
		// from the content being reported
		return nil, huma.Error400BadRequest("subforum rule validation not implemented")
	} else {
		return nil, huma.Error400BadRequest("invalid rule type")
	}

	// Create report in database
	status := sql.Null[string]{V: "pending", Valid: true}
	ruleCode := sql.Null[string]{V: input.Body.RuleCode, Valid: true}
	ruleType := sql.Null[string]{V: input.Body.RuleType, Valid: true}

	reportSetter := &dbmodels.ReportSetter{
		ReporterPseudonymID: &userCtx.ActivePseudonymID,
		ContentType:         &input.Body.ContentType,
		ReportReason:        &input.Body.RuleCode, // Use rule code as reason
		Status:              &status,
		RuleCode:            &ruleCode,
		RuleType:            &ruleType,
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
		log.Error().Err(err).Msg("Failed to create rule violation report")
		return nil, err
	}

	response := &apimodels.RuleViolationResponse{
		ReportID:     int(report.ReportID),
		RuleCode:     input.Body.RuleCode,
		RuleType:     input.Body.RuleType,
		ReportStatus: "pending",
		CreatedAt:    time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("endpoint", "reports/rule-violation").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int("report_id", int(report.ReportID)).
		Msg("Rule violation report created")

	return response, nil
}

// ForwardReportToPlatform forwards a report to platform-level moderators
func (h *RulesHandler) ForwardReportToPlatform(ctx context.Context, input *apimodels.ReportForwardInput) (*apimodels.ReportForwardResponse, error) {
	// Extract user from context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	// Check specific capability for forwarding reports
	if !userCtx.HasCapability(constants.CapabilityForwardReports) {
		log.Error().Int("user_id", int(userCtx.UserID)).Msg("User lacks forward_reports capability")
		return nil, fmt.Errorf("insufficient permissions: forward_reports capability required")
	}

	log.Info().
		Str("endpoint", "reports/forward").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int("report_id", input.ReportID).
		Msg("Forward report to platform requested")

	// Get the report
	report, err := h.reportDAO.GetReportByID(ctx, int64(input.ReportID))
	if err != nil {
		log.Error().Err(err).Int("report_id", input.ReportID).Msg("Failed to get report")
		return nil, fmt.Errorf("report not found: %w", err)
	}

	if report == nil {
		return nil, huma.Error404NotFound("report not found")
	}

	// Check if already forwarded
	if report.ForwardedToPlatform.Valid && report.ForwardedToPlatform.V {
		return nil, huma.Error400BadRequest("report already forwarded to platform")
	}

	// Update report to mark as forwarded
	forwardedToPlatform := sql.Null[bool]{V: true, Valid: true}
	forwardingNotes := sql.Null[string]{V: input.Body.ForwardingNotes, Valid: true}
	forwardedByUserID := sql.Null[int64]{V: userCtx.UserID, Valid: true}
	forwardedAt := sql.Null[time.Time]{V: time.Now(), Valid: true}

	updateSetter := &dbmodels.ReportSetter{
		ForwardedToPlatform: &forwardedToPlatform,
		ForwardingNotes:     &forwardingNotes,
		ForwardedByUserID:   &forwardedByUserID,
		ForwardedAt:         &forwardedAt,
	}

	if err := h.reportDAO.UpdateReport(ctx, int64(input.ReportID), updateSetter); err != nil {
		log.Error().Err(err).Int("report_id", input.ReportID).Msg("Failed to update report")
		return nil, fmt.Errorf("failed to forward report: %w", err)
	}

	response := &apimodels.ReportForwardResponse{
		ReportID:            input.ReportID,
		ForwardedToPlatform: true,
		ForwardingNotes:     input.Body.ForwardingNotes,
		ForwardedBy:         userCtx.DisplayName,
		ForwardedAt:         time.Now().Format(time.RFC3339),
	}

	log.Info().
		Str("endpoint", "reports/forward").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int("report_id", input.ReportID).
		Msg("Report forwarded to platform")

	return response, nil
}

func (h *RulesHandler) validateModeratorPermissionsForSubforum(ctx context.Context, userCtx *middleware.UserContext, communityType, subforumName string) error {
	// Get subforum by community type and name
	subforum, err := h.subforumDAO.GetSubforumByCommunityTypeAndName(ctx, communityType, subforumName)
	if err != nil {
		log.Error().Err(err).Str("community_type", communityType).Str("subforum_name", subforumName).Msg("Failed to get subforum")
		return fmt.Errorf("subforum not found: %w", err)
	}

	if subforum == nil {
		return fmt.Errorf("subforum not found")
	}

	// Check capability for managing subforum rules using unified permission system
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityManageSubforumRules, &subforum.SubforumID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to check unified capability")
		return fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Warn().Int64("user_id", userCtx.UserID).Msg("User lacks manage_subforum_rules capability")
		return fmt.Errorf("insufficient permissions to manage subforum rules")
	}

	return nil
}
