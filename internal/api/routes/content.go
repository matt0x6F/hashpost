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
}
