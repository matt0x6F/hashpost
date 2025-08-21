package constants

// Roles define the different user roles in the system.
// Each role has specific capabilities and access levels.
const (
	// RoleUser is the basic user role with standard platform access
	RoleUser = "user"

	// RoleModerator is for subforum moderators with content moderation capabilities
	RoleModerator = "moderator"

	// RoleSubforumOwner is for subforum owners with full subforum management
	RoleSubforumOwner = "subforum_owner"

	// RoleElectedModerator is for community-elected moderators in democratic subforums
	RoleElectedModerator = "elected_moderator"

	// RoleAppointedModerator is for platform-appointed moderators during crisis management
	RoleAppointedModerator = "appointed_moderator"

	// RolePlatformAdmin is for platform administrators with full system access
	RolePlatformAdmin = "platform_admin"

	// RoleTrustSafety is for trust and safety team members
	RoleTrustSafety = "trust_safety"

	// RoleLegalTeam is for legal team members with compliance access
	RoleLegalTeam = "legal_team"
)

// RoleDefinition defines the structure for role configuration
type RoleDefinition struct {
	RoleName     string
	Scopes       []string
	Capabilities map[string][]string
}

// GetRoleDefinitions returns all role definitions with their scopes and capabilities
func GetRoleDefinitions() []RoleDefinition {
	return []RoleDefinition{
		{
			RoleName: RoleUser,
			Scopes:   []string{ScopeAuthentication, ScopeSelfCorrelation},
			Capabilities: map[string][]string{
				ScopeAuthentication: {
					CapabilityAccessOwnPseudonyms,
					CapabilityLogin,
					CapabilitySessionManagement,
				},
				ScopeSelfCorrelation: {
					CapabilityVerifyOwnPseudonymOwnership,
					CapabilityManageOwnProfile,
				},
			},
		},
		{
			RoleName: RoleModerator,
			Scopes:   []string{ScopeAuthentication, ScopeSelfCorrelation, ScopeModeration, ScopeCorrelation},
			Capabilities: map[string][]string{
				ScopeAuthentication: {
					CapabilityAccessOwnPseudonyms,
					CapabilityLogin,
					CapabilitySessionManagement,
				},
				ScopeSelfCorrelation: {
					CapabilityVerifyOwnPseudonymOwnership,
					CapabilityManageOwnProfile,
				},
				ScopeModeration: {
					CapabilityModerateContent,
					CapabilityBanUsers,
					CapabilityRemoveContent,
					CapabilityReviewReports,
					CapabilityForwardReports,
					CapabilityManageSubforumRules,
					CapabilityManageSubforumSettings,
				},
				ScopeCorrelation: {
					CapabilityAccessSubforumPseudonyms,
					CapabilityCorrelateFingerprints,
				},
			},
		},
		{
			RoleName: RoleSubforumOwner,
			Scopes:   []string{ScopeAuthentication, ScopeSelfCorrelation, ScopeModeration, ScopeCorrelation},
			Capabilities: map[string][]string{
				ScopeAuthentication: {
					CapabilityAccessOwnPseudonyms,
					CapabilityLogin,
					CapabilitySessionManagement,
				},
				ScopeSelfCorrelation: {
					CapabilityVerifyOwnPseudonymOwnership,
					CapabilityManageOwnProfile,
				},
				ScopeModeration: {
					CapabilityModerateContent,
					CapabilityBanUsers,
					CapabilityRemoveContent,
					CapabilityManageModerators,
					CapabilityReviewReports,
					CapabilityForwardReports,
					CapabilityManageSubforumRules,
					CapabilityManageSubforumSettings,
					CapabilityStickyPost,
					CapabilityLockPost,
				},
				ScopeCorrelation: {
					CapabilityAccessSubforumPseudonyms,
					CapabilityCorrelateFingerprints,
				},
			},
		},
		{
			RoleName: RoleElectedModerator,
			Scopes:   []string{ScopeAuthentication, ScopeSelfCorrelation, ScopeModeration, ScopeCorrelation},
			Capabilities: map[string][]string{
				ScopeAuthentication: {
					CapabilityAccessOwnPseudonyms,
					CapabilityLogin,
					CapabilitySessionManagement,
				},
				ScopeSelfCorrelation: {
					CapabilityVerifyOwnPseudonymOwnership,
					CapabilityManageOwnProfile,
				},
				ScopeModeration: {
					CapabilityModerateContent,
					CapabilityBanUsers,
					CapabilityRemoveContent,
					CapabilityManageModerators,
					CapabilityReviewReports,
					CapabilityForwardReports,
					CapabilityManageSubforumRules,
					CapabilityManageSubforumSettings,
					CapabilityStickyPost,
					CapabilityLockPost,
				},
				ScopeCorrelation: {
					CapabilityAccessSubforumPseudonyms,
					CapabilityCorrelateFingerprints,
				},
			},
		},
		{
			RoleName: RoleAppointedModerator,
			Scopes:   []string{ScopeAuthentication, ScopeSelfCorrelation, ScopeModeration, ScopeCorrelation},
			Capabilities: map[string][]string{
				ScopeAuthentication: {
					CapabilityAccessOwnPseudonyms,
					CapabilityLogin,
					CapabilitySessionManagement,
				},
				ScopeSelfCorrelation: {
					CapabilityVerifyOwnPseudonymOwnership,
					CapabilityManageOwnProfile,
				},
				ScopeModeration: {
					CapabilityModerateContent,
					CapabilityBanUsers,
					CapabilityRemoveContent,
					CapabilityManageModerators,
					CapabilityReviewReports,
					CapabilityForwardReports,
					CapabilityManageSubforumRules,
					CapabilityManageSubforumSettings,
					CapabilityStickyPost,
					CapabilityLockPost,
				},
				ScopeCorrelation: {
					CapabilityAccessSubforumPseudonyms,
					CapabilityCorrelateFingerprints,
				},
			},
		},
		{
			RoleName: RolePlatformAdmin,
			Scopes:   []string{ScopeAuthentication, ScopeSelfCorrelation, ScopeAdministration, ScopeCorrelation},
			Capabilities: map[string][]string{
				ScopeAuthentication: {
					CapabilityAccessOwnPseudonyms,
					CapabilityLogin,
					CapabilitySessionManagement,
				},
				ScopeSelfCorrelation: {
					CapabilityVerifyOwnPseudonymOwnership,
					CapabilityManageOwnProfile,
				},
				ScopeAdministration: {
					CapabilityModeration,
					CapabilityCompliance,
					CapabilityLegalRequests,
					CapabilitySystemAdmin,
					CapabilityUserManagement,
				},
				ScopeCorrelation: {
					CapabilityAccessAllPseudonyms,
					CapabilityAccessSubforumPseudonyms,
					CapabilityCrossUserCorrelation,
					CapabilityCorrelateFingerprints,
				},
			},
		},
		{
			RoleName: RoleTrustSafety,
			Scopes:   []string{ScopeAuthentication, ScopeSelfCorrelation, ScopeAdministration, ScopeCorrelation},
			Capabilities: map[string][]string{
				ScopeAuthentication: {
					CapabilityAccessOwnPseudonyms,
					CapabilityLogin,
					CapabilitySessionManagement,
				},
				ScopeSelfCorrelation: {
					CapabilityVerifyOwnPseudonymOwnership,
					CapabilityManageOwnProfile,
				},
				ScopeAdministration: {
					CapabilityModeration,
					CapabilityCompliance,
					CapabilityLegalRequests,
				},
				ScopeCorrelation: {
					CapabilityAccessAllPseudonyms,
					CapabilityAccessSubforumPseudonyms,
					CapabilityCrossUserCorrelation,
					CapabilityCorrelateFingerprints,
				},
			},
		},
		{
			RoleName: RoleLegalTeam,
			Scopes:   []string{ScopeAuthentication, ScopeSelfCorrelation, ScopeAdministration, ScopeCorrelation},
			Capabilities: map[string][]string{
				ScopeAuthentication: {
					CapabilityAccessOwnPseudonyms,
					CapabilityLogin,
					CapabilitySessionManagement,
				},
				ScopeSelfCorrelation: {
					CapabilityVerifyOwnPseudonymOwnership,
					CapabilityManageOwnProfile,
				},
				ScopeAdministration: {
					CapabilityCompliance,
					CapabilityLegalRequests,
				},
				ScopeCorrelation: {
					CapabilityAccessAllPseudonyms,
					CapabilityAccessSubforumPseudonyms,
					CapabilityCrossUserCorrelation,
					CapabilityCorrelateFingerprints,
				},
			},
		},
	}
}

// GetRoleDefinition returns the role definition for a specific role
func GetRoleDefinition(roleName string) *RoleDefinition {
	definitions := GetRoleDefinitions()
	for _, def := range definitions {
		if def.RoleName == roleName {
			return &def
		}
	}
	return nil
}

// GetRoleCapabilities returns all capabilities for a specific role
func GetRoleCapabilities(roleName string) []string {
	def := GetRoleDefinition(roleName)
	if def == nil {
		return []string{}
	}

	var capabilities []string
	for _, scope := range def.Scopes {
		if caps, exists := def.Capabilities[scope]; exists {
			capabilities = append(capabilities, caps...)
		}
	}
	return capabilities
}

// GetRoleScopes returns all scopes for a specific role
func GetRoleScopes(roleName string) []string {
	def := GetRoleDefinition(roleName)
	if def == nil {
		return []string{}
	}
	return def.Scopes
}

// IsValidRole checks if the given role is valid
func IsValidRole(roleName string) bool {
	return GetRoleDefinition(roleName) != nil
}

// GetAllRoles returns all available roles in the system
func GetAllRoles() []string {
	definitions := GetRoleDefinitions()
	roles := make([]string, len(definitions))
	for i, def := range definitions {
		roles[i] = def.RoleName
	}
	return roles
}
