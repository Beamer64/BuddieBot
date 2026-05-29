package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"

	"github.com/Beamer64/BuddieBot/pkg/helper"
)

// initialRatingSeedCount is how many random rating types get assigned to a
// brand-new user row, so a fresh profile isn't empty.
const initialRatingSeedCount = 3

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

	res, err := db.ExecContext(ctx,
		`INSERT OR IGNORE INTO User (Discord_UserID, GuildID) VALUES (?, ?)`,
		discordUserID, guild.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("ensure user %s in guild %d: %w", discordUserID, guild.ID, err)
	}
	// RowsAffected == 1 means the OR-IGNORE actually inserted; on a race
	// between concurrent EnsureUser calls only the winner sees 1, so seeding
	// runs exactly once per real creation.
	created, _ := res.RowsAffected()

	user, err := db.GetUserByDiscordID(ctx, discordUserID, guild.ID)
	if err != nil || user == nil {
		return user, err
	}

	if created == 1 {
		if seedErr := db.seedInitialRatings(ctx, user.ID); seedErr != nil {
			// Seeding is "nice to have" — don't fail user creation over it. A
			// future /rate-this against this user will start populating ratings.
			return user, fmt.Errorf("seed initial ratings for user %d: %w", user.ID, seedErr)
		}
	}
	return user, nil
}

// seedInitialRatings picks initialRatingSeedCount distinct rating types from
// helper.RatingNames and stores a random value for each, so a freshly-created
// profile has something interesting in the recent-ratings corner instead of
// blank space. Schmeat uses its own ASCII-strip range via helper.RandomRatingValue.
func (db *DB) seedInitialRatings(ctx context.Context, userID int64) error {
	names := make([]string, len(helper.RatingNames))
	copy(names, helper.RatingNames)
	rand.Shuffle(len(names), func(i, j int) { names[i], names[j] = names[j], names[i] })

	n := initialRatingSeedCount
	if n > len(names) {
		n = len(names)
	}
	for _, name := range names[:n] {
		if err := db.SetUserRating(ctx, userID, name, helper.RandomRatingValue(name)); err != nil {
			return err
		}
	}
	return nil
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
