package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Blacklist holds revoked-but-not-yet-expired access token jtis.
type Blacklist struct {
	rdb *redis.Client
}

func NewBlacklist(rdb *redis.Client) *Blacklist {
	return &Blacklist{rdb: rdb}
}

func (b *Blacklist) Add(ctx context.Context, jti string, ttl time.Duration) error {
	if b == nil || b.rdb == nil || ttl <= 0 {
		return nil
	}
	return b.rdb.Set(ctx, "blacklist:"+jti, "1", ttl).Err()
}

// IsBlacklisted fails OPEN: if Redis can't be reached, it reports "not
// blacklisted" rather than rejecting every authenticated request because
// of a cache blip. The accepted trade-off (documented in the architecture
// spec, §3.5) is that a just-logged-out token can keep working until its
// natural 15-minute expiry during a Redis outage -- small and time-bounded,
// versus taking the whole API down on every Redis hiccup.
func (b *Blacklist) IsBlacklisted(ctx context.Context, jti string) bool {
	if b == nil || b.rdb == nil {
		return false
	}
	n, err := b.rdb.Exists(ctx, "blacklist:"+jti).Result()
	if err != nil {
		return false
	}
	return n > 0
}
