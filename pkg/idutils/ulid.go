package idutils

import "github.com/oklog/ulid/v2"

// NewUlid returns a new, time-sortable ULID string.
func NewUlid() string {
	return ulid.Make().String()
}
