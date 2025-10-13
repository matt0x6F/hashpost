package appview

import (
	"embed"
	"html/template"
	"net/http"
)

//go:embed openapi.yaml
var openAPISpec embed.FS

// SwaggerUIHandler serves the Swagger UI for the OpenAPI specification
func SwaggerUIHandler(w http.ResponseWriter, r *http.Request) {
	// Get the base URL for the API spec
	baseURL := "http://" + r.Host
	if r.TLS != nil {
		baseURL = "https://" + r.Host
	}
	specURL := baseURL + "/openapi.yaml"

	// Create the Swagger UI HTML template
	swaggerHTML := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HashPost AppView API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
        .swagger-ui .topbar {
            background-color: #1f2937;
        }
        .swagger-ui .topbar .download-url-wrapper {
            display: none;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "{{.SpecURL}}",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                tryItOutEnabled: true,
                requestInterceptor: function(request) {
                    // Add any custom request headers or modifications here
                    return request;
                },
                responseInterceptor: function(response) {
                    // Add any custom response handling here
                    return response;
                }
            });
        };
    </script>
</body>
</html>`

	// Parse and execute the template
	tmpl, err := template.New("swagger").Parse(swaggerHTML)
	if err != nil {
		http.Error(w, "Failed to parse template", http.StatusInternalServerError)
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Execute template with the spec URL
	err = tmpl.Execute(w, map[string]interface{}{
		"SpecURL": specURL,
	})
	if err != nil {
		http.Error(w, "Failed to execute template", http.StatusInternalServerError)
		return
	}
}

// OpenAPISpecHandler serves the raw OpenAPI specification
func OpenAPISpecHandler(w http.ResponseWriter, r *http.Request) {
	// Read the OpenAPI spec
	specData, err := openAPISpec.ReadFile("openapi.yaml")
	if err != nil {
		http.Error(w, "Failed to read OpenAPI spec", http.StatusInternalServerError)
		return
	}

	// Set appropriate headers
	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Write the spec
	w.Write(specData)
}
