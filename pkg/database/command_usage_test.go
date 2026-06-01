package database

import (
	"context"
	"testing"
)

// TestTrackCommandInvocationAggregateOnly covers the DM / bot-invoker path:
// usageKey is non-empty but the per-user IDs aren't, so only the aggregate
// CommandUsage row gets bumped — no User row materializes.
func TestTrackCommandInvocationAggregateOnly(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if err := db.TrackCommandInvocation(ctx, "image filter blur", "", ""); err != nil {
		t.Fatalf("track: %v", err)
	}

	count, err := db.CommandUsageCount(ctx, "image filter blur")
	if err != nil {
		t.Fatalf("aggregate count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected aggregate=1, got %d", count)
	}

	// Aggregate-only path must NOT materialize a User row.
	var userCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM User`).Scan(&userCount); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if userCount != 0 {
		t.Errorf("expected 0 users on aggregate-only path, got %d", userCount)
	}
}

// TestTrackCommandInvocationFullPath covers a normal slash invocation: both
// counters fire in one call and the User row appears on first contact.
func TestTrackCommandInvocationFullPath(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if err := db.TrackCommandInvocation(ctx, "image filter blur", "g", "u"); err != nil {
		t.Fatalf("track: %v", err)
	}

	// Aggregate bumped.
	count, err := db.CommandUsageCount(ctx, "image filter blur")
	if err != nil {
		t.Fatalf("aggregate count: %v", err)
	}
	if count != 1 {
		t.Errorf("expected aggregate=1, got %d", count)
	}

	// User materialized + per-user count = 1.
	g, err := db.GuildByDiscordID(ctx, "g")
	if err != nil || g == nil {
		t.Fatalf("guild lookup: g=%v err=%v", g, err)
	}
	u, err := db.GetUserByDiscordID(ctx, "u", g.ID)
	if err != nil || u == nil {
		t.Fatalf("user lookup: u=%v err=%v", u, err)
	}
	total, err := db.GetUserCommandTotal(ctx, u.ID)
	if err != nil {
		t.Fatalf("user total: %v", err)
	}
	if total != 1 {
		t.Errorf("expected per-user total=1, got %d", total)
	}
}

// TestTrackCommandInvocationPrefixKey verifies the "$"-canonical keys used by
// prefix commands flow through unchanged — they're just another usage key as
// far as the tracker is concerned. Two different prefix-style keys produce
// two separate aggregate rows.
func TestTrackCommandInvocationPrefixKey(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	for range 3 {
		if err := db.TrackCommandInvocation(ctx, "$goodboy", "g", "u"); err != nil {
			t.Fatalf("track goodboy: %v", err)
		}
	}
	if err := db.TrackCommandInvocation(ctx, "$romans", "g", "u"); err != nil {
		t.Fatalf("track romans: %v", err)
	}

	if got, err := db.CommandUsageCount(ctx, "$goodboy"); err != nil || got != 3 {
		t.Errorf("$goodboy aggregate = %d (err=%v), want 3", got, err)
	}
	if got, err := db.CommandUsageCount(ctx, "$romans"); err != nil || got != 1 {
		t.Errorf("$romans aggregate = %d (err=%v), want 1", got, err)
	}
}

// TestTrackCommandInvocationEmptyKeyNoOp guards the early-return: an empty
// usageKey is a misconfigured call site, but it should be silent rather than
// inserting a sentinel empty row.
func TestTrackCommandInvocationEmptyKeyNoOp(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if err := db.TrackCommandInvocation(ctx, "", "g", "u"); err != nil {
		t.Fatalf("track empty key: %v", err)
	}

	var cmdCount, userCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM CommandUsage`).Scan(&cmdCount); err != nil {
		t.Fatalf("count CommandUsage: %v", err)
	}
	if cmdCount != 0 {
		t.Errorf("expected 0 CommandUsage rows on empty-key call, got %d", cmdCount)
	}
	if err := db.QueryRow(`SELECT COUNT(*) FROM User`).Scan(&userCount); err != nil {
		t.Fatalf("count User: %v", err)
	}
	if userCount != 0 {
		t.Errorf("expected 0 User rows on empty-key call, got %d", userCount)
	}
}
