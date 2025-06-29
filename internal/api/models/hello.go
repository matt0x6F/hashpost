package models

// HelloInput represents the hello endpoint input (empty for GET)
type HelloInput struct{}

// HelloResponseBody represents the body of hello endpoint response
type HelloResponseBody struct {
	Message string `json:"message" example:"Hello, World!"`
}

// HelloResponse represents the hello endpoint response
type HelloResponse struct {
	Status int               `json:"-" example:"200"`
	Body   HelloResponseBody `json:"body"`
}

// NewHelloResponse creates a new hello response
func NewHelloResponse() *HelloResponse {
	return &HelloResponse{
		Status: 200,
		Body: HelloResponseBody{
			Message: "Hello, World!",
		},
	}
}
