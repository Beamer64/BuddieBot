package database

import (
	"context"
	"fmt"
)

// CommandUsage mirrors a row in the CommandUsage table — an aggregate
// invocation count per top-level command.
type CommandUsage struct {
	ID          int64  `db:"ID"`
	CommandName string `db:"CommandName"`
	UsageCount  int64  `db:"UsageCount"`
	LastUsedAt  string `db:"LastUsedAt"`
}

// TrackCommandInvocation records one command invocation end-to-end: bumps
// the bot-wide aggregate (CommandUsage), then materializes the (guild, user)
// row and bumps their per-(user, command) counter (UserCommandUsage). Both
// the slash dispatcher (events.CommandHandler) and the prefix dispatcher
// (prefix.ParsePrefixCmds) call this so every invocation lands in both
// counters with one call.
//
// Empty discordGuildID or discordUserID skips the per-user step (DM context
// or bot invoker — the caller is expected to zero discordUserID for bots).
// The aggregate always fires when usageKey is non-empty.
func (db *DB) TrackCommandInvocation(ctx context.Context, usageKey, discordGuildID, discordUserID string) error {
	if usageKey == "" {
		return nil
	}
	if err := db.IncrementCommandUsage(ctx, usageKey); err != nil {
		return err
	}
	if discordGuildID == "" || discordUserID == "" {
		return nil
	}
	return db.RecordUserCommandUsage(ctx, discordGuildID, discordUserID, usageKey)
}

// IncrementCommandUsage bumps the count for commandName, creating the row on
// first use. Atomic upsert — safe to call on every invocation.
func (db *DB) IncrementCommandUsage(ctx context.Context, commandName string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO CommandUsage (CommandName, UsageCount, LastUsedAt)
		VALUES (?, 1, CURRENT_TIMESTAMP)
		ON CONFLICT(CommandName) DO UPDATE SET
			UsageCount = UsageCount + 1,
			LastUsedAt = CURRENT_TIMESTAMP
	`, commandName)
	if err != nil {
		return fmt.Errorf("increment command usage %s: %w", commandName, err)
	}
	return nil
}

// CommandUsageCount returns the recorded count for commandName (0 if unseen).
func (db *DB) CommandUsageCount(ctx context.Context, commandName string) (int64, error) {
	var count int64
	err := db.GetContext(ctx, &count,
		`SELECT COALESCE((SELECT UsageCount FROM CommandUsage WHERE CommandName = ?), 0)`,
		commandName,
	)
	if err != nil {
		return 0, fmt.Errorf("command usage count %s: %w", commandName, err)
	}
	return count, nil
}
