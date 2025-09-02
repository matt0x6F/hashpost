package dao

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/rs/zerolog/log"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql/sm"
)

// ReportDAO implements ReportDAOInterface
type ReportDAO struct {
	db bob.Executor
}

// NewReportDAO creates a new ReportDAO
func NewReportDAO(db bob.Executor) *ReportDAO {
	return &ReportDAO{db: db}
}

// CreateReport creates a new report
func (dao *ReportDAO) CreateReport(ctx context.Context, report *models.ReportSetter) (*models.Report, error) {
	log.Debug().
		Str("reporter_pseudonym_id", *report.ReporterPseudonymID).
		Str("content_type", *report.ContentType).
		Str("report_reason", *report.ReportReason).
		Msg("Creating report")

	// Set default values if not provided
	if report.CreatedAt == nil {
		now := sql.Null[time.Time]{}
		now.Scan(time.Now())
		report.CreatedAt = &now
	}

	if report.Status == nil {
		status := sql.Null[string]{}
		status.Scan("pending")
		report.Status = &status
	}

	// Use the generated Reports table helper
	createdReport, err := models.Reports.Insert(report).One(ctx, dao.db)
	if err != nil {
		return nil, fmt.Errorf("failed to create report: %w", err)
	}

	return createdReport, nil
}

// GetReportByID retrieves a report by ID
func (dao *ReportDAO) GetReportByID(ctx context.Context, reportID int64) (*models.Report, error) {
	// Use the generated FindReport function
	report, err := models.FindReport(ctx, dao.db, reportID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get report by ID: %w", err)
	}

	return report, nil
}

// GetReports retrieves reports with filtering and pagination
func (dao *ReportDAO) GetReports(ctx context.Context, status string, page, limit int) ([]*models.Report, error) {
	offset := (page - 1) * limit

	// Build query with optional status filter
	var reports []*models.Report
	var err error

	if status != "" {
		reports, err = models.Reports.Query(
			models.SelectWhere.Reports.Status.EQ(status),
			sm.Limit(limit),
			sm.Offset(offset),
		).All(ctx, dao.db)
	} else {
		reports, err = models.Reports.Query(
			sm.Limit(limit),
			sm.Offset(offset),
		).All(ctx, dao.db)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get reports: %w", err)
	}

	return reports, nil
}

// CountReports counts reports with optional status filter
func (dao *ReportDAO) CountReports(ctx context.Context, status string) (int64, error) {
	// Build query with optional status filter
	var count int64
	var err error

	if status != "" {
		count, err = models.Reports.Query(
			models.SelectWhere.Reports.Status.EQ(status),
		).Count(ctx, dao.db)
	} else {
		count, err = models.Reports.Query().Count(ctx, dao.db)
	}

	if err != nil {
		return 0, fmt.Errorf("failed to count reports: %w", err)
	}

	return count, nil
}

// UpdateReport updates a report
func (dao *ReportDAO) UpdateReport(ctx context.Context, reportID int64, updates *models.ReportSetter) error {
	// First get the report
	report, err := dao.GetReportByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("failed to get report for update: %w", err)
	}
	if report == nil {
		return fmt.Errorf("report not found")
	}

	// Use the generated Update method
	err = report.Update(ctx, dao.db, updates)
	if err != nil {
		return fmt.Errorf("failed to update report: %w", err)
	}

	return nil
}

// ResolveReport resolves a report with resolution details
func (dao *ReportDAO) ResolveReport(ctx context.Context, reportID int64, resolverUserID int64, resolverPseudonymID string, resolutionNotes string, action string) error {
	// Set resolution details
	now := sql.Null[time.Time]{}
	now.Scan(time.Now())

	status := sql.Null[string]{}
	// Set status based on action
	if action == "dismiss" {
		status.Scan("dismissed")
	} else {
		status.Scan("resolved")
	}

	resolvedByUserID := sql.Null[int64]{}
	resolvedByUserID.Scan(resolverUserID)

	resolvedByPseudonymID := sql.Null[string]{}
	resolvedByPseudonymID.Scan(resolverPseudonymID)

	resolutionNotesNull := sql.Null[string]{}
	resolutionNotesNull.Scan(resolutionNotes)

	updates := &models.ReportSetter{
		Status:                &status,
		ResolvedByUserID:      &resolvedByUserID,
		ResolvedByPseudonymID: &resolvedByPseudonymID,
		ResolutionNotes:       &resolutionNotesNull,
		ResolvedAt:            &now,
	}

	return dao.UpdateReport(ctx, reportID, updates)
}

// GetPendingReportsCount returns the count of pending reports for a subforum
func (dao *ReportDAO) GetPendingReportsCount(ctx context.Context, subforumPath string) (int, error) {
	log.Debug().
		Str("subforum_path", subforumPath).
		Msg("Getting pending reports count")

	// Parse subforum path to get community type and name
	parts := strings.Split(subforumPath, "/")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid subforum path format: %s", subforumPath)
	}
	communityType := parts[0]
	subforumName := parts[1]

	// First get the subforum ID
	subforum, err := models.Subforums.Query(
		models.SelectWhere.Subforums.CommunityType.EQ(communityType),
		models.SelectWhere.Subforums.Name.EQ(subforumName),
	).One(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get subforum: %w", err)
	}

	// Get post IDs in the subforum
	posts, err := models.Posts.Query(
		models.SelectWhere.Posts.SubforumID.EQ(subforum.SubforumID),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get posts in subforum: %w", err)
	}

	if len(posts) == 0 {
		return 0, nil
	}

	// Extract post IDs
	postIDs := make([]int64, len(posts))
	for i, post := range posts {
		postIDs[i] = post.PostID
	}

	// Get comment IDs in the subforum
	comments, err := models.Comments.Query(
		models.SelectWhere.Comments.PostID.In(postIDs...),
	).All(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to get comments in subforum: %w", err)
	}

	// Extract comment IDs
	commentIDs := make([]int64, len(comments))
	for i, comment := range comments {
		commentIDs[i] = comment.CommentID
	}

	// Count pending reports for posts in the subforum
	postReportsCount, err := models.Reports.Query(
		models.SelectWhere.Reports.Status.EQ("pending"),
		models.SelectWhere.Reports.ContentType.EQ("post"),
		models.SelectWhere.Reports.ContentID.In(postIDs...),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count post reports: %w", err)
	}

	// Count pending reports for comments in the subforum
	commentReportsCount, err := models.Reports.Query(
		models.SelectWhere.Reports.Status.EQ("pending"),
		models.SelectWhere.Reports.ContentType.EQ("comment"),
		models.SelectWhere.Reports.ContentID.In(commentIDs...),
	).Count(ctx, dao.db)
	if err != nil {
		return 0, fmt.Errorf("failed to count comment reports: %w", err)
	}

	return int(postReportsCount + commentReportsCount), nil
}

// UpdateReportWithForwarding updates a report with forwarding information
func (dao *ReportDAO) UpdateReportWithForwarding(ctx context.Context, reportID int64, forwardingNotes string, forwardedByUserID int64) error {
	log.Debug().
		Int64("report_id", reportID).
		Str("forwarding_notes", forwardingNotes).
		Int64("forwarded_by_user_id", forwardedByUserID).
		Msg("Updating report with forwarding information")

	// Create update setter with forwarding fields
	updateSetter := &models.ReportSetter{
		ForwardedToPlatform: &sql.Null[bool]{V: true, Valid: true},
		ForwardingNotes:     &sql.Null[string]{V: forwardingNotes, Valid: true},
		ForwardedByUserID:   &sql.Null[int64]{V: forwardedByUserID, Valid: true},
		ForwardedAt:         &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	// Use the existing UpdateReport method
	return dao.UpdateReport(ctx, reportID, updateSetter)
}

// CreateRuleViolationReport creates a report for a rule violation
func (dao *ReportDAO) CreateRuleViolationReport(ctx context.Context, reporterPseudonymID, contentType, ruleCode, ruleType string, contentID *int64, reportedPseudonymID, reportDetails string) (*models.Report, error) {
	log.Debug().
		Str("reporter_pseudonymID", reporterPseudonymID).
		Str("content_type", contentType).
		Str("rule_code", ruleCode).
		Str("rule_type", ruleType).
		Msg("Creating rule violation report")

	// Create report setter with rule violation information
	reportSetter := &models.ReportSetter{
		ReporterPseudonymID: &reporterPseudonymID,
		ContentType:         &contentType,
		ReportReason:        &ruleCode, // Use rule code as reason
		Status:              &sql.Null[string]{V: "pending", Valid: true},
		RuleCode:            &sql.Null[string]{V: ruleCode, Valid: true},
		RuleType:            &sql.Null[string]{V: ruleType, Valid: true},
		CreatedAt:           &sql.Null[time.Time]{V: time.Now(), Valid: true},
	}

	// Set content ID if provided
	if contentID != nil {
		reportSetter.ContentID = &sql.Null[int64]{V: *contentID, Valid: true}
	}

	// Set reported pseudonym ID if provided
	if reportedPseudonymID != "" {
		reportSetter.ReportedPseudonymID = &sql.Null[string]{V: reportedPseudonymID, Valid: true}
	}

	// Set report details if provided
	if reportDetails != "" {
		reportSetter.ReportDetails = &sql.Null[string]{V: reportDetails, Valid: true}
	}

	// Use the existing CreateReport method
	return dao.CreateReport(ctx, reportSetter)
}
