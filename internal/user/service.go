package user

import (
	"context"

	"github.com/google/uuid"

	"rbac-platform/internal/domain"
)

type Service struct {
	users domain.UserRepository
}

func NewService(users domain.UserRepository) *Service {
	return &Service{users: users}
}

func (s *Service) Get(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return s.users.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, limit, offset int) ([]*domain.User, int64, error) {
	return s.users.List(ctx, limit, offset)
}

func (s *Service) UpdateProfile(ctx context.Context, id uuid.UUID, fullName string) (*domain.User, error) {
	u, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	u.FullName = fullName
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}
