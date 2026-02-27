package model

import (
	"testing"
)

func TestIntentRecordValidate(t *testing.T) {
	base := IntentRecord{
		ID:         "01HZYFQ7T9ZV54X2G4A8M4J2C1",
		CreatedAt:  "2026-02-09T10:00:00Z",
		Author:     "alice",
		SourceType: "cli",
		Title:      "",
		Prompt:     "hello",
		Response:   "world",
		Hash:       "abc123",
	}

	if err := base.Validate(); err != nil {
		t.Fatalf("expected valid record, got %v", err)
	}

	cases := []struct {
		name    string
		mutate  func(IntentRecord) IntentRecord
		wantErr string
	}{
		{
			name: "missing id",
			mutate: func(r IntentRecord) IntentRecord {
				r.ID = ""
				return r
			},
			wantErr: "id is required",
		},
		{
			name: "missing created_at",
			mutate: func(r IntentRecord) IntentRecord {
				r.CreatedAt = ""
				return r
			},
			wantErr: "created_at is required",
		},
		{
			name: "bad created_at",
			mutate: func(r IntentRecord) IntentRecord {
				r.CreatedAt = "not-a-time"
				return r
			},
			wantErr: "created_at must be RFC3339",
		},
		{
			name: "missing author",
			mutate: func(r IntentRecord) IntentRecord {
				r.Author = ""
				return r
			},
			wantErr: "author is required",
		},
		{
			name: "missing source_type",
			mutate: func(r IntentRecord) IntentRecord {
				r.SourceType = ""
				return r
			},
			wantErr: "source_type is required",
		},
		{
			name: "missing prompt",
			mutate: func(r IntentRecord) IntentRecord {
				r.Prompt = ""
				return r
			},
			wantErr: "prompt is required",
		},
		{
			name: "missing response",
			mutate: func(r IntentRecord) IntentRecord {
				r.Response = ""
				return r
			},
			wantErr: "response is required",
		},
		{
			name: "missing hash",
			mutate: func(r IntentRecord) IntentRecord {
				r.Hash = ""
				return r
			},
			wantErr: "hash is required",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			record := tc.mutate(base)
			err := record.Validate()
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.wantErr)
			}
			if err.Error() != tc.wantErr {
				t.Fatalf("expected error %q, got %q", tc.wantErr, err.Error())
			}
		})
	}
}

func TestIntentRecordNormalizeNewlines(t *testing.T) {
	record := IntentRecord{
		Author:     "alice\r\nline2",
		SourceType: "cli\rline2",
		Title:      "title\r\nline2",
		Prompt:     "prompt\rline2",
		Response:   "resp\r\nline2",
		PrevHash:   "prev\rline2",
	}

	normalized := record.Normalize()
	if normalized.Author != "alice\nline2" {
		t.Fatalf("expected author normalized, got %q", normalized.Author)
	}
	if normalized.SourceType != "cli\nline2" {
		t.Fatalf("expected source_type normalized, got %q", normalized.SourceType)
	}
	if normalized.Title != "title\nline2" {
		t.Fatalf("expected title normalized, got %q", normalized.Title)
	}
	if normalized.Prompt != "prompt\nline2" {
		t.Fatalf("expected prompt normalized, got %q", normalized.Prompt)
	}
	if normalized.Response != "resp\nline2" {
		t.Fatalf("expected response normalized, got %q", normalized.Response)
	}
	if normalized.PrevHash != "prev\nline2" {
		t.Fatalf("expected prev_hash normalized, got %q", normalized.PrevHash)
	}
}
