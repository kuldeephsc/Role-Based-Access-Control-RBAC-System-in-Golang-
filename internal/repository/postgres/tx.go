package postgres

import (
	"context"

	"gorm.io/gorm"

	"rbac-platform/internal/domain"
)

type txKeyType struct{}

var txKey = txKeyType{}

type TxRunner struct {
	db *gorm.DB
}

func NewTxRunner(db *gorm.DB) *TxRunner {
	return &TxRunner{db: db}
}

var _ domain.TxRunner = (*TxRunner)(nil)

func (t *TxRunner) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txKey, tx))
	})
}

// dbFromContext returns the active transaction if called from inside
// RunInTx, otherwise falls back to the base connection. Every repository
// method in this package goes through this instead of touching r.db
// directly, so any sequence of repository calls made inside one RunInTx
// block automatically shares a single transaction.
func dbFromContext(ctx context.Context, base *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey).(*gorm.DB); ok {
		return tx
	}
	return base.WithContext(ctx)
}
