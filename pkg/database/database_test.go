package database

import (
	"context"
	"errors"
	"testing"
)

// newTestDB returns a fresh in-memory DB with all migrations applied.
// Cleanup is registered on the test so callers don't have to.
func newTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func TestOpenAppliesMigrations(t *testing.T) {
	db := newTestDB(t)

	// If the migration ran, both tables exist and a SELECT against them succeeds.
	for _, table := range []string{"Guild", "User"} {
		if _, err := db.Exec("SELECT 1 FROM " + table + " WHERE 1=0"); err != nil {
			t.Errorf("table %s missing after migrations: %v", table, err)
		}
	}
}

func TestCreateAndFetchGuild(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	created, err := db.CreateGuild(ctx, "123456789")
	if err != nil {
		t.Fatalf("create guild: %v", err)
	}
	if created.ID == 0 {
		t.Errorf("expected non-zero ID, got 0")
	}
	if created.DiscordGuildID != "123456789" {
		t.Errorf("expected Discord_GuildID=123456789, got %s", created.DiscordGuildID)
	}
	if created.AudioEnabled {
		t.Errorf("expected AudioEnabled=false by default, got true")
	}

	fetched, err := db.GuildByDiscordID(ctx, "123456789")
	if err != nil {
		t.Fatalf("fetch guild: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected guild, got nil")
	}
	if fetched.ID != created.ID {
		t.Errorf("expected ID=%d, got %d", created.ID, fetched.ID)
	}
}

func TestGuildByDiscordIDReturnsNilOnMiss(t *testing.T) {
	db := newTestDB(t)

	g, err := db.GuildByDiscordID(context.Background(), "doesnt-exist")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if g != nil {
		t.Errorf("expected nil, got %+v", g)
	}
}

func TestDuplicateGuildRejected(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if _, err := db.CreateGuild(ctx, "abc"); err != nil {
		t.Fatalf("first insert: %v", err)
	}
	if _, err := db.CreateGuild(ctx, "abc"); err == nil {
		t.Errorf("expected UNIQUE constraint violation on duplicate Discord_GuildID")
	}
}

func TestCreateAndFetchUser(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	guild, err := db.CreateGuild(ctx, "guild-1")
	if err != nil {
		t.Fatalf("create guild: %v", err)
	}

	user, err := db.CreateUser(ctx, "user-1", guild.ID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if user.GuildID != guild.ID {
		t.Errorf("expected GuildID=%d, got %d", guild.ID, user.GuildID)
	}

	fetched, err := db.GetUserByDiscordID(ctx, "user-1", guild.ID)
	if err != nil {
		t.Fatalf("fetch user: %v", err)
	}
	if fetched == nil || fetched.ID != user.ID {
		t.Errorf("expected user ID=%d, got %+v", user.ID, fetched)
	}
}

func TestEnsureGuildExistsInsertsThenSkips(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if err := db.EnsureGuildExists(ctx, "seed-1", true); err != nil {
		t.Fatalf("first ensure: %v", err)
	}
	g, _ := db.GuildByDiscordID(ctx, "seed-1")
	if g == nil || !g.AudioEnabled {
		t.Fatalf("expected row with AudioEnabled=true, got %+v", g)
	}
	originalID := g.ID

	// Second call with audioEnabledOnInsert=false should NOT overwrite — the
	// row already exists, so the new value is ignored.
	if err := db.EnsureGuildExists(ctx, "seed-1", false); err != nil {
		t.Fatalf("second ensure: %v", err)
	}
	g2, _ := db.GuildByDiscordID(ctx, "seed-1")
	if g2.ID != originalID {
		t.Errorf("ID changed: was %d, now %d", originalID, g2.ID)
	}
	if !g2.AudioEnabled {
		t.Errorf("AudioEnabled was clobbered; expected true, got false")
	}
}

func TestMarkGuildJoinedCreatesAndClearsLeftAt(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// First call creates the row, LeftAt NULL.
	if err := db.MarkGuildJoined(ctx, "g"); err != nil {
		t.Fatalf("first join: %v", err)
	}
	g, _ := db.GuildByDiscordID(ctx, "g")
	if g == nil {
		t.Fatal("expected row created")
	}
	if g.LeftAt.Valid {
		t.Errorf("expected LeftAt NULL on fresh join, got %q", g.LeftAt.String)
	}

	// Simulate a departure, then rejoin — LeftAt should clear.
	if err := db.MarkGuildLeft(ctx, "g"); err != nil {
		t.Fatalf("leave: %v", err)
	}
	left, _ := db.GuildByDiscordID(ctx, "g")
	if !left.LeftAt.Valid {
		t.Errorf("expected LeftAt set after leave")
	}
	if err := db.MarkGuildJoined(ctx, "g"); err != nil {
		t.Fatalf("rejoin: %v", err)
	}
	rejoined, _ := db.GuildByDiscordID(ctx, "g")
	if rejoined.LeftAt.Valid {
		t.Errorf("expected LeftAt cleared on rejoin, got %q", rejoined.LeftAt.String)
	}
}

func TestMarkGuildLeftPreservesUsers(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	g, _ := db.CreateGuild(ctx, "g")
	if _, err := db.CreateUser(ctx, "u", g.ID); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	if err := db.MarkGuildLeft(ctx, "g"); err != nil {
		t.Fatalf("mark left: %v", err)
	}

	// Marking left must NOT cascade-delete the user (unlike a hard delete).
	if u, _ := db.GetUserByDiscordID(ctx, "u", g.ID); u == nil {
		t.Errorf("user was wrongly removed when guild marked left")
	}
}

func TestGuildPrefixDefaultOverrideReset(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// No row → default.
	if p, err := db.GetGuildPrefixOverride(ctx, "noguild"); err != nil || p != "$" {
		t.Fatalf("missing guild: expected $/nil, got %q/%v", p, err)
	}

	// Row with NULL prefix → default.
	if _, err := db.CreateGuild(ctx, "g"); err != nil {
		t.Fatalf("create guild: %v", err)
	}
	if p, err := db.GetGuildPrefixOverride(ctx, "g"); err != nil || p != "$" {
		t.Errorf("NULL prefix: expected $/nil, got %q/%v", p, err)
	}

	// Set an override.
	if err := db.SetGuildPrefixOverride(ctx, "g", "!"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if p, _ := db.GetGuildPrefixOverride(ctx, "g"); p != "!" {
		t.Errorf("after set: expected !, got %q", p)
	}

	// Reset (empty) → back to default.
	if err := db.SetGuildPrefixOverride(ctx, "g", ""); err != nil {
		t.Fatalf("reset: %v", err)
	}
	if p, _ := db.GetGuildPrefixOverride(ctx, "g"); p != "$" {
		t.Errorf("after reset: expected $, got %q", p)
	}
}

func TestGuildPrefixCacheReflectsSet(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	if _, err := db.CreateGuild(ctx, "g"); err != nil {
		t.Fatalf("create guild: %v", err)
	}

	// Prime the cache with the default, then change via the setter.
	if p, _ := db.GetGuildPrefixOverride(ctx, "g"); p != "$" {
		t.Fatalf("prime: expected $, got %q", p)
	}
	if err := db.SetGuildPrefixOverride(ctx, "g", "?"); err != nil {
		t.Fatalf("set: %v", err)
	}
	// The read should reflect the new value (cache kept in sync by the setter).
	if p, _ := db.GetGuildPrefixOverride(ctx, "g"); p != "?" {
		t.Errorf("cache not updated after set: got %q", p)
	}
}

func TestCachedGuildPrefix(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()
	if _, err := db.CreateGuild(ctx, "g"); err != nil {
		t.Fatalf("create guild: %v", err)
	}

	// Nothing loaded yet → miss (no DB touched).
	if _, ok := db.CachedGuildPrefix("g"); ok {
		t.Errorf("expected cache miss before any load")
	}

	// A full lookup populates the cache.
	if _, err := db.GetGuildPrefixOverride(ctx, "g"); err != nil {
		t.Fatalf("load: %v", err)
	}
	if p, ok := db.CachedGuildPrefix("g"); !ok || p != "$" {
		t.Errorf("expected cache hit with default, got %q/%v", p, ok)
	}
}

func TestGuildAudioEnabledMissingRow(t *testing.T) {
	db := newTestDB(t)

	enabled, err := db.IsGuildAudioEnabled(context.Background(), "never-seen")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if enabled {
		t.Errorf("expected false for missing row, got true")
	}
}

func TestUserDefaultsForOptionalColumns(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	g, _ := db.CreateGuild(ctx, "g")
	u, err := db.CreateUser(ctx, "u", g.ID)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	if u.Dosh != 0 {
		t.Errorf("expected Dosh default 0, got %d", u.Dosh)
	}
	if u.IsDayOne {
		t.Errorf("expected IsDayOne default false, got true")
	}
}

func TestUserCanExistInMultipleGuilds(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	g1, _ := db.CreateGuild(ctx, "g1")
	g2, _ := db.CreateGuild(ctx, "g2")

	if _, err := db.CreateUser(ctx, "same-user", g1.ID); err != nil {
		t.Fatalf("insert in g1: %v", err)
	}
	if _, err := db.CreateUser(ctx, "same-user", g2.ID); err != nil {
		t.Fatalf("insert in g2: %v", err)
	}

	// Same (user, guild) pair should fail the composite UNIQUE.
	if _, err := db.CreateUser(ctx, "same-user", g1.ID); err == nil {
		t.Errorf("expected UNIQUE(Discord_UserID, GuildID) violation")
	}
}

func TestUserGuildForeignKeyCascade(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	guild, _ := db.CreateGuild(ctx, "to-be-deleted")
	if _, err := db.CreateUser(ctx, "user", guild.ID); err != nil {
		t.Fatalf("create user: %v", err)
	}

	// Deleting the guild should cascade to its User rows.
	if _, err := db.Exec(`DELETE FROM Guild WHERE ID = ?`, guild.ID); err != nil {
		t.Fatalf("delete guild: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM User WHERE GuildID = ?`, guild.ID).Scan(&count); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if count != 0 {
		t.Errorf("expected cascade delete; %d user rows remain", count)
	}
}

func TestUserCreateFailsWithoutGuild(t *testing.T) {
	db := newTestDB(t)

	// GuildID=999 doesn't exist; FK constraint should reject.
	if _, err := db.CreateUser(context.Background(), "orphan", 999); err == nil {
		t.Errorf("expected FK violation for non-existent guild")
	}
}

func TestEnsureUserCreatesGuildAndUser(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Neither the guild nor the user exists yet.
	u, err := db.EnsureUser(ctx, "fresh-guild", "fresh-user")
	if err != nil {
		t.Fatalf("ensure user: %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.Dosh != 0 || u.IsDayOne {
		t.Errorf("expected defaults (Dosh=0, IsDayOne=false), got %+v", u)
	}

	// The guild should have been auto-created, audio-disabled.
	g, _ := db.GuildByDiscordID(ctx, "fresh-guild")
	if g == nil {
		t.Fatal("expected guild auto-created by EnsureUser")
	}
	if g.AudioEnabled {
		t.Errorf("expected new guild audio-disabled by default")
	}
}

func TestEnsureUserIsIdempotent(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	u1, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("first ensure: %v", err)
	}
	u2, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("second ensure: %v", err)
	}
	if u1.ID != u2.ID {
		t.Errorf("expected same row both times, got %d then %d", u1.ID, u2.ID)
	}
}

func TestForgetUserDeletesAllGuilds(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Same Discord user across two guilds, plus an unrelated user.
	g1, _ := db.CreateGuild(ctx, "g1")
	g2, _ := db.CreateGuild(ctx, "g2")
	if _, err := db.CreateUser(ctx, "victim", g1.ID); err != nil {
		t.Fatalf("seed victim g1: %v", err)
	}
	if _, err := db.CreateUser(ctx, "victim", g2.ID); err != nil {
		t.Fatalf("seed victim g2: %v", err)
	}
	if _, err := db.CreateUser(ctx, "bystander", g1.ID); err != nil {
		t.Fatalf("seed bystander: %v", err)
	}

	n, err := db.ForgetUser(ctx, "victim")
	if err != nil {
		t.Fatalf("forget user: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 rows deleted (both guilds), got %d", n)
	}

	// Victim gone from both guilds.
	if u, _ := db.GetUserByDiscordID(ctx, "victim", g1.ID); u != nil {
		t.Errorf("victim still present in g1")
	}
	if u, _ := db.GetUserByDiscordID(ctx, "victim", g2.ID); u != nil {
		t.Errorf("victim still present in g2")
	}
	// Bystander untouched.
	if u, _ := db.GetUserByDiscordID(ctx, "bystander", g1.ID); u == nil {
		t.Errorf("bystander was wrongly deleted")
	}
}

func TestIncrementCommandUsage(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Unseen command starts at 0.
	if c, _ := db.CommandUsageCount(ctx, "image"); c != 0 {
		t.Errorf("expected 0 for unseen command, got %d", c)
	}

	// Three increments → 3; an unrelated command stays independent.
	for range 3 {
		if err := db.IncrementCommandUsage(ctx, "image"); err != nil {
			t.Fatalf("increment: %v", err)
		}
	}
	if err := db.IncrementCommandUsage(ctx, "audio"); err != nil {
		t.Fatalf("increment audio: %v", err)
	}

	if c, _ := db.CommandUsageCount(ctx, "image"); c != 3 {
		t.Errorf("expected image=3, got %d", c)
	}
	if c, _ := db.CommandUsageCount(ctx, "audio"); c != 1 {
		t.Errorf("expected audio=1, got %d", c)
	}
}

func TestForgetUserNoRows(t *testing.T) {
	db := newTestDB(t)

	n, err := db.ForgetUser(context.Background(), "never-existed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 rows deleted, got %d", n)
	}
}

func TestEnsureUserDoesNotClobberData(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	u, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	// Simulate accumulated economy state.
	if _, err := db.ExecContext(ctx, `UPDATE User SET Dosh = 500, IsDayOne = 1 WHERE ID = ?`, u.ID); err != nil {
		t.Fatalf("seed data: %v", err)
	}

	// Re-ensuring must not reset Dosh / IsDayOne.
	u2, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("re-ensure: %v", err)
	}
	if u2.Dosh != 500 {
		t.Errorf("EnsureUser clobbered Dosh: expected 500, got %d", u2.Dosh)
	}
	if !u2.IsDayOne {
		t.Errorf("EnsureUser clobbered IsDayOne: expected true")
	}
}

func TestGetApiURLReturnsSeededValue(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// The migration seeds rows for known names with placeholder URLs.
	// We just verify lookup works end-to-end; the placeholder is fine here.
	url, err := db.GetApiURL(ctx, "xkcd")
	if err != nil {
		t.Fatalf("expected seeded xkcd row, got error: %v", err)
	}
	if url == "" {
		t.Errorf("expected non-empty URL")
	}
}

func TestGetApiURLMissingNameReturnsNotAvailable(t *testing.T) {
	db := newTestDB(t)

	_, err := db.GetApiURL(context.Background(), "no-such-api")
	if !errors.Is(err, ErrApiURLNotAvailable) {
		t.Errorf("expected ErrApiURLNotAvailable, got %v", err)
	}
}

func TestGetApiURLInactiveReturnsNotAvailable(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// Flip xkcd to inactive directly via SQL — that's the path admins will
	// use in prod, so it's worth covering here.
	if _, err := db.ExecContext(ctx, `UPDATE ApiURL SET IsActive = 0 WHERE ApiName = ?`, "xkcd"); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	_, err := db.GetApiURL(ctx, "xkcd")
	if !errors.Is(err, ErrApiURLNotAvailable) {
		t.Errorf("expected ErrApiURLNotAvailable for inactive row, got %v", err)
	}
}

func TestGetApiURLRowReturnsInactiveRows(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	if _, err := db.ExecContext(ctx, `UPDATE ApiURL SET IsActive = 0 WHERE ApiName = ?`, "xkcd"); err != nil {
		t.Fatalf("deactivate: %v", err)
	}

	row, err := db.GetApiURLRow(ctx, "xkcd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if row == nil {
		t.Fatal("expected row, got nil")
	}
	if row.IsActive {
		t.Errorf("expected IsActive=false after manual update")
	}
}
