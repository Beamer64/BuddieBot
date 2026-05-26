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

func TestGuildAudioEnabledMissingRow(t *testing.T) {
	db := newTestDB(t)

	enabled, err := db.GuildAudioEnabled(context.Background(), "never-seen")
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
