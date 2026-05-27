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
	LeftAt                     sql.NullString `db:"LeftAt"` // NULL while the bot is in the guild
}

// GuildByDiscordID returns nil (without error) when no row matches.
func (db *DB) GuildByDiscordID(ctx context.Context, discordGuildID string) (*Guild, error) {
	var g Guild
	err := db.GetContext(
		ctx, &g,
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
	err := db.GetContext(
		ctx, &g,
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
	_, err := db.ExecContext(
		ctx,
		`INSERT OR IGNORE INTO Guild (Discord_GuildID, AudioEnabled) VALUES (?, ?)`,
		discordGuildID, audioEnabledOnInsert,
	)
	if err != nil {
		return fmt.Errorf("ensure guild %s: %w", discordGuildID, err)
	}
	return nil
}

// MarkGuildJoined records that the bot is present in the guild: it creates
// the row if missing and clears LeftAt either way. Called from GuildCreate,
// which fires for every guild on connect (backfill) and on each new join.
func (db *DB) MarkGuildJoined(ctx context.Context, discordGuildID string) error {
	_, err := db.ExecContext(
		ctx, `
		INSERT INTO Guild (Discord_GuildID) VALUES (?)
		ON CONFLICT(Discord_GuildID) DO UPDATE SET LeftAt = NULL
	`, discordGuildID,
	)
	if err != nil {
		return fmt.Errorf("mark guild joined %s: %w", discordGuildID, err)
	}
	return nil
}

// MarkGuildLeft stamps LeftAt without deleting anything, so per-guild data
// survives a removal and is restored if the bot rejoins. No-op if the guild
// has no row.
func (db *DB) MarkGuildLeft(ctx context.Context, discordGuildID string) error {
	_, err := db.ExecContext(
		ctx,
		`UPDATE Guild SET LeftAt = CURRENT_TIMESTAMP WHERE Discord_GuildID = ?`,
		discordGuildID,
	)
	if err != nil {
		return fmt.Errorf("mark guild left %s: %w", discordGuildID, err)
	}
	return nil
}

// IsGuildAudioEnabled returns the AudioEnabled flag for the row matching
// discordGuildID. A missing row returns (false, nil) — unknown guilds
// default to disabled.
func (db *DB) IsGuildAudioEnabled(ctx context.Context, discordGuildID string) (bool, error) {
	var enabled bool
	err := db.GetContext(
		ctx, &enabled,
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

// defaultPrefix is used when a guild has no row or no override set.
const defaultPrefix = "$"

// CachedGuildPrefix returns the in-memory cached prefix for a guild, or
// ("", false) on a miss. Pure map read — no context, no DB. The per-message
// fast path: callers fall back to GetGuildPrefixOverride only on a miss.
func (db *DB) CachedGuildPrefix(discordGuildID string) (string, bool) {
	return db.prefixCache.get(discordGuildID)
}

// GetGuildPrefixOverride returns the guild's command prefix: the per-guild
// override when set, otherwise defaultPrefix. Memoized — the first lookup per
// guild hits the DB, later ones read the cache (this runs on every message).
// Any DB problem degrades to the default, returned alongside the error so the
// caller keeps a usable value while still able to log.
func (db *DB) GetGuildPrefixOverride(ctx context.Context, discordGuildID string) (string, error) {
	if p, ok := db.prefixCache.get(discordGuildID); ok {
		return p, nil
	}

	var override sql.NullString
	err := db.GetContext(ctx, &override,
		`SELECT PrefixOverride FROM Guild WHERE Discord_GuildID = ?`,
		discordGuildID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		// No row yet — don't cache; the guild lifecycle will add one shortly.
		return defaultPrefix, nil
	}
	if err != nil {
		return defaultPrefix, fmt.Errorf("guild prefix override %s: %w", discordGuildID, err)
	}

	prefix := defaultPrefix
	if override.Valid && override.String != "" {
		prefix = override.String
	}
	db.prefixCache.set(discordGuildID, prefix)
	return prefix, nil
}

// SetGuildPrefixOverride sets the guild's prefix, or resets to the default
// when prefix is empty (stored as NULL). Upserts so a missing row is created
// rather than silently skipped, and updates the cache so the change is live
// immediately.
func (db *DB) SetGuildPrefixOverride(ctx context.Context, discordGuildID, prefix string) error {
	var stored any // NULL resets to the default
	if prefix != "" {
		stored = prefix
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO Guild (Discord_GuildID, PrefixOverride) VALUES (?, ?)
		ON CONFLICT(Discord_GuildID) DO UPDATE SET PrefixOverride = excluded.PrefixOverride
	`, discordGuildID, stored)
	if err != nil {
		return fmt.Errorf("set guild prefix override %s: %w", discordGuildID, err)
	}

	resolved := defaultPrefix
	if prefix != "" {
		resolved = prefix
	}
	db.prefixCache.set(discordGuildID, resolved)
	return nil
}
