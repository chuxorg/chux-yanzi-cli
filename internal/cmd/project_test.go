package cmd

import (
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

	if !strings.Contains(output, "created_at: ") {
		t.Fatalf("expected created_at output, got %q", output)
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
