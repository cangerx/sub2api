package service

import (
	"context"
	"errors"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/Wei-Shaw/ccapi/internal/config"
)

// ErrNoRoutableGroup is returned when multi-group routing is enabled but no
// bound group is currently usable (all disabled / unavailable / cooling and
// excluded).
var ErrNoRoutableGroup = errors.New("no routable group available")

// GroupRouter selects a target group for an API key request. With multi-group
// routing disabled it returns the key's single bound group (zero behaviour
// change). With it enabled, it picks among the key's GroupBindings by
// priority (lower first) then weighted-random within the top bucket, skipping
// subscription / unavailable / cooling / excluded groups.
type GroupRouter struct {
	cooldown GroupCooldownStore
	cfg      *config.Config
}

// NewGroupRouter builds a GroupRouter. cooldown must be non-nil.
func NewGroupRouter(cooldown GroupCooldownStore, cfg *config.Config) *GroupRouter {
	return &GroupRouter{cooldown: cooldown, cfg: cfg}
}

// MaxGroupSwitches is the per-request group switch cap (failover bound).
func (r *GroupRouter) MaxGroupSwitches() int {
	if r.cfg != nil && r.cfg.Gateway.MaxGroupSwitches > 0 {
		return r.cfg.Gateway.MaxGroupSwitches
	}
	return 5
}

// GroupCooldown is the TTL applied when a group is cooled down after a failure.
func (r *GroupRouter) GroupCooldown() time.Duration {
	seconds := 60
	if r.cfg != nil && r.cfg.Gateway.GroupCooldownSeconds > 0 {
		seconds = r.cfg.Gateway.GroupCooldownSeconds
	}
	return time.Duration(seconds) * time.Second
}

// MarkGroupCooldown cools down a group for this key (failover support).
func (r *GroupRouter) MarkGroupCooldown(ctx context.Context, apiKeyID, groupID int64) {
	r.cooldown.MarkCooldown(ctx, apiKeyID, groupID, r.GroupCooldown())
}

// MultiGroupEnabled reports whether the key uses multi-group routing with at
// least one usable (enabled, non-subscription, active) binding.
func (r *GroupRouter) MultiGroupEnabled(apiKey *APIKey) bool {
	if apiKey == nil || !apiKey.MultiGroupRouting {
		return false
	}
	for i := range apiKey.GroupBindings {
		if isRoutableBinding(&apiKey.GroupBindings[i], "") {
			return true
		}
	}
	return false
}

// SelectGroup returns the group to use for this request.
//   - scope: when non-empty, only groups whose platform matches are eligible
//     (an endpoint serves a single platform). Empty = no platform filter.
//   - excluded: group IDs already tried in this request (failover).
//
// Single-group keys return apiKey.Group unchanged.
func (r *GroupRouter) SelectGroup(ctx context.Context, apiKey *APIKey, scope string, excluded map[int64]struct{}) (*Group, error) {
	if apiKey == nil {
		return nil, ErrNoRoutableGroup
	}
	if !apiKey.MultiGroupRouting {
		return apiKey.Group, nil
	}

	type candidate struct {
		group    *Group
		priority int
		weight   int
	}
	var eligible []candidate   // not cooling, usable
	var coolingFallback *Group // shortest-remaining cooling group (last resort)
	var coolingFallbackRemaining time.Duration

	for i := range apiKey.GroupBindings {
		b := &apiKey.GroupBindings[i]
		if !isRoutableBinding(b, scope) {
			continue
		}
		if excluded != nil {
			if _, skip := excluded[b.GroupID]; skip {
				continue
			}
		}
		cooling, remaining := r.cooldown.IsCooling(ctx, apiKey.ID, b.GroupID)
		if cooling {
			if coolingFallback == nil || remaining < coolingFallbackRemaining {
				coolingFallback = b.Group
				coolingFallbackRemaining = remaining
			}
			continue
		}
		weight := b.Weight
		if weight <= 0 {
			weight = 1
		}
		eligible = append(eligible, candidate{group: b.Group, priority: b.Priority, weight: weight})
	}

	if len(eligible) == 0 {
		// Everything usable is cooling: fall back to the soonest-recovering group
		// rather than hard-failing (avoids a full black-out).
		if coolingFallback != nil {
			return coolingFallback, nil
		}
		return nil, ErrNoRoutableGroup
	}

	// Highest-priority bucket = smallest priority value.
	minPriority := eligible[0].priority
	for _, c := range eligible[1:] {
		if c.priority < minPriority {
			minPriority = c.priority
		}
	}
	bucket := eligible[:0:0]
	totalWeight := 0
	for _, c := range eligible {
		if c.priority == minPriority {
			bucket = append(bucket, c)
			totalWeight += c.weight
		}
	}
	if len(bucket) == 1 || totalWeight <= 0 {
		return bucket[0].group, nil
	}

	// Weighted random within the bucket (roulette wheel).
	target := rand.IntN(totalWeight)
	acc := 0
	for _, c := range bucket {
		acc += c.weight
		if target < acc {
			return c.group, nil
		}
	}
	return bucket[len(bucket)-1].group, nil
}

// isRoutableBinding applies the phase-1 eligibility rules: enabled, group
// loaded & active, non-subscription, and platform matches scope when given.
func isRoutableBinding(b *APIKeyGroupBinding, scope string) bool {
	if b == nil || !b.Enabled || b.Group == nil {
		return false
	}
	g := b.Group
	if !g.IsActive() {
		return false
	}
	// Phase 1: multi-group routing only for non-subscription (standard) groups.
	if g.IsSubscriptionType() {
		return false
	}
	if scope != "" && !strings.EqualFold(strings.TrimSpace(g.Platform), strings.TrimSpace(scope)) {
		return false
	}
	return true
}
