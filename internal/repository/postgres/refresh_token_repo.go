package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rbac-platform/internal/domain"
)

type RefreshTokenRepository struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

var _ domain.RefreshTokenRepository = (*RefreshTokenRepository)(nil)

func (r *RefreshTokenRepository) Create(ctx context.Context, t *domain.RefreshToken) error {
	m := &RefreshTokenModel{
		UserID:    t.UserID,
		TokenHash: t.TokenHash,
		ExpiresAt: t.ExpiresAt,
	}
	if err := dbFromContext(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	t.ID, t.CreatedAt = m.ID, m.CreatedAt
	return nil
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	var m RefreshTokenModel
	if err := dbFromContext(ctx, r.db).First(&m, "token_hash = ?", hash).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &domain.RefreshToken{
		ID: m.ID, UserID: m.UserID, TokenHash: m.TokenHash,
		ExpiresAt: m.ExpiresAt, RevokedAt: m.RevokedAt, CreatedAt: m.CreatedAt,
	}, nil
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	return dbFromContext(ctx, r.db).Model(&RefreshTokenModel{}).
		Where("id = ?", id).
		Update("revoked_at", now).Error
}

func (r *RefreshTokenRepository) CountActiveByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	var count int64
	err := dbFromContext(ctx, r.db).Model(&RefreshTokenModel{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, time.Now()).
		Count(&count).Error
	return count, err
}
