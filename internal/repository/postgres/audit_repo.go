package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rbac-platform/internal/domain"
)

type AuditLogModel struct {
	ID           int64      `gorm:"primaryKey"`
	ActorID      *uuid.UUID `gorm:"type:uuid"`
	Action       string     `gorm:"not null"`
	ResourceType string
	ResourceID   string
	Metadata     []byte `gorm:"type:jsonb"`
	CreatedAt    time.Time
}

func (AuditLogModel) TableName() string { return "audit_logs" }

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

var _ domain.AuditRepository = (*AuditRepository)(nil)

func (r *AuditRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	meta, err := json.Marshal(log.Metadata)
	if err != nil {
		return err
	}
	m := &AuditLogModel{
		ActorID:      log.ActorID,
		Action:       log.Action,
		ResourceType: log.ResourceType,
		ResourceID:   log.ResourceID,
		Metadata:     meta,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	log.ID, log.CreatedAt = m.ID, m.CreatedAt
	return nil
}

func (r *AuditRepository) List(ctx context.Context, limit, offset int) ([]*domain.AuditLog, int64, error) {
	var rows []AuditLogModel
	var total int64
	if err := r.db.WithContext(ctx).Model(&AuditLogModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	logs := make([]*domain.AuditLog, len(rows))
	for i, m := range rows {
		var meta map[string]interface{}
		_ = json.Unmarshal(m.Metadata, &meta)
		logs[i] = &domain.AuditLog{
			ID: m.ID, ActorID: m.ActorID, Action: m.Action,
			ResourceType: m.ResourceType, ResourceID: m.ResourceID,
			Metadata: meta, CreatedAt: m.CreatedAt,
		}
	}
	return logs, total, nil
}
