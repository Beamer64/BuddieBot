// Package database owns the SQLite connection and exposes typed CRUD
// for each table. Migrations run automatically on Open.
package database

import (
	"context"
	"fmt"
	"sync"

	"github.com/Beamer64/BuddieBot/pkg/database/migrations"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// DB wraps *sqlx.DB so callers get struct-scanning (Get/Select) for free
// while keeping the embedded *sql.DB methods (Exec, QueryRow, Begin) available.
type DB struct {
	*sqlx.DB
	prefixCache *prefixCache
}

// prefixCache memoizes per-guild command prefixes. ParsePrefixCmds runs on
// every message, so this turns that hot path into a map read instead of a DB
// query. Writes go through SetGuildPrefixOverride, which keeps it in sync.
type prefixCache struct {
	mu sync.RWMutex
	m  map[string]string
}

func newPrefixCache() *prefixCache {
	return &prefixCache{m: make(map[string]string)}
}

func (c *prefixCache) get(guildID string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	p, ok := c.m[guildID]
	return p, ok
}

func (c *prefixCache) set(guildID, prefix string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.m[guildID] = prefix
}

// Open dials the SQLite file at path (":memory:" for tests), applies the
// required pragmas, and runs any pending migrations.
func Open(path string) (*DB, error) {
	sqlxDB, err := sqlx.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite at %s: %w", path, err)
	}

	// SQLite is single-writer at the file level; serializing through one
	// connection avoids SQLITE_BUSY at the cost of light queuing. Bot
	// traffic doesn't justify the read/write-pool split.
	sqlxDB.SetMaxOpenConns(1)

	// Pragmas — foreign keys default OFF, journal default DELETE. Both
	// matter and both must be set explicitly per-connection (hence the
	// single-conn pool above).
	pragmas := []string{
		`PRAGMA journal_mode = WAL`,
		`PRAGMA foreign_keys = ON`,
		`PRAGMA busy_timeout = 5000`,
		`PRAGMA synchronous = NORMAL`,
	}
	for _, p := range pragmas {
		if _, err := sqlxDB.ExecContext(context.Background(), p); err != nil {
			_ = sqlxDB.Close()
			return nil, fmt.Errorf("apply %q: %w", p, err)
		}
	}

	if err := runMigrations(sqlxDB); err != nil {
		_ = sqlxDB.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}

	return &DB{DB: sqlxDB, prefixCache: newPrefixCache()}, nil
}

func runMigrations(sqlxDB *sqlx.DB) error {
	goose.SetBaseFS(migrations.FS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("set dialect: %w", err)
	}
	// goose wants *sql.DB; sqlx embeds it as .DB.
	return goose.Up(sqlxDB.DB, ".")
}
