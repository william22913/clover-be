// Package seed embeds the fixed v1 catalogue.
package seed

import "embed"

// FS holds movies.sql, applied once at startup after migrations.
//
//go:embed *.sql
var FS embed.FS
