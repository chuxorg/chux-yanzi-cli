package yanzilibrary

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// HashCheckpoint computes a deterministic SHA-256 hash for a Checkpoint.
// The hash preimage excludes the hash field and uses canonical field order.
func HashCheckpoint(checkpoint Checkpoint) (string, error) {
	normalized := checkpoint.Normalize()
	preimage, err := canonicalCheckpointPreimage(normalized)
	if err != nil {
		return "", err
	}

	sum := sha256.Sum256(preimage)
	return hex.EncodeToString(sum[:]), nil
}

// canonicalCheckpointPreimage renders a normalized checkpoint payload in canonical JSON key order.
func canonicalCheckpointPreimage(checkpoint Checkpoint) ([]byte, error) {
	if strings.TrimSpace(checkpoint.Project) == "" {
		return nil, errors.New("project is required for hashing")
	}
	if strings.TrimSpace(checkpoint.Summary) == "" {
		return nil, errors.New("summary is required for hashing")
	}
	if checkpoint.CreatedAt == "" {
		return nil, errors.New("created_at is required for hashing")
	}
	createdAt, err := normalizeRFC3339(checkpoint.CreatedAt)
	if err != nil {
		return nil, errors.New("created_at must be RFC3339")
	}

	artifactIDs := checkpoint.ArtifactIDs
	if artifactIDs == nil {
		artifactIDs = []string{}
	}
	artifactJSON, err := json.Marshal(artifactIDs)
	if err != nil {
		return nil, err
	}

	var b strings.Builder
	b.WriteByte('{')
	first := true

	addStringField(&b, &first, "project", checkpoint.Project)
	addStringField(&b, &first, "created_at", createdAt)
	addStringField(&b, &first, "summary", checkpoint.Summary)
	addRawField(&b, &first, "artifact_ids", artifactJSON)
	if checkpoint.PreviousCheckpointID != "" {
		addStringField(&b, &first, "previous_checkpoint_id", checkpoint.PreviousCheckpointID)
	}
	b.WriteByte('}')

	return []byte(b.String()), nil
}

// normalizeRFC3339 parses and rewrites timestamps into canonical UTC RFC3339Nano form.
func normalizeRFC3339(value string) (string, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return "", err
	}
	return parsed.UTC().Format(time.RFC3339Nano), nil
}

// addStringField appends a quoted JSON string field to a builder preserving caller-supplied order.
func addStringField(b *strings.Builder, first *bool, name string, value string) {
	if !*first {
		b.WriteByte(',')
	}
	*first = false
	b.WriteByte('"')
	b.WriteString(name)
	b.WriteString(`":`)
	encoded, _ := json.Marshal(value)
	b.Write(encoded)
}

// addRawField appends a JSON field whose value is already encoded.
func addRawField(b *strings.Builder, first *bool, name string, raw json.RawMessage) {
	if !*first {
		b.WriteByte(',')
	}
	*first = false
	b.WriteByte('"')
	b.WriteString(name)
	b.WriteString(`":`)
	b.Write(raw)
}
