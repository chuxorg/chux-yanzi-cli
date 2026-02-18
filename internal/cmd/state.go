package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type projectState struct {
	ActiveProject string `json:"active_project"`
}

func loadActiveProject() (string, error) {
	path, err := statePath()
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("read state file: %w", err)
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return "", nil
	}

	var state projectState
	if err := json.Unmarshal(data, &state); err != nil {
		return "", fmt.Errorf("invalid state file: %w", err)
	}

	return strings.TrimSpace(state.ActiveProject), nil
}

func statePath() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("resolve working dir: %w", err)
	}
	return filepath.Join(cwd, ".yanzi", "state.json"), nil
}
