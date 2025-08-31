package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
	"github.com/matt0x6f/hashpost/internal/database/dao"
	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/matt0x6f/hashpost/internal/services"
	"github.com/stephenafamo/bob"
)

// RegisterMessagesRoutes registers direct message routes
func RegisterMessagesRoutes(api huma.API, db bob.Executor, pseudonymDAO dao.PseudonymDAOInterface, ibeSystem *ibe.IBESystem) {
	// Create DAOs for encrypted messaging
	userEncryptionKeyDAO := dao.NewUserEncryptionKeyDAO(db)
	conversationKeyDAO := dao.NewConversationKeyDAO(db)
	encryptedMessageDAO := dao.NewEncryptedMessageDAO(db)

	// Create services for encrypted messaging
	encryptionService := services.NewEncryptionService()
	roleKeyDAO := dao.NewRoleKeyDAO(db, nil)
	keyManagementService := services.NewKeyManagementService(encryptionService, roleKeyDAO, ibeSystem)

	messagesHandler := handlers.NewMessagesHandler(
		dao.NewDirectMessageDAO(db),
		dao.NewUserDAO(db),
		pseudonymDAO,
		userEncryptionKeyDAO,
		conversationKeyDAO,
		encryptedMessageDAO,
		encryptionService,
		keyManagementService,
		dao.NewPermissionDAO(db),
		db,
	)

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

	// Encrypted messaging endpoints
	// Send encrypted message
	huma.Register(api, huma.Operation{
		OperationID: "send-encrypted-message",
		Method:      http.MethodPost,
		Path:        "/api/v1/messages/send",
		Summary:     "Send an encrypted message to another user",
		Description: "Send an end-to-end encrypted message to another user",
		Tags:        []string{"Encrypted Messages"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, messagesHandler.SendEncryptedMessage)

	// Get conversations
	huma.Register(api, huma.Operation{
		OperationID: "get-conversations",
		Method:      http.MethodGet,
		Path:        "/api/v1/messages/conversations",
		Summary:     "Get all conversations for the authenticated user",
		Description: "Retrieve a list of all encrypted conversations for the current user",
		Tags:        []string{"Encrypted Messages"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, messagesHandler.GetConversations)

	// Get conversation messages
	huma.Register(api, huma.Operation{
		OperationID: "get-conversation-messages",
		Method:      http.MethodGet,
		Path:        "/api/v1/messages/conversations/{conversation_id}/messages",
		Summary:     "Get messages from a specific conversation",
		Description: "Retrieve encrypted messages from a specific conversation",
		Tags:        []string{"Encrypted Messages"},
		Security:    []map[string][]string{{"jwt": {}}},
	}, messagesHandler.GetConversationMessages)
}
