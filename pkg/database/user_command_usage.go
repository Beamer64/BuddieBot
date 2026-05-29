package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// UserCommandUsage is one (user, command) invocation counter. New rows are
// only created for users who already have a User row — tracking never
// materializes users on its own; it just records activity for tracked ones.
type UserCommandUsage struct {
	ID          int64  `db:"ID"`
	UserID      int64  `db:"UserID"`
	CommandName string `db:"CommandName"`
	UsageCount  int64  `db:"UsageCount"`
	LastUsedAt  string `db:"LastUsedAt"`
}

// IncrementUserCommandUsage bumps the per-(user, command) counter — but only
// when the (Discord user, guild) pair already has a User row. The SELECT
// drives the INSERT, so a missing user returns zero source rows and the
// statement is a clean no-op. Preserves the opt-in privacy stance: no rows
// get created just from tracking.
func (db *DB) IncrementUserCommandUsage(ctx context.Context, discordGuildID, discordUserID, commandName string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO UserCommandUsage (UserID, CommandName, UsageCount, LastUsedAt)
		SELECT u.ID, ?, 1, CURRENT_TIMESTAMP
		  FROM User u
		  JOIN Guild g ON u.GuildID = g.ID
		 WHERE u.Discord_UserID = ? AND g.Discord_GuildID = ?
		ON CONFLICT(UserID, CommandName) DO UPDATE
		   SET UsageCount = UsageCount + 1,
		       LastUsedAt = CURRENT_TIMESTAMP
	`, commandName, discordUserID, discordGuildID)
	if err != nil {
		return fmt.Errorf("increment user command usage %s for %s in %s: %w",
			commandName, discordUserID, discordGuildID, err)
	}
	return nil
}

// GetUserCommandTotal returns the sum of UsageCount across every command this
// user has recorded. COALESCE keeps "no rows at all" returning 0 cleanly.
func (db *DB) GetUserCommandTotal(ctx context.Context, userID int64) (int64, error) {
	var total int64
	if err := db.GetContext(ctx, &total,
		`SELECT COALESCE(SUM(UsageCount), 0) FROM UserCommandUsage WHERE UserID = ?`,
		userID,
	); err != nil {
		return 0, fmt.Errorf("user command total for %d: %w", userID, err)
	}
	return total, nil
}

// GetUserTopCommand returns the user's most-used command name + its count.
// When a user has no usage rows yet, returns ("", 0, nil) — caller treats that
// as the "no commands recorded" rendering case. Ties on UsageCount break by
// freshest LastUsedAt so a recent burst surfaces over an old equal-count one.
func (db *DB) GetUserTopCommand(ctx context.Context, userID int64) (string, int64, error) {
	var row struct {
		CommandName string `db:"CommandName"`
		UsageCount  int64  `db:"UsageCount"`
	}
	err := db.GetContext(ctx, &row,
		`SELECT CommandName, UsageCount FROM UserCommandUsage
		 WHERE UserID = ?
		 ORDER BY UsageCount DESC, LastUsedAt DESC
		 LIMIT 1`,
		userID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return "", 0, nil
	}
	if err != nil {
		return "", 0, fmt.Errorf("user top command for %d: %w", userID, err)
	}
	return row.CommandName, row.UsageCount, nil
}
