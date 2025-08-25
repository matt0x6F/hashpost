package models

import (
	"github.com/matt0x6f/hashpost/internal/api/middleware"
)

// Rule represents a configurable rule (platform or subforum)
type Rule struct {
	Code        string `json:"code" example:"harassment"`
	Name        string `json:"name" example:"No Harassment"`
	Description string `json:"description" example:"Harassment, bullying, or targeted abuse of any kind is not allowed..."`
	Category    string `json:"category" example:"safety"`
	Severity    string `json:"severity" example:"high"`
	Active      bool   `json:"active" example:"true"`
}

// PlatformRulesInput represents the input for getting platform rules
type PlatformRulesInput struct {
	ActiveOnly bool `query:"active_only" example:"true"`
}

// PlatformRulesResponse represents the response for platform rules
type PlatformRulesResponse struct {
	Status int `json:"status" example:"200"`
	Body   struct {
		Rules []Rule `json:"rules"`
	} `json:"body"`
}

// SubforumRulesInput represents the input for getting subforum rules
type SubforumRulesInput struct {
	CommunityType string `path:"community_type" example:"t"`
	SubforumName  string `path:"subforum_name" example:"golang"`
	ActiveOnly    bool   `query:"active_only" example:"true"`
}

// SubforumRulesResponse represents the response for subforum rules
type SubforumRulesResponse struct {
	Status int `json:"status" example:"200"`
	Body   struct {
		SubforumID int32  `json:"subforum_id" example:"1"`
		Name       string `json:"name" example:"golang"`
		Rules      []Rule `json:"rules"`
	} `json:"body"`
}

// RuleCreateInputBody represents the input for creating a new rule
type RuleCreateInputBody struct {
	Code        string `json:"code" example:"no_politics" required:"true"`
	Name        string `json:"name" example:"No Political Discussion" required:"true"`
	Description string `json:"description" example:"Political discussions are not allowed in this community" required:"true"`
	Category    string `json:"category" example:"content" required:"true"`
	Severity    string `json:"severity" example:"medium" required:"true"`
	Active      bool   `json:"active" example:"true"`
}

// RuleCreateInput represents the input for creating a new rule
type RuleCreateInput struct {
	middleware.AuthInput
	CommunityType string              `path:"community_type" example:"t"`
	SubforumName  string              `path:"subforum_name" example:"golang"`
	Body          RuleCreateInputBody `json:"body"`
}

// RuleUpdateInputBody represents the input for updating a rule
type RuleUpdateInputBody struct {
	Name        *string `json:"name,omitempty" example:"No Political Discussion"`
	Description *string `json:"description,omitempty" example:"Political discussions are not allowed in this community"`
	Category    *string `json:"category,omitempty" example:"content"`
	Severity    *string `json:"severity,omitempty" example:"medium"`
	Active      *bool   `json:"active,omitempty" example:"true"`
}

// RuleUpdateInput represents the input for updating a rule
type RuleUpdateInput struct {
	middleware.AuthInput
	CommunityType string              `path:"community_type" example:"t"`
	SubforumName  string              `path:"subforum_name" example:"golang"`
	RuleCode      string              `path:"rule_code" example:"no_politics"`
	Body          RuleUpdateInputBody `json:"body"`
}

// RuleDeleteInput represents the input for deleting a rule
type RuleDeleteInput struct {
	middleware.AuthInput
	CommunityType string `path:"community_type" example:"t"`
	SubforumName  string `path:"subforum_name" example:"golang"`
	RuleCode      string `path:"rule_code" example:"no_politics"`
}

// RuleDeleteResponse represents the response for deleting a rule
type RuleDeleteResponse struct {
	Code      string `json:"code" example:"no_politics"`
	DeletedAt string `json:"deleted_at" example:"2024-01-01T16:00:00Z"`
	DeletedBy string `json:"deleted_by" example:"moderator1"`
}

// ReportWithRule represents a report with rule information
type ReportWithRule struct {
	ReportID            int          `json:"report_id" example:"789"`
	ContentType         string       `json:"content_type" example:"post"`
	ContentID           *int         `json:"content_id" example:"123"`
	ReportedPseudonymID string       `json:"reported_pseudonym_id" example:"def789ghi012..."`
	ReportReason        string       `json:"report_reason" example:"spam"`
	ReportDetails       string       `json:"report_details" example:"This post violates community guidelines..."`
	ReportStatus        string       `json:"status" example:"pending"`
	CreatedAt           string       `json:"created_at" example:"2024-01-01T16:00:00Z"`
	ResolvedBy          *ResolvedBy  `json:"resolved_by"`
	ResolvedAt          string       `json:"resolved_at" example:"2024-01-01T17:00:00Z"`
	ResolutionNotes     string       `json:"resolution_notes" example:"Post removed for violation of community guidelines"`
	Reporter            Reporter     `json:"reporter"`
	ReportedUser        ReportedUser `json:"reported_user"`
	Content             *Content     `json:"content"`
	Rule                *Rule        `json:"rule,omitempty"`
	RuleType            string       `json:"rule_type,omitempty" example:"platform"`
	ForwardedToPlatform bool         `json:"forwarded_to_platform,omitempty" example:"false"`
	ForwardingNotes     string       `json:"forwarding_notes,omitempty" example:"This appears to be a systemic issue"`
	ForwardedBy         string       `json:"forwarded_by,omitempty" example:"moderator1"`
	ForwardedAt         string       `json:"forwarded_at,omitempty" example:"2024-01-01T18:00:00Z"`
}

// ReportForwardInputBody represents the input for forwarding a report to platform
type ReportForwardInputBody struct {
	ForwardingNotes string `json:"forwarding_notes" example:"This appears to be a systemic issue that requires platform-level attention" required:"true"`
}

// ReportForwardInput represents the input for forwarding a report
type ReportForwardInput struct {
	middleware.AuthInput
	ReportID int                    `path:"report_id" example:"789"`
	Body     ReportForwardInputBody `json:"body"`
}

// ReportForwardResponse represents the response for forwarding a report
type ReportForwardResponse struct {
	ReportID            int    `json:"report_id" example:"789"`
	ForwardedToPlatform bool   `json:"forwarded_to_platform" example:"true"`
	ForwardingNotes     string `json:"forwarding_notes" example:"This appears to be a systemic issue"`
	ForwardedBy         string `json:"forwarded_by" example:"moderator1"`
	ForwardedAt         string `json:"forwarded_at" example:"2024-01-01T18:00:00Z"`
}

// RuleViolationInputBody represents the input for reporting a rule violation
type RuleViolationInputBody struct {
	ContentType         string `json:"content_type" example:"post" required:"true"`
	ContentID           *int   `json:"content_id" example:"123"`
	ReportedPseudonymID string `json:"reported_pseudonym_id" example:"def789ghi012..."`
	RuleCode            string `json:"rule_code" example:"harassment" required:"true"`
	RuleType            string `json:"rule_type" example:"platform" required:"true"` // 'platform' or 'subforum'
	ReportDetails       string `json:"report_details" example:"This post violates the harassment rule..." required:"true"`
}

// RuleViolationInput represents the input for reporting a rule violation
type RuleViolationInput struct {
	middleware.AuthInput
	Body RuleViolationInputBody `json:"body"`
}

// RuleViolationResponse represents the response for reporting a rule violation
type RuleViolationResponse struct {
	ReportID     int    `json:"report_id" example:"789"`
	RuleCode     string `json:"rule_code" example:"harassment"`
	RuleType     string `json:"rule_type" example:"platform"`
	ReportStatus string `json:"status" example:"pending"`
	CreatedAt    string `json:"created_at" example:"2024-01-01T16:00:00Z"`
}

// PlatformRulesUpdateInputBody represents the input for updating platform rules
type PlatformRulesUpdateInputBody struct {
	Rules []Rule `json:"rules" required:"true"`
}

// PlatformRulesUpdateInput represents the input for updating platform rules
type PlatformRulesUpdateInput struct {
	middleware.AuthInput
	Body PlatformRulesUpdateInputBody `json:"body"`
}
