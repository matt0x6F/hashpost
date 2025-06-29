package models

import "time"

// Path parameter models for routes
type PseudonymIDPathParam struct {
	PseudonymID string `path:"pseudonym_id" example:"abc123def456..." required:"true"`
}

// PseudonymProfileBody represents the body of pseudonym profile update request
type PseudonymProfileBody struct {
	DisplayName         string `json:"display_name" example:"new_display_name"`
	Bio                 string `json:"bio" example:"Updated bio text"`
	WebsiteURL          string `json:"website_url" example:"https://newwebsite.com"`
	ShowKarma           *bool  `json:"show_karma" example:"false"`
	AllowDirectMessages *bool  `json:"allow_direct_messages" example:"false"`
}

// PseudonymProfileInput represents pseudonym profile update request
// For Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type PseudonymProfileInput struct {
	Body PseudonymProfileBody
}

// UserPreferencesBody represents the body of user preferences update request
type UserPreferencesBody struct {
	Timezone           string `json:"timezone" example:"America/New_York"`
	Language           string `json:"language" example:"en"`
	Theme              string `json:"theme" example:"dark"`
	EmailNotifications *bool  `json:"email_notifications" example:"false"`
	PushNotifications  *bool  `json:"push_notifications" example:"true"`
	AutoHideNSFW       *bool  `json:"auto_hide_nsfw" example:"false"`
	AutoHideSpoilers   *bool  `json:"auto_hide_spoilers" example:"true"`
}

// UserPreferencesInput represents user preferences update request
// For Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type UserPreferencesInput struct {
	Body UserPreferencesBody
}

// CreatePseudonymBody represents the body of pseudonym creation request
type CreatePseudonymBody struct {
	DisplayName         string `json:"display_name" required:"true"`
	Bio                 string `json:"bio"`
	WebsiteURL          string `json:"website_url"`
	ShowKarma           *bool  `json:"show_karma"`
	AllowDirectMessages *bool  `json:"allow_direct_messages"`
}

// CreatePseudonymInput is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type CreatePseudonymInput struct {
	Body CreatePseudonymBody
}

// BlockUserBody represents the body of user block request
type BlockUserBody struct {
	BlockAllPersonas *bool `json:"block_all_personas" example:"true"`
}

// BlockUserInput represents user block request
// For Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type BlockUserInput struct {
	Body BlockUserBody
}

// PseudonymProfile represents pseudonym profile information
type PseudonymProfile struct {
	PseudonymID         string `json:"pseudonym_id" example:"abc123def456..."`
	DisplayName         string `json:"display_name" example:"user_display_name"`
	KarmaScore          int    `json:"karma_score" example:"1250"`
	CreatedAt           string `json:"created_at" example:"2024-01-01T12:00:00Z"`
	LastActiveAt        string `json:"last_active_at" example:"2024-01-01T18:00:00Z"`
	IsActive            bool   `json:"is_active" example:"true"`
	Bio                 string `json:"bio" example:"User bio text"`
	WebsiteURL          string `json:"website_url" example:"https://example.com"`
	ShowKarma           bool   `json:"show_karma" example:"true"`
	AllowDirectMessages bool   `json:"allow_direct_messages" example:"true"`
	PostCount           int    `json:"post_count" example:"45"`
	CommentCount        int    `json:"comment_count" example:"230"`
}

// UserProfile represents user profile with multiple pseudonyms
type UserProfile struct {
	UserID       int                `json:"user_id" example:"123"`
	Email        string             `json:"email" example:"user@example.com"`
	CreatedAt    string             `json:"created_at" example:"2024-01-01T12:00:00Z"`
	LastActiveAt string             `json:"last_active_at" example:"2024-01-01T18:00:00Z"`
	IsActive     bool               `json:"is_active" example:"true"`
	IsSuspended  bool               `json:"is_suspended" example:"false"`
	Roles        []string           `json:"roles" example:"user"`
	Capabilities []string           `json:"capabilities" example:"create_content,vote,message,report"`
	Pseudonyms   []PseudonymProfile `json:"pseudonyms"`
}

// UserPreferences represents user preferences
type UserPreferences struct {
	Timezone           string `json:"timezone" example:"UTC"`
	Language           string `json:"language" example:"en"`
	Theme              string `json:"theme" example:"light"`
	EmailNotifications bool   `json:"email_notifications" example:"true"`
	PushNotifications  bool   `json:"push_notifications" example:"true"`
	AutoHideNSFW       bool   `json:"auto_hide_nsfw" example:"true"`
	AutoHideSpoilers   bool   `json:"auto_hide_spoilers" example:"true"`
}

// PseudonymProfileResponseBody represents the body of pseudonym profile response
type PseudonymProfileResponseBody struct {
	PseudonymProfile
	UpdatedAt string `json:"updated_at" example:"2024-01-01T13:00:00Z"`
}

// UserProfileResponseBody represents the body of user profile response
type UserProfileResponseBody struct {
	UserProfile
	UpdatedAt string `json:"updated_at" example:"2024-01-01T13:00:00Z"`
}

// UserPreferencesResponse represents user preferences response
type UserPreferencesResponse struct {
	Status int             `json:"-" example:"200"`
	Body   UserPreferences `json:"body"`
}

// CreatePseudonymResponseBody represents the body of pseudonym creation response
type CreatePseudonymResponseBody struct {
	PseudonymProfile
	CreatedAt string `json:"created_at" example:"2024-01-01T12:00:00Z"`
}

// BlockUserResponseBody represents the body of user block response
type BlockUserResponseBody struct {
	BlockedPseudonymID     string `json:"blocked_pseudonym_id" example:"def789ghi012..."`
	BlockedUserFingerprint string `json:"blocked_user_fingerprint" example:"a1b2c3d4e5f6..."`
	BlockedAt              string `json:"blocked_at" example:"2024-01-01T18:00:00Z"`
}

// UnblockUserResponseBody represents the body of user unblock response
type UnblockUserResponseBody struct {
	BlockedUserID      int    `json:"blocked_user_id" example:"456"`
	BlockedPseudonymID string `json:"blocked_pseudonym_id" example:"def789ghi012..."`
	UnblockedAt        string `json:"unblocked_at" example:"2024-01-01T19:00:00Z"`
}

// PseudonymProfileResponse represents pseudonym profile response
type PseudonymProfileResponse struct {
	Status int                          `json:"-" example:"200"`
	Body   PseudonymProfileResponseBody `json:"body"`
}

// UserProfileResponse represents user profile response
type UserProfileResponse struct {
	Status int                     `json:"-" example:"200"`
	Body   UserProfileResponseBody `json:"body"`
}

// CreatePseudonymResponse represents pseudonym creation response
type CreatePseudonymResponse struct {
	Status int                         `json:"-" example:"200"`
	Body   CreatePseudonymResponseBody `json:"body"`
}

// BlockUserResponse represents user block response
type BlockUserResponse struct {
	Status int                   `json:"-" example:"200"`
	Body   BlockUserResponseBody `json:"body"`
}

// UnblockUserResponse represents user unblock response
type UnblockUserResponse struct {
	Status int                     `json:"-" example:"200"`
	Body   UnblockUserResponseBody `json:"body"`
}

// NewPseudonymProfileResponse creates a new pseudonym profile response
func NewPseudonymProfileResponse(pseudonymID, displayName, bio, websiteURL string, karmaScore, postCount, commentCount int, showKarma, allowDirectMessages bool, createdAt, lastActiveAt string) *PseudonymProfileResponse {
	return &PseudonymProfileResponse{
		Status: 200,
		Body: PseudonymProfileResponseBody{
			PseudonymProfile: PseudonymProfile{
				PseudonymID:         pseudonymID,
				DisplayName:         displayName,
				KarmaScore:          karmaScore,
				CreatedAt:           createdAt,
				LastActiveAt:        lastActiveAt,
				IsActive:            true,
				Bio:                 bio,
				WebsiteURL:          websiteURL,
				ShowKarma:           showKarma,
				AllowDirectMessages: allowDirectMessages,
				PostCount:           postCount,
				CommentCount:        commentCount,
			},
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// NewUserProfileResponse creates a new user profile response
func NewUserProfileResponse(userID int, email string, roles, capabilities []string, pseudonyms []PseudonymProfile) *UserProfileResponse {
	return &UserProfileResponse{
		Status: 200,
		Body: UserProfileResponseBody{
			UserProfile: UserProfile{
				UserID:       userID,
				Email:        email,
				CreatedAt:    time.Now().UTC().Format(time.RFC3339),
				LastActiveAt: time.Now().UTC().Format(time.RFC3339),
				IsActive:     true,
				IsSuspended:  false,
				Roles:        roles,
				Capabilities: capabilities,
				Pseudonyms:   pseudonyms,
			},
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// NewUserPreferencesResponse creates a new user preferences response
func NewUserPreferencesResponse(timezone, language, theme string, emailNotifications, pushNotifications, autoHideNSFW, autoHideSpoilers bool) *UserPreferencesResponse {
	return &UserPreferencesResponse{
		Status: 200,
		Body: UserPreferences{
			Timezone:           timezone,
			Language:           language,
			Theme:              theme,
			EmailNotifications: emailNotifications,
			PushNotifications:  pushNotifications,
			AutoHideNSFW:       autoHideNSFW,
			AutoHideSpoilers:   autoHideSpoilers,
		},
	}
}

// NewCreatePseudonymResponse creates a new pseudonym creation response
func NewCreatePseudonymResponse(pseudonymID, displayName, bio, websiteURL string, showKarma, allowDirectMessages bool) *CreatePseudonymResponse {
	now := time.Now().UTC().Format(time.RFC3339)
	return &CreatePseudonymResponse{
		Status: 200,
		Body: CreatePseudonymResponseBody{
			PseudonymProfile: PseudonymProfile{
				PseudonymID:         pseudonymID,
				DisplayName:         displayName,
				KarmaScore:          0,
				CreatedAt:           now,
				LastActiveAt:        now,
				IsActive:            true,
				Bio:                 bio,
				WebsiteURL:          websiteURL,
				ShowKarma:           showKarma,
				AllowDirectMessages: allowDirectMessages,
				PostCount:           0,
				CommentCount:        0,
			},
			CreatedAt: now,
		},
	}
}

// NewBlockUserResponse creates a new user block response
func NewBlockUserResponse(blockedPseudonymID, blockedUserFingerprint string) *BlockUserResponse {
	return &BlockUserResponse{
		Status: 200,
		Body: BlockUserResponseBody{
			BlockedPseudonymID:     blockedPseudonymID,
			BlockedUserFingerprint: blockedUserFingerprint,
			BlockedAt:              time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// NewUnblockUserResponse creates a new user unblock response
func NewUnblockUserResponse(blockedUserID int, blockedPseudonymID string) *UnblockUserResponse {
	return &UnblockUserResponse{
		Status: 200,
		Body: UnblockUserResponseBody{
			BlockedUserID:      blockedUserID,
			BlockedPseudonymID: blockedPseudonymID,
			UnblockedAt:        time.Now().UTC().Format(time.RFC3339),
		},
	}
}
