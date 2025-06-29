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
	// Empty for now, but could include subscription preferences
	SubforumName string `path:"name" example:"golang"`
}

// SubforumsListResponseBody represents the body of subforums list response
type SubforumsListResponseBody struct {
	Subforums  []Subforum `json:"subforums"`
	Pagination Pagination `json:"pagination"`
}

// SubforumSubscriptionResponseBody represents the body of subforum subscription response
type SubforumSubscriptionResponseBody struct {
	SubforumID      int    `json:"subforum_id" example:"1"`
	Name            string `json:"name" example:"golang"`
	Subscribed      bool   `json:"subscribed" example:"true"`
	SubscriberCount int    `json:"subscriber_count" example:"125001"`
}

// SubforumsListResponse represents subforums list response
type SubforumsListResponse struct {
	Status int                       `json:"-" example:"200"`
	Body   SubforumsListResponseBody `json:"body"`
}

// SubforumDetailsResponseBody represents the body of subforum details response
type SubforumDetailsResponseBody struct {
	Subforum
	Moderators   []SubforumModerator `json:"moderators"`
	IsSubscribed bool                `json:"is_subscribed"`
	IsFavorite   bool                `json:"is_favorite"`
}

// SubforumDetailsResponse represents the response for subforum details
type SubforumDetailsResponse struct {
	Status int                         `json:"-" example:"200"`
	Body   SubforumDetailsResponseBody `json:"body"`
}

// SubforumSubscriptionResponse represents subforum subscription response
type SubforumSubscriptionResponse struct {
	Status int                              `json:"-" example:"200"`
	Body   SubforumSubscriptionResponseBody `json:"body"`
}

// SubforumCreateBody represents the JSON body for creating a new subforum
// slug, name, and description are required; others are optional
// slug is the unique identifier (e.g., "golang")
type SubforumCreateBody struct {
	Slug         string `json:"slug" example:"golang" required:"true"`
	Name         string `json:"name" example:"Golang" required:"true"`
	Description  string `json:"description" example:"The Go programming language" required:"true"`
	SidebarText  string `json:"sidebar_text,omitempty" example:"Welcome to r/golang..."`
	RulesText    string `json:"rules_text,omitempty" example:"1. Be respectful..."`
	IsNSFW       bool   `json:"is_nsfw,omitempty" example:"false"`
	IsPrivate    bool   `json:"is_private,omitempty" example:"false"`
	IsRestricted bool   `json:"is_restricted,omitempty" example:"false"`
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
		Body: SubforumsListResponseBody{
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
		Body: SubforumDetailsResponseBody{
			Subforum:     subforum,
			Moderators:   moderators,
			IsSubscribed: isSubscribed,
			IsFavorite:   isFavorite,
		},
	}
}

// NewSubforumSubscriptionResponse creates a new subforum subscription response
func NewSubforumSubscriptionResponse(subforumID int, name string, subscribed bool, subscriberCount int) *SubforumSubscriptionResponse {
	return &SubforumSubscriptionResponse{
		Status: 200,
		Body: SubforumSubscriptionResponseBody{
			SubforumID:      subforumID,
			Name:            name,
			Subscribed:      subscribed,
			SubscriberCount: subscriberCount,
		},
	}
}
