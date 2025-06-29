package models

// Pagination represents pagination information
type Pagination struct {
	Page  int `json:"page" example:"1"`
	Limit int `json:"limit" example:"25"`
	Total int `json:"total" example:"1500"`
	Pages int `json:"pages" example:"60"`
}

// Author represents a user author in responses
type Author struct {
	PseudonymID string `json:"pseudonym_id" example:"abc123def456..."`
	DisplayName string `json:"display_name" example:"user_display_name"`
}

// SubforumInfo represents basic subforum information
type SubforumInfo struct {
	SubforumID  int    `json:"subforum_id" example:"1"`
	Name        string `json:"name" example:"golang"`
	DisplayName string `json:"display_name" example:"Golang"`
}

// Moderator represents a moderator in responses
type Moderator struct {
	PseudonymID string `json:"pseudonym_id" example:"mod_pseudonym_id"`
	DisplayName string `json:"display_name" example:"moderator_name"`
	Role        string `json:"role" example:"owner"`
}

// Reporter represents a user who reported content
type Reporter struct {
	PseudonymID string `json:"pseudonym_id" example:"reporter_pseudonym_id"`
	DisplayName string `json:"display_name" example:"reporter_name"`
}

// ReportedUser represents a user who was reported
type ReportedUser struct {
	PseudonymID string `json:"pseudonym_id" example:"reported_pseudonym_id"`
	DisplayName string `json:"display_name" example:"reported_user_name"`
}

// Content represents reported content
type Content struct {
	Title   string `json:"title" example:"Reported Post Title"`
	Content string `json:"content" example:"Reported post content..."`
}

// ResolvedBy represents who resolved a report
type ResolvedBy struct {
	PseudonymID string `json:"pseudonym_id" example:"mod_pseudonym_id"`
	DisplayName string `json:"display_name" example:"moderator_name"`
}

// RemovedBy represents who removed content
type RemovedBy struct {
	PseudonymID string `json:"pseudonym_id" example:"mod_pseudonym_id"`
	DisplayName string `json:"display_name" example:"moderator_name"`
}

// BannedBy represents who banned a user
type BannedBy struct {
	PseudonymID string `json:"pseudonym_id" example:"mod_pseudonym_id"`
	DisplayName string `json:"display_name" example:"moderator_name"`
}

// ActionDetails represents details of a moderation action
type ActionDetails struct {
	RemovalReason string `json:"removal_reason" example:"violates community guidelines"`
}
