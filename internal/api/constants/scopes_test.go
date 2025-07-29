package constants

import (
	"testing"
)

func TestAllScopes(t *testing.T) {
	scopes := AllScopes()
	expectedScopes := []string{
		ScopeAuthentication,
		ScopeSelfCorrelation,
		ScopeModeration,
		ScopeAdministration,
		ScopeCorrelation,
	}

	if len(scopes) != len(expectedScopes) {
		t.Errorf("Expected %d scopes, got %d", len(expectedScopes), len(scopes))
	}

	for _, expectedScope := range expectedScopes {
		found := false
		for _, scope := range scopes {
			if scope == expectedScope {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected scope %s not found in AllScopes()", expectedScope)
		}
	}
}

func TestIsValidScope(t *testing.T) {
	tests := []struct {
		name     string
		scope    string
		expected bool
	}{
		// Positive tests
		{"authentication scope", ScopeAuthentication, true},
		{"self_correlation scope", ScopeSelfCorrelation, true},
		{"moderation scope", ScopeModeration, true},
		{"administration scope", ScopeAdministration, true},
		{"correlation scope", ScopeCorrelation, true},

		// Negative tests
		{"empty scope", "", false},
		{"invalid scope", "invalid_scope", false},
		{"case sensitive", "AUTHENTICATION", false},
		{"partial match", "auth", false},
		{"extra text", "authentication_extra", false},
		{"special characters", "auth@#$%", false},
		{"numbers only", "123", false},
		{"spaces", "authentication ", false},
		{"leading space", " authentication", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidScope(tt.scope)
			if result != tt.expected {
				t.Errorf("IsValidScope(%q) = %v, expected %v", tt.scope, result, tt.expected)
			}
		})
	}
}

func TestGetScopeDescription(t *testing.T) {
	tests := []struct {
		name        string
		scope       string
		expected    string
		shouldExist bool
	}{
		// Positive tests
		{
			name:        "authentication scope description",
			scope:       ScopeAuthentication,
			expected:    "Basic user authentication and session management",
			shouldExist: true,
		},
		{
			name:        "self_correlation scope description",
			scope:       ScopeSelfCorrelation,
			expected:    "User management of their own identity and pseudonyms",
			shouldExist: true,
		},
		{
			name:        "moderation scope description",
			scope:       ScopeModeration,
			expected:    "Content moderation, report reviews, and rule management",
			shouldExist: true,
		},
		{
			name:        "administration scope description",
			scope:       ScopeAdministration,
			expected:    "Platform administration and user management",
			shouldExist: true,
		},
		{
			name:        "correlation scope description",
			scope:       ScopeCorrelation,
			expected:    "Cross-user identity correlation (admin only)",
			shouldExist: true,
		},

		// Negative tests
		{
			name:        "unknown scope",
			scope:       "unknown_scope",
			expected:    "Unknown scope",
			shouldExist: false,
		},
		{
			name:        "empty scope",
			scope:       "",
			expected:    "Unknown scope",
			shouldExist: false,
		},
		{
			name:        "case sensitive scope",
			scope:       "AUTHENTICATION",
			expected:    "Unknown scope",
			shouldExist: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetScopeDescription(tt.scope)
			if result != tt.expected {
				t.Errorf("GetScopeDescription(%q) = %q, expected %q", tt.scope, result, tt.expected)
			}
		})
	}
}

func TestScopeConstants(t *testing.T) {
	// Test that all scope constants are defined and unique
	scopes := map[string]string{
		"ScopeAuthentication":  ScopeAuthentication,
		"ScopeSelfCorrelation": ScopeSelfCorrelation,
		"ScopeModeration":      ScopeModeration,
		"ScopeAdministration":  ScopeAdministration,
		"ScopeCorrelation":     ScopeCorrelation,
	}

	// Check that all constants are non-empty
	for name, value := range scopes {
		if value == "" {
			t.Errorf("Constant %s is empty", name)
		}
	}

	// Check that all values are unique
	seen := make(map[string]string)
	for name, value := range scopes {
		if existingName, exists := seen[value]; exists {
			t.Errorf("Duplicate scope value %q: %s and %s", value, existingName, name)
		}
		seen[value] = name
	}

	// Verify expected values
	expectedValues := map[string]string{
		"ScopeAuthentication":  "authentication",
		"ScopeSelfCorrelation": "self_correlation",
		"ScopeModeration":      "moderation",
		"ScopeAdministration":  "administration",
		"ScopeCorrelation":     "correlation",
	}

	for name, expectedValue := range expectedValues {
		if scopes[name] != expectedValue {
			t.Errorf("Constant %s = %q, expected %q", name, scopes[name], expectedValue)
		}
	}
}

func TestScopeLogicalGrouping(t *testing.T) {
	// Test that scopes are logically grouped
	userScopes := []string{ScopeAuthentication, ScopeSelfCorrelation}
	moderatorScopes := []string{ScopeAuthentication, ScopeSelfCorrelation, ScopeModeration, ScopeCorrelation}
	adminScopes := []string{ScopeAuthentication, ScopeSelfCorrelation, ScopeModeration, ScopeAdministration, ScopeCorrelation}

	// Verify user scopes are basic
	for _, scope := range userScopes {
		if !IsValidScope(scope) {
			t.Errorf("User scope %s is not valid", scope)
		}
	}

	// Verify moderator scopes include user scopes plus moderation
	for _, userScope := range userScopes {
		found := false
		for _, moderatorScope := range moderatorScopes {
			if moderatorScope == userScope {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Moderator scopes should include user scope %s", userScope)
		}
	}

	// Verify admin scopes include moderator scopes plus administration
	for _, moderatorScope := range moderatorScopes {
		found := false
		for _, adminScope := range adminScopes {
			if adminScope == moderatorScope {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Admin scopes should include moderator scope %s", moderatorScope)
		}
	}

	// Verify that administration scope is only for admins
	adminOnlyScopes := []string{ScopeAdministration}
	for _, adminScope := range adminOnlyScopes {
		found := false
		for _, scope := range adminScopes {
			if scope == adminScope {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Admin-only scope %s should be in admin scopes", adminScope)
		}

		// Verify it's not in user or moderator scopes
		for _, scope := range userScopes {
			if scope == adminScope {
				t.Errorf("Admin-only scope %s should not be in user scopes", adminScope)
			}
		}
		for _, scope := range moderatorScopes {
			if scope == adminScope {
				t.Errorf("Admin-only scope %s should not be in moderator scopes", adminScope)
			}
		}
	}
}
