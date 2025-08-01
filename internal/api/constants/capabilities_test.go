package constants

import (
	"testing"
)

func TestGetCapabilitiesByScope(t *testing.T) {
	tests := []struct {
		name           string
		scope          string
		expectedCaps   []string
		shouldHaveCaps bool
	}{
		// Positive tests
		{
			name:  "authentication scope capabilities",
			scope: ScopeAuthentication,
			expectedCaps: []string{
				CapabilityAccessOwnPseudonyms,
				CapabilityLogin,
				CapabilitySessionManagement,
			},
			shouldHaveCaps: true,
		},
		{
			name:  "self_correlation scope capabilities",
			scope: ScopeSelfCorrelation,
			expectedCaps: []string{
				CapabilityVerifyOwnPseudonymOwnership,
				CapabilityManageOwnProfile,
				CapabilityManageOwnPseudonyms,
			},
			shouldHaveCaps: true,
		},
		{
			name:  "moderation scope capabilities",
			scope: ScopeModeration,
			expectedCaps: []string{
				CapabilityModerateContent,
				CapabilityBanUsers,
				CapabilityRemoveContent,
				CapabilityManageModerators,
				CapabilityReviewReports,
				CapabilityForwardReports,
				CapabilityManageSubforumRules,
				CapabilityManageSubforumSettings,
			},
			shouldHaveCaps: true,
		},
		{
			name:  "administration scope capabilities",
			scope: ScopeAdministration,
			expectedCaps: []string{
				CapabilityModeration,
				CapabilityCompliance,
				CapabilityLegalRequests,
				CapabilitySystemAdmin,
				CapabilityUserManagement,
			},
			shouldHaveCaps: true,
		},
		{
			name:  "correlation scope capabilities",
			scope: ScopeCorrelation,
			expectedCaps: []string{
				CapabilityAccessAllPseudonyms,
				CapabilityAccessSubforumPseudonyms,
				CapabilityCrossUserCorrelation,
				CapabilityCorrelateFingerprints,
			},
			shouldHaveCaps: true,
		},

		// Negative tests
		{
			name:           "unknown scope",
			scope:          "unknown_scope",
			expectedCaps:   []string{},
			shouldHaveCaps: false,
		},
		{
			name:           "empty scope",
			scope:          "",
			expectedCaps:   []string{},
			shouldHaveCaps: false,
		},
		{
			name:           "case sensitive scope",
			scope:          "AUTHENTICATION",
			expectedCaps:   []string{},
			shouldHaveCaps: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCapabilitiesByScope(tt.scope)

			if tt.shouldHaveCaps {
				if len(result) == 0 {
					t.Errorf("Expected capabilities for scope %s, got none", tt.scope)
				}

				// Check that all expected capabilities are present
				for _, expectedCap := range tt.expectedCaps {
					found := false
					for _, cap := range result {
						if cap == expectedCap {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected capability %s not found in scope %s", expectedCap, tt.scope)
					}
				}

				// Check that no unexpected capabilities are present
				for _, cap := range result {
					found := false
					for _, expectedCap := range tt.expectedCaps {
						if cap == expectedCap {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Unexpected capability %s found in scope %s", cap, tt.scope)
					}
				}
			} else {
				if len(result) != 0 {
					t.Errorf("Expected no capabilities for scope %s, got %v", tt.scope, result)
				}
			}
		})
	}
}

func TestCapabilityScopeSeparation(t *testing.T) {
	// Test that capabilities are properly separated by scope
	moderationCaps := GetCapabilitiesByScope(ScopeModeration)
	correlationCaps := GetCapabilitiesByScope(ScopeCorrelation)
	administrationCaps := GetCapabilitiesByScope(ScopeAdministration)

	// Verify that moderation capabilities are not in correlation scope
	for _, modCap := range moderationCaps {
		for _, corrCap := range correlationCaps {
			if modCap == corrCap {
				t.Errorf("Moderation capability %s should not be in correlation scope", modCap)
			}
		}
	}

	// Verify that administration capabilities are not in correlation scope
	for _, adminCap := range administrationCaps {
		for _, corrCap := range correlationCaps {
			if adminCap == corrCap {
				t.Errorf("Administration capability %s should not be in correlation scope", adminCap)
			}
		}
	}

	// Verify that correlation capabilities are not in moderation scope
	for _, corrCap := range correlationCaps {
		for _, modCap := range moderationCaps {
			if corrCap == modCap {
				t.Errorf("Correlation capability %s should not be in moderation scope", corrCap)
			}
		}
	}

	// Verify that administration capabilities are not in moderation scope
	for _, adminCap := range administrationCaps {
		for _, modCap := range moderationCaps {
			if adminCap == modCap {
				t.Errorf("Administration capability %s should not be in moderation scope", adminCap)
			}
		}
	}
}

func TestCapabilityUniqueness(t *testing.T) {
	// Test that all capabilities are unique across all scopes
	allScopes := AllScopes()
	allCapabilities := make(map[string]string) // capability -> scope

	for _, scope := range allScopes {
		capabilities := GetCapabilitiesByScope(scope)
		for _, capability := range capabilities {
			if existingScope, exists := allCapabilities[capability]; exists {
				t.Errorf("Capability %s appears in both scope %s and scope %s", capability, existingScope, scope)
			}
			allCapabilities[capability] = scope
		}
	}
}

func TestCapabilityConstants(t *testing.T) {
	// Test that all capability constants are defined and non-empty
	capabilities := []string{
		// Authentication capabilities
		CapabilityAccessOwnPseudonyms,
		CapabilityLogin,
		CapabilitySessionManagement,

		// Self-correlation capabilities
		CapabilityVerifyOwnPseudonymOwnership,
		CapabilityManageOwnProfile,
		CapabilityManageOwnPseudonyms,

		// Moderation capabilities
		CapabilityModerateContent,
		CapabilityBanUsers,
		CapabilityRemoveContent,
		CapabilityManageModerators,
		CapabilityReviewReports,
		CapabilityForwardReports,
		CapabilityManageSubforumRules,

		// Administration capabilities
		CapabilityModeration,
		CapabilityCompliance,
		CapabilityLegalRequests,
		CapabilitySystemAdmin,
		CapabilityUserManagement,

		// Correlation capabilities
		CapabilityAccessAllPseudonyms,
		CapabilityAccessSubforumPseudonyms,
		CapabilityCrossUserCorrelation,
		CapabilityCorrelateFingerprints,
	}

	// Check that all capabilities are non-empty
	for i, capability := range capabilities {
		if capability == "" {
			t.Errorf("Capability at index %d is empty", i)
		}
	}

	// Check that all capabilities are unique
	seen := make(map[string]bool)
	for _, capability := range capabilities {
		if seen[capability] {
			t.Errorf("Duplicate capability: %s", capability)
		}
		seen[capability] = true
	}
}

func TestGetAllCapabilities(t *testing.T) {
	allCaps := GetAllCapabilities()
	allScopes := AllScopes()

	// Verify that all capabilities from all scopes are included
	for _, scope := range allScopes {
		scopeCaps := GetCapabilitiesByScope(scope)
		for _, scopeCap := range scopeCaps {
			found := false
			for _, allCap := range allCaps {
				if allCap == scopeCap {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("Capability %s from scope %s not found in GetAllCapabilities()", scopeCap, scope)
			}
		}
	}

	// Verify that GetAllCapabilities() doesn't contain duplicates
	seen := make(map[string]bool)
	for _, cap := range allCaps {
		if seen[cap] {
			t.Errorf("Duplicate capability in GetAllCapabilities(): %s", cap)
		}
		seen[cap] = true
	}
}

func TestRoleCapabilityConsistency(t *testing.T) {
	// Test that role definitions are consistent with scope capabilities
	roleDefs := GetRoleDefinitions()

	for _, roleDef := range roleDefs {
		for scope, capabilities := range roleDef.Capabilities {
			// Verify that the scope is valid
			if !IsValidScope(scope) {
				t.Errorf("Role %s has invalid scope: %s", roleDef.RoleName, scope)
			}

			// Verify that all capabilities for this scope are valid for the scope
			validCaps := GetCapabilitiesByScope(scope)
			for _, capability := range capabilities {
				found := false
				for _, validCap := range validCaps {
					if capability == validCap {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Role %s has invalid capability %s for scope %s", roleDef.RoleName, capability, scope)
				}
			}
		}
	}
}
