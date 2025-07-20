package fixtures

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob/types"
)

// CreateTestReport creates a test report
func CreateTestReport() *models.Report {
	contentID := int64(123)
	reportedPseudonymID := "reported-pseudonym-id"

	return &models.Report{
		ReportID:              789,
		ReporterPseudonymID:   "reporter-pseudonym-id",
		ContentType:           "post",
		ContentID:             sql.Null[int64]{V: contentID, Valid: true},
		ReportedPseudonymID:   sql.Null[string]{V: reportedPseudonymID, Valid: true},
		ReportReason:          "spam",
		ReportDetails:         sql.Null[string]{V: "This post violates community guidelines...", Valid: true},
		Status:                sql.Null[string]{V: "pending", Valid: true},
		CreatedAt:             sql.Null[time.Time]{V: time.Now(), Valid: true},
		ResolvedByUserID:      sql.Null[int64]{Valid: false},
		ResolvedByPseudonymID: sql.Null[string]{Valid: false},
		ResolutionNotes:       sql.Null[string]{Valid: false},
		ResolvedAt:            sql.Null[time.Time]{Valid: false},
	}
}

// CreateTestReportWithResolution creates a test report that has been resolved
func CreateTestReportWithResolution() *models.Report {
	contentID := int64(123)
	reportedPseudonymID := "reported-pseudonym-id"
	resolverUserID := int64(456)
	resolverPseudonymID := "moderator-pseudonym-id"

	return &models.Report{
		ReportID:              789,
		ReporterPseudonymID:   "reporter-pseudonym-id",
		ContentType:           "post",
		ContentID:             sql.Null[int64]{V: contentID, Valid: true},
		ReportedPseudonymID:   sql.Null[string]{V: reportedPseudonymID, Valid: true},
		ReportReason:          "spam",
		ReportDetails:         sql.Null[string]{V: "This post violates community guidelines...", Valid: true},
		Status:                sql.Null[string]{V: "resolved", Valid: true},
		CreatedAt:             sql.Null[time.Time]{V: time.Now().Add(-time.Hour), Valid: true},
		ResolvedByUserID:      sql.Null[int64]{V: resolverUserID, Valid: true},
		ResolvedByPseudonymID: sql.Null[string]{V: resolverPseudonymID, Valid: true},
		ResolutionNotes:       sql.Null[string]{V: "Content removed for violating community guidelines", Valid: true},
		ResolvedAt:            sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// CreateTestModerationAction creates a test moderation action
func CreateTestModerationAction() *models.ModerationAction {
	subforumID := int32(1)
	targetContentID := int64(123)

	return &models.ModerationAction{
		ActionID:             123,
		ModeratorUserID:      456,
		ModeratorPseudonymID: "moderator-pseudonym-id",
		SubforumID:           sql.Null[int32]{V: subforumID, Valid: true},
		ActionType:           "remove_post",
		TargetContentType:    sql.Null[string]{V: "post", Valid: true},
		TargetContentID:      sql.Null[int64]{V: targetContentID, Valid: true},
		TargetUserID:         sql.Null[int64]{Valid: false},
		ActionDetails:        sql.Null[types.JSON[json.RawMessage]]{Valid: false},
		CreatedAt:            sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// CreateTestModerationActionWithDetails creates a test moderation action with action details
func CreateTestModerationActionWithDetails() *models.ModerationAction {
	subforumID := int32(1)
	targetContentID := int64(123)
	actionDetails := types.NewJSON[json.RawMessage]([]byte(`{"removal_reason": "violates community guidelines"}`))

	return &models.ModerationAction{
		ActionID:             123,
		ModeratorUserID:      456,
		ModeratorPseudonymID: "moderator-pseudonym-id",
		SubforumID:           sql.Null[int32]{V: subforumID, Valid: true},
		ActionType:           "remove_post",
		TargetContentType:    sql.Null[string]{V: "post", Valid: true},
		TargetContentID:      sql.Null[int64]{V: targetContentID, Valid: true},
		TargetUserID:         sql.Null[int64]{Valid: false},
		ActionDetails:        sql.Null[types.JSON[json.RawMessage]]{V: actionDetails, Valid: true},
		CreatedAt:            sql.Null[time.Time]{V: time.Now(), Valid: true},
	}
}

// CreateTestUserBan creates a test user ban
func CreateTestUserBan() *models.UserBan {
	return &models.UserBan{
		BanID:               123,
		SubforumID:          1,
		BannedUserID:        789,
		BannedByUserID:      456,
		BannedByPseudonymID: "moderator-pseudonym-id",
		BanReason:           "repeated violations",
		IsPermanent:         sql.Null[bool]{V: false, Valid: true},
		ExpiresAt:           sql.Null[time.Time]{V: time.Now().AddDate(0, 1, 0), Valid: true}, // 1 month
		CreatedAt:           sql.Null[time.Time]{V: time.Now(), Valid: true},
		IsActive:            sql.Null[bool]{V: true, Valid: true},
	}
}

// CreateTestPermanentUserBan creates a test permanent user ban
func CreateTestPermanentUserBan() *models.UserBan {
	return &models.UserBan{
		BanID:               124,
		SubforumID:          1,
		BannedUserID:        790,
		BannedByUserID:      456,
		BannedByPseudonymID: "moderator-pseudonym-id",
		BanReason:           "severe violations",
		IsPermanent:         sql.Null[bool]{V: true, Valid: true},
		ExpiresAt:           sql.Null[time.Time]{Valid: false},
		CreatedAt:           sql.Null[time.Time]{V: time.Now(), Valid: true},
		IsActive:            sql.Null[bool]{V: true, Valid: true},
	}
}

// CreateTestInactiveUserBan creates a test inactive user ban
func CreateTestInactiveUserBan() *models.UserBan {
	return &models.UserBan{
		BanID:               125,
		SubforumID:          1,
		BannedUserID:        791,
		BannedByUserID:      456,
		BannedByPseudonymID: "moderator-pseudonym-id",
		BanReason:           "temporary ban",
		IsPermanent:         sql.Null[bool]{V: false, Valid: true},
		ExpiresAt:           sql.Null[time.Time]{V: time.Now().AddDate(0, 0, 7), Valid: true},  // 1 week
		CreatedAt:           sql.Null[time.Time]{V: time.Now().AddDate(0, 0, -7), Valid: true}, // 1 week ago
		IsActive:            sql.Null[bool]{V: false, Valid: true},
	}
}
