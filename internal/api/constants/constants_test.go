package constants

import (
	"testing"
)

func TestScopes(t *testing.T) {
	// Test that all scopes are valid
	for _, scope := range AllScopes() {
		if !IsValidScope(scope) {
			t.Errorf("Scope %s should be valid", scope)
		}
	}

	// Test that invalid scopes are not valid
	invalidScopes := []string{"invalid", "test", ""}
	for _, scope := range invalidScopes {
		if IsValidScope(scope) {
			t.Errorf("Scope %s should not be valid", scope)
		}
	}
}

func TestCapabilities(t *testing.T) {
	// Test that capabilities are valid for their scopes
	authCaps := GetCapabilitiesByScope(ScopeAuthentication)
	if len(authCaps) == 0 {
		t.Error("Authentication scope should have capabilities")
	}

	selfCorrCaps := GetCapabilitiesByScope(ScopeSelfCorrelation)
	if len(selfCorrCaps) == 0 {
		t.Error("Self-correlation scope should have capabilities")
	}

	corrCaps := GetCapabilitiesByScope(ScopeCorrelation)
	if len(corrCaps) == 0 {
		t.Error("Correlation scope should have capabilities")
	}

	// Test that invalid scopes return empty capabilities
	invalidCaps := GetCapabilitiesByScope("invalid")
	if len(invalidCaps) != 0 {
		t.Error("Invalid scope should return empty capabilities")
	}
}

func TestRoles(t *testing.T) {
	// Test that all roles are valid
	for _, role := range GetAllRoles() {
		if !IsValidRole(role) {
			t.Errorf("Role %s should be valid", role)
		}
	}

	// Test that invalid roles are not valid
	invalidRoles := []string{"invalid", "test", ""}
	for _, role := range invalidRoles {
		if IsValidRole(role) {
			t.Errorf("Role %s should not be valid", role)
		}
	}

	// Test role capabilities
	userCaps := GetRoleCapabilities(RoleUser)
	if len(userCaps) == 0 {
		t.Error("User role should have capabilities")
	}

	adminCaps := GetRoleCapabilities(RolePlatformAdmin)
	if len(adminCaps) == 0 {
		t.Error("Platform admin role should have capabilities")
	}

	// Test role scopes
	userScopes := GetRoleScopes(RoleUser)
	if len(userScopes) == 0 {
		t.Error("User role should have scopes")
	}

	adminScopes := GetRoleScopes(RolePlatformAdmin)
	if len(adminScopes) == 0 {
		t.Error("Platform admin role should have scopes")
	}
}

func TestRoleDefinitions(t *testing.T) {
	definitions := GetRoleDefinitions()
	if len(definitions) == 0 {
		t.Error("Should have role definitions")
	}

	// Test that each role definition has the expected structure
	for _, def := range definitions {
		if def.RoleName == "" {
			t.Error("Role definition should have a role name")
		}
		if len(def.Scopes) == 0 {
			t.Errorf("Role %s should have scopes", def.RoleName)
		}
		if len(def.Capabilities) == 0 {
			t.Errorf("Role %s should have capabilities", def.RoleName)
		}

		// Test that each scope has capabilities
		for _, scope := range def.Scopes {
			if caps, exists := def.Capabilities[scope]; !exists || len(caps) == 0 {
				t.Errorf("Role %s scope %s should have capabilities", def.RoleName, scope)
			}
		}
	}
}
