package models

import (
	"time"
)

// Subforum represents subforum information
type Subforum struct {
	Name            string    `json:"name" example:"golang"`
	DisplayName     string    `json:"display_name" example:"Golang"`
	Description     string    `json:"description" example:"The Go programming language"`
	SidebarText     string    `json:"sidebar_text" example:"Welcome to r/golang..."`
	RulesText       string    `json:"rules_text" example:"1. Be respectful..."`
	IsNSFW          bool      `json:"is_nsfw" example:"false"`
	IsPrivate       bool      `json:"is_private" example:"false"`
	IsRestricted    bool      `json:"is_restricted" example:"false"`
	SubscriberCount int       `json:"subscriber_count" example:"1234"`
	PostCount       int       `json:"post_count" example:"5678"`
	CreatedAt       time.Time `json:"created_at" example:"2023-01-01T00:00:00Z"`
	UpdatedAt       time.Time `json:"updated_at" example:"2023-01-01T00:00:00Z"`
	// Community type system fields
	CommunityType    string `json:"community_type" example:"t"`
	GovernanceStyle  string `json:"governance_style" example:"democratic"`
	OwnerPseudonymID string `json:"owner_pseudonym_id,omitempty" example:"abc123"`
}

// SubforumModerator represents a subforum moderator
type SubforumModerator struct {
	PseudonymID   string `json:"pseudonym_id" example:"abc123"`
	DisplayName   string `json:"display_name" example:"moderator1"`
	ModeratorType string `json:"moderator_type" example:"admin"`
	AddedAt       string `json:"added_at" example:"2023-01-01T00:00:00Z"`
}

// SubforumDetails represents detailed subforum information
type SubforumDetails struct {
	Subforum
	Moderators   []SubforumModerator `json:"moderators"`
	IsSubscribed bool                `json:"is_subscribed"`
	IsFavorite   bool                `json:"is_favorite"`
}

// SubforumListInput represents subforum list request parameters
type SubforumListInput struct {
	Page  int    `query:"page" example:"1"`
	Limit int    `query:"limit" example:"25"`
	Sort  string `query:"sort" example:"subscribers"`
}

// SubforumSubscriptionInput represents subforum subscription request
type SubforumSubscriptionInput struct {
	CommunityType string `path:"type" example:"t"`
	SubforumName  string `path:"name" example:"programming"`
}

// SubforumSubscriptionResponse represents subforum subscription response
type SubforumSubscriptionResponse struct {
	Status int `json:"status" example:"200"`
	Body   struct {
		SubforumID      int    `json:"subforum_id" example:"1"`
		SubforumName    string `json:"subforum_name" example:"golang"`
		IsSubscribed    bool   `json:"is_subscribed" example:"true"`
		SubscriberCount int    `json:"subscriber_count" example:"1234"`
	} `json:"body"`
}

// SubforumsListResponse represents a list of subforums
type SubforumsListResponse struct {
	Status int `json:"status" example:"200"`
	Body   struct {
		Subforums  []Subforum `json:"subforums"`
		Pagination Pagination `json:"pagination"`
	} `json:"body"`
}

// SubforumDetailsResponse represents detailed subforum information
type SubforumDetailsResponse struct {
	Status int `json:"status" example:"200"`
	Body   struct {
		Subforum     Subforum            `json:"subforum"`
		Moderators   []SubforumModerator `json:"moderators"`
		IsSubscribed bool                `json:"is_subscribed"`
		IsFavorite   bool                `json:"is_favorite"`
	} `json:"body"`
}

// SubforumCreateBody represents the body for creating a new subforum
type SubforumCreateBody struct {
	Slug         string `json:"slug" example:"golang" required:"true"`
	Name         string `json:"name" example:"Golang" required:"true"`
	Description  string `json:"description" example:"The Go programming language" required:"true"`
	SidebarText  string `json:"sidebar_text,omitempty" example:"Welcome to r/golang..."`
	RulesText    string `json:"rules_text,omitempty" example:"1. Be respectful..."`
	IsNSFW       bool   `json:"is_nsfw,omitempty" example:"false"`
	IsPrivate    bool   `json:"is_private,omitempty" example:"false"`
	IsRestricted bool   `json:"is_restricted,omitempty" example:"false"`
	// Community type system fields
	CommunityType string `json:"community_type" example:"t" required:"true"`
	// Governance style is automatically determined based on community type:
	// - t/ and g/ communities use "democratic" governance
	// - b/ and c/ communities use "owned" governance
}

// SubforumCreateInput represents the input for creating a new subforum
// Includes authentication headers and a Body field for the JSON body
type SubforumCreateInput struct {
	Authorization string             `header:"Authorization" doc:"Bearer token for API authentication"`
	AccessToken   string             `cookie:"access_token" doc:"JWT access token from cookie"`
	Body          SubforumCreateBody // JSON body
}

// NewSubforumListResponse creates a new subforum list response
func NewSubforumListResponse(subforums []Subforum, page, limit, total int) *SubforumsListResponse {
	pages := (total + limit - 1) / limit // Ceiling division

	return &SubforumsListResponse{
		Status: 200,
		Body: struct {
			Subforums  []Subforum `json:"subforums"`
			Pagination Pagination `json:"pagination"`
		}{
			Subforums: subforums,
			Pagination: Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: pages,
			},
		},
	}
}

// NewSubforumDetailsResponse creates a new SubforumDetailsResponse
func NewSubforumDetailsResponse(subforum Subforum, moderators []SubforumModerator, isSubscribed, isFavorite bool) *SubforumDetailsResponse {
	return &SubforumDetailsResponse{
		Status: 200,
		Body: struct {
			Subforum     Subforum            `json:"subforum"`
			Moderators   []SubforumModerator `json:"moderators"`
			IsSubscribed bool                `json:"is_subscribed"`
			IsFavorite   bool                `json:"is_favorite"`
		}{
			Subforum:     subforum,
			Moderators:   moderators,
			IsSubscribed: isSubscribed,
			IsFavorite:   isFavorite,
		},
	}
}

// NewSubforumSubscriptionResponse creates a new subforum subscription response
func NewSubforumSubscriptionResponse(subforumID int, subforumName string, isSubscribed bool, subscriberCount int) *SubforumSubscriptionResponse {
	return &SubforumSubscriptionResponse{
		Status: 200,
		Body: struct {
			SubforumID      int    `json:"subforum_id" example:"1"`
			SubforumName    string `json:"subforum_name" example:"golang"`
			IsSubscribed    bool   `json:"is_subscribed" example:"true"`
			SubscriberCount int    `json:"subscriber_count" example:"1234"`
		}{
			SubforumID:      subforumID,
			SubforumName:    subforumName,
			IsSubscribed:    isSubscribed,
			SubscriberCount: subscriberCount,
		},
	}
}

// SubforumSubscriptionsResponseBody represents the body for pseudonym subforum subscriptions
// Returns a list of subforums the pseudonym is subscribed to
type SubforumSubscriptionsResponseBody struct {
	Subforums []Subforum `json:"subforums"`
}

// SubforumSubscriptionsResponse represents the response for pseudonym subforum subscriptions
type SubforumSubscriptionsResponse struct {
	Status int                               `json:"-" example:"200"`
	Body   SubforumSubscriptionsResponseBody `json:"body"`
}

// PseudonymSubscriptionsInput represents the input for pseudonym subscriptions endpoint
type PseudonymSubscriptionsInput struct {
	PseudonymID string `path:"pseudonym_id" example:"abc123"`
}
