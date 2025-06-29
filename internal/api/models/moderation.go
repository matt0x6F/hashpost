package models

import "time"

// ReportInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type ReportInputBody struct {
	ContentType         string `json:"content_type" example:"post" required:"true"`     // "post", "comment", "user", "subforum"
	ContentID           *int   `json:"content_id" example:"123"`                        // Required for post/comment reports
	ReportedPseudonymID string `json:"reported_pseudonym_id" example:"def789ghi012..."` // Required for user reports
	ReportReason        string `json:"report_reason" example:"spam" required:"true"`    // "spam", "harassment", "violence", "misinformation", etc.
	ReportDetails       string `json:"report_details" example:"This post violates community guidelines..." required:"true"`
}

// ReportInput represents content or user report request (for OpenAPI schema only)
type ReportInput struct {
	Body ReportInputBody `json:"body"`
}

// Report represents a content or user report
type Report struct {
	ReportID            int          `json:"report_id" example:"789"`
	ContentType         string       `json:"content_type" example:"post"`
	ContentID           *int         `json:"content_id" example:"123"`
	ReportedPseudonymID string       `json:"reported_pseudonym_id" example:"def789ghi012..."`
	ReportReason        string       `json:"report_reason" example:"spam"`
	ReportDetails       string       `json:"report_details" example:"This post violates community guidelines..."`
	Status              string       `json:"status" example:"pending"` // "pending", "investigating", "resolved", "dismissed"
	CreatedAt           string       `json:"created_at" example:"2024-01-01T16:00:00Z"`
	ResolvedBy          *ResolvedBy  `json:"resolved_by"`
	ResolvedAt          string       `json:"resolved_at" example:"2024-01-01T17:00:00Z"`
	ResolutionNotes     string       `json:"resolution_notes" example:"Post removed for violation of community guidelines"`
	Reporter            Reporter     `json:"reporter"`
	ReportedUser        ReportedUser `json:"reported_user"`
	Content             *Content     `json:"content"`
}

// ReportsListInput represents reports list request parameters
type ReportsListInput struct {
	SubforumID int    `query:"subforum_id" example:"1"`  // 0 means "all subforums"
	Status     string `query:"status" example:"pending"` // "pending", "investigating", "resolved", "dismissed"
	Page       int    `query:"page" example:"1"`
	Limit      int    `query:"limit" example:"25"`
}

// ContentRemovalInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type ContentRemovalInputBody struct {
	RemovalReason    string `json:"removal_reason" example:"violates community guidelines" required:"true"`
	SendNotification bool   `json:"send_notification" example:"true"`
}

// ContentRemovalInput represents content removal request (for OpenAPI schema only)
type ContentRemovalInput struct {
	ContentType string                  `path:"content_type" example:"post"` // "post", "comment"
	ContentID   int                     `path:"content_id" example:"123"`
	Body        ContentRemovalInputBody `json:"body"`
}

// UserBanInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type UserBanInputBody struct {
	SubforumID       int    `json:"subforum_id" example:"1" required:"true"`
	BanReason        string `json:"ban_reason" example:"Repeated violations of community guidelines" required:"true"`
	IsPermanent      bool   `json:"is_permanent" example:"false"`
	DurationDays     *int   `json:"duration_days" example:"30"`
	SendNotification bool   `json:"send_notification" example:"true"`
}

// UserBanInput represents user ban request (for OpenAPI schema only)
type UserBanInput struct {
	PseudonymID string           `path:"pseudonym_id" example:"def789ghi012..."`
	Body        UserBanInputBody `json:"body"`
}

// ModerationHistoryInput represents moderation history request parameters
type ModerationHistoryInput struct {
	SubforumID int    `query:"subforum_id" example:"1"`           // 0 means "all subforums"
	ActionType string `query:"action_type" example:"remove_post"` // "remove_post", "remove_comment", "ban_user", "unban_user"
	Page       int    `query:"page" example:"1"`
	Limit      int    `query:"limit" example:"25"`
}

// ModerationAction represents a moderation action
type ModerationAction struct {
	ActionID          int               `json:"action_id" example:"123"`
	ActionType        string            `json:"action_type" example:"remove_post"`
	TargetContentType string            `json:"target_content_type" example:"post"`
	TargetContentID   int               `json:"target_content_id" example:"456"`
	ActionDetails     ActionDetails     `json:"action_details"`
	CreatedAt         string            `json:"created_at" example:"2024-01-01T17:00:00Z"`
	Moderator         Moderator         `json:"moderator"`
	Subforum          SubforumModerator `json:"subforum"`
}

// ReportResponseBody represents the body of report creation response
type ReportResponseBody struct {
	ReportID  int    `json:"report_id" example:"789"`
	Status    string `json:"status" example:"pending"`
	CreatedAt string `json:"created_at" example:"2024-01-01T16:00:00Z"`
}

// ReportsListResponseBody represents the body of reports list response
type ReportsListResponseBody struct {
	Reports    []Report   `json:"reports"`
	Pagination Pagination `json:"pagination"`
}

// ContentRemovalResponseBody represents the body of content removal response
type ContentRemovalResponseBody struct {
	ContentID     int       `json:"content_id" example:"123"`
	ContentType   string    `json:"content_type" example:"post"`
	Removed       bool      `json:"removed" example:"true"`
	RemovalReason string    `json:"removal_reason" example:"violates community guidelines"`
	RemovedAt     string    `json:"removed_at" example:"2024-01-01T17:00:00Z"`
	RemovedBy     RemovedBy `json:"removed_by"`
}

// UserBanResponseBody represents the body of user ban response
type UserBanResponseBody struct {
	BanID             int      `json:"ban_id" example:"123"`
	BannedFingerprint string   `json:"banned_fingerprint" example:"a1b2c3d4e5f6..."`
	SubforumID        int      `json:"subforum_id" example:"1"`
	BanReason         string   `json:"ban_reason" example:"Repeated violations of community guidelines"`
	IsPermanent       bool     `json:"is_permanent" example:"false"`
	ExpiresAt         string   `json:"expires_at" example:"2024-02-01T17:00:00Z"`
	CreatedAt         string   `json:"created_at" example:"2024-01-01T17:00:00Z"`
	BannedBy          BannedBy `json:"banned_by"`
}

// ModerationHistoryResponseBody represents the body of moderation history response
type ModerationHistoryResponseBody struct {
	Actions    []ModerationAction `json:"actions"`
	Pagination Pagination         `json:"pagination"`
}

// ReportResponse represents report creation response
type ReportResponse struct {
	Status int                `json:"-" example:"200"`
	Body   ReportResponseBody `json:"body"`
}

// ReportsListResponse represents reports list response
type ReportsListResponse struct {
	Status int                     `json:"-" example:"200"`
	Body   ReportsListResponseBody `json:"body"`
}

// ContentRemovalResponse represents content removal response
type ContentRemovalResponse struct {
	Status int                        `json:"-" example:"200"`
	Body   ContentRemovalResponseBody `json:"body"`
}

// UserBanResponse represents user ban response
type UserBanResponse struct {
	Status int                 `json:"-" example:"200"`
	Body   UserBanResponseBody `json:"body"`
}

// ModerationHistoryResponse represents moderation history response
type ModerationHistoryResponse struct {
	Status int                           `json:"-" example:"200"`
	Body   ModerationHistoryResponseBody `json:"body"`
}

// NewReportResponse creates a new report response
func NewReportResponse(reportID int) *ReportResponse {
	return &ReportResponse{
		Status: 200,
		Body: ReportResponseBody{
			ReportID:  reportID,
			Status:    "pending",
			CreatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// NewReportsListResponse creates a new reports list response
func NewReportsListResponse(reports []Report, page, limit, total int) *ReportsListResponse {
	pages := (total + limit - 1) / limit // Ceiling division

	return &ReportsListResponse{
		Status: 200,
		Body: ReportsListResponseBody{
			Reports: reports,
			Pagination: Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: pages,
			},
		},
	}
}

// NewContentRemovalResponse creates a new content removal response
func NewContentRemovalResponse(contentID int, contentType, removalReason string, moderatorPseudonymID, moderatorDisplayName string) *ContentRemovalResponse {
	return &ContentRemovalResponse{
		Status: 200,
		Body: ContentRemovalResponseBody{
			ContentID:     contentID,
			ContentType:   contentType,
			Removed:       true,
			RemovalReason: removalReason,
			RemovedAt:     time.Now().UTC().Format(time.RFC3339),
			RemovedBy: RemovedBy{
				PseudonymID: moderatorPseudonymID,
				DisplayName: moderatorDisplayName,
			},
		},
	}
}

// NewUserBanResponse creates a new user ban response
func NewUserBanResponse(banID int, bannedFingerprint string, subforumID int, banReason string, isPermanent bool, durationDays *int, moderatorPseudonymID, moderatorDisplayName string) *UserBanResponse {
	expiresAt := ""
	if !isPermanent && durationDays != nil {
		expiresAt = time.Now().AddDate(0, 0, *durationDays).UTC().Format(time.RFC3339)
	}

	return &UserBanResponse{
		Status: 200,
		Body: UserBanResponseBody{
			BanID:             banID,
			BannedFingerprint: bannedFingerprint,
			SubforumID:        subforumID,
			BanReason:         banReason,
			IsPermanent:       isPermanent,
			ExpiresAt:         expiresAt,
			CreatedAt:         time.Now().UTC().Format(time.RFC3339),
			BannedBy: BannedBy{
				PseudonymID: moderatorPseudonymID,
				DisplayName: moderatorDisplayName,
			},
		},
	}
}

// NewModerationHistoryResponse creates a new moderation history response
func NewModerationHistoryResponse(actions []ModerationAction, page, limit, total int) *ModerationHistoryResponse {
	pages := (total + limit - 1) / limit // Ceiling division

	return &ModerationHistoryResponse{
		Status: 200,
		Body: ModerationHistoryResponseBody{
			Actions: actions,
			Pagination: Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: pages,
			},
		},
	}
}
