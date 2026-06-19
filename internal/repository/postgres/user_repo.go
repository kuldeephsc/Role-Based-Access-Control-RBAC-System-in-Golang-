package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"rbac-platform/internal/domain"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

var _ domain.UserRepository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	m := toUserModel(u)
	if err := dbFromContext(ctx, r.db).Create(m).Error; err != nil {
		return err
	}
	u.ID, u.CreatedAt, u.UpdatedAt = m.ID, m.CreatedAt, m.UpdatedAt
	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var m UserModel
	if err := dbFromContext(ctx, r.db).First(&m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return fromUserModel(&m), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var m UserModel
	if err := dbFromContext(ctx, r.db).First(&m, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return fromUserModel(&m), nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*domain.User, int64, error) {
	var models []UserModel
	var total int64
	if err := dbFromContext(ctx, r.db).Model(&UserModel{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := dbFromContext(ctx, r.db).Limit(limit).Offset(offset).Order("created_at desc").Find(&models).Error; err != nil {
		return nil, 0, err
	}
	users := make([]*domain.User, len(models))
	for i := range models {
		users[i] = fromUserModel(&models[i])
	}
	return users, total, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	return dbFromContext(ctx, r.db).Model(&UserModel{}).Where("id = ?", u.ID).
		Updates(map[string]interface{}{
			"full_name": u.FullName,
			"is_active": u.IsActive,
		}).Error
}

func (r *UserRepository) AssignRole(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	ur := UserRoleModel{UserID: userID, RoleID: roleID, AssignedBy: &assignedBy}
	return dbFromContext(ctx, r.db).Clauses(onConflictDoNothing()).Create(&ur).Error
}

func (r *UserRepository) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	return dbFromContext(ctx, r.db).
		Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&UserRoleModel{}).Error
}

func (r *UserRepository) GetRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	var models []RoleModel
	err := dbFromContext(ctx, r.db).
		Joins("JOIN user_roles ON user_roles.role_id = roles.id").
		Where("user_roles.user_id = ?", userID).
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	roles := make([]*domain.Role, len(models))
	for i := range models {
		roles[i] = fromRoleModel(&models[i])
	}
	return roles, nil
}

func (r *UserRepository) GetPermissions(ctx context.Context, userID uuid.UUID) ([]*domain.Permission, error) {
	var models []PermissionModel
	err := dbFromContext(ctx, r.db).
		Distinct().
		Joins("JOIN role_permissions ON role_permissions.permission_id = permissions.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_permissions.role_id").
		Where("user_roles.user_id = ?", userID).
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

func toUserModel(u *domain.User) *UserModel {
	return &UserModel{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		FullName:     u.FullName,
		IsActive:     u.IsActive,
	}
}

func fromUserModel(m *UserModel) *domain.User {
	return &domain.User{
		ID:           m.ID,
		Email:        m.Email,
		PasswordHash: m.PasswordHash,
		FullName:     m.FullName,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}
