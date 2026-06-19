package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Permission struct {
	ID          uuid.UUID
	Name        string
	Resource    string
	Action      string
	Description string
	CreatedAt   time.Time
}

type PermissionRepository interface {
	Create(ctx context.Context, p *Permission) error
	GetByID(ctx context.Context, id uuid.UUID) (*Permission, error)
	GetByName(ctx context.Context, name string) (*Permission, error)
	List(ctx context.Context) ([]*Permission, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
