package database

import (
	"context"
	"fmt"
)

// UserRating is one /rate-this score stored against a (user, guild) pair.
// One row per (UserID, RatingName); new ratings overwrite the previous value.
type UserRating struct {
	ID         int64  `db:"ID"`
	UserID     int64  `db:"UserID"`
	RatingName string `db:"RatingName"`
	Value      int    `db:"Value"`
	UpdatedAt  string `db:"UpdatedAt"`
}

// SetUserRating upserts the latest score for a (user, rating) pair, bumping
// UpdatedAt every time. UNIQUE(UserID, RatingName) means there's only ever one
// "current" value per rating type per user — no history is retained.
func (db *DB) SetUserRating(ctx context.Context, userID int64, name string, value int) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO UserRating (UserID, RatingName, Value)
		 VALUES (?, ?, ?)
		 ON CONFLICT(UserID, RatingName) DO UPDATE
		   SET Value = excluded.Value,
		       UpdatedAt = CURRENT_TIMESTAMP`,
		userID, name, value,
	)
	if err != nil {
		return fmt.Errorf("upsert rating %s for user %d: %w", name, userID, err)
	}
	return nil
}

// GetUserRatings returns every stored rating for the user, sorted by name —
// stable ordering for the future "all ratings" profile page.
func (db *DB) GetUserRatings(ctx context.Context, userID int64) ([]*UserRating, error) {
	var rows []*UserRating
	if err := db.SelectContext(
		ctx, &rows,
		`SELECT * FROM UserRating WHERE UserID = ? ORDER BY RatingName ASC`,
		userID,
	); err != nil {
		return nil, fmt.Errorf("list ratings for user %d: %w", userID, err)
	}
	return rows, nil
}

// GetRecentUserRatings returns the N most recently updated ratings for the
// user, freshest first. Powers the recent-3 corner on /user profile.
func (db *DB) GetRecentUserRatings(ctx context.Context, userID int64, limit int) ([]*UserRating, error) {
	if limit <= 0 {
		return nil, nil
	}
	var rows []*UserRating
	if err := db.SelectContext(
		ctx, &rows,
		`SELECT * FROM UserRating WHERE UserID = ? ORDER BY UpdatedAt DESC, ID DESC LIMIT ?`,
		userID, limit,
	); err != nil {
		return nil, fmt.Errorf("recent ratings for user %d: %w", userID, err)
	}
	return rows, nil
}
