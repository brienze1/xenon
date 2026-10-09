package idutils

import (
	"testing"

	"github.com/oklog/ulid/v2"
)

func TestNewUlidReturnsDistinctValidUlids(t *testing.T) {
	first, second := NewUlid(), NewUlid()
	if first == second {
		t.Fatal("ULIDs must be unique")
	}
	for _, id := range []string{first, second} {
		if _, err := ulid.ParseStrict(id); err != nil {
			t.Fatalf("NewUlid() = %q is not a ULID: %v", id, err)
		}
	}
}
