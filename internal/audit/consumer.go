package audit

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	"rbac-platform/internal/domain"
)

type Consumer struct {
	repo domain.AuditRepository
}

func NewConsumer(repo domain.AuditRepository) *Consumer {
	return &Consumer{repo: repo}
}

// Handle is the rabbitmq.HandlerFunc that processes every event routed to
// audit.queue (bound with "#", so it sees everything). It decodes the
// outbox payload into a generic map, extracts the actor if present, and
// writes one audit_logs row per event. Failures return an error, which
// the generic consumer wrapper (internal/platform/rabbitmq/consumer.go)
// translates into a retry-then-DLQ flow.
func (c *Consumer) Handle(ctx context.Context, eventType string, body []byte) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	var actorID *uuid.UUID
	if v, ok := payload["assigned_by"]; ok {
		if s, ok := v.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				actorID = &id
			}
		}
	} else if v, ok := payload["user_id"]; ok {
		if s, ok := v.(string); ok {
			if id, err := uuid.Parse(s); err == nil {
				actorID = &id
			}
		}
	}

	log := &domain.AuditLog{
		ActorID:  actorID,
		Action:   eventType,
		Metadata: payload,
	}

	if rt, ok := payload["resource_type"]; ok {
		log.ResourceType, _ = rt.(string)
	}
	if ri, ok := payload["resource_id"]; ok {
		log.ResourceID, _ = ri.(string)
	}

	return c.repo.Create(ctx, log)
}
