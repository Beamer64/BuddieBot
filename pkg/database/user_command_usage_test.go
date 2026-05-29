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

// TestIncrementUserCommandUsageNoOpForUnknownUser confirms the SELECT-driven
// INSERT silently does nothing when the (user, guild) pair has no User row.
// This is the privacy-stance contract: tracking never creates rows.
func TestIncrementUserCommandUsageNoOpForUnknownUser(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	// No EnsureUser. Guild also doesn't exist. The statement should still
	// succeed (no SQL error), but write zero rows.
	if err := db.IncrementUserCommandUsage(ctx, "ghost-guild", "ghost-user", "image"); err != nil {
		t.Fatalf("increment for missing user shouldn't error: %v", err)
	}

	// And no row materialized in User table either.
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
