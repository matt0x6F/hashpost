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

	// Correlation scope is used by administrative roles (moderators, admins) to correlate identities
	// across different pseudonyms and access sensitive data. This is the highest privilege scope.
	ScopeCorrelation = "correlation"
)

// AllScopes returns all available scopes in the system.
func AllScopes() []string {
	return []string{
		ScopeAuthentication,
		ScopeSelfCorrelation,
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
