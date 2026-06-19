package postgres

import (
	"context"
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"rbac-platform/internal/domain"
)

type OutboxModel struct {
	ID         uint64 `gorm:"primaryKey"`
	EventType  string `gorm:"column:event_type;not null"`
	Payload    []byte `gorm:"column:payload;type:jsonb;not null"`
	Status     string `gorm:"column:status;not null;default:pending"`
	RetryCount int    `gorm:"column:retry_count;not null;default:0"`
	CreatedAt  time.Time
	SentAt     *time.Time
}

func (OutboxModel) TableName() string { return "outbox_events" }

type OutboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) *OutboxRepository {
	return &OutboxRepository{db: db}
}

var _ domain.OutboxRepository = (*OutboxRepository)(nil)

// Create is always called from inside a TxRunner.RunInTx block alongside
// the business write it's recording -- see rbac.Service.AssignRoleToUser
// for the canonical example. dbFromContext picks up that transaction
// automatically.
func (r *OutboxRepository) Create(ctx context.Context, eventType string, payload interface{}) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	m := &OutboxModel{EventType: eventType, Payload: b, Status: "pending"}
	return dbFromContext(ctx, r.db).Create(m).Error
}

// The methods below are used only by the outbox relay (internal/outbox),
// which polls continuously rather than running inside any business
// transaction -- they intentionally use r.db directly, not dbFromContext.

func (r *OutboxRepository) FetchPending(ctx context.Context, limit int) ([]OutboxModel, error) {
	var rows []OutboxModel
	err := r.db.WithContext(ctx).
		Where("status = ?", "pending").
		Order("created_at").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *OutboxRepository) MarkSent(ctx context.Context, id uint64) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&OutboxModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"status": "sent", "sent_at": now}).Error
}

func (r *OutboxRepository) MarkFailed(ctx context.Context, id uint64) error {
	return r.db.WithContext(ctx).Model(&OutboxModel{}).
		Where("id = ?", id).
		Update("retry_count", gorm.Expr("retry_count + 1")).Error
}

// OldestPendingAge backs the outbox_relay_lag_seconds gauge.
func (r *OutboxRepository) OldestPendingAge(ctx context.Context) (time.Duration, error) {
	var m OutboxModel
	err := r.db.WithContext(ctx).Where("status = ?", "pending").Order("created_at").First(&m).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil
		}
		return 0, err
	}
	return time.Since(m.CreatedAt), nil
}
