package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const permissionTTL = 10 * time.Minute

// PermissionCache is a thin cache-aside wrapper, not a source of truth.
// Every method here is nil-safe and swallows Redis errors rather than
// returning them: a Redis outage degrades this to "always miss, fall
// through to Postgres" rather than ever causing a request to fail or an
// authorize check to return a wrong answer. See architecture spec §3.5.
type PermissionCache struct {
	rdb *redis.Client
}

func NewPermissionCache(rdb *redis.Client) *PermissionCache {
	return &PermissionCache{rdb: rdb}
}

func permissionsKey(userID uuid.UUID) string {
	return fmt.Sprintf("user:%s:permissions", userID)
}

// Get returns the cached permission set and whether the cache was usable.
// A miss and a Redis error look identical to the caller on purpose.
func (c *PermissionCache) Get(ctx context.Context, userID uuid.UUID) (map[string]bool, bool) {
	if c == nil || c.rdb == nil {
		return nil, false
	}
	members, err := c.rdb.SMembers(ctx, permissionsKey(userID)).Result()
	if err != nil || len(members) == 0 {
		return nil, false
	}
	set := make(map[string]bool, len(members))
	for _, m := range members {
		set[m] = true
	}
	return set, true
}

func (c *PermissionCache) Set(ctx context.Context, userID uuid.UUID, permissions []string) {
	if c == nil || c.rdb == nil {
		return
	}
	key := permissionsKey(userID)
	if len(permissions) == 0 {
		// Cache the "no permissions" outcome too (as a short-lived
		// sentinel) so a permission-less user doesn't hammer Postgres on
		// every single request -- but keep the TTL short since this is
		// the unusual case.
		_ = c.rdb.Set(ctx, key+":empty", "1", time.Minute).Err()
		return
	}
	members := make([]interface{}, len(permissions))
	for i, p := range permissions {
		members[i] = p
	}
	pipe := c.rdb.TxPipeline()
	pipe.Del(ctx, key)
	pipe.SAdd(ctx, key, members...)
	pipe.Expire(ctx, key, permissionTTL)
	_, _ = pipe.Exec(ctx)
}

// Invalidate is called synchronously, right after a role/permission write
// commits -- this is the "synchronous invalidation" from architecture
// decision #2 that keeps the cache from ever serving a stale "allow".
func (c *PermissionCache) Invalidate(ctx context.Context, userID uuid.UUID) error {
	if c == nil || c.rdb == nil {
		return nil
	}
	return c.rdb.Del(ctx, permissionsKey(userID), permissionsKey(userID)+":empty").Err()
}
