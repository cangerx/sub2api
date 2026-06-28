package handler

import (
	"context"

	"github.com/Wei-Shaw/ccapi/internal/service"
)

// groupFailover encapsulates per-request multi-group routing for a gateway
// handler. With multi-group routing disabled it is a thin pass-through that
// always yields the key's nominal group (zero behaviour change). With it
// enabled it selects an initial group by priority/weight and, on account
// exhaustion / persistent upstream failure, cools the current group down and
// advances to the next eligible one within the same request.
type groupFailover struct {
	router      *service.GroupRouter
	apiKey      *service.APIKey
	platform    string
	enabled     bool
	excluded    map[int64]struct{}
	cur         *service.Group
	switches    int
	maxSwitches int
}

// newGroupFailover resolves the initial group for the request. platform is the
// endpoint platform (the key's nominal group platform) used as the routing
// scope. It never returns nil; cur falls back to apiKey.Group.
func (h *GatewayHandler) newGroupFailover(ctx context.Context, apiKey *service.APIKey, platform string) *groupFailover {
	gf := &groupFailover{
		router:   h.groupRouter,
		apiKey:   apiKey,
		platform: platform,
		excluded: make(map[int64]struct{}),
	}
	if h.groupRouter != nil && h.groupRouter.MultiGroupEnabled(apiKey) {
		gf.enabled = true
		gf.maxSwitches = h.groupRouter.MaxGroupSwitches()
		if g, err := h.groupRouter.SelectGroup(ctx, apiKey, platform, nil); err == nil && g != nil {
			gf.cur = g
		}
	}
	if gf.cur == nil {
		gf.cur = apiKey.Group
	}
	return gf
}

// group returns the current effective group (may be nil if the key has none).
func (gf *groupFailover) group() *service.Group {
	return gf.cur
}

// groupID returns the current effective group ID pointer for selection/billing.
func (gf *groupFailover) groupID() *int64 {
	if gf.cur != nil {
		return &gf.cur.ID
	}
	return gf.apiKey.GroupID
}

// platformOf returns the current group's platform (falls back to the scope).
func (gf *groupFailover) platformOf() string {
	if gf.cur != nil {
		return gf.cur.Platform
	}
	return gf.platform
}

// advance cools down the current group and selects the next eligible one.
// Returns false when multi-group routing is off, the switch cap is hit, or no
// other group is available — in which case the caller should give up as before.
func (gf *groupFailover) advance(ctx context.Context) bool {
	if !gf.enabled || gf.router == nil {
		return false
	}
	if gf.switches >= gf.maxSwitches {
		return false
	}
	if gf.cur != nil {
		gf.router.MarkGroupCooldown(ctx, gf.apiKey.ID, gf.cur.ID)
		gf.excluded[gf.cur.ID] = struct{}{}
	}
	next, err := gf.router.SelectGroup(ctx, gf.apiKey, gf.platform, gf.excluded)
	if err != nil || next == nil {
		return false
	}
	// SelectGroup may return a cooling fallback already in excluded; treat that
	// as "no new group" to avoid an infinite loop.
	if _, seen := gf.excluded[next.ID]; seen {
		return false
	}
	gf.cur = next
	gf.switches++
	return true
}
