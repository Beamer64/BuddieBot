package database

import (
	"context"
	"testing"

	"github.com/Beamer64/BuddieBot/pkg/helper"
)

// TestSetUserRatingUpserts confirms repeated calls overwrite the value in
// place (one row per user/name) and that subsequent reads see the latest
// value. Timestamp advancement is exercised by TestGetRecentUserRatingsOrder.
func TestSetUserRatingUpserts(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	u, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("ensure user: %v", err)
	}

	if err := db.SetUserRating(ctx, u.ID, "nerd", 10); err != nil {
		t.Fatalf("first set: %v", err)
	}
	if err := db.SetUserRating(ctx, u.ID, "nerd", 88); err != nil {
		t.Fatalf("overwrite set: %v", err)
	}

	// Read it back via the "all ratings" path — exactly one nerd row, value 88.
	all, err := db.GetUserRatings(ctx, u.ID)
	if err != nil {
		t.Fatalf("list ratings: %v", err)
	}
	hits := 0
	for _, r := range all {
		if r.RatingName == "nerd" {
			hits++
			if r.Value != 88 {
				t.Errorf("expected nerd=88 after upsert, got %d", r.Value)
			}
		}
	}
	if hits != 1 {
		t.Errorf("expected exactly one 'nerd' row, got %d", hits)
	}
}

// TestGetRecentUserRatingsOrder confirms freshest-first ordering. SQLite's
// CURRENT_TIMESTAMP is second-precision, so we insert with explicit UpdatedAt
// values to avoid burning real seconds on a sleep.
func TestGetRecentUserRatingsOrder(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	u, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("ensure user: %v", err)
	}

	// Drop whatever seeding produced so the timestamps below dictate order.
	if _, err := db.ExecContext(ctx, `DELETE FROM UserRating WHERE UserID = ?`, u.ID); err != nil {
		t.Fatalf("clear seeded ratings: %v", err)
	}

	inserts := []struct {
		name string
		at   string
	}{
		{"nerd", "2024-01-01 00:00:01"},
		{"stinky", "2024-01-01 00:00:02"},
		{"schmeat", "2024-01-01 00:00:03"},
		{"geek", "2024-01-01 00:00:04"},
	}
	for _, ins := range inserts {
		if _, err := db.ExecContext(ctx,
			`INSERT INTO UserRating (UserID, RatingName, Value, UpdatedAt) VALUES (?, ?, ?, ?)`,
			u.ID, ins.name, 1, ins.at,
		); err != nil {
			t.Fatalf("insert %s: %v", ins.name, err)
		}
	}

	recent, err := db.GetRecentUserRatings(ctx, u.ID, 3)
	if err != nil {
		t.Fatalf("recent: %v", err)
	}
	if len(recent) != 3 {
		t.Fatalf("expected 3 most-recent, got %d", len(recent))
	}
	want := []string{"geek", "schmeat", "stinky"}
	for idx, w := range want {
		if recent[idx].RatingName != w {
			t.Errorf("recent[%d] = %q, want %q", idx, recent[idx].RatingName, w)
		}
	}
}

// TestEnsureUserSeedsRatingsOnCreate locks in the "fresh user gets N random
// ratings" contract. Subsequent EnsureUser calls must not re-seed (would
// clobber any /rate-this writes that have happened since).
func TestEnsureUserSeedsRatingsOnCreate(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	u, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("first ensure: %v", err)
	}

	got, err := db.GetUserRatings(ctx, u.ID)
	if err != nil {
		t.Fatalf("list seeded: %v", err)
	}
	if len(got) != initialRatingSeedCount {
		t.Fatalf("expected %d seeded ratings, got %d", initialRatingSeedCount, len(got))
	}

	// All seeded names must come from the canonical list, and be distinct.
	known := map[string]bool{}
	for _, n := range helper.RatingNames {
		known[n] = true
	}
	seen := map[string]bool{}
	for _, r := range got {
		if !known[r.RatingName] {
			t.Errorf("unexpected rating name %q (not in helper.RatingNames)", r.RatingName)
		}
		if seen[r.RatingName] {
			t.Errorf("duplicate seeded rating %q", r.RatingName)
		}
		seen[r.RatingName] = true
	}

	// Second EnsureUser must NOT add more rows.
	if _, err := db.EnsureUser(ctx, "g", "u"); err != nil {
		t.Fatalf("second ensure: %v", err)
	}
	got2, _ := db.GetUserRatings(ctx, u.ID)
	if len(got2) != initialRatingSeedCount {
		t.Errorf("expected stable %d ratings after re-ensure, got %d", initialRatingSeedCount, len(got2))
	}
}

// TestUserRatingFKCascadeOnUserDelete confirms /user forget-me's path:
// deleting the User row removes its UserRating rows automatically.
func TestUserRatingFKCascadeOnUserDelete(t *testing.T) {
	db := newTestDB(t)
	ctx := context.Background()

	u, err := db.EnsureUser(ctx, "g", "u")
	if err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if err := db.SetUserRating(ctx, u.ID, "nerd", 50); err != nil {
		t.Fatalf("seed rating: %v", err)
	}

	if _, err := db.ExecContext(ctx, `DELETE FROM User WHERE ID = ?`, u.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM UserRating WHERE UserID = ?`, u.ID).Scan(&count); err != nil {
		t.Fatalf("count ratings: %v", err)
	}
	if count != 0 {
		t.Errorf("expected cascade delete; %d rating rows remain", count)
	}
}
