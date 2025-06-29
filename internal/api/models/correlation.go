package models

// FingerprintCorrelationInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type FingerprintCorrelationInputBody struct {
	RequestedPseudonym   string `json:"requested_pseudonym" example:"abc123def456..." required:"true"`
	RequestedFingerprint string `json:"requested_fingerprint" example:"a1b2c3d4e5f6..." required:"true"`
	Justification        string `json:"justification" example:"Investigation of ban evasion in r/golang" required:"true"`
	SubforumID           int    `json:"subforum_id" example:"1" required:"true"`
	IncidentID           string `json:"incident_id" example:"ban_evasion_123" required:"true"`
}

// FingerprintCorrelationInput represents fingerprint correlation request (for OpenAPI schema only)
type FingerprintCorrelationInput struct {
	Body FingerprintCorrelationInputBody `json:"body"`
}

// IdentityCorrelationInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type IdentityCorrelationInputBody struct {
	RequestedPseudonym   string `json:"requested_pseudonym" example:"abc123def456..." required:"true"`
	RequestedFingerprint string `json:"requested_fingerprint" example:"a1b2c3d4e5f6..." required:"true"`
	Justification        string `json:"justification" example:"Investigation of reported harassment across subforums" required:"true"`
	LegalBasis           string `json:"legal_basis" example:"Platform Terms of Service" required:"true"`
	IncidentID           string `json:"incident_id" example:"harassment_case_123" required:"true"`
	Scope                string `json:"scope" example:"platform_wide" required:"true"`
}

// IdentityCorrelationInput represents identity correlation request (for OpenAPI schema only)
type IdentityCorrelationInput struct {
	Body IdentityCorrelationInputBody `json:"body"`
}

// CorrelationHistoryInput represents correlation history request parameters
type CorrelationHistoryInput struct {
	CorrelationType string `query:"correlation_type" example:"fingerprint"` // "fingerprint", "identity"
	Page            int    `query:"page" example:"1"`
	Limit           int    `query:"limit" example:"25"`
}

// CorrelationResult represents a correlation result
type CorrelationResult struct {
	PseudonymID           string    `json:"pseudonym_id" example:"def789ghi012..."`
	DisplayName           string    `json:"display_name" example:"suspected_user"`
	CreatedAt             string    `json:"created_at" example:"2024-01-01T10:00:00Z"`
	PostsInSubforum       *int      `json:"posts_in_subforum" example:"5"`
	CommentsInSubforum    *int      `json:"comments_in_subforum" example:"12"`
	EncryptedRealIdentity *string   `json:"encrypted_real_identity" example:"encrypted_data_here"`
	TotalPosts            *int      `json:"total_posts" example:"45"`
	TotalComments         *int      `json:"total_comments" example:"230"`
	SubforumsActive       *[]string `json:"subforums_active" example:"golang,programming,tech"`
}

// Correlation represents a correlation request
type Correlation struct {
	CorrelationID      string `json:"correlation_id" example:"uuid_here"`
	CorrelationType    string `json:"correlation_type" example:"fingerprint"`
	RequestedPseudonym string `json:"requested_pseudonym" example:"abc123def456..."`
	Justification      string `json:"justification" example:"Investigation of ban evasion"`
	Status             string `json:"status" example:"completed"`
	Timestamp          string `json:"timestamp" example:"2024-01-01T16:00:00Z"`
	ResultsCount       int    `json:"results_count" example:"2"`
}

// FingerprintCorrelationResponseBody represents the body of fingerprint correlation response
type FingerprintCorrelationResponseBody struct {
	CorrelationID   string              `json:"correlation_id" example:"uuid_here"`
	CorrelationType string              `json:"correlation_type" example:"fingerprint"`
	Scope           string              `json:"scope" example:"subforum_specific"`
	TimeWindow      string              `json:"time_window" example:"30_days"`
	Status          string              `json:"status" example:"completed"`
	Results         []CorrelationResult `json:"results"`
	AuditID         string              `json:"audit_id" example:"audit_uuid_here"`
}

// IdentityCorrelationResponseBody represents the body of identity correlation response
type IdentityCorrelationResponseBody struct {
	CorrelationID   string              `json:"correlation_id" example:"uuid_here"`
	CorrelationType string              `json:"correlation_type" example:"identity"`
	Scope           string              `json:"scope" example:"platform_wide"`
	TimeWindow      string              `json:"time_window" example:"unlimited"`
	Status          string              `json:"status" example:"completed"`
	Results         []CorrelationResult `json:"results"`
	AuditID         string              `json:"audit_id" example:"audit_uuid_here"`
}

// CorrelationHistoryResponseBody represents the body of correlation history response
type CorrelationHistoryResponseBody struct {
	Correlations []Correlation `json:"correlations"`
	Pagination   Pagination    `json:"pagination"`
}

// FingerprintCorrelationResponse represents fingerprint correlation response
type FingerprintCorrelationResponse struct {
	Status int                                `json:"-" example:"200"`
	Body   FingerprintCorrelationResponseBody `json:"body"`
}

// IdentityCorrelationResponse represents identity correlation response
type IdentityCorrelationResponse struct {
	Status int                             `json:"-" example:"200"`
	Body   IdentityCorrelationResponseBody `json:"body"`
}

// CorrelationHistoryResponse represents correlation history response
type CorrelationHistoryResponse struct {
	Status int                            `json:"-" example:"200"`
	Body   CorrelationHistoryResponseBody `json:"body"`
}

// NewFingerprintCorrelationResponse creates a new fingerprint correlation response
func NewFingerprintCorrelationResponse(correlationID string, results []CorrelationResult, auditID string) *FingerprintCorrelationResponse {
	return &FingerprintCorrelationResponse{
		Status: 200,
		Body: FingerprintCorrelationResponseBody{
			CorrelationID:   correlationID,
			CorrelationType: "fingerprint",
			Scope:           "subforum_specific",
			TimeWindow:      "30_days",
			Status:          "completed",
			Results:         results,
			AuditID:         auditID,
		},
	}
}

// NewIdentityCorrelationResponse creates a new identity correlation response
func NewIdentityCorrelationResponse(correlationID string, results []CorrelationResult, auditID string) *IdentityCorrelationResponse {
	return &IdentityCorrelationResponse{
		Status: 200,
		Body: IdentityCorrelationResponseBody{
			CorrelationID:   correlationID,
			CorrelationType: "identity",
			Scope:           "platform_wide",
			TimeWindow:      "unlimited",
			Status:          "completed",
			Results:         results,
			AuditID:         auditID,
		},
	}
}

// NewCorrelationHistoryResponse creates a new correlation history response
func NewCorrelationHistoryResponse(correlations []Correlation, page, limit, total int) *CorrelationHistoryResponse {
	pages := (total + limit - 1) / limit // Ceiling division

	return &CorrelationHistoryResponse{
		Status: 200,
		Body: CorrelationHistoryResponseBody{
			Correlations: correlations,
			Pagination: Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: pages,
			},
		},
	}
}
