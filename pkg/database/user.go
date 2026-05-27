package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// User is a (Discord user, guild) pair. The same Discord user appears as
// multiple rows if they're tracked in multiple guilds.
type User struct {
	ID            int64  `db:"ID"`
	DiscordUserID string `db:"Discord_UserID"`
	GuildID       int64  `db:"GuildID"`
	Dosh          int64  `db:"Dosh"`
	IsDayOne      bool   `db:"IsDayOne"`
	CreatedAt     string `db:"CreatedAt"`
}

// GetUserByDiscordID returns the row for this Discord user within the given
// guild context, or nil when no match exists.
func (db *DB) GetUserByDiscordID(ctx context.Context, discordUserID string, guildID int64) (*User, error) {
	var u User
	err := db.GetContext(
		ctx, &u,
		`SELECT * FROM User WHERE Discord_UserID = ? AND GuildID = ?`,
		discordUserID, guildID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("user by discord id %s in guild %d: %w", discordUserID, guildID, err)
	}

	return &u, nil
}

// CreateUser inserts a (Discord user, guild) row. The guild must already
// exist — the FK constraint is enforced.
func (db *DB) CreateUser(ctx context.Context, discordUserID string, guildID int64) (*User, error) {
	var u User
	err := db.GetContext(
		ctx, &u,
		`INSERT INTO User (Discord_UserID, GuildID) VALUES (?, ?) RETURNING *`,
		discordUserID, guildID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user %s in guild %d: %w", discordUserID, guildID, err)
	}
	return &u, nil
}

// EnsureUser guarantees a (guild, user) row exists and returns it. It creates
// the Guild row first when needed (the FK requires it), defaulting new guilds
// to audio-disabled. Both inserts are idempotent, so repeat calls are cheap
// no-ops that never overwrite existing Dosh / IsDayOne values.
func (db *DB) EnsureUser(ctx context.Context, discordGuildID, discordUserID string) (*User, error) {
	if err := db.EnsureGuildExists(ctx, discordGuildID, false); err != nil {
		return nil, fmt.Errorf("ensure guild for user: %w", err)
	}
	guild, err := db.GuildByDiscordID(ctx, discordGuildID)
	if err != nil {
		return nil, err
	}
	if guild == nil {
		return nil, fmt.Errorf("guild %s missing immediately after ensure", discordGuildID)
	}

	if _, err := db.ExecContext(ctx,
		`INSERT OR IGNORE INTO User (Discord_UserID, GuildID) VALUES (?, ?)`,
		discordUserID, guild.ID,
	); err != nil {
		return nil, fmt.Errorf("ensure user %s in guild %d: %w", discordUserID, guild.ID, err)
	}

	return db.GetUserByDiscordID(ctx, discordUserID, guild.ID)
}

// ForgetUser hard-deletes every row for this Discord user across all guilds
// — the privacy "right to be forgotten" path. Returns the number of rows
// removed (0 if the user had no stored data).
func (db *DB) ForgetUser(ctx context.Context, discordUserID string) (int64, error) {
	res, err := db.ExecContext(ctx,
		`DELETE FROM User WHERE Discord_UserID = ?`,
		discordUserID,
	)
	if err != nil {
		return 0, fmt.Errorf("forget user %s: %w", discordUserID, err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}
