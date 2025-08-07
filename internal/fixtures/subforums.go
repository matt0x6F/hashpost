package fixtures

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob/types"
)

// CreateTestPrivateSubforum creates a test private subforum
func CreateTestPrivateSubforum() *models.Subforum {
	subforum := CreateTestSubforum()
	subforum.Name = "private-test-subforum"
	subforum.DisplayName = "Private Test Subforum"
	subforum.IsPrivate = sql.Null[bool]{V: true, Valid: true}
	return subforum
}

// CreateTestNSFWSubforum creates a test NSFW subforum
func CreateTestNSFWSubforum() *models.Subforum {
	subforum := CreateTestSubforum()
	subforum.Name = "nsfw-test-subforum"
	subforum.DisplayName = "NSFW Test Subforum"
	subforum.IsNSFW = sql.Null[bool]{V: true, Valid: true}
	return subforum
}

// CreateTestRestrictedSubforum creates a test restricted subforum
func CreateTestRestrictedSubforum() *models.Subforum {
	subforum := CreateTestSubforum()
	subforum.Name = "restricted-test-subforum"
	subforum.DisplayName = "Restricted Test Subforum"
	subforum.IsRestricted = sql.Null[bool]{V: true, Valid: true}
	return subforum
}

// CreateTestSubforumSubscription creates a test subforum subscription
func CreateTestSubforumSubscription() *models.SubforumSubscription {
	return &models.SubforumSubscription{
		SubscriptionID: 1,
		PseudonymID:    "test-pseudonym-id",
		SubforumID:     1,
		SubscribedAt:   sql.Null[time.Time]{V: time.Now(), Valid: true},
		IsFavorite:     sql.Null[bool]{V: false, Valid: true},
	}
}

// CreateTestFavoriteSubscription creates a test favorite subscription
func CreateTestFavoriteSubscription() *models.SubforumSubscription {
	subscription := CreateTestSubforumSubscription()
	subscription.IsFavorite = sql.Null[bool]{V: true, Valid: true}
	return subscription
}

// CreateTestModeratorRoleKey creates a test moderator role key
func CreateTestModeratorRoleKey() *models.RoleKey {
	return &models.RoleKey{
		RoleName:     "moderator",
		Scope:        "moderation",
		KeyData:      []byte{},
		KeyVersion:   1,
		Capabilities: types.JSON[json.RawMessage]{},
		CreatedAt:    sql.Null[time.Time]{V: time.Now(), Valid: true},
		ExpiresAt:    time.Now().AddDate(1, 0, 0),
		IsActive:     sql.Null[bool]{V: true, Valid: true},
		CreatedBy:    "admin-pseudonym-id",
		PseudonymID:  "moderator-pseudonym-id",
		SubforumID:   sql.Null[int32]{V: 1, Valid: true},
	}
}

// CreateTestAdminRoleKey creates a test admin role key
func CreateTestAdminRoleKey() *models.RoleKey {
	roleKey := CreateTestModeratorRoleKey()
	roleKey.RoleName = "platform_admin"
	return roleKey
}
