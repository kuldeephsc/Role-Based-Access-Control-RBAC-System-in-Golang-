package outbox

import (
	"context"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"rbac-platform/internal/platform/metrics"
	pgrepo "rbac-platform/internal/repository/postgres"
)

const pollInterval = 500 * time.Millisecond

// Relay is the piece that makes the transactional outbox pattern real.
// The business write and its outbox_events row already committed
// atomically in Postgres (see rbac.Service) before this ever runs --
// RabbitMQ being temporarily unreachable just means rows sit pending a
// little longer, never that an event is silently lost.
type Relay struct {
	outbox *pgrepo.OutboxRepository
	ch     *amqp.Channel
	logger *slog.Logger
}

func NewRelay(outbox *pgrepo.OutboxRepository, ch *amqp.Channel, logger *slog.Logger) *Relay {
	return &Relay{outbox: outbox, ch: ch, logger: logger}
}

func (r *Relay) Run(ctx context.Context) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.relayBatch(ctx)
			r.recordLag(ctx)
		}
	}
}

func (r *Relay) relayBatch(ctx context.Context) {
	rows, err := r.outbox.FetchPending(ctx, 50)
	if err != nil {
		r.logger.Error("outbox fetch failed", "error", err)
		return
	}
	for _, row := range rows {
		err := r.ch.PublishWithContext(ctx, "rbac.events", row.EventType, false, false, amqp.Publishing{
			ContentType: "application/json",
			Body:        row.Payload,
		})
		if err != nil {
			r.logger.Error("outbox publish failed", "event_id", row.ID, "event_type", row.EventType, "error", err)
			_ = r.outbox.MarkFailed(ctx, row.ID)
			continue
		}
		if err := r.outbox.MarkSent(ctx, row.ID); err != nil {
			r.logger.Error("outbox mark-sent failed", "event_id", row.ID, "error", err)
			continue
		}
		metrics.RabbitMQPublishedTotal.WithLabelValues(row.EventType).Inc()
	}
}

func (r *Relay) recordLag(ctx context.Context) {
	age, err := r.outbox.OldestPendingAge(ctx)
	if err != nil {
		return
	}
	metrics.OutboxRelayLagSeconds.Set(age.Seconds())
}
