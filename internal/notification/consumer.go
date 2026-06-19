package notification

import (
	"context"
	"encoding/json"
	"log/slog"
)

// Consumer handles events on notification.queue (user.created, role.assigned).
// Phase 1 just logs — Phase 7 integrates actual email or websocket push.
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
	c.logger.Info("notification event", "event_type", eventType, "payload", payload)
	return nil
}
