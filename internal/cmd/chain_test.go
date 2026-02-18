package cmd

import "testing"

func TestJoinComma(t *testing.T) {
	if got := joinComma(nil); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
	if got := joinComma([]string{"one"}); got != "one" {
		t.Fatalf("expected one, got %q", got)
	}
	if got := joinComma([]string{"one", "two", "three"}); got != "one,two,three" {
		t.Fatalf("expected joined values, got %q", got)
	}
}
