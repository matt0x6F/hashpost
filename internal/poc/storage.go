package poc

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
)

// MemoryDB provides in-memory storage for the proof of concept
type MemoryDB struct {
	mu sync.RWMutex

	users             map[string]*User
	posts             map[int64]*Post
	identityMappings  map[string]*IdentityMapping
	adminUsers        map[string]*AdminUser
	adminRoles        map[string]*AdminRole
	correlationAudits map[string]*CorrelationAudit
	subforums         map[int]*Subforum

	nextPostID int64
}

// NewMemoryDB creates a new in-memory database
func NewMemoryDB() *MemoryDB {
	return &MemoryDB{
		users:             make(map[string]*User),
		posts:             make(map[int64]*Post),
		identityMappings:  make(map[string]*IdentityMapping),
		adminUsers:        make(map[string]*AdminUser),
		adminRoles:        make(map[string]*AdminRole),
		correlationAudits: make(map[string]*CorrelationAudit),
		subforums:         make(map[int]*Subforum),
		nextPostID:        1,
	}
}

// CreateUser stores a new pseudonymous user
func (db *MemoryDB) CreateUser(user *User) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.users[user.PseudonymID] = user
	return nil
}

// GetUser retrieves a user by pseudonym ID
func (db *MemoryDB) GetUser(pseudonymID string) (*User, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	user, exists := db.users[pseudonymID]
	if !exists {
		return nil, fmt.Errorf("user not found: %s", pseudonymID)
	}
	return user, nil
}

// CreatePost stores a new post
func (db *MemoryDB) CreatePost(post *Post) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	post.PostID = db.nextPostID
	db.nextPostID++

	db.posts[post.PostID] = post
	return nil
}

// GetPostsByPseudonym retrieves all posts by a pseudonym
func (db *MemoryDB) GetPostsByPseudonym(pseudonymID string) ([]*Post, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var posts []*Post
	for _, post := range db.posts {
		if post.PseudonymID == pseudonymID {
			posts = append(posts, post)
		}
	}
	return posts, nil
}

// StoreIdentityMapping stores an encrypted identity mapping
func (db *MemoryDB) StoreIdentityMapping(mapping *IdentityMapping) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if mapping.MappingID == "" {
		mapping.MappingID = generateID()
	}

	db.identityMappings[mapping.MappingID] = mapping
	return nil
}

// GetIdentityMapping retrieves an identity mapping by ID
func (db *MemoryDB) GetIdentityMapping(mappingID string) (*IdentityMapping, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	mapping, exists := db.identityMappings[mappingID]
	if !exists {
		return nil, fmt.Errorf("identity mapping not found: %s", mappingID)
	}
	return mapping, nil
}

// CreateAdminUser stores a new admin user
func (db *MemoryDB) CreateAdminUser(admin *AdminUser) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if admin.AdminID == "" {
		admin.AdminID = generateID()
	}

	db.adminUsers[admin.AdminID] = admin
	return nil
}

// GetAdminUser retrieves an admin user by ID
func (db *MemoryDB) GetAdminUser(adminID string) (*AdminUser, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	admin, exists := db.adminUsers[adminID]
	if !exists {
		return nil, fmt.Errorf("admin user not found: %s", adminID)
	}
	return admin, nil
}

// CreateAdminRole stores a new admin role
func (db *MemoryDB) CreateAdminRole(role *AdminRole) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if role.RoleID == "" {
		role.RoleID = generateID()
	}

	db.adminRoles[role.RoleID] = role
	return nil
}

// GetAdminRole retrieves an admin role by ID
func (db *MemoryDB) GetAdminRole(roleID string) (*AdminRole, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	role, exists := db.adminRoles[roleID]
	if !exists {
		return nil, fmt.Errorf("admin role not found: %s", roleID)
	}
	return role, nil
}

// LogCorrelationAudit stores a correlation audit entry
func (db *MemoryDB) LogCorrelationAudit(audit *CorrelationAudit) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	if audit.AuditID == "" {
		audit.AuditID = generateID()
	}

	db.correlationAudits[audit.AuditID] = audit
	return nil
}

// GetCorrelationAudits retrieves all correlation audits
func (db *MemoryDB) GetCorrelationAudits() ([]*CorrelationAudit, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var audits []*CorrelationAudit
	for _, audit := range db.correlationAudits {
		audits = append(audits, audit)
	}
	return audits, nil
}

// CreateSubforum stores a new subforum
func (db *MemoryDB) CreateSubforum(subforum *Subforum) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.subforums[subforum.SubforumID] = subforum
	return nil
}

// GetSubforum retrieves a subforum by ID
func (db *MemoryDB) GetSubforum(subforumID int) (*Subforum, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	subforum, exists := db.subforums[subforumID]
	if !exists {
		return nil, fmt.Errorf("subforum not found: %d", subforumID)
	}
	return subforum, nil
}

// generateID creates a random ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
