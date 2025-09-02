package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/constants"
	"github.com/matt0x6f/hashpost/internal/api/middleware"
	apimodels "github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/rs/zerolog/log"
)

// RulesHandler handles rule-related requests
type RulesHandler struct {
	reportDAO         dao.ReportDAOInterface
	subforumDAO       dao.SubforumDAOInterface
	systemSettingsDAO dao.SystemSettingsDAOInterface
	permissionDAO     dao.PermissionDAOInterface
	pseudonymDAO      dao.PseudonymDAOInterface
}

// NewRulesHandler creates a new rules handler
func NewRulesHandler(
	reportDAO dao.ReportDAOInterface,
	subforumDAO dao.SubforumDAOInterface,
	systemSettingsDAO dao.SystemSettingsDAOInterface,
	permissionDAO dao.PermissionDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
) *RulesHandler {
	return &RulesHandler{
		reportDAO:         reportDAO,
		subforumDAO:       subforumDAO,
		systemSettingsDAO: systemSettingsDAO,
		permissionDAO:     permissionDAO,
		pseudonymDAO:      pseudonymDAO,
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

	// Update subforum with new rules through DAO
	if err := h.subforumDAO.UpdateRules(ctx, subforum.SubforumID, rulesJSON); err != nil {
		log.Error().Err(err).Msg("Failed to update subforum rules")
		return nil, fmt.Errorf("failed to save rule: %w", err)
	}

	// Update last active timestamp for the pseudonym since creating rules represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, userCtx.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", userCtx.ActivePseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
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

	// Update subforum with updated rules through DAO
	if err := h.subforumDAO.UpdateRules(ctx, subforum.SubforumID, rulesJSON); err != nil {
		log.Error().Err(err).Msg("Failed to update subforum rules")
		return nil, fmt.Errorf("failed to save rule: %w", err)
	}

	// Update last active timestamp for the pseudonym since updating rules represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, userCtx.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", userCtx.ActivePseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
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

	// Update subforum with updated rules through DAO
	if err := h.subforumDAO.UpdateRules(ctx, subforum.SubforumID, rulesJSON); err != nil {
		log.Error().Err(err).Msg("Failed to update subforum rules")
		return nil, fmt.Errorf("failed to delete rule: %w", err)
	}

	// Update last active timestamp for the pseudonym since deleting rules represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, userCtx.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", userCtx.ActivePseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
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

	// Create report using the DAO method
	var contentID *int64
	if input.Body.ContentID != nil {
		contentIDVal := int64(*input.Body.ContentID)
		contentID = &contentIDVal
	}

	report, err := h.reportDAO.CreateRuleViolationReport(
		ctx,
		userCtx.ActivePseudonymID,
		input.Body.ContentType,
		input.Body.RuleCode,
		input.Body.RuleType,
		contentID,
		input.Body.ReportedPseudonymID,
		input.Body.ReportDetails,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create rule violation report")
		return nil, err
	}

	// Update last active timestamp for the pseudonym since reporting represents activity
	err = h.pseudonymDAO.UpdateLastActive(ctx, userCtx.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", userCtx.ActivePseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
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
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilityForwardReports, nil)
	if err != nil {
		log.Error().Err(err).Int64("user_id", userCtx.UserID).Msg("Failed to check forward_reports capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
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

	// Update report to mark as forwarded using the DAO method
	if err := h.reportDAO.UpdateReportWithForwarding(ctx, int64(input.ReportID), input.Body.ForwardingNotes, userCtx.UserID); err != nil {
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

// UpdatePlatformRules updates platform-wide rules
func (h *RulesHandler) UpdatePlatformRules(ctx context.Context, input *apimodels.PlatformRulesUpdateInput) (*apimodels.PlatformRulesResponse, error) {
	// Extract user from context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	// Check if user has system admin capability
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, constants.CapabilitySystemAdmin, nil)
	if err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("Failed to check system_admin capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("User lacks system_admin capability")
		return nil, fmt.Errorf("insufficient permissions: system_admin capability required")
	}

	// Validate rules
	if len(input.Body.Rules) == 0 {
		return nil, huma.Error400BadRequest("at least one rule is required")
	}

	// Check for duplicate rule codes
	ruleCodes := make(map[string]bool)
	for _, rule := range input.Body.Rules {
		if rule.Code == "" {
			return nil, huma.Error400BadRequest("rule code is required")
		}
		if ruleCodes[rule.Code] {
			return nil, huma.Error400BadRequest("duplicate rule code: " + rule.Code)
		}
		ruleCodes[rule.Code] = true
	}

	// Convert rules to JSON
	rulesJSON, err := json.Marshal(input.Body.Rules)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal platform rules to JSON")
		return nil, fmt.Errorf("failed to save rules: %w", err)
	}

	// Save to system settings
	err = h.systemSettingsDAO.SetSetting(ctx, "platform_rules", string(rulesJSON), "json", userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Msg("Failed to save platform rules to system settings")
		return nil, fmt.Errorf("failed to save platform rules: %w", err)
	}

	log.Info().
		Str("endpoint", "platform/rules").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Int("rules_count", len(input.Body.Rules)).
		Msg("Platform rules updated")

	return &apimodels.PlatformRulesResponse{
		Status: 200,
		Body: struct {
			Rules []apimodels.Rule `json:"rules"`
		}{
			Rules: input.Body.Rules,
		},
	}, nil
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
