// Package migrations holds the SQL files goose runs at startup. The
// embed makes the files available without shipping them as a separate
// directory next to the binary.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
