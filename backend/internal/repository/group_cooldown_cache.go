package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/Wei-Shaw/ccapi/internal/service"
)

// groupCooldownRedis adapts *redis.Client to service.GroupCooldownRedis for
// multi-group-routing group cooldown. May be nil-backed safely upstream:
// NewGroupCooldownRedis returns nil when rdb is nil so the service falls back
// to its in-memory store.
type groupCooldownRedis struct {
	rdb *redis.Client
}

// NewGroupCooldownRedis returns a Redis-backed service.GroupCooldownRedis, or
// nil when rdb is nil (memory-only cooldown).
func NewGroupCooldownRedis(rdb *redis.Client) service.GroupCooldownRedis {
	if rdb == nil {
		return nil
	}
	return &groupCooldownRedis{rdb: rdb}
}

func (c *groupCooldownRedis) GroupCooldownSet(ctx context.Context, key string, ttl time.Duration) error {
	return c.rdb.Set(ctx, key, "1", ttl).Err()
}

func (c *groupCooldownRedis) GroupCooldownTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.rdb.TTL(ctx, key).Result()
}

func (c *groupCooldownRedis) GroupCooldownDel(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}
