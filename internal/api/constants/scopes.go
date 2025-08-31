package constants

// Scopes define the different operational contexts for IBE keys and role-based access control.
// Each scope represents a specific domain of operations that require different levels of access.
const (
	// Authentication scope is used for general user authentication and accessing their own pseudonyms.
	// This is the most basic scope that all users have access to.
	ScopeAuthentication = "authentication"

	// SelfCorrelation scope is used by users to verify their own pseudonym ownership and manage their profile.
	// This allows users to access and manage their own identity mappings.
	ScopeSelfCorrelation = "self_correlation"

	// Messaging scope is used for user-to-user direct messaging operations.
	// This includes sending, receiving, and managing encrypted messages.
	ScopeMessaging = "messaging"

	// Moderation scope is used for content moderation, report reviews, and rule management.
	// This includes subforum-level and platform-level moderation operations.
	ScopeModeration = "moderation"

	// Administration scope is used for platform administration and user management.
	// This includes system administration, user management, and platform-wide operations.
	ScopeAdministration = "administration"

	// Correlation scope is used by administrative roles to correlate identities across different pseudonyms.
	// This is the highest privilege scope for identity correlation operations.
	ScopeCorrelation = "correlation"
)

// AllScopes returns all available scopes in the system.
func AllScopes() []string {
	return []string{
		ScopeAuthentication,
		ScopeSelfCorrelation,
		ScopeMessaging,
		ScopeModeration,
		ScopeAdministration,
		ScopeCorrelation,
	}
}

// IsValidScope checks if the given scope is valid.
func IsValidScope(scope string) bool {
	for _, validScope := range AllScopes() {
		if validScope == scope {
			return true
		}
	}
	return false
}

// GetScopeDescription returns a human-readable description of each scope.
func GetScopeDescription(scope string) string {
	switch scope {
	case ScopeAuthentication:
		return "Basic user authentication and session management"
	case ScopeSelfCorrelation:
		return "User management of their own identity and pseudonyms"
	case ScopeModeration:
		return "Content moderation, report reviews, and rule management"
	case ScopeAdministration:
		return "Platform administration and user management"
	case ScopeCorrelation:
		return "Cross-user identity correlation (admin only)"
	default:
		return "Unknown scope"
	}
}
