package domain

import "context"

// TxRunner executes fn inside a single database transaction. Repository
// methods called with the context fn receives automatically participate
// in that same transaction (see internal/repository/postgres/tx.go) --
// this is what lets a business write and its outbox_events row commit or
// roll back together, fixing the dual-write problem described in the
// architecture spec (decision #6).
type TxRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
