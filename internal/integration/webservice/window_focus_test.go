package webservice

import (
	"context"
	"testing"
)

type countingFocuser struct{ focused int }

func (c *countingFocuser) Focus() { c.focused++ }

func TestWindowFocusBringsTheWindowToTheFront(t *testing.T) {
	focuser := &countingFocuser{}
	NewWindowFocusWebservice(focuser).Focus(context.Background())
	if focuser.focused != 1 {
		t.Fatalf("focused %d times, want 1", focuser.focused)
	}
}
