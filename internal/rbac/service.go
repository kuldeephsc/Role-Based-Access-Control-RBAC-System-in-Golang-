package rbac

import (
	"context"

	"github.com/google/uuid"

	"rbac-platform/internal/domain"
	"rbac-platform/internal/platform/cache"
	"rbac-platform/internal/platform/metrics"
)

type Service struct {
	users   domain.UserRepository
	roles   domain.RoleRepository
	perms   domain.PermissionRepository
	outbox  domain.OutboxRepository
	txRunner domain.TxRunner
	cache   *cache.PermissionCache
}

func NewService(
	users domain.UserRepository,
	roles domain.RoleRepository,
	perms domain.PermissionRepository,
	outbox domain.OutboxRepository,
	txRunner domain.TxRunner,
	permCache *cache.PermissionCache,
) *Service {
	return &Service{users: users, roles: roles, perms: perms, outbox: outbox, txRunner: txRunner, cache: permCache}
}

// --- Roles ---

func (s *Service) CreateRole(ctx context.Context, name, description string) (*domain.Role, error) {
	r := &domain.Role{Name: name, Description: description}
	if err := s.roles.Create(ctx, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (s *Service) ListRoles(ctx context.Context) ([]*domain.Role, error) {
	return s.roles.List(ctx)
}

func (s *Service) DeleteRole(ctx context.Context, id uuid.UUID) error {
	return s.roles.Delete(ctx, id)
}

// --- Permissions ---

func (s *Service) CreatePermission(ctx context.Context, name, resource, action, description string) (*domain.Permission, error) {
	p := &domain.Permission{Name: name, Resource: resource, Action: action, Description: description}
	if err := s.perms.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Service) ListPermissions(ctx context.Context) ([]*domain.Permission, error) {
	return s.perms.List(ctx)
}

func (s *Service) DeletePermission(ctx context.Context, id uuid.UUID) error {
	return s.perms.Delete(ctx, id)
}

// --- Assignments ---

// AssignRoleToUser writes the assignment and a role.assigned outbox row in
// one Postgres transaction (the dual-write fix from architecture decision
// #6), then -- once that's committed -- synchronously invalidates the
// affected user's permission cache. The outbox relay picks up the event
// asynchronously after this call already returned to the caller.
func (s *Service) AssignRoleToUser(ctx context.Context, userID, roleID, assignedBy uuid.UUID) error {
	err := s.txRunner.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.users.AssignRole(ctx, userID, roleID, assignedBy); err != nil {
			return err
		}
		return s.outbox.Create(ctx, "role.assigned", map[string]interface{}{
			"user_id": userID, "role_id": roleID, "assigned_by": assignedBy,
		})
	})
	if err != nil {
		return err
	}
	metrics.RoleAssignmentsTotal.Inc()
	if s.cache != nil {
		_ = s.cache.Invalidate(ctx, userID)
	}
	return nil
}

func (s *Service) RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error {
	err := s.txRunner.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.users.RemoveRole(ctx, userID, roleID); err != nil {
			return err
		}
		return s.outbox.Create(ctx, "role.revoked", map[string]interface{}{
			"user_id": userID, "role_id": roleID,
		})
	})
	if err != nil {
		return err
	}
	if s.cache != nil {
		_ = s.cache.Invalidate(ctx, userID)
	}
	return nil
}

// AttachPermissionToRole invalidates every user currently holding that
// role, not just one user -- a permission change at the role level fans
// out to everyone who inherits it. Invalidation failures are logged
// implicitly by being swallowed; the 10-minute cache TTL is the backstop
// if a DEL is ever missed.
func (s *Service) AttachPermissionToRole(ctx context.Context, roleID, permissionID uuid.UUID) error {
	err := s.txRunner.RunInTx(ctx, func(ctx context.Context) error {
		if err := s.roles.AttachPermission(ctx, roleID, permissionID); err != nil {
			return err
		}
		return s.outbox.Create(ctx, "permission.attached", map[string]interface{}{
			"role_id": roleID, "permission_id": permissionID,
		})
	})
	if err != nil {
		return err
	}
	if s.cache != nil {
		if userIDs, uErr := s.roles.GetUserIDsForRole(ctx, roleID); uErr == nil {
			for _, uid := range userIDs {
				_ = s.cache.Invalidate(ctx, uid)
			}
		}
	}
	return nil
}

// Authorize is the core RBAC check, used by the /authorize handler, by
// middleware.RequirePermission, and by the user handlers' self-or-permission
// checks. It's cache-aside in front of Postgres: a cache hit answers
// straight from Redis; a miss (including "Redis is down") falls straight
// through to the same Postgres join Phase 1 always used, so a Redis outage
// degrades latency, never correctness or availability (architecture spec §3.5).
func (s *Service) Authorize(ctx context.Context, userID uuid.UUID, permissionName string) (bool, error) {
	if s.cache != nil {
		if cached, ok := s.cache.Get(ctx, userID); ok {
			metrics.RedisCacheHitTotal.Inc()
			allowed := cached[permissionName]
			metrics.PermissionChecksTotal.WithLabelValues("hit", resultLabel(allowed)).Inc()
			return allowed, nil
		}
	}
	metrics.RedisCacheMissTotal.Inc()

	perms, err := s.users.GetPermissions(ctx, userID)
	if err != nil {
		return false, err
	}
	names := make([]string, len(perms))
	allowed := false
	for i, p := range perms {
		names[i] = p.Name
		if p.Name == permissionName {
			allowed = true
		}
	}
	if s.cache != nil {
		s.cache.Set(ctx, userID, names)
	}
	metrics.PermissionChecksTotal.WithLabelValues("miss", resultLabel(allowed)).Inc()
	return allowed, nil
}

func resultLabel(allowed bool) string {
	if allowed {
		return "allowed"
	}
	return "denied"
}
