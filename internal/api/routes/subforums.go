package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/stephenafamo/bob"
)

// RegisterSubforumRoutes registers subforum-related routes
func RegisterSubforumRoutes(api huma.API, db bob.Executor) {
	subforumHandler := handlers.NewSubforumHandler(db)

	// Get subforums
	huma.Register(api, huma.Operation{
		OperationID: "get-subforums",
		Method:      http.MethodGet,
		Path:        "/subforums",
		Summary:     "Get a list of subforums",
		Description: "Retrieves a paginated list of subforums with optional sorting. Supports query parameters: page (default: 1), limit (default: 25), sort (options: name, subscribers, posts, created_at)",
		Tags:        []string{"Subforums"},
	}, subforumHandler.GetSubforums)

	// Get subforum details
	huma.Register(api, huma.Operation{
		OperationID: "get-subforum-details",
		Method:      http.MethodGet,
		Path:        "/subforums/{name}",
		Summary:     "Get detailed information about a specific subforum",
		Description: "Retrieves detailed information about a subforum including moderators and subscription status. Requires authentication for private subforums.",
		Tags:        []string{"Subforums"},
	}, subforumHandler.GetSubforumDetails)

	// Subscribe to subforum
	huma.Register(api, huma.Operation{
		OperationID: "subscribe-to-subforum",
		Method:      http.MethodPost,
		Path:        "/subforums/{name}/subscribe",
		Summary:     "Subscribe to a subforum",
		Description: "Subscribes the authenticated user to a subforum. Requires authentication.",
		Tags:        []string{"Subforums"},
	}, subforumHandler.SubscribeToSubforum)

	// Unsubscribe from subforum
	huma.Register(api, huma.Operation{
		OperationID: "unsubscribe-from-subforum",
		Method:      http.MethodDelete,
		Path:        "/subforums/{name}/subscribe",
		Summary:     "Unsubscribe from a subforum",
		Description: "Unsubscribes the authenticated user from a subforum. Requires authentication.",
		Tags:        []string{"Subforums"},
	}, subforumHandler.UnsubscribeFromSubforum)

	// Create subforum
	huma.Register(api, huma.Operation{
		OperationID: "create-subforum",
		Method:      http.MethodPost,
		Path:        "/subforums",
		Summary:     "Create a new subforum",
		Description: "Creates a new subforum. Requires authentication and the create_subforum capability.",
		Tags:        []string{"Subforums"},
	}, subforumHandler.CreateSubforum)
}
