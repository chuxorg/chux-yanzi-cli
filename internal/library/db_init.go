package yanzilibrary

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

const (
	envDBPath        = "YANZI_DB_PATH"
	defaultDBDirName = ".yanzi"
	defaultDBFile    = "yanzi.db"
)

var (
	resolvedDBPath string
	resolvedMu     sync.RWMutex
)

// InitDB resolves the database path, ensures migrations, and returns a SQLite handle.
func InitDB() (*sql.DB, error) {
	path, err := resolveDBPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA foreign_keys=ON;`); err != nil {
		_ = db.Close()
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA busy_timeout=5000;`); err != nil {
		_ = db.Close()
		return nil, err
	}

	if err := migrateDB(context.Background(), db); err != nil {
		_ = db.Close()
		return nil, err
	}

	setResolvedDBPath(path)
	return db, nil
}

// ResolvedDBPath returns the most recently resolved database path.
func ResolvedDBPath() string {
	resolvedMu.RLock()
	defer resolvedMu.RUnlock()
	return resolvedDBPath
}

// setResolvedDBPath updates the in-memory record of the database path used by InitDB.
func setResolvedDBPath(path string) {
	resolvedMu.Lock()
	resolvedDBPath = path
	resolvedMu.Unlock()
}

// resolveDBPath determines the SQLite path from YANZI_DB_PATH or the default ~/.yanzi/yanzi.db.
func resolveDBPath() (string, error) {
	if override := strings.TrimSpace(os.Getenv(envDBPath)); override != "" {
		return override, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home dir: %w", err)
	}

	dir := filepath.Join(home, defaultDBDirName)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("ensure db dir: %w", err)
	}

	return filepath.Join(dir, defaultDBFile), nil
}

const schemaMigrationsTable = `
CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	applied_at TEXT NOT NULL
);
`

// migrateDB applies embedded SQL migrations that have not yet been recorded.
func migrateDB(ctx context.Context, db *sql.DB) error {
	if db == nil {
		return errors.New("database is nil")
	}

	if _, err := db.ExecContext(ctx, schemaMigrationsTable); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	paths, err := listMigrationFiles(MigrationsFS())
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return errors.New("no migration files found")
	}

	sort.Strings(paths)
	for _, path := range paths {
		version := filepath.Base(path)
		applied, err := isMigrationApplied(ctx, db, version)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		contents, err := fs.ReadFile(MigrationsFS(), path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", version, err)
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, applied_at) VALUES (?, ?)`, version, time.Now().UTC().Format(time.RFC3339Nano)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}

	return nil
}

// listMigrationFiles collects migration SQL files from the embedded migrations directory.
func listMigrationFiles(fsys fs.FS) ([]string, error) {
	entries, err := fs.ReadDir(fsys, "migrations")
	if err != nil {
		return nil, fmt.Errorf("list migrations: %w", err)
	}

	paths := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}
		paths = append(paths, filepath.Join("migrations", name))
	}
	return paths, nil
}

// isMigrationApplied reports whether a migration version is present in schema_migrations.
func isMigrationApplied(ctx context.Context, db *sql.DB, version string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(1) FROM schema_migrations WHERE version = ?`, version).Scan(&count); err != nil {
		return false, fmt.Errorf("check migration %s: %w", version, err)
	}
	return count > 0, nil
}
