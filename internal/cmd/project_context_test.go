package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
