package models

import (
	"time"

	"github.com/matt0x6f/hashpost/internal/api/middleware"
)

// Post represents a post
type Post struct {
	PostID       int    `json:"post_id" example:"123"`
	Slug         string `json:"slug" example:"my-post-title-123"`
	Title        string `json:"title" example:"Post Title"`
	Content      string `json:"content" example:"Post content..."`
	PostType     string `json:"post_type" example:"text"`
	URL          string `json:"url" example:"https://example.com"`
	IsSelfPost   bool   `json:"is_self_post" example:"true"`
	IsNSFW       bool   `json:"is_nsfw" example:"false"`
	IsSpoiler    bool   `json:"is_spoiler" example:"false"`
	IsLocked     bool   `json:"is_locked" example:"false"`
	IsSticky     bool   `json:"is_sticky" example:"false"`
	IsRemoved    bool   `json:"is_removed" example:"false"`
	Score        int    `json:"score" example:"1250"`
	Upvotes      int    `json:"upvotes" example:"1300"`
	Downvotes    int    `json:"downvotes" example:"50"`
	CommentCount int    `json:"comment_count" example:"25"`
	CreatedAt    string `json:"created_at" example:"2024-01-01T12:00:00Z"`
	Author       struct {
		PseudonymID string `json:"pseudonym_id" example:"abc123def456..."`
		DisplayName string `json:"display_name" example:"author_name"`
	} `json:"author"`
	Subforum struct {
		Name          string `json:"name" example:"technology"`
		DisplayName   string `json:"display_name" example:"Technology"`
		CommunityType string `json:"community_type" example:"t"`
	} `json:"subforum"`
	UserVote int       `json:"user_vote" example:"0"`
	Comments []Comment `json:"comments"`
	// User deletion fields
	IsDeleted    bool   `json:"is_deleted" example:"false"`
	DeletedAt    string `json:"deleted_at,omitempty" example:"2024-01-01T16:00:00Z"`
	DeleteReason string `json:"delete_reason,omitempty" example:"User requested deletion"`
	DeletedBy    struct {
		PseudonymID string `json:"pseudonym_id" example:"user_pseudonym_id"`
		DisplayName string `json:"display_name" example:"user_name"`
	} `json:"deleted_by,omitempty"`
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
	// Post information for comments in profile context
	PostTitle           string `json:"post_title,omitempty" example:"My Post Title"`
	PostID              int    `json:"post_id,omitempty" example:"123"`
	SubforumName        string `json:"subforum_name,omitempty" example:"golang"`
	SubforumDisplayName string `json:"subforum_display_name,omitempty" example:"Golang"`
	CommunityType       string `json:"community_type,omitempty" example:"t"`
	// User deletion fields
	IsDeleted    bool   `json:"is_deleted" example:"false"`
	DeletedAt    string `json:"deleted_at,omitempty" example:"2024-01-01T16:00:00Z"`
	DeleteReason string `json:"delete_reason,omitempty" example:"User requested deletion"`
	DeletedBy    struct {
		PseudonymID string `json:"pseudonym_id" example:"user_pseudonym_id"`
		DisplayName string `json:"display_name" example:"user_name"`
	} `json:"deleted_by,omitempty"`
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
	middleware.AuthInput
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

// PostBySlugInput represents post details request parameters by slug
type PostBySlugInput struct {
	middleware.AuthInput
	SubforumName string `path:"subforum" example:"golang" doc:"Subforum name"`
	Slug         string `path:"slug" example:"my-first-post-123" doc:"Post slug"`
	Sort         string `query:"sort" example:"best"` // "best", "top", "new", "controversial", "old", "qa"
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
	Slug         string `json:"slug" example:"my-post-title-124"`
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
func NewPostResponse(postID int, title, content, postType, pseudonymID, displayName, slug string) *PostResponse {
	return &PostResponse{
		Status: 200,
		Body: PostResponseBody{
			PostID:       postID,
			Slug:         slug,
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

// PostCreateBody represents the body of a post creation request
type PostCreateBody struct {
	Title     string  `json:"title" doc:"Post title"`
	Content   string  `json:"content" doc:"Post content"`
	PostType  string  `json:"post_type" doc:"Type of post (text, link, image, etc.)"`
	URL       *string `json:"url,omitempty" doc:"URL for link posts"`
	IsNSFW    bool    `json:"is_nsfw" doc:"Whether the post is NSFW"`
	IsSpoiler bool    `json:"is_spoiler" doc:"Whether the post contains spoilers"`
	IsSticky  bool    `json:"is_sticky" doc:"Whether the post should be sticky (moderator only)"`
	IsLocked  bool    `json:"is_locked" doc:"Whether the post should be locked (moderator only)"`
}

// Lock/Unlock Post
// Input for locking/unlocking a post
// PATCH /posts/{post_id}/lock
// Body: { locked: true|false }
type PostLockInput struct {
	middleware.AuthInput
	PostID int64 `path:"post_id" example:"123" doc:"Post ID"`
	Body   struct {
		Locked bool `json:"locked" example:"true" required:"true"`
	}
}

// Sticky/Unsticky Post
// PATCH /posts/{post_id}/sticky
// Body: { sticky: true|false }
type PostStickyInput struct {
	middleware.AuthInput
	PostID int64 `path:"post_id" example:"123" doc:"Post ID"`
	Body   struct {
		Sticky bool `json:"sticky" example:"true" required:"true"`
	}
}

// Remove/Restore Post
// PATCH /posts/{post_id}/remove
// Body: { removed: true|false }
type PostRemoveInput struct {
	middleware.AuthInput
	PostID int64 `path:"post_id" example:"123" doc:"Post ID"`
	Body   struct {
		Removed bool `json:"removed" example:"true" required:"true"`
	}
}

// PostDeleteInput represents post deletion by user request (for OpenAPI schema only)
type PostDeleteInput struct {
	middleware.AuthInput
	PostID int64 `path:"post_id" example:"123" doc:"Post ID"`
	Body   struct {
		Reason string `json:"reason,omitempty" example:"User requested deletion"`
	}
}

// CommentEditInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type CommentEditInputBody struct {
	Content    string `json:"content" example:"Updated comment text..." required:"true"`
	EditReason string `json:"edit_reason,omitempty" example:"Fixed typo"`
}

// CommentEditInput represents comment editing request (for OpenAPI schema only)
type CommentEditInput struct {
	middleware.AuthInput
	CommentID int64 `path:"comment_id" example:"456" doc:"Comment ID"`
	Body      CommentEditInputBody
}

// CommentRemoveInput represents comment removal request (for OpenAPI schema only)
type CommentRemoveInput struct {
	middleware.AuthInput
	CommentID int64 `path:"comment_id" example:"456" doc:"Comment ID"`
	Body      struct {
		Removed bool   `json:"removed" example:"true" required:"true"`
		Reason  string `json:"reason,omitempty" example:"Violates community guidelines"`
	}
}

// CommentDeleteInput represents comment deletion by user request (for OpenAPI schema only)
type CommentDeleteInput struct {
	middleware.AuthInput
	CommentID int64 `path:"comment_id" example:"456" doc:"Comment ID"`
	Body      struct {
		Reason string `json:"reason,omitempty" example:"User requested deletion"`
	}
}

// CommentReportInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type CommentReportInputBody struct {
	ReportReason  string `json:"report_reason" example:"spam" required:"true"`
	ReportDetails string `json:"report_details,omitempty" example:"This comment violates community guidelines"`
}

// CommentReportInput represents comment reporting request (for OpenAPI schema only)
type CommentReportInput struct {
	middleware.AuthInput
	CommentID int64 `path:"comment_id" example:"456" doc:"Comment ID"`
	Body      CommentReportInputBody
}

// CommentEditResponseBody represents the body of comment editing response
type CommentEditResponseBody struct {
	CommentID       int    `json:"comment_id" example:"456"`
	Content         string `json:"content" example:"Updated comment text..."`
	ParentCommentID *int   `json:"parent_comment_id" example:"123"`
	Score           int    `json:"score" example:"5"`
	CreatedAt       string `json:"created_at" example:"2024-01-01T15:00:00Z"`
	EditedAt        string `json:"edited_at" example:"2024-01-01T16:00:00Z"`
	EditReason      string `json:"edit_reason" example:"Fixed typo"`
	IsEdited        bool   `json:"is_edited" example:"true"`
	Author          struct {
		PseudonymID string `json:"pseudonym_id" example:"abc123def456..."`
		DisplayName string `json:"display_name" example:"user_display_name"`
	} `json:"author"`
}

// CommentEditResponse represents comment editing response
type CommentEditResponse struct {
	Status int `json:"-" example:"200"`
	Body   CommentEditResponseBody
}

// NewCommentEditResponse creates a new comment edit response
func NewCommentEditResponse(commentID int, content string, parentCommentID *int, pseudonymID, displayName, editReason string, isEdited bool) *CommentEditResponse {
	now := time.Now()

	response := &CommentEditResponse{
		Status: 200,
		Body: CommentEditResponseBody{
			CommentID:       commentID,
			Content:         content,
			ParentCommentID: parentCommentID,
			Score:           0,                                                  // Will be updated by vote system
			CreatedAt:       now.Add(-time.Hour).Format("2006-01-02T15:04:05Z"), // Mock creation time
			EditedAt:        now.Format("2006-01-02T15:04:05Z"),
			EditReason:      editReason,
			IsEdited:        isEdited,
		},
	}

	response.Body.Author.PseudonymID = pseudonymID
	response.Body.Author.DisplayName = displayName

	return response
}

// NewCommentRemoveResponse creates a new comment removal response
func NewCommentRemoveResponse(commentID int, removed bool, removalReason, removedByPseudonymID, removedByDisplayName string) *CommentRemoveResponse {
	now := time.Now()

	response := &CommentRemoveResponse{
		Status: 200,
		Body: CommentRemoveResponseBody{
			CommentID:     commentID,
			Removed:       removed,
			RemovalReason: removalReason,
			RemovedAt:     now.Format("2006-01-02T15:04:05Z"),
		},
	}

	response.Body.RemovedBy.PseudonymID = removedByPseudonymID
	response.Body.RemovedBy.DisplayName = removedByDisplayName

	return response
}

// NewCommentReportResponse creates a new comment report response
func NewCommentReportResponse(reportID, commentID int, reportReason, reportDetails, reporterPseudonymID, reporterDisplayName string) *CommentReportResponse {
	now := time.Now()

	response := &CommentReportResponse{
		Status: 200,
		Body: CommentReportResponseBody{
			ReportID:      reportID,
			CommentID:     commentID,
			ReportReason:  reportReason,
			ReportDetails: reportDetails,
			Status:        "pending",
			CreatedAt:     now.Format("2006-01-02T15:04:05Z"),
		},
	}

	response.Body.Reporter.PseudonymID = reporterPseudonymID
	response.Body.Reporter.DisplayName = reporterDisplayName

	return response
}

// CommentRemoveResponseBody represents the body of comment removal response
type CommentRemoveResponseBody struct {
	CommentID     int    `json:"comment_id" example:"456"`
	Removed       bool   `json:"removed" example:"true"`
	RemovalReason string `json:"removal_reason" example:"Violates community guidelines"`
	RemovedAt     string `json:"removed_at" example:"2024-01-01T17:00:00Z"`
	RemovedBy     struct {
		PseudonymID string `json:"pseudonym_id" example:"mod_pseudonym_id"`
		DisplayName string `json:"display_name" example:"moderator_name"`
	} `json:"removed_by"`
}

// CommentRemoveResponse represents comment removal response
type CommentRemoveResponse struct {
	Status int `json:"-" example:"200"`
	Body   CommentRemoveResponseBody
}

// CommentReportResponseBody represents the body of comment reporting response
type CommentReportResponseBody struct {
	ReportID      int    `json:"report_id" example:"789"`
	CommentID     int    `json:"comment_id" example:"456"`
	ReportReason  string `json:"report_reason" example:"spam"`
	ReportDetails string `json:"report_details" example:"This comment violates community guidelines"`
	Status        string `json:"status" example:"pending"`
	CreatedAt     string `json:"created_at" example:"2024-01-01T18:00:00Z"`
	Reporter      struct {
		PseudonymID string `json:"pseudonym_id" example:"reporter_pseudonym_id"`
		DisplayName string `json:"display_name" example:"reporter_name"`
	} `json:"reporter"`
}

// CommentReportResponse represents comment reporting response
type CommentReportResponse struct {
	Status int `json:"-" example:"200"`
	Body   CommentReportResponseBody
}

// PostEditInputBody is for Huma schema definition only. Actual requests should send flat JSON, not nested under 'body'.
type PostEditInputBody struct {
	Title      string `json:"title" example:"Updated Post Title" required:"true"`
	Content    string `json:"content" example:"Updated post content..." required:"true"`
	EditReason string `json:"edit_reason,omitempty" example:"Fixed typo"`
}

// PostEditInput represents post editing request (for OpenAPI schema only)
type PostEditInput struct {
	middleware.AuthInput
	PostID int64 `path:"post_id" example:"123" doc:"Post ID"`
	Body   PostEditInputBody
}

// PostEditResponseBody represents the body of post editing response
type PostEditResponseBody struct {
	PostID     int    `json:"post_id" example:"123"`
	Title      string `json:"title" example:"Updated Post Title"`
	Content    string `json:"content" example:"Updated post content..."`
	EditedAt   string `json:"edited_at" example:"2024-01-01T16:00:00Z"`
	EditReason string `json:"edit_reason" example:"Fixed typo"`
	IsEdited   bool   `json:"is_edited" example:"true"`
	Author     struct {
		PseudonymID string `json:"pseudonym_id" example:"abc123def456..."`
		DisplayName string `json:"display_name" example:"user_display_name"`
	} `json:"author"`
}

// PostEditResponse represents post editing response
type PostEditResponse struct {
	Status int `json:"-" example:"200"`
	Body   PostEditResponseBody
}

// NewPostEditResponse creates a new post edit response
func NewPostEditResponse(postID int, title, content, pseudonymID, displayName, editReason string, isEdited bool) *PostEditResponse {
	now := time.Now()
	response := &PostEditResponse{
		Status: 200,
		Body: PostEditResponseBody{
			PostID:     postID,
			Title:      title,
			Content:    content,
			EditedAt:   now.Format("2006-01-02T15:04:05Z"),
			EditReason: editReason,
			IsEdited:   isEdited,
		},
	}
	response.Body.Author.PseudonymID = pseudonymID
	response.Body.Author.DisplayName = displayName
	return response
}

// CommentDeleteResponseBody represents the body of comment deletion response
type CommentDeleteResponseBody struct {
	CommentID    int    `json:"comment_id" example:"456"`
	DeletedAt    string `json:"deleted_at" example:"2024-01-01T16:00:00Z"`
	DeleteReason string `json:"delete_reason" example:"User requested deletion"`
	DeletedBy    struct {
		PseudonymID string `json:"pseudonym_id" example:"user_pseudonym_id"`
		DisplayName string `json:"display_name" example:"user_name"`
	} `json:"deleted_by"`
}

// CommentDeleteResponse represents comment deletion response
type CommentDeleteResponse struct {
	Status int `json:"-" example:"200"`
	Body   CommentDeleteResponseBody
}

// PostDeleteResponseBody represents the body of post deletion response
type PostDeleteResponseBody struct {
	PostID       int    `json:"post_id" example:"123"`
	DeletedAt    string `json:"deleted_at" example:"2024-01-01T16:00:00Z"`
	DeleteReason string `json:"delete_reason" example:"User requested deletion"`
	DeletedBy    struct {
		PseudonymID string `json:"pseudonym_id" example:"user_pseudonym_id"`
		DisplayName string `json:"display_name" example:"user_name"`
	} `json:"deleted_by"`
}

// PostDeleteResponse represents post deletion response
type PostDeleteResponse struct {
	Status int `json:"-" example:"200"`
	Body   PostDeleteResponseBody
}
