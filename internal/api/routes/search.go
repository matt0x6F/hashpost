package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
)

// RegisterSearchRoutes registers search routes
func RegisterSearchRoutes(api huma.API) {
	searchHandler := handlers.NewSearchHandler()

	// Search posts
	huma.Register(api, huma.Operation{
		OperationID: "search-posts",
		Method:      http.MethodGet,
		Path:        "/search/posts",
		Summary:     "Search for posts across all subforums",
		Description: "Search for posts across all subforums with various filters",
		Tags:        []string{"Search"},
	}, searchHandler.SearchPosts)

	// Search users
	huma.Register(api, huma.Operation{
		OperationID: "search-users",
		Method:      http.MethodGet,
		Path:        "/search/users",
		Summary:     "Search for users by display name",
		Description: "Search for users by display name",
		Tags:        []string{"Search"},
	}, searchHandler.SearchUsers)
}
