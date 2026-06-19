package analytics

import (
	"context"
	"encoding/json"
	"log/slog"
)

// Consumer handles events on analytics.queue (login.success, login.failure,
// user.created). Increments counters via structured logging; a real
// production version would write to a time-series store or data warehouse.
type Consumer struct {
	logger *slog.Logger
}

func NewConsumer(logger *slog.Logger) *Consumer {
	return &Consumer{logger: logger}
}

func (c *Consumer) Handle(ctx context.Context, eventType string, body []byte) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}
	c.logger.Info("analytics event", "event_type", eventType, "payload", payload)
	return nil
}
