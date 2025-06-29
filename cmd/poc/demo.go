package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/matt0x6f/hashpost/internal/ibe"
	"github.com/matt0x6f/hashpost/internal/poc"
)

// Demo orchestrates the IBE proof of concept demonstration
type Demo struct {
	ibeSystem *ibe.IBESystem
	db        *poc.MemoryDB
}

// NewDemo creates a new demonstration instance
func NewDemo(ibeSystem *ibe.IBESystem, db *poc.MemoryDB) *Demo {
	return &Demo{
		ibeSystem: ibeSystem,
		db:        db,
	}
}

// Run executes the complete demonstration
func (demo *Demo) Run() {
	fmt.Println("🚀 Starting IBE Demonstration...")

	// Step 1: Setup administrative roles
	demo.setupAdminRoles()

	// Step 2: Register users with multiple pseudonyms
	userMap := demo.registerUsers()

	// Step 3: Create subforums
	demo.createSubforums()

	// Step 4: Users make posts
	demo.usersMakePosts(userMap)

	// Step 5: Demonstrate admin correlation of all pseudonyms for a user
	demo.showUserPseudonymCorrelation(userMap)

	// Step 6: Show audit trail
	demo.showAuditTrail()

	fmt.Println("✅ IBE Demonstration Complete!")
}

// setupAdminRoles creates the administrative role hierarchy
func (demo *Demo) setupAdminRoles() {
	fmt.Println("📋 Setting up Administrative Roles...")

	roles := []*poc.AdminRole{
		{
			RoleName:     "Site Administrator",
			Scope:        "full_correlation",
			Capabilities: []string{"correlate_any_user", "platform_wide_access"},
			ExpiresAt:    time.Now().AddDate(1, 0, 0),
		},
		{
			RoleName:     "Trust & Safety",
			Scope:        "harassment_investigation",
			Capabilities: []string{"correlate_reported_users", "cross_community_access"},
			ExpiresAt:    time.Now().AddDate(0, 3, 0),
		},
		{
			RoleName:     "Subforum Moderator",
			Scope:        "golang:local_correlation",
			Capabilities: []string{"correlate_local_users", "30_day_window"},
			ExpiresAt:    time.Now().AddDate(0, 1, 0),
		},
		{
			RoleName:     "Anti-Spam Team",
			Scope:        "network_analysis",
			Capabilities: []string{"correlate_48h_window", "automated_detection"},
			ExpiresAt:    time.Now().AddDate(0, 0, 7),
		},
	}

	for _, role := range roles {
		demo.db.CreateAdminRole(role)
		fmt.Printf("  ✅ Created role: %s (%s)\n", role.RoleName, role.Scope)
	}

	// Create admin users
	admins := []*poc.AdminUser{
		{Username: "admin_sarah", RoleID: roles[0].RoleID, IsActive: true},
		{Username: "trust_alex", RoleID: roles[1].RoleID, IsActive: true},
		{Username: "mod_john", RoleID: roles[2].RoleID, IsActive: true},
		{Username: "spam_bot", RoleID: roles[3].RoleID, IsActive: true},
	}

	for _, admin := range admins {
		demo.db.CreateAdminUser(admin)
		fmt.Printf("  ✅ Created admin: %s\n", admin.Username)
	}

	fmt.Println()
}

// registerUsers demonstrates user registration with multiple pseudonyms per real user
func (demo *Demo) registerUsers() map[string][]*poc.User {
	fmt.Println("👥 Registering Users with Multiple Pseudonyms...")

	realIdentities := []string{
		"alice@example.com",
		"bob@example.com",
		"charlie@example.com",
		"diana@example.com",
	}

	pseudonymsPerUser := 2
	userMap := make(map[string][]*poc.User)

	for i, realIdentity := range realIdentities {
		for j := 0; j < pseudonymsPerUser; j++ {
			// Generate user secret (in real system, this would be derived from password/2FA + pseudonym seed)
			userSecret := make([]byte, 32)
			rand.Read(userSecret)

			// Generate pseudonym using IBE
			pseudonymID := demo.ibeSystem.GeneratePseudonym(userSecret)

			// Create user profile
			user := &poc.User{
				PseudonymID: pseudonymID,
				DisplayName: fmt.Sprintf("user_%d_%d", i+1, j+1),
				KarmaScore:  0,
				CreatedAt:   time.Now(),
			}

			// Store user
			demo.db.CreateUser(user)

			// Encrypt and store identity mapping
			adminKey := demo.ibeSystem.GenerateRoleKey("site_admin", "full_correlation", time.Now().AddDate(1, 0, 0))
			encryptedMapping, _ := demo.ibeSystem.EncryptIdentity(realIdentity, pseudonymID, adminKey)

			mapping := &poc.IdentityMapping{
				EncryptedRealIdentity: encryptedMapping,
				KeyVersion:            1,
				CreatedAt:             time.Now(),
			}

			demo.db.StoreIdentityMapping(mapping)

			fmt.Printf("  ✅ User %s registered with pseudonym: %s\n", realIdentity, pseudonymID[:8]+"...")
			userMap[realIdentity] = append(userMap[realIdentity], user)
		}
	}

	fmt.Println()
	return userMap
}

// createSubforums sets up community subforums
func (demo *Demo) createSubforums() {
	fmt.Println("🏛️  Creating Subforums...")

	subforums := []*poc.Subforum{
		{
			SubforumID:  1,
			Name:        "golang",
			Description: "Go programming language discussions",
			OwnerID:     "mod_john",
			CreatedAt:   time.Now(),
		},
		{
			SubforumID:  2,
			Name:        "privacy",
			Description: "Privacy and security discussions",
			OwnerID:     "trust_alex",
			CreatedAt:   time.Now(),
		},
	}

	for _, subforum := range subforums {
		demo.db.CreateSubforum(subforum)
		fmt.Printf("  ✅ Created subforum: r/%s\n", subforum.Name)
	}

	fmt.Println()
}

// usersMakePosts demonstrates users creating content
func (demo *Demo) usersMakePosts(userMap map[string][]*poc.User) {
	fmt.Println("📝 Users Creating Posts...")

	posts := []struct {
		realIdentity string
		pseudonymIdx int
		subforumID   int
		title        string
		content      string
	}{
		{"alice@example.com", 0, 1, "How to implement IBE in Go?", "I'm working on a privacy-focused social platform..."},
		{"alice@example.com", 1, 2, "Throwaway: Privacy Q", "Does anyone use throwaway accounts for privacy?"},
		{"bob@example.com", 0, 1, "Best practices for pseudonymous systems", "Here are some lessons learned from building..."},
		{"bob@example.com", 1, 2, "Throwaway: Anonymous feedback", "How do you handle anonymous feedback?"},
		{"charlie@example.com", 0, 2, "Privacy concerns with social media", "What do you think about the current state..."},
		{"diana@example.com", 0, 1, "Go crypto libraries recommendation", "Which libraries would you recommend for..."},
	}

	for i, postData := range posts {
		user := userMap[postData.realIdentity][postData.pseudonymIdx]
		post := &poc.Post{
			PseudonymID: user.PseudonymID,
			SubforumID:  postData.subforumID,
			Title:       postData.title,
			Content:     postData.content,
			CreatedAt:   time.Now().Add(time.Duration(i) * time.Minute),
		}

		demo.db.CreatePost(post)
		fmt.Printf("  ✅ Post %d: %s (by %s)\n", i+1, postData.title, user.DisplayName)
	}

	fmt.Println()
}

// showUserPseudonymCorrelation demonstrates admin correlation of all pseudonyms for a real user
func (demo *Demo) showUserPseudonymCorrelation(userMap map[string][]*poc.User) {
	fmt.Println("🔗 Demonstrating Admin Correlation of Multiple Pseudonyms per User...")

	for realIdentity, pseudonymUsers := range userMap {
		// Generate fingerprint for this real identity
		fingerprint := demo.ibeSystem.GenerateFingerprint(realIdentity)

		fmt.Printf("\n  Real Identity: %s\n", realIdentity)
		fmt.Printf("  Fingerprint: %s\n", fingerprint)
		fmt.Printf("  Associated Pseudonyms:\n")

		for _, user := range pseudonymUsers {
			// In a real system, admin would decrypt the mapping to get the fingerprint
			// and then use the fingerprint for lookups
			fmt.Printf("    - Pseudonym: %s\n", user.PseudonymID)
		}

		fmt.Printf("  → Admin can correlate all pseudonyms using fingerprint: %s\n", fingerprint)
	}
	fmt.Println()
}

// showAuditTrail displays the correlation audit log
func (demo *Demo) showAuditTrail() {
	fmt.Println("📊 Correlation Audit Trail...")

	audits, _ := demo.db.GetCorrelationAudits()

	fmt.Printf("  Total correlation requests: %d\n\n", len(audits))

	for i, audit := range audits {
		fmt.Printf("  Audit Entry %d:\n", i+1)
		fmt.Printf("    Admin: %s (%s)\n", audit.AdminUserID, audit.AdminRole)
		fmt.Printf("    Target: %s\n", audit.RequestedPseudonym[:8]+"...")
		fmt.Printf("    Justification: %s\n", audit.Justification)
		fmt.Printf("    Result: %s\n", string(audit.CorrelationResult))
		fmt.Printf("    Timestamp: %s\n", audit.Timestamp.Format("2006-01-02 15:04:05"))
		fmt.Printf("    Legal Basis: %s\n", audit.LegalBasis)
		fmt.Println()
	}
}

// generateUserSecret creates a deterministic user secret for demo purposes
func generateUserSecret(email string) []byte {
	hash := sha256.Sum256([]byte(email))
	return hash[:]
}
