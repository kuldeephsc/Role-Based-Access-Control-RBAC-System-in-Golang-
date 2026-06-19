package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RefreshToken rows are the actual source of truth for "active sessions" --
// there is deliberately no separate Redis session store (see the
// architecture spec, decision #3). A session is just a non-revoked,
// non-expired row here.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

type RefreshTokenRepository interface {
	Create(ctx context.Context, t *RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error)
}
