package models

import (
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
)

// Post represents a post
type Post struct {
	PostID       int    `json:"post_id" example:"123"`
	Title        string `json:"title" example:"Post Title"`
	Content      string `json:"content" example:"Post content..."`
	PostType     string `json:"post_type" example:"text"`
	URL          string `json:"url" example:"https://example.com"`
	IsSelfPost   bool   `json:"is_self_post" example:"true"`
	IsNSFW       bool   `json:"is_nsfw" example:"false"`
	IsSpoiler    bool   `json:"is_spoiler" example:"false"`
	Score        int    `json:"score" example:"1250"`
	Upvotes      int    `json:"upvotes" example:"1300"`
	Downvotes    int    `json:"downvotes" example:"50"`
	CommentCount int    `json:"comment_count" example:"45"`
	ViewCount    int    `json:"view_count" example:"5000"`
	CreatedAt    string `json:"created_at" example:"2024-01-01T12:00:00Z"`
	Author       struct {
		PseudonymID string `json:"pseudonym_id" example:"abc123def456..."`
		DisplayName string `json:"display_name" example:"user_display_name"`
	} `json:"author"`
	Subforum struct {
		SubforumID  int    `json:"subforum_id" example:"1"`
		Name        string `json:"name" example:"golang"`
		DisplayName string `json:"display_name" example:"Golang"`
	} `json:"subforum"`
	UserVote int  `json:"user_vote" example:"1"` // 1 for upvote, -1 for downvote, 0 for no vote
	IsSaved  bool `json:"is_saved" example:"false"`
}

// Comment represents a comment
type Comment struct {
	CommentID       int    `json:"comment_id" example:"456"`
	Content         string `json:"content" example:"Comment text..."`
	ParentCommentID *int   `json:"parent_comment_id" example:"123"`
	Score           int    `json:"score" example:"25"`
	CreatedAt       string `json:"created_at" example:"2024-01-01T12:30:00Z"`
	Author          struct {
		PseudonymID string `json:"pseudonym_id" example:"def789ghi012..."`
		DisplayName string `json:"display_name" example:"commenter_name"`
	} `json:"author"`
	UserVote int       `json:"user_vote" example:"0"`
	Replies  []Comment `json:"replies"`
}

// PostInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type PostInputBody struct {
	Title     string `json:"title" example:"Post Title" required:"true"`
	Content   string `json:"content" example:"Post content text..." required:"true"`
	PostType  string `json:"post_type" example:"text" required:"true"`
	URL       string `json:"url" example:"https://example.com"`
	IsNSFW    bool   `json:"is_nsfw" example:"false"`
	IsSpoiler bool   `json:"is_spoiler" example:"false"`
}

// CommentInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type CommentInputBody struct {
	Content         string `json:"content" example:"Comment text..." required:"true"`
	ParentCommentID *int   `json:"parent_comment_id,omitempty" example:"456"`
}

// CommentInput represents comment creation request (for OpenAPI schema only)
type CommentInput struct {
	middleware.AuthInput
	PostID int64 `path:"post_id" example:"123" doc:"Post ID"`
	Body   CommentInputBody
}

// VoteInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type VoteInputBody struct {
	VoteValue int `json:"vote_value" example:"1" required:"true"`
}

// VoteInput represents vote request (for OpenAPI schema only)
type VoteInput struct {
	Body VoteInputBody
}

// PostVoteInput represents post vote request with path parameter (for OpenAPI schema only)
type PostVoteInput struct {
	middleware.AuthInput
	PostID int64 `path:"post_id" example:"123" doc:"Post ID"`
	Body   VoteInputBody
}

// CommentVoteInput represents comment vote request with path parameter (for OpenAPI schema only)
type CommentVoteInput struct {
	middleware.AuthInput
	CommentID int64 `path:"comment_id" example:"789" doc:"Comment ID"`
	Body      VoteInputBody
}

// Post sort options
const (
	PostSortNew      = "new"
	PostSortTop      = "top"
	PostSortOld      = "old"
	PostSortComments = "comments"
	PostSortViews    = "views"
)

// PostListInput represents post list request parameters
// Sort can be one of: "new", "top", "old", "comments", "views"
// Time can be one of: "hour", "day", "week", "month", "year", "all"
type PostListInput struct {
	SubforumName string `path:"name" example:"golang" doc:"Subforum name"`
	Page         int    `query:"page" example:"1"`
	Limit        int    `query:"limit" example:"25"`
	Sort         string `query:"sort" example:"new"` // Allowed: "new", "top", "old", "comments", "views"
	Time         string `query:"time" example:"day"` // Allowed: "hour", "day", "week", "month", "year", "all"
}

// PostDetailsInput represents post details request parameters
type PostDetailsInput struct {
	PostID int64  `path:"post_id" example:"123" doc:"Post ID"`
	Sort   string `query:"sort" example:"best"` // "best", "top", "new", "controversial", "old", "qa"
}

// PostListResponseBody represents the body of post list response
type PostListResponseBody struct {
	Posts      []Post     `json:"posts"`
	Pagination Pagination `json:"pagination"`
}

// PostDetailsResponseBody represents the body of post details response
type PostDetailsResponseBody struct {
	Post
	Comments []Comment `json:"comments"`
}

// PostResponseBody represents the body of post creation response
type PostResponseBody struct {
	PostID       int    `json:"post_id" example:"124"`
	Title        string `json:"title" example:"Post Title"`
	Content      string `json:"content" example:"Post content text..."`
	PostType     string `json:"post_type" example:"text"`
	Score        int    `json:"score" example:"0"`
	CommentCount int    `json:"comment_count" example:"0"`
	CreatedAt    string `json:"created_at" example:"2024-01-01T14:00:00Z"`
	Author       struct {
		PseudonymID string `json:"pseudonym_id" example:"abc123def456..."`
		DisplayName string `json:"display_name" example:"user_display_name"`
	} `json:"author"`
}

// CommentResponseBody represents the body of comment creation response
type CommentResponseBody struct {
	CommentID       int    `json:"comment_id" example:"789"`
	Content         string `json:"content" example:"Comment text..."`
	ParentCommentID *int   `json:"parent_comment_id" example:"456"`
	Score           int    `json:"score" example:"0"`
	CreatedAt       string `json:"created_at" example:"2024-01-01T15:00:00Z"`
	Author          struct {
		PseudonymID string `json:"pseudonym_id" example:"abc123def456..."`
		DisplayName string `json:"display_name" example:"user_display_name"`
	} `json:"author"`
}

// VoteResponseBody represents the body of vote response
type VoteResponseBody struct {
	PostID    int `json:"post_id" example:"123"`
	VoteValue int `json:"vote_value" example:"1"`
	Score     int `json:"score" example:"1251"`
	Upvotes   int `json:"upvotes" example:"1301"`
	Downvotes int `json:"downvotes" example:"50"`
}

// CommentVoteResponseBody represents the body of comment vote response
type CommentVoteResponseBody struct {
	CommentID int `json:"comment_id" example:"789"`
	VoteValue int `json:"vote_value" example:"1"`
	Score     int `json:"score" example:"1"`
	Upvotes   int `json:"upvotes" example:"1"`
	Downvotes int `json:"downvotes" example:"0"`
}

// PostListResponse represents post list response
type PostListResponse struct {
	Status int `json:"-" example:"200"`
	Body   PostListResponseBody
}

// PostDetailsResponse represents post details response
type PostDetailsResponse struct {
	Status int `json:"-" example:"200"`
	Body   PostDetailsResponseBody
}

// PostResponse represents post creation response
type PostResponse struct {
	Status int `json:"-" example:"200"`
	Body   PostResponseBody
}

// CommentResponse represents comment creation response
type CommentResponse struct {
	Status int `json:"-" example:"200"`
	Body   CommentResponseBody
}

// VoteResponse represents vote response
type VoteResponse struct {
	Status int `json:"-" example:"200"`
	Body   VoteResponseBody
}

// CommentVoteResponse represents comment vote response
type CommentVoteResponse struct {
	Status int `json:"-" example:"200"`
	Body   CommentVoteResponseBody
}

// PostsListResponse represents posts list response
type PostsListResponse struct {
	Status int `json:"-" example:"200"`
	Body   PostListResponseBody
}

// NewPostListResponse creates a new post list response
func NewPostListResponse(posts []Post, page, limit, total int) *PostListResponse {
	// Set default limit if zero to prevent division by zero
	if limit <= 0 {
		limit = 25
	}

	pages := (total + limit - 1) / limit // Ceiling division

	return &PostListResponse{
		Status: 200,
		Body: PostListResponseBody{
			Posts: posts,
			Pagination: Pagination{
				Page:  page,
				Limit: limit,
				Total: total,
				Pages: pages,
			},
		},
	}
}

// NewPostDetailsResponse creates a new post details response
func NewPostDetailsResponse(post Post, comments []Comment) *PostDetailsResponse {
	return &PostDetailsResponse{
		Status: 200,
		Body: PostDetailsResponseBody{
			Post:     post,
			Comments: comments,
		},
	}
}

// NewPostResponse creates a new post creation response
func NewPostResponse(postID int, title, content, postType, pseudonymID, displayName string) *PostResponse {
	return &PostResponse{
		Status: 200,
		Body: PostResponseBody{
			PostID:       postID,
			Title:        title,
			Content:      content,
			PostType:     postType,
			Score:        0,
			CommentCount: 0,
			CreatedAt:    time.Now().UTC().Format(time.RFC3339),
			Author: struct {
				PseudonymID string `json:"pseudonym_id" example:"abc123def456..."`
				DisplayName string `json:"display_name" example:"user_display_name"`
			}{
				PseudonymID: pseudonymID,
				DisplayName: displayName,
			},
		},
	}
}

// NewCommentResponse creates a new comment creation response
func NewCommentResponse(commentID int, content string, parentCommentID *int, pseudonymID, displayName string) *CommentResponse {
	return &CommentResponse{
		Status: 200,
		Body: CommentResponseBody{
			CommentID:       commentID,
			Content:         content,
			ParentCommentID: parentCommentID,
			Score:           0,
			CreatedAt:       time.Now().UTC().Format(time.RFC3339),
			Author: struct {
				PseudonymID string `json:"pseudonym_id" example:"abc123def456..."`
				DisplayName string `json:"display_name" example:"user_display_name"`
			}{
				PseudonymID: pseudonymID,
				DisplayName: displayName,
			},
		},
	}
}

// NewVoteResponse creates a new vote response
func NewVoteResponse(postID, voteValue, score, upvotes, downvotes int) *VoteResponse {
	return &VoteResponse{
		Status: 200,
		Body: VoteResponseBody{
			PostID:    postID,
			VoteValue: voteValue,
			Score:     score,
			Upvotes:   upvotes,
			Downvotes: downvotes,
		},
	}
}

// NewCommentVoteResponse creates a new comment vote response
func NewCommentVoteResponse(commentID, voteValue, score, upvotes, downvotes int) *CommentVoteResponse {
	return &CommentVoteResponse{
		Status: 200,
		Body: CommentVoteResponseBody{
			CommentID: commentID,
			VoteValue: voteValue,
			Score:     score,
			Upvotes:   upvotes,
			Downvotes: downvotes,
		},
	}
}

// PostCreateInput represents the input for creating a post
type PostCreateInput struct {
	middleware.AuthInput
	SubforumName string `path:"name" example:"golang" doc:"Subforum name"`
	Body         PostCreateBody
}

// PostCreateBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type PostCreateBody struct {
	Title     string `json:"title" example:"Post Title" required:"true"`
	Content   string `json:"content" example:"Post content text..." required:"true"`
	PostType  string `json:"post_type" example:"text" required:"true"`
	URL       string `json:"url,omitempty" example:"https://example.com"`
	IsNSFW    bool   `json:"is_nsfw,omitempty" example:"false"`
	IsSpoiler bool   `json:"is_spoiler,omitempty" example:"false"`
}
