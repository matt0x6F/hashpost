package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/stephenafamo/bob"
)

// RegisterSearchRoutes registers search routes
func RegisterSearchRoutes(api huma.API, db bob.Executor, ibeSystem *ibe.IBESystem) {
	// Use production mode with the provided IBE system
	searchHandler := handlers.NewSearchHandler(db, nil, nil, nil, nil, nil, ibeSystem)

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
		Security:    []map[string][]string{{"jwt": {}}},
	}, searchHandler.SearchUsers)

	// Search pseudonyms (requires platform admin role)
	huma.Register(api, huma.Operation{
		OperationID: "search-pseudonyms",
		Method:      http.MethodGet,
		Path:        "/search/pseudonyms",
		Summary:     "Search for pseudonyms by display name, slug, or ID",
		Description: "Search for pseudonyms by display name, slug, or ID. Requires platform admin role.",
		Tags:        []string{"Search"},
	}, searchHandler.SearchPseudonyms)

	// Public search pseudonyms (no auth required, for co-moderator selection)
	huma.Register(api, huma.Operation{
		OperationID: "search-pseudonyms-public",
		Method:      http.MethodGet,
		Path:        "/search/pseudonyms/public",
		Summary:     "Search for pseudonyms publicly (for co-moderator selection)",
		Description: "Search for active pseudonyms by display name, slug, or ID. This endpoint is public and doesn't require authentication. Used for selecting co-moderators when creating democratic subforums.",
		Tags:        []string{"Search"},
	}, searchHandler.PublicSearchPseudonyms)
}
