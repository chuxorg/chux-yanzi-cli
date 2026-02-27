package yanzilibrary

import (
	"strings"
	"time"
)

// Checkpoint represents an immutable checkpoint artifact.
type Checkpoint struct {
	Project              string   `json:"project"`
	Summary              string   `json:"summary"`
	CreatedAt            string   `json:"created_at"`
	ArtifactIDs          []string `json:"artifact_ids"`
	PreviousCheckpointID string   `json:"previous_checkpoint_id,omitempty"`
	Hash                 string   `json:"hash"`
}

// CheckpointValidationError reports invalid checkpoint input.
type CheckpointValidationError struct {
	Field   string
	Message string
}

// Error returns the validation error as "<field> <message>".
func (e CheckpointValidationError) Error() string {
	return e.Field + " " + e.Message
}

// ProjectNotFoundError indicates the referenced project was not found.
type ProjectNotFoundError struct {
	Name string
}

// Error returns the project-not-found message including the project name.
func (e ProjectNotFoundError) Error() string {
	return "project not found: " + e.Name
}

// Validate checks required fields for a checkpoint record.
func (c Checkpoint) Validate() error {
	if strings.TrimSpace(c.Project) == "" {
		return CheckpointValidationError{Field: "project", Message: "is required"}
	}
	if strings.TrimSpace(c.Summary) == "" {
		return CheckpointValidationError{Field: "summary", Message: "is required"}
	}
	if c.CreatedAt == "" {
		return CheckpointValidationError{Field: "created_at", Message: "is required"}
	}
	if _, err := time.Parse(time.RFC3339Nano, c.CreatedAt); err != nil {
		return CheckpointValidationError{Field: "created_at", Message: "must be RFC3339"}
	}
	if c.Hash == "" {
		return CheckpointValidationError{Field: "hash", Message: "is required"}
	}
	return nil
}

// Normalize returns a copy with normalized fields for deterministic hashing/storage.
func (c Checkpoint) Normalize() Checkpoint {
	out := c
	out.Project = normalizeNewlines(strings.TrimSpace(c.Project))
	out.Summary = normalizeNewlines(strings.TrimSpace(c.Summary))
	out.PreviousCheckpointID = normalizeNewlines(c.PreviousCheckpointID)
	if len(out.ArtifactIDs) > 0 {
		ids := make([]string, len(out.ArtifactIDs))
		for i, id := range out.ArtifactIDs {
			ids[i] = normalizeNewlines(id)
		}
		out.ArtifactIDs = ids
	}
	return out
}
