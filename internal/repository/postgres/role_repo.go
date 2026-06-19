package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rbac-platform/internal/domain"
)

type RoleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) *RoleRepository {
	return &RoleRepository{db: db}
}

var _ domain.RoleRepository = (*RoleRepository)(nil)

func (r *RoleRepository) Create(ctx context.Context, role *domain.Role) error {
	m := &RoleModel{Name: role.Name, Description: role.Description}
	if err := dbFromContext(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	role.ID, role.CreatedAt = m.ID, m.CreatedAt
	return nil
}

func (r *RoleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	var m RoleModel
	if err := dbFromContext(ctx, r.db).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return fromRoleModel(&m), nil
}

func (r *RoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	var m RoleModel
	if err := dbFromContext(ctx, r.db).First(&m, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return fromRoleModel(&m), nil
}

func (r *RoleRepository) List(ctx context.Context) ([]*domain.Role, error) {
	var models []RoleModel
	if err := dbFromContext(ctx, r.db).Order("name").Find(&models).Error; err != nil {
		return nil, err
	}
	roles := make([]*domain.Role, len(models))
	for i := range models {
		roles[i] = fromRoleModel(&models[i])
	}
	return roles, nil
}

func (r *RoleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return dbFromContext(ctx, r.db).Delete(&RoleModel{}, "id = ?", id).Error
}

func (r *RoleRepository) AttachPermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	rp := RolePermissionModel{RoleID: roleID, PermissionID: permissionID}
	return dbFromContext(ctx, r.db).Clauses(onConflictDoNothing()).Create(&rp).Error
}

func (r *RoleRepository) DetachPermission(ctx context.Context, roleID, permissionID uuid.UUID) error {
	return dbFromContext(ctx, r.db).
		Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&RolePermissionModel{}).Error
}

func (r *RoleRepository) GetPermissions(ctx context.Context, roleID uuid.UUID) ([]*domain.Permission, error) {
	var models []PermissionModel
	err := dbFromContext(ctx, r.db).
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Where("role_permissions.role_id = ?", roleID).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	perms := make([]*domain.Permission, len(models))
	for i := range models {
		perms[i] = fromPermissionModel(&models[i])
	}
	return perms, nil
}

func (r *RoleRepository) GetUserIDsForRole(ctx context.Context, roleID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := dbFromContext(ctx, r.db).Model(&UserRoleModel{}).
		Where("role_id = ?", roleID).
		Pluck("user_id", &ids).Error
	return ids, err
}

func fromRoleModel(m *RoleModel) *domain.Role {
	return &domain.Role{ID: m.ID, Name: m.Name, Description: m.Description, CreatedAt: m.CreatedAt}
}
