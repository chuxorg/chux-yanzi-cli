package cmd

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuxorg/chux-yanzi-cli/internal/config"
	yanzilibrary "github.com/chuxorg/yanzi-library"
)

func TestCheckpointCreateSuccess(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	output, err := captureStdout(func() error {
		return RunCheckpoint([]string{"create", "--summary", "first checkpoint"})
	})
	if err != nil {
		t.Fatalf("RunCheckpoint create: %v", err)
	}
	if !strings.Contains(output, "id: ") {
		t.Fatalf("expected id output, got %q", output)
	}
	if !strings.Contains(output, "summary: first checkpoint") {
		t.Fatalf("expected summary output, got %q", output)
	}
}

func TestCheckpointCreateNoActiveProject(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	err := RunCheckpoint([]string{"create", "--summary", "first checkpoint"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "no active project") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCheckpointListEmpty(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")

	output, err := captureStdout(func() error {
		return RunCheckpoint([]string{"list"})
	})
	if err != nil {
		t.Fatalf("RunCheckpoint list: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected header only, got %q", output)
	}
	if lines[0] != "Index\tCreatedAt\tSummary" {
		t.Fatalf("unexpected header: %q", lines[0])
	}
}

func TestCheckpointListPopulated(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")
	writeStateFile(t, home, "alpha")
	createTestCheckpoint(t, "alpha", "first")
	createTestCheckpoint(t, "alpha", "second")

	output, err := captureStdout(func() error {
		return RunCheckpoint([]string{"list"})
	})
	if err != nil {
		t.Fatalf("RunCheckpoint list: %v", err)
	}
	if !strings.Contains(output, "first") || !strings.Contains(output, "second") {
		t.Fatalf("expected summaries in output, got %q", output)
	}
}

func writeTestConfig(t *testing.T, home string) {
	t.Helper()

	stateDir := filepath.Join(home, ".yanzi")
	if err := os.MkdirAll(stateDir, 0o700); err != nil {
		t.Fatalf("create state dir: %v", err)
	}
	dbPath := filepath.Join(stateDir, "yanzi.db")
	configPath := filepath.Join(stateDir, "config.yaml")
	content := []byte("mode: local\ndb_path: " + dbPath + "\n")
	if err := os.WriteFile(configPath, content, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
}

func writeStateFile(t *testing.T, home, project string) {
	t.Helper()

	stateDir := filepath.Join(home, ".yanzi")
	path := filepath.Join(stateDir, "state.json")
	payload := map[string]string{"active_project": project}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode state: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o600); err != nil {
		t.Fatalf("write state file: %v", err)
	}
}

func createTestProject(t *testing.T, name string) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	ctx := context.Background()
	db, closeFn, err := openLocalCheckpointDB(ctx, cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = closeFn()
	}()

	seedProject(t, db, name)
}

func createTestCheckpoint(t *testing.T, project, summary string) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	ctx := context.Background()
	db, closeFn, err := openLocalCheckpointDB(ctx, cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = closeFn()
	}()

	if _, err := yanzilibrary.CreateCheckpoint(ctx, db, project, summary, []string{}); err != nil {
		t.Fatalf("create checkpoint: %v", err)
	}
}

func seedProject(t *testing.T, db *sql.DB, name string) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO projects (name, description, created_at, prev_hash, hash) VALUES (?, ?, ?, ?, ?)`,
		name,
		nil,
		"2025-01-01T00:00:00Z",
		nil,
		"seed-hash",
	)
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}
}

func captureStdout(fn func() error) (string, error) {
	reader, writer, err := os.Pipe()
	if err != nil {
		return "", err
	}

	stdout := os.Stdout
	os.Stdout = writer
	defer func() {
		os.Stdout = stdout
	}()

	runErr := fn()
	_ = writer.Close()

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	_ = reader.Close()

	return buf.String(), runErr
}
