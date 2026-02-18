package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chuxorg/chux-yanzi-cli/internal/config"
	yanzilibrary "github.com/chuxorg/chux-yanzi-library"
)

func TestProjectUse(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")

	output, err := captureStdout(func() error {
		return RunProject([]string{"use", "alpha"})
	})
	if err != nil {
		t.Fatalf("RunProject use: %v", err)
	}
	if !strings.Contains(output, "Active project set to alpha.") {
		t.Fatalf("unexpected output: %q", output)
	}

	path := filepath.Join(home, ".yanzi", "state.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read state file: %v", err)
	}
	var state struct {
		ActiveProject string `json:"active_project"`
	}
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("decode state file: %v", err)
	}
	if state.ActiveProject != "alpha" {
		t.Fatalf("expected active project alpha, got %q", state.ActiveProject)
	}
}

func TestProjectCurrent(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)
	createTestProject(t, "alpha")

	if err := RunProject([]string{"use", "alpha"}); err != nil {
		t.Fatalf("RunProject use: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunProject([]string{"current"})
	})
	if err != nil {
		t.Fatalf("RunProject current: %v", err)
	}
	if !strings.Contains(output, "Active project: alpha") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestProjectCurrentNone(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	output, err := captureStdout(func() error {
		return RunProject([]string{"current"})
	})
	if err != nil {
		t.Fatalf("RunProject current: %v", err)
	}
	if strings.TrimSpace(output) != "No active project" {
		t.Fatalf("unexpected output: %q", output)
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

func createTestProject(t *testing.T, name string) {
	t.Helper()

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	ctx := context.Background()
	db, closeFn, err := openLocalProjectDB(ctx, cfg)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer func() {
		_ = closeFn()
	}()

	if _, err := yanzilibrary.CreateProject(ctx, db, name, ""); err != nil {
		t.Fatalf("create project: %v", err)
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
