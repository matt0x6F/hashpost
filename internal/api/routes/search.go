package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/stephenafamo/bob"
)

// RegisterSearchRoutes registers search routes
func RegisterSearchRoutes(api huma.API, db bob.Executor) {
	searchHandler := handlers.NewSearchHandler(db, nil, nil, nil, nil, nil)

	// Search posts
	huma.Register(api, huma.Operation{
		OperationID: "search-posts",
		Method:      http.MethodGet,
		Path:        "/search/posts",
		Summary:     "Search for posts across all subforums",
		Description: "Search for posts across all subforums with various filters",
		Tags:        []string{"Search"},
	}, searchHandler.SearchPosts)

	// Search users (platform admin only)
	huma.Register(api, huma.Operation{
		OperationID: "search-users",
		Method:      http.MethodGet,
		Path:        "/search/users",
		Summary:     "Search for users by display name",
		Description: "Search for users by display name. Requires platform admin role.",
		Tags:        []string{"Search"},
	}, searchHandler.SearchUsers)
}
