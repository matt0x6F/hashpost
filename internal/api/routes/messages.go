package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/stephenafamo/bob"
)

// RegisterMessagesRoutes registers direct message routes
func RegisterMessagesRoutes(api huma.API, db bob.Executor, pseudonymDAO dao.PseudonymDAOInterface) {
	messagesHandler := handlers.NewMessagesHandler(dao.NewDirectMessageDAO(db), dao.NewUserDAO(db), pseudonymDAO)

	// Send direct message
	huma.Register(api, huma.Operation{
		OperationID: "send-direct-message",
		Method:      http.MethodPost,
		Path:        "/messages",
		Summary:     "Send a direct message to another user",
		Description: "Send a direct message to another user",
		Tags:        []string{"Messages"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, messagesHandler.SendDirectMessage)

	// Get direct messages
	huma.Register(api, huma.Operation{
		OperationID: "get-direct-messages",
		Method:      http.MethodGet,
		Path:        "/messages",
		Summary:     "Get direct messages for the current user",
		Description: "Get direct messages for the current user",
		Tags:        []string{"Messages"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, messagesHandler.GetDirectMessages)
}
