package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// BannedUser is one row in the BannedUser table. Discord_UserID is UNIQUE so
// a user has at most one ban regardless of guild count — bans are global.
// Reason and BannedBy are nullable for flexibility (pre-emptive bans don't
// need a reason; bans issued via direct DB edit may have no recorded issuer).
type BannedUser struct {
	ID            int64          `db:"ID"`
	DiscordUserID string         `db:"Discord_UserID"`
	Reason        sql.NullString `db:"Reason"`
	BannedAt      string         `db:"BannedAt"`
	BannedBy      sql.NullString `db:"BannedBy"`
}

// IsUserBanned reports whether the given Discord user ID has an active ban.
// Runs in the dispatch hot path (every slash + prefix invocation), so keep
// the query single-row and indexed (Discord_UserID is UNIQUE → automatic
// index). Empty input short-circuits to false without a DB round trip.
func (db *DB) IsUserBanned(ctx context.Context, discordUserID string) (bool, error) {
	if discordUserID == "" {
		return false, nil
	}
	var one int
	err := db.GetContext(ctx, &one,
		`SELECT 1 FROM BannedUser WHERE Discord_UserID = ? LIMIT 1`,
		discordUserID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("is user banned %s: %w", discordUserID, err)
	}
	return true, nil
}

// BanUser inserts or updates a ban row. Returns (true, nil) when a new ban
// was created, (false, nil) when an existing ban's metadata was refreshed.
// Idempotent — issuing a ban twice updates the reason/issuer/timestamp on
// the second call rather than erroring.
func (db *DB) BanUser(ctx context.Context, discordUserID, reason, bannedBy string) (created bool, err error) {
	if discordUserID == "" {
		return false, errors.New("ban user: discordUserID is required")
	}

	// Empty string → SQL NULL for nullable columns. Lets DBeaver tell
	// "no reason given" apart from "reason was literally empty".
	var reasonArg, bannedByArg any
	if reason != "" {
		reasonArg = reason
	}
	if bannedBy != "" {
		bannedByArg = bannedBy
	}

	// Read the prior state so we can report new-vs-update to the caller.
	prior, err := db.GetBannedUser(ctx, discordUserID)
	if err != nil {
		return false, err
	}

	_, err = db.ExecContext(ctx, `
		INSERT INTO BannedUser (Discord_UserID, Reason, BannedBy)
		VALUES (?, ?, ?)
		ON CONFLICT(Discord_UserID) DO UPDATE SET
			Reason = excluded.Reason,
			BannedBy = excluded.BannedBy,
			BannedAt = CURRENT_TIMESTAMP
	`, discordUserID, reasonArg, bannedByArg)
	if err != nil {
		return false, fmt.Errorf("ban user %s: %w", discordUserID, err)
	}
	return prior == nil, nil
}

// UnbanUser deletes the ban row, if any. Returns true when a row was deleted
// (the user was previously banned), false when no row existed. Does NOT
// touch the User table — the user's profile, ratings, and command history
// survive the ban/unban cycle so they pick up where they left off.
func (db *DB) UnbanUser(ctx context.Context, discordUserID string) (removed bool, err error) {
	if discordUserID == "" {
		return false, errors.New("unban user: discordUserID is required")
	}
	res, err := db.ExecContext(ctx,
		`DELETE FROM BannedUser WHERE Discord_UserID = ?`,
		discordUserID,
	)
	if err != nil {
		return false, fmt.Errorf("unban user %s: %w", discordUserID, err)
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// GetBannedUser returns the ban row for the given user, or (nil, nil) if no
// ban is recorded. The (nil, nil) case is the common one and is NOT an error.
func (db *DB) GetBannedUser(ctx context.Context, discordUserID string) (*BannedUser, error) {
	var b BannedUser
	err := db.GetContext(ctx, &b,
		`SELECT * FROM BannedUser WHERE Discord_UserID = ?`,
		discordUserID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get banned user %s: %w", discordUserID, err)
	}
	return &b, nil
}

// ListBannedUsersInGuild returns banned users who have a User row tied to the
// given guild — i.e., users who have invoked the bot in this guild at least
// once. Used by /admin banned-list so server admins see only bans relevant
// to their server, not the global list. Ordered newest-first.
//
// Trade-off: a user banned pre-emptively who has never used the bot in this
// guild won't appear here. That's the right behavior — they're not present
// in the guild's bot-user pool. The maintainer can still see the global
// list via DBeaver / direct DB query if they need it.
func (db *DB) ListBannedUsersInGuild(ctx context.Context, discordGuildID string) ([]*BannedUser, error) {
	rows := []*BannedUser{}
	err := db.SelectContext(ctx, &rows, `
		SELECT b.*
		FROM BannedUser b
		JOIN User u ON u.Discord_UserID = b.Discord_UserID
		JOIN Guild g ON u.GuildID = g.ID
		WHERE g.Discord_GuildID = ?
		ORDER BY b.BannedAt DESC
	`, discordGuildID)
	if err != nil {
		return nil, fmt.Errorf("list banned users in guild %s: %w", discordGuildID, err)
	}
	return rows, nil
}
