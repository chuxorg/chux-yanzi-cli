package yanzilibrary

import (
	"embed"
	"io/fs"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// MigrationsFS exposes embedded migration files for libraryd.
func MigrationsFS() fs.FS {
	return migrationsFS
}
