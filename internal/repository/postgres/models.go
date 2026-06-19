package postgres

import (
	"time"

	"github.com/google/uuid"
)

// These are the GORM-tagged persistence models. They mirror but are kept
// separate from the domain entities in internal/domain -- the domain layer
// stays free of `gorm` struct tags, and each repository file below maps
// between the two. This is the seam that lets Postgres be swapped out
// without touching any service or handler code.

type UserModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	FullName     string
	IsActive     bool `gorm:"default:true"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (UserModel) TableName() string { return "users" }

type RoleModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"uniqueIndex;not null"`
	Description string
	CreatedAt   time.Time
}

func (RoleModel) TableName() string { return "roles" }

type PermissionModel struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string    `gorm:"uniqueIndex;not null"`
	Resource    string    `gorm:"not null"`
	Action      string    `gorm:"not null"`
	Description string
	CreatedAt   time.Time
}

func (PermissionModel) TableName() string { return "permissions" }

type UserRoleModel struct {
	UserID     uuid.UUID  `gorm:"type:uuid;primaryKey"`
	RoleID     uuid.UUID  `gorm:"type:uuid;primaryKey"`
	AssignedBy *uuid.UUID `gorm:"type:uuid"`
	AssignedAt time.Time
}

func (UserRoleModel) TableName() string { return "user_roles" }

type RolePermissionModel struct {
	RoleID       uuid.UUID `gorm:"type:uuid;primaryKey"`
	PermissionID uuid.UUID `gorm:"type:uuid;primaryKey"`
	GrantedAt    time.Time
}

func (RolePermissionModel) TableName() string { return "role_permissions" }

type RefreshTokenModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	TokenHash string    `gorm:"not null"`
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (RefreshTokenModel) TableName() string { return "refresh_tokens" }
