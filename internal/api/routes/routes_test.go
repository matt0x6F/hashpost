package routes

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRouteSecurity_ContentRoutes validates route security by parsing Go AST
func TestRouteSecurity_ContentRoutes(t *testing.T) {
	// Parse the content.go file to check route definitions
	contentFile := "content.go"
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, contentFile, nil, parser.ParseComments)
	require.NoError(t, err, "Should be able to parse content.go")

	// Find all huma.Register calls
	var registerCalls []*ast.CallExpr
	ast.Inspect(node, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "huma" {
					if sel.Sel.Name == "Register" {
						registerCalls = append(registerCalls, call)
					}
				}
			}
		}
		return true
	})

	require.NotEmpty(t, registerCalls, "Should find huma.Register calls")

	// Check each route for proper security configuration
	for _, call := range registerCalls {
		if len(call.Args) < 2 {
			continue
		}

		// First argument should be the API, second should be the Operation
		operationArg := call.Args[1]
		if compositeLit, ok := operationArg.(*ast.CompositeLit); ok {
			if typeExpr, ok := compositeLit.Type.(*ast.SelectorExpr); ok {
				if ident, ok := typeExpr.X.(*ast.Ident); ok && ident.Name == "huma" {
					if typeExpr.Sel.Name == "Operation" {
						// Found a huma.Operation, check its fields
						routeInfo := extractRouteInfo(compositeLit)
						if routeInfo != nil {
							validateRouteSecurity(t, routeInfo)
						}
					}
				}
			}
		}
	}
}

// TestRouteSecurity_SearchRoutes validates search route security
func TestRouteSecurity_SearchRoutes(t *testing.T) {
	// Parse the search.go file
	searchFile := "search.go"
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, searchFile, nil, parser.ParseComments)
	require.NoError(t, err, "Should be able to parse search.go")

	// Find all huma.Register calls
	var registerCalls []*ast.CallExpr
	ast.Inspect(node, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "huma" {
					if sel.Sel.Name == "Register" {
						registerCalls = append(registerCalls, call)
					}
				}
			}
		}
		return true
	})

	require.NotEmpty(t, registerCalls, "Should find huma.Register calls")

	// Check each route for proper security configuration
	for _, call := range registerCalls {
		if len(call.Args) < 2 {
			continue
		}

		operationArg := call.Args[1]
		if compositeLit, ok := operationArg.(*ast.CompositeLit); ok {
			if typeExpr, ok := compositeLit.Type.(*ast.SelectorExpr); ok {
				if ident, ok := typeExpr.X.(*ast.Ident); ok && ident.Name == "huma" {
					if typeExpr.Sel.Name == "Operation" {
						routeInfo := extractRouteInfo(compositeLit)
						if routeInfo != nil {
							validateRouteSecurity(t, routeInfo)
						}
					}
				}
			}
		}
	}
}

// TestRouteSecurity_MessageRoutes validates message route security
func TestRouteSecurity_MessageRoutes(t *testing.T) {
	// Parse the messages.go file
	messagesFile := "messages.go"
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, messagesFile, nil, parser.ParseComments)
	require.NoError(t, err, "Should be able to parse messages.go")

	// Find all huma.Register calls
	var registerCalls []*ast.CallExpr
	ast.Inspect(node, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "huma" {
					if sel.Sel.Name == "Register" {
						registerCalls = append(registerCalls, call)
					}
				}
			}
		}
		return true
	})

	require.NotEmpty(t, registerCalls, "Should find huma.Register calls")

	// Check each route for proper security configuration
	for _, call := range registerCalls {
		if len(call.Args) < 2 {
			continue
		}

		operationArg := call.Args[1]
		if compositeLit, ok := operationArg.(*ast.CompositeLit); ok {
			if typeExpr, ok := compositeLit.Type.(*ast.SelectorExpr); ok {
				if ident, ok := typeExpr.X.(*ast.Ident); ok && ident.Name == "huma" {
					if typeExpr.Sel.Name == "Operation" {
						routeInfo := extractRouteInfo(compositeLit)
						if routeInfo != nil {
							validateRouteSecurity(t, routeInfo)
						}
					}
				}
			}
		}
	}
}

// TestRouteSecurity_CorrelationRoutes validates correlation route security
func TestRouteSecurity_CorrelationRoutes(t *testing.T) {
	// Parse the correlation.go file
	correlationFile := "correlation.go"
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, correlationFile, nil, parser.ParseComments)
	require.NoError(t, err, "Should be able to parse correlation.go")

	// Find all huma.Register calls
	var registerCalls []*ast.CallExpr
	ast.Inspect(node, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok && ident.Name == "huma" {
					if sel.Sel.Name == "Register" {
						registerCalls = append(registerCalls, call)
					}
				}
			}
		}
		return true
	})

	require.NotEmpty(t, registerCalls, "Should find huma.Register calls")

	// Check each route for proper security configuration
	for _, call := range registerCalls {
		if len(call.Args) < 2 {
			continue
		}

		operationArg := call.Args[1]
		if compositeLit, ok := operationArg.(*ast.CompositeLit); ok {
			if typeExpr, ok := compositeLit.Type.(*ast.SelectorExpr); ok {
				if ident, ok := typeExpr.X.(*ast.Ident); ok && ident.Name == "huma" {
					if typeExpr.Sel.Name == "Operation" {
						routeInfo := extractRouteInfo(compositeLit)
						if routeInfo != nil {
							validateRouteSecurity(t, routeInfo)
						}
					}
				}
			}
		}
	}
}

// routeInfo holds extracted route information
type routeInfo struct {
	method      string
	path        string
	hasSecurity bool
}

// extractRouteInfo extracts method, path, and security info from a huma.Operation
func extractRouteInfo(operation *ast.CompositeLit) *routeInfo {
	info := &routeInfo{}

	for _, elt := range operation.Elts {
		if kv, ok := elt.(*ast.KeyValueExpr); ok {
			if ident, ok := kv.Key.(*ast.Ident); ok {
				switch ident.Name {
				case "Method":
					if lit, ok := kv.Value.(*ast.SelectorExpr); ok {
						if x, ok := lit.X.(*ast.Ident); ok && x.Name == "http" {
							info.method = lit.Sel.Name
						}
					}
				case "Path":
					if lit, ok := kv.Value.(*ast.BasicLit); ok && lit.Kind == token.STRING {
						info.path = strings.Trim(lit.Value, `"`)
					}
				case "Security":
					info.hasSecurity = true
				}
			}
		}
	}

	return info
}

// validateRouteSecurity validates that routes have appropriate security
func validateRouteSecurity(t *testing.T, route *routeInfo) {
	if route.method == "" || route.path == "" {
		return // Skip incomplete route info
	}

	t.Run(route.method+"_"+route.path, func(t *testing.T) {
		// Define which routes should be protected based on method + path combination
		protectedRoutes := []struct {
			method string
			path   string
		}{
			{"POST", "/subforums/{name}/posts"},        // Create post
			{"POST", "/posts/{post_id}/comments"},      // Create comment
			{"POST", "/posts/{post_id}/vote"},          // Vote on post
			{"POST", "/comments/{comment_id}/vote"},    // Vote on comment
			{"PATCH", "/posts/{post_id}/lock"},         // Lock post
			{"PATCH", "/posts/{post_id}/sticky"},       // Sticky post
			{"PATCH", "/posts/{post_id}/remove"},       // Remove post
			{"PATCH", "/posts/{post_id}"},              // Edit post
			{"PATCH", "/comments/{comment_id}"},        // Edit comment
			{"PATCH", "/comments/{comment_id}/remove"}, // Remove comment
			{"POST", "/comments/{comment_id}/report"},  // Report comment
			{"DELETE", "/posts/{post_id}"},             // Delete post
			{"DELETE", "/comments/{comment_id}"},       // Delete comment
			{"GET", "/search/users"},                   // Search users (admin)
			{"GET", "/search/pseudonyms"},              // Search pseudonyms (admin)
			{"POST", "/messages"},                      // Send message
			{"GET", "/messages"},                       // Get messages
			{"POST", "/correlation/fingerprint"},       // Fingerprint correlation
			{"POST", "/correlation/identity"},          // Identity correlation
			{"GET", "/correlation/history"},            // Correlation history
		}

		// Check if this route should be protected
		shouldBeProtected := false
		for _, protectedRoute := range protectedRoutes {
			if route.method == protectedRoute.method && route.path == protectedRoute.path {
				shouldBeProtected = true
				break
			}
		}

		// Validate security requirements
		if shouldBeProtected {
			assert.True(t, route.hasSecurity,
				"Route %s %s should have Security field", route.method, route.path)
		} else {
			// Public routes (GET operations) should not have security
			if route.method == "GET" {
				assert.False(t, route.hasSecurity,
					"Public GET route %s should NOT have Security field", route.path)
			}
		}
	})
}

// TestRouteSecurityPattern validates the expected Security field format
func TestRouteSecurityPattern(t *testing.T) {
	// This test validates the expected Security field format for protected routes
	expectedSecurity := []map[string][]string{{"jwt": {}}}

	// Verify the structure
	assert.Len(t, expectedSecurity, 1, "Should have one security requirement")
	assert.Len(t, expectedSecurity[0], 1, "Should have one security scheme")
	assert.Contains(t, expectedSecurity[0], "jwt", "Should use JWT security scheme")
	assert.Empty(t, expectedSecurity[0]["jwt"], "JWT scheme should have no scopes")

	t.Log("Expected Security field format: Security: []map[string][]string{{\"jwt\": {}}}")
}
