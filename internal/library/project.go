package yanzilibrary

import "time"

// Project represents a named project namespace in the library ledger.
type Project struct {
	Name        string
	Description string
	CreatedAt   time.Time
}
