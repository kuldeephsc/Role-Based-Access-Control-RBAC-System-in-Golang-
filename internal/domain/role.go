package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID
	Name        string
	Description string
	CreatedAt   time.Time
}

type RoleRepository interface {
	Create(ctx context.Context, r *Role) error
	GetByID(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	List(ctx context.Context) ([]*Role, error)
	Delete(ctx context.Context, id uuid.UUID) error
	AttachPermission(ctx context.Context, roleID, permissionID uuid.UUID) error
	DetachPermission(ctx context.Context, roleID, permissionID uuid.UUID) error
	GetPermissions(ctx context.Context, roleID uuid.UUID) ([]*Permission, error)
	// GetUserIDsForRole exists purely for cache invalidation: when a
	// permission is attached to or detached from a role, every user
	// holding that role needs their user:{id}:permissions cache entry
	// cleared, not just the role itself.
	GetUserIDsForRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error)
}
