package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrApiURLNotAvailable is returned when no row matches the requested
// ApiName, or the row exists but IsActive is false.
var ErrApiURLNotAvailable = errors.New("api url not available")

// ApiURL mirrors a row in the ApiURL table.
type ApiURL struct {
	ID          int64          `db:"ID"`
	ApiName     string         `db:"ApiName"`
	ApiURL      string         `db:"ApiURL"`
	Description sql.NullString `db:"Description"`
	IsActive    bool           `db:"IsActive"`
	CreatedAt   string         `db:"CreatedAt"`
	UpdatedAt   sql.NullString `db:"UpdatedAt"`
}

// GetApiURL returns the URL string for the given ApiName, but only when
// IsActive is true. Missing names and inactive rows both return
// ErrApiURLNotAvailable so callers can fail uniformly.
func (db *DB) GetApiURL(ctx context.Context, name string) (string, error) {
	var url string
	err := db.GetContext(ctx, &url,
		`SELECT ApiURL FROM ApiURL WHERE ApiName = ? AND IsActive = 1`,
		name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrApiURLNotAvailable
	}
	if err != nil {
		return "", fmt.Errorf("get api url %s: %w", name, err)
	}
	return url, nil
}

// GetApiURLRow returns the full ApiURL row including inactive entries.
// Useful for admin-style introspection; callers wanting just the URL with
// activeness enforced should use GetApiURL.
func (db *DB) GetApiURLRow(ctx context.Context, name string) (*ApiURL, error) {
	var a ApiURL
	err := db.GetContext(ctx, &a,
		`SELECT * FROM ApiURL WHERE ApiName = ?`,
		name,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("api url row %s: %w", name, err)
	}
	return &a, nil
}
