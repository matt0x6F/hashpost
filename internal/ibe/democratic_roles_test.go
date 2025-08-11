package ibe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectDomain_DemocraticRoles(t *testing.T) {
	tests := []struct {
		name           string
		role           string
		expectedDomain string
	}{
		{
			name:           "elected moderator uses mod correlation domain",
			role:           "elected_moderator",
			expectedDomain: DOMAIN_MOD_CORRELATION,
		},
		{
			name:           "appointed moderator uses mod correlation domain",
			role:           "appointed_moderator",
			expectedDomain: DOMAIN_MOD_CORRELATION,
		},
		{
			name:           "regular moderator still uses mod correlation domain",
			role:           "moderator",
			expectedDomain: DOMAIN_MOD_CORRELATION,
		},
		{
			name:           "subforum owner still uses mod correlation domain",
			role:           "subforum_owner",
			expectedDomain: DOMAIN_MOD_CORRELATION,
		},
		{
			name:           "user uses user correlation domain",
			role:           "user",
			expectedDomain: DOMAIN_USER_CORRELATION,
		},
		{
			name:           "platform admin uses admin correlation domain",
			role:           "platform_admin",
			expectedDomain: DOMAIN_ADMIN_CORRELATION,
		},
		{
			name:           "trust safety uses admin correlation domain",
			role:           "trust_safety",
			expectedDomain: DOMAIN_ADMIN_CORRELATION,
		},
		{
			name:           "legal team uses legal correlation domain",
			role:           "legal_team",
			expectedDomain: DOMAIN_LEGAL_CORRELATION,
		},
		{
			name:           "invalid role defaults to user correlation domain",
			role:           "invalid_role",
			expectedDomain: DOMAIN_USER_CORRELATION,
		},
		{
			name:           "empty role defaults to user correlation domain",
			role:           "",
			expectedDomain: DOMAIN_USER_CORRELATION,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := selectDomain(tt.role)
			assert.Equal(t, tt.expectedDomain, result,
				"Role %s should map to domain %s", tt.role, tt.expectedDomain)
		})
	}
}

func TestGetDomainForRole_DemocraticRoles(t *testing.T) {
	// Create a mock IBE system for testing
	ibeSystem := &IBESystem{}

	tests := []struct {
		name           string
		role           string
		expectedDomain string
	}{
		{
			name:           "elected moderator domain mapping",
			role:           "elected_moderator",
			expectedDomain: DOMAIN_MOD_CORRELATION,
		},
		{
			name:           "appointed moderator domain mapping",
			role:           "appointed_moderator",
			expectedDomain: DOMAIN_MOD_CORRELATION,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ibeSystem.GetDomainForRole(tt.role)
			assert.Equal(t, tt.expectedDomain, result,
				"IBESystem.GetDomainForRole(%s) should return %s", tt.role, tt.expectedDomain)
		})
	}
}

func TestDemocraticRoles_DomainConsistency(t *testing.T) {
	// Test that all moderator-level roles use the same domain
	moderatorRoles := []string{
		"moderator",
		"subforum_owner",
		"elected_moderator",
		"appointed_moderator",
	}

	t.Run("all moderator roles use same domain", func(t *testing.T) {
		expectedDomain := DOMAIN_MOD_CORRELATION

		for _, role := range moderatorRoles {
			domain := selectDomain(role)
			assert.Equal(t, expectedDomain, domain,
				"Role %s should use domain %s like other moderator roles", role, expectedDomain)
		}
	})
}

func TestDemocraticRoles_DomainSeparation(t *testing.T) {
	// Test that democratic roles maintain proper domain separation
	tests := []struct {
		name           string
		role           string
		shouldNotEqual string
	}{
		{
			name:           "elected moderator not admin domain",
			role:           "elected_moderator",
			shouldNotEqual: DOMAIN_ADMIN_CORRELATION,
		},
		{
			name:           "elected moderator not legal domain",
			role:           "elected_moderator",
			shouldNotEqual: DOMAIN_LEGAL_CORRELATION,
		},
		{
			name:           "elected moderator not user domain",
			role:           "elected_moderator",
			shouldNotEqual: DOMAIN_USER_CORRELATION,
		},
		{
			name:           "appointed moderator not admin domain",
			role:           "appointed_moderator",
			shouldNotEqual: DOMAIN_ADMIN_CORRELATION,
		},
		{
			name:           "appointed moderator not legal domain",
			role:           "appointed_moderator",
			shouldNotEqual: DOMAIN_LEGAL_CORRELATION,
		},
		{
			name:           "appointed moderator not user domain",
			role:           "appointed_moderator",
			shouldNotEqual: DOMAIN_USER_CORRELATION,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain := selectDomain(tt.role)
			assert.NotEqual(t, tt.shouldNotEqual, domain,
				"Role %s should not use domain %s", tt.role, tt.shouldNotEqual)
		})
	}
}
