// Package model provides shared data model types.
package model

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// IntentRecord represents the v0 intent schema persisted and shared across services.
type IntentRecord struct {
	ID         string          `json:"id"`
	CreatedAt  string          `json:"created_at"`
	Author     string          `json:"author"`
	SourceType string          `json:"source_type"`
	Title      string          `json:"title,omitempty"`
	Prompt     string          `json:"prompt"`
	Response   string          `json:"response"`
	Meta       json.RawMessage `json:"meta,omitempty"`
	PrevHash   string          `json:"prev_hash,omitempty"`
	Hash       string          `json:"hash"`
}

// Validate checks required fields for the v0 schema.
func (r IntentRecord) Validate() error {
	if strings.TrimSpace(r.ID) == "" {
		return errors.New("id is required")
	}
	if len(r.CreatedAt) == 0 {
		return errors.New("created_at is required")
	}
	if _, err := time.Parse(time.RFC3339Nano, r.CreatedAt); err != nil {
		return errors.New("created_at must be RFC3339")
	}
	if len(r.Author) == 0 {
		return errors.New("author is required")
	}
	if len(r.SourceType) == 0 {
		return errors.New("source_type is required")
	}
	if len(r.Prompt) == 0 {
		return errors.New("prompt is required")
	}
	if len(r.Response) == 0 {
		return errors.New("response is required")
	}
	if len(r.Hash) == 0 {
		return errors.New("hash is required")
	}
	return nil
}

// Normalize returns a copy with normalized fields for deterministic hashing/storage.
func (r IntentRecord) Normalize() IntentRecord {
	out := r
	out.Author = normalizeNewlines(r.Author)
	out.SourceType = normalizeNewlines(r.SourceType)
	out.Title = normalizeNewlines(r.Title)
	out.Prompt = normalizeNewlines(r.Prompt)
	out.Response = normalizeNewlines(r.Response)
	out.PrevHash = normalizeNewlines(r.PrevHash)
	return out
}

func normalizeNewlines(value string) string {
	if value == "" {
		return value
	}
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return value
}
