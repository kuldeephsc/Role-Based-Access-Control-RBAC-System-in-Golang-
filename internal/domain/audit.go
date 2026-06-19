package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID           int64
	ActorID      *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   string
	Metadata     map[string]interface{}
	CreatedAt    time.Time
}

type AuditRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, limit, offset int) ([]*AuditLog, int64, error)
}
