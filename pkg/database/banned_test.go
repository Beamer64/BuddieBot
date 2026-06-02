package database

import (
	"context"
	"testing"
)

// TestIsUserBanned covers the hot-path read used before every command
// dispatch: not-banned returns false cleanly, banned returns true, empty
// input short-circuits without errors (defensive default).
func TestIsUserBanned(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if got, err := db.IsUserBanned(ctx, "u1"); err != nil || got {
		t.Errorf("not-yet-banned user: got %v (err=%v), want false", got, err)
	}

	if _, err := db.BanUser(ctx, "u1", "spam", "admin1"); err != nil {
		t.Fatalf("seed ban: %v", err)
	}
	if got, err := db.IsUserBanned(ctx, "u1"); err != nil || !got {
		t.Errorf("banned user: got %v (err=%v), want true", got, err)
	}

	// Empty input must NOT touch the DB — return false cleanly. Defends the
	// dispatcher from spurious "banned" verdicts when invokerID is empty
	// (DM context with no user data, etc.).
	if got, err := db.IsUserBanned(ctx, ""); err != nil || got {
		t.Errorf("empty userID: got %v (err=%v), want false", got, err)
	}
}

// TestBanUser_CreateThenUpdate locks in the "created vs updated" return
// contract — first call returns true, second call (same user, new reason)
// returns false and refreshes the metadata.
func TestBanUser_CreateThenUpdate(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	created, err := db.BanUser(ctx, "u1", "spamming /image", "admin1")
	if err != nil {
		t.Fatalf("first ban: %v", err)
	}
	if !created {
		t.Errorf("first ban should report created=true, got false")
	}

	created, err = db.BanUser(ctx, "u1", "actually it was /generate", "admin2")
	if err != nil {
		t.Fatalf("second ban: %v", err)
	}
	if created {
		t.Errorf("second ban (same user) should report created=false, got true")
	}

	row, err := db.GetBannedUser(ctx, "u1")
	if err != nil || row == nil {
		t.Fatalf("get banned: row=%v err=%v", row, err)
	}
	if row.Reason.String != "actually it was /generate" {
		t.Errorf("expected updated reason, got %q", row.Reason.String)
	}
	if row.BannedBy.String != "admin2" {
		t.Errorf("expected updated issuer, got %q", row.BannedBy.String)
	}
}

// TestBanUser_NullableFields confirms empty reason/issuer arrive as SQL NULL
// (Valid=false) rather than empty strings. The display layer can then
// distinguish "no reason given" from "reason was literally blank".
func TestBanUser_NullableFields(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if _, err := db.BanUser(ctx, "u1", "", ""); err != nil {
		t.Fatalf("ban with empty meta: %v", err)
	}

	row, err := db.GetBannedUser(ctx, "u1")
	if err != nil || row == nil {
		t.Fatalf("get banned: row=%v err=%v", row, err)
	}
	if row.Reason.Valid {
		t.Errorf("expected Reason NULL when empty supplied, got %q (Valid=true)", row.Reason.String)
	}
	if row.BannedBy.Valid {
		t.Errorf("expected BannedBy NULL when empty supplied, got %q (Valid=true)", row.BannedBy.String)
	}
}

// TestUnbanUser covers both the "found and removed" and "no row" paths.
func TestUnbanUser(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Unbanning a non-banned user is a clean no-op, returns false.
	if removed, err := db.UnbanUser(ctx, "ghost"); err != nil || removed {
		t.Errorf("unban ghost: got removed=%v (err=%v), want false", removed, err)
	}

	if _, err := db.BanUser(ctx, "u1", "test", "admin1"); err != nil {
		t.Fatalf("seed ban: %v", err)
	}
	if removed, err := db.UnbanUser(ctx, "u1"); err != nil || !removed {
		t.Errorf("unban existing: got removed=%v (err=%v), want true", removed, err)
	}

	// And the ban check now returns false for that user.
	if banned, _ := db.IsUserBanned(ctx, "u1"); banned {
		t.Error("user still appears banned after unban")
	}
}

// TestUnbanPreservesUserData locks in the "ban/unban cycle doesn't wipe
// stored data" contract — after unban, the user's profile rows, ratings,
// and command-usage counts are intact so they resume where they left off.
func TestUnbanPreservesUserData(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	u, err := db.EnsureUser(ctx, "g", "u1")
	if err != nil {
		t.Fatalf("ensure user: %v", err)
	}
	if err := db.IncrementUserCommandUsage(ctx, "g", "u1", "image filter blur"); err != nil {
		t.Fatalf("seed usage: %v", err)
	}

	if _, err := db.BanUser(ctx, "u1", "abuse", "admin1"); err != nil {
		t.Fatalf("ban: %v", err)
	}
	if _, err := db.UnbanUser(ctx, "u1"); err != nil {
		t.Fatalf("unban: %v", err)
	}

	// User row and usage row both still present.
	got, err := db.GetUserByDiscordID(ctx, "u1", u.GuildID)
	if err != nil || got == nil {
		t.Fatalf("user row gone after ban/unban: got=%v err=%v", got, err)
	}
	total, err := db.GetUserCommandTotal(ctx, u.ID)
	if err != nil {
		t.Fatalf("user command total: %v", err)
	}
	if total != 1 {
		t.Errorf("usage count lost: got %d, want 1", total)
	}
}

// TestListBannedUsersInGuild covers the per-guild filter: only banned users
// who have a User row in this guild appear; bans from other guilds (or
// pre-emptive bans with no User row anywhere) are filtered out.
func TestListBannedUsersInGuild(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Three users, three different setups:
	//   uA — in guild g1, banned (should appear in g1's list)
	//   uB — in guild g2, banned (should NOT appear in g1's list)
	//   uC — never used the bot, pre-emptively banned (should NOT appear)
	if _, err := db.EnsureUser(ctx, "g1", "uA"); err != nil {
		t.Fatalf("ensure uA: %v", err)
	}
	if _, err := db.EnsureUser(ctx, "g2", "uB"); err != nil {
		t.Fatalf("ensure uB: %v", err)
	}
	for _, id := range []string{"uA", "uB", "uC"} {
		if _, err := db.BanUser(ctx, id, "spam", "admin1"); err != nil {
			t.Fatalf("ban %s: %v", id, err)
		}
	}

	got, err := db.ListBannedUsersInGuild(ctx, "g1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("g1 list should return exactly uA, got %d entries", len(got))
	}
	if got[0].DiscordUserID != "uA" {
		t.Errorf("g1 list returned %q, want uA", got[0].DiscordUserID)
	}

	// g2's list returns uB.
	got2, _ := db.ListBannedUsersInGuild(ctx, "g2")
	if len(got2) != 1 || got2[0].DiscordUserID != "uB" {
		t.Errorf("g2 list = %+v, want [uB]", got2)
	}

	// Unknown guild → empty slice, no error.
	got3, err := db.ListBannedUsersInGuild(ctx, "unknown")
	if err != nil || len(got3) != 0 {
		t.Errorf("unknown guild: got %d entries (err=%v), want 0", len(got3), err)
	}
}
