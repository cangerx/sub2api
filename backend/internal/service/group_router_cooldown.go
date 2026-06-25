package service

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// GroupCooldownStore tracks per-(apiKey, group) cooldown for multi-group routing.
// A group entering cooldown is skipped by SelectGroup until the TTL expires
// (auto-recovery on read). Backed by Redis when available, with an in-memory
// fallback so single-node / no-Redis deployments still work.
type GroupCooldownStore interface {
	// MarkCooldown puts (apiKeyID, groupID) into cooldown for ttl.
	MarkCooldown(ctx context.Context, apiKeyID, groupID int64, ttl time.Duration)
	// IsCooling reports whether (apiKeyID, groupID) is currently cooling, and
	// the remaining duration if so.
	IsCooling(ctx context.Context, apiKeyID, groupID int64) (bool, time.Duration)
	// ClearCooldown removes any cooldown for (apiKeyID, groupID).
	ClearCooldown(ctx context.Context, apiKeyID, groupID int64)
}

func groupCooldownKey(apiKeyID, groupID int64) string {
	return fmt.Sprintf("cooldown:group:%d:%d", apiKeyID, groupID)
}

// GroupCooldownRedis is the optional Redis client surface GroupCooldownStore needs.
// It mirrors the subset of *redis.Client used here so the service package stays
// decoupled from the redis import (the repository layer adapts the real client).
type GroupCooldownRedis interface {
	GroupCooldownSet(ctx context.Context, key string, ttl time.Duration) error
	GroupCooldownTTL(ctx context.Context, key string) (time.Duration, error)
	GroupCooldownDel(ctx context.Context, key string) error
}

// memoryGroupCooldown is the in-memory fallback (and also fronts Redis as a
// fast local cache to avoid a round-trip on every selection).
type memoryGroupCooldown struct {
	mu   sync.Mutex
	until map[string]time.Time
}

func newMemoryGroupCooldown() *memoryGroupCooldown {
	return &memoryGroupCooldown{until: make(map[string]time.Time)}
}

func (m *memoryGroupCooldown) set(key string, ttl time.Duration) {
	m.mu.Lock()
	m.until[key] = time.Now().Add(ttl)
	m.mu.Unlock()
}

func (m *memoryGroupCooldown) ttl(key string) (bool, time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	exp, ok := m.until[key]
	if !ok {
		return false, 0
	}
	remaining := time.Until(exp)
	if remaining <= 0 {
		delete(m.until, key) // lazy cleanup on read (auto-recovery)
		return false, 0
	}
	return true, remaining
}

func (m *memoryGroupCooldown) del(key string) {
	m.mu.Lock()
	delete(m.until, key)
	m.mu.Unlock()
}

// groupCooldownStore is the default GroupCooldownStore: Redis-backed when a
// client is provided, always mirrored to an in-memory map for fast reads and
// as the sole backend when Redis is absent.
type groupCooldownStore struct {
	redis GroupCooldownRedis // optional
	mem   *memoryGroupCooldown
}

// NewGroupCooldownStore builds a cooldown store. redis may be nil (memory-only).
func NewGroupCooldownStore(redis GroupCooldownRedis) GroupCooldownStore {
	return &groupCooldownStore{redis: redis, mem: newMemoryGroupCooldown()}
}

func (s *groupCooldownStore) MarkCooldown(ctx context.Context, apiKeyID, groupID int64, ttl time.Duration) {
	if ttl <= 0 {
		return
	}
	key := groupCooldownKey(apiKeyID, groupID)
	s.mem.set(key, ttl)
	if s.redis != nil {
		_ = s.redis.GroupCooldownSet(ctx, key, ttl)
	}
}

func (s *groupCooldownStore) IsCooling(ctx context.Context, apiKeyID, groupID int64) (bool, time.Duration) {
	key := groupCooldownKey(apiKeyID, groupID)
	// Local memory first (covers no-Redis and avoids a round-trip).
	if cooling, remaining := s.mem.ttl(key); cooling {
		return true, remaining
	}
	if s.redis == nil {
		return false, 0
	}
	remaining, err := s.redis.GroupCooldownTTL(ctx, key)
	if err != nil || remaining <= 0 {
		return false, 0
	}
	// Backfill local cache so subsequent reads in this process are cheap.
	s.mem.set(key, remaining)
	return true, remaining
}

func (s *groupCooldownStore) ClearCooldown(ctx context.Context, apiKeyID, groupID int64) {
	key := groupCooldownKey(apiKeyID, groupID)
	s.mem.del(key)
	if s.redis != nil {
		_ = s.redis.GroupCooldownDel(ctx, key)
	}
}
