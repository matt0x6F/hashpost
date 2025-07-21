package constants

// Capabilities define the specific operations that can be performed within each scope.
// These are used to control access to different system features and data.

// Authentication capabilities - used for basic user operations
const (
	// AccessOwnPseudonyms allows users to access their own pseudonyms
	CapabilityAccessOwnPseudonyms = "access_own_pseudonyms"

	// Login allows users to authenticate and create sessions
	CapabilityLogin = "login"

	// SessionManagement allows users to manage their active sessions
	CapabilitySessionManagement = "session_management"
)

// Self-correlation capabilities - used for users to manage their own identity
const (
	// VerifyOwnPseudonymOwnership allows users to verify they own a pseudonym
	CapabilityVerifyOwnPseudonymOwnership = "verify_own_pseudonym_ownership"

	// ManageOwnProfile allows users to manage their own profile information
	CapabilityManageOwnProfile = "manage_own_profile"

	// ManageOwnPseudonyms allows users to manage their own pseudonyms (create, deactivate)
	CapabilityManageOwnPseudonyms = "manage_own_pseudonyms"
)

// Correlation capabilities - used for administrative identity correlation
const (
	// AccessAllPseudonyms allows access to all pseudonyms in the system
	CapabilityAccessAllPseudonyms = "access_all_pseudonyms"

	// AccessSubforumPseudonyms allows access to pseudonyms within a specific subforum
	CapabilityAccessSubforumPseudonyms = "access_subforum_pseudonyms"

	// CrossUserCorrelation allows correlating identities across different users
	CapabilityCrossUserCorrelation = "cross_user_correlation"

	// CorrelateFingerprints allows correlating fingerprints within a subforum
	CapabilityCorrelateFingerprints = "correlate_fingerprints"
)

// Moderation capabilities - used for content and user moderation
const (
	// ModerateContent allows moderating content within a subforum
	CapabilityModerateContent = "moderate_content"

	// BanUsers allows banning users from a subforum
	CapabilityBanUsers = "ban_users"

	// RemoveContent allows removing content from a subforum
	CapabilityRemoveContent = "remove_content"

	// ManageModerators allows managing moderator assignments
	CapabilityManageModerators = "manage_moderators"
)

// Platform-wide capabilities - used for system administration
const (
	// Moderation allows platform-wide moderation operations
	CapabilityModeration = "moderation"

	// Compliance allows compliance-related operations
	CapabilityCompliance = "compliance"

	// LegalRequests allows handling legal requests
	CapabilityLegalRequests = "legal_requests"

	// SystemAdmin allows full system administration
	CapabilitySystemAdmin = "system_admin"

	// UserManagement allows managing user accounts
	CapabilityUserManagement = "user_management"
)

// Basic user capabilities - available to all users
const (
	// CreateContent allows creating posts and comments
	CapabilityCreateContent = "create_content"

	// Vote allows voting on posts and comments
	CapabilityVote = "vote"

	// Message allows sending direct messages
	CapabilityMessage = "message"

	// Report allows reporting content or users
	CapabilityReport = "report"

	// CreateSubforum allows creating new subforums
	CapabilityCreateSubforum = "create_subforum"
)

// GetCapabilitiesByScope returns the capabilities available for a given scope
func GetCapabilitiesByScope(scope string) []string {
	switch scope {
	case ScopeAuthentication:
		return []string{
			CapabilityAccessOwnPseudonyms,
			CapabilityLogin,
			CapabilitySessionManagement,
		}
	case ScopeSelfCorrelation:
		return []string{
			CapabilityVerifyOwnPseudonymOwnership,
			CapabilityManageOwnProfile,
			CapabilityManageOwnPseudonyms,
		}
	case ScopeCorrelation:
		return []string{
			CapabilityAccessAllPseudonyms,
			CapabilityAccessSubforumPseudonyms,
			CapabilityCrossUserCorrelation,
			CapabilityCorrelateFingerprints,
			CapabilityModerateContent,
			CapabilityBanUsers,
			CapabilityRemoveContent,
			CapabilityManageModerators,
			CapabilityModeration,
			CapabilityCompliance,
			CapabilityLegalRequests,
			CapabilitySystemAdmin,
			CapabilityUserManagement,
		}
	default:
		return []string{}
	}
}

// IsValidCapability checks if the given capability is valid for the specified scope
func IsValidCapability(capability, scope string) bool {
	validCapabilities := GetCapabilitiesByScope(scope)
	for _, validCap := range validCapabilities {
		if validCap == capability {
			return true
		}
	}
	return false
}

// GetAllCapabilities returns all available capabilities in the system
func GetAllCapabilities() []string {
	return []string{
		// Authentication capabilities
		CapabilityAccessOwnPseudonyms,
		CapabilityLogin,
		CapabilitySessionManagement,

		// Self-correlation capabilities
		CapabilityVerifyOwnPseudonymOwnership,
		CapabilityManageOwnProfile,
		CapabilityManageOwnPseudonyms,

		// Correlation capabilities
		CapabilityAccessAllPseudonyms,
		CapabilityAccessSubforumPseudonyms,
		CapabilityCrossUserCorrelation,
		CapabilityCorrelateFingerprints,

		// Moderation capabilities
		CapabilityModerateContent,
		CapabilityBanUsers,
		CapabilityRemoveContent,
		CapabilityManageModerators,

		// Platform-wide capabilities
		CapabilityModeration,
		CapabilityCompliance,
		CapabilityLegalRequests,
		CapabilitySystemAdmin,
		CapabilityUserManagement,

		// Basic user capabilities
		CapabilityCreateContent,
		CapabilityVote,
		CapabilityMessage,
		CapabilityReport,
		CapabilityCreateSubforum,
	}
}
