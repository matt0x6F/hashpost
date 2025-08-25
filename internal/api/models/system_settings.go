package models

import (
	"github.com/matt0x6f/hashpost/internal/api/middleware"
)

// SystemSettingUpdateInputBody represents the input for updating a system setting
type SystemSettingUpdateInputBody struct {
	SettingKey   string `json:"setting_key" required:"true" example:"platform_name"`
	SettingValue string `json:"setting_value" required:"true" example:"HashPost"`
	SettingType  string `json:"setting_type" required:"true" example:"string"`
}

// SystemSettingUpdateInput represents the input for updating a system setting
type SystemSettingUpdateInput struct {
	middleware.AuthInput
	Body SystemSettingUpdateInputBody `json:"body"`
}

// SystemSettingGetInput represents the input for getting a system setting
type SystemSettingGetInput struct {
	middleware.AuthInput
	SettingKey string `path:"setting_key" required:"true" example:"platform_name"`
}

// SystemSettingResponseBody represents the response body for system setting operations
type SystemSettingResponseBody struct {
	SettingKey   string `json:"setting_key" example:"platform_name"`
	SettingValue string `json:"setting_value" example:"HashPost"`
	SettingType  string `json:"setting_type" example:"string"`
	Message      string `json:"message" example:"Setting updated successfully"`
}

// SystemSettingResponse represents the response for system setting operations
type SystemSettingResponse struct {
	Status int                       `json:"status" example:"200"`
	Body   SystemSettingResponseBody `json:"body"`
}
