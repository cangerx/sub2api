package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func mgrTestRouter() *GroupRouter {
	return NewGroupRouter(NewGroupCooldownStore(nil), &config.Config{})
}

func mgrGroup(id int64, platform string, sub bool) *Group {
	subType := SubscriptionTypeStandard
	if sub {
		subType = SubscriptionTypeSubscription
	}
	return &Group{ID: id, Platform: platform, Status: "active", SubscriptionType: subType}
}

func mgrKey(bindings ...APIKeyGroupBinding) *APIKey {
	return &APIKey{ID: 1, MultiGroupRouting: true, GroupBindings: bindings}
}

func TestSelectGroup_SingleGroupModeUnchanged(t *testing.T) {
	r := mgrTestRouter()
	g := mgrGroup(7, "openai", false)
	key := &APIKey{ID: 1, MultiGroupRouting: false, Group: g}
	got, err := r.SelectGroup(context.Background(), key, "openai", nil)
	if err != nil || got == nil || got.ID != 7 {
		t.Fatalf("single-group mode should return bound group; got=%v err=%v", got, err)
	}
}

func TestSelectGroup_PriorityBucketWins(t *testing.T) {
	r := mgrTestRouter()
	key := mgrKey(
		APIKeyGroupBinding{GroupID: 10, Priority: 0, Weight: 100, Enabled: true, Group: mgrGroup(10, "openai", false)},
		APIKeyGroupBinding{GroupID: 20, Priority: 5, Weight: 100, Enabled: true, Group: mgrGroup(20, "openai", false)},
	)
	for i := 0; i < 50; i++ {
		got, err := r.SelectGroup(context.Background(), key, "openai", nil)
		if err != nil || got.ID != 10 {
			t.Fatalf("highest-priority group(10) should always win; got=%v err=%v", got, err)
		}
	}
}

func TestSelectGroup_WeightedWithinBucket(t *testing.T) {
	r := mgrTestRouter()
	key := mgrKey(
		APIKeyGroupBinding{GroupID: 10, Priority: 0, Weight: 90, Enabled: true, Group: mgrGroup(10, "openai", false)},
		APIKeyGroupBinding{GroupID: 20, Priority: 0, Weight: 10, Enabled: true, Group: mgrGroup(20, "openai", false)},
	)
	counts := map[int64]int{}
	for i := 0; i < 2000; i++ {
		got, _ := r.SelectGroup(context.Background(), key, "openai", nil)
		counts[got.ID]++
	}
	if counts[10] <= counts[20] {
		t.Fatalf("weight 90 vs 10 should favour group 10; counts=%v", counts)
	}
}

func TestSelectGroup_SkipsSubscriptionAndDisabled(t *testing.T) {
	r := mgrTestRouter()
	key := mgrKey(
		APIKeyGroupBinding{GroupID: 10, Priority: 0, Weight: 100, Enabled: true, Group: mgrGroup(10, "openai", true)}, // subscription -> skip
		APIKeyGroupBinding{GroupID: 20, Priority: 0, Weight: 100, Enabled: false, Group: mgrGroup(20, "openai", false)}, // disabled -> skip
		APIKeyGroupBinding{GroupID: 30, Priority: 0, Weight: 100, Enabled: true, Group: mgrGroup(30, "openai", false)}, // only valid
	)
	got, err := r.SelectGroup(context.Background(), key, "openai", nil)
	if err != nil || got.ID != 30 {
		t.Fatalf("should pick the only non-subscription enabled group(30); got=%v err=%v", got, err)
	}
}

func TestSelectGroup_PlatformScopeFilter(t *testing.T) {
	r := mgrTestRouter()
	key := mgrKey(
		APIKeyGroupBinding{GroupID: 10, Priority: 0, Weight: 100, Enabled: true, Group: mgrGroup(10, "anthropic", false)},
		APIKeyGroupBinding{GroupID: 20, Priority: 1, Weight: 100, Enabled: true, Group: mgrGroup(20, "openai", false)},
	)
	got, err := r.SelectGroup(context.Background(), key, "openai", nil)
	if err != nil || got.ID != 20 {
		t.Fatalf("scope=openai should pick group 20; got=%v err=%v", got, err)
	}
}

func TestSelectGroup_CoolingSkippedThenFallback(t *testing.T) {
	r := mgrTestRouter()
	ctx := context.Background()
	key := mgrKey(
		APIKeyGroupBinding{GroupID: 10, Priority: 0, Weight: 100, Enabled: true, Group: mgrGroup(10, "openai", false)},
		APIKeyGroupBinding{GroupID: 20, Priority: 1, Weight: 100, Enabled: true, Group: mgrGroup(20, "openai", false)},
	)
	// Cool down the top-priority group -> should switch to 20.
	r.MarkGroupCooldown(ctx, key.ID, 10)
	got, err := r.SelectGroup(ctx, key, "openai", nil)
	if err != nil || got.ID != 20 {
		t.Fatalf("cooling group 10 should be skipped for 20; got=%v err=%v", got, err)
	}
	// Cool down both -> fallback to shortest-remaining (no hard fail).
	r.MarkGroupCooldown(ctx, key.ID, 20)
	got, err = r.SelectGroup(ctx, key, "openai", nil)
	if err != nil || got == nil {
		t.Fatalf("all-cooling should fall back to a group, not error; got=%v err=%v", got, err)
	}
}

func TestSelectGroup_ExcludedSkipped(t *testing.T) {
	r := mgrTestRouter()
	key := mgrKey(
		APIKeyGroupBinding{GroupID: 10, Priority: 0, Weight: 100, Enabled: true, Group: mgrGroup(10, "openai", false)},
		APIKeyGroupBinding{GroupID: 20, Priority: 1, Weight: 100, Enabled: true, Group: mgrGroup(20, "openai", false)},
	)
	got, err := r.SelectGroup(context.Background(), key, "openai", map[int64]struct{}{10: {}})
	if err != nil || got.ID != 20 {
		t.Fatalf("excluded group 10 should yield 20; got=%v err=%v", got, err)
	}
}

func TestGroupCooldown_AutoRecovers(t *testing.T) {
	store := NewGroupCooldownStore(nil)
	store.MarkCooldown(context.Background(), 1, 5, 40*time.Millisecond)
	if cooling, _ := store.IsCooling(context.Background(), 1, 5); !cooling {
		t.Fatal("should be cooling right after mark")
	}
	time.Sleep(60 * time.Millisecond)
	if cooling, _ := store.IsCooling(context.Background(), 1, 5); cooling {
		t.Fatal("cooldown should auto-expire on read")
	}
}
