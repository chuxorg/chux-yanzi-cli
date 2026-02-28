package yanzilibrary

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func TestInitializeCreatesRuntimeState(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	initialized, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	if !initialized {
		t.Fatalf("expected first initialization")
	}

	dir := filepath.Join(home, defaultDBDirName)
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("stat runtime dir: %v", err)
	}

	dbPath := filepath.Join(dir, defaultDBFile)
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("stat db file: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	assertTableExists(t, db, "schema_version")
	assertTableExists(t, db, "schema_migrations")
	assertTableExists(t, db, "intents")
	assertTableExists(t, db, "projects")
	assertTableExists(t, db, "checkpoints")

	var version int
	if err := db.QueryRow(`SELECT version FROM schema_version LIMIT 1`).Scan(&version); err != nil {
		t.Fatalf("query schema_version: %v", err)
	}
	if version != 1 {
		t.Fatalf("expected schema version 1, got %d", version)
	}
}

func TestInitializeIsIdempotent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	first, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize first: %v", err)
	}
	if !first {
		t.Fatalf("expected first initialization")
	}

	second, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize second: %v", err)
	}
	if second {
		t.Fatalf("expected subsequent initialization to be silent")
	}

	dbPath := filepath.Join(home, defaultDBDirName, defaultDBFile)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow(`SELECT COUNT(1) FROM schema_version`).Scan(&count); err != nil {
		t.Fatalf("count schema_version rows: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected single schema_version row, got %d", count)
	}
}

func TestInitializeRecreatesDatabaseWhenDeleted(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(envDBPath, "")

	if _, err := Initialize(); err != nil {
		t.Fatalf("Initialize first: %v", err)
	}

	dbPath := filepath.Join(home, defaultDBDirName, defaultDBFile)
	if err := os.Remove(dbPath); err != nil {
		t.Fatalf("remove db: %v", err)
	}

	initialized, err := Initialize()
	if err != nil {
		t.Fatalf("Initialize after delete: %v", err)
	}
	if !initialized {
		t.Fatalf("expected initialization after db deletion")
	}
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("stat recreated db: %v", err)
	}
}

func assertTableExists(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	var count int
	if err := db.QueryRow(`SELECT COUNT(1) FROM sqlite_master WHERE type = 'table' AND name = ?`, name).Scan(&count); err != nil {
		t.Fatalf("query sqlite_master for %s: %v", name, err)
	}
	if count != 1 {
		t.Fatalf("expected table %s to exist", name)
	}
}
