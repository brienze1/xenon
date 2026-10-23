package entity

import "testing"

func TestInEnvironmentRecordsTheActiveEnvironmentAsItIsThen(t *testing.T) {
	active := NewEnvironment("hml", "HML")
	active.Color = "green"
	entry := HistoryEntry{ID: "h1"}.InEnvironment(&active)
	active.Name = "Renamed"
	if entry.Environment == nil || entry.Environment.ID != "hml" || entry.Environment.Name != "HML" || entry.Environment.Color != "green" {
		t.Fatalf("Environment = %+v, want the snapshot", entry.Environment)
	}
	if got := entry.InEnvironment(nil).Environment; got != nil {
		t.Errorf("no active environment records none, got %+v", got)
	}
}
