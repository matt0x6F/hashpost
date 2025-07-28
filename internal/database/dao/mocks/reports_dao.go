package mocks

import (
	"context"
	"database/sql"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stretchr/testify/mock"
)

// MockReportDAO is a mock implementation of ReportDAOInterface with data injection support
type MockReportDAO struct {
	mock.Mock
	// Custom data that can be injected for testing
	reports         map[int64]*models.Report
	reportsByStatus map[string][]*models.Report
	counts          map[string]int64 // key: "status"
}

// NewMockReportDAO creates a new mock ReportDAO with optional initial data
func NewMockReportDAO() *MockReportDAO {
	return &MockReportDAO{
		reports:         make(map[int64]*models.Report),
		reportsByStatus: make(map[string][]*models.Report),
		counts:          make(map[string]int64),
	}
}

// InjectReport injects a report into the mock for testing
func (m *MockReportDAO) InjectReport(report *models.Report) {
	m.reports[report.ReportID] = report
}

// InjectReportsByStatus injects reports that should be returned when querying by status
func (m *MockReportDAO) InjectReportsByStatus(status string, reports []*models.Report) {
	m.reportsByStatus[status] = reports
}

// InjectCount injects a count that should be returned for count operations
func (m *MockReportDAO) InjectCount(status string, count int64) {
	m.counts[status] = count
}

// SetDefaultBehavior sets up default mock behavior for common operations
func (m *MockReportDAO) SetDefaultBehavior() {
	// Default behavior for CreateReport
	m.On("CreateReport", mock.Anything, mock.AnythingOfType("*models.ReportSetter")).Return(
		func(ctx context.Context, reportSetter *models.ReportSetter) (*models.Report, error) {
			// Create a mock report with the provided data
			report := &models.Report{
				ReportID:            123, // Mock ID
				ReporterPseudonymID: *reportSetter.ReporterPseudonymID,
				ContentType:         *reportSetter.ContentType,
				ReportReason:        *reportSetter.ReportReason,
				Status:              *reportSetter.Status,
				CreatedAt:           *reportSetter.CreatedAt,
			}

			// Set optional fields if provided
			if reportSetter.ContentID != nil {
				report.ContentID = *reportSetter.ContentID
			}
			if reportSetter.ReportedPseudonymID != nil {
				report.ReportedPseudonymID = *reportSetter.ReportedPseudonymID
			}
			if reportSetter.ReportDetails != nil {
				report.ReportDetails = *reportSetter.ReportDetails
			}

			return report, nil
		},
	)

	// Default behavior for GetReportByID
	m.On("GetReportByID", mock.Anything, mock.AnythingOfType("int64")).Return(
		func(ctx context.Context, reportID int64) (*models.Report, error) {
			if report, exists := m.reports[reportID]; exists {
				return report, nil
			}
			return nil, sql.ErrNoRows
		},
	)

	// Default behavior for GetReports
	m.On("GetReports", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("int"), mock.AnythingOfType("int")).Return(
		func(ctx context.Context, status string, page, limit int) ([]*models.Report, error) {
			if reports, exists := m.reportsByStatus[status]; exists {
				return reports, nil
			}
			return []*models.Report{}, nil
		},
	)

	// Default behavior for CountReports
	m.On("CountReports", mock.Anything, mock.AnythingOfType("string")).Return(
		func(ctx context.Context, status string) (int64, error) {
			count := m.counts[status]
			return count, nil
		},
	)
}

// CreateReport creates a new report
func (m *MockReportDAO) CreateReport(ctx context.Context, report *models.ReportSetter) (*models.Report, error) {
	args := m.Called(ctx, report)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, *models.ReportSetter) (*models.Report, error)); ok {
		return fn(ctx, report)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Report), args.Error(1)
}

// GetReportByID retrieves a report by ID
func (m *MockReportDAO) GetReportByID(ctx context.Context, reportID int64) (*models.Report, error) {
	args := m.Called(ctx, reportID)
	if args.Get(0) == nil {
		return nil, sql.ErrNoRows
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, int64) (*models.Report, error)); ok {
		return fn(ctx, reportID)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Report), args.Error(1)
}

// GetReports retrieves reports with filtering and pagination
func (m *MockReportDAO) GetReports(ctx context.Context, status string, page, limit int) ([]*models.Report, error) {
	args := m.Called(ctx, status, page, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string, int, int) ([]*models.Report, error)); ok {
		return fn(ctx, status, page, limit)
	}

	// Fallback to direct return values
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Report), args.Error(1)
}

// CountReports counts reports with optional status filter
func (m *MockReportDAO) CountReports(ctx context.Context, status string) (int64, error) {
	args := m.Called(ctx, status)

	// Check if the return value is a function
	if fn, ok := args.Get(0).(func(context.Context, string) (int64, error)); ok {
		return fn(ctx, status)
	}

	// Fallback to direct return values
	return args.Get(0).(int64), args.Error(1)
}

// UpdateReport updates a report
func (m *MockReportDAO) UpdateReport(ctx context.Context, reportID int64, updates *models.ReportSetter) error {
	args := m.Called(ctx, reportID, updates)
	return args.Error(0)
}

// ResolveReport resolves a report with resolution details
func (m *MockReportDAO) ResolveReport(ctx context.Context, reportID int64, resolverUserID int64, resolverPseudonymID string, resolutionNotes string) error {
	args := m.Called(ctx, reportID, resolverUserID, resolverPseudonymID, resolutionNotes)
	return args.Error(0)
}

// GetPendingReportsCount returns the count of pending reports for a subforum
func (m *MockReportDAO) GetPendingReportsCount(ctx context.Context, subforumPath string) (int, error) {
	args := m.Called(ctx, subforumPath)
	return args.Get(0).(int), args.Error(1)
}
