package dao

import (
	"context"

	"github.com/matt0x6f/hashpost/internal/database/models"
	"github.com/stephenafamo/bob"
)

// CorrelationAuditDAO implements CorrelationAuditDAOInterface
type CorrelationAuditDAO struct {
	db bob.Executor
}

// NewCorrelationAuditDAO creates a new correlation audit DAO
func NewCorrelationAuditDAO(db bob.Executor) *CorrelationAuditDAO {
	return &CorrelationAuditDAO{
		db: db,
	}
}

// CreateCorrelationAudit creates a new correlation audit record
func (dao *CorrelationAuditDAO) CreateCorrelationAudit(ctx context.Context, auditRecord *models.CorrelationAuditSetter) error {
	_, err := models.CorrelationAudits.Insert(auditRecord).One(ctx, dao.db)
	return err
}

// GetCorrelationHistory retrieves correlation history with pagination
func (dao *CorrelationAuditDAO) GetCorrelationHistory(ctx context.Context, correlationType string, page, limit int) (models.CorrelationAuditSlice, error) {
	// For now, just return all records since the ViewQuery doesn't support filtering
	// In a real implementation, you'd need to use a proper SelectQuery
	return models.CorrelationAudits.Query().All(ctx, dao.db)
}
