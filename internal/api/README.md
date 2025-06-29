# API Organization

This directory contains the organized API structure for the HashPost server.

## Structure

```
internal/api/
├── models/          # Request/response models
├── handlers/        # Operation handlers
├── middleware/      # Router-specific and router-agnostic middleware
├── routes/          # Route registration
├── logger/          # Zerolog configuration and utilities
└── server.go        # API server setup and configuration
```

## Logging

The API uses [zerolog](https://github.com/rs/zerolog) for structured logging. Logs are configured with:

- **Pretty console output** for development
- **Structured JSON format** for production
- **Component-based logging** for easy filtering
- **Request/response logging** with timing and metadata

### Log Levels

- `DEBUG`: Detailed debugging information
- `INFO`: General information about application flow
- `WARN`: Warning messages for potentially harmful situations
- `ERROR`: Error events that might still allow the application to continue
- `FATAL`: Severe errors that cause the application to exit

### Request Logging

Each HTTP request is logged with structured fields:

```json
{
  "level": "info",
  "component": "http",
  "method": "GET",
  "path": "/health",
  "remote_addr": "127.0.0.1:12345",
  "user_agent": "curl/7.68.0",
  "time": "2025-06-22T07:50:59Z",
  "message": "Request started"
}
```

Response logging includes timing information:

```json
{
  "level": "info",
  "component": "http",
  "method": "GET",
  "path": "/health",
  "status": 200,
  "duration": "118.4µs",
  "remote_addr": "127.0.0.1:12345",
  "time": "2025-06-22T07:50:59Z",
  "message": "Request completed"
}
```

## Adding New Endpoints

To add a new endpoint, follow these steps:

### 1. Create Models (`models/your_endpoint.go`)

```go
package models

// YourEndpointInput represents the input for your endpoint
type YourEndpointInput struct {
    // Define your input fields with Huma tags
    Name string `path:"name" maxLength:"30" example:"world" doc:"Name parameter"`
}

// YourEndpointResponse represents the response for your endpoint
type YourEndpointResponse struct {
    Status int `json:"-" example:"200"`
    Body   struct {
        Message string `json:"message" example:"Hello, world!"`
    } `json:"body"`
}

// NewYourEndpointResponse creates a new response
func NewYourEndpointResponse(message string) *YourEndpointResponse {
    return &YourEndpointResponse{
        Status: 200,
        Body: struct {
            Message string `json:"message" example:"Hello, world!"`
        }{
            Message: message,
        },
    }
}
```

### 2. Create Handler (`handlers/your_endpoint.go`)

```go
package handlers

import (
    "context"
    "github.com/matt0x6f/hashpost/internal/api/models"
    "github.com/rs/zerolog/log"
)

// YourEndpointHandler handles your endpoint requests
func YourEndpointHandler(ctx context.Context, input *models.YourEndpointInput) (*models.YourEndpointResponse, error) {
    // Log with structured fields
    log.Info().
        Str("endpoint", "your-endpoint").
        Str("name", input.Name).
        Msg("Processing request")
    
    // Your business logic here
    return models.NewYourEndpointResponse("Hello, " + input.Name + "!"), nil
}
```

### 3. Create Route (`routes/your_endpoint.go`)

```go
package routes

import (
    "net/http"
    "github.com/danielgtaylor/huma/v2"
    "github.com/matt0x6f/hashpost/internal/api/handlers"
)

// RegisterYourEndpointRoutes registers your endpoint routes
func RegisterYourEndpointRoutes(api huma.API) {
    huma.Register(api, huma.Operation{
        OperationID: "your-endpoint",
        Method:      http.MethodGet,
        Path:        "/your-endpoint/{name}",
        Summary:     "Your endpoint summary",
        Description: "Your endpoint description",
    }, handlers.YourEndpointHandler)
}
```

### 4. Register the Route

Add your route registration to `server.go`:

```go
// In NewServer() function, add:
routes.RegisterYourEndpointRoutes(api)
```

## Middleware

### Router-Specific Middleware

Router-specific middleware runs at the HTTP level before Huma processing. It's applied to the entire mux in `server.go`:

```go
// Example: Add to router.go
func (m *RouterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Your router-specific logic here
    m.handler.ServeHTTP(w, r)
}
```

### Router-Agnostic Middleware

Router-agnostic middleware runs in the Huma processing chain. Add it to `server.go`:

```go
// In NewServer() function:
api.UseMiddleware(middleware.YourMiddleware)
```

### Custom Middleware with Logging

When creating custom middleware, use the logger package for consistent logging:

```go
package middleware

import (
    "github.com/danielgtaylor/huma/v2"
    "github.com/matt0x6f/hashpost/internal/api/logger"
)

func YourCustomMiddleware(ctx huma.Context, next func(huma.Context)) {
    log := logger.GetRequestLogger()
    
    log.Info().
        Str("middleware", "your-custom").
        Str("method", ctx.Method()).
        Str("path", ctx.URL().Path).
        Msg("Custom middleware processing")
    
    next(ctx)
}
```

## Testing

To test your new endpoint:

1. Build and run the server: `go run cmd/server/main.go`
2. Test with curl: `curl http://localhost:8888/your-endpoint/world`
3. Check the API docs: `curl http://localhost:8888/docs`
4. Monitor logs for structured output

## Best Practices

1. **Separation of Concerns**: Keep models, handlers, and routes in separate files
2. **Consistent Naming**: Use consistent naming conventions across all files
3. **Error Handling**: Always return appropriate errors from handlers
4. **Documentation**: Add proper documentation and examples to your models
5. **Middleware Order**: Router-specific middleware runs first, then router-agnostic middleware
6. **Structured Logging**: Use zerolog for all logging with appropriate log levels and structured fields
7. **Component Logging**: Use component-based logging for easy filtering and debugging 