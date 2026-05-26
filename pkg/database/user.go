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
