package routes

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/matt0x6f/hashpost/internal/api/handlers"
)

// RegisterHelloRoutes registers hello-related routes
func RegisterHelloRoutes(api huma.API) {
	huma.Register(api, huma.Operation{
		OperationID: "hello",
		Method:      http.MethodGet,
		Path:        "/hello",
		Summary:     "Hello world endpoint",
		Description: "Returns a simple hello world message",
	}, handlers.HelloHandler)
}
