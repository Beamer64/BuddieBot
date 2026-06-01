package database

import (
	"context"
	"testing"
)

// TestIncrementUserCommandUsage exercises the happy path: increment a known
// (user, guild) for a few command names, then read back the total + top.
func TestIncrementUserCommandUsage(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	u, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("ensure user: %v", err)
	}

	// Three hits on "image filter blur", one on "daily horoscope".
	for range 3 {
		if err := db.IncrementUserCommandUsage(ctx, "g", "u", "image filter blur"); err != nil {
			t.Fatalf("increment image: %v", err)
		}
	}
	if err := db.IncrementUserCommandUsage(ctx, "g", "u", "daily horoscope"); err != nil {
		t.Fatalf("increment daily: %v", err)
	}

	total, err := db.GetUserCommandTotal(ctx, u.ID)
	if err != nil {
		t.Fatalf("total: %v", err)
	}
	if total != 4 {
		t.Errorf("expected total 4, got %d", total)
	}

	name, count, err := db.GetUserTopCommand(ctx, u.ID)
	if err != nil {
		t.Fatalf("top: %v", err)
	}
	if name != "image filter blur" || count != 3 {
		t.Errorf("expected top=(image filter blur, 3), got (%q, %d)", name, count)
	}
}

// TestIncrementUserCommandUsageSafeWhenUserMissing exercises the
// defense-in-depth property of the bare IncrementUserCommandUsage: even if
// the orchestrator skips or fails the EnsureUser step, the SELECT-driven
// INSERT still produces a clean no-op (no SQL error, no rows written, no
// constraint violation) rather than corrupting state.
func TestIncrementUserCommandUsageSafeWhenUserMissing(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// No EnsureUser. Guild also doesn't exist. The statement should still
	// succeed (no SQL error), but write zero rows.
	if err := db.IncrementUserCommandUsage(ctx, "ghost-guild", "ghost-user", "image"); err != nil {
		t.Fatalf("increment for missing user shouldn't error: %v", err)
	}

	// And no row materialized in User table either — bare Increment never
	// creates users on its own; that's RecordUserCommandUsage's job.
	var userCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM User`).Scan(&userCount); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if userCount != 0 {
		t.Errorf("expected 0 users after no-op increment, got %d", userCount)
	}

	var usageCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM UserCommandUsage`).Scan(&usageCount); err != nil {
		t.Fatalf("count usage: %v", err)
	}
	if usageCount != 0 {
		t.Errorf("expected 0 usage rows after no-op increment, got %d", usageCount)
	}
}

// TestRecordUserCommandUsageMaterializesUserOnFirstCall locks in the
// "track from first command" contract — a fresh (guild, user) pair gets a
// User row AND a usage row in a single call. This is the change that makes
// /user profile's Commands field accurate from the very first interaction.
func TestRecordUserCommandUsageMaterializesUserOnFirstCall(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if err := db.RecordUserCommandUsage(ctx, "g", "u", "image filter blur"); err != nil {
		t.Fatalf("record: %v", err)
	}

	u, err := db.GetUserByDiscordID(ctx, "u", mustGuildID(t, db, "g"))
	if err != nil || u == nil {
		t.Fatalf("expected User row materialized, got user=%v err=%v", u, err)
	}

	total, err := db.GetUserCommandTotal(ctx, u.ID)
	if err != nil {
		t.Fatalf("total: %v", err)
	}
	if total != 1 {
		t.Errorf("expected total 1 after first record, got %d", total)
	}
}

// TestRecordUserCommandUsageIdempotentExistingUser confirms repeat calls bump
// the counter on the existing row instead of duplicating User rows or
// resetting the count.
func TestRecordUserCommandUsageIdempotentExistingUser(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	for range 3 {
		if err := db.RecordUserCommandUsage(ctx, "g", "u", "image filter blur"); err != nil {
			t.Fatalf("record: %v", err)
		}
	}

	var userCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM User WHERE Discord_UserID = 'u'`).Scan(&userCount); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if userCount != 1 {
		t.Errorf("expected exactly 1 User row across repeat records, got %d", userCount)
	}

	u, err := db.GetUserByDiscordID(ctx, "u", mustGuildID(t, db, "g"))
	if err != nil || u == nil {
		t.Fatalf("user lookup: u=%v err=%v", u, err)
	}
	total, err := db.GetUserCommandTotal(ctx, u.ID)
	if err != nil {
		t.Fatalf("total: %v", err)
	}
	if total != 3 {
		t.Errorf("expected total 3 after three records, got %d", total)
	}
}

// mustGuildID looks up the internal Guild ID for the test guild, failing
// the test if EnsureUser somehow didn't create the parent Guild row.
func mustGuildID(t *testing.T, db *DB, discordGuildID string) int64 {
	t.Helper()
	g, err := db.GuildByDiscordID(context.Background(), discordGuildID)
	if err != nil || g == nil {
		t.Fatalf("guild %q lookup: g=%v err=%v", discordGuildID, g, err)
	}
	return g.ID
}

// TestGetUserTopCommandEmpty confirms a tracked user with zero usage gets the
// "no commands" return shape: empty name + zero count + nil error.
func TestGetUserTopCommandEmpty(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	u, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}

	name, count, err := db.GetUserTopCommand(ctx, u.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if name != "" || count != 0 {
		t.Errorf("expected ('', 0), got (%q, %d)", name, count)
	}

	total, err := db.GetUserCommandTotal(ctx, u.ID)
	if err != nil {
		t.Errorf("total error: %v", err)
	}
	if total != 0 {
		t.Errorf("expected total 0, got %d", total)
	}
}

// TestUserCommandUsageFKCascade confirms /user forget-me's deletion path —
// removing the User row removes its UserCommandUsage rows automatically.
func TestUserCommandUsageFKCascade(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	u, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if err := db.IncrementUserCommandUsage(ctx, "g", "u", "image"); err != nil {
		t.Fatalf("seed usage: %v", err)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM User WHERE ID = ?`, u.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM UserCommandUsage WHERE UserID = ?`, u.ID).Scan(&count); err != nil {
		t.Fatalf("count usage: %v", err)
	}
	if count != 0 {
		t.Errorf("expected cascade delete; %d rows remain", count)
	}
}
