package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
	"github.com/matt0x6f/hashpost/internal/api/models"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/rs/zerolog/log"
)

// SystemSettingsHandler handles system-wide settings operations
type SystemSettingsHandler struct {
	systemSettingsDAO dao.SystemSettingsDAOInterface
	permissionDAO     dao.PermissionDAOInterface
	pseudonymDAO      dao.PseudonymDAOInterface
	ibeSystem         ibe.IBESystemInterface
}

// NewSystemSettingsHandler creates a new SystemSettingsHandler
func NewSystemSettingsHandler(
	systemSettingsDAO dao.SystemSettingsDAOInterface,
	permissionDAO dao.PermissionDAOInterface,
	pseudonymDAO dao.PseudonymDAOInterface,
	ibeSystem ibe.IBESystemInterface,
) *SystemSettingsHandler {
	return &SystemSettingsHandler{
		systemSettingsDAO: systemSettingsDAO,
		permissionDAO:     permissionDAO,
		pseudonymDAO:      pseudonymDAO,
		ibeSystem:         ibeSystem,
	}
}

// UpdateSystemSetting updates a single system setting
func (h *SystemSettingsHandler) UpdateSystemSetting(ctx context.Context, input *models.SystemSettingUpdateInput) (*models.SystemSettingResponse, error) {
	// Extract user from context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	// Check if user has system admin capability
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, "system_admin", nil)
	if err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("Failed to check system_admin capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("User lacks system_admin capability")
		return nil, fmt.Errorf("insufficient permissions: system_admin capability required")
	}

	// Validate setting value based on type
	if err := h.validateSettingValue(input.Body.SettingType, input.Body.SettingValue); err != nil {
		return nil, fmt.Errorf("invalid setting value: %w", err)
	}

	// Save setting to database
	err = h.systemSettingsDAO.SetSetting(ctx, input.Body.SettingKey, input.Body.SettingValue, input.Body.SettingType, userCtx.UserID)
	if err != nil {
		log.Error().Err(err).Str("setting_key", input.Body.SettingKey).Msg("Failed to save system setting")
		return nil, fmt.Errorf("failed to save setting: %w", err)
	}

	// Update last active timestamp for the pseudonym
	err = h.pseudonymDAO.UpdateLastActive(ctx, userCtx.ActivePseudonymID)
	if err != nil {
		log.Error().Err(err).Str("pseudonym_id", userCtx.ActivePseudonymID).Msg("Failed to update pseudonym last active timestamp")
		// Don't fail the request for this error
	}

	log.Info().
		Str("endpoint", "system/settings").
		Str("component", "handler").
		Int("user_id", int(userCtx.UserID)).
		Str("setting_key", input.Body.SettingKey).
		Msg("System setting updated")

	return &models.SystemSettingResponse{
		Status: 200,
		Body: models.SystemSettingResponseBody{
			SettingKey:   input.Body.SettingKey,
			SettingValue: input.Body.SettingValue,
			SettingType:  input.Body.SettingType,
			Message:      "Setting updated successfully",
		},
	}, nil
}

// GetSystemSetting retrieves a single system setting
func (h *SystemSettingsHandler) GetSystemSetting(ctx context.Context, input *models.SystemSettingGetInput) (*models.SystemSettingResponse, error) {
	// Extract user from context
	userCtx, err := middleware.ExtractUserFromHumaInput(&input.AuthInput)
	if err != nil {
		log.Error().Err(err).Msg("Failed to extract user from context")
		return nil, fmt.Errorf("authentication required")
	}

	// Check if user has system admin capability
	hasCapability, err := h.permissionDAO.HasUnifiedCapability(ctx, userCtx.UserID, userCtx.ActivePseudonymID, "system_admin", nil)
	if err != nil {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("Failed to check system_admin capability")
		return nil, fmt.Errorf("failed to check permissions: %w", err)
	}
	if !hasCapability {
		log.Error().Err(err).Int("user_id", int(userCtx.UserID)).Msg("User lacks system_admin capability")
		return nil, fmt.Errorf("insufficient permissions: system_admin capability required")
	}

	// Get setting from database
	setting, err := h.systemSettingsDAO.GetSetting(ctx, input.SettingKey)
	if err != nil {
		log.Error().Err(err).Str("setting_key", input.SettingKey).Msg("Failed to get system setting")
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	if setting == nil {
		return nil, fmt.Errorf("setting not found: %s", input.SettingKey)
	}

	return &models.SystemSettingResponse{
		Status: 200,
		Body: models.SystemSettingResponseBody{
			SettingKey:   setting.SettingKey,
			SettingValue: setting.SettingValue,
			SettingType:  setting.SettingType,
			Message:      "Setting retrieved successfully",
		},
	}, nil
}

// validateSettingValue validates the setting value based on its type
func (h *SystemSettingsHandler) validateSettingValue(settingType, settingValue string) error {
	switch settingType {
	case "string":
		// String values are always valid
		return nil
	case "boolean":
		if settingValue != "true" && settingValue != "false" {
			return fmt.Errorf("boolean value must be 'true' or 'false'")
		}
	case "number":
		// Try to parse as number
		var num float64
		if err := json.Unmarshal([]byte(settingValue), &num); err != nil {
			return fmt.Errorf("invalid number format")
		}
	case "json":
		// Try to parse as JSON
		var js json.RawMessage
		if err := json.Unmarshal([]byte(settingValue), &js); err != nil {
			return fmt.Errorf("invalid JSON format")
		}
	default:
		return fmt.Errorf("unsupported setting type: %s", settingType)
	}
	return nil
}
