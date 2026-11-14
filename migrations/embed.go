// Package migrations holds the embedded SQL migration files applied by
// internal/db at startup. The files are applied in filename order and are
// never edited after they have shipped; a change is a new file.
package migrations

import "embed"

// FS contains every .sql file in this directory.
//
//go:embed *.sql
var FS embed.FS
