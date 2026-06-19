package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rbac-platform/internal/domain"
)

type PermissionRepository struct {
	db *gorm.DB
}

func NewPermissionRepository(db *gorm.DB) *PermissionRepository {
	return &PermissionRepository{db: db}
}

var _ domain.PermissionRepository = (*PermissionRepository)(nil)

func (r *PermissionRepository) Create(ctx context.Context, p *domain.Permission) error {
	m := &PermissionModel{Name: p.Name, Resource: p.Resource, Action: p.Action, Description: p.Description}
	if err := dbFromContext(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	p.ID, p.CreatedAt = m.ID, m.CreatedAt
	return nil
}

func (r *PermissionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Permission, error) {
	var m PermissionModel
	if err := dbFromContext(ctx, r.db).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return fromPermissionModel(&m), nil
}

func (r *PermissionRepository) GetByName(ctx context.Context, name string) (*domain.Permission, error) {
	var m PermissionModel
	if err := dbFromContext(ctx, r.db).First(&m, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return fromPermissionModel(&m), nil
}

func (r *PermissionRepository) List(ctx context.Context) ([]*domain.Permission, error) {
	var models []PermissionModel
	if err := dbFromContext(ctx, r.db).Order("name").Find(&models).Error; err != nil {
		return nil, err
	}
	perms := make([]*domain.Permission, len(models))
	for i := range models {
		perms[i] = fromPermissionModel(&models[i])
	}
	return perms, nil
}

func (r *PermissionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return dbFromContext(ctx, r.db).Delete(&PermissionModel{}, "id = ?", id).Error
}

func fromPermissionModel(m *PermissionModel) *domain.Permission {
	return &domain.Permission{
		ID: m.ID, Name: m.Name, Resource: m.Resource, Action: m.Action,
		Description: m.Description, CreatedAt: m.CreatedAt,
	}
}
