package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"rbac-platform/internal/domain"
	"rbac-platform/internal/platform/cache"
	"rbac-platform/internal/platform/jwt"
	"rbac-platform/internal/platform/metrics"
)

type Service struct {
	users      domain.UserRepository
	roles      domain.RoleRepository
	tokens     domain.RefreshTokenRepository
	jwtMgr     *jwt.Manager
	refreshTTL time.Duration
	blacklist  *cache.Blacklist
}

func NewService(
	users domain.UserRepository,
	roles domain.RoleRepository,
	tokens domain.RefreshTokenRepository,
	jwtMgr *jwt.Manager,
	refreshTTL time.Duration,
	blacklist *cache.Blacklist,
) *Service {
	return &Service{users: users, roles: roles, tokens: tokens, jwtMgr: jwtMgr, refreshTTL: refreshTTL, blacklist: blacklist}
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

func (s *Service) Signup(ctx context.Context, email, password, fullName string) (*domain.User, error) {
	_, err := s.users.GetByEmail(ctx, email)
	if err == nil {
		return nil, domain.ErrAlreadyExists
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	u := &domain.User{
		Email:        email,
		PasswordHash: string(hash),
		FullName:     fullName,
		IsActive:     true,
	}
	if err := s.users.Create(ctx, u); err != nil {
		return nil, err
	}
	metrics.UsersCreatedTotal.Inc()

	// Best-effort default role. If the "developer" seed role is missing
	// (e.g. migrations not run), signup still succeeds -- the user just
	// starts with no roles instead of failing outright.
	if defaultRole, rErr := s.roles.GetByName(ctx, "developer"); rErr == nil {
		_ = s.users.AssignRole(ctx, u.ID, defaultRole.ID, u.ID)
	}

	return u, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	u, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			metrics.LoginFailureTotal.Inc()
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if !u.IsActive {
		metrics.LoginFailureTotal.Inc()
		return nil, domain.ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		metrics.LoginFailureTotal.Inc()
		return nil, domain.ErrInvalidCredentials
	}
	metrics.LoginSuccessTotal.Inc()
	return s.issueTokenPair(ctx, u)
}

func (s *Service) Refresh(ctx context.Context, rawRefresh string) (*TokenPair, error) {
	hash := hashToken(rawRefresh)
	stored, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	if stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		return nil, domain.ErrTokenExpired
	}
	// Rotate: the presented refresh token is single-use. Revoking it here,
	// before issuing the replacement, means a stolen-and-replayed old
	// token can never mint a second valid session.
	if err := s.tokens.Revoke(ctx, stored.ID); err != nil {
		return nil, err
	}

	u, err := s.users.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, err
	}
	return s.issueTokenPair(ctx, u)
}

// Logout revokes the refresh token and blacklists the access token's jti
// (if Redis is configured) so it's rejected immediately instead of being
// allowed to keep working until its natural 15-minute expiry.
func (s *Service) Logout(ctx context.Context, rawRefresh, jti string, accessTokenRemaining time.Duration) error {
	if s.blacklist != nil && jti != "" {
		_ = s.blacklist.Add(ctx, jti, accessTokenRemaining)
	}
	hash := hashToken(rawRefresh)
	stored, err := s.tokens.GetByHash(ctx, hash)
	if err != nil {
		return nil // logout is idempotent: an unknown token is already "logged out"
	}
	return s.tokens.Revoke(ctx, stored.ID)
}

func (s *Service) issueTokenPair(ctx context.Context, u *domain.User) (*TokenPair, error) {
	roles, err := s.users.GetRoles(ctx, u.ID)
	if err != nil {
		return nil, err
	}
	roleNames := make([]string, len(roles))
	for i, r := range roles {
		roleNames[i] = r.Name
	}

	access, _, err := s.jwtMgr.GenerateAccessToken(u.ID, roleNames)
	if err != nil {
		return nil, err
	}

	rawRefresh, hash := generateRefreshToken()
	if err := s.tokens.Create(ctx, &domain.RefreshToken{
		UserID:    u.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(s.refreshTTL),
	}); err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: access, RefreshToken: rawRefresh}, nil
}

func generateRefreshToken() (raw string, hash string) {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	raw = hex.EncodeToString(b)
	return raw, hashToken(raw)
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
