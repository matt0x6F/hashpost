package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

// ExamplePermissionMiddleware demonstrates how to use the permission middleware
func ExamplePermissionMiddleware() {
	// This is an example of how to set up and use the permission middleware
	// in a real application

	// 1. Create a mock permission DAO (in real app, use real DAO with database)
	mockDAO := &MockPermissionDAO{
		hasSubforumCapabilityFunc: func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
			// Mock logic: user 123 has moderation capability for subforum 456
			if userID == 123 && subforumID == 456 && capability == "moderate_content" {
				return true, nil
			}
			return false, nil
		},
		canAccessPrivateSubforumFunc: func(ctx context.Context, userID int64, subforumID int32) (bool, error) {
			// Mock logic: user 123 can access private subforum 789
			if userID == 123 && subforumID == 789 {
				return true, nil
			}
			return false, nil
		},
	}

	// 2. Create permission middleware with the DAO
	permissionMiddleware := NewPermissionMiddlewareWithDAO(mockDAO)

	// 3. Create a test handler that requires moderation capability
	moderationHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get subforum ID from context (set by middleware)
		subforumID, ok := GetSubforumIDFromContext(r)
		if !ok {
			http.Error(w, "Subforum ID not found in context", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Moderation action performed for subforum " + string(rune(subforumID))))
	})

	// 4. Apply the middleware
	protectedHandler := permissionMiddleware.RequireModerationCapability()(moderationHandler)

	// 5. Create a test request with user context
	req := httptest.NewRequest("POST", "/moderate?subforum_id=456", nil)
	userCtx := &UserContext{
		UserID: 123,
		Email:  "moderator@example.com",
		Roles:  []string{"moderator"},
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	// 6. Test the protected endpoint
	w := httptest.NewRecorder()
	protectedHandler.ServeHTTP(w, req)

	// The request should succeed because user 123 has moderation capability for subforum 456
}

// ExamplePermissionChecker demonstrates how to use the permission checker
func ExamplePermissionChecker() {
	// This is an example of how to use the permission checker in handlers

	// 1. Create a mock permission DAO
	mockDAO := &MockPermissionDAO{
		hasSubforumCapabilityFunc: func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
			// Mock logic: user 123 has all capabilities for subforum 456
			if userID == 123 && subforumID == 456 {
				return true, nil
			}
			return false, nil
		},
		getUserSubforumRolesFunc: func(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
			// Mock logic: user 123 is a moderator for subforum 456
			if userID == 123 && subforumID == 456 {
				return []string{"moderator", "admin"}, nil
			}
			return []string{}, nil
		},
		getUserSubforumCapabilitiesFunc: func(ctx context.Context, userID int64, subforumID int32) ([]string, error) {
			// Mock logic: user 123 has all capabilities for subforum 456
			if userID == 123 && subforumID == 456 {
				return []string{"moderate_content", "ban_users", "remove_content", "manage_moderators"}, nil
			}
			return []string{}, nil
		},
	}

	// 2. Create permission checker with the DAO
	checker := NewPermissionCheckerWithDAO(mockDAO)

	// 3. Example handler that uses the permission checker
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user context from request
		userCtx, err := ExtractUserFromRequest(r)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get subforum ID from query parameters
		subforumIDStr := r.URL.Query().Get("subforum_id")
		if subforumIDStr == "" {
			http.Error(w, "Subforum ID required", http.StatusBadRequest)
			return
		}

		// Parse subforum ID (simplified for example)
		subforumID := int32(456) // In real app, parse from subforumIDStr

		// Check specific capability
		canModerate, err := checker.CheckSubforumCapability(r.Context(), userCtx.UserID, subforumID, "moderate_content")
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if !canModerate {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Get user's roles and capabilities for this subforum
		roles, err := checker.GetUserSubforumRoles(r.Context(), userCtx.UserID, subforumID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		capabilities, err := checker.GetUserSubforumCapabilities(r.Context(), userCtx.UserID, subforumID)
		if err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Use the information to customize the response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User has roles: " + strings.Join(roles, ", ") + " and capabilities: " + strings.Join(capabilities, ", ")))
	})

	// 4. Test the handler
	req := httptest.NewRequest("GET", "/user-permissions?subforum_id=456", nil)
	userCtx := &UserContext{
		UserID: 123,
		Email:  "user@example.com",
	}
	ctx := SetUserContext(req.Context(), userCtx)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	// The request should succeed and return the user's roles and capabilities
}

// TestPermissionMiddlewareIntegration tests the middleware in a more realistic scenario
func TestPermissionMiddlewareIntegration(t *testing.T) {
	// Create a mock DAO with realistic behavior
	mockDAO := &MockPermissionDAO{
		hasSubforumCapabilityFunc: func(ctx context.Context, userID int64, subforumID int32, capability string) (bool, error) {
			// Mock logic: user 123 is a moderator for subforum 456
			if userID == 123 && subforumID == 456 {
				switch capability {
				case "moderate_content", "ban_users", "remove_content":
					return true, nil
				default:
					return false, nil
				}
			}
			return false, nil
		},
		canAccessPrivateSubforumFunc: func(ctx context.Context, userID int64, subforumID int32) (bool, error) {
			// Mock logic: user 123 can access private subforum 789
			if userID == 123 && subforumID == 789 {
				return true, nil
			}
			return false, nil
		},
	}

	// Create middleware
	permissionMiddleware := NewPermissionMiddlewareWithDAO(mockDAO)

	// Test cases
	testCases := []struct {
		name           string
		userID         int64
		subforumID     int32
		capability     string
		expectedStatus int
		description    string
	}{
		{
			name:           "Moderator can moderate content",
			userID:         123,
			subforumID:     456,
			capability:     "moderate_content",
			expectedStatus: http.StatusOK,
			description:    "User 123 should be able to moderate content in subforum 456",
		},
		{
			name:           "Moderator can ban users",
			userID:         123,
			subforumID:     456,
			capability:     "ban_users",
			expectedStatus: http.StatusOK,
			description:    "User 123 should be able to ban users in subforum 456",
		},
		{
			name:           "Moderator cannot manage moderators",
			userID:         123,
			subforumID:     456,
			capability:     "manage_moderators",
			expectedStatus: http.StatusForbidden,
			description:    "User 123 should not be able to manage moderators in subforum 456",
		},
		{
			name:           "User can access private subforum",
			userID:         123,
			subforumID:     789,
			capability:     "", // Not used for private subforum access
			expectedStatus: http.StatusOK,
			description:    "User 123 should be able to access private subforum 789",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test handler
			handlerCalled := false
			testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				w.WriteHeader(http.StatusOK)
			})

			// Create request
			req := httptest.NewRequest("GET", "/test?subforum_id="+strconv.Itoa(int(tc.subforumID)), nil)
			userCtx := &UserContext{
				UserID: tc.userID,
				Email:  "test@example.com",
			}
			ctx := SetUserContext(req.Context(), userCtx)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// Apply appropriate middleware based on test case
			var middleware http.Handler
			if tc.capability == "" {
				// Test private subforum access
				middleware = permissionMiddleware.RequirePrivateSubforumAccess()(testHandler)
			} else {
				// Test capability requirement
				middleware = permissionMiddleware.RequireSubforumCapability(tc.capability)(testHandler)
			}

			// Execute middleware
			middleware.ServeHTTP(w, req)

			// Verify results
			if tc.expectedStatus == http.StatusOK && !handlerCalled {
				t.Error("Expected handler to be called")
			}
			if tc.expectedStatus != http.StatusOK && handlerCalled {
				t.Error("Expected handler not to be called")
			}
			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status %d, got %d: %s", tc.expectedStatus, w.Code, tc.description)
			}
		})
	}
}
