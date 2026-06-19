package domain

import "context"

// OutboxRepository is written to inside the same transaction as the
// business change it's recording (via TxRunner.RunInTx), and read by the
// separate relay process that publishes to RabbitMQ. The Postgres
// implementation deliberately does NOT implement Read/MarkSent through
// dbFromContext -- the relay runs outside any business transaction, by
// design, since it polls continuously rather than reacting to one write.
type OutboxRepository interface {
	Create(ctx context.Context, eventType string, payload interface{}) error
}
