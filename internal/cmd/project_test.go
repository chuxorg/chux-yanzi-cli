package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunProjectCreate(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	output, err := captureStdout(func() error {
		return RunProject([]string{"create", "alpha", "--description", "first project"})
	})
	if err != nil {
		t.Fatalf("RunProject create: %v", err)
	}

	if !strings.Contains(output, "hash: ") {
		t.Fatalf("expected hash output, got %q", output)
	}
	if !strings.Contains(output, "Project created.") {
		t.Fatalf("expected confirmation output, got %q", output)
	}
}

func TestRunProjectList(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	writeTestConfig(t, home)

	if err := RunProject([]string{"create", "alpha", "--description", "first project"}); err != nil {
		t.Fatalf("RunProject create: %v", err)
	}

	output, err := captureStdout(func() error {
		return RunProject([]string{"list"})
	})
	if err != nil {
		t.Fatalf("RunProject list: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header + rows, got %q", output)
	}
	if lines[0] != "Name\tCreatedAt\tDescription" {
		t.Fatalf("unexpected header: %q", lines[0])
	}
	if !strings.Contains(output, "alpha") {
		t.Fatalf("expected project name in output, got %q", output)
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
