package routes

import (
	"database/sql"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob"
)

// RegisterContentRoutes registers content-related routes
func RegisterContentRoutes(api huma.API, db bob.Executor, rawDB *sql.DB, ibeSystem *ibe.IBESystem, identityMappingDAO *dao.IdentityMappingDAO, userDAO *dao.UserDAO) {
	contentHandler := handlers.NewContentHandler(db, rawDB, ibeSystem, identityMappingDAO, userDAO)

	// Get posts from subforum
	huma.Register(api, huma.Operation{
		OperationID: "get-subforum-posts",
		Method:      http.MethodGet,
		Path:        "/subforums/{name}/posts",
		Summary:     "Get posts from a subforum",
		Description: "Retrieves a paginated list of posts from a specific subforum with optional sorting",
		Tags:        []string{"Content"},
	}, contentHandler.GetPosts)

	// Create post in subforum
	huma.Register(api, huma.Operation{
		OperationID: "create-post",
		Method:      http.MethodPost,
		Path:        "/subforums/{name}/posts",
		Summary:     "Create a new post",
		Description: "Creates a new post in the specified subforum",
		Tags:        []string{"Content"},
	}, contentHandler.CreatePost)

	// Get post details
	huma.Register(api, huma.Operation{
		OperationID: "get-post-details",
		Method:      http.MethodGet,
		Path:        "/posts/{post_id}",
		Summary:     "Get detailed information about a specific post",
		Description: "Retrieves detailed information about a post including comments",
		Tags:        []string{"Content"},
	}, contentHandler.GetPostDetails)

	// Get post details by slug
	huma.Register(api, huma.Operation{
		OperationID: "get-post-by-slug",
		Method:      http.MethodGet,
		Path:        "/subforums/{subforum}/posts/{slug}",
		Summary:     "Get detailed information about a specific post by slug",
		Description: "Retrieves detailed information about a post including comments using subforum name and post slug",
		Tags:        []string{"Content"},
	}, contentHandler.GetPostBySlug)

	// Vote on post
	huma.Register(api, huma.Operation{
		OperationID: "vote-on-post",
		Method:      http.MethodPost,
		Path:        "/posts/{post_id}/vote",
		Summary:     "Vote on a post",
		Description: "Votes on a post (upvote, downvote, or remove vote)",
		Tags:        []string{"Content"},
	}, contentHandler.VoteOnPost)

	// Create comment on post
	huma.Register(api, huma.Operation{
		OperationID: "create-comment",
		Method:      http.MethodPost,
		Path:        "/posts/{post_id}/comments",
		Summary:     "Create a comment on a post",
		Description: "Creates a comment on a post, optionally as a reply to another comment",
		Tags:        []string{"Content"},
	}, contentHandler.CreateComment)

	// Vote on comment
	huma.Register(api, huma.Operation{
		OperationID: "vote-on-comment",
		Method:      http.MethodPost,
		Path:        "/comments/{comment_id}/vote",
		Summary:     "Vote on a comment",
		Description: "Votes on a comment (upvote, downvote, or remove vote)",
		Tags:        []string{"Content"},
	}, contentHandler.VoteOnComment)

	// Lock/Unlock post
	huma.Register(api, huma.Operation{
		OperationID: "lock-post",
		Method:      http.MethodPatch,
		Path:        "/posts/{post_id}/lock",
		Summary:     "Lock or unlock a post (moderators only)",
		Description: "Locks or unlocks a post. Requires moderator permission.",
		Tags:        []string{"Content", "Moderation"},
	}, contentHandler.LockPost)

	// Sticky/Unsticky post
	huma.Register(api, huma.Operation{
		OperationID: "sticky-post",
		Method:      http.MethodPatch,
		Path:        "/posts/{post_id}/sticky",
		Summary:     "Sticky or unsticky a post (moderators only)",
		Description: "Stickies or unstickies a post. Requires moderator permission.",
		Tags:        []string{"Content", "Moderation"},
	}, contentHandler.StickyPost)

	// Remove/Restore post
	huma.Register(api, huma.Operation{
		OperationID: "remove-post",
		Method:      http.MethodPatch,
		Path:        "/posts/{post_id}/remove",
		Summary:     "Remove or restore a post (moderators only)",
		Description: "Removes or restores a post. Requires moderator permission.",
		Tags:        []string{"Content", "Moderation"},
	}, contentHandler.RemovePost)

	// Edit comment
	huma.Register(api, huma.Operation{
		OperationID: "edit-comment",
		Method:      http.MethodPatch,
		Path:        "/comments/{comment_id}",
		Summary:     "Edit a comment",
		Description: "Edits a comment. Users can only edit their own comments.",
		Tags:        []string{"Content"},
	}, contentHandler.EditComment)

	// Remove/Restore comment
	huma.Register(api, huma.Operation{
		OperationID: "remove-comment",
		Method:      http.MethodPatch,
		Path:        "/comments/{comment_id}/remove",
		Summary:     "Remove or restore a comment",
		Description: "Removes or restores a comment. Users can remove their own comments, moderators can remove any comment.",
		Tags:        []string{"Content", "Moderation"},
	}, contentHandler.RemoveComment)

	// Report comment
	huma.Register(api, huma.Operation{
		OperationID: "report-comment",
		Method:      http.MethodPost,
		Path:        "/comments/{comment_id}/report",
		Summary:     "Report a comment",
		Description: "Reports a comment for moderation review. Users cannot report their own comments.",
		Tags:        []string{"Content", "Moderation"},
	}, contentHandler.ReportComment)
}
