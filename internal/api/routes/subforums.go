package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/stephenafamo/bob"
)

// RegisterSubforumRoutes registers subforum-related routes
func RegisterSubforumRoutes(api huma.API, db bob.Executor, pseudonymDAO dao.PseudonymDAOInterface) {
	// Create DAOs needed for the subforum handler
	subforumDAO := dao.NewSubforumDAO(db)
	subforumSubscriptionDAO := dao.NewSubforumSubscriptionDAO(db)
	permissionDAO := dao.NewPermissionDAO(db)
	identityMappingDAO := dao.NewIdentityMappingDAO(db)
	postDAO := dao.NewPostDAO(db)
	roleKeyDAO := dao.NewRoleKeyDAO(db)

	subforumHandler := handlers.NewSubforumHandler(db, subforumDAO, subforumSubscriptionDAO, permissionDAO, identityMappingDAO, pseudonymDAO, postDAO, roleKeyDAO)

	// Admin routes (more specific) - register first to avoid conflicts
	// Subforum settings routes
	huma.Register(api, huma.Operation{
		OperationID: "get-subforum-settings",
		Method:      http.MethodGet,
		Path:        "/subforums/{type}/{name}/admin/settings",
		Summary:     "Get subforum settings",
		Description: "Retrieves settings for a specific subforum. Requires manage_subforum_settings capability.",
		Tags:        []string{"Subforums"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, subforumHandler.GetSubforumSettings)

	huma.Register(api, huma.Operation{
		OperationID: "update-subforum-settings",
		Method:      http.MethodPut,
		Path:        "/subforums/{type}/{name}/admin/settings",
		Summary:     "Update subforum settings",
		Description: "Updates settings for a specific subforum. Requires manage_subforum_settings capability.",
		Tags:        []string{"Subforums"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, subforumHandler.UpdateSubforumSettings)

	// Moderator team management routes
	huma.Register(api, huma.Operation{
		OperationID: "get-moderator-team",
		Method:      http.MethodGet,
		Path:        "/subforums/{type}/{name}/admin/moderators",
		Summary:     "Get moderator team",
		Description: "Retrieves the moderator team for a specific subforum. Requires manage_moderators capability.",
		Tags:        []string{"Subforums"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, subforumHandler.GetModeratorTeam)

	huma.Register(api, huma.Operation{
		OperationID: "add-moderator",
		Method:      http.MethodPost,
		Path:        "/subforums/{type}/{name}/admin/moderators",
		Summary:     "Add moderator",
		Description: "Adds a new moderator to the subforum. Requires manage_moderators capability.",
		Tags:        []string{"Subforums"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, subforumHandler.AddModerator)

	huma.Register(api, huma.Operation{
		OperationID: "update-moderator",
		Method:      http.MethodPut,
		Path:        "/subforums/{type}/{name}/admin/moderators/{pseudonym_id}",
		Summary:     "Update moderator",
		Description: "Updates a moderator's role and permissions. Requires manage_moderators capability.",
		Tags:        []string{"Subforums"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, subforumHandler.UpdateModerator)

	huma.Register(api, huma.Operation{
		OperationID: "remove-moderator",
		Method:      http.MethodDelete,
		Path:        "/subforums/{type}/{name}/admin/moderators/{pseudonym_id}",
		Summary:     "Remove moderator",
		Description: "Removes a moderator from the subforum. Requires manage_moderators capability.",
		Tags:        []string{"Subforums"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, subforumHandler.RemoveModerator)

	// General subforum routes (less specific) - register after admin routes
	// Get subforums
	huma.Register(api, huma.Operation{
		OperationID: "get-subforums",
		Method:      http.MethodGet,
		Path:        "/subforums",
		Summary:     "Get a list of subforums",
		Description: "Retrieves a paginated list of subforums with optional sorting. Supports query parameters: page (default: 1), limit (default: 25), sort (options: name, subscribers, posts, created_at)",
		Tags:        []string{"Subforums"},
	}, subforumHandler.GetSubforums)

	// Get subforum details - support community type prefixes (t/, g/, b/, c/)
	huma.Register(api, huma.Operation{
		OperationID: "get-subforum-details",
		Method:      http.MethodGet,
		Path:        "/subforums/{type}/{name}",
		Summary:     "Get detailed information about a specific subforum",
		Description: "Retrieves detailed information about a subforum including moderators and subscription status. Supports community type prefixes (t/, g/, b/, c/).",
		Tags:        []string{"Subforums"},
	}, subforumHandler.GetSubforumDetails)

	// Subscribe to subforum - support community type prefixes
	huma.Register(api, huma.Operation{
		OperationID: "subscribe-to-subforum",
		Method:      http.MethodPost,
		Path:        "/subforums/{type}/{name}/subscribe",
		Summary:     "Subscribe to a subforum",
		Description: "Subscribes the authenticated user to a subforum. Supports community type prefixes (t/, g/, b/, c/). Requires authentication.",
		Tags:        []string{"Subforums"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, subforumHandler.SubscribeToSubforum)

	// Unsubscribe from subforum - support community type prefixes
	huma.Register(api, huma.Operation{
		OperationID: "unsubscribe-from-subforum",
		Method:      http.MethodDelete,
		Path:        "/subforums/{type}/{name}/subscribe",
		Summary:     "Unsubscribe from a subforum",
		Description: "Unsubscribes the authenticated user from a subforum. Supports community type prefixes (t/, g/, b/, c/). Requires authentication.",
		Tags:        []string{"Subforums"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, subforumHandler.UnsubscribeFromSubforum)

	// Create subforum
	huma.Register(api, huma.Operation{
		OperationID: "create-subforum",
		Method:      http.MethodPost,
		Path:        "/subforums",
		Summary:     "Create a new subforum",
		Description: "Creates a new subforum with community type and governance style. Requires authentication and the create_subforum capability.",
		Tags:        []string{"Subforums"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, subforumHandler.CreateSubforum)

	// Get pseudonym subscriptions
	huma.Register(api, huma.Operation{
		OperationID: "get-pseudonym-subscriptions",
		Method:      http.MethodGet,
		Path:        "/pseudonyms/{pseudonym_id}/subscriptions",
		Summary:     "Get pseudonym subscriptions",
		Description: "Retrieves all subforums that a pseudonym is subscribed to. Only the pseudonym owner can access this endpoint.",
		Tags:        []string{"Subforums"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, subforumHandler.GetPseudonymSubscriptions)
}
