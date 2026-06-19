package rabbitmq

import (
	"context"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"

	"rbac-platform/internal/platform/metrics"
)

const maxRetries = 3

// HandlerFunc processes one event. routingKey is the event_type (e.g.
// "role.assigned"); body is the JSON payload written to outbox_events.
type HandlerFunc func(ctx context.Context, eventType string, body []byte) error

// Consume runs handler for every message on queueName until ctx is
// cancelled. On handler error, the message is republished to the events
// exchange with an incremented x-retry-count header, up to maxRetries;
// past that it's nacked without requeue, and the queue's
// x-dead-letter-exchange (declared in setupTopology) routes it to that
// queue's DLQ instead of losing it.
func Consume(ctx context.Context, ch *amqp.Channel, queueName, consumerName string, logger *slog.Logger, handler HandlerFunc) error {
	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-msgs:
				if !ok {
					return
				}
				retryCount := headerRetryCount(d.Headers)

				err := handler(ctx, d.RoutingKey, d.Body)
				if err == nil {
					_ = d.Ack(false)
					metrics.EventsConsumedTotal.WithLabelValues(d.RoutingKey, consumerName).Inc()
					continue
				}

				logger.Error("consumer handler failed",
					"queue", queueName, "routing_key", d.RoutingKey, "retry", retryCount, "error", err)

				if retryCount >= maxRetries {
					_ = d.Nack(false, false) // dead-lettered to this queue's DLQ
					continue
				}
				_ = ch.PublishWithContext(ctx, EventsExchange, d.RoutingKey, false, false, amqp.Publishing{
					ContentType: "application/json",
					Body:        d.Body,
					Headers:     amqp.Table{"x-retry-count": int32(retryCount + 1)},
				})
				_ = d.Ack(false)
			}
		}
	}()
	return nil
}

func headerRetryCount(headers amqp.Table) int {
	if headers == nil {
		return 0
	}
	if v, ok := headers["x-retry-count"]; ok {
		if n, ok := v.(int32); ok {
			return int(n)
		}
	}
	return 0
}
