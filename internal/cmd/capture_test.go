package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestKVPairsToJSONLastValueWins(t *testing.T) {
	pairs := kvPairs{"area=auth", "area=billing", "tags=migration,security"}

	raw, err := pairs.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}

	var meta map[string]string
	if err := json.Unmarshal(raw, &meta); err != nil {
		t.Fatalf("unmarshal meta: %v", err)
	}
	if meta["area"] != "billing" {
		t.Fatalf("expected last value to win for duplicate key, got %q", meta["area"])
	}
	if meta["tags"] != "migration,security" {
		t.Fatalf("unexpected tags value: %q", meta["tags"])
	}
}

func TestKVPairsToJSONMalformedArgument(t *testing.T) {
	pairs := kvPairs{"missing-separator"}

	_, err := pairs.ToJSON()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "expected key=value") {
		t.Fatalf("unexpected error: %v", err)
	}
}
