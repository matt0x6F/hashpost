package poc

import (
	"time"
)

// User represents a pseudonymous user profile
type User struct {
	PseudonymID string    `json:"pseudonym_id"`
	DisplayName string    `json:"display_name"`
	KarmaScore  int       `json:"karma_score"`
	CreatedAt   time.Time `json:"created_at"`
}

// Post represents a post made by a pseudonymous user
type Post struct {
	PostID      int64     `json:"post_id"`
	PseudonymID string    `json:"pseudonym_id"`
	SubforumID  int       `json:"subforum_id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}

// IdentityMapping represents the encrypted mapping between real identity and pseudonym
type IdentityMapping struct {
	MappingID             string    `json:"mapping_id"`
	EncryptedRealIdentity []byte    `json:"encrypted_real_identity"`
	EncryptedPseudonymMap []byte    `json:"encrypted_pseudonym_map"`
	KeyVersion            int       `json:"key_version"`
	CreatedAt             time.Time `json:"created_at"`
}

// AdminRole represents an administrative role with specific permissions
type AdminRole struct {
	RoleID       string    `json:"role_id"`
	RoleName     string    `json:"role_name"`
	Scope        string    `json:"scope"`
	Capabilities []string  `json:"capabilities"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// CorrelationAudit represents an audit log entry for correlation activities
type CorrelationAudit struct {
	AuditID            string    `json:"audit_id"`
	AdminUserID        string    `json:"admin_user_id"`
	AdminRole          string    `json:"admin_role"`
	RequestedPseudonym string    `json:"requested_pseudonym"`
	Justification      string    `json:"justification"`
	ApprovedBy         string    `json:"approved_by"`
	CorrelationResult  []byte    `json:"correlation_result"`
	Timestamp          time.Time `json:"timestamp"`
	LegalBasis         string    `json:"legal_basis"`
}

// Subforum represents a community subforum
type Subforum struct {
	SubforumID  int       `json:"subforum_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// AdminUser represents an administrative user
type AdminUser struct {
	AdminID   string    `json:"admin_id"`
	Username  string    `json:"username"`
	RoleID    string    `json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
	IsActive  bool      `json:"is_active"`
}
