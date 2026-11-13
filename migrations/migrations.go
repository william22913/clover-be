// Package migrations embeds the numbered SQL files applied at startup.
package migrations

import "embed"

// FS holds the ordered migration files (0001_init.sql, ...).
//
//go:embed *.sql
var FS embed.FS
