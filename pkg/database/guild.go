package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Guild mirrors a row in the Guild table. db tags map struct fields to
// column names so sqlx can populate the struct via Get/Select.
type Guild struct {
	ID                         int64          `db:"ID"`
	DiscordGuildID             string         `db:"Discord_GuildID"`
	AudioEnabled               bool           `db:"AudioEnabled"`
	PrefixOverride             sql.NullString `db:"PrefixOverride"`
	DiscordEventNotifChannelID sql.NullString `db:"Discord_EventNotifChannelID"`
	JoinedAt                   string         `db:"JoinedAt"`
}

// GuildByDiscordID returns nil (without error) when no row matches.
func (db *DB) GuildByDiscordID(ctx context.Context, discordGuildID string) (*Guild, error) {
	var g Guild
	err := db.GetContext(ctx, &g,
		`SELECT * FROM Guild WHERE Discord_GuildID = ?`,
		discordGuildID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("guild by discord id %s: %w", discordGuildID, err)
	}
	return &g, nil
}

// GuildByID looks up by our internal surrogate ID. Useful when following a
// foreign key from another row.
func (db *DB) GuildByID(ctx context.Context, id int64) (*Guild, error) {
	var g Guild
	err := db.GetContext(ctx, &g, `SELECT * FROM Guild WHERE ID = ?`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("guild by id %d: %w", id, err)
	}
	return &g, nil
}

// CreateGuild inserts a new row with defaults applied for everything but
// the Discord ID. RETURNING gives us the populated row back in one round trip.
func (db *DB) CreateGuild(ctx context.Context, discordGuildID string) (*Guild, error) {
	var g Guild
	err := db.GetContext(ctx, &g,
		`INSERT INTO Guild (Discord_GuildID) VALUES (?) RETURNING *`,
		discordGuildID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert guild %s: %w", discordGuildID, err)
	}
	return &g, nil
}

// EnsureGuildExists inserts a row only if one isn't already present. Used
// at bot startup to seed known guilds without clobbering admin changes
// from previous runs.
func (db *DB) EnsureGuildExists(ctx context.Context, discordGuildID string, audioEnabledOnInsert bool) error {
	_, err := db.ExecContext(ctx,
		`INSERT OR IGNORE INTO Guild (Discord_GuildID, AudioEnabled) VALUES (?, ?)`,
		discordGuildID, audioEnabledOnInsert,
	)
	if err != nil {
		return fmt.Errorf("ensure guild %s: %w", discordGuildID, err)
	}
	return nil
}

// GuildAudioEnabled returns the AudioEnabled flag for the row matching
// discordGuildID. A missing row returns (false, nil) — unknown guilds
// default to disabled.
func (db *DB) GuildAudioEnabled(ctx context.Context, discordGuildID string) (bool, error) {
	var enabled bool
	err := db.GetContext(ctx, &enabled,
		`SELECT AudioEnabled FROM Guild WHERE Discord_GuildID = ?`,
		discordGuildID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("guild audio enabled %s: %w", discordGuildID, err)
	}
	return enabled, nil
}
