package dao

import (
	"context"
	"database/sql"

	"github.com/stephenafamo/bob"
)

// BaseDAO provides common database operations
type BaseDAO struct {
	db bob.DB
}

// NewBaseDAO creates a new base DAO
func NewBaseDAO(db bob.DB) *BaseDAO {
	return &BaseDAO{
		db: db,
	}
}

// GetDB returns the database connection
func (dao *BaseDAO) GetDB() bob.DB {
	return dao.db
}

// BeginTx starts a new transaction
func (dao *BaseDAO) BeginTx(ctx context.Context, opts *sql.TxOptions) (bob.Transaction, error) {
	return dao.db.BeginTx(ctx, opts)
}

// HealthCheck performs a health check on the database
func (dao *BaseDAO) HealthCheck(ctx context.Context) error {
	return dao.db.PingContext(ctx)
}
