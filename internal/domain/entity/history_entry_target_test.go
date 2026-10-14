package entity

import "testing"

func TestHistoryEntryActivityTarget(t *testing.T) {
	target := HistoryEntry{ID: "h-1"}.ActivityTarget()
	if target != (AIActivityTarget{ID: "h-1"}) {
		t.Errorf("target %+v", target)
	}
}
