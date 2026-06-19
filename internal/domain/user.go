package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User is the core identity entity. It has no persistence-layer
// annotations on purpose -- GORM tags live on the model structs in
// internal/repository/postgres, not here. Keeping this struct free of
// infrastructure concerns is what lets the service layer be tested and
// reasoned about without a database.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// UserRepository is implemented by internal/repository/postgres.UserRepository.
// Services depend on this interface, never on the concrete GORM type --
// that's what makes swapping the storage engine later a one-package change.
type UserRepository interface {
	Create(ctx context.Context, u *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	List(ctx context.Context, limit, offset int) ([]*User, int64, error)
	Update(ctx context.Context, u *User) error
	AssignRole(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error
	RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error
	GetRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error)
	GetPermissions(ctx context.Context, userID uuid.UUID) ([]*Permission, error)
}
