package yanzilibrary

import "strings"

// normalizeNewlines converts CRLF/CR line endings to LF for stable storage and hashing.
func normalizeNewlines(value string) string {
	if value == "" {
		return value
	}
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return value
}
